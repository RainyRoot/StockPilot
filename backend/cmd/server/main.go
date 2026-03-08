package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/rainyroot/stockpilot/backend/internal/config"
	"github.com/rainyroot/stockpilot/backend/internal/handler"
	"github.com/rainyroot/stockpilot/backend/internal/repository/sqlite"
	"github.com/rainyroot/stockpilot/backend/internal/scraper"
	"github.com/rainyroot/stockpilot/backend/internal/service"
)

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
	yahooClient := scraper.NewYahooClient()
	stockService := service.NewStockService(stockRepo, yahooClient, cfg.CacheTTLSecs)
	watchlistService := service.NewWatchlistService(watchlistRepo, stockRepo, yahooClient, cfg.CacheTTLSecs)
	portfolioService := service.NewPortfolioService(portfolioRepo, tradeRepo, stockRepo, yahooClient)
	stockHandler := handler.NewStockHandler(stockService)
	watchlistHandler := handler.NewWatchlistHandler(watchlistService)
	portfolioHandler := handler.NewPortfolioHandler(portfolioService)
	router := handler.NewRouter(cfg.FrontendURL, stockHandler, watchlistHandler, portfolioHandler)

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("StockPilot API server starting on %s", addr)
	log.Printf("Frontend URL: %s", cfg.FrontendURL)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
