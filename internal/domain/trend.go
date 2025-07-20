package domain

type TrendPoint struct {
	Date     string  `json:"date"    db:"date"`
	Income   float64 `json:"income"  db:"income"`
	Expenses float64 `json:"expenses" db:"expenses"`
}
