# Posdigi Microservices

A Go-based microservices system for employee and attendance management, built with Echo, GORM, PostgreSQL, MongoDB, and JWT authentication.

## 🚀 Quick Start

```bash
# Clone the repository
git clone <repository-url>
cd posdigi-microservices

# Start all services (Docker required)
make docker-up

# Access the API via Gateway
curl http://localhost:8000/health

# View Swagger documentation
# Auth Service:  http://localhost:8001/docs/index.html
# User Service:  http://localhost:8002/docs/index.html
# Attendance Service: http://localhost:8003/docs/index.html
# Gateway:        http://localhost:8000/docs/index.html
```

## 📋 Production Deployment

**📖 Complete Deployment Guide:** [DEPLOYMENT.md](DEPLOYMENT.md)
**🔒 Security Checklist:** [SECURITY_CHECKLIST.md](SECURITY_CHECKLIST.md)

**Quick Deploy:**
```bash
# 1. Prepare production environment
cp .env.example .env
# Edit .env with production values

# 2. Deploy to VPS
scp -r . root@your-vps-ip:/var/www/posdigi-microservices/

# 3. On VPS
cd /var/www/posdigi-microservices
make docker-up
```

## Architecture Overview

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ HTTP
       ▼
┌─────────────────┐
│  Gateway :8000  │  ← Rate limiting, JWT auth, reverse proxy
└────────┬────────┘
         │
    ┌────┴──────────────┬────────────────┐
    ▼                   ▼                ▼
┌─────────┐      ┌──────────┐    ┌──────────────┐
│  Auth   │      │   User   │    │  Attendance  │
│  :8001  │─────▶│  :8002   │    │    :8003     │
└─────────┘      └──────────┘    └──────────────┘
    │                  │                 │
    └──────────────────┴─────────┬───────┘
                              ▼
                    ┌──────────────────────┐
                    │    PostgreSQL :5432   │
                    │  (Users, Employees)   │
                    └──────────────────────┘
                              │
                    ┌──────────────────────┐
                    │    MongoDB :27017     │
                    │  (Activity Logs)      │
                    └──────────────────────┘
