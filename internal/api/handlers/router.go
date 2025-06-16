package handlers

import (
	_ "ariand/docs"
	"ariand/internal/db"
	"github.com/swaggo/http-swagger/v2"
	"net/http"
)

func SetupRoutes(store db.Store) *http.ServeMux {
	mux := http.NewServeMux()
	txn := &TransactionHandler{Store: store}
	acc := &AccountHandler{Store: store}
	dash := &DashboardHandler{Store: store}

	// transactions
	mux.Handle("GET    /api/transactions", http.HandlerFunc(txn.List))
	mux.Handle("POST   /api/transactions", http.HandlerFunc(txn.Create))
	mux.Handle("GET    /api/transactions/{id}", http.HandlerFunc(txn.Get))
	mux.Handle("PATCH  /api/transactions/{id}", http.HandlerFunc(txn.Patch))
	mux.Handle("DELETE /api/transactions/{id}", http.HandlerFunc(txn.Delete))

	// dashboard
	mux.Handle("GET /api/dashboard/balance", http.HandlerFunc(dash.Balance))
	mux.Handle("GET /api/dashboard/debt", http.HandlerFunc(dash.Debt))
	mux.Handle("GET /api/transactions/trends", http.HandlerFunc(dash.Trends))

	// accounts
	mux.Handle("GET    /api/accounts", http.HandlerFunc(acc.List))
	mux.Handle("GET    /api/accounts/{id}", http.HandlerFunc(acc.Get))
	mux.Handle("POST   /api/accounts/{id}/anchor", http.HandlerFunc(acc.SetAnchor))
	mux.Handle("GET    /api/accounts/{id}/balance", http.HandlerFunc(acc.Balance))

	// docs
	mux.Handle("/docs/", http.StripPrefix("/docs/", http.FileServer(http.Dir("docs"))))
	mux.Handle("/swagger/", httpSwagger.Handler(httpSwagger.URL("http://localhost:8080/docs/swagger.json")))

	return mux
}
