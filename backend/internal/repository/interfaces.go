package repository

import (
	"context"

	"github.com/rainyroot/stockpilot/backend/internal/domain"
)

type StockRepo interface {
	GetByTicker(ctx context.Context, ticker string) (*domain.Stock, error)
	Upsert(ctx context.Context, stock *domain.Stock) error
	CacheQuote(ctx context.Context, stockID int64, quote *domain.Quote) error
	GetCachedQuote(ctx context.Context, stockID int64, maxAgeSecs int) (*domain.Quote, error)
}
