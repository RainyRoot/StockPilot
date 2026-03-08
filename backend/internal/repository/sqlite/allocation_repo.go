package sqlite

import (
	"context"
	"database/sql"

	"github.com/rainyroot/stockpilot/backend/internal/domain"
)

type AllocationRepo struct {
	db *sql.DB
}

func NewAllocationRepo(db *sql.DB) *AllocationRepo {
	return &AllocationRepo{db: db}
}

func (r *AllocationRepo) GetTargets(ctx context.Context, portfolioID int64) ([]domain.AllocationTarget, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, portfolio_id, category, target_percent FROM allocation_targets WHERE portfolio_id = ? ORDER BY category`, portfolioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var targets []domain.AllocationTarget
	for rows.Next() {
		var t domain.AllocationTarget
		if err := rows.Scan(&t.ID, &t.PortfolioID, &t.Category, &t.TargetPercent); err != nil {
			return nil, err
		}
		targets = append(targets, t)
	}
	return targets, rows.Err()
}

func (r *AllocationRepo) SetTargets(ctx context.Context, portfolioID int64, targets []domain.AllocationTarget) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM allocation_targets WHERE portfolio_id = ?`, portfolioID); err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO allocation_targets (portfolio_id, category, target_percent) VALUES (?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, t := range targets {
		if _, err := stmt.ExecContext(ctx, portfolioID, t.Category, t.TargetPercent); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *AllocationRepo) DeleteTarget(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM allocation_targets WHERE id = ?`, id)
	return err
}
