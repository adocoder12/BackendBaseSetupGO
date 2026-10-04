package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/adocoder12/BackendBaseSetupGO/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

const migrationsPath = "file://internal/db/migrations"

// NewPool creates a pgxpool with explicit connection limits from config.
// The caller is responsible for calling pool.Close() on shutdown.
func NewPool(ctx context.Context, cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("db - parse config: %w", err)
	}
	poolConfig.MaxConns = cfg.MaxConns              // Absolute ceiling on active connections
	poolConfig.MinConns = cfg.MinConns              // Keep cold connections alive to eliminate cold starts
	poolConfig.MaxConnLifetime = 30 * time.Second   // Prevent memory leaks / stale connections
	poolConfig.MaxConnIdleTime = 5 * time.Minute    // Terminate idle connections to free DB memory
	poolConfig.HealthCheckPeriod = 30 * time.Second // Automatically prune dropped/broken sockets

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("db - new pool: %w", err)
	}

	// Ping makes the real connection, so we know the DB is reachable.
	if err := pool.Ping(ctx); err != nil {
		pool.Close() // don't leak the pool on failure
		return nil, fmt.Errorf("db - ping: %w", err)
	}

	return pool, nil
}

// Migrate runs all pending up migrations.
// Uses MigrationDSN() (postgres:// URL), which golang-migrate requires.
// Safe to call on every startup: applied migrations are skipped.
func Migrate(cfg config.DatabaseConfig) error {
	m, err := migrate.New(migrationsPath, cfg.MigrationDSN())
	if err != nil {
		return fmt.Errorf("db - migrate new: %w", err)
	}
	defer func() {
		_, _ = m.Close()
	}()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("db - migrate up: %w", err)
	}

	return nil
}
