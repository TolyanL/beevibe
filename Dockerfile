# ---------- frontend ----------
FROM oven/bun:1-alpine AS frontend

# vue-tsc (type-check part of `bun run build`) needs Node, not the Bun runtime
RUN apk add --no-cache nodejs

WORKDIR /app

COPY frontend/package.json frontend/bun.lock ./
RUN bun install --frozen-lockfile

COPY frontend/ .
RUN bun run build

# ---------- backend ----------
FROM golang:1.27-alpine AS backend

RUN apk add --no-cache git

WORKDIR /src

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/beevibe ./cmd/app \
    && CGO_ENABLED=0 GOOS=linux GOBIN=/out go install github.com/pressly/goose/v3/cmd/goose@v3.27.3

# ---------- runtime ----------
FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -u 10001 app

WORKDIR /app

COPY --from=backend /out/beevibe /usr/local/bin/beevibe
COPY --from=backend /out/goose /usr/local/bin/goose
COPY --chown=app:app backend/configs/config.example.yaml ./configs/config.example.yaml
COPY --chown=app:app backend/migrations ./migrations
COPY --from=frontend --chown=app:app /app/dist ./static

# NOTE: no configs/config.yaml in the image; mount it or point APP_CONFIG_PATH at it
ENV APP_CONFIG_PATH=/app/configs/config.yaml

USER app

EXPOSE 8080

CMD ["beevibe"]
