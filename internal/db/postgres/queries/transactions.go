package queries

import (
	"ariand/internal/db"
	"ariand/internal/domain"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

type Transactions struct {
	db *sqlx.DB
}

func NewTransactions(db *sqlx.DB) *Transactions {
	return &Transactions{db}
}

// SyncAccountBalances recalculates the entire balance_after chain for a given account
func (q *Transactions) SyncAccountBalances(ctx context.Context, tx *sqlx.Tx, accountID int64) error {
	const syncQuery = `
 		WITH
 		-- Step 1: Calculate the running delta for every transaction from the beginning of time.
 		transaction_deltas AS (
 			SELECT
 				id,
 				SUM(CASE WHEN tx_direction = 'in' THEN tx_amount ELSE -tx_amount END)
 					OVER (PARTITION BY account_id ORDER BY tx_date, id) as running_delta
 			FROM transactions
 			WHERE account_id = $1
 		),
 		-- Step 2: Calculate the state of the account at the anchor point.
 		-- This includes the anchor balance itself and the sum of all transaction
 		-- deltas that occurred *before* the anchor date.
 		anchor_point AS (
 			SELECT
 				a.anchor_balance,
 				COALESCE(SUM(CASE WHEN t.tx_direction = 'in' THEN t.tx_amount ELSE -t.tx_amount END), 0.0) as delta_at_anchor
 			FROM
 				accounts a
 			LEFT JOIN
 				-- Join only those transactions that happened strictly before the anchor date
 				transactions t ON t.account_id = a.id AND t.tx_date < a.anchor_date
 			WHERE
 				a.id = $1
 			GROUP BY
 				a.id, a.anchor_balance
 		)
 		-- Step 3: Update every transaction for the account.
 		UPDATE
 			transactions
 		SET
 			-- The new balance is calculated by taking this transaction's running delta,
 			-- subtracting the delta at the anchor point, and adding the anchor balance.
 			-- This effectively "rebases" the entire history around the anchor point.
 			balance_after = ap.anchor_balance + td.running_delta - ap.delta_at_anchor
 		FROM
 			transaction_deltas td,
 			anchor_point ap
 		WHERE
 			transactions.id = td.id
 			AND transactions.account_id = $1;
 	`
	_, err := tx.ExecContext(ctx, syncQuery, accountID)
	return err
}

// ListTransactions constructs a dynamic query for fetching transactions
func (q *Transactions) ListTransactions(ctx context.Context, opts db.ListOpts) ([]domain.Transaction, error) {
	var args []any
	var conditions []string

	// --- build filter conditions ---
	if opts.CursorDate != nil && opts.CursorID != nil {
		conditions = append(conditions, fmt.Sprintf("(tx_date, id) < ($%d, $%d)", len(args)+1, len(args)+2))
		args = append(args, *opts.CursorDate, *opts.CursorID)
	}
	if opts.Start != nil {
		conditions = append(conditions, fmt.Sprintf("tx_date >= $%d", len(args)+1))
		args = append(args, *opts.Start)
	}
	if opts.End != nil {
		conditions = append(conditions, fmt.Sprintf("tx_date <= $%d", len(args)+1))
		args = append(args, *opts.End)
	}
	if opts.AmountMin != nil {
		conditions = append(conditions, fmt.Sprintf("tx_amount >= $%d", len(args)+1))
		args = append(args, *opts.AmountMin)
	}
	if opts.AmountMax != nil {
		conditions = append(conditions, fmt.Sprintf("tx_amount <= $%d", len(args)+1))
		args = append(args, *opts.AmountMax)
	}
	if opts.Direction != "" {
		conditions = append(conditions, fmt.Sprintf("tx_direction = $%d", len(args)+1))
		args = append(args, opts.Direction)
	}
	if len(opts.Categories) > 0 {
		conditions = append(conditions, fmt.Sprintf("category = ANY($%d)", len(args)+1))
		args = append(args, opts.Categories)
	}
	if opts.MerchantSearch != nil {
		conditions = append(conditions, fmt.Sprintf("merchant ILIKE $%d", len(args)+1))
		args = append(args, "%"+*opts.MerchantSearch+"%")
	}
	if opts.DescriptionSearch != nil {
		conditions = append(conditions, fmt.Sprintf("tx_desc ILIKE $%d", len(args)+1))
		args = append(args, "%"+*opts.DescriptionSearch+"%")
	}
	if opts.Currency != nil {
		conditions = append(conditions, fmt.Sprintf("tx_currency = $%d", len(args)+1))
		args = append(args, *opts.Currency)
	}
	if opts.TimeOfDayStart != nil {
		conditions = append(conditions, fmt.Sprintf("tx_date::time >= $%d", len(args)+1))
		args = append(args, *opts.TimeOfDayStart)
	}
	if opts.TimeOfDayEnd != nil {
		conditions = append(conditions, fmt.Sprintf("tx_date::time <= $%d", len(args)+1))
		args = append(args, *opts.TimeOfDayEnd)
	}
	if len(opts.AccountIDs) > 0 {
		conditions = append(conditions, fmt.Sprintf("account_id = ANY($%d)", len(args)+1))
		args = append(args, opts.AccountIDs)
	}

	// --- construct the final query ---
	query := "SELECT * FROM transactions"
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// ORDER BY clause must be fixed and match the cursor logic
	query += " ORDER BY tx_date DESC, id DESC"

	if opts.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", opts.Limit)
	}

	var out []domain.Transaction
	if err := q.db.SelectContext(ctx, &out, query, args...); err != nil {
		return nil, err
	}

	return out, nil
}

