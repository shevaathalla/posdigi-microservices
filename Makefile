# Makefile for Posdigi Microservices
# Load environment variables from .env
include .env

##@ General

.PHONY: help
help: ## Display this help message
	@echo "Posdigi Microservices - Available Commands:"
	@echo ""
	@grep -E '^### ?@[-a-zA-Z]+' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$2, $$3}'
	@echo ""

##@ Development

.PHONY: run-auth
run-auth: ## Run auth service locally
	@echo "Starting Auth Service on port $(AUTH_PORT)..."
	@cd services/auth && go run main.go

.PHONY: run-user
run-user: ## Run user service locally
	@echo "Starting User Service on port $(USER_PORT)..."
	@cd services/user && go run main.go

.PHONY: run-attendance
run-attendance: ## Run attendance service locally
	@echo "Starting Attendance Service on port $(ATTENDANCE_PORT)..."
	@cd services/attendance && go run main.go

.PHONY: run-gateway
run-gateway: ## Run gateway service locally
	@echo "Starting Gateway Service on port $(GATEWAY_PORT)..."
	@cd services/gateway && go run main.go

.PHONY: run-all
run-all: ## Run all services locally (in background)
	@echo "Starting all services..."
	@make -j run-auth run-user run-attendance run-gateway

##@ Database Migrations

.PHONY: migrate-up
migrate-up: ## Run all UP migrations
	@echo "Running all migrations..."
	@migrate -path migrations -database "$(DATABASE_URL)" up
	@echo "✅ All migrations completed!"

.PHONY: migrate-down
migrate-down: ## Run all DOWN migrations
	@echo "Rolling back all migrations..."
	@migrate -path migrations -database "$(DATABASE_URL)" down
	@echo "✅ All migrations rolled back!"

.PHONY: migrate-down-one
migrate-down-one: ## Rollback only the last migration
	@echo "Rolling back last migration..."
	@migrate -path migrations -database "$(DATABASE_URL)" down 1
	@echo "✅ Last migration rolled back!"

.PHONY: migrate-redo
migrate-redo: ## Redo all migrations (down + up)
	@echo "Redoing all migrations..."
	@echo "Step 1: Rolling back all migrations..."
	@migrate -path migrations -database "$(DATABASE_URL)" down
	@echo "Step 2: Re-applying all migrations..."
	@migrate -path migrations -database "$(DATABASE_URL)" up
	@echo "✅ All migrations redone!"

.PHONY: migrate-redo-one
migrate-redo-one: ## Redo the last migration (down 1 + up)
	@echo "Redoing last migration..."
	@echo "Step 1: Rolling back last migration..."
	@migrate -path migrations -database "$(DATABASE_URL)" down 1
	@echo "Step 2: Re-applying migrations..."
	@migrate -path migrations -database "$(DATABASE_URL)" up
	@echo "✅ Last migration redone!"

.PHONY: migrate-create
migrate-create: ## Create new migration (usage: make migrate-create NAME=add_users_table)
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME parameter required"; \
		echo "Usage: make migrate-create NAME=add_users_table"; \
		exit 1; \
	fi
	@echo "Creating migration: $(NAME)"
	@migrate create -ext sql -dir migrations $(NAME)
	@echo "✅ Migration created!"

.PHONY: migrate-status
migrate-status: ## Show migration status
	@echo "Checking migration status..."
	@migrate -path migrations -database "$(DATABASE_URL)" version

.PHONY: migrate-force
migrate-force: ## Force migration version (usage: make migrate-force VERSION=1)
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION parameter required"; \
		echo "Usage: make migrate-force VERSION=1"; \
		exit 1; \
	fi
	@migrate -path migrations -database "$(DATABASE_URL)" force $(VERSION)

.PHONY: migrate-fix-dirty
migrate-fix-dirty: ## Fix dirty database state (interactive - asks for version)
	@echo "⚠️  Database is in a DIRTY state - a migration failed!"
	@echo ""
	@echo "To fix this, you need to:"
	@echo "1. Check the error and understand what went wrong"
	@echo "2. Manually fix the database state if needed"
	@echo "3. Force the version to the last successful migration"
	@echo ""
	@echo "Current migration status:"
	@migrate -path migrations -database "$(DATABASE_URL)" version
	@echo ""
	@read -p "Enter the version number to force (or press Ctrl+C to cancel): " version; \
	echo ""; \
	echo "Forcing database to version $$version..."; \
	migrate -path migrations -database "$(DATABASE_URL)" force $$version; \
	echo "✅ Dirty state fixed! You can now run migrations again."

