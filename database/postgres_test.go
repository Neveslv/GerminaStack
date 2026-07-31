package database

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestOpenRejectsEmptyDatabaseURL(t *testing.T) {
	t.Parallel()

	db, err := Open(context.Background(), "", DefaultPoolConfig())
	if err == nil {
		if db != nil {
			db.Close()
		}
		t.Fatal("Open() error = nil, want non-nil")
	}
}

func TestConfigurePoolAppliesLimits(t *testing.T) {
	t.Parallel()

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	cfg := PoolConfig{
		MaxOpenConnections: 17,
		MaxIdleConnections: 7,
		ConnectionMaxLife:  30 * time.Minute,
		ConnectionMaxIdle:  5 * time.Minute,
		PingTimeout:        2 * time.Second,
	}
	configurePool(db, cfg)

	if got := db.Stats().MaxOpenConnections; got != 17 {
		t.Fatalf("MaxOpenConnections = %d, want 17", got)
	}
}

func TestOpenStopsWhenPingContextIsCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	db, err := Open(ctx, "postgres://user:password@127.0.0.1:1/app?sslmode=disable", DefaultPoolConfig())
	if err == nil {
		if db != nil {
			db.Close()
		}
		t.Fatal("Open() error = nil, want canceled ping error")
	}
}
