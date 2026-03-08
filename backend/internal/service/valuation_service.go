package service

import (
	"context"
	"math"

	"github.com/rainyroot/stockpilot/backend/internal/domain"
	"github.com/rainyroot/stockpilot/backend/internal/repository"
	"github.com/rainyroot/stockpilot/backend/internal/scraper"
)

type ValuationService struct {
	stockRepo repository.StockRepo
	provider  scraper.DataProvider
}

func NewValuationService(
	stockRepo repository.StockRepo,
	provider scraper.DataProvider,
) *ValuationService {
	return &ValuationService{
		stockRepo: stockRepo,
		provider:  provider,
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

	methods := []domain.MethodResult{}

	if dcf := s.calcDCF(fundamentals); dcf != nil {
		methods = append(methods, *dcf)
	}
	if graham := s.calcGrahamNumber(fundamentals); graham != nil {
		methods = append(methods, *graham)
	}
	if peg := s.calcPEGValuation(fundamentals, quote); peg != nil {
		methods = append(methods, *peg)
	}
	if owner := s.calcOwnerEarnings(fundamentals); owner != nil {
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
func (s *ValuationService) calcDCF(f *domain.Fundamentals) *domain.MethodResult {
	if f.FreeCashFlowCents <= 0 || f.SharesOutstanding <= 0 {
		return nil
	}

	fcfPerShare := float64(f.FreeCashFlowCents) / float64(f.SharesOutstanding)
	discountRate := 0.10
	growthRate := 0.05
	terminalGrowth := 0.025

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
		Description: "10-year FCF projection with 5% growth, 10% discount rate, 2.5% terminal growth",
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

	growthRate := f.ROE * 100
	if growthRate <= 0 {
		growthRate = 5.0
	}
	if growthRate > 30 {
		growthRate = 30.0
	}

	eps := float64(f.EPSCents) / 100.0
	fairPE := growthRate
	fairValue := eps * fairPE
	fairCents := int64(math.Round(fairValue * 100))

	return &domain.MethodResult{
		Name:        "PEG Ratio Fair Value",
		ValueCents:  fairCents,
		Weight:      0.15,
		Description: "Fair value assuming PEG ratio of 1.0 with estimated earnings growth",
	}
}

// calcOwnerEarnings — Buffett's Owner Earnings method
func (s *ValuationService) calcOwnerEarnings(f *domain.Fundamentals) *domain.MethodResult {
	if f.NetIncomeCents <= 0 || f.SharesOutstanding <= 0 {
		return nil
	}

	var ownerEarnings float64
	if f.FreeCashFlowCents > 0 {
		ownerEarnings = float64(f.FreeCashFlowCents) / float64(f.SharesOutstanding)
	} else {
		ownerEarnings = float64(f.NetIncomeCents) / float64(f.SharesOutstanding) * 0.8
	}

	intrinsicPerShare := ownerEarnings * 10.0
	intrinsicCents := int64(math.Round(intrinsicPerShare))

	return &domain.MethodResult{
		Name:        "Owner Earnings (Buffett)",
		ValueCents:  intrinsicCents,
		Weight:      0.25,
		Description: "Owner earnings capitalized at 10% required return",
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
