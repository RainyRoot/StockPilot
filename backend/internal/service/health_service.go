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

type HealthService struct {
	portfolioRepo     repository.PortfolioRepo
	tradeRepo         repository.TradeRepo
	stockRepo         repository.StockRepo
	allocationRepo    repository.AllocationRepo
	watchlistRepo     repository.WatchlistRepo
	provider          scraper.DataProvider
	valuationService  *ValuationService
	allocationService *AllocationService
	portfolioService  *PortfolioService
	settingsRepo      settingsGetter
}

func NewHealthService(
	portfolioRepo repository.PortfolioRepo,
	tradeRepo repository.TradeRepo,
	stockRepo repository.StockRepo,
	allocationRepo repository.AllocationRepo,
	watchlistRepo repository.WatchlistRepo,
	provider scraper.DataProvider,
	valuationService *ValuationService,
	allocationService *AllocationService,
	portfolioService *PortfolioService,
	settingsRepo settingsGetter,
) *HealthService {
	return &HealthService{
		portfolioRepo:     portfolioRepo,
		tradeRepo:         tradeRepo,
		stockRepo:         stockRepo,
		allocationRepo:    allocationRepo,
		watchlistRepo:     watchlistRepo,
		provider:          provider,
		valuationService:  valuationService,
		allocationService: allocationService,
		portfolioService:  portfolioService,
		settingsRepo:      settingsRepo,
	}
}

func (s *HealthService) loadMaxPositionPct(ctx context.Context) float64 {
	if s.settingsRepo != nil {
		if settings, err := s.settingsRepo.Get(ctx); err == nil && settings.MaxPositionPct > 0 {
			return settings.MaxPositionPct
		}
	}
	return 10.0
}

func (s *HealthService) GetHealthScore(ctx context.Context, portfolioID int64) (*domain.HealthScore, error) {
	holdings, err := s.portfolioService.GetHoldings(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("get holdings: %w", err)
	}

	components := s.calcComponents(ctx, portfolioID, holdings)

	var totalScore float64
	var totalWeight float64
	for _, c := range components {
		totalScore += float64(c.Score) * c.Weight
		totalWeight += c.Weight
	}
	overall := 0
	if totalWeight > 0 {
		overall = int(math.Round(totalScore / totalWeight))
	}

	grade := scoreToGrade(overall)

	alerts := s.getAlerts(ctx, portfolioID, holdings)
	topPicks := s.getTopPicks(ctx, portfolioID, holdings)

	return &domain.HealthScore{
		PortfolioID:  portfolioID,
		OverallScore: overall,
		Grade:        grade,
		Components:   components,
		Alerts:       alerts,
		TopPicks:     topPicks,
	}, nil
}

func (s *HealthService) GetAlerts(ctx context.Context, portfolioID int64) ([]domain.HealthAlert, error) {
	holdings, err := s.portfolioService.GetHoldings(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("get holdings: %w", err)
	}
	return s.getAlerts(ctx, portfolioID, holdings), nil
}

func (s *HealthService) GetTopPicks(ctx context.Context, portfolioID int64) ([]domain.TopPick, error) {
	holdings, err := s.portfolioService.GetHoldings(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("get holdings: %w", err)
	}
	return s.getTopPicks(ctx, portfolioID, holdings), nil
}

func (s *HealthService) calcComponents(ctx context.Context, portfolioID int64, holdings []domain.Holding) []domain.HealthComponent {
	components := []domain.HealthComponent{
		s.calcDiversification(holdings),
		s.calcValuationQuality(ctx, holdings),
		s.calcAllocationAdherence(ctx, portfolioID, holdings),
		s.calcRiskBalance(ctx, holdings),
	}
	return components
}

