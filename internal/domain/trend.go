package domain

type TrendPoint struct {
	Date    string  `json:"date"    db:"date"`
	Income  float64 `json:"income"  db:"income"`
	Expense float64 `json:"expense" db:"expense"`
}
