package handlers

import (
	"ariand/internal/db"
	"ariand/internal/domain"
	"ariand/internal/service"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type AccountHandler struct {
	service service.AccountService
}

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
func (h *AccountHandler) Create(r *http.Request) (any, *HTTPError) {
	var req CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid json")
	}

	account := &domain.Account{
		Name:          req.Name,
		Bank:          req.Bank,
		Type:          domain.AccountType(req.Type),
		Alias:         req.Alias,
		AnchorDate:    time.Now(),
		AnchorBalance: req.AnchorBalance,
	}

	created, err := h.service.Create(r.Context(), account)
	if err != nil {
		return nil, NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return created, nil
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
func (h *AccountHandler) List(r *http.Request) (any, *HTTPError) {
	accounts, err := h.service.List(r.Context())
	if err != nil {
		return nil, NewHTTPError(
			http.StatusInternalServerError,
			"internal server error",
		)
	}
	return accounts, nil
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
func (h *AccountHandler) Get(r *http.Request) (any, *HTTPError) {
	id, err := parseIDFromRequest(r)
	if err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid id format")
	}

	account, err := h.service.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, NewHTTPError(http.StatusNotFound, "resource not found")
		}
		return nil, NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return account, nil
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
func (h *AccountHandler) Delete(r *http.Request) (any, *HTTPError) {
	id, err := parseIDFromRequest(r)
	if err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid id format")
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, NewHTTPError(http.StatusNotFound, "resource not found")
		}
		return nil, NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return nil, nil
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
func (h *AccountHandler) SetAnchor(r *http.Request) (any, *HTTPError) {
	id, err := parseIDFromRequest(r)
	if err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid id format")
	}

	var in SetAnchorRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid json payload")
	}

	if err := h.service.SetAnchor(r.Context(), id, in.Balance); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, NewHTTPError(http.StatusNotFound, "resource not found")
		}
		return nil, NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return nil, nil
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
func (h *AccountHandler) Balance(r *http.Request) (any, *HTTPError) {
	id, err := parseIDFromRequest(r)
	if err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid id format")
	}

	bal, err := h.service.Balance(r.Context(), id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, NewHTTPError(http.StatusNotFound, "resource not found")
		}
		return nil, NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return BalanceResponse{Balance: bal}, nil
}
