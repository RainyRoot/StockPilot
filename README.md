# StockPilot

A self-hosted stock portfolio tracker and analysis tool. StockPilot combines a Go REST API backend with a SvelteKit frontend to give you real-time quotes, portfolio tracking, valuation models, allocation analysis, and more — all powered by Yahoo Finance data.

## Features

- **Portfolio Management** — Create multiple portfolios, log trades (buy/sell), and track holdings in real time
- **Live Stock Quotes** — Search tickers and get current prices from Yahoo Finance (US + German markets)
- **Dashboard** — At-a-glance view of portfolio performance, P&L breakdown, and historical performance charts
- **Valuation Models** — PE, PEG, and fundamental analysis with historical EPS data and backtesting
- **Allocation Targets** — Set target allocations by sector/region and get rebalancing suggestions
- **Portfolio Health** — Automated alerts for concentration risk, overweight positions, and top pick recommendations
- **Watchlists** — Track stocks you're interested in and run screening filters
- **Strategy Spread** — Organize trades into buckets (core, growth, speculative, etc.)
- **Data Export** — Export holdings and trades as CSV
- **Configurable Settings** — Customize currency display, alert thresholds, and strategy types

## Quick Start

```bash
git clone https://github.com/RainyRoot/StockPilot.git
cd StockPilot
make dev
```

The backend API starts on `http://localhost:8080` and the SvelteKit frontend on `http://localhost:5173`.

See [INSTALL.md](INSTALL.md) for detailed setup and [USAGE.md](USAGE.md) for usage examples.

## Tech Stack

| Component | Technology |
|-----------|------------|
| Backend | Go 1.25+, chi router, raw SQL |
| Database | SQLite (modernc.org/sqlite, pure Go) |
| Frontend | SvelteKit 2, Svelte 5, TypeScript, Vite |
| Charts | Chart.js |
| Data Source | Yahoo Finance (unofficial JSON API) |
| Markets | US (NYSE, NASDAQ) + DE (XETRA) |
| Package Manager | pnpm |

## Project Structure

```
StockPilot/
├── backend/
│   ├── cmd/server/          # Application entrypoint
│   ├── internal/
│   │   ├── config/          # Environment-based configuration
│   │   ├── domain/          # Pure domain types
│   │   ├── handler/         # HTTP handlers, router, middleware
│   │   ├── repository/      # Repository interfaces + SQLite implementation
│   │   ├── scraper/         # Yahoo Finance data provider
│   │   └── service/         # Business logic layer
│   └── pkg/
│       ├── httputil/        # JSON response helpers
│       └── money/           # Cent-based arithmetic (int64)
├── frontend/
│   └── src/
│       ├── lib/
│       │   ├── api/         # Typed API client modules
│       │   ├── components/  # Svelte components (charts, skeleton, toast)
│       │   └── utils/       # Formatting utilities
│       └── routes/          # SvelteKit pages
└── Makefile
```

## Docker

Pull and run the complete application:

```bash
docker pull ghcr.io/rainyroot/stockpilot:latest

docker run -d \
  -p 8080:8080 \
  -v stockpilot-data:/data \
  -e DB_PATH=/data/stockpilot.db \
  ghcr.io/rainyroot/stockpilot:latest
```

Access the web UI at `http://localhost:8080`.

## API Overview

All endpoints live under `/api/v1`. Key routes:

- `GET /api/v1/stocks/search?q=AAPL` — Search tickers
- `GET /api/v1/stocks/{ticker}` — Live quote
- `GET /api/v1/portfolios` — List portfolios
- `POST /api/v1/portfolios/{id}/trades` — Record a trade
- `GET /api/v1/dashboard` — Dashboard overview
- `GET /api/v1/valuation/{ticker}` — Valuation analysis
- `GET /api/v1/settings` — User settings

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Open a pull request
