package domain

import "time"

type Account struct {
	ID    int64   `db:"id"             json:"id"`
	Name  string  `db:"name"           json:"name"`
	Bank  string  `db:"bank"           json:"bank"`
	Type  string  `db:"type"           json:"type"`
	Alias *string `db:"alias"          json:"alias,omitempty"`

	AnchorDate    time.Time `db:"anchor_date"    json:"anchor_date"`
	AnchorBalance float64   `db:"anchor_balance" json:"anchor_balance"`

	CreatedAt time.Time `db:"created_at"     json:"created_at"`
}
