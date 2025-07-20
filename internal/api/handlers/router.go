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
	mux.HandleFunc("GET /api/transactions", HandleGet(txn.List))
	mux.HandleFunc("POST /api/transactions", HandleCreate(txn.Create))
	mux.HandleFunc("GET /api/transactions/{id}", HandleGet(txn.Get))
	mux.HandleFunc("PATCH /api/transactions/{id}", HandleUpdate(txn.Patch))
	mux.HandleFunc("DELETE /api/transactions/{id}", HandleUpdate(txn.Delete))
	// mux.HandleFunc("POST /api/transactions/{id}/categorize", HandleUpdate(txn.Categorize))
	// mux.HandleFunc("POST /api/transactions/{id}/identify-merchant", HandleUpdate(txn.IdentifyMerchant))

	// categories
	cat := &CategoryHandler{Store: store}
	mux.HandleFunc("GET /api/categories", HandleGet(cat.List))
	mux.HandleFunc("POST /api/categories", HandleCreate(cat.Create))
	mux.HandleFunc("GET /api/categories/{id}", HandleGet(cat.Get))
	mux.HandleFunc("PATCH /api/categories/{id}", HandleUpdate(cat.Patch))
	mux.HandleFunc("DELETE /api/categories/{id}", HandleUpdate(cat.Delete))

	// dashboard
	dash := &DashboardHandler{Store: store}
	mux.HandleFunc("GET /api/dashboard/balance", HandleGet(dash.Balance))
	mux.HandleFunc("GET /api/dashboard/debt", HandleGet(dash.Debt))
	mux.HandleFunc("GET /api/dashboard/trends", HandleGet(dash.Trends))

	// accounts
	acc := &AccountHandler{Store: store}
	mux.HandleFunc("GET /api/accounts", HandleGet(acc.List))
	mux.HandleFunc("POST /api/accounts", HandleCreate(acc.Create))
	mux.HandleFunc("GET /api/accounts/{id}", HandleGet(acc.Get))
	mux.HandleFunc("DELETE /api/accounts/{id}", HandleUpdate(acc.Delete))
	mux.HandleFunc("POST /api/accounts/{id}/anchor", HandleUpdate(acc.SetAnchor))
	mux.HandleFunc("GET /api/accounts/{id}/balance", HandleGet(acc.Balance))

	// docs & health
	mux.Handle("/swagger/", httpSwagger.Handler(httpSwagger.URL("/docs/swagger.json")))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	return mux
}
