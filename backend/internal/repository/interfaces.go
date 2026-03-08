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

type WatchlistRepo interface {
	List(ctx context.Context) ([]domain.Watchlist, error)
	GetByID(ctx context.Context, id int64) (*domain.Watchlist, error)
	Create(ctx context.Context, name string) (*domain.Watchlist, error)
	Update(ctx context.Context, id int64, name string) error
	Delete(ctx context.Context, id int64) error
	AddItem(ctx context.Context, watchlistID int64, stockID int64, notes string) error
	RemoveItem(ctx context.Context, watchlistID int64, stockID int64) error
	GetItems(ctx context.Context, watchlistID int64) ([]domain.WatchlistItem, error)
}
