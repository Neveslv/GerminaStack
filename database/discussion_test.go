package database

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"germinaStack/model"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestPostgresDiscussionRepositoryListsPostsWithFiltersAndPagination(t *testing.T) {
	t.Parallel()
	db, mock := newCatalogMock(t)
	subjectID, authorID := int64(3), int64(7)
	created := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	const query = `SELECT id, id_user, id_subject, title, image_url, image_description, content, likes, dislikes, created_at
FROM posts WHERE id_subject = $1 AND id_user = $2 ORDER BY created_at DESC, id DESC LIMIT $3 OFFSET $4`
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(subjectID, authorID, 20, int64(20)).WillReturnRows(
		sqlmock.NewRows([]string{"id", "id_user", "id_subject", "title", "image_url", "image_description", "content", "likes", "dislikes", "created_at"}).
			AddRow(int64(9), authorID, subjectID, "Title", nil, nil, "Content", int64(2), int64(1), created),
	)

	posts, err := NewPostgresDiscussionRepository(db).ListPosts(context.Background(), PostFilter{SubjectID: &subjectID, AuthorID: &authorID, Pagination: Pagination{Page: 2, PageSize: 20}})
	if err != nil || len(posts) != 1 || posts[0].ID != 9 || posts[0].CreatedAt == nil || !posts[0].CreatedAt.Equal(created) {
		t.Fatalf("ListPosts() = (%#v, %v)", posts, err)
	}
	assertCatalogExpectations(t, mock)
}

func TestPostgresDiscussionRepositoryReturnsEmptyReadLists(t *testing.T) {
	t.Parallel()
	db, mock := newCatalogMock(t)
	mock.ExpectQuery("SELECT id, id_post, id_user, content").WillReturnRows(sqlmock.NewRows([]string{"id", "id_post", "id_user", "content", "likes", "dislikes", "created_at"}))
	mock.ExpectQuery("SELECT id, id_comment, id_user, content").WillReturnRows(sqlmock.NewRows([]string{"id", "id_comment", "id_user", "content", "likes", "dislikes", "created_at"}))
	mock.ExpectQuery("SELECT id, id_post, id_user, text_show").WillReturnRows(sqlmock.NewRows([]string{"id", "id_post", "id_user", "text_show", "is_read", "created_at"}))
	repository := NewPostgresDiscussionRepository(db)
	pagination := Pagination{Page: 1, PageSize: 20}
	comments, commentErr := repository.ListComments(context.Background(), 1, pagination)
	replies, replyErr := repository.ListReplies(context.Background(), 1, pagination)
	notifications, notificationErr := repository.ListNotifications(context.Background(), 1, NotificationFilter{Pagination: pagination})
	if commentErr != nil || replyErr != nil || notificationErr != nil || comments == nil || replies == nil || notifications == nil {
		t.Fatalf("lists/errors = %#v/%#v/%#v; %v/%v/%v", comments, replies, notifications, commentErr, replyErr, notificationErr)
	}
	assertCatalogExpectations(t, mock)
}

func TestPostgresDiscussionRepositoryUsesDatabaseFunctions(t *testing.T) {
	t.Parallel()
	db, mock := newCatalogMock(t)
	mock.ExpectExec(regexp.QuoteMeta(`SELECT reaction($1, $2, $3, $4)`)).WithArgs(int64(2), int64(3), "post", model.ReactionTypeLike).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`SELECT mark_notifications_as_read($1)`)).WithArgs(int64(2)).WillReturnResult(sqlmock.NewResult(0, 1))
	repository := NewPostgresDiscussionRepository(db)
	if err := repository.React(context.Background(), 2, 3, "post", model.ReactionTypeLike); err != nil {
		t.Fatalf("React() error = %v", err)
	}
	if err := repository.MarkNotificationsRead(context.Background(), 2); err != nil {
		t.Fatalf("MarkNotificationsRead() error = %v", err)
	}
	assertCatalogExpectations(t, mock)
}

func TestPostgresDiscussionRepositoryCreatesMessagesThroughDatabaseFunction(t *testing.T) {
	t.Parallel()
	db, mock := newCatalogMock(t)
	const createQuery = `SELECT create_message($1, $2, $3, $4, $5, $6, $7)`
	mock.ExpectQuery(regexp.QuoteMeta(createQuery)).WithArgs("post", int64(42), int64(3), "Body", "Title", nil, nil).WillReturnRows(sqlmock.NewRows([]string{"create_message"}).AddRow(int64(8)))
	mock.ExpectQuery("SELECT id, id_user, id_subject, title").WithArgs(int64(8)).WillReturnRows(
		sqlmock.NewRows([]string{"id", "id_user", "id_subject", "title", "image_url", "image_description", "content", "likes", "dislikes", "created_at"}).
			AddRow(int64(8), int64(42), int64(3), "Title", nil, nil, "Body", int64(0), int64(0), nil),
	)
	post, err := NewPostgresDiscussionRepository(db).CreatePost(context.Background(), 42, PostInput{SubjectID: 3, Title: "Title", Content: "Body"})
	if err != nil || post.ID != 8 || post.UserID != 42 {
		t.Fatalf("CreatePost() = (%#v, %v)", post, err)
	}
	assertCatalogExpectations(t, mock)
}

func TestDiscussionMutationErrorHidesDatabaseDetails(t *testing.T) {
	t.Parallel()
	err := discussionMutationError("react", &pgconn.PgError{Code: "23503", Message: "private detail"})
	if !errors.Is(err, ErrDiscussionNotFound) || err.Error() != ErrDiscussionNotFound.Error() {
		t.Fatalf("discussionMutationError() = %v", err)
	}
}
