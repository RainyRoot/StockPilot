package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rainyroot/stockpilot/backend/internal/service"
	"github.com/rainyroot/stockpilot/backend/pkg/httputil"
)

type DashboardHandler struct {
	dashboardService *service.DashboardService
}

func NewDashboardHandler(dashboardService *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardService: dashboardService}
}

// GetDashboard returns aggregated data across all portfolios.
func (h *DashboardHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	data, err := h.dashboardService.GetDashboard(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, data)
}

// GetSummary returns the summary for a single portfolio.
func (h *DashboardHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid portfolio id")
		return
	}

	summary, err := h.dashboardService.GetSummary(r.Context(), id)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, summary)
}

// GetHoldingsPL returns per-stock P&L breakdown.
func (h *DashboardHandler) GetHoldingsPL(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid portfolio id")
		return
	}

	pl, err := h.dashboardService.GetHoldingsPL(r.Context(), id)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, pl)
}

// GetPerformance returns portfolio value time series.
func (h *DashboardHandler) GetPerformance(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid portfolio id")
		return
	}

	rangeStr := r.URL.Query().Get("range")
	if rangeStr == "" {
		rangeStr = "3m"
	}

	snapshots, err := h.dashboardService.GetPerformanceHistory(r.Context(), id, rangeStr)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, snapshots)
}
