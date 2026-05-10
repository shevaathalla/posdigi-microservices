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

- **Gateway**: Single entry point, rate limiting (in-memory), authentication middleware, reverse proxy routing
- **Auth Service**: User registration, login, JWT generation/validation, bcrypt password hashing, employee profile creation
- **User Service**: User CRUD operations with pagination, filtering, and soft delete; employee profile management with department hierarchy
- **Attendance Service**: Clock-in/clockout, attendance history, work hours calculation, attendance summaries by date range

### Databases

- **PostgreSQL**: Users, employees, credentials, roles, attendance records (structured data)
- **ORM**: GORM for database operations and migrations

### Tech Stack

- **Framework**: Echo v4.15.2 (high-performance Go web framework)
- **ORM**: GORM for PostgreSQL with automated migrations
- **Authentication**: JWT (v5.3.1) with 24-hour expiry, bcrypt password hashing
- **Validation**: Custom validators with go-playground/validator wrapping
- **Documentation**: Swagger/OpenAPI annotations (accessible at `/docs`)
- **Containerization**: Multi-stage Docker builds + Docker Compose orchestration
- **Go Workspace**: Go 1.25.0 workspace feature (`go.work`) for multi-module management
- **Database**: PostgreSQL 15 with golang-migrate for schema management

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
# Run tests for specific service (when tests are implemented)
make test-auth
make test-user
make test-attendance

