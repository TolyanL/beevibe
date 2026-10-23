BINARY := app
PKG := ./...
COMPOSE ?= docker compose -f ../docker-compose.dev.yaml

DB_USER        ?= bee
DB_PASSWORD    ?= 12345
DB_NAME        ?= bee
DB_PORT        ?= 5432
DB_DSN         ?= postgres://$(DB_USER):$(DB_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)?sslmode=disable
MIGRATION_DIR  ?= migrations

.PHONY: build run config lint test up down migrate-up migrate-down migrate-status migrate-create mocks tidy clean

build:
	go build -o bin/$(BINARY) ./cmd/$(BINARY)

run:
	go run ./cmd/$(BINARY)

config:
	@test -f configs/config.yaml || cp configs/config.example.yaml configs/config.yaml

lint:
	golangci-lint run

test:
	go test -v -cover $(PKG)

up:
	$(COMPOSE) up -d

down:
	$(COMPOSE) down

migrate-up:
	goose -dir $(MIGRATION_DIR) postgres "$(DB_DSN)" up

migrate-down:
	goose -dir $(MIGRATION_DIR) postgres "$(DB_DSN)" down

migrate-status:
	goose -dir $(MIGRATION_DIR) postgres "$(DB_DSN)" status

migrate-create:
	@test -n "$(name)" || (echo "usage: make migrate-create name=<name>"; exit 1)
	goose -dir $(MIGRATION_DIR) create $(name) sql

mocks:
	mockery

tidy:
	go mod tidy

clean:
	rm -rf bin coverage.out coverage.html
