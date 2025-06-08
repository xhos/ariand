package handlers

import (
	"ariand/internal/db"
	"net/http"
	"time"
)

type DashboardHandler struct{ Store db.Store }

func (h *DashboardHandler) Balance(w http.ResponseWriter, r *http.Request) {
	val, err := h.Store.GetDashboardBalance(r.Context())
	if err != nil {
		internalErr(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]float64{"balance": val})
}

func (h *DashboardHandler) Debt(w http.ResponseWriter, r *http.Request) {
	val, err := h.Store.GetDashboardDebt(r.Context())
	if err != nil {
		internalErr(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]float64{"debt": val})
}

func (h *DashboardHandler) Trends(w http.ResponseWriter, r *http.Request) {
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
	data, err := h.Store.GetDashboardTrends(r.Context(), opt)
	if err != nil {
		internalErr(w)
		return
	}
	writeJSON(w, http.StatusOK, data)
}
