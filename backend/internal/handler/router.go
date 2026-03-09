package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rainyroot/stockpilot/backend/pkg/httputil"
)

func NewRouter(frontendURL string, stockHandler *StockHandler, watchlistHandler *WatchlistHandler, portfolioHandler *PortfolioHandler, dashboardHandler *DashboardHandler, valuationHandler *ValuationHandler, allocationHandler *AllocationHandler, healthHandler *HealthHandler, settingsHandler *SettingsHandler, exportHandler *ExportHandler) http.Handler {
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
		r.Get("/{id}/screen", valuationHandler.ScreenWatchlist)
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
		r.Get("/{id}/allocation/targets", allocationHandler.GetTargets)
		r.Put("/{id}/allocation/targets", allocationHandler.SetTargets)
		r.Get("/{id}/allocation/actual", allocationHandler.GetActual)
		r.Get("/{id}/allocation/comparison", allocationHandler.GetComparison)
		r.Get("/{id}/allocation/rebalance", allocationHandler.GetRebalance)
		r.Get("/{id}/health", healthHandler.GetHealth)
		r.Get("/{id}/health/alerts", healthHandler.GetAlerts)
		r.Get("/{id}/health/top-picks", healthHandler.GetTopPicks)
		r.Get("/{id}/export/holdings.csv", exportHandler.ExportHoldings)
		r.Get("/{id}/export/trades.csv", exportHandler.ExportTrades)
	})

	r.Route("/api/v1/dashboard", func(r chi.Router) {
		r.Get("/", dashboardHandler.GetDashboard)
		r.Get("/portfolios/{id}/summary", dashboardHandler.GetSummary)
		r.Get("/portfolios/{id}/holdings-pl", dashboardHandler.GetHoldingsPL)
		r.Get("/portfolios/{id}/performance", dashboardHandler.GetPerformance)
	})

	r.Route("/api/v1/valuation", func(r chi.Router) {
		r.Get("/{ticker}", valuationHandler.GetValuation)
		r.Get("/{ticker}/fundamentals", valuationHandler.GetFundamentals)
	})

	r.Route("/api/v1/settings", func(r chi.Router) {
		r.Get("/", settingsHandler.Get)
		r.Put("/", settingsHandler.Update)
	})

	return r
}
