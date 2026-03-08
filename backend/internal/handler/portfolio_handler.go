package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rainyroot/stockpilot/backend/internal/domain"
	"github.com/rainyroot/stockpilot/backend/internal/service"
	"github.com/rainyroot/stockpilot/backend/pkg/httputil"
)

type PortfolioHandler struct {
	portfolioService *service.PortfolioService
}

func NewPortfolioHandler(portfolioService *service.PortfolioService) *PortfolioHandler {
	return &PortfolioHandler{portfolioService: portfolioService}
}

func (h *PortfolioHandler) List(w http.ResponseWriter, r *http.Request) {
	portfolios, err := h.portfolioService.List(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, portfolios)
}

func (h *PortfolioHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid portfolio id")
		return
	}

	portfolio, err := h.portfolioService.Get(r.Context(), id)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, portfolio)
}

type createPortfolioRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Currency    string `json:"currency"`
}

func (h *PortfolioHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createPortfolioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	portfolio, err := h.portfolioService.Create(r.Context(), req.Name, req.Description, req.Currency)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.JSON(w, http.StatusCreated, portfolio)
}

type updatePortfolioRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *PortfolioHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid portfolio id")
		return
	}

	var req updatePortfolioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if err := h.portfolioService.Update(r.Context(), id, req.Name, req.Description); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *PortfolioHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid portfolio id")
		return
	}

	if err := h.portfolioService.Delete(r.Context(), id); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Trades ---

type addTradeRequest struct {
	Ticker     string  `json:"ticker"`
	TradeType  string  `json:"trade_type"`
	Quantity   float64 `json:"quantity"`
	PriceCents int64   `json:"price_cents"`
	FeeCents   int64   `json:"fee_cents"`
	Currency   string  `json:"currency"`
	ExecutedAt string  `json:"executed_at"` // ISO 8601
	Notes      string  `json:"notes"`
}

func (h *PortfolioHandler) AddTrade(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid portfolio id")
		return
	}

	var req addTradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}
	if req.Ticker == "" {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "ticker is required")
		return
	}
	if req.TradeType != "BUY" && req.TradeType != "SELL" {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "trade_type must be BUY or SELL")
		return
	}
	if req.Quantity <= 0 {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "quantity must be positive")
		return
	}
	if req.PriceCents <= 0 {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "price_cents must be positive")
		return
	}

	executedAt := time.Now()
	if req.ExecutedAt != "" {
		parsed, err := time.Parse(time.RFC3339, req.ExecutedAt)
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "executed_at must be RFC3339 format")
			return
		}
		executedAt = parsed
	}

	trade, err := h.portfolioService.AddTrade(
		r.Context(), id, req.Ticker,
		domain.TradeType(req.TradeType), req.Quantity,
		req.PriceCents, req.FeeCents, req.Currency,
		executedAt, req.Notes,
	)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.JSON(w, http.StatusCreated, trade)
}

func (h *PortfolioHandler) DeleteTrade(w http.ResponseWriter, r *http.Request) {
	tradeID, err := strconv.ParseInt(chi.URLParam(r, "tradeID"), 10, 64)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid trade id")
		return
	}

	if err := h.portfolioService.DeleteTrade(r.Context(), tradeID); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *PortfolioHandler) ListTrades(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid portfolio id")
		return
	}

	limit := 50
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	trades, total, err := h.portfolioService.ListTrades(r.Context(), id, limit, offset)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]interface{}{
		"trades": trades,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *PortfolioHandler) GetHoldings(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid portfolio id")
		return
	}

	holdings, err := h.portfolioService.GetHoldings(r.Context(), id)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, holdings)
}
