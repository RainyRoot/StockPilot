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
	CreatedAt   time.Time `json:"created_at"`
}
