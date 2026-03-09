package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"math"
	"time"

	"github.com/rainyroot/stockpilot/backend/internal/repository"
)

type ExportService struct {
	portfolioService *PortfolioService
	tradeRepo        repository.TradeRepo
}

func NewExportService(portfolioService *PortfolioService, tradeRepo repository.TradeRepo) *ExportService {
	return &ExportService{
		portfolioService: portfolioService,
		tradeRepo:        tradeRepo,
	}
}

func (s *ExportService) ExportHoldingsCSV(ctx context.Context, portfolioID int64) ([]byte, error) {
	holdings, err := s.portfolioService.GetHoldings(ctx, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("get holdings: %w", err)
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	w.Write([]string{"Ticker", "Name", "Quantity", "Avg Cost", "Current Price", "Market Value", "Unrealized P&L", "P&L %"})

	for _, h := range holdings {
		w.Write([]string{
			h.Stock.Ticker,
			h.Stock.Name,
			fmt.Sprintf("%.4f", h.Quantity),
			formatCentsCSV(h.AvgCostCents),
			formatCentsCSV(h.CurrentCents),
			formatCentsCSV(h.MarketValue),
			formatCentsCSV(h.UnrealizedPL),
			fmt.Sprintf("%.2f%%", h.UnrealizedPct*100),
		})
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("write csv: %w", err)
	}

	return buf.Bytes(), nil
}

func (s *ExportService) ExportTradesCSV(ctx context.Context, portfolioID int64) ([]byte, error) {
	trades, _, err := s.tradeRepo.ListByPortfolio(ctx, portfolioID, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("list trades: %w", err)
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	w.Write([]string{"Date", "Type", "Ticker", "Quantity", "Price", "Fee", "Total", "Currency", "Notes"})

	for _, t := range trades {
		total := int64(math.Round(float64(t.PriceCents)*t.Quantity)) + t.FeeCents
		w.Write([]string{
			t.ExecutedAt.Format(time.DateOnly),
			string(t.TradeType),
			t.Ticker,
			fmt.Sprintf("%.4f", t.Quantity),
			formatCentsCSV(t.PriceCents),
			formatCentsCSV(t.FeeCents),
			formatCentsCSV(total),
			t.Currency,
			t.Notes,
		})
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("write csv: %w", err)
	}

	return buf.Bytes(), nil
}

func formatCentsCSV(cents int64) string {
	return fmt.Sprintf("%.2f", float64(cents)/100.0)
}
