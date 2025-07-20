package domain

type Category struct {
	ID    int64  `db:"id"   json:"id"`
	Slug  string `db:"slug" json:"slug"`
	Label string `db:"label" json:"label"`
	Color string `db:"color" json:"color"`
}
