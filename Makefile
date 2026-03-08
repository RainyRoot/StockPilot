.PHONY: dev dev-backend dev-frontend build build-backend build-frontend clean

# Development
dev:
	@echo "Starting StockPilot development servers..."
	@$(MAKE) dev-backend &
	@$(MAKE) dev-frontend &
	@wait

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

# Clean
clean:
	rm -f backend/stockpilot-server
	rm -f backend/*.db
	rm -rf frontend/.svelte-kit
	rm -rf frontend/build
