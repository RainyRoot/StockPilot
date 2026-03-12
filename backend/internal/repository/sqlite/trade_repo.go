package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/rainyroot/stockpilot/backend/internal/domain"
	"github.com/rainyroot/stockpilot/backend/internal/repository"
)

type TradeRepo struct {
	db *sql.DB
}

func NewTradeRepo(db *sql.DB) *TradeRepo {
	return &TradeRepo{db: db}
}

func (r *TradeRepo) Create(ctx context.Context, trade *domain.Trade) (*domain.Trade, error) {
	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO trades (portfolio_id, stock_id, trade_type, quantity, price_cents, fee_cents, currency, executed_at, notes, bucket, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		trade.PortfolioID, trade.StockID, string(trade.TradeType),
		trade.Quantity, trade.PriceCents, trade.FeeCents, trade.Currency,
		trade.ExecutedAt.Unix(), nullString(trade.Notes), trade.Bucket, now,
	)
	if err != nil {
		return nil, fmt.Errorf("create trade: %w", err)
	}
	id, _ := result.LastInsertId()
	trade.ID = id
	trade.CreatedAt = time.Unix(now, 0)
	return trade, nil
}

func (r *TradeRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM trades WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete trade: %w", err)
	}
	return nil
}

func (r *TradeRepo) UpdateBucket(ctx context.Context, id int64, bucket string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE trades SET bucket = ? WHERE id = ?", bucket, id)
	if err != nil {
		return fmt.Errorf("update trade bucket: %w", err)
	}
	return nil
}

func (r *TradeRepo) ListByPortfolio(ctx context.Context, portfolioID int64, limit int, offset int) ([]domain.Trade, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM trades WHERE portfolio_id = ?", portfolioID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count trades: %w", err)
	}

	if limit <= 0 {
		limit = 50
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT t.id, t.portfolio_id, t.stock_id, s.ticker, t.trade_type, t.quantity,
		        t.price_cents, t.fee_cents, t.currency, t.executed_at, t.notes, t.bucket, t.created_at
		 FROM trades t
		 JOIN stocks s ON s.id = t.stock_id
		 WHERE t.portfolio_id = ?
		 ORDER BY t.executed_at DESC
		 LIMIT ? OFFSET ?`,
		portfolioID, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list trades: %w", err)
	}
	defer rows.Close()

	var trades []domain.Trade
	for rows.Next() {
		var t domain.Trade
		var notes sql.NullString
		var bucket sql.NullString
		var executedAt, createdAt int64
		if err := rows.Scan(
			&t.ID, &t.PortfolioID, &t.StockID, &t.Ticker, &t.TradeType,
			&t.Quantity, &t.PriceCents, &t.FeeCents, &t.Currency,
			&executedAt, &notes, &bucket, &createdAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan trade: %w", err)
		}
		t.ExecutedAt = time.Unix(executedAt, 0)
		t.Notes = notes.String
		t.Bucket = bucket.String
		t.CreatedAt = time.Unix(createdAt, 0)
		trades = append(trades, t)
	}
	return trades, total, nil
}

func (r *TradeRepo) GetHoldings(ctx context.Context, portfolioID int64) ([]repository.HoldingRow, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT
			s.id, s.ticker, s.name, s.exchange, s.currency,
			SUM(CASE WHEN t.trade_type = 'BUY' THEN t.quantity ELSE -t.quantity END) AS net_qty,
			SUM(CASE WHEN t.trade_type = 'BUY' THEN t.quantity * t.price_cents ELSE -t.quantity * t.price_cents END) AS total_cost,
			SUM(t.fee_cents) AS total_fees
		 FROM trades t
		 JOIN stocks s ON s.id = t.stock_id
		 WHERE t.portfolio_id = ?
		 GROUP BY s.id
		 HAVING net_qty > 0
		 ORDER BY s.ticker ASC`,
		portfolioID,
	)
	if err != nil {
		return nil, fmt.Errorf("get holdings: %w", err)
	}
	defer rows.Close()

	var holdings []repository.HoldingRow
	for rows.Next() {
		var h repository.HoldingRow
		if err := rows.Scan(
			&h.StockID, &h.Ticker, &h.Name, &h.Exchange, &h.Currency,
			&h.NetQuantity, &h.TotalCostCents, &h.TotalFeeCents,
		); err != nil {
			return nil, fmt.Errorf("scan holding: %w", err)
		}
		holdings = append(holdings, h)
	}
	return holdings, nil
}

func (r *TradeRepo) GetHoldingsByBucket(ctx context.Context, portfolioID int64) ([]repository.BucketHoldingRow, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT
			s.id, s.ticker, s.name, s.exchange, s.currency,
			COALESCE(t.bucket, '') AS bucket,
			SUM(CASE WHEN t.trade_type = 'BUY' THEN t.quantity ELSE -t.quantity END) AS net_qty,
			SUM(CASE WHEN t.trade_type = 'BUY' THEN t.quantity * t.price_cents ELSE -t.quantity * t.price_cents END) AS total_cost,
			SUM(t.fee_cents) AS total_fees
		 FROM trades t
		 JOIN stocks s ON s.id = t.stock_id
		 WHERE t.portfolio_id = ?
		 GROUP BY s.id, t.bucket
		 HAVING net_qty > 0
		 ORDER BY bucket, s.ticker ASC`,
		portfolioID,
	)
	if err != nil {
		return nil, fmt.Errorf("get holdings by bucket: %w", err)
	}
	defer rows.Close()

	var holdings []repository.BucketHoldingRow
	for rows.Next() {
		var h repository.BucketHoldingRow
		if err := rows.Scan(
			&h.StockID, &h.Ticker, &h.Name, &h.Exchange, &h.Currency,
			&h.Bucket, &h.NetQuantity, &h.TotalCostCents, &h.TotalFeeCents,
		); err != nil {
			return nil, fmt.Errorf("scan bucket holding: %w", err)
		}
		holdings = append(holdings, h)
	}
	return holdings, nil
}
