package handlers

import (
	"ariand/internal/db"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type AccountHandler struct{ Store db.Store }

func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	accts, err := h.Store.ListAccounts(r.Context())
	if err != nil {
		internalErr(w)
		return
	}
	writeJSON(w, http.StatusOK, accts)
}

func (h *AccountHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	acc, err := h.Store.GetAccount(r.Context(), id)
	if err != nil {
		internalErr(w)
		return
	}
	writeJSON(w, http.StatusOK, acc)
}

func (h *AccountHandler) SetAnchor(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)

	var in struct {
		Date    string  `json:"date"`
		Balance float64 `json:"balance"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		badRequest(w, "invalid json")
		return
	}

	dt, err := time.Parse("2006-01-02", in.Date)
	if err != nil {
		badRequest(w, "bad date")
		return
	}

	if err := h.Store.SetAccountAnchor(r.Context(), id, dt, in.Balance); err != nil {
		internalErr(w)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AccountHandler) Balance(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	bal, err := h.Store.GetAccountBalance(r.Context(), id)
	if err != nil {
		internalErr(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]float64{"balance": bal})
}
