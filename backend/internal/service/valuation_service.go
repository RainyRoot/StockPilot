package service

import (
	"context"
	"fmt"
	"math"
	"sort"

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
	quote, err := s.provider.GetQuote(ctx, ticker)
	if err != nil {
		return nil, err
	}

	fundamentals, err := s.provider.GetFundamentals(ctx, ticker)
	if err != nil {
		return nil, err
	}

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

	return &domain.ValuationResult{
		Ticker:         ticker,
		CurrentCents:   int64(quote.PriceCents),
		IntrinsicCents: intrinsicCents,
		MarginOfSafety: marginOfSafety,
		Verdict:        verdict,
		QualityScore:   qualityScore,
		Methods:        methods,
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
