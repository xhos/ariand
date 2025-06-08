package domain

import "time"

type Transaction struct {
	ID        int64  `db:"id"             json:"id"`
	EmailID   string `db:"email_id"       json:"email_id"`
	AccountID int64  `db:"account_id"     json:"account_id"`

	TxDate       time.Time `db:"tx_date"        json:"tx_date"`
	TxAmount     float64   `db:"tx_amount"      json:"tx_amount"`
	TxCurrency   string    `db:"tx_currency"    json:"tx_currency"`
	TxDirection  string    `db:"tx_direction"   json:"tx_direction"`
	TxDesc       *string   `db:"tx_desc"        json:"tx_desc,omitempty"`
	BalanceAfter *float64  `db:"balance_after"  json:"balance_after,omitempty"`

	Category  *string `db:"category"       json:"category,omitempty"`
	Merchant  *string `db:"merchant"       json:"merchant,omitempty"`
	UserNotes *string `db:"user_notes"     json:"user_notes,omitempty"`

	ForeignAmount   *float64 `db:"foreign_amount"   json:"foreign_amount,omitempty"`
	ForeignCurrency *string  `db:"foreign_currency" json:"foreign_currency,omitempty"`
	ExchangeRate    *float64 `db:"exchange_rate"    json:"exchange_rate,omitempty"`
}
