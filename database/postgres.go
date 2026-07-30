package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PoolConfig struct {
	MaxOpenConnections int
	MaxIdleConnections int
	ConnectionMaxLife  time.Duration
	ConnectionMaxIdle  time.Duration
	PingTimeout        time.Duration
}

func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxOpenConnections: 20,
		MaxIdleConnections: 5,
		ConnectionMaxLife:  30 * time.Minute,
		ConnectionMaxIdle:  5 * time.Minute,
		PingTimeout:        5 * time.Second,
	}
}

func Open(ctx context.Context, databaseURL string, cfg PoolConfig) (*sql.DB, error) {
	if databaseURL == "" {
		return nil, errors.New("database URL is required")
	}
	if err := validatePoolConfig(cfg); err != nil {
		return nil, err
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	configurePool(db, cfg)

	pingCtx, cancel := context.WithTimeout(ctx, cfg.PingTimeout)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return db, nil
}

func validatePoolConfig(cfg PoolConfig) error {
	if cfg.MaxOpenConnections <= 0 ||
		cfg.MaxIdleConnections < 0 ||
		cfg.MaxIdleConnections > cfg.MaxOpenConnections ||
		cfg.ConnectionMaxLife <= 0 ||
		cfg.ConnectionMaxIdle <= 0 ||
		cfg.PingTimeout <= 0 {
		return errors.New("invalid database pool configuration")
	}
	return nil
}

func configurePool(db *sql.DB, cfg PoolConfig) {
	db.SetMaxOpenConns(cfg.MaxOpenConnections)
	db.SetMaxIdleConns(cfg.MaxIdleConnections)
	db.SetConnMaxLifetime(cfg.ConnectionMaxLife)
	db.SetConnMaxIdleTime(cfg.ConnectionMaxIdle)
}
