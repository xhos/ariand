package handlers

import (
	"ariand/internal/domain"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionHandlers(t *testing.T) {
	app, cleanup := newTestApp(t)
	defer cleanup()

	server := app.server
	ctx := context.Background()

	truncateTables(t, app.db.DB, "transactions", "categories", "accounts")

	acc1, err := app.db.CreateAccount(ctx, &domain.Account{Name: "Test Chequing", Type: "chequing", Bank: "Bank A", AnchorBalance: 1000, AnchorDate: time.Now()})
	require.NoError(t, err)

	acc2, err := app.db.CreateAccount(ctx, &domain.Account{Name: "Test Visa", Type: "credit_card", Bank: "Bank B", AnchorBalance: -500, AnchorDate: time.Now()})
	require.NoError(t, err)

	cat1ID, err := app.db.CreateCategory(ctx, "food.groceries", "Groceries", "#ff0000")
	require.NoError(t, err)

	t.Run("transaction lifecycle", func(t *testing.T) {
		var createdTransaction domain.Transaction

		t.Run("create", func(t *testing.T) {
			desc := "Superstore Groceries"
			emailID := "unique-email-1"
			payload := domain.Transaction{
				EmailID:     &emailID,
				AccountID:   acc1.ID,
				TxDate:      time.Now(),
				TxAmount:    55.43,
				TxCurrency:  "CAD",
				TxDirection: "out",
				TxDesc:      &desc,
				CategoryID:  &cat1ID,
			}
			body, err := json.Marshal(payload)
			require.NoError(t, err)

			req := newAuthedRequest(t, http.MethodPost, "/api/transactions", app.apiKey, bytes.NewReader(body))
			rr := httptest.NewRecorder()
			server.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusCreated, rr.Code)

			var resp CreateTransactionResponse
			err = json.NewDecoder(rr.Body).Decode(&resp)
			require.NoError(t, err)
			require.NotZero(t, resp.ID)
			createdTransaction = payload
			createdTransaction.ID = resp.ID
		})

		t.Run("create duplicate", func(t *testing.T) {
			emailID := "unique-email-1"
			payload := domain.Transaction{
				EmailID:     &emailID,
				AccountID:   acc1.ID,
				TxAmount:    1.0,
				TxCurrency:  "CAD",
				TxDirection: "out",
			}
			body, _ := json.Marshal(payload)
			req := newAuthedRequest(t, http.MethodPost, "/api/transactions", app.apiKey, bytes.NewReader(body))
			rr := httptest.NewRecorder()
			server.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusConflict, rr.Code)
		})

		t.Run("get", func(t *testing.T) {
			require.NotZero(t, createdTransaction.ID)
			url := fmt.Sprintf("/api/transactions/%d", createdTransaction.ID)
			req := newAuthedRequest(t, http.MethodGet, url, app.apiKey, nil)
			rr := httptest.NewRecorder()
			server.ServeHTTP(rr, req)

			require.Equal(t, http.StatusOK, rr.Code)
			var resp domain.Transaction
			err := json.NewDecoder(rr.Body).Decode(&resp)
			require.NoError(t, err)
			assert.Equal(t, createdTransaction.EmailID, resp.EmailID)
			assert.Equal(t, *createdTransaction.CategoryID, *resp.CategoryID)
		})

		t.Run("patch", func(t *testing.T) {
			require.NotZero(t, createdTransaction.ID)
			url := fmt.Sprintf("/api/transactions/%d", createdTransaction.ID)
			payload := map[string]any{"user_notes": "this was for the party"}
			body, _ := json.Marshal(payload)

			req := newAuthedRequest(t, http.MethodPatch, url, app.apiKey, bytes.NewReader(body))
			rr := httptest.NewRecorder()
			server.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusNoContent, rr.Code)

			getReq := newAuthedRequest(t, http.MethodGet, url, app.apiKey, nil)
			getRR := httptest.NewRecorder()
			server.ServeHTTP(getRR, getReq)
			require.Equal(t, http.StatusOK, getRR.Code)
			var resp domain.Transaction
			json.NewDecoder(getRR.Body).Decode(&resp)
			assert.Equal(t, "this was for the party", *resp.UserNotes)
		})

		t.Run("list and filter", func(t *testing.T) {
			tx2EmailID := "tx2"
			tx3EmailID := "tx3"
			_, err := app.db.CreateTransaction(ctx, &domain.Transaction{EmailID: &tx2EmailID, AccountID: acc1.ID, TxDate: time.Now(), TxAmount: 1200.00, TxDirection: "in", TxCurrency: "CAD"})
			require.NoError(t, err)
			_, err = app.db.CreateTransaction(ctx, &domain.Transaction{EmailID: &tx3EmailID, AccountID: acc2.ID, TxDate: time.Now(), TxAmount: 99.99, TxDirection: "out", TxCurrency: "CAD"})
			require.NoError(t, err)

			u, _ := url.Parse("/api/transactions")
			q := u.Query()
			q.Set("account_ids", fmt.Sprintf("%d", acc1.ID))
			q.Set("direction", "in")
			u.RawQuery = q.Encode()

			req := newAuthedRequest(t, http.MethodGet, u.String(), app.apiKey, nil)
			rr := httptest.NewRecorder()
			server.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusOK, rr.Code)

			var resp ListTransactionsResponse
			err = json.NewDecoder(rr.Body).Decode(&resp)
			require.NoError(t, err)
			assert.Len(t, resp.Transactions, 1)
			assert.Equal(t, float64(1200.00), resp.Transactions[0].TxAmount)
		})

		t.Run("list with pagination", func(t *testing.T) {
			req := newAuthedRequest(t, http.MethodGet, "/api/transactions?limit=2", app.apiKey, nil)
			rr := httptest.NewRecorder()
			server.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusOK, rr.Code)

			var resp ListTransactionsResponse
			err = json.NewDecoder(rr.Body).Decode(&resp)
			require.NoError(t, err)
			assert.Len(t, resp.Transactions, 2)
			require.NotNil(t, resp.NextCursor)
			assert.NotZero(t, resp.NextCursor.ID)
		})

		t.Run("delete", func(t *testing.T) {
			require.NotZero(t, createdTransaction.ID)
			url := fmt.Sprintf("/api/transactions/%d", createdTransaction.ID)
			req := newAuthedRequest(t, http.MethodDelete, url, app.apiKey, nil)
			rr := httptest.NewRecorder()
			server.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusNoContent, rr.Code)

			getReq := newAuthedRequest(t, http.MethodGet, url, app.apiKey, nil)
			getRR := httptest.NewRecorder()
			server.ServeHTTP(getRR, getReq)
			assert.Equal(t, http.StatusNotFound, getRR.Code)
		})
	})
}
