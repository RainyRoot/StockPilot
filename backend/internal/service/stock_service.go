package service

import (
	"context"
	"fmt"

	"github.com/rainyroot/stockpilot/backend/internal/domain"
	"github.com/rainyroot/stockpilot/backend/internal/repository"
	"github.com/rainyroot/stockpilot/backend/internal/scraper"
)

type StockService struct {
	repo     repository.StockRepo
	provider scraper.DataProvider
	cacheTTL int
}

func NewStockService(repo repository.StockRepo, provider scraper.DataProvider, cacheTTL int) *StockService {
	return &StockService{
		repo:     repo,
		provider: provider,
		cacheTTL: cacheTTL,
	}
}

func (s *StockService) GetQuote(ctx context.Context, ticker string) (*domain.Quote, *domain.Stock, error) {
	stock, err := s.repo.GetByTicker(ctx, ticker)
	if err != nil {
		return nil, nil, fmt.Errorf("get stock: %w", err)
	}

	if stock != nil {
		cached, err := s.repo.GetCachedQuote(ctx, stock.ID, s.cacheTTL)
		if err != nil {
			return nil, nil, fmt.Errorf("get cached quote: %w", err)
		}
		if cached != nil {
			cached.Ticker = stock.Ticker
			cached.Name = stock.Name
			cached.Currency = stock.Currency
			cached.Exchange = stock.Exchange
			cached.StockID = stock.ID
			return cached, stock, nil
		}
	}

	quote, err := s.provider.GetQuote(ctx, ticker)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch quote from provider: %w", err)
	}

	if stock == nil {
		stock = &domain.Stock{
			Ticker:   quote.Ticker,
			Name:     quote.Name,
			Exchange: quote.Exchange,
			Currency: quote.Currency,
		}
	}
	if err := s.repo.Upsert(ctx, stock); err != nil {
		return nil, nil, fmt.Errorf("upsert stock: %w", err)
	}

	quote.StockID = stock.ID
	if err := s.repo.CacheQuote(ctx, stock.ID, quote); err != nil {
		fmt.Printf("warning: failed to cache quote for %s: %v\n", ticker, err)
	}

	return quote, stock, nil
}

func (s *StockService) Search(ctx context.Context, query string) ([]domain.StockSearchResult, error) {
	results, err := s.provider.SearchStock(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("search stocks: %w", err)
	}
	return results, nil
}

func (s *StockService) GetHistory(ctx context.Context, ticker string, rangeStr string) ([]domain.PricePoint, error) {
	points, err := s.provider.GetHistory(ctx, ticker, rangeStr, "1d")
	if err != nil {
		return nil, fmt.Errorf("get history: %w", err)
	}
	return points, nil
}
