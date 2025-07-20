package handlers

import (
	_ "ariand/docs"
	"ariand/internal/ai"
	"ariand/internal/db"
	"ariand/internal/service"
	"fmt"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func SetupRoutes(
	store db.Store,
	aiManager *ai.Manager,
	categorizer *service.Categorizer,
) *http.ServeMux {
	mux := http.NewServeMux()

	// transactions
	txn := &TransactionHandler{Store: store, Categorizer: categorizer}
	mux.HandleFunc("GET /api/transactions", txn.List)
	mux.HandleFunc("POST /api/transactions", txn.Create)
	mux.HandleFunc("GET /api/transactions/{id}", txn.Get)
	mux.HandleFunc("PATCH /api/transactions/{id}", txn.Patch)
	mux.HandleFunc("DELETE /api/transactions/{id}", txn.Delete)

	// categories
	cat := &CategoryHandler{Store: store}
	mux.HandleFunc("GET /api/categories", cat.List)
	mux.HandleFunc("POST /api/categories", cat.Create)
	mux.HandleFunc("GET /api/categories/{id}", cat.Get)
	mux.HandleFunc("PATCH /api/categories/{id}", cat.Patch)
	mux.HandleFunc("DELETE /api/categories/{id}", cat.Delete)

	// dashboard
	dash := &DashboardHandler{Store: store}
	mux.HandleFunc("GET /api/dashboard/balance", dash.Balance)
	mux.HandleFunc("GET /api/dashboard/debt", dash.Debt)
	mux.HandleFunc("GET /api/dashboard/trends", dash.Trends)

	// accounts
	acc := &AccountHandler{Store: store}
	mux.HandleFunc("GET /api/accounts", acc.List)
	mux.HandleFunc("POST /api/accounts", acc.Create)
	mux.HandleFunc("GET /api/accounts/{id}", acc.Get)
	mux.HandleFunc("DELETE /api/accounts/{id}", acc.Delete)
	mux.HandleFunc("POST /api/accounts/{id}/anchor", acc.SetAnchor)
	mux.HandleFunc("GET /api/accounts/{id}/balance", acc.Balance)

	// docs & health
	mux.Handle("/swagger/", httpSwagger.Handler(httpSwagger.URL("/docs/swagger.json")))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	return mux
}
