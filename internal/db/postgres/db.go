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

//go:embed schema.sql
var schemaDDL string

func ensureSchema(db *sqlx.DB) error {
	_, err := db.Exec(schemaDDL)
	return err
}

type DB struct {
	*sqlx.DB
	log *log.Logger

	*queries.Accounts
	*queries.Dashboard
	*queries.Transactions
	*queries.Categories
	*queries.Receipts
}

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

	return initStore(conn), nil
}

func NewFromDB(conn *sqlx.DB) (*DB, error) {
	if err := ensureSchema(conn); err != nil {
		return nil, err
	}

	return initStore(conn), nil
}

func initStore(conn *sqlx.DB) *DB {
	return &DB{
		DB:           conn,
		log:          log.WithPrefix("db"),
		Accounts:     queries.NewAccounts(conn),
		Dashboard:    queries.NewDashboard(conn),
		Transactions: queries.NewTransactions(conn),
		Categories:   queries.NewCategories(conn),
		Receipts:     queries.NewReceipts(conn), // ← NEW
	}
}

func (db *DB) Close() error { return db.DB.Close() }

var _ db.Store = (*DB)(nil)
