package service

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/rainyroot/stockpilot/backend/internal/domain"
	"github.com/rainyroot/stockpilot/backend/internal/repository"
	"github.com/rainyroot/stockpilot/backend/internal/scraper"
)

type ValuationService struct {
	stockRepo     repository.StockRepo
	watchlistRepo repository.WatchlistRepo
	provider      scraper.DataProvider
	settingsRepo  settingsGetter
}

// settingsGetter is a minimal interface to avoid depending on the concrete SettingsRepo.
type settingsGetter interface {
	Get(ctx context.Context) (*domain.UserSettings, error)
}

func NewValuationService(
	stockRepo repository.StockRepo,
	watchlistRepo repository.WatchlistRepo,
	provider scraper.DataProvider,
	settingsRepo settingsGetter,
) *ValuationService {
	return &ValuationService{
		stockRepo:     stockRepo,
		watchlistRepo: watchlistRepo,
		provider:      provider,
		settingsRepo:  settingsRepo,
	}
}

// loadSettings returns user settings, falling back to defaults on error.
func (s *ValuationService) loadSettings(ctx context.Context) *domain.UserSettings {
	if s.settingsRepo != nil {
		if settings, err := s.settingsRepo.Get(ctx); err == nil {
			return settings
		}
	}
	return &domain.UserSettings{
		DiscountRate:   0.10,
		GrowthRate:     0.05,
		TerminalGrowth: 0.025,
	}
}

// GetValuation computes the full valuation analysis for a single stock.
func (s *ValuationService) GetValuation(ctx context.Context, ticker string) (*domain.ValuationResult, error) {
	// Fetch quote and fundamentals in parallel
	type quoteResult struct {
		quote *domain.Quote
		err   error
	}
	type fundResult struct {
		fund *domain.Fundamentals
		err  error
	}
	qCh := make(chan quoteResult, 1)
	fCh := make(chan fundResult, 1)
	go func() {
		q, err := s.provider.GetQuote(ctx, ticker)
		qCh <- quoteResult{q, err}
	}()
	go func() {
		f, err := s.provider.GetFundamentals(ctx, ticker)
		fCh <- fundResult{f, err}
	}()
	qr := <-qCh
	fr := <-fCh
	if qr.err != nil {
		return nil, qr.err
	}
	if fr.err != nil {
		return nil, fr.err
	}
	quote := qr.quote
	fundamentals := fr.fund

	settings := s.loadSettings(ctx)
	methods := []domain.MethodResult{}

	if dcf := s.calcDCF(fundamentals, settings); dcf != nil {
		methods = append(methods, *dcf)
	}
	if graham := s.calcGrahamNumber(fundamentals); graham != nil {
		methods = append(methods, *graham)
	}
	if peg := s.calcPEGValuation(fundamentals, quote); peg != nil {
		methods = append(methods, *peg)
	}
	if owner := s.calcOwnerEarnings(fundamentals, settings); owner != nil {
		methods = append(methods, *owner)
	}

	var totalWeight float64
	var weightedValue float64
	for _, m := range methods {
		totalWeight += m.Weight
		weightedValue += float64(m.ValueCents) * m.Weight
	}
	var intrinsicCents int64
	if totalWeight > 0 {
		intrinsicCents = int64(math.Round(weightedValue / totalWeight))
	}

	var marginOfSafety float64
	if quote.PriceCents > 0 {
		marginOfSafety = float64(intrinsicCents-int64(quote.PriceCents)) / float64(intrinsicCents)
	}

	qualityScore := s.calcQualityScore(fundamentals)
	verdict := s.determineVerdict(marginOfSafety, qualityScore)

	trend := s.DetectValueTrap(fundamentals)
	if trend.IsValueTrap && verdict == domain.VerdictUndervalued {
		verdict = domain.VerdictFairLow
	}

	return &domain.ValuationResult{
		Ticker:         ticker,
		CurrentCents:   int64(quote.PriceCents),
		IntrinsicCents: intrinsicCents,
		MarginOfSafety: marginOfSafety,
		Verdict:        verdict,
		QualityScore:   qualityScore,
		Methods:        methods,
		Trend:          trend,
	}, nil
}

