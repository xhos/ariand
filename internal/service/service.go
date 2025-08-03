package service

import (
	"ariand/internal/ai"
	"ariand/internal/config"
	sqlc "ariand/internal/db/sqlc"

	"github.com/charmbracelet/log"
)

type Services struct {
	Transactions TransactionService
	Categories   CategoryService
	Accounts     AccountService
	Dashboard    DashboardService
	Receipts     ReceiptService
	Users        UserService
	Auth         AuthService
}

func New(queries *sqlc.Queries, lg *log.Logger, cfg *config.Config, aiMgr *ai.Manager) *Services {
	catSvc := newCatSvc(queries, lg.WithPrefix("cat"))
	// parserClient := receiptparser.New(cfg.ReceiptParserURL, cfg.ReceiptParserTimeout)

	return &Services{
		Transactions: newTxnSvc(queries, lg.WithPrefix("txn"), catSvc, aiMgr),
		Categories:   catSvc,
		Accounts:     newAcctSvc(queries, lg.WithPrefix("acct")),
		Dashboard:    newDashSvc(queries),
		Users:        newUserSvc(queries, lg.WithPrefix("user")),
		Auth:         newAuthSvc(queries, lg.WithPrefix("auth")),
		// Receipts:     newReceiptSvc(queries, parserClient, lg.WithPrefix("receipt")),
	}
}
