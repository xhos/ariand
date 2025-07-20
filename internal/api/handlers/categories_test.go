package handlers

import (
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

func TestCategoryHandlers(t *testing.T) {
	app, cleanup := newTestApp(t)
	defer cleanup()

	truncateTables(t, app.db.DB, "transactions", "categories", "accounts")

	vars := make(map[string]any)

	testCases := []struct {
		name           string
		method         string
		url            string
		body           any
		wantStatusCode int
		wantBody       string
		setupVars      func(t *testing.T, body []byte, vars map[string]any)
		cmpOpts        []cmp.Option
	}{
		{
			name:           "create first category",
			method:         http.MethodPost,
			url:            "/api/categories",
			body:           CreateCategoryRequest{Slug: "test.food", Label: "Test Food", Color: "#FF0000"},
			wantStatusCode: http.StatusCreated,
			setupVars: func(t *testing.T, body []byte, vars map[string]any) {
				var resp CreateCategoryResponse
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err)
				vars["categoryID"] = resp.ID
			},
		},
		{
			name:           "get created category",
			method:         http.MethodGet,
			url:            "/api/categories/%d",
			body:           nil,
			wantStatusCode: http.StatusOK,
			wantBody:       `{"slug":"test.food", "label":"Test Food", "color":"#FF0000"}`,
		},
		{
			name:           "create second category with random color",
			method:         http.MethodPost,
			url:            "/api/categories",
			body:           CreateCategoryRequest{Slug: "test.shopping", Label: "Shopping"},
			wantStatusCode: http.StatusCreated,
		},
		{
			name:           "list all categories",
			method:         http.MethodGet,
			url:            "/api/categories",
			body:           nil,
			wantStatusCode: http.StatusOK,
			wantBody: `[
				{"slug":"test.food", "label":"Test Food", "color":"#FF0000"},
				{"slug":"test.shopping", "label":"Shopping"}
			]`,
			cmpOpts: []cmp.Option{
				cmpopts.SortSlices(func(a, b map[string]any) bool {
					return a["slug"].(string) < b["slug"].(string)
				}),
				cmpopts.IgnoreMapEntries(func(key string, val any) bool {
					return key == "color" && val != "#FF0000"
				}),
			},
		},
		{
			name:           "patch category",
			method:         http.MethodPatch,
			url:            "/api/categories/%d",
			body:           map[string]any{"label": "New Test Label"},
			wantStatusCode: http.StatusNoContent,
		},
		{
			name:           "get patched category",
			method:         http.MethodGet,
			url:            "/api/categories/%d",
			body:           nil,
			wantStatusCode: http.StatusOK,
			wantBody:       `{"slug":"test.food", "label":"New Test Label", "color":"#FF0000"}`,
		},
		{
			name:           "delete category",
			method:         http.MethodDelete,
			url:            "/api/categories/%d",
			body:           nil,
			wantStatusCode: http.StatusNoContent,
		},
		{
			name:           "get deleted category",
			method:         http.MethodGet,
			url:            "/api/categories/%d",
			body:           nil,
			wantStatusCode: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := tc.url
			if id, ok := vars["categoryID"]; ok && strings.Contains(tc.url, "%d") {
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

				opts := []cmp.Option{
					cmpopts.IgnoreMapEntries(func(key string, val any) bool {
						return key == "id"
					}),
				}
				opts = append(opts, tc.cmpOpts...)

				diff := cmp.Diff(want, got, opts...)
				assert.Empty(t, diff, "response body mismatch")
			}
		})
	}
}
