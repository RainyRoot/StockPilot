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

type AllocationService struct {
	allocationRepo repository.AllocationRepo
	portfolioRepo  repository.PortfolioRepo
	tradeRepo      repository.TradeRepo
	stockRepo      repository.StockRepo
	provider       scraper.DataProvider
}

func NewAllocationService(
	allocationRepo repository.AllocationRepo,
	portfolioRepo repository.PortfolioRepo,
	tradeRepo repository.TradeRepo,
	stockRepo repository.StockRepo,
	provider scraper.DataProvider,
) *AllocationService {
	return &AllocationService{
		allocationRepo: allocationRepo,
		portfolioRepo:  portfolioRepo,
		tradeRepo:      tradeRepo,
		stockRepo:      stockRepo,
		provider:       provider,
	}
}

func (s *AllocationService) GetTargets(ctx context.Context, portfolioID int64) ([]domain.AllocationTarget, error) {
	return s.allocationRepo.GetTargets(ctx, portfolioID)
}

func (s *AllocationService) SetTargets(ctx context.Context, portfolioID int64, targets []domain.AllocationTarget) error {
	var sum float64
	for _, t := range targets {
		sum += t.TargetPercent
	}
	if math.Abs(sum-100.0) > 0.01 {
		return fmt.Errorf("target percentages must sum to 100, got %.2f", sum)
	}
	return s.allocationRepo.SetTargets(ctx, portfolioID, targets)
}

func (s *AllocationService) GetActual(ctx context.Context, portfolioID int64) ([]domain.AllocationActual, error) {
	holdings, err := s.getHoldingsWithSectors(ctx, portfolioID)
	if err != nil {
		return nil, err
	}

	sectorValue := make(map[string]int64)
	var totalValue int64
	for _, h := range holdings {
		sector := h.Category
		if sector == "" {
			sector = "Other"
		}
		sectorValue[sector] += h.MarketValue
		totalValue += h.MarketValue
	}

	var actuals []domain.AllocationActual
	for cat, val := range sectorValue {
		var pct float64
		if totalValue > 0 {
			pct = float64(val) / float64(totalValue) * 100
		}
		actuals = append(actuals, domain.AllocationActual{
			Category:      cat,
			ActualPercent: math.Round(pct*100) / 100,
			ValueCents:    val,
		})
	}

	sort.Slice(actuals, func(i, j int) bool {
		return actuals[i].Category < actuals[j].Category
	})

	return actuals, nil
}

func (s *AllocationService) GetComparison(ctx context.Context, portfolioID int64) ([]domain.AllocationComparison, error) {
	targets, err := s.GetTargets(ctx, portfolioID)
	if err != nil {
		return nil, err
	}
	actuals, err := s.GetActual(ctx, portfolioID)
	if err != nil {
		return nil, err
	}

	targetMap := make(map[string]float64)
	for _, t := range targets {
		targetMap[t.Category] = t.TargetPercent
	}
	actualMap := make(map[string]domain.AllocationActual)
	for _, a := range actuals {
		actualMap[a.Category] = a
	}

	// Union of all categories
	categories := make(map[string]bool)
	for _, t := range targets {
		categories[t.Category] = true
	}
	for _, a := range actuals {
		categories[a.Category] = true
	}

	// Compute total portfolio value for DiffCents
	var totalValue int64
	for _, a := range actuals {
		totalValue += a.ValueCents
	}

	var comparisons []domain.AllocationComparison
	for cat := range categories {
		tp := targetMap[cat]
		ap := actualMap[cat].ActualPercent
		diff := ap - tp
		diffCents := int64(math.Round(diff / 100 * float64(totalValue)))

		comparisons = append(comparisons, domain.AllocationComparison{
			Category:      cat,
			TargetPercent: tp,
			ActualPercent: ap,
			DiffPercent:   math.Round(diff*100) / 100,
			DiffCents:     diffCents,
		})
	}

	sort.Slice(comparisons, func(i, j int) bool {
		return comparisons[i].Category < comparisons[j].Category
	})

	return comparisons, nil
}

