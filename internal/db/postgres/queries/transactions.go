package queries

import (
	"ariand/internal/db"
	"ariand/internal/domain"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type Transactions struct {
	db *sqlx.DB
}

func NewTransactions(db *sqlx.DB) *Transactions {
	return &Transactions{db}
}

func (q *Transactions) ListTransactions(ctx context.Context, options db.ListOpts) ([]domain.Transaction, error) {
	builder, args := strings.Builder{}, []any{}
	builder.WriteString("SELECT * FROM transactions")
	filters := []string{}

	if options.Start != nil {
		filters = append(filters, fmt.Sprintf("tx_date >= $%d", len(args)+1))
		args = append(args, *options.Start)
	}

	if options.End != nil {
		filters = append(filters, fmt.Sprintf("tx_date <= $%d", len(args)+1))
		args = append(args, *options.End)
	}

	if len(options.Accounts) > 0 {
		filters = append(filters, fmt.Sprintf("account_id = ANY($%d)", len(args)+1))
		args = append(args, options.Accounts)
	}

	if options.Direction == "in" || options.Direction == "out" {
		filters = append(filters, fmt.Sprintf("tx_direction = $%d", len(args)+1))
		args = append(args, options.Direction)
	}

	if len(filters) > 0 {
		builder.WriteString(" WHERE ")
		builder.WriteString(strings.Join(filters, " AND "))
	}

	builder.WriteString(" ORDER BY tx_date DESC, id DESC")

	if options.Limit > 0 {
		builder.WriteString(fmt.Sprintf(" LIMIT %d", options.Limit))
	}

	if options.Offset > 0 {
		builder.WriteString(fmt.Sprintf(" OFFSET %d", options.Offset))
	}

	var out []domain.Transaction
	if err := q.db.SelectContext(ctx, &out, builder.String(), args...); err != nil {
		return nil, err
	}

	return out, nil
}

func (q *Transactions) GetTransaction(ctx context.Context, id int64) (*domain.Transaction, error) {
	var transaction domain.Transaction
	err := q.db.GetContext(ctx, &transaction, `SELECT * FROM transactions WHERE id=$1`, id)
	if err != nil && strings.Contains(err.Error(), "no rows") {
		return nil, errors.New("not found")
	}
	return &transaction, err
}

func (q *Transactions) CreateTransaction(ctx context.Context, transaction *domain.Transaction) (int64, error) {
	const query = `
	INSERT INTO transactions (
	  email_id, account_id, tx_date, tx_amount, tx_currency,
	  tx_direction, tx_desc, category, merchant, user_notes,
	  foreign_currency, foreign_amount, exchange_rate
	) VALUES (
	  :email_id, :account_id, :tx_date, :tx_amount, :tx_currency,
	  :tx_direction, :tx_desc, :category, :merchant, :user_notes,
	  :foreign_currency, :foreign_amount, :exchange_rate
	) RETURNING id`

	rows, err := q.db.NamedQueryContext(ctx, query, transaction)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return 0, errors.New("conflict")
		}
		return 0, err
	}

	defer rows.Close()

	var id int64
	if rows.Next() {
		_ = rows.Scan(&id)
	}

	return id, nil
}

func (q *Transactions) UpdateTransaction(ctx context.Context, id int64, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}

	set, args := []string{}, []any{}

	i := 1
	for k, v := range fields {
		set = append(set, fmt.Sprintf("%s=$%d", k, i))
		args = append(args, v)
		i++
	}

	args = append(args, id)
	_, err := q.db.ExecContext(ctx,
		fmt.Sprintf(`UPDATE transactions SET %s WHERE id=$%d`, strings.Join(set, ","), i),
		args...)

	return err
}

func (q *Transactions) DeleteTransaction(ctx context.Context, id int64) error {
	res, err := q.db.ExecContext(ctx, `DELETE FROM transactions WHERE id=$1`, id)
	if err != nil {
		return err
	}

	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("not found")
	}

	return nil
}
