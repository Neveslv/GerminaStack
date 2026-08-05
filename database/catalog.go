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
	ErrSubjectNotFound   = errors.New("subject not found")
	ErrCatalogConflict   = errors.New("catalog label already exists")
	ErrCatalogReferenced = errors.New("catalog item is referenced")
)

type PostgresCatalogRepository struct {
	db *sql.DB
}

func NewPostgresCatalogRepository(db *sql.DB) *PostgresCatalogRepository {
	return &PostgresCatalogRepository{db: db}
}

func (r *PostgresCatalogRepository) ListYears(ctx context.Context) ([]model.Year, error) {
	const query = `SELECT id, year, created_at FROM years ORDER BY year, id`
	rows, err := r.db.QueryContext(ctx, query)
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

func (r *PostgresCatalogRepository) CreateYear(ctx context.Context, label string) (model.Year, error) {
	const query = `INSERT INTO years (year) VALUES ($1) RETURNING id, year, created_at`
	var year model.Year
	err := r.db.QueryRowContext(ctx, query, label).Scan(&year.ID, &year.Year, &year.CreatedAt)
	if err != nil {
		return model.Year{}, catalogMutationError("create year", err, ErrYearNotFound, false)
	}
	return year, nil
}

func (r *PostgresCatalogRepository) UpdateYear(ctx context.Context, id int64, label string) (model.Year, error) {
	const query = `UPDATE years SET year = $1 WHERE id = $2 RETURNING id, year, created_at`
	var year model.Year
	err := r.db.QueryRowContext(ctx, query, label, id).Scan(&year.ID, &year.Year, &year.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Year{}, ErrYearNotFound
	}
	if err != nil {
		return model.Year{}, catalogMutationError("update year", err, ErrYearNotFound, false)
	}
	return year, nil
}

func (r *PostgresCatalogRepository) DeleteYear(ctx context.Context, id int64) error {
	return deleteCatalogItem(ctx, r.db, `DELETE FROM years WHERE id = $1`, id, ErrYearNotFound, "delete year")
}

func (r *PostgresCatalogRepository) ListSubjects(ctx context.Context, yearID *int64) ([]model.Subject, error) {
	const allQuery = `SELECT id, id_year, subject, created_at FROM subjects ORDER BY subject, id`
	const filteredQuery = `SELECT id, id_year, subject, created_at FROM subjects WHERE id_year = $1 ORDER BY subject, id`
	var rows *sql.Rows
	var err error
	if yearID == nil {
		rows, err = r.db.QueryContext(ctx, allQuery)
	} else {
		rows, err = r.db.QueryContext(ctx, filteredQuery, *yearID)
	}
	if err != nil {
		return nil, fmt.Errorf("list subjects: %w", err)
	}
	defer rows.Close()
	subjects := make([]model.Subject, 0)
	for rows.Next() {
		var subject model.Subject
		if err := rows.Scan(&subject.ID, &subject.YearID, &subject.Subject, &subject.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan subject: %w", err)
		}
		subjects = append(subjects, subject)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate subjects: %w", err)
	}
	return subjects, nil
}

func (r *PostgresCatalogRepository) CreateSubject(ctx context.Context, label string, yearID *int64) (model.Subject, error) {
	const query = `INSERT INTO subjects (id_year, subject) VALUES ($1, $2) RETURNING id, id_year, subject, created_at`
	var subject model.Subject
	err := r.db.QueryRowContext(ctx, query, yearID, label).Scan(&subject.ID, &subject.YearID, &subject.Subject, &subject.CreatedAt)
	if err != nil {
		return model.Subject{}, catalogMutationError("create subject", err, ErrYearNotFound, false)
	}
	return subject, nil
}

func (r *PostgresCatalogRepository) UpdateSubject(ctx context.Context, id int64, label string, yearID *int64) (model.Subject, error) {
	const query = `UPDATE subjects SET id_year = $1, subject = $2 WHERE id = $3 RETURNING id, id_year, subject, created_at`
	var subject model.Subject
	err := r.db.QueryRowContext(ctx, query, yearID, label, id).Scan(&subject.ID, &subject.YearID, &subject.Subject, &subject.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Subject{}, ErrSubjectNotFound
	}
	if err != nil {
		return model.Subject{}, catalogMutationError("update subject", err, ErrYearNotFound, false)
	}
	return subject, nil
}

func (r *PostgresCatalogRepository) DeleteSubject(ctx context.Context, id int64) error {
	return deleteCatalogItem(ctx, r.db, `DELETE FROM subjects WHERE id = $1`, id, ErrSubjectNotFound, "delete subject")
}

func deleteCatalogItem(ctx context.Context, db *sql.DB, query string, id int64, missing error, operation string) error {
	result, err := db.ExecContext(ctx, query, id)
	if err != nil {
		return catalogMutationError(operation, err, missing, true)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s result: %w", operation, err)
	}
	if affected == 0 {
		return missing
	}
	return nil
}

func catalogMutationError(operation string, err, relatedMissing error, deleting bool) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "23505":
			return ErrCatalogConflict
		case "23503":
			if deleting {
				return ErrCatalogReferenced
			}
			return relatedMissing
		}
	}
	return fmt.Errorf("%s: %w", operation, err)
}