```

| Service    | Port | Description |
|------------|------|-------------|
| Gateway    | 8000 | API gateway — routes, rate limits, JWT validation |
| Auth       | 8001 | Registration, login, JWT token issuance |
| User       | 8002 | User profiles, employee records |
| Attendance | 8003 | Clock-in/out, attendance history |

## Services

### Gateway Service
- Transparent reverse proxy to backend services
- JWT authentication middleware for protected routes
- Rate limiting
- CORS handling
- Health check aggregation (`GET /health`)

**Public routes** (no token required):
```
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/validate
GET  /api/v1/auth/validate
```

**Protected routes** (JWT required):
```
ANY  /api/v1/users/*
ANY  /api/v1/attendance/*
```

---

### Auth Service
Handles user registration and login. Delegates user storage and credential validation to the User Service. Issues JWT tokens.

**Endpoints:**
```
POST /api/v1/auth/register   — Register new user (+ optional employee profile)
POST /api/v1/auth/login      — Login, returns JWT token
POST /api/v1/auth/validate   — Validate JWT token
GET  /api/v1/auth/validate   — Validate JWT token (query param)
GET  /health
```

**Registration flow:**
1. Auth Service checks if user exists (via User Service `GET /users/email/:email`)
2. Creates user in User Service (`POST /users`)
3. Creates employee profile if `employee_data` provided (`POST /employees`)
4. Generates and returns JWT token

**Register request body:**
```json
{
  "email": "john@example.com",
  "password": "secret123",
  "employee_data": {
    "full_name": "John Doe",
    "phone": "08123456789",
    "department": "Engineering",
    "position": "Backend Developer",
    "salary": 8000000,
    "hire_date": "2026-01-15",
    "employment_status": "active",
    "manager_id": null,
    "emergency_contact": "Jane Doe",
    "emergency_phone": "08987654321",
    "address": "Jakarta, Indonesia"
  }
}
```

**Login request body:**
```json
{
  "email": "john@example.com",
  "password": "secret123"
}
```

---

### User Service
Owns all user and employee data. Handles password hashing (bcrypt). Exposes internal API secured with `X-Service-Auth` header.

**User endpoints:**
```
POST   /api/v1/users                    — Create user
GET    /api/v1/users                    — List users (paginated)
GET    /api/v1/users/email/:email       — Get user by email
POST   /api/v1/users/authenticate       — Validate credentials (used by auth service)
GET    /api/v1/users/:id                — Get user by ID
PUT    /api/v1/users/:id                — Update user
DELETE /api/v1/users/:id                — Delete user (soft delete)
```

**Employee endpoints:**
```
POST   /api/v1/employees                          — Create employee profile
GET    /api/v1/employees                          — List employees (paginated)
GET    /api/v1/employees/active                   — List active employees
GET    /api/v1/employees/:id                      — Get employee by ID
GET    /api/v1/employees/:id/profile              — Get full employee + user profile
PUT    /api/v1/employees/:id                      — Update employee
DELETE /api/v1/employees/:id                      — Delete employee (soft delete)
PATCH  /api/v1/employees/:id/status               — Update employment status
GET    /api/v1/employees/user/:userId             — Get employee by user ID
GET    /api/v1/employees/code/:code               — Get employee by code (e.g. EMP001)
GET    /api/v1/employees/department/:department   — List by department
GET    /api/v1/employees/manager/:managerId/subordinates — Get subordinates
```

**Query params for list endpoints:** `?page=1&page_size=10&search=keyword`

---

### Attendance Service
Tracks clock-in and clock-out records per user.

**Endpoints:**
```
POST /api/v1/attendance/clock-in          — Record clock-in
POST /api/v1/attendance/clock-out         — Record clock-out
GET  /api/v1/attendance/history/:userId   — Attendance history
GET  /api/v1/attendance/summary/:userId   — Attendance summary
GET  /health
```

---

## Database Schema

All services share a single PostgreSQL database (`attendance_db`).

### `users` table
| Column     | Type         | Notes |
|------------|--------------|-------|
| id         | UUID (PK)    | Auto-generated |
| email      | VARCHAR(255) | Unique |
| password   | VARCHAR(255) | bcrypt hashed, never returned in JSON |
| full_name  | VARCHAR(255) | |
| role       | VARCHAR(50)  | Default: `user` |
| created_at | TIMESTAMP    | |
| updated_at | TIMESTAMP    | |
| deleted_at | TIMESTAMP    | Soft delete |

### `employees` table
| Column            | Type          | Notes |
|-------------------|---------------|-------|
| id                | UUID (PK)     | Default: `uuid_generate_v4()` |
| user_id           | UUID (FK)     | References `users.id`, unique |
| employee_code     | VARCHAR(20)   | Auto-generated: EMP001, EMP002... |
| full_name         | VARCHAR(100)  | |
| phone             | VARCHAR(20)   | |
| department        | VARCHAR(50)   | |
| position          | VARCHAR(50)   | |
| salary            | DECIMAL(10,2) | |
| hire_date         | DATE          | |
| employment_status | VARCHAR(20)   | `active`, `terminated`, `on_leave`, `suspended` |
| manager_id        | UUID (FK)     | Self-referencing, references `employees.id` |
| emergency_contact | VARCHAR(100)  | |
| emergency_phone   | VARCHAR(20)   | |
| address           | TEXT          | |
| profile_image     | VARCHAR(255)  | |
| metadata          | JSONB         | |
| created_at        | TIMESTAMP     | |
| updated_at        | TIMESTAMP     | |
| deleted_at        | TIMESTAMP     | Soft delete |

### `attendances` table
| Column     | Type        | Notes |
|------------|-------------|-------|
| id         | VARCHAR(36) | UUID |
| user_id    | VARCHAR(36) | References `users.id` |
| clock_in   | TIMESTAMP   | |
| clock_out  | TIMESTAMP   | Nullable |
| notes      | TEXT        | |
| created_at | TIMESTAMP   | |
| updated_at | TIMESTAMP   | |
| deleted_at | TIMESTAMP   | Soft delete |

---

## Getting Started

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- `golang-migrate` CLI (for migrations)
- `make`

### Quick Setup

```bash
# 1. Clone the repo
git clone <repo-url>
cd posdigi-microservices

# 2. Copy and configure environment
cp .env.example .env

# 3. One-command setup (Docker + migrations)
make setup
```

This starts all services and runs all migrations. The gateway will be available at `http://localhost:8000`.

---

## Environment Variables

Copy `.env.example` to `.env` and adjust values:

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=posdigi_microservices
DB_USER=postgres
DB_PASSWORD=postgres

# Ports
AUTH_PORT=8001
USER_PORT=8002
ATTENDANCE_PORT=8003
GATEWAY_PORT=8000

# Service URLs (inter-service communication)
USER_SERVICE_URL=http://localhost:8002
AUTH_SERVICE_URL=http://localhost:8001
ATTENDANCE_SERVICE_URL=http://localhost:8003

# JWT
JWT_SECRET=your-secret-key-change-in-production
JWT_EXPIRY=24

# Security
INTERNAL_SERVICE_KEY=internal-service-key-change-in-production

# Logging
LOG_LEVEL=INFO
ENVIRONMENT=development
```

> ⚠️ **Security:** Always change `JWT_SECRET` and `INTERNAL_SERVICE_KEY` in production.

---

## Running Locally (without Docker)

Start each service individually (requires PostgreSQL running locally):

```bash
make run-auth        # Auth service on :8001
make run-user        # User service on :8002
make run-attendance  # Attendance service on :8003
make run-gateway     # Gateway on :8000

# Or all at once (parallel)
make run-all
```

---

## Docker

```bash
make docker-up        # Start all services
make docker-down      # Stop all services
make docker-build     # Build images
make docker-rebuild   # Rebuild from scratch (no cache)
make docker-logs      # Tail all logs
make docker-logs-auth # Tail auth service logs
make docker-ps        # Show container status
```

---

## Database Migrations

Uses [golang-migrate](https://github.com/golang-migrate/migrate).

```bash
make migrate-up           # Apply all pending migrations
make migrate-down         # Roll back all migrations
make migrate-down-one     # Roll back last migration
make migrate-redo         # Roll back all + re-apply
make migrate-status       # Show current version

# Create a new migration
make migrate-create NAME=add_departments_table

# Fix dirty migration state
make migrate-fix-dirty
```

Migration files are in `migrations/` and apply to all services (shared DB).

---

## API Response Format

All endpoints return a consistent JSON envelope:

**Success:**
```json
{
  "success": true,
  "message": "Operation successful",
  "data": { ... }
}
```

**Error:**
```json
{
  "success": false,
  "message": "Human-readable error message"
}
```

---

## Authentication

### Getting a token
```bash
curl -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "john@example.com", "password": "secret123"}'
```

### Using a token
```bash
curl http://localhost:8000/api/v1/users \
  -H "Authorization: Bearer <token>"
```

### Token payload
```json
{
  "user_id": "uuid",
  "email": "john@example.com",
  "role": "user",
  "exp": 1234567890
}
```

---

## Inter-Service Communication

Backend services authenticate each other using the `X-Service-Auth` header:
```
X-Service-Auth: <INTERNAL_SERVICE_KEY>
```

The Auth Service communicates with the User Service via HTTP. It **never** accesses the database directly — all user and employee CRUD is delegated to the User Service.

---

## Project Structure

```
posdigi-microservices/
├── services/
│   ├── auth/               # Auth service
│   │   ├── client/         # HTTP client for User Service
│   │   ├── config/
│   │   ├── dto/
│   │   ├── handler/
│   │   ├── middleware/
│   │   ├── router/
│   │   └── service/
│   ├── user/               # User service
│   │   ├── config/
│   │   ├── database/
│   │   ├── dto/
│   │   ├── handler/
│   │   ├── middleware/
│   │   ├── model/
│   │   ├── repository/
│   │   ├── router/
│   │   └── service/
│   ├── attendance/         # Attendance service
│   │   ├── dto/
│   │   ├── handler/
│   │   ├── middleware/
│   │   ├── model/
│   │   ├── repository/
│   │   ├── router/
│   │   └── service/
│   └── gateway/            # API gateway
│       ├── client/         # Service proxy client
│       ├── config/
│       ├── handler/        # Proxy handler
│       ├── middleware/
│       ├── router/
│       └── service/        # Health checker
├── migrations/             # SQL migration files
├── docker-compose.yml
├── Makefile
├── go.work                 # Go workspace
└── .env.example
```

---

## Make Commands Reference

| Command | Description |
|---------|-------------|
| `make setup` | Full setup: copy env, start Docker, run migrations |
| `make run-all` | Run all services locally |
| `make docker-up` | Start all Docker services |
| `make docker-down` | Stop all Docker services |
| `make migrate-up` | Apply all migrations |
| `make migrate-down` | Roll back all migrations |
| `make migrate-create NAME=x` | Create a new migration |
| `make test-all` | Run all tests |
| `make tidy` | `go mod tidy` all services |
| `make install-deps` | Download all Go dependencies |
| `make swagger-generate` | Generate Swagger docs |
| `make db-shell` | Open PostgreSQL shell |
| `make db-backup` | Backup the database |
| `make clean` | Remove build artifacts |

---

## Health Checks

> ⚠️ Health endpoints are at the **root path** — NOT under `/api/v1/`. Calling `/api/v1/users/health` will hit the JWT-protected group and return `401 Missing authorization header`.

| URL | Auth required | Description |
|-----|---------------|-------------|
| `GET http://localhost:8000/health` | ❌ None | Gateway health — aggregates all services |
| `GET http://localhost:8001/health` | ❌ None | Auth service health (direct) |
| `GET http://localhost:8002/health` | ❌ None | User service health (direct) |
| `GET http://localhost:8003/health` | ❌ None | Attendance service health (direct) |

```bash
curl http://localhost:8000/health
```

```json
{
  "success": true,
  "service": "gateway",
  "status": "healthy",
  "services": {
    "auth": true,
    "user": true,
    "attendance": true
  }
}
```
