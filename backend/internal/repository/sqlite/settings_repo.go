package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/rainyroot/stockpilot/backend/internal/domain"
)

type SettingsRepo struct {
	db *sql.DB
}

func NewSettingsRepo(db *sql.DB) *SettingsRepo {
	return &SettingsRepo{db: db}
}

func (r *SettingsRepo) Get(ctx context.Context) (*domain.UserSettings, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT key, value FROM user_settings")
	if err != nil {
		return nil, fmt.Errorf("query settings: %w", err)
	}
	defer rows.Close()

	kv := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, fmt.Errorf("scan setting: %w", err)
		}
		kv[k] = v
	}

	settings := &domain.UserSettings{
		Currency:       "USD",
		DiscountRate:   0.10,
		GrowthRate:     0.05,
		TerminalGrowth: 0.025,
		MaxPositionPct: 10.0,
		Theme:          "dark",
	}

	if v, ok := kv["currency"]; ok {
		settings.Currency = v
	}
	if v, ok := kv["discount_rate"]; ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			settings.DiscountRate = f
		}
	}
	if v, ok := kv["growth_rate"]; ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			settings.GrowthRate = f
		}
	}
	if v, ok := kv["terminal_growth"]; ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			settings.TerminalGrowth = f
		}
	}
	if v, ok := kv["max_position_pct"]; ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			settings.MaxPositionPct = f
		}
	}
	if v, ok := kv["theme"]; ok {
		settings.Theme = v
	}

	return settings, nil
}

func (r *SettingsRepo) Update(ctx context.Context, settings *domain.UserSettings) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "INSERT OR REPLACE INTO user_settings (key, value) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Close()

	pairs := map[string]string{
		"currency":         settings.Currency,
		"discount_rate":    strconv.FormatFloat(settings.DiscountRate, 'f', -1, 64),
		"growth_rate":      strconv.FormatFloat(settings.GrowthRate, 'f', -1, 64),
		"terminal_growth":  strconv.FormatFloat(settings.TerminalGrowth, 'f', -1, 64),
		"max_position_pct": strconv.FormatFloat(settings.MaxPositionPct, 'f', -1, 64),
		"theme":            settings.Theme,
	}

	for k, v := range pairs {
		if _, err := stmt.ExecContext(ctx, k, v); err != nil {
			return fmt.Errorf("upsert %s: %w", k, err)
		}
	}

	return tx.Commit()
}