func (q *Transactions) GetTransaction(ctx context.Context, id int64) (*domain.Transaction, error) {
	var transaction domain.Transaction

	err := q.db.GetContext(ctx, &transaction, `SELECT * FROM transactions WHERE id=$1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, db.ErrNotFound
		}
		return nil, err
	}

	return &transaction, nil
}

func (q *Transactions) CreateTransaction(ctx context.Context, t *domain.Transaction) (int64, error) {
	tx, err := q.db.Beginx()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var anchorBalance float64
	err = tx.GetContext(ctx, &anchorBalance, `SELECT anchor_balance FROM accounts WHERE id=$1 FOR UPDATE`, t.AccountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("account with id %d not found", t.AccountID)
		}
		return 0, fmt.Errorf("could not lock account for update: %w", err)
	}

	var previousBalance float64
	err = tx.GetContext(ctx, &previousBalance, `
 		SELECT balance_after FROM transactions WHERE account_id = $1 ORDER BY tx_date DESC, id DESC LIMIT 1
 	`, t.AccountID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			previousBalance = anchorBalance
		} else {
			return 0, fmt.Errorf("could not get previous balance: %w", err)
		}
	}

	if t.TxDirection == "in" {
		t.BalanceAfter = previousBalance + t.TxAmount
	} else {
		t.BalanceAfter = previousBalance - t.TxAmount
	}

	const insertQuery = `
 	INSERT INTO transactions (
 	  email_id, account_id, tx_date, tx_amount, tx_currency, tx_direction,
 	  tx_desc, balance_after, category, merchant, user_notes,
 	  foreign_currency, foreign_amount, exchange_rate
 	) VALUES (
 	  :email_id, :account_id, :tx_date, :tx_amount, :tx_currency, :tx_direction,
 	  :tx_desc, :balance_after, :category, :merchant, :user_notes,
 	  :foreign_currency, :foreign_amount, :exchange_rate
 	) RETURNING id`

	stmt, err := tx.PrepareNamedContext(ctx, insertQuery)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	var id int64
	err = stmt.GetContext(ctx, &id, t)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, db.ErrConflict
		}
		return 0, fmt.Errorf("failed to execute insert statement: %w", err)
	}

	if err := q.SyncAccountBalances(ctx, tx, t.AccountID); err != nil {
		return 0, fmt.Errorf("failed to sync balances: %w", err)
	}

	return id, tx.Commit()
}

func (q *Transactions) UpdateTransaction(ctx context.Context, id int64, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}

	allowedCols := map[string]bool{
		"tx_date": true, "tx_amount": true, "tx_currency": true,
		"tx_direction": true, "tx_desc": true, "category": true,
		"merchant": true, "user_notes": true, "foreign_currency": true,
		"foreign_amount": true, "exchange_rate": true,
	}

	var accountID int64
	tx, err := q.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	err = tx.GetContext(ctx, &accountID, `SELECT account_id FROM transactions WHERE id=$1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.ErrNotFound
		}
		return err
	}

	setClauses, args, err := buildUpdateClauses(fields, allowedCols)
	if err != nil {
		return err
	}
	args = append(args, id)

	query := fmt.Sprintf("UPDATE transactions SET %s WHERE id = $%d", setClauses, len(args))

	res, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return db.ErrNotFound
	}

	if err := q.SyncAccountBalances(ctx, tx, accountID); err != nil {
		return fmt.Errorf("failed to sync balances: %w", err)
	}

	return tx.Commit()
}

func (q *Transactions) DeleteTransaction(ctx context.Context, id int64) error {
	var accountID int64
	tx, err := q.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	err = tx.GetContext(ctx, &accountID, `SELECT account_id FROM transactions WHERE id=$1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.ErrNotFound
		}
		return err
	}

	res, err := tx.ExecContext(ctx, `DELETE FROM transactions WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return db.ErrNotFound
	}

	if err := q.SyncAccountBalances(ctx, tx, accountID); err != nil {
		return fmt.Errorf("failed to sync balances: %w", err)
	}

	return tx.Commit()
}

func buildUpdateClauses(fields map[string]any, allowedCols map[string]bool) (string, []any, error) {
	var setClauses []string
	var args []any

	i := 1

	for k, v := range fields {
		if !allowedCols[k] {
			continue
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", k, i))
		args = append(args, v)
		i++
	}

	if len(setClauses) == 0 {
		return "", nil, errors.New("no valid fields provided for update")
	}

	return strings.Join(setClauses, ", "), args, nil
}
