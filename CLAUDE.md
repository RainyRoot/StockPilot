# StockPilot — Project Conventions

## Repository
- Commit frequently with descriptive messages
- Never add "Co-Authored-By: Claude" to commits
- Ask before pushing to any remote

## Technology Stack
- Backend: Go (REST API, chi router, raw SQL)
- Frontend: SvelteKit + TypeScript (pnpm)
- Database: SQLite (modernc.org/sqlite, pure Go)
- Data Source: Yahoo Finance (unofficial JSON API)
- Markets: US (NYSE, NASDAQ) + DE (XETRA)

## Project Structure
- `backend/cmd/server/` — Entrypoint
- `backend/internal/domain/` — Pure domain types (no external imports)
- `backend/internal/service/` — Business logic
- `backend/internal/repository/` — DB interfaces + SQLite implementation
- `backend/internal/scraper/` — Yahoo Finance client (behind DataProvider interface)
- `backend/internal/handler/` — HTTP handlers (thin: validate, call service, serialize)
- `backend/internal/config/` — Environment-based config
- `backend/pkg/money/` — Cent arithmetic (int64)
- `backend/pkg/httputil/` — JSON response helpers
- `frontend/` — SvelteKit app

## Code Style (Go)
- No ORM — raw SQL with database/sql
- All monetary values in cents (int64), never float
- All errors wrapped with context
- Use context.Context for cancellation
- Handlers are thin — business logic in services

## Code Style (Frontend)
- TypeScript strict mode
- pnpm only (never npm/yarn)
- Single quotes, 2-space indentation, semicolons

## Running
- `make dev` — starts both backend (8080) and frontend (5173)
- `make build` — builds both
- `make migrate` — runs DB migrations

## Environment
- pnpm: always `export PATH="$HOME/.local/bin:$PATH"` before pnpm commands
- Go: standard system install
