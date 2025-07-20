package handlers

import (
	"ariand/internal/db"
	"ariand/internal/domain"
	"ariand/internal/service"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
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
	NextCursor   *Cursor              `json:"nextCursor,omitempty"`
}

type TransactionHandler struct {
	Store       db.Store
	Categorizer *service.Categorizer
}

type CreateTransactionResponse struct {
	ID int64 `json:"id" example:"101"`
}

type UpdateTransactionRequest map[string]any

// bindlistoptions parses query parameters into a db.listopts struct
func bindListOptions(q url.Values) (db.ListOpts, error) {
	opts := db.ListOpts{}

	// pagination
	opts.Limit = 25
	if limitStr := q.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			opts.Limit = limit
		}
	}
	if cursorDateStr := q.Get("cursor_date"); cursorDateStr != "" {
		if t, err := time.Parse(time.RFC3339Nano, cursorDateStr); err == nil {
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
			return opts, fmt.Errorf("invalid start_date format, expected yyyy-mm-dd")
		}
		opts.Start = &t
	}
	if e := q.Get("end_date"); e != "" {
		t, err := time.Parse("2006-01-02", e)
		if err != nil {
			return opts, fmt.Errorf("invalid end_date format, expected yyyy-mm-dd")
		}
		opts.End = &t
	}
	if min := q.Get("amount_min"); min != "" {
		v, err := strconv.ParseFloat(min, 64)
		if err != nil {
			return opts, fmt.Errorf("invalid amount_min")
		}
		opts.AmountMin = &v
	}
	if max := q.Get("amount_max"); max != "" {
		v, err := strconv.ParseFloat(max, 64)
		if err != nil {
			return opts, fmt.Errorf("invalid amount_max")
		}
		opts.AmountMax = &v
	}
	if cats := q.Get("categories"); cats != "" {
		opts.Categories = strings.Split(cats, ",")
	}
	if accountIDsStr := q.Get("account_ids"); accountIDsStr != "" {
		idStrs := strings.Split(accountIDsStr, ",")
		ids := make([]int64, 0, len(idStrs))
		for _, idStr := range idStrs {
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return opts, fmt.Errorf("invalid account_id in list: %q", idStr)
			}
			ids = append(ids, id)
		}
		opts.AccountIDs = ids
	}

	opts.MerchantSearch = q.Get("merchant")
	opts.DescriptionSearch = q.Get("description")
	opts.Currency = q.Get("currency")

	if dir := q.Get("direction"); dir == "in" || dir == "out" {
		opts.Direction = dir
	}

	return opts, nil
}

// List godoc
// @Summary      list transactions
// @Description  returns a paginated and filtered list of transactions
// @Tags         transactions
// @Produce      json
// @Param        limit        query int    false "page size" default(25)
// @Param        cursor_date  query string false "cursor date from previous page (rfc3339nano)"
// @Param        cursor_id    query int    false "cursor id from previous page"
// @Param        account_ids  query string false "comma-separated list of account ids"
// @Success      200          {object} ListTransactionsResponse
// @Failure      400          {object} ErrorResponse "invalid query parameter"
// @Failure      500          {object} ErrorResponse
// @Router       /api/transactions [get]
// @Security     BearerAuth
func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	opts, err := bindListOptions(r.URL.Query())
	if err != nil {
		badRequest(w, err.Error())
		return
	}

	queryLimit := opts.Limit
	opts.Limit = queryLimit + 1

	out, err := h.Store.ListTransactions(r.Context(), opts)
	if err != nil {
		internalErr(w)
		return
	}

	var nextCursor *Cursor
	if len(out) > queryLimit {
		lastTxn := out[queryLimit]
		nextCursor = &Cursor{
			ID:   lastTxn.ID,
			Date: lastTxn.TxDate,
		}
		out = out[:queryLimit]
	}

	writeJSON(w, http.StatusOK, ListTransactionsResponse{
		Transactions: out,
		NextCursor:   nextCursor,
	})
}

// Get godoc
// @Summary      get a single transaction
// @Description  retrieves a transaction by its numeric id
// @Tags         transactions
// @Produce      json
// @Param        id   path      int  true  "transaction id"
// @Success      200  {object}  domain.Transaction
// @Failure      400  {object}  ErrorResponse "invalid id format"
// @Failure      404  {object}  ErrorResponse "transaction not found"
// @Failure      500  {object}  ErrorResponse
// @Router       /api/transactions/{id} [get]
// @Security     BearerAuth
func (h *TransactionHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromRequest(r)
	if err != nil {
		badRequest(w, "invalid id format")
		return
	}

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
// @Summary      create a new transaction
// @Description  adds a new transaction to the database
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        transaction body     domain.Transaction        true "transaction object"
// @Success      201         {object} CreateTransactionResponse
// @Failure      400         {object} ErrorResponse "invalid request body"
// @Failure      409         {object} ErrorResponse "duplicate transaction"
// @Failure      500         {object} ErrorResponse
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
		writeJSON(w, http.StatusConflict, ErrorResponse{Code: http.StatusConflict, Message: "duplicate transaction"})
	default:
		internalErr(w)
	}
}

// Patch godoc
// @Summary      update a transaction
// @Description  partially updates a transaction's fields
// @Tags         transactions
// @Accept       json
// @Param        id      path      int                     true  "transaction id"
// @Param        fields  body      UpdateTransactionRequest  true  "fields to update"
// @Success      204
// @Failure      400     {object}  ErrorResponse "invalid request body or id format"
// @Failure      404     {object}  ErrorResponse "transaction not found"
// @Failure      500     {object}  ErrorResponse
// @Router       /api/transactions/{id} [patch]
// @Security     BearerAuth
func (h *TransactionHandler) Patch(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromRequest(r)
	if err != nil {
		badRequest(w, "invalid id format")
		return
	}

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
// @Summary      delete a transaction
// @Description  deletes a transaction by its numeric id
// @Tags         transactions
// @Param        id  path  int  true  "transaction id"
// @Success      204
// @Failure      400  {object}  ErrorResponse "invalid id format"
// @Failure      404  {object}  ErrorResponse "transaction not found"
// @Failure      500  {object}  ErrorResponse
// @Router       /api/transactions/{id} [delete]
// @Security     BearerAuth
func (h *TransactionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromRequest(r)
	if err != nil {
		badRequest(w, "invalid id format")
		return
	}

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
