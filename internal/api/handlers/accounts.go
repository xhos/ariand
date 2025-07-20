package handlers

import (
	"ariand/internal/db"
	"ariand/internal/domain"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type AccountHandler struct{ Store db.Store }

type CreateAccountRequest struct {
	Name          string  `json:"name" example:"main chequing"`
	Bank          string  `json:"bank" example:"big bank inc."`
	Type          string  `json:"type" example:"chequing"`
	Alias         *string `json:"alias,omitempty" example:"main"`
	AnchorBalance float64 `json:"anchor_balance" example:"1234.56"`
}

type SetAnchorRequest struct {
	Balance float64 `json:"balance" example:"1234.56"`
}

// Create godoc
// @Summary      Create a new account
// @Description  Adds a new account to the database.
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        account  body      CreateAccountRequest  true  "new account object"
// @Success      201      {object}  domain.Account
// @Failure      400      {object}  ErrorResponse "invalid request body"
// @Failure      500      {object}  ErrorResponse
// @Router       /api/accounts [post]
// @Security     BearerAuth
func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		badRequest(w, "invalid json")
		return
	}

	acc := &domain.Account{
		Name:          req.Name,
		Bank:          req.Bank,
		Type:          req.Type,
		Alias:         req.Alias,
		AnchorDate:    time.Now(),
		AnchorBalance: req.AnchorBalance,
	}

	created, err := h.Store.CreateAccount(r.Context(), acc)
	if err != nil {
		internalErr(w)
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

// List godoc
// @Summary      List all accounts
// @Description  Returns a list of all accounts.
// @Tags         accounts
// @Produce      json
// @Success      200  {array}   domain.Account
// @Failure      500  {object}  ErrorResponse
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
// @Failure      400  {object}  ErrorResponse "invalid id format"
// @Failure      404  {object}  ErrorResponse "account not found"
// @Failure      500  {object}  ErrorResponse
// @Router       /api/accounts/{id} [get]
// @Security     BearerAuth
func (h *AccountHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromRequest(r)
	if err != nil {
		badRequest(w, "invalid id format")
		return
	}

	acc, err := h.Store.GetAccount(r.Context(), id)
	switch {
	case errors.Is(err, db.ErrNotFound):
		notFound(w)
	case err != nil:
		internalErr(w)
	default:
		writeJSON(w, http.StatusOK, acc)
	}
}

// Delete godoc
// @Summary      Delete an account
// @Description  Deletes an account and all of its associated transactions.
// @Tags         accounts
// @Param        id  path      int  true  "Account ID"
// @Success      204
// @Failure      400 {object}  ErrorResponse "invalid id format"
// @Failure      404 {object}  ErrorResponse "account not found"
// @Failure      500 {object}  ErrorResponse
// @Router       /api/accounts/{id} [delete]
// @Security     BearerAuth
func (h *AccountHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromRequest(r)
	if err != nil {
		badRequest(w, "invalid id format")
		return
	}

	if err := h.Store.DeleteAccount(r.Context(), id); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			notFound(w)
			return
		}
		internalErr(w)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// SetAnchor godoc
// @Summary      Update account anchor
// @Description  Updates the anchor balance for an account and sets the anchor date to now.
// @Tags         accounts
// @Accept       json
// @Param        id       path      int               true  "Account ID"
// @Param        payload  body      SetAnchorRequest  true  "Anchor Payload (balance only)"
// @Success      204
// @Failure      400      {object}  ErrorResponse "invalid request payload or id format"
// @Failure      404      {object}  ErrorResponse "account not found"
// @Failure      500      {object}  ErrorResponse
// @Router       /api/accounts/{id}/anchor [post]
// @Security     BearerAuth
func (h *AccountHandler) SetAnchor(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromRequest(r)
	if err != nil {
		badRequest(w, "invalid id format")
		return
	}

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
// @Description  Returns the current calculated balance of an account.
// @Tags         accounts
// @Produce      json
// @Param        id   path      int  true  "Account ID"
// @Success      200  {object}  BalanceResponse
// @Failure      400  {object}  ErrorResponse "invalid id format"
// @Failure      404  {object}  ErrorResponse "account not found"
// @Failure      500  {object}  ErrorResponse
// @Router       /api/accounts/{id}/balance [get]
// @Security     BearerAuth
func (h *AccountHandler) Balance(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromRequest(r)
	if err != nil {
		badRequest(w, "invalid id format")
		return
	}

	bal, err := h.Store.GetAccountBalance(r.Context(), id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			notFound(w)
			return
		}
		internalErr(w)
		return
	}

	writeJSON(w, http.StatusOK, BalanceResponse{Balance: bal})
}
