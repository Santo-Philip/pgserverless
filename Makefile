.PHONY: all build run dev clean test lint migrate

GO := go
GOFLAGS := -ldflags="-s -w"

all: build

build:
	$(GO) build -o bin/gateway $(GOFLAGS) ./gateway
	$(GO) build -o bin/management-api $(GOFLAGS) ./management-api
	$(GO) build -o bin/worker $(GOFLAGS) ./worker
	$(GO) build -o bin/migrate $(GOFLAGS) ./cmd/migrate
	@echo "Built: gateway, management-api, worker, migrate"

run-gateway:
	$(GO) run ./gateway

run-management-api:
	$(GO) run ./management-api

run-worker:
	$(GO) run ./worker

migrate:
	$(GO) run ./cmd/migrate

run-all:
	@echo "Starting all services in background..."
	$(GO) run ./gateway &
	$(GO) run ./management-api &
	$(GO) run ./worker &
	@echo "All services started. Press Ctrl+C to stop."

dev:
	@echo "Starting development environment..."
	$(GO) run ./gateway &

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

.PHONY: dashboard-dev dashboard-build
dashboard-dev:
	cd dashboard && npm run dev

dashboard-build:
	cd dashboard && npm run build

.PHONY: db-migrate
db-migrate:
	@echo "Running database migrations..."
	PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -U $(DB_USER) -d $(DB_NAME) -f postgres/init/001-schema.sql
	PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -U $(DB_USER) -d $(DB_NAME) -f postgres/init/002-functions.sql
	PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -U $(DB_USER) -d $(DB_NAME) -f postgres/init/003-roles.sql
	PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -U $(DB_USER) -d $(DB_NAME) -f postgres/init/004-platform.sql
	PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -U $(DB_USER) -d $(DB_NAME) -f postgres/init/005-domains.sql
	PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -U $(DB_USER) -d $(DB_NAME) -f postgres/init/006-super-admin.sql
	PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -U $(DB_USER) -d $(DB_NAME) -f postgres/init/007-platform-settings.sql
	@echo "Database migrations complete."

help:
	@echo "Targets:"
	@echo "  build             - Build all Go services"
	@echo "  run-gateway       - Run gateway service"
	@echo "  run-management-api - Run management API"
	@echo "  run-worker        - Run worker"
	@echo "  migrate           - Run database migrations (Go tool)"
	@echo "  run-all           - Run all services"
	@echo "  dev               - Start dev environment"
	@echo "  lint              - Run linters (go vet)"
	@echo "  test              - Run tests"
	@echo "  test-cover        - Run tests with coverage"
	@echo "  clean             - Clean build artifacts"
	@echo "  docker-up         - Start Docker Compose"
	@echo "  docker-down       - Stop Docker Compose"
	@echo "  docker-build      - Build Docker images"
	@echo "  dashboard-dev     - Start dashboard dev server"
	@echo "  dashboard-build   - Build dashboard"
	@echo "  db-migrate        - Initialize database schema (psql)"
