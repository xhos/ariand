package handlers

import (
	_ "ariand/docs"
	"ariand/internal/service"
	"fmt"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func SetupRoutes(services *service.Services) *http.ServeMux {
	mux := http.NewServeMux()

	// transactions
	txnH := &TransactionHandler{service: services.Transactions}
	mux.HandleFunc("GET /api/transactions", HandleGet(txnH.List))
	mux.HandleFunc("POST /api/transactions", HandleCreate(txnH.Create))
	mux.HandleFunc("GET /api/transactions/{id}", HandleGet(txnH.Get))
	mux.HandleFunc("PATCH /api/transactions/{id}", HandleUpdate(txnH.Patch))
	mux.HandleFunc("DELETE /api/transactions/{id}", HandleUpdate(txnH.Delete))
	mux.HandleFunc("POST /api/transactions/{id}/categorize", HandleUpdate(txnH.Categorize))
	mux.HandleFunc("POST /api/transactions/{id}/identify-merchant", HandleUpdate(txnH.IdentifyMerchant))

	// categories
	catH := &CategoryHandler{service: services.Categories}
	mux.HandleFunc("GET /api/categories", HandleGet(catH.List))
	mux.HandleFunc("POST /api/categories", HandleCreate(catH.Create))
	mux.HandleFunc("GET /api/categories/{id}", HandleGet(catH.Get))
	mux.HandleFunc("PATCH /api/categories/{id}", HandleUpdate(catH.Patch))
	mux.HandleFunc("DELETE /api/categories/{id}", HandleUpdate(catH.Delete))

	// dashboard
	dashH := &DashboardHandler{service: services.Dashboard}
	mux.HandleFunc("GET /api/dashboard/balance", HandleGet(dashH.Balance))
	mux.HandleFunc("GET /api/dashboard/debt", HandleGet(dashH.Debt))
	mux.HandleFunc("GET /api/dashboard/trends", HandleGet(dashH.Trends))

	// accounts
	accH := &AccountHandler{service: services.Accounts}
	mux.HandleFunc("GET /api/accounts", HandleGet(accH.List))
	mux.HandleFunc("POST /api/accounts", HandleCreate(accH.Create))
	mux.HandleFunc("GET /api/accounts/{id}", HandleGet(accH.Get))
	mux.HandleFunc("DELETE /api/accounts/{id}", HandleUpdate(accH.Delete))
	mux.HandleFunc("POST /api/accounts/{id}/anchor", HandleUpdate(accH.SetAnchor))
	mux.HandleFunc("GET /api/accounts/{id}/balance", HandleGet(accH.Balance))

	// swagger & health
	mux.Handle("/swagger/", httpSwagger.Handler(httpSwagger.PersistAuthorization(true)))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	return mux
}
