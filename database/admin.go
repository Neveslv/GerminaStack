package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"germinaStack/domain/moderation"
	"germinaStack/model"
)

var ErrAdminUserNotFound = moderation.ErrUserNotFound

type PostgresAdminRepository struct{ db *sql.DB }

type AdminFilter = moderation.Filter
type AdminPost = moderation.Post

func NewPostgresAdminRepository(db *sql.DB) *PostgresAdminRepository {
	return &PostgresAdminRepository{db: db}
}

func (r *PostgresAdminRepository) ListUsers(ctx context.Context, filter AdminFilter) ([]model.User, int, error) {
	term := "%" + filter.Search + "%"
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE name ILIKE $1 OR username ILIKE $1`, term).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, id_year, name, profile_image_url, profile_image_description, username, email, is_admin, is_banned, created_at FROM users WHERE name ILIKE $1 OR username ILIKE $1 ORDER BY name, id LIMIT $2 OFFSET $3`, term, filter.Pagination.PageSize, pageOffset(filter.Pagination))
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()
	users := make([]model.User, 0)
	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.ID, &user.YearID, &user.Name, &user.ProfileImageURL, &user.ProfileImageDescription, &user.Username, &user.Email, &user.IsAdmin, &user.IsBanned, &user.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, user)
	}
	return users, total, rows.Err()
}

func (r *PostgresAdminRepository) ListPosts(ctx context.Context, filter AdminFilter) ([]AdminPost, int, error) {
	term := "%" + filter.Search + "%"
	const where = ` FROM posts p JOIN users u ON u.id = p.id_user WHERE p.title ILIKE $1 OR p.content ILIKE $1 OR u.name ILIKE $1 OR u.username ILIKE $1`
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*)`+where, term).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count posts: %w", err)
	}
	rows, err := r.db.QueryContext(ctx, `SELECT p.id, p.title, u.name, u.username`+where+` ORDER BY p.created_at DESC, p.id DESC LIMIT $2 OFFSET $3`, term, filter.Pagination.PageSize, pageOffset(filter.Pagination))
	if err != nil {
		return nil, 0, fmt.Errorf("list posts: %w", err)
	}
	defer rows.Close()
	posts := make([]AdminPost, 0)
	for rows.Next() {
		var post AdminPost
		if err := rows.Scan(&post.ID, &post.Title, &post.AuthorName, &post.AuthorUsername); err != nil {
			return nil, 0, fmt.Errorf("scan post: %w", err)
		}
		posts = append(posts, post)
	}
	return posts, total, rows.Err()
}

func (r *PostgresAdminRepository) GetUser(ctx context.Context, id int64) (model.User, error) {
	var user model.User
	err := r.db.QueryRowContext(ctx, `SELECT id, id_year, name, profile_image_url, profile_image_description, username, email, is_admin, is_banned, created_at FROM users WHERE id = $1`, id).Scan(&user.ID, &user.YearID, &user.Name, &user.ProfileImageURL, &user.ProfileImageDescription, &user.Username, &user.Email, &user.IsAdmin, &user.IsBanned, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return model.User{}, ErrAdminUserNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}

func (r *PostgresAdminRepository) SetBanned(ctx context.Context, id int64, banned bool) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin ban: %w", err)
	}
	defer tx.Rollback()
	if banned {
		if _, err := tx.ExecContext(ctx, `DELETE FROM posts WHERE id_user = $1`, id); err != nil {
			return fmt.Errorf("delete banned user posts: %w", err)
		}
	}
	result, err := tx.ExecContext(ctx, `UPDATE users SET is_banned = $1 WHERE id = $2`, banned, id)
	if err != nil {
		return fmt.Errorf("set ban: %w", err)
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return ErrAdminUserNotFound
	}
	return tx.Commit()
}

func (r *PostgresAdminRepository) SetAdmin(ctx context.Context, id int64, admin bool) error {
	result, err := r.db.ExecContext(ctx, `UPDATE users SET is_admin = $1 WHERE id = $2`, admin, id)
	if err != nil {
		return fmt.Errorf("set admin: %w", err)
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return ErrAdminUserNotFound
	}
	return nil
}

func (r *PostgresAdminRepository) DeletePost(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM posts WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete post: %w", err)
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return ErrAdminUserNotFound
	}
	return nil
}
