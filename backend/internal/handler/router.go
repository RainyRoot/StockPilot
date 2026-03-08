package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rainyroot/stockpilot/backend/pkg/httputil"
)

func NewRouter(frontendURL string, stockHandler *StockHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(Logger)
	r.Use(CORS(frontendURL))

	r.Get("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		httputil.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/api/v1/stocks", func(r chi.Router) {
		r.Get("/search", stockHandler.Search)
		r.Get("/{ticker}", stockHandler.GetQuote)
		r.Get("/{ticker}/history", stockHandler.GetHistory)
	})

	return r
}
