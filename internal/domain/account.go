package domain

import "time"

type AccountType string

const (
	AccountTypeChequing   AccountType = "chequing"
	AccountTypeSavings    AccountType = "savings"
	AccountTypeCreditCard AccountType = "credit_card"
	AccountTypeInvestment AccountType = "investment"
	AccountTypeOther      AccountType = "other"
)

type Account struct {
	ID    int64       `db:"id"   json:"id"`
	Name  string      `db:"name" json:"name"`
	Bank  string      `db:"bank" json:"bank"`
	Type  AccountType `db:"account_type" json:"type"`
	Alias *string     `db:"alias" json:"alias,omitempty"`

	AnchorDate    time.Time `db:"anchor_date"    json:"anchorDate"`
	AnchorBalance float64   `db:"anchor_balance" json:"anchorBalance"`

	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
}
