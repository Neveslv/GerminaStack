package database

import (
	"context"
	"database/sql"
	"fmt"

	"germinaStack/model"
)

type PostgresNotificationRepository struct {
	db *sql.DB
}

func NewPostgresNotificationRepository(db *sql.DB) *PostgresNotificationRepository {
	return &PostgresNotificationRepository{db: db}
}

func (r *PostgresNotificationRepository) ListNotifications(ctx context.Context, userID int64) ([]model.Notification, error) {
	const query = `SELECT id, id_post, id_user, text_show, is_read, created_at FROM notifications WHERE id_user = $1 ORDER BY created_at DESC, id DESC`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()
	notifications := make([]model.Notification, 0)
	for rows.Next() {
		var notification model.Notification
		if err := rows.Scan(&notification.ID, &notification.PostID, &notification.UserID, &notification.TextShow, &notification.IsRead, &notification.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		notifications = append(notifications, notification)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notifications: %w", err)
	}
	return notifications, nil
}

func (r *PostgresNotificationRepository) MarkNotificationsAsRead(ctx context.Context, userID int64) error {
	if _, err := r.db.ExecContext(ctx, `SELECT mark_notifications_as_read($1)`, userID); err != nil {
		return fmt.Errorf("mark notifications as read: %w", err)
	}
	return nil
}
