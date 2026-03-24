# Usage

## Getting Started

Once both backend and frontend are running (`make dev`), open `http://localhost:5173` in your browser.

## Web Interface

### Dashboard

The main dashboard shows an overview of all your portfolios with:

- Total portfolio value and daily change
- Per-portfolio summary cards
- P&L breakdown by holding (bar chart)
- Historical performance chart

Select a portfolio from the dashboard to dive into detailed views.

### Managing Portfolios

1. Navigate to the **Portfolio** page
2. Click **Create Portfolio** to add a new one
3. Record trades (buy/sell) with ticker, quantity, price, and optional bucket assignment
4. View current holdings with real-time prices and P&L

### Logging Trades

Each trade records:
- **Ticker** — Stock symbol (e.g., `AAPL`, `SAP.DE` for XETRA)
- **Side** — Buy or Sell
- **Quantity** — Number of shares
- **Price** — Price per share in cents (integer)
- **Bucket** — Optional strategy label (core, growth, speculative, etc.)

### Watchlists

Create watchlists to track stocks you're interested in:

1. Go to the **Watchlist** page
2. Create a new watchlist and add tickers
3. Use the **Screen** feature to filter by valuation metrics

### Valuation Analysis

The **Valuation** page provides fundamental analysis for any ticker:

- P/E ratio and fair value estimate
- PEG ratio with historical EPS growth
- Fundamental data summary
- Backtesting valuation models against historical prices

### Allocation & Rebalancing

On the **Allocation** page:

1. Set target allocation percentages (by sector, region, or custom category)
2. View actual vs. target allocation
3. Get rebalancing suggestions showing what to buy/sell to reach targets

### Strategy Spread

The **Strategy** page shows how your trades are distributed across strategy buckets, helping you maintain your intended risk profile.

### Settings

Configure:
- Display currency
- Alert thresholds (concentration limits, position size warnings)
- Portfolio strategy types
- Yahoo Finance cache duration

## API Usage

### Search for a Stock

```bash
curl "http://localhost:8080/api/v1/stocks/search?q=apple"
```

### Get a Live Quote

```bash
curl http://localhost:8080/api/v1/stocks/AAPL
```

### Create a Portfolio

```bash
curl -X POST http://localhost:8080/api/v1/portfolios \
  -H "Content-Type: application/json" \
  -d '{"name":"Tech Picks","currency":"USD"}'
```

### Record a Trade

```bash
curl -X POST http://localhost:8080/api/v1/portfolios/1/trades \
  -H "Content-Type: application/json" \
  -d '{"ticker":"AAPL","side":"buy","quantity":10,"price_cents":17500,"bucket":"core"}'
```

### Get Portfolio Holdings

```bash
curl http://localhost:8080/api/v1/portfolios/1/holdings
```

### Run a Valuation

```bash
curl http://localhost:8080/api/v1/valuation/MSFT
```

### Export Data

```bash
# Export holdings as CSV
curl http://localhost:8080/api/v1/portfolios/1/export/holdings.csv -o holdings.csv

# Export trades as CSV
curl http://localhost:8080/api/v1/portfolios/1/export/trades.csv -o trades.csv
```

## Configuration Reference

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | API server listen port |
| `DB_PATH` | `stockpilot.db` | Path to SQLite database file |
| `CACHE_TTL_SECS` | `60` | Yahoo Finance cache TTL in seconds |
| `FRONTEND_URL` | `http://localhost:5173` | Allowed CORS origin for the frontend |

## Troubleshooting

### Yahoo Finance requests are slow or timing out

Yahoo Finance is an unofficial API and may rate-limit requests. Increase `CACHE_TTL_SECS` to reduce the number of outgoing requests:

```bash
export CACHE_TTL_SECS=300
```

### Frontend shows "Failed to fetch" errors

Make sure the backend is running and the `FRONTEND_URL` environment variable matches the URL where your frontend is served. For local development, the defaults (`localhost:8080` and `localhost:5173`) should work out of the box.

### German stocks not showing correct prices

Use the `.DE` suffix for XETRA-listed stocks (e.g., `SAP.DE`, `BMW.DE`). Yahoo Finance uses this convention for German market tickers.
