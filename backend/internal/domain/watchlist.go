package domain

import "time"

type Watchlist struct {
	ID        int64           `json:"id"`
	Name      string          `json:"name"`
	Items     []WatchlistItem `json:"items,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type WatchlistItem struct {
	ID          int64     `json:"id"`
	WatchlistID int64     `json:"watchlist_id"`
	StockID     int64     `json:"stock_id"`
	Stock       *Stock    `json:"stock,omitempty"`
	Quote       *Quote    `json:"quote,omitempty"`
	AddedAt     time.Time `json:"added_at"`
	Notes       string    `json:"notes,omitempty"`
}
