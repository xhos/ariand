package service

import (
	"ariand/internal/config"
	"ariand/internal/db"
	"ariand/internal/receiptparser"

	"github.com/charmbracelet/log"
)

type Services struct {
	Transactions TransactionService
	Categories   CategoryService
	Accounts     AccountService
	Dashboard    DashboardService
	Receipts     ReceiptService
}

func New(store db.Store, lg *log.Logger, cfg *config.Config) *Services {
	catSvc := newCatSvc(store, lg.WithPrefix("cat"))
	parserClient := receiptparser.New(cfg.ReceiptParserURL, cfg.ReceiptParserTimeout)

	return &Services{
		Transactions: newTxnSvc(store, lg.WithPrefix("txn"), catSvc),
		Categories:   catSvc,
		Accounts:     newAcctSvc(store, lg.WithPrefix("acct")),
		Dashboard:    newDashSvc(store),
		Receipts:     newReceiptSvc(store, parserClient, lg.WithPrefix("receipt")),
	}
}
