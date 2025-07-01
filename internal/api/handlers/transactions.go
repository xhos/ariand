package handlers

import (
	"ariand/internal/db"
	"ariand/internal/domain"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Cursor struct {
	ID   int64     `json:"id"`
	Date time.Time `json:"date"`
}

type ListTransactionsResponse struct {
	Transactions []domain.Transaction `json:"transactions"`
	NextCursor   *Cursor              `json:"next_cursor"`
}

type TransactionHandler struct{ Store db.Store }

type CreateTransactionResponse struct {
	ID int64 `json:"id" example:"101"`
}

type UpdateTransactionRequest map[string]any

// List godoc
// @Summary      List transactions
// @Description  Returns a paginated and filtered list of transactions, ideal for infinite scrolling.
// @Tags         transactions
// @Produce      json
// @Param        limit         query int     false "Number of transactions to return per page" default(25)
// @Param        cursor_date   query string  false "Cursor date from the previous page (RFC3339)"
// @Param        cursor_id     query int     false "Cursor ID from the previous page"
// @Param        start_date    query string  false "Filter by start date (YYYY-MM-DD)"
// @Param        end_date      query string  false "Filter by end date (YYYY-MM-DD)"
// @Param        amount_min    query number  false "Filter by minimum transaction amount"
// @Param        amount_max    query number  false "Filter by maximum transaction amount"
// @Param        direction     query string  false "Filter by transaction direction ('in' or 'out')"
// @Param        currency      query string  false "Filter by a specific currency (e.g., 'USD')"
// @Param        categories    query string  false "Comma-separated list of categories to filter by"
// @Param        merchant      query string  false "Search term for the merchant field (case-insensitive)"
// @Param        description   query string  false "Search term for the description field (case-insensitive)"
// @Param        time_start    query string  false "Filter by start time of day (HH:MM:SS)"
// @Param        time_end      query string  false "Filter by end time of day (HH:MM:SS)"
// @Success      200           {object}  ListTransactionsResponse
// @Failure      400           {object}  HTTPError "invalid query parameter"
// @Failure      500           {object}  HTTPError
// @Router       /api/transactions [get]
// @Security     BearerAuth
func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	opts := db.ListOpts{}
	var err error

	// pagination
	opts.Limit = 25
	if limitStr := q.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			opts.Limit = limit
		}
	}
	if cursorDateStr := q.Get("cursor_date"); cursorDateStr != "" {
		if t, err := time.Parse(time.RFC3339, cursorDateStr); err == nil {
			opts.CursorDate = &t
		}
	}
	if cursorIDStr := q.Get("cursor_id"); cursorIDStr != "" {
		if id, err := strconv.ParseInt(cursorIDStr, 10, 64); err == nil {
			opts.CursorID = &id
		}
	}

	// filtering
	if s := q.Get("start_date"); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			badRequest(w, "invalid start_date format, expected YYYY-MM-DD")
			return
		}
		opts.Start = &t
	}
	if e := q.Get("end_date"); e != "" {
		t, err := time.Parse("2006-01-02", e)
		if err != nil {
			badRequest(w, "invalid end_date format, expected YYYY-MM-DD")
			return
		}
		opts.End = &t
	}
	if min := q.Get("amount_min"); min != "" {
		v, err := strconv.ParseFloat(min, 64)
		if err != nil {
			badRequest(w, "invalid amount_min")
			return
		}
		opts.AmountMin = &v
	}
	if max := q.Get("amount_max"); max != "" {
		v, err := strconv.ParseFloat(max, 64)
		if err != nil {
			badRequest(w, "invalid amount_max")
			return
		}
		opts.AmountMax = &v
	}
	if cats := q.Get("categories"); cats != "" {
		opts.Categories = strings.Split(cats, ",")
	}
	if m := q.Get("merchant"); m != "" {
		opts.MerchantSearch = &m
	}
	if d := q.Get("description"); d != "" {
		opts.DescriptionSearch = &d
	}
	if c := q.Get("currency"); c != "" {
		opts.Currency = &c
	}
	if ts := q.Get("time_start"); ts != "" {
		_, err := time.Parse("15:04:05", ts)
		if err != nil {
			badRequest(w, "invalid time_start format, expected HH:MM:SS")
			return
		}
		opts.TimeOfDayStart = &ts
	}
	if te := q.Get("time_end"); te != "" {
		_, err := time.Parse("15:04:05", te)
		if err != nil {
			badRequest(w, "invalid time_end format, expected HH:MM:SS")
			return
		}
		opts.TimeOfDayEnd = &te
	}
	if dir := q.Get("direction"); dir == "in" || dir == "out" {
		opts.Direction = dir
	}

	// fetch one more than the limit to check for a next page
	queryLimit := opts.Limit
	opts.Limit = queryLimit + 1

	out, err := h.Store.ListTransactions(r.Context(), opts)
	if err != nil {
		internalErr(w)
		return
	}

	// Determine the next cursor
	var nextCursor *Cursor
	if len(out) > queryLimit {
		// the extra item is the cursor for the next page
		lastTxn := out[queryLimit]
		nextCursor = &Cursor{
			ID:   lastTxn.ID,
			Date: lastTxn.TxDate,
		}
		// trim the extra item from the response
		out = out[:queryLimit]
	}

	writeJSON(w, http.StatusOK, ListTransactionsResponse{
		Transactions: out,
		NextCursor:   nextCursor,
	})
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
