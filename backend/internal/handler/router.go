package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rainyroot/stockpilot/backend/pkg/httputil"
)

func NewRouter(frontendURL string, stockHandler *StockHandler, watchlistHandler *WatchlistHandler, portfolioHandler *PortfolioHandler) http.Handler {
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

	r.Route("/api/v1/portfolios", func(r chi.Router) {
		r.Get("/", portfolioHandler.List)
		r.Post("/", portfolioHandler.Create)
		r.Get("/{id}", portfolioHandler.Get)
		r.Put("/{id}", portfolioHandler.Update)
		r.Delete("/{id}", portfolioHandler.Delete)
		r.Post("/{id}/trades", portfolioHandler.AddTrade)
		r.Get("/{id}/trades", portfolioHandler.ListTrades)
		r.Delete("/{id}/trades/{tradeID}", portfolioHandler.DeleteTrade)
		r.Get("/{id}/holdings", portfolioHandler.GetHoldings)
	})

	return r
}
