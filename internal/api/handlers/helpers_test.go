// internal/api/handlers/helpers_test.go
package handlers

import (
	"ariand/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHelpers(t *testing.T) {
	t.Run("internalErr", func(t *testing.T) {
		rr := httptest.NewRecorder()
		internalErr(rr)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		var body ErrorResponse
		err := json.NewDecoder(rr.Body).Decode(&body)
		require.NoError(t, err)

		assert.Equal(t, http.StatusInternalServerError, body.Code)
		assert.Equal(t, "internal server error", body.Message)
	})

	t.Run("notFound", func(t *testing.T) {
		rr := httptest.NewRecorder()
		notFound(rr)

		assert.Equal(t, http.StatusNotFound, rr.Code)

		var body ErrorResponse
		err := json.NewDecoder(rr.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "resource not found", body.Message)
	})

	t.Run("badRequest", func(t *testing.T) {
		rr := httptest.NewRecorder()
		customMessage := "invalid field: email"
		badRequest(rr, customMessage)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var body ErrorResponse
		err := json.NewDecoder(rr.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, customMessage, body.Message)
	})
}

// newauthedrequest creates an http request with the required auth header
func newAuthedRequest(t *testing.T, method, url, apiKey string, body io.Reader) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, url, body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	return req
}

// truncatetables clears all data from a list of test tables
func truncateTables(t *testing.T, db *sqlx.DB, tables ...string) {
	t.Helper()
	for _, table := range tables {
		_, err := db.ExecContext(context.Background(), fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
		require.NoError(t, err, "failed to truncate table "+table)
	}
}

// createtestaccount is a helper to insert an account directly for testing purposes
func createTestAccount(t *testing.T, app *testApp, acc domain.Account) domain.Account {
	t.Helper()
	created, err := app.db.CreateAccount(context.Background(), &acc)
	require.NoError(t, err)
	return *created
}
