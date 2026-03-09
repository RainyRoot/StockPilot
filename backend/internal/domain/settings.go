package domain

type UserSettings struct {
	Currency       string  `json:"currency"`          // Default: "USD"
	DiscountRate   float64 `json:"discount_rate"`     // Default: 0.10
	GrowthRate     float64 `json:"growth_rate"`       // Default: 0.05
	TerminalGrowth float64 `json:"terminal_growth"`   // Default: 0.025
	MaxPositionPct float64 `json:"max_position_pct"`  // Default: 10.0
	Theme          string  `json:"theme"`             // "dark"
}