// GetFundamentals returns raw fundamental data for display.
func (s *ValuationService) GetFundamentals(ctx context.Context, ticker string) (*domain.Fundamentals, error) {
	return s.provider.GetFundamentals(ctx, ticker)
}

// calcDCF — Discounted Cash Flow (10-year projection)
func (s *ValuationService) calcDCF(f *domain.Fundamentals, settings *domain.UserSettings) *domain.MethodResult {
	if f.FreeCashFlowCents <= 0 || f.SharesOutstanding <= 0 {
		return nil
	}

	fcfPerShare := float64(f.FreeCashFlowCents) / float64(f.SharesOutstanding)
	discountRate := settings.DiscountRate
	growthRate := settings.GrowthRate
	terminalGrowth := settings.TerminalGrowth

	var pvSum float64
	currentFCF := fcfPerShare
	for year := 1; year <= 10; year++ {
		currentFCF *= (1 + growthRate)
		pv := currentFCF / math.Pow(1+discountRate, float64(year))
		pvSum += pv
	}

	terminalFCF := currentFCF * (1 + terminalGrowth)
	terminalValue := terminalFCF / (discountRate - terminalGrowth)
	pvTerminal := terminalValue / math.Pow(1+discountRate, 10)

	intrinsicPerShare := pvSum + pvTerminal
	intrinsicCents := int64(math.Round(intrinsicPerShare))

	return &domain.MethodResult{
		Name:        "DCF (Discounted Cash Flow)",
		ValueCents:  intrinsicCents,
		Weight:      0.35,
		Description: fmt.Sprintf("10-year FCF projection with %.1f%% growth, %.1f%% discount rate, %.1f%% terminal growth", growthRate*100, discountRate*100, terminalGrowth*100),
	}
}

// calcGrahamNumber — Benjamin Graham's intrinsic value formula
func (s *ValuationService) calcGrahamNumber(f *domain.Fundamentals) *domain.MethodResult {
	if f.EPSCents <= 0 || f.BookValueCents <= 0 || f.SharesOutstanding <= 0 {
		return nil
	}

	eps := float64(f.EPSCents) / 100.0
	bvps := float64(f.BookValueCents) / 100.0

	grahamValue := math.Sqrt(22.5 * eps * bvps)
	grahamCents := int64(math.Round(grahamValue * 100))

	return &domain.MethodResult{
		Name:        "Graham Number",
		ValueCents:  grahamCents,
		Weight:      0.25,
		Description: "sqrt(22.5 * EPS * Book Value Per Share)",
	}
}

// calcPEGValuation — Fair value based on PEG ratio
func (s *ValuationService) calcPEGValuation(f *domain.Fundamentals, q *domain.Quote) *domain.MethodResult {
	if f.EPSCents <= 0 || q.PriceCents <= 0 {
		return nil
	}

	growthRate := s.calcEarningsGrowth(f)

	eps := float64(f.EPSCents) / 100.0
	fairPE := growthRate
	fairValue := eps * fairPE
	fairCents := int64(math.Round(fairValue * 100))

	return &domain.MethodResult{
		Name:        "PEG Ratio Fair Value",
		ValueCents:  fairCents,
		Weight:      0.15,
		Description: fmt.Sprintf("PEG=1.0 with %.1f%% earnings growth (historical CAGR)", growthRate),
	}
}

// calcEarningsGrowth computes annualized earnings growth from historical net income.
// Falls back to ROE-based estimate if insufficient history.
func (s *ValuationService) calcEarningsGrowth(f *domain.Fundamentals) float64 {
	// Need at least 2 years of positive data for CAGR
	if len(f.NetIncomeHistory) >= 2 {
		newest := f.NetIncomeHistory[0]
		oldest := f.NetIncomeHistory[len(f.NetIncomeHistory)-1]
		years := float64(len(f.NetIncomeHistory) - 1)

		if oldest > 0 && newest > 0 && years > 0 {
			cagr := math.Pow(float64(newest)/float64(oldest), 1.0/years) - 1.0
			growthPct := cagr * 100
			if growthPct < 5 {
				growthPct = 5.0
			}
			if growthPct > 30 {
				growthPct = 30.0
			}
			return growthPct
		}
	}

	// Fallback: ROE-based estimate
	growthRate := f.ROE * 100
	if growthRate <= 0 {
		growthRate = 5.0
	}
	if growthRate > 30 {
		growthRate = 30.0
	}
	return growthRate
}

