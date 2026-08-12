package database

import (
	"context"
	"regexp"
	"testing"

	"germinaStack/model"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPostgresReactionRepositoryTogglesReactionAndReadsCounters(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM posts WHERE id = $1")).WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(7)))
	mock.ExpectExec(regexp.QuoteMeta("SELECT reaction($1, $2, $3, $4)")).WithArgs(int64(42), int64(7), "post", "like").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT reaction_type FROM reactions WHERE id_user = $1 AND id_post = $2")).WithArgs(int64(42), int64(7)).WillReturnRows(sqlmock.NewRows([]string{"reaction_type"}).AddRow("like"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT likes, dislikes FROM posts WHERE id = $1")).WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"likes", "dislikes"}).AddRow(int64(3), int64(1)))

	got, err := NewPostgresReactionRepository(db).ToggleReaction(context.Background(), 42, "post", 7, model.ReactionTypeLike)
	if err != nil || got.Reaction == nil || *got.Reaction != model.ReactionTypeLike || got.Likes != 3 || got.Dislikes != 1 {
		t.Fatalf("ToggleReaction() = %#v, error = %v", got, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL expectations: %v", err)
	}
}

func TestPostgresNotificationRepositoryListsOnlyUserNotificationsAndMarksRead(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	listQuery := `SELECT id, id_post, id_user, text_show, is_read, created_at FROM notifications WHERE id_user = $1 ORDER BY created_at DESC, id DESC`
	mock.ExpectQuery(regexp.QuoteMeta(listQuery)).WithArgs(int64(42)).WillReturnRows(sqlmock.NewRows([]string{"id", "id_post", "id_user", "text_show", "is_read", "created_at"}).AddRow(int64(1), int64(7), int64(42), "mentioned", false, nil))
	mock.ExpectExec(regexp.QuoteMeta("SELECT mark_notifications_as_read($1)")).WithArgs(int64(42)).WillReturnResult(sqlmock.NewResult(0, 0))

	repository := NewPostgresNotificationRepository(db)
	items, err := repository.ListNotifications(context.Background(), 42)
	if err != nil || len(items) != 1 || items[0].UserID != 42 || items[0].PostID == nil || *items[0].PostID != 7 {
		t.Fatalf("ListNotifications() = %#v, error = %v", items, err)
	}
	if err := repository.MarkNotificationsAsRead(context.Background(), 42); err != nil {
		t.Fatalf("MarkNotificationsAsRead() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL expectations: %v", err)
	}
}