.PHONY: migrate-diagnostics
migrate-diagnostics: ## Run migration diagnostics to check for issues
	@echo "=== Migration Diagnostics ==="
	@echo ""
	@echo "1. Checking migration status:"
	@migrate -path migrations -database "$(DATABASE_URL)" version
	@echo ""
	@echo "2. Listing migration files:"
	@ls -la migrations/*.sql 2>/dev/null | grep -E '\.(up|down)\.sql$$' || echo "No migration files found"
	@echo ""
	@echo "3. Checking for dirty state:"
	@migrate -path migrations -database "$(DATABASE_URL)" version 2>&1 | grep -i dirty && echo "⚠️  DATABASE IS DIRTY!" || echo "✅ Database is clean"
	@echo ""
	@echo "4. Testing database connection:"
	@migrate -path migrations -database "$(DATABASE_URL)" version 2>&1 | head -1

##@ Docker

.PHONY: docker-up
docker-up: ## Start all Docker services
	@echo "Starting Docker services..."
	docker-compose up -d
	@echo "✅ Docker services started!"
	@echo "Gateway: http://localhost:$(GATEWAY_PORT)"
	@echo "Swagger Docs: http://localhost:$(GATEWAY_PORT)/docs"

.PHONY: docker-down
docker-down: ## Stop all Docker services
	@echo "Stopping Docker services..."
	docker-compose down
	@echo "✅ Docker services stopped!"

.PHONY: docker-build
docker-build: ## Build all Docker images
	@echo "Building Docker images..."
	docker-compose build
	@echo "✅ Docker images built!"

.PHONY: docker-rebuild
docker-rebuild: ## Rebuild and restart all services
	@echo "Rebuilding Docker images..."
	docker-compose build --no-cache
	docker-compose up -d
	@echo "✅ Services rebuilt and restarted!"

.PHONY: docker-logs
docker-logs: ## Show logs from all services
	docker-compose logs -f

.PHONY: docker-logs-auth
docker-logs-auth: ## Show logs from auth service
	docker-compose logs -f auth

.PHONY: docker-logs-user
docker-logs-user: ## Show logs from user service
	docker-compose logs -f user

.PHONY: docker-logs-attendance
docker-logs-attendance: ## Show logs from attendance service
	docker-compose logs -f attendance

.PHONY: docker-logs-gateway
docker-logs-gateway: ## Show logs from gateway service
	docker-compose logs -f gateway

.PHONY: docker-ps
docker-ps: ## Show running Docker containers
	docker-compose ps

##@ Database

.PHONY: db-reset
db-reset: ## Drop and recreate database (DEV ONLY!)
	@echo "⚠️  WARNING: This will delete all data!"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		echo "Dropping database..."; \
		docker-compose down -v; \
		echo "Starting fresh database..."; \
		docker-compose up -d postgres mongo redis; \
		echo "Waiting for databases to be ready..."; \
		sleep 5; \
		echo "Running migrations..."; \
		make migrate-up; \
		echo "✅ Database reset complete!"; \
	else \
		echo "Cancelled."; \
	fi

.PHONY: db-shell
db-shell: ## Open PostgreSQL shell
	docker-compose exec postgres psql -U $(DB_USER) -d $(DB_NAME)

.PHONY: db-backup
db-backup: ## Backup PostgreSQL database
	@echo "Backing up database..."
	@docker-compose exec postgres pg_dump -U $(DB_USER) $(DB_NAME) > backup_$$(date +%Y%m%d_%H%M%S).sql
	@echo "✅ Backup complete!"

##@ Setup

.PHONY: setup
setup: ## Full setup for reviewers (docker + migrations)
	@echo "Setting up Posdigi Microservices..."
	@echo "→ Copying environment file"
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "✅ .env created from .env.example"; \
	fi
	@echo "→ Starting Docker services"
	@make docker-up
	@echo "→ Waiting for services to be ready"
	@sleep 5
	@echo "→ Running migrations"
	@make migrate-up
	@echo ""
	@echo "✅ Setup complete!"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Gateway:       http://localhost:$(GATEWAY_PORT)"
	@echo "Swagger Docs:  http://localhost:$(GATEWAY_PORT)/docs"
	@echo "Auth Service:  http://localhost:$(AUTH_PORT)"
	@echo "User Service:  http://localhost:$(USER_PORT)"
	@echo "Attendance:    http://localhost:$(ATTENDANCE_PORT)"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

##@ Testing

.PHONY: test-auth
test-auth: ## Run auth service tests
	@echo "Running Auth Service tests..."
	@cd services/auth && go test -v ./...

.PHONY: test-user
test-user: ## Run user service tests
	@echo "Running User Service tests..."
	@cd services/user && go test -v ./...

.PHONY: test-attendance
test-attendance: ## Run attendance service tests
	@echo "Running Attendance Service tests..."
	@cd services/attendance && go test -v ./...

.PHONY: test-all
test-all: ## Run all tests
	@echo "Running all tests..."
	@make test-auth test-user test-attendance

##@ Utilities

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@find . -name "*.log" -delete
	@find . -name "*.tmp" -delete
	@echo "✅ Cleaned!"

.PHONY: install-deps
install-deps: ## Install all Go dependencies
	@echo "Installing dependencies for all services..."
	@cd services/auth && go mod download
	@cd services/user && go mod download
	@cd services/attendance && go mod download
	@cd services/gateway && go mod download
	@echo "✅ Dependencies installed!"

.PHONY: tidy
tidy: ## Tidy all go.mod files
	@echo "Tidying Go modules..."
	@cd services/auth && go mod tidy
	@cd services/user && go mod tidy
	@cd services/attendance && go mod tidy
	@cd services/gateway && go mod tidy
	@echo "✅ All modules tidied!"

.PHONY: swagger-generate
swagger-generate: ## Generate Swagger docs for all services
	@echo "Generating Swagger documentation..."
	@cd services/auth && swag init
	@cd services/user && swag init
	@cd services/attendance && swag init
	@cd services/gateway && swag init
	@echo "✅ Swagger docs generated!"
