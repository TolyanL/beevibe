BINARY := gamegen
PKG := ./...

.PHONY: build run lint test integration-test migrate-up migrate-down migrate-status mocks generate tidy clean

build:
	go build -o bin/$(BINARY) ./cmd/$(BINARY)

run:
	go run ./cmd/$(BINARY)

lint:
	golangci-lint run

test:
	go test -v -cover $(PKG)

integration-test:
	go test -count=1 -v -tags=integration ./test/integration

migrate-up:
	go run ./cmd/migrate -command up

migrate-down:
	go run ./cmd/migrate -command down

migrate-status:
	go run ./cmd/migrate -command status

mocks:
	go generate ./internal/usecase/...

generate:
	go generate ./...

tidy:
	go mod tidy

clean:
	rm -rf bin coverage.out coverage.html