// calcOwnerEarnings — Buffett's Owner Earnings method
func (s *ValuationService) calcOwnerEarnings(f *domain.Fundamentals, settings *domain.UserSettings) *domain.MethodResult {
	if f.NetIncomeCents <= 0 || f.SharesOutstanding <= 0 {
		return nil
	}

	var ownerEarnings float64
	if f.FreeCashFlowCents > 0 {
		ownerEarnings = float64(f.FreeCashFlowCents) / float64(f.SharesOutstanding)
	} else {
		ownerEarnings = float64(f.NetIncomeCents) / float64(f.SharesOutstanding) * 0.8
	}

	requiredReturn := settings.DiscountRate
	intrinsicPerShare := ownerEarnings / requiredReturn
	intrinsicCents := int64(math.Round(intrinsicPerShare))

	return &domain.MethodResult{
		Name:        "Owner Earnings (Buffett)",
		ValueCents:  intrinsicCents,
		Weight:      0.25,
		Description: fmt.Sprintf("Owner earnings capitalized at %.1f%% required return", requiredReturn*100),
	}
}

// calcQualityScore — Moat indicators (0-100 points)
func (s *ValuationService) calcQualityScore(f *domain.Fundamentals) int {
	score := 0

	if f.ROE > 0.15 {
		score += 20
	} else if f.ROE > 0.10 {
		score += 10
	}

	if f.ProfitMargin > 0.20 {
		score += 20
	} else if f.ProfitMargin > 0.10 {
		score += 10
	}

	if f.DebtCents > 0 || f.CashCents > 0 {
		total := float64(f.DebtCents + f.CashCents)
		if total > 0 {
			debtRatio := float64(f.DebtCents) / total
			if debtRatio < 0.5 {
				score += 15
			} else if debtRatio < 0.7 {
				score += 8
			}
		}
	} else {
		score += 15
	}

	if f.FreeCashFlowCents > 0 {
		score += 15
	}

	if f.NetIncomeCents > 0 {
		score += 10
	}

	if f.RevenueCents > 0 {
		score += 5
	}

	if f.EPSCents > 0 {
		score += 10
	}

	if score > 100 {
		score = 100
	}

	return score
}

// determineVerdict — Valuation verdict based on margin of safety and quality
func (s *ValuationService) determineVerdict(marginOfSafety float64, qualityScore int) domain.ValuationVerdict {
	if marginOfSafety > 0.30 && qualityScore >= 60 {
		return domain.VerdictUndervalued
	}
	if marginOfSafety > 0.10 {
		return domain.VerdictFairLow
	}
	if marginOfSafety > -0.10 {
		return domain.VerdictFair
	}
	return domain.VerdictOvervalued
}

