package domain

import "time"

// PortfolioSummary is the top-level overview of a portfolio.
type PortfolioSummary struct {
	PortfolioID     int64   `json:"portfolio_id"`
	PortfolioName   string  `json:"portfolio_name"`
	Currency        string  `json:"currency"`
	TotalValueCents int64   `json:"total_value_cents"`
	TotalCostCents  int64   `json:"total_cost_cents"`
	UnrealizedPL    int64   `json:"unrealized_pl_cents"`
	UnrealizedPct   float64 `json:"unrealized_pl_percent"`
	RealizedPL      int64   `json:"realized_pl_cents"`
	DayChangeCents  int64   `json:"day_change_cents"`
	DayChangePct    float64 `json:"day_change_percent"`
	HoldingsCount   int     `json:"holdings_count"`
}

// PortfolioSnapshot is a single data point in the portfolio value time series.
type PortfolioSnapshot struct {
	Date       string `json:"date"`       // YYYY-MM-DD
	ValueCents int64  `json:"value_cents"`
	CostCents  int64  `json:"cost_cents"`
}

// HoldingPL is the per-stock profit/loss breakdown.
type HoldingPL struct {
	Ticker        string  `json:"ticker"`
	Name          string  `json:"name"`
	UnrealizedPL  int64   `json:"unrealized_pl_cents"`
	UnrealizedPct float64 `json:"unrealized_pl_percent"`
	Weight        float64 `json:"weight"` // percentage of total portfolio value
}

// RealizedTrade represents a closed position's P&L (SELL matched against BUY).
type RealizedTrade struct {
	Ticker    string    `json:"ticker"`
	Quantity  float64   `json:"quantity"`
	BuyCents  int64     `json:"buy_price_cents"`
	SellCents int64     `json:"sell_price_cents"`
	PLCents   int64     `json:"pl_cents"`
	ClosedAt  time.Time `json:"closed_at"`
}

// DashboardData is the aggregated response for the main dashboard page.
type DashboardData struct {
	Summaries       []PortfolioSummary `json:"summaries"`
	TotalValueCents int64              `json:"total_value_cents"`
	TotalPLCents    int64              `json:"total_pl_cents"`
	TotalPLPct      float64            `json:"total_pl_percent"`
}
