package domain

import "time"

type TradeType string

const (
	TradeTypeBuy  TradeType = "BUY"
	TradeTypeSell TradeType = "SELL"
)

type Trade struct {
	ID          int64     `json:"id"`
	PortfolioID int64     `json:"portfolio_id"`
	StockID     int64     `json:"stock_id"`
	Ticker      string    `json:"ticker,omitempty"`
	TradeType   TradeType `json:"trade_type"`
	Quantity    float64   `json:"quantity"`
	PriceCents  int64     `json:"price_cents"`
	FeeCents    int64     `json:"fee_cents"`
	Currency    string    `json:"currency"`
	ExecutedAt  time.Time `json:"executed_at"`
	Notes       string    `json:"notes,omitempty"`
	Bucket      string    `json:"bucket,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type BucketSummary struct {
	Bucket        string  `json:"bucket"`
	TargetPercent float64 `json:"target_percent"`
	InvestedCents int64   `json:"invested_cents"`
	CurrentCents  int64   `json:"current_cents"`
	PLCents       int64   `json:"pl_cents"`
	PLPercent     float64 `json:"pl_percent"`
	ActualPercent float64 `json:"actual_percent"`
	HoldingCount  int     `json:"holding_count"`
}

type StrategyOverview struct {
	Buckets       []BucketSummary `json:"buckets"`
	TotalInvested int64           `json:"total_invested_cents"`
	TotalCurrent  int64           `json:"total_current_cents"`
	TotalPL       int64           `json:"total_pl_cents"`
	TotalPLPct    float64         `json:"total_pl_percent"`
}