// DetectValueTrap analyzes fundamental trends to identify potential value traps.
func (s *ValuationService) DetectValueTrap(f *domain.Fundamentals) *domain.FundamentalTrend {
	trend := &domain.FundamentalTrend{
		Flags:   []domain.TrapFlag{},
		Reasons: []string{},
	}
	trapPoints := 0

	if len(f.RevenueHistory) >= 2 {
		cagr := calcCAGR(f.RevenueHistory)
		trend.RevenueCAGR = cagr
		if cagr < -0.03 {
			trapPoints += 25
			trend.Flags = append(trend.Flags, domain.TrapDecliningRevenue)
			trend.Reasons = append(trend.Reasons,
				fmt.Sprintf("Revenue declining at %.1f%% per year", cagr*100))
		}
	}

	if len(f.NetIncomeHistory) >= 2 {
		cagr := calcCAGR(f.NetIncomeHistory)
		trend.EarningsCAGR = cagr
		if cagr < -0.05 {
			trapPoints += 25
			trend.Flags = append(trend.Flags, domain.TrapDecliningEarnings)
			trend.Reasons = append(trend.Reasons,
				fmt.Sprintf("Earnings declining at %.1f%% per year", cagr*100))
		}
	}

	if len(f.MarginHistory) >= 2 {
		newest := f.MarginHistory[0]
		oldest := f.MarginHistory[len(f.MarginHistory)-1]
		if oldest > 0 && newest < oldest*0.8 {
			trapPoints += 20
			trend.Flags = append(trend.Flags, domain.TrapShrinkingMargins)
			trend.Reasons = append(trend.Reasons,
				fmt.Sprintf("Operating margin fell from %.1f%% to %.1f%%",
					oldest*100, newest*100))
		}
	}

	if len(f.DebtHistory) >= 2 {
		newest := f.DebtHistory[0]
		oldest := f.DebtHistory[len(f.DebtHistory)-1]
		if oldest > 0 && newest > 0 {
			growth := float64(newest-oldest) / float64(oldest)
			if growth > 0.30 {
				trapPoints += 15
				trend.Flags = append(trend.Flags, domain.TrapGrowingDebt)
				trend.Reasons = append(trend.Reasons,
					fmt.Sprintf("Long-term debt increased %.0f%% over %d years",
						growth*100, len(f.DebtHistory)-1))
			}
			trend.DebtTrend = trendLabel(growth)
		}
	}

	if len(f.FCFHistory) >= 2 {
		newest := f.FCFHistory[0]
		oldest := f.FCFHistory[len(f.FCFHistory)-1]
		if oldest > 0 && newest < oldest/2 {
			trapPoints += 15
			trend.Flags = append(trend.Flags, domain.TrapDecliningFCF)
			trend.Reasons = append(trend.Reasons, "Free cash flow deteriorating significantly")
		}
		if oldest > 0 {
			fcfGrowth := float64(newest-oldest) / float64(oldest)
			trend.FCFTrend = trendLabel(fcfGrowth)
		} else if newest > 0 {
			trend.FCFTrend = "improving"
		} else {
			trend.FCFTrend = "declining"
		}
	}

	trend.TrapScore = trapPoints
	if trend.TrapScore > 100 {
		trend.TrapScore = 100
	}
	trend.IsValueTrap = trapPoints >= 40

	return trend
}

func calcCAGR(values []int64) float64 {
	if len(values) < 2 {
		return 0
	}
	newest := float64(values[0])
	oldest := float64(values[len(values)-1])
	years := float64(len(values) - 1)
	if oldest <= 0 || newest <= 0 {
		return 0
	}
	return math.Pow(newest/oldest, 1.0/years) - 1.0
}

func trendLabel(change float64) string {
	if change > 0.10 {
		return "growing"
	}
	if change < -0.10 {
		return "declining"
	}
	return "stable"
}

