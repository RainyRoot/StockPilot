.PHONY: dev dev-backend dev-frontend build build-backend build-frontend build-release clean

# Development
dev:
	@echo "Starting StockPilot development servers..."
	@lsof -ti:8080 | xargs -r kill 2>/dev/null || true
	@lsof -ti:5173 | xargs -r kill 2>/dev/null || true
	@trap 'kill 0' EXIT; \
		(cd backend && go run ./cmd/server/) & \
		(cd frontend && PATH="$$HOME/.local/bin:$$PATH" pnpm dev) & \
		wait

dev-backend:
	cd backend && go run ./cmd/server/

dev-frontend:
	cd frontend && pnpm dev

# Build
build: build-backend build-frontend

build-backend:
	cd backend && go build -o stockpilot-server ./cmd/server/

build-frontend:
	cd frontend && pnpm build

# Release build: embeds frontend into Go binary
build-release:
	cd frontend && PATH="$$HOME/.local/bin:$$PATH" pnpm install && PATH="$$HOME/.local/bin:$$PATH" pnpm build
	rm -rf backend/cmd/server/static
	cp -r frontend/build backend/cmd/server/static
	cd backend && CGO_ENABLED=0 go build -ldflags="-s -w" -o stockpilot ./cmd/server/

# Clean
clean:
	rm -f backend/stockpilot-server
	rm -f backend/stockpilot
	rm -f backend/*.db
	rm -rf frontend/.svelte-kit
	rm -rf frontend/build
