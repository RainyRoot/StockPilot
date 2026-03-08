package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rainyroot/stockpilot/backend/internal/service"
	"github.com/rainyroot/stockpilot/backend/pkg/httputil"
)

type StockHandler struct {
	stockService *service.StockService
}

func NewStockHandler(stockService *service.StockService) *StockHandler {
	return &StockHandler{stockService: stockService}
}

func (h *StockHandler) GetQuote(w http.ResponseWriter, r *http.Request) {
	ticker := chi.URLParam(r, "ticker")
	if ticker == "" {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "ticker is required")
		return
	}

	quote, stock, err := h.stockService.GetQuote(r.Context(), ticker)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]interface{}{
		"stock": stock,
		"quote": quote,
	})
}

func (h *StockHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "query parameter 'q' is required")
		return
	}

	results, err := h.stockService.Search(r.Context(), query)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	httputil.JSON(w, http.StatusOK, results)
}

func (h *StockHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	ticker := chi.URLParam(r, "ticker")
	if ticker == "" {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "ticker is required")
		return
	}

	rangeStr := r.URL.Query().Get("range")
	if rangeStr == "" {
		rangeStr = "1y"
	}

	points, err := h.stockService.GetHistory(r.Context(), ticker, rangeStr)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	httputil.JSON(w, http.StatusOK, points)
}
