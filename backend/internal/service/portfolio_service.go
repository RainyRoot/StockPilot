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

type PortfolioService struct {
	portfolioRepo repository.PortfolioRepo
	tradeRepo     repository.TradeRepo
	stockRepo     repository.StockRepo
	provider      scraper.DataProvider
}

func NewPortfolioService(
	portfolioRepo repository.PortfolioRepo,
	tradeRepo repository.TradeRepo,
	stockRepo repository.StockRepo,
	provider scraper.DataProvider,
) *PortfolioService {
	return &PortfolioService{
		portfolioRepo: portfolioRepo,
		tradeRepo:     tradeRepo,
		stockRepo:     stockRepo,
		provider:      provider,
	}
}

func (s *PortfolioService) List(ctx context.Context) ([]domain.Portfolio, error) {
	return s.portfolioRepo.List(ctx)
}

func (s *PortfolioService) Get(ctx context.Context, id int64) (*domain.Portfolio, error) {
	p, err := s.portfolioRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, fmt.Errorf("portfolio %d not found", id)
	}
	return p, nil
}

func (s *PortfolioService) Create(ctx context.Context, name string, description string, currency string) (*domain.Portfolio, error) {
	if name == "" {
		name = "My Portfolio"
	}
	if currency == "" {
		currency = "EUR"
	}
	return s.portfolioRepo.Create(ctx, name, description, currency)
}

func (s *PortfolioService) Update(ctx context.Context, id int64, name string, description string) error {
	return s.portfolioRepo.Update(ctx, id, name, description)
}

func (s *PortfolioService) Delete(ctx context.Context, id int64) error {
	return s.portfolioRepo.Delete(ctx, id)
}

// AddTrade records a BUY or SELL. Ensures the stock exists in the DB first.
func (s *PortfolioService) AddTrade(ctx context.Context, portfolioID int64, ticker string, tradeType domain.TradeType, quantity float64, priceCents int64, feeCents int64, currency string, executedAt time.Time, notes string) (*domain.Trade, error) {
	// Validate portfolio exists
	p, err := s.portfolioRepo.GetByID(ctx, portfolioID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, fmt.Errorf("portfolio %d not found", portfolioID)
	}

	// Ensure stock exists in DB
	stock, err := s.stockRepo.GetByTicker(ctx, ticker)
	if err != nil {
		return nil, fmt.Errorf("get stock: %w", err)
	}
	if stock == nil {
		quote, err := s.provider.GetQuote(ctx, ticker)
		if err != nil {
			return nil, fmt.Errorf("fetch quote for %s: %w", ticker, err)
		}
		stock = &domain.Stock{
			Ticker:   quote.Ticker,
			Name:     quote.Name,
			Exchange: quote.Exchange,
			Currency: quote.Currency,
		}
		if err := s.stockRepo.Upsert(ctx, stock); err != nil {
			return nil, fmt.Errorf("upsert stock: %w", err)
		}
	}

	if currency == "" {
		currency = stock.Currency
	}

	trade := &domain.Trade{
		PortfolioID: portfolioID,
		StockID:     stock.ID,
		Ticker:      ticker,
		TradeType:   tradeType,
		Quantity:    quantity,
		PriceCents:  priceCents,
		FeeCents:    feeCents,
		Currency:    currency,
		ExecutedAt:  executedAt,
		Notes:       notes,
	}

	return s.tradeRepo.Create(ctx, trade)
}

func (s *PortfolioService) DeleteTrade(ctx context.Context, id int64) error {
	return s.tradeRepo.Delete(ctx, id)
}

func (s *PortfolioService) ListTrades(ctx context.Context, portfolioID int64, limit int, offset int) ([]domain.Trade, int, error) {
	return s.tradeRepo.ListByPortfolio(ctx, portfolioID, limit, offset)
}

// GetHoldings computes current holdings with live prices and P&L.
func (s *PortfolioService) GetHoldings(ctx context.Context, portfolioID int64) ([]domain.Holding, error) {
	portfolio, err := s.portfolioRepo.GetByID(ctx, portfolioID)
	if err != nil {
		return nil, err
	}
	if portfolio == nil {
		return nil, fmt.Errorf("portfolio %d not found", portfolioID)
	}

	rows, err := s.tradeRepo.GetHoldings(ctx, portfolioID)
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return []domain.Holding{}, nil
	}

	// Fetch batch quotes
	tickers := make([]string, len(rows))
	for i, h := range rows {
		tickers[i] = h.Ticker
	}
	quotes, err := s.provider.GetBatchQuotes(ctx, tickers)
	if err != nil {
		fmt.Printf("warning: failed to fetch batch quotes for holdings: %v\n", err)
	}
	quoteMap := make(map[string]*domain.Quote)
	for i := range quotes {
		quoteMap[quotes[i].Ticker] = &quotes[i]
	}

	// Fetch exchange rate if portfolio currency differs from stock currency
	rateCache := make(map[string]float64)

	holdings := make([]domain.Holding, 0, len(rows))
	for _, row := range rows {
		avgCost := int64(math.Round(float64(row.TotalCostCents) / row.NetQuantity))

		var currentCents int64
		if q, ok := quoteMap[row.Ticker]; ok {
			currentCents = q.PriceCents

			// Convert to portfolio currency if needed
			if row.Currency != portfolio.Currency {
				rateKey := row.Currency + "->" + portfolio.Currency
				rate, cached := rateCache[rateKey]
				if !cached {
					r, err := s.provider.GetExchangeRate(ctx, row.Currency, portfolio.Currency)
					if err != nil {
						fmt.Printf("warning: exchange rate %s: %v\n", rateKey, err)
						rate = 1.0
					} else {
						rate = r
					}
					rateCache[rateKey] = rate
				}
				currentCents = int64(math.Round(float64(currentCents) * rate))
				avgCost = int64(math.Round(float64(avgCost) * rate))
			}
		}

		marketValue := int64(math.Round(float64(currentCents) * row.NetQuantity))
		costBasis := int64(math.Round(float64(avgCost) * row.NetQuantity))
		unrealizedPL := marketValue - costBasis

		var unrealizedPct float64
		if costBasis != 0 {
			unrealizedPct = float64(unrealizedPL) / float64(costBasis)
		}

		holdings = append(holdings, domain.Holding{
			Stock: domain.Stock{
				ID:       row.StockID,
				Ticker:   row.Ticker,
				Name:     row.Name,
				Exchange: row.Exchange,
				Currency: row.Currency,
			},
			Quantity:      row.NetQuantity,
			AvgCostCents:  avgCost,
			CurrentCents:  currentCents,
			MarketValue:   marketValue,
			UnrealizedPL:  unrealizedPL,
			UnrealizedPct: unrealizedPct,
		})
	}

	return holdings, nil
}
