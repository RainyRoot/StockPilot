# Installation

## Prerequisites

- **Go** 1.25 or later — [golang.org/dl](https://golang.org/dl/)
- **Node.js** 18+ — [nodejs.org](https://nodejs.org/)
- **pnpm** — `npm install -g pnpm` (or via [pnpm.io](https://pnpm.io/installation))
- **Make** — pre-installed on macOS/Linux; on Windows use WSL or `choco install make`

## Clone the Repository

```bash
git clone https://github.com/RainyRoot/StockPilot.git
cd StockPilot
```

## Backend Setup

The Go backend compiles to a single binary with an embedded SQLite database — no external database server needed.

```bash
cd backend
go build -o stockpilot-server ./cmd/server/
```

Or use Make from the project root:

```bash
make build-backend
```

### Configuration

All settings are read from environment variables with sensible defaults:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | API server listen port |
| `DB_PATH` | `stockpilot.db` | Path to SQLite database file |
| `CACHE_TTL_SECS` | `60` | Yahoo Finance response cache duration |
| `FRONTEND_URL` | `http://localhost:5173` | Allowed CORS origin |

Create a `.env` file or export variables before starting:

```bash
export DB_PATH="./data/stockpilot.db"
export FRONTEND_URL="http://localhost:5173"
```

### Run the Backend

```bash
make dev-backend
# or: cd backend && go run ./cmd/server/
```

The API starts on `http://localhost:8080`. Verify:

```bash
curl http://localhost:8080/api/v1/health
# {"status":"ok"}
```

Database migrations run automatically on first startup.

## Frontend Setup

```bash
cd frontend
export PATH="$HOME/.local/bin:$PATH"
pnpm install
```

### Development Server

```bash
pnpm dev
# Frontend available at http://localhost:5173
```

### Production Build

```bash
pnpm build
pnpm preview  # Preview the production build locally
```

## Run Everything at Once

From the project root:

```bash
make dev
```

This starts both the Go API server (port 8080) and the SvelteKit dev server (port 5173) with automatic cleanup on Ctrl+C.

## Docker Setup

The easiest way to run StockPilot in production:

```bash
# Pull the pre-built image
docker pull ghcr.io/rainyroot/stockpilot:latest

# Run with a persistent volume for the database
docker run -d \
  --name stockpilot \
  -p 8080:8080 \
  -v stockpilot-data:/data \
  -e DB_PATH=/data/stockpilot.db \
  ghcr.io/rainyroot/stockpilot:latest
```

### Build Locally

```bash
docker build -t stockpilot .
docker run -d -p 8080:8080 -v stockpilot-data:/data -e DB_PATH=/data/stockpilot.db stockpilot
```

Access the application at `http://localhost:8080`.
