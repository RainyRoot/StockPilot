package scraper

import (
	"context"

	"github.com/rainyroot/stockpilot/backend/internal/domain"
)

type DataProvider interface {
	GetQuote(ctx context.Context, ticker string) (*domain.Quote, error)
	GetBatchQuotes(ctx context.Context, tickers []string) ([]domain.Quote, error)
	SearchStock(ctx context.Context, query string) ([]domain.StockSearchResult, error)
	GetHistory(ctx context.Context, ticker string, rangeStr string, interval string) ([]domain.PricePoint, error)
	GetExchangeRate(ctx context.Context, from, to string) (float64, error)
	GetFundamentals(ctx context.Context, ticker string) (*domain.Fundamentals, error)
}
