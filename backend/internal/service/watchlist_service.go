package service

import (
	"context"
	"fmt"

	"github.com/rainyroot/stockpilot/backend/internal/domain"
	"github.com/rainyroot/stockpilot/backend/internal/repository"
	"github.com/rainyroot/stockpilot/backend/internal/scraper"
)

type WatchlistService struct {
	watchlistRepo repository.WatchlistRepo
	stockRepo     repository.StockRepo
	provider      scraper.DataProvider
	cacheTTL      int
}

func NewWatchlistService(
	watchlistRepo repository.WatchlistRepo,
	stockRepo repository.StockRepo,
	provider scraper.DataProvider,
	cacheTTL int,
) *WatchlistService {
	return &WatchlistService{
		watchlistRepo: watchlistRepo,
		stockRepo:     stockRepo,
		provider:      provider,
		cacheTTL:      cacheTTL,
	}
}

func (s *WatchlistService) List(ctx context.Context) ([]domain.Watchlist, error) {
	watchlists, err := s.watchlistRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list watchlists: %w", err)
	}
	return watchlists, nil
}

func (s *WatchlistService) Get(ctx context.Context, id int64) (*domain.Watchlist, error) {
	watchlist, err := s.watchlistRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get watchlist: %w", err)
	}
	if watchlist == nil {
		return nil, fmt.Errorf("watchlist %d not found", id)
	}

	items, err := s.watchlistRepo.GetItems(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get watchlist items: %w", err)
	}

	// Fetch batch quotes for all tickers
	if len(items) > 0 {
		tickers := make([]string, len(items))
		for i, item := range items {
			tickers[i] = item.Stock.Ticker
		}

		quotes, err := s.provider.GetBatchQuotes(ctx, tickers)
		if err != nil {
			// Log but don't fail — still return items without quotes
			fmt.Printf("warning: failed to fetch batch quotes: %v\n", err)
		} else {
			quoteMap := make(map[string]*domain.Quote)
			for i := range quotes {
				quoteMap[quotes[i].Ticker] = &quotes[i]
			}
			for i := range items {
				if q, ok := quoteMap[items[i].Stock.Ticker]; ok {
					items[i].Quote = q
				}
			}
		}
	}

	watchlist.Items = items
	return watchlist, nil
}

func (s *WatchlistService) Create(ctx context.Context, name string) (*domain.Watchlist, error) {
	if name == "" {
		name = "My Watchlist"
	}
	watchlist, err := s.watchlistRepo.Create(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("create watchlist: %w", err)
	}
	return watchlist, nil
}

func (s *WatchlistService) Update(ctx context.Context, id int64, name string) error {
	if err := s.watchlistRepo.Update(ctx, id, name); err != nil {
		return fmt.Errorf("update watchlist: %w", err)
	}
	return nil
}

func (s *WatchlistService) Delete(ctx context.Context, id int64) error {
	if err := s.watchlistRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete watchlist: %w", err)
	}
	return nil
}

func (s *WatchlistService) AddStock(ctx context.Context, watchlistID int64, ticker string, notes string) error {
	// Ensure stock exists in DB (fetch from Yahoo if needed)
	stock, err := s.stockRepo.GetByTicker(ctx, ticker)
	if err != nil {
		return fmt.Errorf("get stock: %w", err)
	}

	if stock == nil {
		// Fetch from Yahoo to get stock info
		quote, err := s.provider.GetQuote(ctx, ticker)
		if err != nil {
			return fmt.Errorf("fetch quote for %s: %w", ticker, err)
		}
		stock = &domain.Stock{
			Ticker:   quote.Ticker,
			Name:     quote.Name,
			Exchange: quote.Exchange,
			Currency: quote.Currency,
		}
		if err := s.stockRepo.Upsert(ctx, stock); err != nil {
			return fmt.Errorf("upsert stock: %w", err)
		}
	}

	if err := s.watchlistRepo.AddItem(ctx, watchlistID, stock.ID, notes); err != nil {
		return fmt.Errorf("add stock to watchlist: %w", err)
	}
	return nil
}

func (s *WatchlistService) RemoveStock(ctx context.Context, watchlistID int64, ticker string) error {
	stock, err := s.stockRepo.GetByTicker(ctx, ticker)
	if err != nil {
		return fmt.Errorf("get stock: %w", err)
	}
	if stock == nil {
		return fmt.Errorf("stock %s not found", ticker)
	}

	if err := s.watchlistRepo.RemoveItem(ctx, watchlistID, stock.ID); err != nil {
		return fmt.Errorf("remove stock from watchlist: %w", err)
	}
	return nil
}
