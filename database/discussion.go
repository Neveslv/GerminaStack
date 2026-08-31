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
var ErrDiscussionInvalid = errors.New("invalid discussion item")

type PostInput struct {
	SubjectID        int64
	Title            string
	ImageURL         *string
	ImageDescription *string
	Content          string
}

type CommentInput struct{ Content string }

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

func (r *PostgresDiscussionRepository) GetPost(ctx context.Context, postID int64) (model.Post, error) {
	const query = `SELECT p.id, p.id_user, p.id_subject, p.title, p.image_url, p.image_description, p.content, p.likes, p.dislikes,
	       (SELECT COUNT(*) FROM comments WHERE id_post = p.id), u.name, u.username, u.profile_image_url, u.profile_image_description, p.created_at
FROM posts p JOIN users u ON u.id = p.id_user WHERE p.id = $1`
	var post model.Post
	err := r.db.QueryRowContext(ctx, query, postID).Scan(&post.ID, &post.UserID, &post.SubjectID, &post.Title, &post.ImageURL, &post.ImageDescription, &post.Content, &post.Likes, &post.Dislikes, &post.CommentsCount, &post.AuthorName, &post.AuthorUsername, &post.AuthorImageURL, &post.AuthorImageDescription, &post.CreatedAt)
	return post, discussionReadError("get post", err)
}

func (r *PostgresDiscussionRepository) CreatePost(ctx context.Context, userID int64, input PostInput) (model.Post, error) {
	var postID int64
	err := r.db.QueryRowContext(ctx, `INSERT INTO posts (id_user, id_subject, title, image_url, image_description, content)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id`, userID, input.SubjectID, input.Title, input.ImageURL, input.ImageDescription, input.Content).Scan(&postID)
	if err != nil {
		return model.Post{}, discussionMutationError("create post", err)
	}
	return r.GetPost(ctx, postID)
}

func (r *PostgresDiscussionRepository) UpdatePost(ctx context.Context, postID int64, input PostInput) (model.Post, error) {
	const query = `UPDATE posts SET id_subject = $1, title = $2, image_url = $3, image_description = $4, content = $5
WHERE id = $6
RETURNING id`
	var updatedID int64
	err := r.db.QueryRowContext(ctx, query, input.SubjectID, input.Title, input.ImageURL, input.ImageDescription, input.Content, postID).Scan(&updatedID)
	if err != nil {
		return model.Post{}, discussionMutationReadError("update post", err)
	}
	return r.GetPost(ctx, updatedID)
}

func (r *PostgresDiscussionRepository) DeletePost(ctx context.Context, postID int64) error {
	return deleteDiscussionItem(ctx, r.db, "DELETE FROM posts WHERE id = $1", postID, "delete post")
}

func (r *PostgresDiscussionRepository) GetComment(ctx context.Context, commentID int64) (model.Comment, error) {
	const query = `SELECT c.id, c.id_post, c.id_user, c.content, c.likes, c.dislikes, u.name, u.username, c.created_at
FROM comments c JOIN users u ON u.id = c.id_user WHERE c.id = $1`
	var comment model.Comment
	err := r.db.QueryRowContext(ctx, query, commentID).Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Content, &comment.Likes, &comment.Dislikes, &comment.AuthorName, &comment.AuthorUsername, &comment.CreatedAt)
	return comment, discussionReadError("get comment", err)
}

func (r *PostgresDiscussionRepository) CreateComment(ctx context.Context, userID, postID int64, input CommentInput) (model.Comment, error) {
	var commentID int64
	err := r.db.QueryRowContext(ctx, `SELECT create_message($1, $2, $3, $4, NULL, NULL, NULL)`, "comment", userID, postID, input.Content).Scan(&commentID)
	if err != nil {
		return model.Comment{}, discussionMutationError("create comment", err)
	}
	return r.GetComment(ctx, commentID)
}

func (r *PostgresDiscussionRepository) UpdateComment(ctx context.Context, commentID int64, input CommentInput) (model.Comment, error) {
	const query = `UPDATE comments SET content = $1 WHERE id = $2 RETURNING id`
	var updatedID int64
	err := r.db.QueryRowContext(ctx, query, input.Content, commentID).Scan(&updatedID)
	if err != nil {
		return model.Comment{}, discussionMutationReadError("update comment", err)
	}
	return r.GetComment(ctx, updatedID)
}

func (r *PostgresDiscussionRepository) DeleteComment(ctx context.Context, commentID int64) error {
	return deleteDiscussionItem(ctx, r.db, "DELETE FROM comments WHERE id = $1", commentID, "delete comment")
}

func (r *PostgresDiscussionRepository) GetReply(ctx context.Context, replyID int64) (model.CommentOnComment, error) {
	const query = `SELECT c.id, c.id_comment, c.id_user, c.content, c.likes, c.dislikes, u.name, u.username, c.created_at
FROM comments_on_comments c JOIN users u ON u.id = c.id_user WHERE c.id = $1`
	var reply model.CommentOnComment
	err := r.db.QueryRowContext(ctx, query, replyID).Scan(&reply.ID, &reply.CommentID, &reply.UserID, &reply.Content, &reply.Likes, &reply.Dislikes, &reply.AuthorName, &reply.AuthorUsername, &reply.CreatedAt)
	return reply, discussionReadError("get reply", err)
}

