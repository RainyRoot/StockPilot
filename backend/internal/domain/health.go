package domain

// HealthScore represents the overall health assessment of a portfolio.
type HealthScore struct {
	PortfolioID  int64             `json:"portfolio_id"`
	OverallScore int               `json:"overall_score"` // 0-100
	Grade        string            `json:"grade"`         // A, B, C, D, F
	Components   []HealthComponent `json:"components"`
	Alerts       []HealthAlert     `json:"alerts"`
	TopPicks     []TopPick         `json:"top_picks"`
}

// HealthComponent represents a single scoring dimension.
type HealthComponent struct {
	Name        string  `json:"name"`
	Score       int     `json:"score"` // 0-100
	Weight      float64 `json:"weight"`
	Description string  `json:"description"`
}

// HealthAlert represents a warning or informational notice about the portfolio.
type HealthAlert struct {
	Severity string `json:"severity"` // "info", "warning", "critical"
	Title    string `json:"title"`
	Message  string `json:"message"`
	Ticker   string `json:"ticker,omitempty"`
}

// TopPick represents a stock recommendation.
type TopPick struct {
	Ticker         string  `json:"ticker"`
	Name           string  `json:"name"`
	Action         string  `json:"action"` // "BUY", "HOLD", "WATCH"
	Reason         string  `json:"reason"`
	MarginOfSafety float64 `json:"margin_of_safety"`
	QualityScore   int     `json:"quality_score"`
}
