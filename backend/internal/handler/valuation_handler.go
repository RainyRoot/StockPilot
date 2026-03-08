package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rainyroot/stockpilot/backend/internal/service"
	"github.com/rainyroot/stockpilot/backend/pkg/httputil"
)

type ValuationHandler struct {
	valuationService *service.ValuationService
}

func NewValuationHandler(valuationService *service.ValuationService) *ValuationHandler {
	return &ValuationHandler{valuationService: valuationService}
}

func (h *ValuationHandler) GetValuation(w http.ResponseWriter, r *http.Request) {
	ticker := chi.URLParam(r, "ticker")
	if ticker == "" {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "ticker is required")
		return
	}

	valuation, err := h.valuationService.GetValuation(r.Context(), ticker)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, valuation)
}

func (h *ValuationHandler) GetFundamentals(w http.ResponseWriter, r *http.Request) {
	ticker := chi.URLParam(r, "ticker")
	if ticker == "" {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "ticker is required")
		return
	}

	fundamentals, err := h.valuationService.GetFundamentals(r.Context(), ticker)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, fundamentals)
}
