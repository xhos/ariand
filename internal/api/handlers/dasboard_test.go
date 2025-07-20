package handlers

import (
	"ariand/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDashboardHandlers(t *testing.T) {
	app, cleanup := newTestApp(t)
	defer cleanup()
	ctx := context.Background()

	// setup: create a consistent set of data for all test cases
	truncateTables(t, app.db.DB, "transactions", "accounts")

	anchorDate := time.Date(2025, 7, 18, 0, 0, 0, 0, time.UTC)
	today := time.Date(2025, 7, 20, 12, 0, 0, 0, time.UTC)
	yesterday := today.AddDate(0, 0, -1)

	chequing, err := app.db.CreateAccount(ctx, &domain.Account{
		Name: "Chequing Account", Bank: "Test Bank", Type: "Chequing",
		AnchorDate: anchorDate, AnchorBalance: 1200.50,
	})
	require.NoError(t, err)

	visa, err := app.db.CreateAccount(ctx, &domain.Account{
		Name: "Visa Card", Bank: "Credit Corp", Type: "Credit Card",
		AnchorDate: anchorDate, AnchorBalance: -250.25,
	})
	require.NoError(t, err)

	amex, err := app.db.CreateAccount(ctx, &domain.Account{
		Name: "Amex Card", Bank: "Credit Corp", Type: "Credit Card",
		AnchorDate: anchorDate, AnchorBalance: -1300.00,
	})
	require.NoError(t, err)

	_, err = app.db.CreateTransaction(ctx, &domain.Transaction{EmailID: "tx1", AccountID: chequing.ID, TxDate: today, TxAmount: 100.00, TxCurrency: "CAD", TxDirection: "in"})
	require.NoError(t, err)
	_, err = app.db.CreateTransaction(ctx, &domain.Transaction{EmailID: "tx2", AccountID: chequing.ID, TxDate: yesterday, TxAmount: 25.50, TxCurrency: "CAD", TxDirection: "out"})
	require.NoError(t, err)
	_, err = app.db.CreateTransaction(ctx, &domain.Transaction{EmailID: "tx3", AccountID: visa.ID, TxDate: today, TxAmount: 40.10, TxCurrency: "CAD", TxDirection: "out"})
	require.NoError(t, err)
	_, err = app.db.CreateTransaction(ctx, &domain.Transaction{EmailID: "tx4", AccountID: amex.ID, TxDate: yesterday, TxAmount: 80.00, TxCurrency: "CAD", TxDirection: "out"})
	require.NoError(t, err)

	testCases := []struct {
		name           string
		url            string
		wantStatusCode int
		wantBody       string
		cmpOpts        []cmp.Option
	}{
		{
			name:           "get balance",
			url:            "/api/dashboard/balance",
			wantStatusCode: http.StatusOK,
			// expected = -349.75 (anchors) + (100 - 25.50 - 40.10 - 80.00) (transactions) = -395.35
			wantBody: `{"balance": -395.35}`,
		},
		{
			name:           "get debt",
			url:            "/api/dashboard/debt",
			wantStatusCode: http.StatusOK,
			// expected = -1550.25 (credit anchors) + (-40.10 - 80.00) (credit transactions) = -1670.35
			wantBody: `{"debt": -1670.35}`,
		},
		{
			name:           "get trends",
			url:            "/api/dashboard/trends",
			wantStatusCode: http.StatusOK,
			wantBody: fmt.Sprintf(`[
				{"date": "%s", "income": 0, "expenses": 105.50},
				{"date": "%s", "income": 100.00, "expenses": 40.10}
			]`, yesterday.Format("2006-01-02"), today.Format("2006-01-02")),
			cmpOpts: []cmp.Option{
				// sort slices by date to ensure a stable comparison order
				cmpopts.SortSlices(func(a, b map[string]any) bool {
					return a["date"].(string) < b["date"].(string)
				}),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rr := app.do(t, http.MethodGet, tc.url, nil)
			assert.Equal(t, tc.wantStatusCode, rr.Code, "status code mismatch")

			if tc.wantBody != "" {
				var got, want any
				err := json.Unmarshal(rr.Body.Bytes(), &got)
				require.NoError(t, err, "failed to unmarshal actual response body")

				err = json.Unmarshal([]byte(tc.wantBody), &want)
				require.NoError(t, err, "failed to unmarshal wantBody")

				// use a small tolerance for floating point comparisons
				opts := []cmp.Option{cmpopts.EquateApprox(0, 0.001)}
				opts = append(opts, tc.cmpOpts...)

				diff := cmp.Diff(want, got, opts...)
				assert.Empty(t, diff, "response body mismatch")
			}
		})
	}
}
