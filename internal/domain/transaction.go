package domain

import (
	"time"

	"github.com/lib/pq"
)

type Transaction struct {
	ID        int64   `db:"id"         json:"id"`
	EmailID   *string `db:"email_id"   json:"emailId,omitempty"`
	AccountID int64   `db:"account_id" json:"accountId"`

	TxDate      time.Time `db:"tx_date"      json:"txDate"`
	TxAmount    float64   `db:"tx_amount"    json:"txAmount"`
	TxCurrency  string    `db:"tx_currency"  json:"txCurrency"`
	TxDirection string    `db:"tx_direction" json:"txDirection"`
	TxDesc      *string   `db:"tx_desc"      json:"txDesc,omitempty"`

	BalanceAfter *float64 `db:"balance_after" json:"balanceAfter,omitempty"`

	CategoryID    *int64         `db:"category_id"   json:"categoryId,omitempty"`
	CategorySlug  *string        `db:"category_slug" json:"categorySlug,omitempty"`
	CategoryLabel *string        `db:"category_label" json:"categoryLabel,omitempty"`
	CategoryColor *string        `db:"category_color" json:"categoryColor,omitempty"`
	CatStatus     string         `db:"cat_status"    json:"catStatus"`
	Merchant      *string        `db:"merchant"      json:"merchant,omitempty"`
	UserNotes     *string        `db:"user_notes"    json:"userNotes,omitempty"`
	Suggestions   pq.StringArray `json:"suggestions" db:"suggestions" swaggertype:"array,string"`

	ReceiptID *int64 `db:"receipt_id" json:"receiptId,omitempty"`

	ForeignAmount   *float64 `db:"foreign_amount"   json:"foreignAmount,omitempty"`
	ForeignCurrency *string  `db:"foreign_currency" json:"foreignCurrency,omitempty"`
	ExchangeRate    *float64 `db:"exchange_rate"    json:"exchangeRate,omitempty"`

	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
}

type TransactionWithScore struct {
	Transaction
	MerchantScore float64 `db:"merchant_score" json:"merchantScore"`
}
