package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
)

// ErrorResponse is a standard struct for api error responses
type ErrorResponse struct {
	Code    int    `json:"code" example:"404"`
	Message string `json:"message" example:"resource not found"`
}

// HTTPError represents an error with an associated HTTP status code.
type HTTPError struct {
	Code    int
	Message string
}

// Error makes it compatible with the error interface.
func (e *HTTPError) Error() string {
	return e.Message
}

// Write sends the error as a JSON response.
func (e *HTTPError) Write(w http.ResponseWriter) {
	errResp := ErrorResponse{Code: e.Code, Message: e.Message}
	writeJSON(w, e.Code, errResp)
}

// NewHTTPError is a constructor for HTTPError.
func NewHTTPError(code int, message string) *HTTPError {
	return &HTTPError{Code: code, Message: message}
}

// APIFunc is a handler that returns a value to be JSON encoded or an error.
type APIFunc func(r *http.Request) (any, *HTTPError)

// HandleGet converts an APIFunc to an http.HandlerFunc for GET requests (200 OK).
func HandleGet(fn APIFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := fn(r)
		if err != nil {
			err.Write(w)
			return
		}
		writeJSON(w, http.StatusOK, data)
	}
}

// HandleCreate converts an APIFunc to an http.HandlerFunc for POST requests (201 Created).
func HandleCreate(fn APIFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := fn(r)
		if err != nil {
			err.Write(w)
			return
		}
		writeJSON(w, http.StatusCreated, data)
	}
}

// HandleUpdate converts an APIFunc to an http.HandlerFunc for PATCH/POST/DELETE requests (204 No Content).
func HandleUpdate(fn APIFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := fn(r) // data is ignored
		if err != nil {
			err.Write(w)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// BalanceResponse is a generic response for endpoints returning a single balance value
type BalanceResponse struct {
	Balance float64 `json:"balance" example:"1234.56"`
}

// writeJSON is a helper for writing json responses
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Errorf("failed to write json response: %v", err)
	}
}

// parseIDFromRequest extracts and validates the numeric id from the url path
func parseIDFromRequest(r *http.Request) (int64, error) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// randomHexColor generates a random 6-digit hex color string
func randomHexColor() string {
	return "#" + uuid.NewString()[:6]
}