// calcDiversification scores based on sector diversity using Herfindahl-Hirschman Index.
func (s *HealthService) calcDiversification(holdings []domain.Holding) domain.HealthComponent {
	if len(holdings) == 0 {
		return domain.HealthComponent{
			Name:        "Diversification",
			Score:       0,
			Weight:      0.25,
			Description: "No holdings to evaluate",
		}
	}

	sectorValue := make(map[string]int64)
	var totalValue int64
	for _, h := range holdings {
		sector := h.Stock.Sector
		if sector == "" {
			sector = "Other"
		}
		sectorValue[sector] += h.MarketValue
		totalValue += h.MarketValue
	}

	if totalValue == 0 {
		return domain.HealthComponent{
			Name:        "Diversification",
			Score:       0,
			Weight:      0.25,
			Description: "Portfolio has no market value",
		}
	}

	// HHI: sum of squared market shares
	var hhi float64
	for _, val := range sectorValue {
		share := float64(val) / float64(totalValue)
		hhi += share * share
	}

	// Score: 100 when HHI < 0.15, 0 when HHI >= 1.0
	var score int
	if hhi < 0.15 {
		score = 100
	} else if hhi >= 1.0 {
		score = 0
	} else {
		score = int(math.Round(100 * (1.0 - hhi) / 0.85))
	}

	return domain.HealthComponent{
		Name:        "Diversification",
		Score:       score,
		Weight:      0.25,
		Description: fmt.Sprintf("HHI: %.2f across %d sectors", hhi, len(sectorValue)),
	}
}

// calcValuationQuality scores based on average quality score of holdings.
func (s *HealthService) calcValuationQuality(ctx context.Context, holdings []domain.Holding) domain.HealthComponent {
	if len(holdings) == 0 {
		return domain.HealthComponent{
			Name:        "Valuation Quality",
			Score:       50,
			Weight:      0.25,
			Description: "No holdings to evaluate",
		}
	}

	var totalWeight float64
	var weightedQuality float64
	for _, h := range holdings {
		val, err := s.valuationService.GetValuation(ctx, h.Stock.Ticker)
		if err != nil {
			continue
		}
		weight := float64(h.MarketValue)
		totalWeight += weight
		weightedQuality += float64(val.QualityScore) * weight
	}

	score := 50
	if totalWeight > 0 {
		score = int(math.Round(weightedQuality / totalWeight))
	}

	return domain.HealthComponent{
		Name:        "Valuation Quality",
		Score:       score,
		Weight:      0.25,
		Description: fmt.Sprintf("Weighted avg quality score of %d holdings", len(holdings)),
	}
}

// calcAllocationAdherence scores how well actual allocation matches targets.
func (s *HealthService) calcAllocationAdherence(ctx context.Context, portfolioID int64, holdings []domain.Holding) domain.HealthComponent {
	targets, err := s.allocationService.GetTargets(ctx, portfolioID)
	if err != nil || len(targets) == 0 {
		return domain.HealthComponent{
			Name:        "Allocation Adherence",
			Score:       50,
			Weight:      0.25,
			Description: "No allocation targets set",
		}
	}

	comparison, err := s.allocationService.GetComparison(ctx, portfolioID)
	if err != nil {
		return domain.HealthComponent{
			Name:        "Allocation Adherence",
			Score:       50,
			Weight:      0.25,
			Description: "Could not compute allocation comparison",
		}
	}

	var totalDiff float64
	for _, c := range comparison {
		totalDiff += math.Abs(c.DiffPercent)
	}
	avgDiff := totalDiff / float64(len(comparison))

	// Score: 100 when avgDiff=0, linearly to 0 when avgDiff>=20
	score := int(math.Round(math.Max(0, 100*(1-avgDiff/20))))

	return domain.HealthComponent{
		Name:        "Allocation Adherence",
		Score:       score,
		Weight:      0.25,
		Description: fmt.Sprintf("Avg deviation: %.1f%% across %d categories", avgDiff, len(comparison)),
	}
}

// calcRiskBalance scores based on position concentration and overall P&L.
func (s *HealthService) calcRiskBalance(ctx context.Context, holdings []domain.Holding) domain.HealthComponent {
	if len(holdings) == 0 {
		return domain.HealthComponent{
			Name:        "Risk Balance",
			Score:       50,
			Weight:      0.25,
			Description: "No holdings to evaluate",
		}
	}

	var totalValue int64
	var totalPL int64
	for _, h := range holdings {
		totalValue += h.MarketValue
		totalPL += h.UnrealizedPL
	}

	score := 100

	// Penalize concentration based on user-configured max position size
	maxPct := s.loadMaxPositionPct(ctx)
	for _, h := range holdings {
		if totalValue == 0 {
			break
		}
		pct := float64(h.MarketValue) / float64(totalValue) * 100
		if pct > maxPct*2 {
			score -= 30
		} else if pct > maxPct {
			score -= 15
		}
	}

	// Bonus for positive P&L
	if totalPL > 0 {
		score += 10
	} else if totalPL < 0 {
		score -= 10
	}

	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}

	return domain.HealthComponent{
		Name:        "Risk Balance",
		Score:       score,
		Weight:      0.25,
		Description: fmt.Sprintf("Concentration check across %d positions", len(holdings)),
	}
}

