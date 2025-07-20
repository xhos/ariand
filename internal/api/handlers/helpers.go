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

// badRequest sends a 400 bad request response
func badRequest(w http.ResponseWriter, message string) {
	err := ErrorResponse{Code: http.StatusBadRequest, Message: message}
	writeJSON(w, http.StatusBadRequest, err)
}

// notFound sends a 404 not found response
func notFound(w http.ResponseWriter) {
	err := ErrorResponse{Code: http.StatusNotFound, Message: "resource not found"}
	writeJSON(w, http.StatusNotFound, err)
}

// internalErr sends a 500 internal server error response
func internalErr(w http.ResponseWriter) {
	err := ErrorResponse{Code: http.StatusInternalServerError, Message: "internal server error"}
	writeJSON(w, http.StatusInternalServerError, err)
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
