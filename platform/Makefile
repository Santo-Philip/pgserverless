.PHONY: all build run dev clean test lint

GO := go
GOFLAGS := -ldflags="-s -w"

all: build

build:
	$(GO) build -o bin/gateway $(GOFLAGS) ./gateway
	$(GO) build -o bin/management-api $(GOFLAGS) ./management-api
	$(GO) build -o bin/worker $(GOFLAGS) ./worker
	@echo "Built: gateway, management-api, worker"

run-gateway:
	$(GO) run ./gateway

run-management-api:
	$(GO) run ./management-api

run-worker:
	$(GO) run ./worker

run-all:
	@echo "Starting all services in background..."
	$(GO) run ./gateway &
	$(GO) run ./management-api &
	$(GO) run ./worker &
	@echo "All services started. Press Ctrl+C to stop."

dev:
	@echo "Starting development environment..."
	@echo "1. Start PostgreSQL, Redis, PostgREST via Docker Compose"
	@echo "2. Run services individually: make run-gateway, make run-management-api, make run-worker"
	@echo ""
	$(GO) run ./gateway &

lint:
	$(GO) vet ./...
	$(GO) fmt ./...

test:
	$(GO) test ./... -v -count=1

clean:
	rm -rf bin/
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

.PHONY: db-init
db-init:
	@echo "Initializing database..."
	PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -U $(DB_USER) -d $(DB_NAME) -f postgres/init/001-schema.sql
	PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -U $(DB_USER) -d $(DB_NAME) -f postgres/init/002-functions.sql
	PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -U $(DB_USER) -d $(DB_NAME) -f postgres/init/003-roles.sql
	PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -U $(DB_USER) -d $(DB_NAME) -f postgres/init/004-platform.sql
	@echo "Database initialized."

help:
	@echo "Targets:"
	@echo "  build             - Build all Go services"
	@echo "  run-gateway       - Run gateway service"
	@echo "  run-management-api - Run management API"
	@echo "  run-worker        - Run worker"
	@echo "  run-all           - Run all services"
	@echo "  dev               - Start dev environment"
	@echo "  lint              - Run linters"
	@echo "  test              - Run tests"
	@echo "  clean             - Clean build artifacts"
	@echo "  docker-up         - Start Docker Compose"
	@echo "  docker-down       - Stop Docker Compose"
	@echo "  docker-build      - Build Docker images"
	@echo "  dashboard-dev     - Start dashboard dev server"
	@echo "  dashboard-build   - Build dashboard"
	@echo "  db-init           - Initialize database schema"
