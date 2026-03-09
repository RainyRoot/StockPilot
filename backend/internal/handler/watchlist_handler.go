package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rainyroot/stockpilot/backend/internal/domain"
	"github.com/rainyroot/stockpilot/backend/internal/service"
	"github.com/rainyroot/stockpilot/backend/pkg/httputil"
)

type WatchlistHandler struct {
	watchlistService *service.WatchlistService
}

func NewWatchlistHandler(watchlistService *service.WatchlistService) *WatchlistHandler {
	return &WatchlistHandler{watchlistService: watchlistService}
}

func (h *WatchlistHandler) List(w http.ResponseWriter, r *http.Request) {
	watchlists, err := h.watchlistService.List(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	if watchlists == nil {
		watchlists = []domain.Watchlist{}
	}
	httputil.JSON(w, http.StatusOK, watchlists)
}

func (h *WatchlistHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid watchlist id")
		return
	}

	watchlist, err := h.watchlistService.Get(r.Context(), id)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, watchlist)
}

type createWatchlistRequest struct {
	Name string `json:"name"`
}

func (h *WatchlistHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createWatchlistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	watchlist, err := h.watchlistService.Create(r.Context(), req.Name)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.JSON(w, http.StatusCreated, watchlist)
}

type updateWatchlistRequest struct {
	Name string `json:"name"`
}

func (h *WatchlistHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid watchlist id")
		return
	}

	var req updateWatchlistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if err := h.watchlistService.Update(r.Context(), id, req.Name); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *WatchlistHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid watchlist id")
		return
	}

	if err := h.watchlistService.Delete(r.Context(), id); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

type addStockRequest struct {
	Ticker string `json:"ticker"`
	Notes  string `json:"notes"`
}

func (h *WatchlistHandler) AddStock(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid watchlist id")
		return
	}

	var req addStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}
	if req.Ticker == "" {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "ticker is required")
		return
	}

	if err := h.watchlistService.AddStock(r.Context(), id, req.Ticker, req.Notes); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.JSON(w, http.StatusCreated, map[string]string{"status": "added"})
}

type removeStockRequest struct {
	Ticker string `json:"ticker"`
}

func (h *WatchlistHandler) RemoveStock(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid watchlist id")
		return
	}

	var req removeStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}
	if req.Ticker == "" {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "ticker is required")
		return
	}

	if err := h.watchlistService.RemoveStock(r.Context(), id, req.Ticker); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]string{"status": "removed"})
}
