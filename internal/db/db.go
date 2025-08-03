package db

import (
	"context"
	"fmt"

	sqlc "ariand/internal/db/sqlc"

	"github.com/charmbracelet/log"
	"github.com/jackc/pgx/v5/pgxpool"
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
