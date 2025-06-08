package queries

import (
	"ariand/internal/domain"
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type Accounts struct {
	db *sqlx.DB
}

func NewAccounts(db *sqlx.DB) *Accounts {
	return &Accounts{db}
}

func (q *Accounts) ListAccounts(ctx context.Context) ([]domain.Account, error) {
	var accounts []domain.Account
	err := q.db.SelectContext(ctx, &accounts, `SELECT * FROM accounts ORDER BY created_at`)
	return accounts, err
}

func (q *Accounts) GetAccount(ctx context.Context, id int64) (*domain.Account, error) {
	var account domain.Account
	err := q.db.GetContext(ctx, &account, `SELECT * FROM accounts WHERE id=$1`, id)
	return &account, err
}

func (q *Accounts) SetAccountAnchor(ctx context.Context, id int64, date time.Time, balance float64) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE accounts
		    SET anchor_date=$1, anchor_balance=$2
		  WHERE id=$3`, date, balance, id)
	return err
}

func (q *Accounts) GetAccountBalance(ctx context.Context, id int64) (float64, error) {
	var current float64
	err := q.db.GetContext(ctx, &current, `
		SELECT a.anchor_balance
		     + COALESCE(SUM(CASE WHEN t.tx_direction='in'
		                         THEN  t.tx_amount
		                         ELSE -t.tx_amount END),0)
		  FROM accounts      a
		  LEFT JOIN transactions t
		         ON t.account_id = a.id
		        AND t.tx_date    > a.anchor_date
		 WHERE a.id=$1
		 GROUP BY a.anchor_balance`, id)
	return current, err
}
