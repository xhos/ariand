package service

import (
	"ariand/internal/db"

	"github.com/charmbracelet/log"
)

type Services struct {
	Transactions TransactionService
	Categories   CategoryService
	Accounts     AccountService
	Dashboard    DashboardService
	Receipts     ReceiptService
}

func New(store db.Store, lg *log.Logger) *Services {
	catSvc := newCatSvc(store, lg.WithPrefix("cat"))
	return &Services{
		Transactions: newTxnSvc(store, lg.WithPrefix("txn"), catSvc),
		Categories:   catSvc,
		Accounts:     newAcctSvc(store, lg.WithPrefix("acct")),
		Dashboard:    newDashSvc(store),
		Receipts:     newReceiptSvc(),
	}
}
