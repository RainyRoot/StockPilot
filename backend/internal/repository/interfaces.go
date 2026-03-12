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

type PortfolioRepo interface {
	List(ctx context.Context) ([]domain.Portfolio, error)
	GetByID(ctx context.Context, id int64) (*domain.Portfolio, error)
	Create(ctx context.Context, name string, description string, currency string, strategy string) (*domain.Portfolio, error)
	Update(ctx context.Context, id int64, name string, description string) error
	Delete(ctx context.Context, id int64) error
}

type TradeRepo interface {
	Create(ctx context.Context, trade *domain.Trade) (*domain.Trade, error)
	Delete(ctx context.Context, id int64) error
	UpdateBucket(ctx context.Context, id int64, bucket string) error
	ListByPortfolio(ctx context.Context, portfolioID int64, limit int, offset int) ([]domain.Trade, int, error)
	GetHoldings(ctx context.Context, portfolioID int64) ([]HoldingRow, error)
	GetHoldingsByBucket(ctx context.Context, portfolioID int64) ([]BucketHoldingRow, error)
}

type AllocationRepo interface {
	GetTargets(ctx context.Context, portfolioID int64) ([]domain.AllocationTarget, error)
	SetTargets(ctx context.Context, portfolioID int64, targets []domain.AllocationTarget) error
	DeleteTarget(ctx context.Context, id int64) error
}

// HoldingRow is the raw aggregation from the DB before enrichment with live prices.
type HoldingRow struct {
	StockID        int64
	Ticker         string
	Name           string
	Exchange       string
	Currency       string
	NetQuantity    float64
	TotalCostCents int64 // sum of (qty * price) for BUY minus SELL proceeds
	TotalFeeCents  int64
}

// BucketHoldingRow is like HoldingRow but grouped by bucket as well.
type BucketHoldingRow struct {
	HoldingRow
	Bucket string
}
