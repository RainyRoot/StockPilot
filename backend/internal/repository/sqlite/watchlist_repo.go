package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/rainyroot/stockpilot/backend/internal/domain"
)

type WatchlistRepo struct {
	db *sql.DB
}

func NewWatchlistRepo(db *sql.DB) *WatchlistRepo {
	return &WatchlistRepo{db: db}
}

func (r *WatchlistRepo) List(ctx context.Context) ([]domain.Watchlist, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, name, created_at, updated_at FROM watchlists ORDER BY created_at ASC")
	if err != nil {
		return nil, fmt.Errorf("list watchlists: %w", err)
	}
	defer rows.Close()

	var watchlists []domain.Watchlist
	for rows.Next() {
		var w domain.Watchlist
		var createdAt, updatedAt int64
		if err := rows.Scan(&w.ID, &w.Name, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan watchlist: %w", err)
		}
		w.CreatedAt = time.Unix(createdAt, 0)
		w.UpdatedAt = time.Unix(updatedAt, 0)
		watchlists = append(watchlists, w)
	}
	return watchlists, nil
}

func (r *WatchlistRepo) GetByID(ctx context.Context, id int64) (*domain.Watchlist, error) {
	var w domain.Watchlist
	var createdAt, updatedAt int64
	err := r.db.QueryRowContext(ctx,
		"SELECT id, name, created_at, updated_at FROM watchlists WHERE id = ?", id,
	).Scan(&w.ID, &w.Name, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get watchlist: %w", err)
	}
	w.CreatedAt = time.Unix(createdAt, 0)
	w.UpdatedAt = time.Unix(updatedAt, 0)
	return &w, nil
}

func (r *WatchlistRepo) Create(ctx context.Context, name string) (*domain.Watchlist, error) {
	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx,
		"INSERT INTO watchlists (name, created_at, updated_at) VALUES (?, ?, ?)",
		name, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("create watchlist: %w", err)
	}
	id, _ := result.LastInsertId()
	return &domain.Watchlist{
		ID:        id,
		Name:      name,
		CreatedAt: time.Unix(now, 0),
		UpdatedAt: time.Unix(now, 0),
	}, nil
}

func (r *WatchlistRepo) Update(ctx context.Context, id int64, name string) error {
	now := time.Now().Unix()
	_, err := r.db.ExecContext(ctx,
		"UPDATE watchlists SET name = ?, updated_at = ? WHERE id = ?",
		name, now, id,
	)
	if err != nil {
		return fmt.Errorf("update watchlist: %w", err)
	}
	return nil
}

func (r *WatchlistRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM watchlists WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete watchlist: %w", err)
	}
	return nil
}

func (r *WatchlistRepo) AddItem(ctx context.Context, watchlistID int64, stockID int64, notes string) error {
	now := time.Now().Unix()
	_, err := r.db.ExecContext(ctx,
		"INSERT OR IGNORE INTO watchlist_items (watchlist_id, stock_id, added_at, notes) VALUES (?, ?, ?, ?)",
		watchlistID, stockID, now, nullString(notes),
	)
	if err != nil {
		return fmt.Errorf("add watchlist item: %w", err)
	}
	// Update watchlist timestamp
	r.db.ExecContext(ctx, "UPDATE watchlists SET updated_at = ? WHERE id = ?", now, watchlistID)
	return nil
}

func (r *WatchlistRepo) RemoveItem(ctx context.Context, watchlistID int64, stockID int64) error {
	now := time.Now().Unix()
	_, err := r.db.ExecContext(ctx,
		"DELETE FROM watchlist_items WHERE watchlist_id = ? AND stock_id = ?",
		watchlistID, stockID,
	)
	if err != nil {
		return fmt.Errorf("remove watchlist item: %w", err)
	}
	r.db.ExecContext(ctx, "UPDATE watchlists SET updated_at = ? WHERE id = ?", now, watchlistID)
	return nil
}

func (r *WatchlistRepo) GetItems(ctx context.Context, watchlistID int64) ([]domain.WatchlistItem, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT wi.id, wi.watchlist_id, wi.stock_id, wi.added_at, wi.notes,
		        s.id, s.ticker, s.name, s.exchange, s.currency, s.sector, s.industry, s.market_cap_cents, s.created_at, s.updated_at
		 FROM watchlist_items wi
		 JOIN stocks s ON s.id = wi.stock_id
		 WHERE wi.watchlist_id = ?
		 ORDER BY wi.added_at ASC`, watchlistID)
	if err != nil {
		return nil, fmt.Errorf("get watchlist items: %w", err)
	}
	defer rows.Close()

	var items []domain.WatchlistItem
	for rows.Next() {
		var item domain.WatchlistItem
		var stock domain.Stock
		var addedAt int64
		var notes sql.NullString
		var sector, industry sql.NullString
		var marketCap sql.NullInt64
		var sCreatedAt, sUpdatedAt int64

		if err := rows.Scan(
			&item.ID, &item.WatchlistID, &item.StockID, &addedAt, &notes,
			&stock.ID, &stock.Ticker, &stock.Name, &stock.Exchange, &stock.Currency,
			&sector, &industry, &marketCap, &sCreatedAt, &sUpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan watchlist item: %w", err)
		}

		item.AddedAt = time.Unix(addedAt, 0)
		item.Notes = notes.String
		stock.Sector = sector.String
		stock.Industry = industry.String
		stock.MarketCapCents = marketCap.Int64
		stock.CreatedAt = time.Unix(sCreatedAt, 0)
		stock.UpdatedAt = time.Unix(sUpdatedAt, 0)
		item.Stock = &stock
		items = append(items, item)
	}
	return items, nil
}
