package domain

import "time"

type Category struct {
	ID    int64  `db:"id"    json:"id"`
	Slug  string `db:"slug"  json:"slug"`
	Label string `db:"label" json:"label"`
	Color string `db:"color" json:"color"`

	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
}
