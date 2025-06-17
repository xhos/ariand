package handlers

import (
	"ariand/internal/db"
	_ "ariand/internal/domain"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

type AccountHandler struct{ Store db.Store }

type HTTPError struct {
	Code    int    `json:"code" example:"500"`
	Message string `json:"message" example:"internal server error"`
}

type SetAnchorRequest struct {
	Balance float64 `json:"balance" example:"1234.56"`
}

type BalanceResponse struct {
	Balance float64 `json:"balance" example:"1234.56"`
}

// List godoc
// @Summary      List all accounts
// @Description  Returns a list of all accounts.
// @Tags         accounts
// @Produce      json
// @Success      200  {array}   domain.Account
// @Failure      500  {object}  HTTPError
// @Router       /api/accounts [get]
// @Security     BearerAuth
func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	accts, err := h.Store.ListAccounts(r.Context())
	if err != nil {
		internalErr(w)
		return
	}

	writeJSON(w, http.StatusOK, accts)
}

// Get godoc
// @Summary      Get account by ID
// @Description  Returns a single account by its numeric ID.
// @Tags         accounts
// @Produce      json
// @Param        id   path      int  true  "Account ID"
// @Success      200  {object}  domain.Account
// @Failure      404  {object}  HTTPError "account not found"
// @Failure      500  {object}  HTTPError
// @Router       /api/accounts/{id} [get]
// @Security     BearerAuth
func (h *AccountHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	acc, err := h.Store.GetAccount(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { // Use db.ErrNotFound if you have it
			notFound(w)
			return
		}
		internalErr(w)
		return
	}

	writeJSON(w, http.StatusOK, acc)
}

// SetAnchor godoc
// @Summary      Set account anchor to now
// @Description  Defines a true balance for an account at the current time. This anchor is the starting point for all balance calculations.
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        id       path      int               true  "Account ID"
// @Param        payload  body      SetAnchorRequest  true  "Anchor Payload (balance only)"
// @Success      204
// @Failure      400      {object}  HTTPError "invalid request payload"
// @Failure      404      {object}  HTTPError "account not found"
// @Failure      500      {object}  HTTPError
// @Router       /api/accounts/{id}/anchor [post]
// @Security     BearerAuth
func (h *AccountHandler) SetAnchor(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)

	var in SetAnchorRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		badRequest(w, "invalid json payload")
		return
	}

	if err := h.Store.SetAccountAnchor(r.Context(), id, in.Balance); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			notFound(w)
			return
		}
		internalErr(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Balance godoc
// @Summary      Get current balance
// @Description  Returns the current calculated balance of an account based on its anchor and subsequent transactions.
// @Tags         accounts
// @Produce      json
// @Param        id   path      int  true  "Account ID"
// @Success      200  {object}  BalanceResponse
// @Failure      404  {object}  HTTPError "account not found"
// @Failure      500  {object}  HTTPError
// @Router       /api/accounts/{id}/balance [get]
// @Security     BearerAuth
func (h *AccountHandler) Balance(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	bal, err := h.Store.GetAccountBalance(r.Context(), id)
	if err != nil {
		// Your GetAccountBalance query returns 0 for a non-existent ID.
		// For a more robust API, you could first check if the account exists,
		// or change the query to error out if the ID is not found.
		// Assuming the latter for this doc:
		if errors.Is(err, db.ErrNotFound) {
			notFound(w)
			return
		}
		internalErr(w)
		return
	}

	writeJSON(w, http.StatusOK, BalanceResponse{Balance: bal})
}
