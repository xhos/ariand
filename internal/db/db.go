package db

import (
	"context"
	"database/sql"
	"fmt"

	sqlc "ariand/internal/db/sqlc"

	"github.com/charmbracelet/log"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

type DB struct {
	*sqlc.Queries
	log  *log.Logger
	pool *pgxpool.Pool
}

func New(dsn string) (*DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("empty dsn")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.new: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return &DB{
		Queries: sqlc.New(pool),
		log:     log.WithPrefix("db"),
		pool:    pool,
	}, nil
}

func (s *DB) Close() error {
	s.pool.Close()
	return nil
}

func (s *DB) Pool() *pgxpool.Pool {
	return s.pool
}

func RunMigrations(dsn string, migrationsDir string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database for migrations: %w", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
