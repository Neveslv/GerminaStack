package database

import (
	"context"
	"database/sql"
	"fmt"

	"germinaStack/model"
)

type PostgresFrokRepository struct {
	db *sql.DB
}

func NewPostgresFrokRepository(db *sql.DB) *PostgresFrokRepository {
	return &PostgresFrokRepository{db: db}
}

func (r *PostgresFrokRepository) BotUserID(ctx context.Context) (int64, error) {
	var userID int64
	err := r.db.QueryRowContext(ctx, `SELECT id FROM users WHERE username = 'frok'`).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("get Frok user: %w", err)
	}
	return userID, nil
}

func (r *PostgresFrokRepository) Username(ctx context.Context, userID int64) (string, error) {
	var username string
	err := r.db.QueryRowContext(ctx, `SELECT username FROM users WHERE id = $1`, userID).Scan(&username)
	if err != nil {
		return "", fmt.Errorf("get Frok recipient: %w", err)
	}
	return username, nil
}

func (r *PostgresFrokRepository) GetPost(ctx context.Context, postID int64) (model.Post, error) {
	return NewPostgresDiscussionRepository(r.db).GetPost(ctx, postID)
}

func (r *PostgresFrokRepository) GetComment(ctx context.Context, commentID int64) (model.Comment, error) {
	return NewPostgresDiscussionRepository(r.db).GetComment(ctx, commentID)
}

func (r *PostgresFrokRepository) ListReplies(ctx context.Context, commentID int64) ([]model.CommentOnComment, error) {
	// ponytail: 50 replies keep one Groq prompt bounded; add token budgeting if threads grow beyond this.
	return NewPostgresDiscussionRepository(r.db).ListReplies(ctx, commentID, Pagination{Page: 1, PageSize: 50})
}

func (r *PostgresFrokRepository) CreateComment(ctx context.Context, userID, postID int64, content string) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO comments (id_user, id_post, content) VALUES ($1, $2, $3)`, userID, postID, content)
	if err != nil {
		return fmt.Errorf("create Frok comment: %w", err)
	}
	return nil
}

func (r *PostgresFrokRepository) CreateReply(ctx context.Context, userID, commentID int64, content string) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO comments_on_comments (id_user, id_comment, content) VALUES ($1, $2, $3)`, userID, commentID, content)
	if err != nil {
		return fmt.Errorf("create Frok reply: %w", err)
	}
	return nil
}
