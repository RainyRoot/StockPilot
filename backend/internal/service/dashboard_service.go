package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/rainyroot/stockpilot/backend/internal/domain"
	"github.com/rainyroot/stockpilot/backend/internal/repository"
	"github.com/rainyroot/stockpilot/backend/internal/scraper"
)

type DashboardService struct {
	portfolioRepo repository.PortfolioRepo
	tradeRepo     repository.TradeRepo
	stockRepo     repository.StockRepo
	provider      scraper.DataProvider
	portfolioSvc  *PortfolioService
}

func NewDashboardService(
	portfolioRepo repository.PortfolioRepo,
	tradeRepo repository.TradeRepo,
	stockRepo repository.StockRepo,
	provider scraper.DataProvider,
	portfolioSvc *PortfolioService,
) *DashboardService {
	return &DashboardService{
		portfolioRepo: portfolioRepo,
		tradeRepo:     tradeRepo,
		stockRepo:     stockRepo,
		provider:      provider,
		portfolioSvc:  portfolioSvc,
	}
}

// GetSummary returns the top-level overview for a single portfolio.
func (s *DashboardService) GetSummary(ctx context.Context, portfolioID int64) (*domain.PortfolioSummary, error) {
	portfolio, err := s.portfolioRepo.GetByID(ctx, portfolioID)
	if err != nil {
		return nil, err
	}
	if portfolio == nil {
		return nil, fmt.Errorf("portfolio %d not found", portfolioID)
	}

	holdings, err := s.portfolioSvc.GetHoldings(ctx, portfolioID)
	if err != nil {
		return nil, err
	}

	var totalValue, totalCost, unrealizedPL, dayChange int64
	for _, h := range holdings {
		totalValue += h.MarketValue
		totalCost += int64(math.Round(float64(h.AvgCostCents) * h.Quantity))
		unrealizedPL += h.UnrealizedPL
	}

	// Fetch day change from quotes
	if len(holdings) > 0 {
		tickers := make([]string, len(holdings))
		for i, h := range holdings {
			tickers[i] = h.Stock.Ticker
		}
		quotes, err := s.provider.GetBatchQuotes(ctx, tickers)
		if err == nil {
			quoteMap := make(map[string]*domain.Quote)
			for i := range quotes {
				quoteMap[quotes[i].Ticker] = &quotes[i]
			}
			for _, h := range holdings {
				if q, ok := quoteMap[h.Stock.Ticker]; ok {
					// day change = current value * change_percent
					change := int64(math.Round(float64(h.MarketValue) * q.ChangePercent / 100.0))
					dayChange += change
				}
			}
		}
	}

	var unrealizedPct, dayChangePct float64
	if totalCost > 0 {
		unrealizedPct = float64(unrealizedPL) / float64(totalCost)
	}
	if totalValue > 0 {
		dayChangePct = float64(dayChange) / float64(totalValue-dayChange)
	}

	return &domain.PortfolioSummary{
		PortfolioID:     portfolio.ID,
		PortfolioName:   portfolio.Name,
		Currency:        portfolio.Currency,
		TotalValueCents: totalValue,
		TotalCostCents:  totalCost,
		UnrealizedPL:    unrealizedPL,
		UnrealizedPct:   unrealizedPct,
		DayChangeCents:  dayChange,
		DayChangePct:    dayChangePct,
		HoldingsCount:   len(holdings),
	}, nil
}

// GetDashboard returns aggregated data across all portfolios for the main dashboard.
func (s *DashboardService) GetDashboard(ctx context.Context) (*domain.DashboardData, error) {
	portfolios, err := s.portfolioRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	data := &domain.DashboardData{}
	for _, p := range portfolios {
		summary, err := s.GetSummary(ctx, p.ID)
		if err != nil {
			fmt.Printf("warning: dashboard summary for portfolio %d: %v\n", p.ID, err)
			continue
		}
		data.Summaries = append(data.Summaries, *summary)
		data.TotalValueCents += summary.TotalValueCents
		data.TotalPLCents += summary.UnrealizedPL
	}
	if data.TotalValueCents > 0 {
		totalCost := data.TotalValueCents - data.TotalPLCents
		if totalCost > 0 {
			data.TotalPLPct = float64(data.TotalPLCents) / float64(totalCost)
		}
	}
	if data.Summaries == nil {
		data.Summaries = []domain.PortfolioSummary{}
	}
	return data, nil
}