# Run all tests
make test-all
```

**Note:** Test infrastructure is defined in Makefile but no unit or integration tests are currently implemented in the codebase.

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

## API Endpoints

### Gateway Service (:8000)
- `GET /health` - Aggregated health check of all services
- `POST /api/v1/auth/*` - Public authentication routes (no JWT required)
- `ANY /api/v1/users/*` - Protected user routes (JWT required)
- `ANY /api/v1/employees/*` - Protected employee routes (JWT required)
- `ANY /api/v1/attendance/*` - Protected attendance routes (JWT required)

### Auth Service (:8001)
- `POST /api/v1/auth/register` - User registration with optional employee profile
- `POST /api/v1/auth/login` - User authentication with JWT generation
- `POST /api/v1/auth/validate` - Token validation endpoint
- `GET /health` - Service health check
- `GET /docs/*` - Swagger API documentation

### User Service (:8002)
**User Endpoints:**
- `POST /api/v1/users` - Create new user
- `GET /api/v1/users` - List users (paginated, searchable)
- `GET /api/v1/users/email/:email` - Get user by email
- `POST /api/v1/users/authenticate` - Validate user credentials
- `GET /api/v1/users/:id` - Get user by ID
- `PUT /api/v1/users/:id` - Update user information
- `DELETE /api/v1/users/:id` - Soft delete user

**Employee Endpoints:**
- `POST /api/v1/employees` - Create employee profile
- `GET /api/v1/employees` - List employees (paginated, filterable)
- `GET /api/v1/employees/active` - Get all active employees
- `GET /api/v1/employees/:id` - Get employee by ID
- `GET /api/v1/employees/:id/profile` - Get complete employee profile
- `PUT /api/v1/employees/:id` - Update employee information
- `DELETE /api/v1/employees/:id` - Soft delete employee
- `PATCH /api/v1/employees/:id/status` - Update employment status
- `GET /api/v1/employees/user/:userId` - Get employee by user ID
- `GET /api/v1/employees/code/:code` - Get employee by employee code
- `GET /api/v1/employees/department/:department` - Get employees by department
- `GET /api/v1/employees/manager/:managerId/subordinates` - Get subordinates by manager ID

### Attendance Service (:8003)
- `POST /api/v1/attendance/clock-in` - Record user clock-in
- `POST /api/v1/attendance/clock-out` - Record user clock-out
- `GET /api/v1/attendance/history/:userId` - Get attendance history (paginated)
- `GET /api/v1/attendance/summary/:userId` - Get attendance summary by date range
- `GET /health` - Service health check
- `GET /docs/*` - Swagger API documentation

## Development Workflow

1. **Local Development**: Use `make run-{service}` to start services individually. Each service loads `.env` from project root.

2. **Testing via Gateway**: All external requests should go through Gateway (`:8000`) which proxies to internal services.

3. **API Documentation**: Access Swagger UI at `http://localhost:{port}/docs` for each service.

4. **Health Checks**: Each service has a `/health` endpoint returning service status.

5. **Authentication Flow**:
   - Public routes (register/login) bypass Gateway authentication
   - Protected routes require valid JWT in `Authorization: Bearer <token>` header
   - Inter-service communication uses `X-Service-Auth` header with `INTERNAL_SERVICE_KEY`

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

**Current Migrations:**
1. `20260509130708_create_users_table` - Users table with UUID, email (unique), bcrypt password, role
2. `20260509130709_create_attendances_table` - Attendance records with clock_in/clock_out timestamps
3. `20260510160921_create_employees_table` - Employee profiles with department hierarchy and manager relationships

## Database Schema

**Users Table:**
- UUID primary key with timestamps
- Unique email constraint
- Bcrypt hashed passwords
- Role-based access (user/admin)
- Soft delete support

**Employees Table:**
- UUID primary key with timestamps
- Foreign key relationship to users table
- Unique employee code for business identification
- Department and position information
- Self-referencing manager_id for hierarchical relationships
- Employment status tracking (active/inactive/on_leave)
- JSONB metadata field for flexible data storage
- Soft delete support

**Attendances Table:**
- UUID primary key with timestamps
- Foreign key relationship to users table
- Clock-in and clock-out timestamp tracking
- Nullable clock_out to support active shifts
- Notes field for shift annotations
- Soft delete support

## Middleware Components

**Gateway Middleware Stack:**
1. `RequestIDMiddleware` - Unique request tracking
2. `LoggerMiddleware` - Structured request/response logging
3. `RateLimitMiddleware` - In-memory rate limiting (100 req/min)
4. `AuthMiddleware` - JWT token validation
5. `ProxyHandler` - Reverse proxy routing to backend services

**Service Middleware Stack:**
1. `RequestIDMiddleware` - Request tracing
2. `LoggerMiddleware` - Request logging
3. `CORS` - Cross-origin resource sharing
4. `GZip` - Response compression
5. `Recover` - Panic recovery
6. `CustomValidator` - Request validation with snake_case error messages
7. Route Handler - Service-specific business logic

**Custom Validation Middleware:**
- Wraps go-playground/validator with custom error handling
- Provides human-readable validation error messages
- Supports snake_case JSON tag names for consistent API responses
- Structured validation error responses with field-level details

## Inter-Service Communication

**Service Client Implementations:**
- **Auth → User**: HTTP client for user CRUD operations during registration
- **Gateway → All Services**: Reverse proxy with service authentication
- **Authentication Method**: Shared `INTERNAL_SERVICE_KEY` via `X-Service-Auth` header

**Communication Patterns:**
- Auth Service delegates user storage to User Service
- Services communicate over HTTP with internal authentication
- Service URLs configured via environment variables
- Graceful degradation when services are unavailable

## Current Development Status

**Recently Implemented:**
- Custom validation middleware across all services
- Enhanced error response formatting
- Employee management with hierarchical relationships
- Department-based filtering and organization

**Known Limitations:**
- No unit or integration tests currently implemented
- Rate limiting uses in-memory storage (lost on restart)
- No MongoDB or Redis implementation (PostgreSQL-only)
- No distributed tracing or advanced monitoring
- Limited error recovery mechanisms

**Production Considerations:**
- Change `JWT_SECRET` and `INTERNAL_SERVICE_KEY` in production
- Configure appropriate database connection pooling
- Implement proper logging aggregation
- Add health check monitoring and alerting
- Consider implementing Redis for distributed rate limiting
- Add comprehensive test coverage before production deployment
