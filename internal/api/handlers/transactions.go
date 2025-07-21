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
	service service.TransactionService
}

type CreateTransactionResponse struct {
	ID int64 `json:"id" example:"101"`
}

type UpdateTransactionRequest map[string]any

// bindListOptions parses query parameters into a db.ListOpts struct
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
// @Param        start_date   query string false "filter by start date (yyyy-mm-dd)"
// @Param        end_date     query string false "filter by end date (yyyy-mm-dd)"
// @Param        amount_min   query number false "filter by minimum amount"
// @Param        amount_max   query number false "filter by maximum amount"
// @Param        categories   query string false "comma-separated list of category slugs"
// @Param        merchant     query string false "search merchant name (case-insensitive)"
// @Param        description  query string false "search description (case-insensitive)"
// @Param        currency     query string false "filter by currency code (e.g., CAD)"
// @Param        direction    query string false "filter by direction ('in' or 'out')"
// @Success      200          {object} ListTransactionsResponse
// @Failure      400          {object} ErrorResponse "invalid query parameter"
// @Failure      500          {object} ErrorResponse
// @Router       /api/transactions [get]
// @Security     BearerAuth
func (handler *TransactionHandler) List(r *http.Request) (any, *HTTPError) {
	opts, err := bindListOptions(r.URL.Query())
	if err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, err.Error())
	}

	queryLimit := opts.Limit
	opts.Limit = queryLimit + 1

	out, err := handler.service.List(r.Context(), opts)
	if err != nil {
		return nil, NewHTTPError(http.StatusInternalServerError, "internal server error")
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

	return ListTransactionsResponse{
		Transactions: out,
		NextCursor:   nextCursor,
	}, nil
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
func (handler *TransactionHandler) Get(r *http.Request) (any, *HTTPError) {
	id, err := parseIDFromRequest(r)
	if err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid id format")
	}

	txn, err := handler.service.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, NewHTTPError(http.StatusNotFound, "resource not found")
		}
		return nil, NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return txn, nil
}

// Create godoc
// @Summary      create a new transaction
// @Description  adds a new transaction to the database
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        transaction body     domain.Transaction      true "transaction object"
// @Success      201         {object} CreateTransactionResponse
// @Failure      400         {object} ErrorResponse "invalid request body"
// @Failure      409         {object} ErrorResponse "duplicate transaction"
// @Failure      500         {object} ErrorResponse
// @Router       /api/transactions [post]
// @Security     BearerAuth
func (handler *TransactionHandler) Create(r *http.Request) (any, *HTTPError) {
	var in domain.Transaction
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid json payload")
	}

	id, err := handler.service.Create(r.Context(), &in)
	if err != nil {
		if errors.Is(err, db.ErrConflict) {
			return nil, NewHTTPError(http.StatusConflict, "duplicate transaction")
		}
		return nil, NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return CreateTransactionResponse{ID: id}, nil
}

// Patch godoc
// @Summary      update a transaction
// @Description  partially updates a transaction's fields
// @Tags         transactions
// @Accept       json
// @Param        id     path      int                       true  "transaction id"
// @Param        fields body      UpdateTransactionRequest  true  "fields to update"
// @Success      204
// @Failure      400    {object}  ErrorResponse "invalid request body or id format"
// @Failure      404    {object}  ErrorResponse "transaction not found"
// @Failure      500    {object}  ErrorResponse
// @Router       /api/transactions/{id} [patch]
// @Security     BearerAuth
func (handler *TransactionHandler) Patch(r *http.Request) (any, *HTTPError) {
	id, err := parseIDFromRequest(r)
	if err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid id format")
	}

	var fields map[string]any
	if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid json payload")
	}

	if err := handler.service.Update(r.Context(), id, fields); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, NewHTTPError(http.StatusNotFound, "resource not found")
		}
		return nil, NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return nil, nil
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
func (handler *TransactionHandler) Delete(r *http.Request) (any, *HTTPError) {
	id, err := parseIDFromRequest(r)
	if err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid id format")
	}

	if err := handler.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, NewHTTPError(http.StatusNotFound, "resource not found")
		}
		return nil, NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return nil, nil
}

// Categorize godoc
// @Summary      categorize a transaction
// @Description  re-evaluates and updates the category of a transaction based on its description using AI
// @Tags         transactions
// @Param        id   path      int  true  "transaction id"
// @Success      204
// @Failure      400  {object}  ErrorResponse "invalid id format"
// @Failure      404  {object}  ErrorResponse "transaction not found"
// @Failure      500  {object}  ErrorResponse
// @Router       /api/transactions/{id}/categorize [post]
// @Security     BearerAuth
func (h *TransactionHandler) Categorize(r *http.Request) (any, *HTTPError) {
	id, err := parseIDFromRequest(r)
	if err != nil {
		return nil, NewHTTPError(400, "invalid id")
	}

	if err := h.service.CategorizeTransaction(r.Context(), id); err != nil {
		return nil, NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return nil, nil
}

// IdentifyMerchant godoc
// @Summary      identify merchant for a transaction
// @Description  extracts and updates the merchant name for a transaction based on its description using AI
// @Tags         transactions
// @Param        id   path      int  true  "transaction id"
// @Success      204
// @Failure      400  {object}  ErrorResponse "invalid id format"
// @Failure      404  {object}  ErrorResponse "transaction not found"
// @Failure      500  {object}  ErrorResponse
// @Router       /api/transactions/{id}/identify-merchant [post]
// @Security     BearerAuth
func (h *TransactionHandler) IdentifyMerchant(r *http.Request) (any, *HTTPError) {
	id, err := parseIDFromRequest(r)
	if err != nil {
		return nil, NewHTTPError(400, "invalid id")
	}

	if err := h.service.IdentifyMerchantForTransaction(r.Context(), id); err != nil {
		return nil, NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return nil, nil
}