func (s *HealthService) getAlerts(ctx context.Context, portfolioID int64, holdings []domain.Holding) []domain.HealthAlert {
	var alerts []domain.HealthAlert
	var totalValue int64
	for _, h := range holdings {
		totalValue += h.MarketValue
	}

	maxPct := s.loadMaxPositionPct(ctx)
	for _, h := range holdings {
		// Concentration alerts based on user-configured max position size
		if totalValue > 0 {
			pct := float64(h.MarketValue) / float64(totalValue) * 100
			if pct > maxPct*2 {
				alerts = append(alerts, domain.HealthAlert{
					Severity: "critical",
					Title:    "Extreme concentration",
					Message:  fmt.Sprintf("%s makes up %.1f%% of your portfolio (max: %.0f%%)", h.Stock.Ticker, pct, maxPct),
					Ticker:   h.Stock.Ticker,
				})
			} else if pct > maxPct {
				alerts = append(alerts, domain.HealthAlert{
					Severity: "warning",
					Title:    "Position too large",
					Message:  fmt.Sprintf("%s makes up %.1f%% of your portfolio (max: %.0f%%)", h.Stock.Ticker, pct, maxPct),
					Ticker:   h.Stock.Ticker,
				})
			}
		}

		// Critical: unrealized loss > 20%
		if h.UnrealizedPct < -0.20 {
			alerts = append(alerts, domain.HealthAlert{
				Severity: "critical",
				Title:    "Large unrealized loss",
				Message:  fmt.Sprintf("%s is down %.1f%%", h.Stock.Ticker, h.UnrealizedPct*100),
				Ticker:   h.Stock.Ticker,
			})
		}
	}

	// Warning: no allocation targets
	targets, err := s.allocationService.GetTargets(ctx, portfolioID)
	if err == nil && len(targets) == 0 {
		alerts = append(alerts, domain.HealthAlert{
			Severity: "warning",
			Title:    "No allocation targets",
			Message:  "Set target allocations to track your portfolio balance",
		})
	}

	// Warning/Info: check allocation drift
	if len(targets) > 0 {
		comparison, err := s.allocationService.GetComparison(ctx, portfolioID)
		if err == nil {
			// Sectors in targets but missing in actual
			for _, c := range comparison {
				if c.TargetPercent > 0 && c.ActualPercent == 0 {
					alerts = append(alerts, domain.HealthAlert{
						Severity: "warning",
						Title:    "Missing sector",
						Message:  fmt.Sprintf("Target sector '%s' has no holdings", c.Category),
					})
				}
				if math.Abs(c.DiffPercent) > 5 {
					alerts = append(alerts, domain.HealthAlert{
						Severity: "info",
						Title:    "Rebalancing recommended",
						Message:  fmt.Sprintf("'%s' deviates %.1f%% from target", c.Category, c.DiffPercent),
					})
				}
			}
		}
	}

	// Info: undervalued stock in watchlist
	watchlists, err := s.watchlistRepo.List(ctx)
	if err == nil {
		for _, wl := range watchlists {
			items, err := s.watchlistRepo.GetItems(ctx, wl.ID)
			if err != nil {
				continue
			}
			for _, item := range items {
				if item.Stock == nil {
					continue
				}
				val, err := s.valuationService.GetValuation(ctx, item.Stock.Ticker)
				if err != nil {
					continue
				}
				if val.Verdict == domain.VerdictUndervalued {
					alerts = append(alerts, domain.HealthAlert{
						Severity: "info",
						Title:    "Undervalued watchlist stock",
						Message:  fmt.Sprintf("%s appears undervalued (margin of safety: %.0f%%)", item.Stock.Ticker, val.MarginOfSafety*100),
						Ticker:   item.Stock.Ticker,
					})
				}
			}
		}
	}

	if alerts == nil {
		alerts = []domain.HealthAlert{}
	}

	return alerts
}

