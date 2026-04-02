FROM node:22-alpine AS frontend-builder

RUN corepack enable && corepack prepare pnpm@latest --activate

WORKDIR /build

COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY frontend/ .
RUN pnpm build


FROM golang:1.25-alpine AS backend-builder

WORKDIR /build

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .
COPY --from=frontend-builder /build/build ./cmd/server/static

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /server ./cmd/server/


FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -h /app appuser

WORKDIR /app

COPY --from=backend-builder /server ./server

RUN mkdir -p /data && chown -R appuser:appuser /app /data

USER appuser

ENV PORT=8080
ENV DB_PATH=/data/stockpilot.db

EXPOSE 8080

VOLUME ["/data"]

ENTRYPOINT ["./server"]
