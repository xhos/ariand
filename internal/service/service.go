package service

import (
	"ariand/internal/ai"
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
	Users        UserService
	Auth         AuthService
}

func New(database *db.DB, lg *log.Logger, cfg *config.Config, aiMgr *ai.Manager) (*Services, error) {
	queries := database.Queries
	catSvc := newCatSvc(queries, lg.WithPrefix("cat"))

	// Initialize the gRPC receipt parser client
	parserClient, err := receiptparser.New(cfg.ReceiptParserURL, cfg.ReceiptParserTimeout)
	if err != nil {
		return nil, err
	}

	return &Services{
		Transactions: newTxnSvc(queries, lg.WithPrefix("txn"), catSvc, aiMgr),
		Categories:   catSvc,
		Accounts:     newAcctSvc(queries, lg.WithPrefix("acct")),
		Dashboard:    newDashSvc(queries),
		Users:        newUserSvc(queries, database, lg.WithPrefix("user")), //TODO: WHY PASS DB?
		Auth:         newAuthSvc(queries, lg.WithPrefix("auth")),
		Receipts:     newReceiptSvc(queries, parserClient, lg.WithPrefix("receipt")),
	}, nil
}
