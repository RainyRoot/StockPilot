package domain

import "time"

type Stock struct {
	ID             int64     `json:"id"`
	Ticker         string    `json:"ticker"`
	Name           string    `json:"name"`
	Exchange       string    `json:"exchange"`
	Currency       string    `json:"currency"`
	Sector         string    `json:"sector,omitempty"`
	Industry       string    `json:"industry,omitempty"`
	MarketCapCents int64     `json:"market_cap_cents,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Quote struct {
	StockID       int64     `json:"stock_id"`
	Ticker        string    `json:"ticker"`
	Name          string    `json:"name"`
	PriceCents    int64     `json:"price_cents"`
	OpenCents     int64     `json:"open_cents,omitempty"`
	HighCents     int64     `json:"high_cents,omitempty"`
	LowCents      int64     `json:"low_cents,omitempty"`
	Volume        int64     `json:"volume,omitempty"`
	ChangePercent float64   `json:"change_percent"`
	Currency      string    `json:"currency"`
	Exchange      string    `json:"exchange"`
	FetchedAt     time.Time `json:"fetched_at"`
}

type PricePoint struct {
	Date          string `json:"date"`
	OpenCents     int64  `json:"open_cents"`
	HighCents     int64  `json:"high_cents"`
	LowCents      int64  `json:"low_cents"`
	CloseCents    int64  `json:"close_cents"`
	AdjCloseCents int64  `json:"adj_close_cents"`
	Volume        int64  `json:"volume"`
}

type StockSearchResult struct {
	Ticker   string `json:"ticker"`
	Name     string `json:"name"`
	Exchange string `json:"exchange"`
	Type     string `json:"type"`
}
