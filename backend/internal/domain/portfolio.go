package domain

import "time"

type Portfolio struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Currency    string    `json:"currency"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Holding struct {
	Stock         Stock   `json:"stock"`
	Quantity      float64 `json:"quantity"`
	AvgCostCents  int64   `json:"avg_cost_cents"`
	CurrentCents  int64   `json:"current_price_cents"`
	MarketValue   int64   `json:"market_value_cents"`
	UnrealizedPL  int64   `json:"unrealized_pl_cents"`
	UnrealizedPct float64 `json:"unrealized_pl_percent"`
	Category      string  `json:"category,omitempty"`
}