// GetHoldingsPL returns per-stock P&L breakdown for a portfolio.
func (s *DashboardService) GetHoldingsPL(ctx context.Context, portfolioID int64) ([]domain.HoldingPL, error) {
	holdings, err := s.portfolioSvc.GetHoldings(ctx, portfolioID)
	if err != nil {
		return nil, err
	}

	var totalValue int64
	for _, h := range holdings {
		totalValue += h.MarketValue
	}

	result := make([]domain.HoldingPL, 0, len(holdings))
	for _, h := range holdings {
		var weight float64
		if totalValue > 0 {
			weight = float64(h.MarketValue) / float64(totalValue)
		}
		result = append(result, domain.HoldingPL{
			Ticker:        h.Stock.Ticker,
			Name:          h.Stock.Name,
			UnrealizedPL:  h.UnrealizedPL,
			UnrealizedPct: h.UnrealizedPct,
			Weight:        weight,
		})
	}
	return result, nil
}

// GetPerformanceHistory returns the portfolio value over time.
func (s *DashboardService) GetPerformanceHistory(ctx context.Context, portfolioID int64, rangeStr string) ([]domain.PortfolioSnapshot, error) {
	now := time.Now()
	var start time.Time
	switch rangeStr {
	case "1m":
		start = now.AddDate(0, -1, 0)
	case "3m":
		start = now.AddDate(0, -3, 0)
	case "6m":
		start = now.AddDate(0, -6, 0)
	case "ytd":
		start = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	case "1y":
		start = now.AddDate(-1, 0, 0)
	case "all":
		start = now.AddDate(-10, 0, 0)
	default:
		start = now.AddDate(0, -3, 0)
	}

	trades, _, err := s.tradeRepo.ListByPortfolio(ctx, portfolioID, 10000, 0)
	if err != nil {
		return nil, err
	}
	if len(trades) == 0 {
		return []domain.PortfolioSnapshot{}, nil
	}

	// Find the earliest trade date
	earliestTrade := trades[len(trades)-1].ExecutedAt // trades are DESC sorted
	if start.Before(earliestTrade) {
		start = earliestTrade
	}

	// Get current holdings to know which tickers to fetch history for
	holdings, err := s.portfolioSvc.GetHoldings(ctx, portfolioID)
	if err != nil {
		return nil, err
	}

	// Fetch price history for all held tickers
	priceHistories := make(map[string]map[string]int64) // ticker -> date -> close_cents
	for _, h := range holdings {
		history, err := s.provider.GetHistory(ctx, h.Stock.Ticker, rangeStr, "1d")
		if err != nil {
			fmt.Printf("warning: history for %s: %v\n", h.Stock.Ticker, err)
			continue
		}
		dateMap := make(map[string]int64)
		for _, pp := range history {
			dateMap[pp.Date] = pp.AdjCloseCents
		}
		priceHistories[h.Stock.Ticker] = dateMap
	}

	// Build snapshots: for each date, compute the portfolio value
	snapshots := []domain.PortfolioSnapshot{}
	for d := start; !d.After(now); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
			continue
		}

		// Compute holdings as of this date
		holdingsAtDate := make(map[string]float64) // ticker -> net qty
		costAtDate := make(map[string]int64)        // ticker -> total cost
		for _, t := range trades {
			if t.ExecutedAt.After(d.Add(24 * time.Hour)) {
				continue
			}
			switch t.TradeType {
			case domain.TradeTypeBuy:
				holdingsAtDate[t.Ticker] += t.Quantity
				costAtDate[t.Ticker] += int64(math.Round(t.Quantity * float64(t.PriceCents)))
			case domain.TradeTypeSell:
				holdingsAtDate[t.Ticker] -= t.Quantity
				costAtDate[t.Ticker] -= int64(math.Round(t.Quantity * float64(t.PriceCents)))
			}
		}

		var totalValue, totalCost int64
		hasPrice := false
		for ticker, qty := range holdingsAtDate {
			if qty <= 0 {
				continue
			}
			totalCost += costAtDate[ticker]
			if prices, ok := priceHistories[ticker]; ok {
				if price, ok := prices[dateStr]; ok {
					totalValue += int64(math.Round(float64(price) * qty))
					hasPrice = true
				}
			}
		}

		if hasPrice {
			snapshots = append(snapshots, domain.PortfolioSnapshot{
				Date:       dateStr,
				ValueCents: totalValue,
				CostCents:  totalCost,
			})
		}
	}

	return snapshots, nil
}