func (s *AllocationService) GetRebalanceActions(ctx context.Context, portfolioID int64) ([]domain.RebalanceAction, error) {
	comparisons, err := s.GetComparison(ctx, portfolioID)
	if err != nil {
		return nil, err
	}
	holdings, err := s.getHoldingsWithSectors(ctx, portfolioID)
	if err != nil {
		return nil, err
	}

	var totalValue int64
	for _, h := range holdings {
		totalValue += h.MarketValue
	}

	// Group holdings by sector
	sectorHoldings := make(map[string][]domain.Holding)
	for _, h := range holdings {
		cat := h.Category
		if cat == "" {
			cat = "Other"
		}
		sectorHoldings[cat] = append(sectorHoldings[cat], h)
	}

	var actions []domain.RebalanceAction
	for _, c := range comparisons {
		if math.Abs(c.DiffPercent) <= 2 {
			continue
		}

		diffAmount := int64(math.Round(math.Abs(c.DiffPercent) / 100 * float64(totalValue)))
		sectorList := sectorHoldings[c.Category]

		if c.DiffPercent > 0 {
			// Over-allocated: SELL the largest holding in this sector
			if len(sectorList) == 0 {
				continue
			}
			sort.Slice(sectorList, func(i, j int) bool {
				return sectorList[i].MarketValue > sectorList[j].MarketValue
			})
			h := sectorList[0]
			var qty float64
			if h.CurrentCents > 0 {
				qty = math.Floor(float64(diffAmount) / float64(h.CurrentCents))
			}
			if qty < 1 {
				qty = 1
			}
			if qty > h.Quantity {
				qty = h.Quantity
			}
			actions = append(actions, domain.RebalanceAction{
				Ticker:     h.Stock.Ticker,
				Action:     "SELL",
				Quantity:   qty,
				ValueCents: int64(qty) * h.CurrentCents,
				Reason:     fmt.Sprintf("%s is %.1f%% over target allocation", c.Category, c.DiffPercent),
			})
		} else {
			// Under-allocated: BUY the smallest holding in this sector (or suggest any)
			if len(sectorList) > 0 {
				sort.Slice(sectorList, func(i, j int) bool {
					return sectorList[i].MarketValue < sectorList[j].MarketValue
				})
				h := sectorList[0]
				var qty float64
				if h.CurrentCents > 0 {
					qty = math.Floor(float64(diffAmount) / float64(h.CurrentCents))
				}
				if qty < 1 {
					qty = 1
				}
				actions = append(actions, domain.RebalanceAction{
					Ticker:     h.Stock.Ticker,
					Action:     "BUY",
					Quantity:   qty,
					ValueCents: int64(qty) * h.CurrentCents,
					Reason:     fmt.Sprintf("%s is %.1f%% under target allocation", c.Category, math.Abs(c.DiffPercent)),
				})
			}
		}
	}

	return actions, nil
}

// getHoldingsWithSectors fetches holdings and enriches them with sector info.
func (s *AllocationService) getHoldingsWithSectors(ctx context.Context, portfolioID int64) ([]domain.Holding, error) {
	portfolio, err := s.portfolioRepo.GetByID(ctx, portfolioID)
	if err != nil {
		return nil, err
	}
	if portfolio == nil {
		return nil, fmt.Errorf("portfolio %d not found", portfolioID)
	}

	rows, err := s.tradeRepo.GetHoldings(ctx, portfolioID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []domain.Holding{}, nil
	}

	tickers := make([]string, len(rows))
	for i, h := range rows {
		tickers[i] = h.Ticker
	}
	quotes, err := s.provider.GetBatchQuotes(ctx, tickers)
	if err != nil {
		fmt.Printf("warning: failed to fetch batch quotes: %v\n", err)
	}
	quoteMap := make(map[string]*domain.Quote)
	for i := range quotes {
		quoteMap[quotes[i].Ticker] = &quotes[i]
	}

	holdings := make([]domain.Holding, 0, len(rows))
	for _, row := range rows {
		avgCost := int64(math.Round(float64(row.TotalCostCents) / row.NetQuantity))

		var currentCents int64
		sector := ""
		if q, ok := quoteMap[row.Ticker]; ok {
			currentCents = q.PriceCents
			sector = q.Sector
		}

		// Fallback: check stored stock data
		if sector == "" {
			stock, err := s.stockRepo.GetByTicker(ctx, row.Ticker)
			if err == nil && stock != nil {
				sector = stock.Sector
			}
		}
		if sector == "" {
			sector = "Other"
		}

		marketValue := int64(math.Round(float64(currentCents) * row.NetQuantity))
		costBasis := int64(math.Round(float64(avgCost) * row.NetQuantity))
		unrealizedPL := marketValue - costBasis

		var unrealizedPct float64
		if costBasis != 0 {
			unrealizedPct = float64(unrealizedPL) / float64(costBasis)
		}

		holdings = append(holdings, domain.Holding{
			Stock: domain.Stock{
				ID:       row.StockID,
				Ticker:   row.Ticker,
				Name:     row.Name,
				Exchange: row.Exchange,
				Currency: row.Currency,
				Sector:   sector,
			},
			Quantity:      row.NetQuantity,
			AvgCostCents:  avgCost,
			CurrentCents:  currentCents,
			MarketValue:   marketValue,
			UnrealizedPL:  unrealizedPL,
			UnrealizedPct: unrealizedPct,
			Category:      sector,
		})
	}

	return holdings, nil
}
