package queries

import (
	"ariand/internal/db"
	"ariand/internal/domain"
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

type Accounts struct {
	db *sqlx.DB
}

func NewAccounts(db *sqlx.DB) *Accounts {
	return &Accounts{db}
}

func (q *Accounts) CreateAccount(ctx context.Context, acc *domain.Account) (*domain.Account, error) {
	query := `
		INSERT INTO accounts (name, bank, account_type, anchor_date, anchor_balance, alias)
		VALUES (:name, :bank, :account_type, :anchor_date, :anchor_balance, :alias)
		RETURNING *
	`
	var createdAccount domain.Account
	stmt, err := q.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	err = stmt.GetContext(ctx, &createdAccount, acc)
	if err != nil {
		return nil, err
	}

	return &createdAccount, nil
}

func (q *Accounts) ListAccounts(ctx context.Context) ([]domain.Account, error) {
	var accounts []domain.Account
	err := q.db.SelectContext(ctx, &accounts, `SELECT * FROM accounts ORDER BY created_at`)
	return accounts, err
}

func (q *Accounts) GetAccount(ctx context.Context, id int64) (*domain.Account, error) {
	var account domain.Account

	err := q.db.GetContext(ctx, &account, `SELECT * FROM accounts WHERE id=$1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, db.ErrNotFound
		}
		return nil, err
	}

	return &account, err
}

func (q *Accounts) DeleteAccount(ctx context.Context, id int64) error {
	tx, err := q.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// first, delete associated transactions to avoid foreign key violations
	_, err = tx.ExecContext(ctx, `DELETE FROM transactions WHERE account_id = $1`, id)
	if err != nil {
		return err
	}

	// now, delete the account itself
	res, err := tx.ExecContext(ctx, `DELETE FROM accounts WHERE id = $1`, id)
	if err != nil {
		return err
	}

	// check if a row was actually deleted
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return db.ErrNotFound
	}

	return tx.Commit()
}

func (q *Accounts) SetAccountAnchor(ctx context.Context, id int64, balance float64) error {
	tx, err := q.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`UPDATE accounts SET anchor_date=NOW(), anchor_balance=$1 WHERE id=$2`,
		balance, id)

	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return db.ErrNotFound
	}

	tq := NewTransactions(q.db)
	if err := tq.SyncAccountBalances(ctx, tx, id); err != nil {
		return err
	}

	return tx.Commit()
}

func (q *Accounts) GetAccountBalance(ctx context.Context, id int64) (float64, error) {
	var currentBalance float64
	err := q.db.GetContext(ctx, &currentBalance, `
    SELECT balance_after FROM transactions WHERE account_id=$1 ORDER BY tx_date DESC, id DESC LIMIT 1
  `, id)

	if err == nil {
		return currentBalance, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	err = q.db.GetContext(ctx, &currentBalance, `SELECT anchor_balance FROM accounts WHERE id=$1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, db.ErrNotFound
		}
		return 0, err
	}

	return currentBalance, nil
}
