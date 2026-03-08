package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/rainyroot/stockpilot/backend/internal/domain"
)

type PortfolioRepo struct {
	db *sql.DB
}

func NewPortfolioRepo(db *sql.DB) *PortfolioRepo {
	return &PortfolioRepo{db: db}
}

func (r *PortfolioRepo) List(ctx context.Context) ([]domain.Portfolio, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, name, description, currency, created_at, updated_at FROM portfolios ORDER BY created_at ASC")
	if err != nil {
		return nil, fmt.Errorf("list portfolios: %w", err)
	}
	defer rows.Close()

	var portfolios []domain.Portfolio
	for rows.Next() {
		var p domain.Portfolio
		var desc sql.NullString
		var createdAt, updatedAt int64
		if err := rows.Scan(&p.ID, &p.Name, &desc, &p.Currency, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan portfolio: %w", err)
		}
		p.Description = desc.String
		p.CreatedAt = time.Unix(createdAt, 0)
		p.UpdatedAt = time.Unix(updatedAt, 0)
		portfolios = append(portfolios, p)
	}
	return portfolios, nil
}

func (r *PortfolioRepo) GetByID(ctx context.Context, id int64) (*domain.Portfolio, error) {
	var p domain.Portfolio
	var desc sql.NullString
	var createdAt, updatedAt int64
	err := r.db.QueryRowContext(ctx,
		"SELECT id, name, description, currency, created_at, updated_at FROM portfolios WHERE id = ?", id,
	).Scan(&p.ID, &p.Name, &desc, &p.Currency, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get portfolio: %w", err)
	}
	p.Description = desc.String
	p.CreatedAt = time.Unix(createdAt, 0)
	p.UpdatedAt = time.Unix(updatedAt, 0)
	return &p, nil
}

func (r *PortfolioRepo) Create(ctx context.Context, name string, description string, currency string) (*domain.Portfolio, error) {
	if currency == "" {
		currency = "EUR"
	}
	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx,
		"INSERT INTO portfolios (name, description, currency, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		name, nullString(description), currency, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("create portfolio: %w", err)
	}
	id, _ := result.LastInsertId()
	return &domain.Portfolio{
		ID:          id,
		Name:        name,
		Description: description,
		Currency:    currency,
		CreatedAt:   time.Unix(now, 0),
		UpdatedAt:   time.Unix(now, 0),
	}, nil
}

func (r *PortfolioRepo) Update(ctx context.Context, id int64, name string, description string) error {
	now := time.Now().Unix()
	_, err := r.db.ExecContext(ctx,
		"UPDATE portfolios SET name = ?, description = ?, updated_at = ? WHERE id = ?",
		name, nullString(description), now, id,
	)
	if err != nil {
		return fmt.Errorf("update portfolio: %w", err)
	}
	return nil
}

func (r *PortfolioRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM portfolios WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete portfolio: %w", err)
	}
	return nil
}
