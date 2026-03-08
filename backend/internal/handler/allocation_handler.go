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

type AllocationHandler struct {
	allocationService *service.AllocationService
}

func NewAllocationHandler(allocationService *service.AllocationService) *AllocationHandler {
	return &AllocationHandler{allocationService: allocationService}
}

func (h *AllocationHandler) GetTargets(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid portfolio id")
		return
	}

	targets, err := h.allocationService.GetTargets(r.Context(), id)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	if targets == nil {
		targets = []domain.AllocationTarget{}
	}
	httputil.JSON(w, http.StatusOK, targets)
}

func (h *AllocationHandler) SetTargets(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid portfolio id")
		return
	}

	var req struct {
		Targets []domain.AllocationTarget `json:"targets"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if err := h.allocationService.SetTargets(r.Context(), id, req.Targets); err != nil {
		httputil.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AllocationHandler) GetActual(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid portfolio id")
		return
	}

	actual, err := h.allocationService.GetActual(r.Context(), id)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	if actual == nil {
		actual = []domain.AllocationActual{}
	}
	httputil.JSON(w, http.StatusOK, actual)
}

func (h *AllocationHandler) GetComparison(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid portfolio id")
		return
	}

	comparison, err := h.allocationService.GetComparison(r.Context(), id)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	if comparison == nil {
		comparison = []domain.AllocationComparison{}
	}
	httputil.JSON(w, http.StatusOK, comparison)
}

func (h *AllocationHandler) GetRebalance(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid portfolio id")
		return
	}

	actions, err := h.allocationService.GetRebalanceActions(r.Context(), id)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	if actions == nil {
		actions = []domain.RebalanceAction{}
	}
	httputil.JSON(w, http.StatusOK, actions)
}
