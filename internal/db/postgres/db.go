package postgres

import (
	"ariand/internal/db"
	"ariand/internal/db/postgres/queries"
	_ "embed"
	"fmt"

	"github.com/charmbracelet/log"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

// DB holds the database connection and embeds all query types
type DB struct {
	*sqlx.DB
	log *log.Logger

	*queries.Accounts
	*queries.Dashboard
	*queries.Transactions
	*queries.Categories
}

//go:embed schema.sql
var schemaDDL string

// statically assert that *DB satisfies the db.Store interface
// this will cause a compile-time error if the interface is not fully implemented
var _ db.Store = (*DB)(nil)

func ensureSchema(db *sqlx.DB) error {
	_, err := db.Exec(string(schemaDDL))
	return err
}

// New creates a new DB connection
func New(dsn string) (*DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("empty DSN")
	}

	conn, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlx.Open: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, err
	}
	if err := ensureSchema(conn); err != nil {
		return nil, err
	}

	return &DB{
		DB:           conn,
		log:          log.WithPrefix("db"),
		Accounts:     queries.NewAccounts(conn),
		Dashboard:    queries.NewDashboard(conn),
		Transactions: queries.NewTransactions(conn),
		Categories:   queries.NewCategories(conn),
	}, nil
}

// NewFromDB initializes a DB instance from an existing sqlx.DB connection.
// It assumes the connection is alive and applies the schema.
func NewFromDB(conn *sqlx.DB) (*DB, error) {
	if err := ensureSchema(conn); err != nil {
		return nil, fmt.Errorf("failed to ensure schema: %w", err)
	}

	return &DB{
		DB:           conn,
		log:          log.WithPrefix("db"),
		Accounts:     queries.NewAccounts(conn),
		Dashboard:    queries.NewDashboard(conn),
		Transactions: queries.NewTransactions(conn),
		Categories:   queries.NewCategories(conn),
	}, nil
}

func (db *DB) Close() error { return db.DB.Close() }
