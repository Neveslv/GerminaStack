package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"germinaStack/model"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrMessageNotFound       = errors.New("message not found")
	ErrMessageParentNotFound = errors.New("message parent not found")
)

type MessageRepository interface {
	CreatePost(context.Context, int64, int64, string, *string, *string, string) (model.Post, error)
	CreateComment(context.Context, int64, int64, string) (model.Comment, error)
	CreateReply(context.Context, int64, int64, string) (model.CommentOnComment, error)
	GetPost(context.Context, int64) (model.Post, error)
	ListPosts(context.Context, *int64) ([]model.Post, error)
	GetComment(context.Context, int64) (model.Comment, error)
	ListComments(context.Context, int64) ([]model.Comment, error)
	GetReply(context.Context, int64) (model.CommentOnComment, error)
	ListReplies(context.Context, int64) ([]model.CommentOnComment, error)
}

type PostgresMessageRepository struct {
	db *sql.DB
}

func NewPostgresMessageRepository(db *sql.DB) *PostgresMessageRepository {
	return &PostgresMessageRepository{db: db}
}

func (r *PostgresMessageRepository) CreatePost(ctx context.Context, userID, subjectID int64, content string, imageURL, imageDescription *string, title string) (model.Post, error) {
	id, err := r.createMessage(ctx, "post", userID, subjectID, content, &title, imageURL, imageDescription)
	if err != nil {
		return model.Post{}, err
	}
	return r.GetPost(ctx, id)
}

func (r *PostgresMessageRepository) CreateComment(ctx context.Context, userID, postID int64, content string) (model.Comment, error) {
	id, err := r.createMessage(ctx, "comment", userID, postID, content, nil, nil, nil)
	if err != nil {
		return model.Comment{}, err
	}
	return r.GetComment(ctx, id)
}

func (r *PostgresMessageRepository) CreateReply(ctx context.Context, userID, commentID int64, content string) (model.CommentOnComment, error) {
	id, err := r.createMessage(ctx, "comment_on_comment", userID, commentID, content, nil, nil, nil)
	if err != nil {
		return model.CommentOnComment{}, err
	}
	return r.GetReply(ctx, id)
}

func (r *PostgresMessageRepository) createMessage(ctx context.Context, messageType string, userID, parentID int64, content string, title any, imageURL, imageDescription *string) (int64, error) {
	const query = `SELECT create_message($1, $2, $3, $4, $5, $6, $7)`

	var id int64
	err := r.db.QueryRowContext(ctx, query, messageType, userID, parentID, content, title, imageURL, imageDescription).Scan(&id)
	if err == nil {
		return id, nil
	}
	if isForeignKeyViolation(err) {
		return 0, ErrMessageParentNotFound
	}
	return 0, fmt.Errorf("create %s: %w", messageType, err)
}

func (r *PostgresMessageRepository) GetPost(ctx context.Context, id int64) (model.Post, error) {
	const query = `SELECT id, id_user, id_subject, title, image_url, image_description, content, likes, dislikes, created_at FROM posts WHERE id = $1`

	var post model.Post
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.UserID,
		&post.SubjectID,
		&post.Title,
		&post.ImageURL,
		&post.ImageDescription,
		&post.Content,
		&post.Likes,
		&post.Dislikes,
		&post.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Post{}, ErrMessageNotFound
	}
	if err != nil {
		return model.Post{}, fmt.Errorf("get post: %w", err)
	}
	return post, nil
}

func (r *PostgresMessageRepository) ListPosts(ctx context.Context, subjectID *int64) ([]model.Post, error) {
	query := `SELECT id, id_user, id_subject, title, image_url, image_description, content, likes, dislikes, created_at FROM posts ORDER BY created_at DESC, id DESC`
	args := []any{}
	if subjectID != nil {
		query = `SELECT id, id_user, id_subject, title, image_url, image_description, content, likes, dislikes, created_at FROM posts WHERE id_subject = $1 ORDER BY created_at DESC, id DESC`
		args = append(args, *subjectID)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
	}
	defer rows.Close()

	posts := make([]model.Post, 0)
	for rows.Next() {
		var post model.Post
		if err := scanPost(rows, &post); err != nil {
			return nil, fmt.Errorf("scan post: %w", err)
		}
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate posts: %w", err)
	}
	return posts, nil
}

func (r *PostgresMessageRepository) GetComment(ctx context.Context, id int64) (model.Comment, error) {
	const query = `SELECT id, id_post, id_user, content, likes, dislikes, created_at FROM comments WHERE id = $1`

	var comment model.Comment
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&comment.ID,
		&comment.PostID,
		&comment.UserID,
		&comment.Content,
		&comment.Likes,
		&comment.Dislikes,
		&comment.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Comment{}, ErrMessageNotFound
	}
	if err != nil {
		return model.Comment{}, fmt.Errorf("get comment: %w", err)
	}
	return comment, nil
}

func (r *PostgresMessageRepository) ListComments(ctx context.Context, postID int64) ([]model.Comment, error) {
	const query = `SELECT id, id_post, id_user, content, likes, dislikes, created_at FROM comments WHERE id_post = $1 ORDER BY created_at DESC, id DESC`
	return queryComments(ctx, r.db, query, postID)
}

func (r *PostgresMessageRepository) GetReply(ctx context.Context, id int64) (model.CommentOnComment, error) {
	const query = `SELECT id, id_comment, id_user, content, likes, dislikes, created_at FROM comments_on_comments WHERE id = $1`

	var reply model.CommentOnComment
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&reply.ID,
		&reply.CommentID,
		&reply.UserID,
		&reply.Content,
		&reply.Likes,
		&reply.Dislikes,
		&reply.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.CommentOnComment{}, ErrMessageNotFound
	}
	if err != nil {
		return model.CommentOnComment{}, fmt.Errorf("get reply: %w", err)
	}
	return reply, nil
}

func (r *PostgresMessageRepository) ListReplies(ctx context.Context, commentID int64) ([]model.CommentOnComment, error) {
	const query = `SELECT id, id_comment, id_user, content, likes, dislikes, created_at FROM comments_on_comments WHERE id_comment = $1 ORDER BY created_at DESC, id DESC`

	rows, err := r.db.QueryContext(ctx, query, commentID)
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

type rowScanner interface {
	Scan(...any) error
}

func scanPost(row rowScanner, post *model.Post) error {
	return row.Scan(
		&post.ID,
		&post.UserID,
		&post.SubjectID,
		&post.Title,
		&post.ImageURL,
		&post.ImageDescription,
		&post.Content,
		&post.Likes,
		&post.Dislikes,
		&post.CreatedAt,
	)
}

func queryComments(ctx context.Context, db *sql.DB, query string, parentID int64) ([]model.Comment, error) {
	rows, err := db.QueryContext(ctx, query, parentID)
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

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
