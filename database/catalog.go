package database

import (
	"context"
	"database/sql"
	"fmt"

	"germinaStack/model"
)

const academicYear = "2"

type PostgresCatalogRepository struct {
	db *sql.DB
}

func NewPostgresCatalogRepository(db *sql.DB) *PostgresCatalogRepository {
	return &PostgresCatalogRepository{db: db}
}

func (r *PostgresCatalogRepository) ListYears(ctx context.Context) ([]model.Year, error) {
	const query = `SELECT id, year, created_at FROM years WHERE year = $1 ORDER BY id`
	rows, err := r.db.QueryContext(ctx, query, academicYear)
	if err != nil {
		return nil, fmt.Errorf("list years: %w", err)
	}
	defer rows.Close()

	years := make([]model.Year, 0)
	for rows.Next() {
		var year model.Year
		if err := rows.Scan(&year.ID, &year.Year, &year.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan year: %w", err)
		}
		years = append(years, year)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate years: %w", err)
	}
	return years, nil
}

func (r *PostgresCatalogRepository) ListSubjects(ctx context.Context, userID int64) ([]model.Subject, error) {
	const query = `SELECT s.id, s.id_year, s.subject, s.created_at, COUNT(p.id) AS posts_count
FROM subjects s LEFT JOIN posts p ON p.id_subject = s.id
WHERE s.id_year = (SELECT id_year FROM users WHERE id = $1) OR (s.id_year IS NULL AND s.subject = 'Geral')
GROUP BY s.id, s.id_year, s.subject, s.created_at ORDER BY s.subject, s.id`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list subjects: %w", err)
	}
	defer rows.Close()

	subjects := make([]model.Subject, 0)
	for rows.Next() {
		var subject model.Subject
		if err := rows.Scan(&subject.ID, &subject.YearID, &subject.Subject, &subject.CreatedAt, &subject.PostsCount); err != nil {
			return nil, fmt.Errorf("scan subject: %w", err)
		}
		subjects = append(subjects, subject)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate subjects: %w", err)
	}
	return subjects, nil
}