func (s *HealthService) getTopPicks(ctx context.Context, portfolioID int64, holdings []domain.Holding) []domain.TopPick {
	var picks []domain.TopPick

	// Holdings with valuation: top 3 undervalued with high quality -> "HOLD"
	type holdingVal struct {
		ticker         string
		name           string
		marginOfSafety float64
		qualityScore   int
	}
	var holdVals []holdingVal
	for _, h := range holdings {
		val, err := s.valuationService.GetValuation(ctx, h.Stock.Ticker)
		if err != nil {
			continue
		}
		holdVals = append(holdVals, holdingVal{
			ticker:         h.Stock.Ticker,
			name:           h.Stock.Name,
			marginOfSafety: val.MarginOfSafety,
			qualityScore:   val.QualityScore,
		})
	}
	sort.Slice(holdVals, func(i, j int) bool {
		return holdVals[i].marginOfSafety > holdVals[j].marginOfSafety
	})
	for i, hv := range holdVals {
		if i >= 3 {
			break
		}
		if hv.marginOfSafety > 0 && hv.qualityScore >= 50 {
			picks = append(picks, domain.TopPick{
				Ticker:         hv.ticker,
				Name:           hv.name,
				Action:         "HOLD",
				Reason:         fmt.Sprintf("Undervalued with %.0f%% margin of safety", hv.marginOfSafety*100),
				MarginOfSafety: hv.marginOfSafety,
				QualityScore:   hv.qualityScore,
			})
		}
	}

	// Watchlist items: top 3 undervalued -> "WATCH"
	watchlists, err := s.watchlistRepo.List(ctx)
	if err == nil {
		var watchVals []holdingVal
		for _, wl := range watchlists {
			items, err := s.watchlistRepo.GetItems(ctx, wl.ID)
			if err != nil {
				continue
			}
			for _, item := range items {
				if item.Stock == nil {
					continue
				}
				val, err := s.valuationService.GetValuation(ctx, item.Stock.Ticker)
				if err != nil {
					continue
				}
				if val.MarginOfSafety > 0 {
					watchVals = append(watchVals, holdingVal{
						ticker:         item.Stock.Ticker,
						name:           item.Stock.Name,
						marginOfSafety: val.MarginOfSafety,
						qualityScore:   val.QualityScore,
					})
				}
			}
		}
		sort.Slice(watchVals, func(i, j int) bool {
			return watchVals[i].marginOfSafety > watchVals[j].marginOfSafety
		})
		for i, wv := range watchVals {
			if i >= 3 {
				break
			}
			picks = append(picks, domain.TopPick{
				Ticker:         wv.ticker,
				Name:           wv.name,
				Action:         "WATCH",
				Reason:         fmt.Sprintf("Watchlist: undervalued with %.0f%% margin of safety", wv.marginOfSafety*100),
				MarginOfSafety: wv.marginOfSafety,
				QualityScore:   wv.qualityScore,
			})
		}
	}

	// Holdings in under-allocated sectors -> "BUY"
	comparison, err := s.allocationService.GetComparison(ctx, portfolioID)
	if err == nil {
		for _, c := range comparison {
			if c.DiffPercent < -5 {
				// Find a holding in this sector
				for _, h := range holdings {
					sector := h.Stock.Sector
					if sector == "" {
						sector = "Other"
					}
					if sector == c.Category {
						val, err := s.valuationService.GetValuation(ctx, h.Stock.Ticker)
						qs := 0
						mos := 0.0
						if err == nil {
							qs = val.QualityScore
							mos = val.MarginOfSafety
						}
						picks = append(picks, domain.TopPick{
							Ticker:         h.Stock.Ticker,
							Name:           h.Stock.Name,
							Action:         "BUY",
							Reason:         fmt.Sprintf("Sector '%s' is %.1f%% under target", c.Category, math.Abs(c.DiffPercent)),
							MarginOfSafety: mos,
							QualityScore:   qs,
						})
						break
					}
				}
			}
		}
	}

	if picks == nil {
		picks = []domain.TopPick{}
	}

	return picks
}

func scoreToGrade(score int) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	default:
		return "F"
	}
}
