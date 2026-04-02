package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/rainyroot/stockpilot/backend/internal/config"
	"github.com/rainyroot/stockpilot/backend/internal/handler"
	"github.com/rainyroot/stockpilot/backend/internal/repository/sqlite"
	"github.com/rainyroot/stockpilot/backend/internal/scraper"
	"github.com/rainyroot/stockpilot/backend/internal/service"
)

//go:embed static
var staticFiles embed.FS

func main() {
	cfg := config.Load()

	db, err := sqlite.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	stockRepo := sqlite.NewStockRepo(db)
	watchlistRepo := sqlite.NewWatchlistRepo(db)
	portfolioRepo := sqlite.NewPortfolioRepo(db)
	tradeRepo := sqlite.NewTradeRepo(db)
	allocationRepo := sqlite.NewAllocationRepo(db)
	settingsRepo := sqlite.NewSettingsRepo(db)
	yahooClient := scraper.NewYahooClient()

	stockService := service.NewStockService(stockRepo, yahooClient, cfg.CacheTTLSecs)
	watchlistService := service.NewWatchlistService(watchlistRepo, stockRepo, yahooClient, cfg.CacheTTLSecs)
	portfolioService := service.NewPortfolioService(portfolioRepo, tradeRepo, stockRepo, yahooClient)
	dashboardService := service.NewDashboardService(portfolioRepo, tradeRepo, stockRepo, yahooClient, portfolioService)
	valuationService := service.NewValuationService(stockRepo, watchlistRepo, yahooClient, settingsRepo)
	allocationService := service.NewAllocationService(allocationRepo, portfolioRepo, tradeRepo, stockRepo, yahooClient)
	settingsService := service.NewSettingsService(settingsRepo)
	healthService := service.NewHealthService(portfolioRepo, tradeRepo, stockRepo, allocationRepo, watchlistRepo, yahooClient, valuationService, allocationService, portfolioService, settingsRepo)
	exportService := service.NewExportService(portfolioService, tradeRepo)

	stockHandler := handler.NewStockHandler(stockService)
	watchlistHandler := handler.NewWatchlistHandler(watchlistService)
	portfolioHandler := handler.NewPortfolioHandler(portfolioService)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)
	valuationHandler := handler.NewValuationHandler(valuationService)
	allocationHandler := handler.NewAllocationHandler(allocationService)
	healthHandler := handler.NewHealthHandler(healthService)
	settingsHandler := handler.NewSettingsHandler(settingsService)
	exportHandler := handler.NewExportHandler(exportService)

	frontendFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("failed to load embedded frontend: %v", err)
	}

	router := handler.NewRouter(frontendFS, cfg.FrontendURL, stockHandler, watchlistHandler, portfolioHandler, dashboardHandler, valuationHandler, allocationHandler, healthHandler, settingsHandler, exportHandler)

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("StockPilot API server starting on %s", addr)
	log.Printf("Frontend URL: %s", cfg.FrontendURL)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
