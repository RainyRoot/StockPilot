package service

import (
	"context"
	"fmt"

	"github.com/rainyroot/stockpilot/backend/internal/domain"
	"github.com/rainyroot/stockpilot/backend/internal/repository/sqlite"
)

type SettingsService struct {
	repo *sqlite.SettingsRepo
}

func NewSettingsService(repo *sqlite.SettingsRepo) *SettingsService {
	return &SettingsService{repo: repo}
}

func (s *SettingsService) Get(ctx context.Context) (*domain.UserSettings, error) {
	return s.repo.Get(ctx)
}

func (s *SettingsService) Update(ctx context.Context, settings *domain.UserSettings) error {
	if settings.Currency != "USD" && settings.Currency != "EUR" {
		return fmt.Errorf("currency must be USD or EUR, got %s", settings.Currency)
	}
	if settings.DiscountRate < 0.01 || settings.DiscountRate > 0.30 {
		return fmt.Errorf("discount_rate must be between 0.01 and 0.30")
	}
	if settings.GrowthRate < 0.0 || settings.GrowthRate > 0.30 {
		return fmt.Errorf("growth_rate must be between 0.0 and 0.30")
	}
	if settings.TerminalGrowth < 0.0 || settings.TerminalGrowth > 0.10 {
		return fmt.Errorf("terminal_growth must be between 0.0 and 0.10")
	}
	return s.repo.Update(ctx, settings)
}
