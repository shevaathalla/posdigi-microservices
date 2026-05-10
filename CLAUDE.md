# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A microservice backend system for employee attendance with clock-in/clockout functionality. Built with **Golang** using the Echo framework, following clean architecture principles.

### Architecture

```
Client ──► Gateway (:8000) ──┬──► Auth Service (:8001)
                             │
                             ├──► User Service (:8002)
                             │
                             └──► Attendance Service (:8003)
```

- **Gateway**: Single entry point, rate limiting (Redis-backed), authentication middleware, reverse proxy routing
- **Auth Service**: User registration, login, JWT generation/validation, bcrypt password hashing
- **User Service**: Employee CRUD operations with pagination, filtering, and soft delete
- **Attendance Service**: Clock-in/clockout, attendance history, work hours calculation

### Databases

- **PostgreSQL**: Users, credentials, roles, attendance records (structured data)
- **MongoDB**: Activity logs, audit trails, user profile metadata (flexible data)
- **Redis (Upstash)**: Caching layer, rate limiting counters

### Tech Stack

- **Framework**: Echo v4 (high-performance Go web framework)
- **ORM**: GORM for PostgreSQL
- **Documentation**: Swagger/OpenAPI (accessible at `/docs`)
- **Containerization**: Docker + Docker Compose
- **Go Workspace**: Uses `go.work` for multi-module project management

## Common Commands

### Development

```bash
# Run individual services locally
make run-auth      # Auth service on :8001
make run-user      # User service on :8002
make run-attendance # Attendance service on :8003
make run-gateway   # Gateway service on :8000

# Run all services (parallel)
make run-all

# Install dependencies
make install-deps

# Tidy Go modules
make tidy

# Generate Swagger documentation
make swagger-generate
```

### Docker Operations

```bash
# Start all services (including databases)
make docker-up

# Stop all services
make docker-down

# Rebuild images (no cache)
make docker-rebuild

# View logs
make docker-logs          # All services
make docker-logs-auth     # Auth service only
make docker-logs-user     # User service only
make docker-logs-gateway  # Gateway only
```

### Database Migrations

```bash
# Run all migrations
make migrate-up

# Rollback all migrations
make migrate-down

# Rollback last migration only
make migrate-down-one

# Redo migrations (down + up)
make migrate-redo

# Check migration status
make migrate-status

# Create new migration
make migrate-create NAME=add_users_table

# Fix dirty database state
make migrate-fix-dirty
```

### Testing

```bash
# Run tests for specific service
make test-auth
make test-user
make test-attendance

# Run all tests
make test-all
```

### Full Setup

```bash
# Complete setup for reviewers (copies .env, starts docker, runs migrations)
make setup
```

## Project Structure

Each service follows this clean architecture pattern:

```
services/{service-name}/
├── main.go              # Entry point with Swagger annotations
├── app.go               # Bootstrap and dependency injection
├── handler/             # HTTP request handlers (controllers)
├── service/             # Business logic layer
├── repository/          # Data access layer
├── dto/                 # Data Transfer Objects (request/response)
├── middleware/          # Echo middleware (auth, logging, etc.)
├── config/              # Configuration and logger setup
├── database/            # Database initialization
├── model/               # Database models (GORM structs)
├── client/              # HTTP clients for inter-service communication
├── Dockerfile           # Service container definition
└── go.mod               # Go module definition
```

### Key Architecture Patterns

1. **Bootstrap Pattern**: Each service has a `Bootstrap()` function in `app.go` that initializes all layers (config → DB → repository → service → handler → router)

2. **Layered Architecture**:
   - Handlers receive HTTP requests and call services
   - Services contain business logic and use repositories
   - Repositories handle database operations
   - DTOs define API contracts

3. **Inter-Service Communication**: Services communicate via HTTP clients (e.g., Auth Service calls User Service via `userClient`)

4. **Middleware Stack**: RequestID → Logger → CORS → GZip → Recover → Route Handler

## Configuration

Services use a shared `.env` file (see `.env.example`):

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=posdigi_microservices
DB_USER=postgres
DB_PASSWORD=postgres

# Service Ports
AUTH_PORT=8001
USER_PORT=8002
GATEWAY_PORT=8000
ATTENDANCE_PORT=8003

# Service URLs (for inter-service communication)
USER_SERVICE_URL=http://localhost:8002
AUTH_SERVICE_URL=http://localhost:8001

# JWT
JWT_SECRET=your-secret-key-change-in-production
JWT_EXPIRY=24

# Internal Service Communication
INTERNAL_SERVICE_KEY=internal-service-key-change-in-production

# Logging
LOG_LEVEL=INFO
ENVIRONMENT=development
```

## Development Workflow

1. **Local Development**: Use `make run-{service}` to start services individually. Each service loads `.env` from project root.

2. **Testing via Gateway**: All external requests should go through Gateway (`:8000`) which proxies to internal services.

3. **API Documentation**: Access Swagger UI at `http://localhost:{port}/docs` for each service.

4. **Health Checks**: Each service has a `/health` endpoint returning service status.

## Go Workspace

This project uses Go 1.25.0 workspace feature (`go.work`) to manage multiple modules:
- `posdigi-auth` (services/auth)
- `posdigi-user` (services/user)
- `posdigi-attendance` (services/attendance)
- `posdigi-gateway` (services/gateway)

Run commands from service directories: `cd services/auth && go run main.go`

## Migrations

Database migrations use `golang-migrate` with SQL files in `migrations/` directory:
- Filename format: `{timestamp}_{description}.up.sql` / `.down.sql`
- Run migrations before starting services
- Use `make migrate-status` to check current version
