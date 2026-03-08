package domain

type AllocationTarget struct {
	ID            int64   `json:"id"`
	PortfolioID   int64   `json:"portfolio_id"`
	Category      string  `json:"category"`
	TargetPercent float64 `json:"target_percent"`
}

type AllocationActual struct {
	Category      string  `json:"category"`
	ActualPercent float64 `json:"actual_percent"`
	ValueCents    int64   `json:"value_cents"`
}

type AllocationComparison struct {
	Category      string  `json:"category"`
	TargetPercent float64 `json:"target_percent"`
	ActualPercent float64 `json:"actual_percent"`
	DiffPercent   float64 `json:"diff_percent"`
	DiffCents     int64   `json:"diff_cents"`
}

type RebalanceAction struct {
	Ticker     string  `json:"ticker"`
	Action     string  `json:"action"`
	Quantity   float64 `json:"quantity"`
	ValueCents int64   `json:"value_cents"`
	Reason     string  `json:"reason"`
}
