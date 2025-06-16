package handlers

import (
	"ariand/internal/db"
	"ariand/internal/domain"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
)

type TransactionHandler struct{ Store db.Store }

type CreateTransactionResponse struct {
	ID int64 `json:"id" example:"101"`
}

type UpdateTransactionRequest map[string]any

// List godoc
// @Summary      List transactions
// @Description  Returns a list of transactions, with optional filtering by date range.
// @Tags         transactions
// @Produce      json
// @Param        start  query     string  false  "Start date filter (YYYY-MM-DD)"
// @Param        end    query     string  false  "End date filter (YYYY-MM-DD)"
// @Success      200    {array}   domain.Transaction
// @Failure      400    {object}  HTTPError "invalid date format"
// @Failure      500    {object}  HTTPError
// @Router       /api/transactions [get]
// @Security     BearerAuth
func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	var opt db.ListOpts
	q := r.URL.Query()

	if s := q.Get("start"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			badRequest(w, "invalid start date format, expected YYYY-MM-DD")
			return
		}
		opt.Start = &t
	}
	if e := q.Get("end"); e != "" {
		t, err := time.Parse("2006-01-02", e)
		if err != nil {
			badRequest(w, "invalid end date format, expected YYYY-MM-DD")
			return
		}
		opt.End = &t
	}

	out, err := h.Store.ListTransactions(r.Context(), opt)
	if err != nil {
		internalErr(w)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// Get godoc
// @Summary      Get a single transaction
// @Description  Retrieves a transaction by its numeric ID.
// @Tags         transactions
// @Produce      json
// @Param        id   path      int  true  "Transaction ID"
// @Success      200  {object}  domain.Transaction
// @Failure      404  {object}  HTTPError "transaction not found"
// @Failure      500  {object}  HTTPError
// @Router       /api/transactions/{id} [get]
// @Security     BearerAuth
func (h *TransactionHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	txn, err := h.Store.GetTransaction(r.Context(), id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			notFound(w)
			return
		}
		internalErr(w)
		return
	}

	writeJSON(w, http.StatusOK, txn)
}

// Create godoc
// @Summary      Create a new transaction
// @Description  Adds a new transaction to the database.
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        transaction  body      domain.Transaction         true  "Transaction object"
// @Success      201          {object}  CreateTransactionResponse
// @Failure      400          {object}  HTTPError "invalid request body"
// @Failure      409          {object}  HTTPError "duplicate transaction"
// @Failure      500          {object}  HTTPError
// @Router       /api/transactions [post]
// @Security     BearerAuth
func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in domain.Transaction
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		badRequest(w, "invalid json payload")
		return
	}

	id, err := h.Store.CreateTransaction(r.Context(), &in)
	switch {
	case err == nil:
		writeJSON(w, http.StatusCreated, CreateTransactionResponse{ID: id})
	case errors.Is(err, db.ErrConflict):
		writeJSON(w, http.StatusConflict, HTTPError{Code: http.StatusConflict, Message: "duplicate transaction"})
	default:
		internalErr(w)
	}
}

// Patch godoc
// @Summary      Update a transaction
// @Description  Partially updates a transaction's fields. Only the provided fields will be changed.
// @Tags         transactions
// @Accept       json
// @Param        id      path      int                     true  "Transaction ID"
// @Param        fields  body      UpdateTransactionRequest  true  "Fields to update"
// @Success      204
// @Failure      400     {object}  HTTPError "invalid request body"
// @Failure      404     {object}  HTTPError "transaction not found"
// @Failure      500     {object}  HTTPError
// @Router       /api/transactions/{id} [patch]
// @Security     BearerAuth
func (h *TransactionHandler) Patch(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	var fields map[string]any
	if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
		badRequest(w, "invalid json payload")
		return
	}

	if err := h.Store.UpdateTransaction(r.Context(), id, fields); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			notFound(w)
			return
		}
		internalErr(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Delete godoc
// @Summary      Delete a transaction
// @Description  Deletes a transaction by its numeric ID.
// @Tags         transactions
// @Param        id  path  int  true  "Transaction ID"
// @Success      204
// @Failure      404  {object}  HTTPError "transaction not found"
// @Failure      500  {object}  HTTPError
// @Router       /api/transactions/{id} [delete]
// @Security     BearerAuth
func (h *TransactionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err := h.Store.DeleteTransaction(r.Context(), id); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			notFound(w)
			return
		}
		internalErr(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