func (r *PostgresDiscussionRepository) CreateReply(ctx context.Context, userID, commentID int64, input CommentInput) (model.CommentOnComment, error) {
	var replyID int64
	err := r.db.QueryRowContext(ctx, `SELECT create_message($1, $2, $3, $4, NULL, NULL, NULL)`, "comment_on_comment", userID, commentID, input.Content).Scan(&replyID)
	if err != nil {
		return model.CommentOnComment{}, discussionMutationError("create reply", err)
	}
	return r.GetReply(ctx, replyID)
}

func (r *PostgresDiscussionRepository) UpdateReply(ctx context.Context, replyID int64, input CommentInput) (model.CommentOnComment, error) {
	const query = `UPDATE comments_on_comments SET content = $1 WHERE id = $2 RETURNING id`
	var updatedID int64
	err := r.db.QueryRowContext(ctx, query, input.Content, replyID).Scan(&updatedID)
	if err != nil {
		return model.CommentOnComment{}, discussionMutationReadError("update reply", err)
	}
	return r.GetReply(ctx, updatedID)
}

func (r *PostgresDiscussionRepository) DeleteReply(ctx context.Context, replyID int64) error {
	return deleteDiscussionItem(ctx, r.db, "DELETE FROM comments_on_comments WHERE id = $1", replyID, "delete reply")
}

func (r *PostgresDiscussionRepository) ListPosts(ctx context.Context, filter PostFilter) ([]model.Post, error) {
	query := `SELECT p.id, p.id_user, p.id_subject, p.title, p.image_url, p.image_description, p.content, p.likes, p.dislikes,
	       (SELECT COUNT(*) FROM comments WHERE id_post = p.id), u.name, u.username, u.profile_image_url, u.profile_image_description, p.created_at
FROM posts p JOIN users u ON u.id = p.id_user`
	args := make([]any, 0, 4)
	switch {
	case filter.SubjectID != nil && filter.AuthorID != nil:
		query += ` WHERE p.id_subject = $1 AND p.id_user = $2`
		args = append(args, *filter.SubjectID, *filter.AuthorID)
	case filter.SubjectID != nil:
		query += ` WHERE p.id_subject = $1`
		args = append(args, *filter.SubjectID)
	case filter.AuthorID != nil:
		query += ` WHERE p.id_user = $1`
		args = append(args, *filter.AuthorID)
	}
	query += fmt.Sprintf(` ORDER BY p.created_at DESC, p.id DESC LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
	args = append(args, filter.Pagination.PageSize, pageOffset(filter.Pagination))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
	}
	defer rows.Close()
	posts := make([]model.Post, 0)
	for rows.Next() {
		var post model.Post
		if err := rows.Scan(&post.ID, &post.UserID, &post.SubjectID, &post.Title, &post.ImageURL, &post.ImageDescription, &post.Content, &post.Likes, &post.Dislikes, &post.CommentsCount, &post.AuthorName, &post.AuthorUsername, &post.AuthorImageURL, &post.AuthorImageDescription, &post.CreatedAt); err != nil {
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
	const query = `SELECT c.id, c.id_post, c.id_user, c.content, c.likes, c.dislikes, u.name, u.username, c.created_at
FROM comments c JOIN users u ON u.id = c.id_user
WHERE c.id_post = $1
ORDER BY c.created_at ASC, c.id ASC
LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, postID, pagination.PageSize, pageOffset(pagination))
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}
	defer rows.Close()
	comments := make([]model.Comment, 0)
	for rows.Next() {
		var comment model.Comment
		if err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Content, &comment.Likes, &comment.Dislikes, &comment.AuthorName, &comment.AuthorUsername, &comment.CreatedAt); err != nil {
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
	const query = `SELECT c.id, c.id_comment, c.id_user, c.content, c.likes, c.dislikes, u.name, u.username, c.created_at
FROM comments_on_comments c JOIN users u ON u.id = c.id_user
WHERE c.id_comment = $1
ORDER BY c.created_at ASC, c.id ASC
LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, commentID, pagination.PageSize, pageOffset(pagination))
	if err != nil {
		return nil, fmt.Errorf("list replies: %w", err)
	}
	defer rows.Close()
	replies := make([]model.CommentOnComment, 0)
	for rows.Next() {
		var reply model.CommentOnComment
		if err := rows.Scan(&reply.ID, &reply.CommentID, &reply.UserID, &reply.Content, &reply.Likes, &reply.Dislikes, &reply.AuthorName, &reply.AuthorUsername, &reply.CreatedAt); err != nil {
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

func discussionReadError(operation string, err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrDiscussionNotFound
	}
	if err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}
	return nil
}

func discussionMutationReadError(operation string, err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrDiscussionNotFound
	}
	return discussionMutationError(operation, err)
}

func deleteDiscussionItem(ctx context.Context, db *sql.DB, query string, id int64, operation string) error {
	result, err := db.ExecContext(ctx, query, id)
	if err != nil {
		return discussionMutationError(operation, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s result: %w", operation, err)
	}
	if affected == 0 {
		return ErrDiscussionNotFound
	}
	return nil
}

func discussionMutationError(operation string, err error) error {
	if err == nil {
		return nil
	}
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "23503", "P0001":
			return ErrDiscussionNotFound
		case "23514":
			return ErrDiscussionInvalid
		}
	}
	return fmt.Errorf("%s: %w", operation, err)
}
