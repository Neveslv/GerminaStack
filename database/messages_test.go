package database

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"germinaStack/model"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestPostgresMessageRepositoryCreatesPostThroughFunctionAndReadsIt(t *testing.T) {
	db, mock := newMessagesMock(t)
	created := time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT create_message($1, $2, $3, $4, $5, $6, $7)")).WithArgs("post", int64(42), int64(7), "Body", "Title", "image.png", "Image").WillReturnRows(sqlmock.NewRows([]string{"create_message"}).AddRow(int64(99)))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, id_user, id_subject, title, image_url, image_description, content, likes, dislikes, created_at FROM posts WHERE id = $1")).WithArgs(int64(99)).WillReturnRows(sqlmock.NewRows([]string{"id", "id_user", "id_subject", "title", "image_url", "image_description", "content", "likes", "dislikes", "created_at"}).
		AddRow(int64(99), int64(42), int64(7), "Title", "image.png", "Image", "Body", int64(2), int64(1), created))

	got, err := NewPostgresMessageRepository(db).CreatePost(context.Background(), 42, 7, "Body", stringPointer("image.png"), stringPointer("Image"), "Title")
	if err != nil {
		t.Fatalf("CreatePost() error = %v", err)
	}
	if got.ID != 99 || got.UserID != 42 || got.SubjectID != 7 || got.Title != "Title" || got.Content != "Body" || got.CreatedAt == nil || !got.CreatedAt.Equal(created) {
		t.Fatalf("CreatePost() = %#v", got)
	}
	assertMessagesExpectations(t, mock)
}

func TestPostgresMessageRepositoryCreatesCommentsThroughFunction(t *testing.T) {
	tests := []struct {
		name        string
		messageType string
		create      func(*PostgresMessageRepository) (any, error)
		selectSQL   string
		wantID      int64
	}{
		{
			name:        "comment",
			messageType: "comment",
			create: func(repository *PostgresMessageRepository) (any, error) {
				return repository.CreateComment(context.Background(), 42, 99, "Comment")
			},
			selectSQL: "SELECT id, id_post, id_user, content, likes, dislikes, created_at FROM comments WHERE id = $1",
			wantID:    100,
		},
		{
			name:        "reply",
			messageType: "comment_on_comment",
			create: func(repository *PostgresMessageRepository) (any, error) {
				return repository.CreateReply(context.Background(), 42, 100, "Reply")
			},
			selectSQL: "SELECT id, id_comment, id_user, content, likes, dislikes, created_at FROM comments_on_comments WHERE id = $1",
			wantID:    101,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db, mock := newMessagesMock(t)
			parentID := int64(99)
			content := "Comment"
			if test.messageType == "comment_on_comment" {
				parentID = 100
				content = "Reply"
			}
			mock.ExpectQuery(regexp.QuoteMeta("SELECT create_message($1, $2, $3, $4, $5, $6, $7)")).WithArgs(test.messageType, int64(42), parentID, content, nil, nil, nil).WillReturnRows(sqlmock.NewRows([]string{"create_message"}).AddRow(test.wantID))
			mock.ExpectQuery(regexp.QuoteMeta(test.selectSQL)).
				WithArgs(test.wantID).
				WillReturnRows(sqlmock.NewRows([]string{"id", "parent", "id_user", "content", "likes", "dislikes", "created_at"}).
					AddRow(test.wantID, parentID, int64(42), test.name, int64(0), int64(0), nil))

			got, err := test.create(NewPostgresMessageRepository(db))
			if err != nil {
				t.Fatalf("create error = %v", err)
			}
			if gotID := messageID(got); gotID != test.wantID {
				t.Fatalf("created ID = %d, want %d", gotID, test.wantID)
			}
			assertMessagesExpectations(t, mock)
		})
	}
}

func TestPostgresMessageRepositoryListsMessagesDeterministically(t *testing.T) {
	db, mock := newMessagesMock(t)
	created := time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)
	postsQuery := "SELECT id, id_user, id_subject, title, image_url, image_description, content, likes, dislikes, created_at FROM posts WHERE id_subject = $1 ORDER BY created_at DESC, id DESC"
	mock.ExpectQuery(regexp.QuoteMeta(postsQuery)).WithArgs(int64(7)).WillReturnRows(
		sqlmock.NewRows([]string{"id", "id_user", "id_subject", "title", "image_url", "image_description", "content", "likes", "dislikes", "created_at"}).
			AddRow(int64(2), int64(42), int64(7), "New", nil, nil, "new", int64(0), int64(0), created).
			AddRow(int64(1), int64(42), int64(7), "Old", nil, nil, "old", int64(0), int64(0), created))

	got, err := NewPostgresMessageRepository(db).ListPosts(context.Background(), int64Pointer(7))
	if err != nil || len(got) != 2 || got[0].ID != 2 || got[1].ID != 1 {
		t.Fatalf("ListPosts() = %#v, error = %v", got, err)
	}
	assertMessagesExpectations(t, mock)
}

func TestPostgresMessageRepositoryReadsParentsAndMapsMissingRows(t *testing.T) {
	db, mock := newMessagesMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, id_user, id_subject, title, image_url, image_description, content, likes, dislikes, created_at FROM posts WHERE id = $1")).
		WithArgs(int64(404)).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, id_post, id_user, content, likes, dislikes, created_at FROM comments WHERE id = $1")).
		WithArgs(int64(405)).WillReturnError(sql.ErrNoRows)

	repository := NewPostgresMessageRepository(db)
	if _, err := repository.GetPost(context.Background(), 404); !errors.Is(err, ErrMessageNotFound) {
		t.Fatalf("GetPost() error = %v, want ErrMessageNotFound", err)
	}
	if _, err := repository.GetComment(context.Background(), 405); !errors.Is(err, ErrMessageNotFound) {
		t.Fatalf("GetComment() error = %v, want ErrMessageNotFound", err)
	}
	assertMessagesExpectations(t, mock)
}

func TestPostgresMessageRepositoryMapsForeignKeyAndWrapsUnexpectedCreationErrors(t *testing.T) {
	for _, test := range []struct {
		name string
		err  error
		want error
	}{
		{name: "missing parent", err: &pgconn.PgError{Code: "23503", Message: "private detail"}, want: ErrMessageParentNotFound},
		{name: "database unavailable", err: context.DeadlineExceeded, want: context.DeadlineExceeded},
	} {
		t.Run(test.name, func(t *testing.T) {
			db, mock := newMessagesMock(t)
			mock.ExpectQuery(regexp.QuoteMeta("SELECT create_message($1, $2, $3, $4, $5, $6, $7)")).
				WillReturnError(test.err)

			_, err := NewPostgresMessageRepository(db).CreateComment(context.Background(), 42, 404, "Comment")
			if !errors.Is(err, test.want) {
				t.Fatalf("CreateComment() error = %v, want errors.Is(_, %v)", err, test.want)
			}
			if test.name == "missing parent" && err.Error() != ErrMessageParentNotFound.Error() {
				t.Fatalf("CreateComment() exposed database detail: %q", err)
			}
			assertMessagesExpectations(t, mock)
		})
	}
}

func messageID(value any) int64 {
	switch message := value.(type) {
	case model.Comment:
		return message.ID
	case model.CommentOnComment:
		return message.ID
	default:
		return 0
	}
}

func stringPointer(value string) *string { return &value }

func newMessagesMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db, mock
}

func assertMessagesExpectations(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL expectations: %v", err)
	}
}
