package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/rainyroot/stockpilot/backend/internal/domain"
)

type StockRepo struct {
	db *sql.DB
}

func NewStockRepo(db *sql.DB) *StockRepo {
	return &StockRepo{db: db}
}

func (r *StockRepo) GetByTicker(ctx context.Context, ticker string) (*domain.Stock, error) {
	var s domain.Stock
	var createdAt, updatedAt int64
	var sector, industry sql.NullString
	var marketCap sql.NullInt64

	err := r.db.QueryRowContext(ctx,
		"SELECT id, ticker, name, exchange, currency, sector, industry, market_cap_cents, created_at, updated_at FROM stocks WHERE ticker = ?",
		ticker,
	).Scan(&s.ID, &s.Ticker, &s.Name, &s.Exchange, &s.Currency, &sector, &industry, &marketCap, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get stock by ticker: %w", err)
	}

	s.Sector = sector.String
	s.Industry = industry.String
	s.MarketCapCents = marketCap.Int64
	s.CreatedAt = time.Unix(createdAt, 0)
	s.UpdatedAt = time.Unix(updatedAt, 0)

	return &s, nil
}

func (r *StockRepo) Upsert(ctx context.Context, stock *domain.Stock) error {
	now := time.Now().Unix()

	result, err := r.db.ExecContext(ctx,
		`INSERT INTO stocks (ticker, name, exchange, currency, sector, industry, market_cap_cents, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(ticker) DO UPDATE SET
			name = excluded.name,
			exchange = excluded.exchange,
			currency = excluded.currency,
			sector = COALESCE(excluded.sector, stocks.sector),
			industry = COALESCE(excluded.industry, stocks.industry),
			market_cap_cents = COALESCE(excluded.market_cap_cents, stocks.market_cap_cents),
			updated_at = excluded.updated_at`,
		stock.Ticker, stock.Name, stock.Exchange, stock.Currency,
		nullString(stock.Sector), nullString(stock.Industry),
		nullInt64(stock.MarketCapCents), now, now,
	)
	if err != nil {
		return fmt.Errorf("upsert stock: %w", err)
	}

	if stock.ID == 0 {
		id, _ := result.LastInsertId()
		if id > 0 {
			stock.ID = id
		} else {
			existing, err := r.GetByTicker(ctx, stock.Ticker)
			if err != nil {
				return err
			}
			if existing != nil {
				stock.ID = existing.ID
			}
		}
	}

	return nil
}

func (r *StockRepo) CacheQuote(ctx context.Context, stockID int64, quote *domain.Quote) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO price_cache (stock_id, price_cents, open_cents, high_cents, low_cents, volume, change_percent, fetched_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		stockID, quote.PriceCents, quote.OpenCents, quote.HighCents, quote.LowCents,
		quote.Volume, quote.ChangePercent, quote.FetchedAt.Unix(),
	)
	if err != nil {
		return fmt.Errorf("cache quote: %w", err)
	}
	return nil
}

func (r *StockRepo) GetCachedQuote(ctx context.Context, stockID int64, maxAgeSecs int) (*domain.Quote, error) {
	minTime := time.Now().Unix() - int64(maxAgeSecs)

	var q domain.Quote
	var fetchedAt int64
	var changePercent sql.NullFloat64

	err := r.db.QueryRowContext(ctx,
		`SELECT price_cents, open_cents, high_cents, low_cents, volume, change_percent, fetched_at
		 FROM price_cache
		 WHERE stock_id = ? AND fetched_at > ?
		 ORDER BY fetched_at DESC LIMIT 1`,
		stockID, minTime,
	).Scan(&q.PriceCents, &q.OpenCents, &q.HighCents, &q.LowCents, &q.Volume, &changePercent, &fetchedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get cached quote: %w", err)
	}

	q.ChangePercent = changePercent.Float64
	q.FetchedAt = time.Unix(fetchedAt, 0)

	return &q, nil
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullInt64(i int64) sql.NullInt64 {
	if i == 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: i, Valid: true}
}
