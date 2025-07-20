package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// newAuthedRequest creates an http request with the required auth header
func newAuthedRequest(t *testing.T, method, url, apiKey string, body io.Reader) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, url, body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	return req
}

// truncateTables clears all data from a list of test tables
func truncateTables(t *testing.T, db *sqlx.DB, tables ...string) {
	t.Helper()
	for _, table := range tables {
		_, err := db.ExecContext(context.Background(), fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
		require.NoError(t, err, "failed to truncate table "+table)
	}
}