// BacktestValuation runs a historical "what if" analysis.
func (s *ValuationService) BacktestValuation(ctx context.Context, ticker string, entryDateStr string) (*domain.BacktestResult, error) {
	entryDate, err := time.Parse("2006-01-02", entryDateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date format, use YYYY-MM-DD: %w", err)
	}

	daysSince := int(time.Since(entryDate).Hours() / 24)
	if daysSince < 1 {
		return nil, fmt.Errorf("entry date must be in the past")
	}

	rangeStr := "1y"
	if daysSince > 365*5 {
		rangeStr = "10y"
	} else if daysSince > 365*2 {
		rangeStr = "5y"
	} else if daysSince > 365 {
		rangeStr = "2y"
	}

	history, err := s.provider.GetHistory(ctx, ticker, rangeStr, "1d")
	if err != nil {
		return nil, fmt.Errorf("fetch history: %w", err)
	}

	entryPriceCents := findClosestPrice(history, entryDateStr)
	if entryPriceCents == 0 {
		return nil, fmt.Errorf("no price data found near %s (market may have been closed or date is too old)", entryDateStr)
	}

	quote, err := s.provider.GetQuote(ctx, ticker)
	if err != nil {
		return nil, fmt.Errorf("fetch current quote: %w", err)
	}
	currentCents := int64(quote.PriceCents)

	fundamentals, err := s.provider.GetFundamentals(ctx, ticker)
	if err != nil {
		return nil, fmt.Errorf("fetch fundamentals: %w", err)
	}

	settings := s.loadSettings(ctx)
	methods := []domain.MethodResult{}

	if dcf := s.calcDCF(fundamentals, settings); dcf != nil {
		methods = append(methods, *dcf)
	}
	if graham := s.calcGrahamNumber(fundamentals); graham != nil {
		methods = append(methods, *graham)
	}
	if peg := s.calcPEGValuation(fundamentals, &domain.Quote{PriceCents: entryPriceCents}); peg != nil {
		methods = append(methods, *peg)
	}
	if owner := s.calcOwnerEarnings(fundamentals, settings); owner != nil {
		methods = append(methods, *owner)
	}

	var totalWeight, weightedValue float64
	for _, m := range methods {
		totalWeight += m.Weight
		weightedValue += float64(m.ValueCents) * m.Weight
	}
	var intrinsicCents int64
	if totalWeight > 0 {
		intrinsicCents = int64(math.Round(weightedValue / totalWeight))
	}

	var marginOfSafety float64
	if entryPriceCents > 0 {
		marginOfSafety = float64(intrinsicCents-entryPriceCents) / float64(intrinsicCents)
	}

	qualityScore := s.calcQualityScore(fundamentals)
	verdict := s.determineVerdict(marginOfSafety, qualityScore)

	entryValuation := &domain.ValuationResult{
		Ticker:         ticker,
		CurrentCents:   entryPriceCents,
		IntrinsicCents: intrinsicCents,
		MarginOfSafety: marginOfSafety,
		Verdict:        verdict,
		QualityScore:   qualityScore,
		Methods:        methods,
	}

	returnPct := float64(currentCents-entryPriceCents) / float64(entryPriceCents)
	holdingDays := int(time.Since(entryDate).Hours() / 24)
	years := float64(holdingDays) / 365.25
	var annualizedPct float64
	if years > 0 && currentCents > 0 && entryPriceCents > 0 {
		annualizedPct = math.Pow(float64(currentCents)/float64(entryPriceCents), 1.0/years) - 1.0
	}

	actualOutcome := "GAIN"
	if returnPct < 0 {
		actualOutcome = "LOSS"
	}

	modelCorrect := false
	switch verdict {
	case domain.VerdictUndervalued, domain.VerdictFairLow:
		modelCorrect = returnPct > 0
	case domain.VerdictOvervalued:
		modelCorrect = returnPct < 0
	case domain.VerdictFair:
		modelCorrect = math.Abs(returnPct) < 0.20
	}

	return &domain.BacktestResult{
		Ticker:          ticker,
		EntryDate:       entryDateStr,
		EntryPriceCents: entryPriceCents,
		CurrentCents:    currentCents,
		ReturnPct:       returnPct,
		AnnualizedPct:   annualizedPct,
		HoldingDays:     holdingDays,
		EntryValuation:  entryValuation,
		ModelCorrect:    modelCorrect,
		ModelVerdict:    string(verdict),
		ActualOutcome:   actualOutcome,
	}, nil
}

func findClosestPrice(points []domain.PricePoint, targetDate string) int64 {
	if len(points) == 0 {
		return 0
	}
	for _, p := range points {
		if p.Date == targetDate {
			return p.CloseCents
		}
	}
	target, _ := time.Parse("2006-01-02", targetDate)
	var closest int64
	minDiff := time.Duration(math.MaxInt64)
	for _, p := range points {
		d, _ := time.Parse("2006-01-02", p.Date)
		diff := target.Sub(d)
		if diff < 0 {
			diff = -diff
		}
		if diff < minDiff {
			minDiff = diff
			closest = p.CloseCents
		}
	}
	if minDiff <= 5*24*time.Hour {
		return closest
	}
	return 0
}

// ScreenWatchlist runs valuation on all stocks in a watchlist and returns
// results sorted by margin of safety (best opportunities first).
func (s *ValuationService) ScreenWatchlist(ctx context.Context, watchlistID int64) ([]domain.ValuationResult, error) {
	items, err := s.watchlistRepo.GetItems(ctx, watchlistID)
	if err != nil {
		return nil, fmt.Errorf("get watchlist items: %w", err)
	}

	var results []domain.ValuationResult
	for _, item := range items {
		if item.Stock == nil {
			continue
		}
		val, err := s.GetValuation(ctx, item.Stock.Ticker)
		if err != nil {
			// Skip stocks that fail (e.g. missing fundamentals)
			continue
		}
		results = append(results, *val)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].MarginOfSafety > results[j].MarginOfSafety
	})

	if results == nil {
		results = []domain.ValuationResult{}
	}

	return results, nil
}
