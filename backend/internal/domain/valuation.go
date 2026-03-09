package domain

type ValuationResult struct {
	Ticker         string           `json:"ticker"`
	CurrentCents   int64            `json:"current_price_cents"`
	IntrinsicCents int64            `json:"intrinsic_value_cents"`
	MarginOfSafety float64          `json:"margin_of_safety"`
	Verdict        ValuationVerdict `json:"verdict"`
	QualityScore   int              `json:"quality_score"`
	Methods        []MethodResult   `json:"methods"`
}

type ValuationVerdict string

const (
	VerdictUndervalued ValuationVerdict = "UNDERVALUED"
	VerdictFairLow     ValuationVerdict = "FAIRLY_VALUED_LOW"
	VerdictFair        ValuationVerdict = "FAIR_VALUE"
	VerdictOvervalued  ValuationVerdict = "OVERVALUED"
)

type MethodResult struct {
	Name        string  `json:"name"`
	ValueCents  int64   `json:"value_cents"`
	Weight      float64 `json:"weight"`
	Description string  `json:"description,omitempty"`
}

type Fundamentals struct {
	StockID           int64   `json:"stock_id"`
	FiscalYear        int     `json:"fiscal_year"`
	RevenueCents      int64   `json:"revenue_cents"`
	NetIncomeCents    int64   `json:"net_income_cents"`
	EPSCents          int64   `json:"eps_cents"`
	BookValueCents    int64   `json:"book_value_cents"`
	FreeCashFlowCents int64   `json:"free_cash_flow_cents"`
	DividendsCents    int64   `json:"dividends_cents"`
	SharesOutstanding int64   `json:"shares_outstanding"`
	DebtCents         int64   `json:"debt_cents"`
	CashCents         int64   `json:"cash_cents"`
	ROE               float64 `json:"roe"`
	ProfitMargin      float64 `json:"profit_margin"`
	// Historical net income (most recent first) for earnings growth calculation.
	NetIncomeHistory []int64 `json:"net_income_history,omitempty"`
}
