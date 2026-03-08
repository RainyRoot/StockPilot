package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rainyroot/stockpilot/backend/pkg/httputil"
)

func NewRouter(frontendURL string, stockHandler *StockHandler, watchlistHandler *WatchlistHandler) http.Handler {
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

	r.Route("/api/v1/watchlists", func(r chi.Router) {
		r.Get("/", watchlistHandler.List)
		r.Post("/", watchlistHandler.Create)
		r.Get("/{id}", watchlistHandler.Get)
		r.Put("/{id}", watchlistHandler.Update)
		r.Delete("/{id}", watchlistHandler.Delete)
		r.Post("/{id}/stocks", watchlistHandler.AddStock)
		r.Delete("/{id}/stocks", watchlistHandler.RemoveStock)
	})

	return r
}
