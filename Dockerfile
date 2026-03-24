FROM golang:1.25-alpine AS backend-builder

WORKDIR /build

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /server ./cmd/server/

FROM node:22-alpine AS frontend-builder

RUN corepack enable && corepack prepare pnpm@latest --activate

WORKDIR /build

COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile && pnpm add -D @sveltejs/adapter-node

COPY frontend/ .

RUN sed -i "s|adapter-auto|adapter-node|g" svelte.config.js && pnpm build

FROM node:22-alpine

RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -h /app appuser

WORKDIR /app

COPY --from=backend-builder /server ./server
COPY --from=frontend-builder /build/build ./frontend/

RUN mkdir -p /data && chown -R appuser:appuser /app /data

USER appuser

ENV PORT=8080
ENV DB_PATH=/data/stockpilot.db
ENV FRONTEND_URL=http://localhost:3000

EXPOSE 8080 3000

VOLUME ["/data"]

COPY <<'ENTRYPOINT' /app/start.sh
#!/bin/sh
node /app/frontend/index.js &
exec /app/server
ENTRYPOINT

ENTRYPOINT ["sh", "/app/start.sh"]
