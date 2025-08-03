package db

import (
	"time"

	"github.com/shopspring/decimal"
)

type ListOpts struct {
	// cursor for pagination
	CursorID   *int64
	CursorDate *time.Time

	// filtering
	Start             *time.Time
	End               *time.Time
	AccountIDs        []int64
	Categories        []string
	Direction         string
	MerchantSearch    string
	DescriptionSearch string
	AmountMin         *decimal.Decimal
	AmountMax         *decimal.Decimal
	Currency          string
	TimeOfDayStart    *string
	TimeOfDayEnd      *string

	// pagination limit
	Limit int
}
