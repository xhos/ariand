package handlers

import (
	"ariand/internal/db"
	_ "ariand/internal/domain"
	"net/http"
	"time"
)

type DashboardHandler struct{ Store db.Store }

// DebtResponse wraps a single float64 for the total debt amount.
type DebtResponse struct {
	Debt float64 `json:"debt" example:"2550.75"`
}

// Balance godoc
// @Summary      Get total balance
// @Description  Calculates and returns the sum of current balances across all accounts.
// @Tags         dashboard
// @Produce      json
// @Success      200  {object}  BalanceResponse
// @Failure      500  {object}  HTTPError
// @Router       /api/dashboard/balance [get]
// @Security     BearerAuth
func (h *DashboardHandler) Balance(w http.ResponseWriter, r *http.Request) {
	val, err := h.Store.GetDashboardBalance(r.Context())
	if err != nil {
		internalErr(w)
		return
	}

	writeJSON(w, http.StatusOK, BalanceResponse{Balance: val})
}

// Debt godoc
// @Summary      Get total debt
// @Description  Calculates and returns the sum of current balances for all 'credit_card' type accounts.
// @Tags         dashboard
// @Produce      json
// @Success      200  {object}  DebtResponse
// @Failure      500  {object}  HTTPError
// @Router       /api/dashboard/debt [get]
// @Security     BearerAuth
func (h *DashboardHandler) Debt(w http.ResponseWriter, r *http.Request) {
	val, err := h.Store.GetDashboardDebt(r.Context())
	if err != nil {
		internalErr(w)
		return
	}

	writeJSON(w, http.StatusOK, DebtResponse{Debt: val})
}

// Trends godoc
// @Summary      Get income & expense trends
// @Description  Returns daily aggregated income and expense totals over a specified date range.
// @Tags         dashboard
// @Produce      json
// @Param        start  query     string  false  "Start date for trend data (YYYY-MM-DD)"
// @Param        end    query     string  false  "End date for trend data (YYYY-MM-DD)"
// @Success      200    {array}   domain.TrendPoint
// @Failure      400    {object}  HTTPError "invalid date format"
// @Failure      500    {object}  HTTPError
// @Router       /api/dashboard/trends [get]
// @Security     BearerAuth
func (h *DashboardHandler) Trends(w http.ResponseWriter, r *http.Request) {
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

	data, err := h.Store.GetDashboardTrends(r.Context(), opt)
	if err != nil {
		internalErr(w)
		return
	}

	writeJSON(w, http.StatusOK, data)
}
