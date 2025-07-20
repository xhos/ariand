package domain

import (
	"time"

	"github.com/lib/pq"
)

type Transaction struct {
	ID        int64  `db:"id"          json:"id"`
	EmailID   string `db:"email_id"    json:"emailId"`
	AccountID int64  `db:"account_id"  json:"accountId"`

	TxDate      time.Time `db:"tx_date"      json:"txDate"`
	TxAmount    float64   `db:"tx_amount"    json:"txAmount"`
	TxCurrency  string    `db:"tx_currency"  json:"txCurrency"`
	TxDirection string    `db:"tx_direction" json:"txDirection"`
	TxDesc      *string   `db:"tx_desc"      json:"txDesc,omitempty"`

	BalanceAfter float64 `db:"balance_after" json:"balanceAfter"`

	CategoryID   *int64         `db:"category_id"   json:"categoryId,omitempty"`
	CategorySlug *string        `db:"category_slug" json:"categorySlug,omitempty"`
	CatStatus    string         `db:"cat_status"    json:"catStatus"`
	Merchant     *string        `db:"merchant"      json:"merchant,omitempty"`
	UserNotes    *string        `db:"user_notes"    json:"userNotes,omitempty"`
	Suggestions  pq.StringArray `json:"suggestions" db:"suggestions" swaggertype:"array,string"`

	ForeignAmount   *float64 `db:"foreign_amount"    json:"foreignAmount,omitempty"`
	ForeignCurrency *string  `db:"foreign_currency"  json:"foreignCurrency,omitempty"`
	ExchangeRate    *float64 `db:"exchange_rate"     json:"exchangeRate,omitempty"`
}
