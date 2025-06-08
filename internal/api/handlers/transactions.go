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

/* ---------- LIST ---------- */
func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	var opt db.ListOpts
	q := r.URL.Query()
	if s := q.Get("start"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			opt.Start = &t
		}
	}
	if e := q.Get("end"); e != "" {
		if t, err := time.Parse("2006-01-02", e); err == nil {
			opt.End = &t
		}
	}

	out, err := h.Store.ListTransactions(r.Context(), opt)
	if err != nil {
		internalErr(w)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

/* ---------- GET ---------- */
func (h *TransactionHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	txn, err := h.Store.GetTransaction(r.Context(), id)
	if err != nil {
		internalErr(w)
		return
	}
	writeJSON(w, http.StatusOK, txn)
}

/* ---------- CREATE ---------- */
func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in domain.Transaction
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		badRequest(w, "invalid json")
		return
	}

	id, err := h.Store.CreateTransaction(r.Context(), &in)
	switch {
	case err == nil:
		writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
	case errors.Is(err, db.ErrConflict):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "duplicate"})
	default:
		internalErr(w)
	}
}

/* ---------- PATCH ---------- */
func (h *TransactionHandler) Patch(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	var fields map[string]any
	if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
		badRequest(w, "invalid json")
		return
	}

	if err := h.Store.UpdateTransaction(r.Context(), id, fields); err != nil {
		internalErr(w)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

/* ---------- DELETE ---------- */
func (h *TransactionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err := h.Store.DeleteTransaction(r.Context(), id); err != nil {
		internalErr(w)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
