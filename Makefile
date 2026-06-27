.PHONY: all build run dev clean test lint migrate docker-up docker-down

GO := go
GOFLAGS := -ldflags="-s -w"

all: build

build:
	$(GO) build -o bin/server $(GOFLAGS) ./cmd/server
	$(GO) build -o bin/migrate $(GOFLAGS) ./cmd/migrate
	@echo "Built: server, migrate"

run:
	$(GO) run ./cmd/server

migrate:
	$(GO) run ./cmd/migrate

dev:
	$(GO) run ./cmd/server &

lint:
	$(GO) vet ./...

test:
	$(GO) test ./... -v -count=1 -timeout 60s

test-cover:
	$(GO) test ./... -coverprofile=coverage.out -count=1 -timeout 60s
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

clean:
	rm -rf bin/ coverage.out coverage.html
	$(GO) clean -cache

docker-up:
	docker compose -f docker-compose.yml up -d

docker-down:
	docker compose -f docker-compose.yml down

docker-build:
	docker compose -f docker-compose.yml build

help:
	@echo "Targets:"
	@echo "  build             - Build all Go services"
	@echo "  run               - Run server"
	@echo "  migrate           - Run database migrations"
	@echo "  dev               - Start dev environment"
	@echo "  lint              - Run linters (go vet)"
	@echo "  test              - Run tests"
	@echo "  test-cover        - Run tests with coverage"
	@echo "  clean             - Clean build artifacts"
	@echo "  docker-up         - Start Docker Compose"
	@echo "  docker-down       - Stop Docker Compose"
	@echo "  docker-build      - Build Docker images"
