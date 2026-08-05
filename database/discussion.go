package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"germinaStack/model"

	"github.com/jackc/pgx/v5/pgconn"
)

var ErrDiscussionNotFound = errors.New("discussion item not found")

type PostFilter struct {
	SubjectID  *int64
	AuthorID   *int64
	Pagination Pagination
}

type NotificationFilter struct {
	Unread     bool
	Pagination Pagination
}

type PostgresDiscussionRepository struct{ db *sql.DB }

func NewPostgresDiscussionRepository(db *sql.DB) *PostgresDiscussionRepository {
	return &PostgresDiscussionRepository{db: db}
}

func (r *PostgresDiscussionRepository) ListPosts(ctx context.Context, filter PostFilter) ([]model.Post, error) {
	query := `SELECT id, id_user, id_subject, title, image_url, image_description, content, likes, dislikes, created_at
FROM posts`
	args := make([]any, 0, 4)
	switch {
	case filter.SubjectID != nil && filter.AuthorID != nil:
		query += ` WHERE id_subject = $1 AND id_user = $2`
		args = append(args, *filter.SubjectID, *filter.AuthorID)
	case filter.SubjectID != nil:
		query += ` WHERE id_subject = $1`
		args = append(args, *filter.SubjectID)
	case filter.AuthorID != nil:
		query += ` WHERE id_user = $1`
		args = append(args, *filter.AuthorID)
	}
	query += fmt.Sprintf(` ORDER BY created_at DESC, id DESC LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
	args = append(args, filter.Pagination.PageSize, pageOffset(filter.Pagination))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
	}
	defer rows.Close()
	posts := make([]model.Post, 0)
	for rows.Next() {
		var post model.Post
		if err := rows.Scan(&post.ID, &post.UserID, &post.SubjectID, &post.Title, &post.ImageURL, &post.ImageDescription, &post.Content, &post.Likes, &post.Dislikes, &post.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan post: %w", err)
		}
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate posts: %w", err)
	}
	return posts, nil
}

func (r *PostgresDiscussionRepository) ListComments(ctx context.Context, postID int64, pagination Pagination) ([]model.Comment, error) {
	const query = `SELECT id, id_post, id_user, content, likes, dislikes, created_at
FROM comments
WHERE id_post = $1
ORDER BY created_at DESC, id DESC
LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, postID, pagination.PageSize, pageOffset(pagination))
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}
	defer rows.Close()
	comments := make([]model.Comment, 0)
	for rows.Next() {
		var comment model.Comment
		if err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Content, &comment.Likes, &comment.Dislikes, &comment.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan comment: %w", err)
		}
		comments = append(comments, comment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate comments: %w", err)
	}
	return comments, nil
}

func (r *PostgresDiscussionRepository) ListReplies(ctx context.Context, commentID int64, pagination Pagination) ([]model.CommentOnComment, error) {
	const query = `SELECT id, id_comment, id_user, content, likes, dislikes, created_at
FROM comments_on_comments
WHERE id_comment = $1
ORDER BY created_at DESC, id DESC
LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, commentID, pagination.PageSize, pageOffset(pagination))
	if err != nil {
		return nil, fmt.Errorf("list replies: %w", err)
	}
	defer rows.Close()
	replies := make([]model.CommentOnComment, 0)
	for rows.Next() {
		var reply model.CommentOnComment
		if err := rows.Scan(&reply.ID, &reply.CommentID, &reply.UserID, &reply.Content, &reply.Likes, &reply.Dislikes, &reply.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan reply: %w", err)
		}
		replies = append(replies, reply)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate replies: %w", err)
	}
	return replies, nil
}

func (r *PostgresDiscussionRepository) React(ctx context.Context, userID, messageID int64, messageType string, reactionType model.ReactionType) error {
	_, err := r.db.ExecContext(ctx, `SELECT reaction($1, $2, $3, $4)`, userID, messageID, messageType, reactionType)
	return discussionMutationError("react", err)
}

func (r *PostgresDiscussionRepository) ListNotifications(ctx context.Context, userID int64, filter NotificationFilter) ([]model.Notification, error) {
	query := `SELECT id, id_post, id_user, text_show, is_read, created_at
FROM notifications
WHERE id_user = $1`
	args := []any{userID}
	if filter.Unread {
		query += ` AND is_read = FALSE`
	}
	query += fmt.Sprintf(` ORDER BY created_at DESC, id DESC LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
	args = append(args, filter.Pagination.PageSize, pageOffset(filter.Pagination))

	rows, err := r.db.QueryContext(ctx, query, args...)
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

func (r *PostgresDiscussionRepository) MarkNotificationsRead(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `SELECT mark_notifications_as_read($1)`, userID)
	return discussionMutationError("mark notifications read", err)
}

func pageOffset(pagination Pagination) int64 {
	return int64(pagination.Page-1) * int64(pagination.PageSize)
}

func discussionMutationError(operation string, err error) error {
	if err == nil {
		return nil
	}
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && (postgresError.Code == "23503" || postgresError.Code == "P0001") {
		return ErrDiscussionNotFound
	}
	return fmt.Errorf("%s: %w", operation, err)
}
