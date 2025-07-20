package handlers

import (
	"ariand/internal/domain"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountHandlers(t *testing.T) {
	app, cleanup := newTestApp(t)
	defer cleanup()

	truncateTables(t, app.db.DB, "transactions", "accounts")

	vars := make(map[string]any)

	testCases := []struct {
		name           string
		method         string
		url            string
		body           any
		wantStatusCode int
		wantBody       string
		setupVars      func(t *testing.T, body []byte, vars map[string]any)
	}{
		{
			name:           "create account",
			method:         http.MethodPost,
			url:            "/api/accounts",
			body:           CreateAccountRequest{Name: "test chequing", Bank: "test bank", Type: "chequing", AnchorBalance: 1000.00},
			wantStatusCode: http.StatusCreated,
			wantBody:       `{"name":"test chequing", "bank":"test bank", "type":"chequing", "anchorBalance":1000}`,
			setupVars: func(t *testing.T, body []byte, vars map[string]any) {
				var resp domain.Account
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err)
				vars["accountID"] = resp.ID
			},
		},
		{
			name:           "get created account",
			method:         http.MethodGet,
			url:            "/api/accounts/%d",
			body:           nil,
			wantStatusCode: http.StatusOK,
			wantBody:       `{"name":"test chequing", "bank":"test bank", "type":"chequing", "anchorBalance":1000}`,
		},
		{
			name:           "set new anchor",
			method:         http.MethodPost,
			url:            "/api/accounts/%d/anchor",
			body:           SetAnchorRequest{Balance: 1500.75},
			wantStatusCode: http.StatusNoContent,
		},
		{
			name:           "get balance after anchor update",
			method:         http.MethodGet,
			url:            "/api/accounts/%d/balance",
			body:           nil,
			wantStatusCode: http.StatusOK,
			wantBody:       `{"balance": 1500.75}`,
		},
		{
			name:           "delete account",
			method:         http.MethodDelete,
			url:            "/api/accounts/%d",
			body:           nil,
			wantStatusCode: http.StatusNoContent,
		},
		{
			name:           "get deleted account",
			method:         http.MethodGet,
			url:            "/api/accounts/%d",
			body:           nil,
			wantStatusCode: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := tc.url
			if id, ok := vars["accountID"]; ok && strings.Contains(tc.url, "%d") {
				url = fmt.Sprintf(tc.url, id)
			}

			rr := app.do(t, tc.method, url, tc.body)
			assert.Equal(t, tc.wantStatusCode, rr.Code, "status code mismatch")

			if tc.setupVars != nil {
				tc.setupVars(t, rr.Body.Bytes(), vars)
			}

			if tc.wantBody != "" {
				var got, want any
				err := json.Unmarshal(rr.Body.Bytes(), &got)
				require.NoError(t, err, "failed to unmarshal actual response body")

				err = json.Unmarshal([]byte(tc.wantBody), &want)
				require.NoError(t, err, "failed to unmarshal wantBody")

				opts := cmpopts.IgnoreMapEntries(func(key string, val any) bool {
					return key == "id" || key == "anchorDate" || key == "createdAt"
				})

				diff := cmp.Diff(want, got, opts)
				assert.Empty(t, diff, "response body mismatch")
			}
		})
	}
}
