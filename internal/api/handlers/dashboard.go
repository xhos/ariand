package handlers

import (
	"ariand/internal/db"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type DashboardHandler struct{ Store db.Store }

type DebtResponse struct {
	Debt float64 `json:"debt" example:"2550.75"`
}

// parseQueryDate is a helper to parse a date from a url query value
func parseQueryDate(q url.Values, key string) (*time.Time, error) {
	val := q.Get(key)
	if val == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", val)
	if err != nil {
		return nil, fmt.Errorf("invalid %s date format, expected YYYY-MM-DD", key)
	}
	return &t, nil
}

// Balance godoc
// @Summary      Get total balance
// @Description  Calculates and returns the sum of current balances across all accounts.
// @Tags         dashboard
// @Produce      json
// @Success      200  {object}  BalanceResponse
// @Failure      500  {object}  ErrorResponse
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
// @Failure      500  {object}  ErrorResponse
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
// @Failure      400    {object}  ErrorResponse "invalid date format"
// @Failure      500    {object}  ErrorResponse
// @Router       /api/dashboard/trends [get]
// @Security     BearerAuth
func (h *DashboardHandler) Trends(w http.ResponseWriter, r *http.Request) {
	var err error
	var opt db.ListOpts
	q := r.URL.Query()

	opt.Start, err = parseQueryDate(q, "start")
	if err != nil {
		badRequest(w, err.Error())
		return
	}

	opt.End, err = parseQueryDate(q, "end")
	if err != nil {
		badRequest(w, err.Error())
		return
	}

	data, err := h.Store.GetDashboardTrends(r.Context(), opt)
	if err != nil {
		internalErr(w)
		return
	}

	writeJSON(w, http.StatusOK, data)
}
