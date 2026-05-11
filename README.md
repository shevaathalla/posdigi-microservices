# Posdigi Microservices

A Go-based microservices system for employee and attendance management, built with Echo, GORM, PostgreSQL, MongoDB, and JWT authentication.

## ✨ Key Features

- **🔐 Secure Authentication** - JWT-based auth with bcrypt password hashing
- **👥 User Management** - Complete CRUD operations with employee profiles
- **⏰ Attendance Tracking** - Clock-in/clockout with history and summaries
- **📊 Activity Logging** - MongoDB-powered audit trail for all user actions
- **🚀 High Performance** - Clean architecture with proper separation of concerns
- **🐳 Production Ready** - Docker setup with health checks and auto-restart
- **📖 Well Documented** - Comprehensive API documentation with Swagger

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
**🎯 Deployment Readiness:** [DEPLOYMENT_READINESS.md](DEPLOYMENT_READINESS.md)

**Quick Deploy:**
```bash
# 1. Generate production secrets
./generate-secrets.sh

# 2. Prepare production environment
cp .env.production.example .env
# Add generated secrets to .env

# 3. Deploy to VPS
scp -r . root@your-vps-ip:/var/www/posdigi-microservices/

# 4. On VPS
cd /var/www/posdigi-microservices
docker-compose up -d --build
```

**Production Features:**
- ✅ Health checks with automatic restart
- ✅ MongoDB and PostgreSQL with proper indexing
- ✅ Activity logging for security auditing
- ✅ Rate limiting and request tracking
- ✅ SSL/HTTPS support
- ✅ Automated backup scripts
- ✅ Monitoring and alerting setup

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

**Required fields:** `email`, `password`
**Optional fields:** `employee_data` (object)
**Note:** If `employee_data` is provided, `full_name` and `hire_date` (format: YYYY-MM-DD) are required

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

**Create employee request:**
```json
{
  "user_id": "user-uuid-here",
  "full_name": "John Doe",
  "phone": "08123456789",
  "department": "Engineering",
  "position": "Software Developer",
  "salary": 75000,
  "hire_date": "2024-01-15",
  "employment_status": "active",
  "emergency_contact": "Jane Doe",
  "emergency_phone": "08987654321",
  "address": "Jakarta, Indonesia"
}
```

**Update employee status request:**
```json
{
  "employment_status": "active"
}
```
**Valid status values:** `active`, `terminated`, `on_leave`, `suspended`

---

### Attendance Service
Tracks clock-in and clock-out records per user.

**Endpoints:**
```
POST /api/v1/attendance/clock-in          — Record clock-in
POST /api/v1/attendance/clock-out         — Record clock-out
GET  /api/v1/attendance/history/:userId   — Attendance history
GET  /api/v1/attendance/summary/:userId   — Attendance summary (requires date range)
GET  /health
```

**Clock-in request:**
```json
{
  "user_id": "user-uuid-here",
  "note": "Starting morning shift"
}
```

**Clock-out request:**
```json
{
  "attendance_id": "attendance-uuid-here",
  "note": "Ending shift"
}
```

**Attendance history query parameters:**
- `page` (optional, default: 1)
- `limit` (optional, default: 10, max: 100)

**Attendance summary query parameters (required):**
- `start_date` (required, format: YYYY-MM-DD) - example: `2024-01-01`
- `end_date` (required, format: YYYY-MM-DD) - example: `2024-12-31`

**Example attendance summary request:**
```bash
curl "http://localhost:8000/api/v1/attendance/summary/user-uuid-here?start_date=2024-01-01&end_date=2024-12-31" \
  -H "Authorization: Bearer your-token"
```

---

## API Usage Examples

### Authentication Examples

**Register a new user with employee profile:**
```bash
curl -X POST http://localhost:8000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com",
    "password": "securePassword123",
    "employee_data": {
      "full_name": "John Doe",
      "phone": "08123456789",
      "department": "Engineering",
      "position": "Backend Developer",
      "salary": 8000000,
      "hire_date": "2024-01-15",
      "employment_status": "active"
    }
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "user": {
      "id": "uuid-here",
      "email": "john.doe@example.com",
      "role": "user"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**Login:**
```bash
curl -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com",
    "password": "securePassword123"
  }'
```

### User Management Examples

**List all users with pagination:**
```bash
curl "http://localhost:8000/api/v1/users?page=1&limit=10&search=john" \
  -H "Authorization: Bearer your-token-here"
```

**Update user information:**
```bash
curl -X PUT http://localhost:8000/api/v1/users/user-uuid-here \
  -H "Authorization: Bearer your-token-here" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe.updated@example.com",
    "password": "newSecurePassword123"
  }'
```

### Employee Management Examples

**Create employee profile:**
```bash
curl -X POST http://localhost:8000/api/v1/employees \
  -H "Authorization: Bearer your-token-here" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-uuid-here",
    "full_name": "John Doe",
    "phone": "08123456789",
    "department": "Engineering",
    "position": "Backend Developer",
    "salary": 8000000,
    "hire_date": "2024-01-15",
    "employment_status": "active",
    "emergency_contact": "Jane Doe",
    "emergency_phone": "08987654321",
    "address": "Jakarta, Indonesia"
  }'
```

**Update employee status:**
```bash
curl -X PATCH http://localhost:8000/api/v1/employees/employee-uuid-here/status \
  -H "Authorization: Bearer your-token-here" \
  -H "Content-Type: application/json" \
  -d '{
    "employment_status": "on_leave"
  }'
```

**Get employees by department:**
```bash
curl "http://localhost:8000/api/v1/employees/department/Engineering?page=1&page_size=10" \
  -H "Authorization: Bearer your-token-here"
```

### Attendance Examples

**Clock-in:**
```bash
curl -X POST http://localhost:8000/api/v1/attendance/clock-in \
  -H "Authorization: Bearer your-token-here" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-uuid-here",
    "note": "Starting morning shift"
  }'
```

**Clock-out:**
```bash
curl -X POST http://localhost:8000/api/v1/attendance/clock-out \
  -H "Authorization: Bearer your-token-here" \
  -H "Content-Type: application/json" \
  -d '{
    "attendance_id": "attendance-uuid-here",
    "note": "Completed work day"
  }'
```

**Get attendance history:**
```bash
curl "http://localhost:8000/api/v1/attendance/history/user-uuid-here?page=1&limit=10" \
  -H "Authorization: Bearer your-token-here"
```

**Get attendance summary (with required date parameters):**
```bash
curl "http://localhost:8000/api/v1/attendance/summary/user-uuid-here?start_date=2024-01-01&end_date=2024-12-31" \
  -H "Authorization: Bearer your-token-here"
```

### Error Response Examples

**Validation error response:**
```json
{
  "success": false,
  "message": "Validation failed: user_id is required. Required fields: user_id, full_name, hire_date (format: YYYY-MM-DD)"
}
```

**Authentication error response:**
```json
{
  "success": false,
  "message": "Invalid or expired token"
}
```

**Not found error response:**
```json
{
  "success": false,
  "message": "Employee not found"
}
```

---

## Error Messages and Solutions

### Common Error Messages:

| Error Message | Cause | Solution |
|---------------|-------|----------|
| "Status is required" | Employee status update missing employment_status field | Include `{"employment_status": "active|terminated|on_leave|suspended"}` in request body |
| "start_date query parameter is required" | Attendance summary missing date parameters | Add `?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD` to URL |
| "Invalid date format" | Date not in YYYY-MM-DD format | Use format like `2024-01-15` for dates |
| "User already exists" | Email already registered | Use a different email or login with existing account |
| "Invalid email or password" | Wrong credentials | Check email and password, or register new account |
| "Employee not found" | No employee profile for user | Create employee profile first |
| "User already has an active clock-in" | Attempting to clock-in while already clocked-in | Clock-out first before new clock-in |

---

## Database Schema

All services share a single PostgreSQL database (`posdigi_microservices`) and use MongoDB for activity logging.

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

### MongoDB Activity Logs Collection

The system uses MongoDB for comprehensive activity logging and audit trails:

**`activity_logs` collection structure:**
| Field          | Type    | Description |
|---------------|---------|-------------|
| _id           | ObjectId | Auto-generated MongoDB ID |
| user_id       | UUID    | Reference to user who performed action |
| employee_id   | UUID    | Reference to employee (if applicable) |
| service       | String  | Service name (auth, user, attendance, gateway) |
| action        | String  | Action type (LOGIN_SUCCESS, USER_CREATED, CLOCK_IN, etc.) |
| endpoint      | String  | API endpoint called |
| method        | String  | HTTP method (GET, POST, PUT, DELETE) |
| ip_address    | String  | Client IP address |
| user_agent    | String  | Client user agent |
| request_id    | String  | Request ID for distributed tracing |
| status_code   | Int     | HTTP response status |
| success       | Boolean | Action success status |
| error_message | String  | Error details if failed |
| metadata      | Object  | Additional data (before/after states, changes, etc.) |
| timestamp     | Date    | When action occurred |
| created_at    | Date    | Log creation time |

**Automatic Activity Tracking:**
- ✅ User registration and authentication events
- ✅ User and employee CRUD operations
- ✅ Attendance clock-in/clock-out events
- ✅ Failed login attempts and errors
- ✅ Inter-service communication
- ✅ Request/response logging with correlation IDs

**Querying Activity Logs:**
```bash
# Get recent activity via MongoDB shell
docker exec -it posdigi-mongodb mongosh
use posdigi_activity_logs
db.activity_logs.find().sort({timestamp: -1}).limit(10)

# Get user activity history
db.activity_logs.find({"user_id": "user-uuid"}).sort({timestamp: -1})

# Get failed login attempts
db.activity_logs.find({"action": "LOGIN_FAILED"}).sort({timestamp: -1})
```
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
- Go 1.25+
- Docker & Docker Compose
- MongoDB (local or Docker)
- PostgreSQL (local or Docker)
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
# PostgreSQL Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=posdigi_microservices
DB_USER=postgres
DB_PASSWORD=postgres

# MongoDB Configuration (Activity Logs)
MONGODB_HOST=localhost
MONGODB_PORT=27017
MONGODB_DATABASE=posdigi_activity_logs
MONGODB_USERNAME=
MONGODB_PASSWORD=
MONGODB_AUTH_DB=admin
MONGODB_CONNECTION_TIMEOUT=10
MONGODB_POOL_LIMIT=100

# Service Ports
AUTH_PORT=8001
USER_PORT=8002
ATTENDANCE_PORT=8003
GATEWAY_PORT=8000

# Service URLs (inter-service communication)
USER_SERVICE_URL=http://localhost:8002
AUTH_SERVICE_URL=http://localhost:8001
ATTENDANCE_SERVICE_URL=http://localhost:8003

# JWT Configuration
JWT_SECRET=your-secret-key-change-in-production
JWT_EXPIRY=24

# Internal Service Communication
INTERNAL_SERVICE_KEY=internal-service-key-change-in-production

# Logging
LOG_LEVEL=INFO

# Environment
ENVIRONMENT=development

# Gateway Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m
```

> ⚠️ **Security:** Always change `JWT_SECRET` and `INTERNAL_SERVICE_KEY` in production. Use `./generate-secrets.sh` to generate secure secrets.

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

The Docker setup includes all services plus both PostgreSQL and MongoDB databases:

```bash
make docker-up        # Start all services (including databases)
make docker-down      # Stop all services
make docker-build     # Build images
make docker-rebuild   # Rebuild from scratch (no cache)
make docker-logs      # Tail all logs
make docker-logs-auth # Tail auth service logs
make docker-ps        # Show container status
```

**Docker Services:**
- `postgres` - PostgreSQL database on port 5432
- `mongodb` - MongoDB database on port 27017
- `auth-service` - Authentication service on port 8001
- `user-service` - User management service on port 8002
- `attendance-service` - Attendance tracking service on port 8003
- `gateway-service` - API Gateway on port 8000

**Local MongoDB Setup:**
If you prefer to use local MongoDB instead of Docker:

```bash
# Make sure MongoDB is running on localhost:27017
# Update .env with your MongoDB credentials
# Services will automatically connect to local MongoDB
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
│   ├── shared/             # Shared packages and utilities
│   │   ├── activitylogger/ # MongoDB activity logging system
│   │   │   ├── logger.go   # Main logging interface
│   │   │   ├── middleware.go # HTTP logging middleware
│   │   │   ├── model.go    # Activity log data structures
│   │   │   └── repository.go # MongoDB operations
│   │   └── mongodb/        # MongoDB client and configuration
│   │       └── client.go   # Connection management
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
├── migrations/             # SQL migration files for PostgreSQL
├── mongo-init.js           # MongoDB initialization script
├── docker-compose.yml      # Multi-service Docker setup
├── Makefile                # Build automation
├── go.work                 # Go workspace configuration
├── DEPLOYMENT.md           # Production deployment guide
├── SECURITY_CHECKLIST.md   # Security hardening checklist
└── .env.example            # Environment variable template
```

---

## Make Commands Reference

| Command | Description |
|---------|-------------|
| `make setup` | Full setup: copy env, start Docker, run migrations |
| `make run-all` | Run all services locally |
| `make docker-up` | Start all Docker services |
| `make docker-down` | Stop all Docker services |
| `make migrate-up` | Apply all PostgreSQL migrations |
| `make migrate-down` | Roll back all migrations |
| `make migrate-create NAME=x` | Create a new migration |
| `make test-all` | Run all tests |
| `make tidy` | `go mod tidy` all services |
| `make install-deps` | Download all Go dependencies |
| `make swagger-generate` | Generate Swagger docs |
| `make db-shell` | Open PostgreSQL shell |
| `make db-backup` | Backup the PostgreSQL database |
| `make clean` | Remove build artifacts |

### MongoDB Commands:
```bash
# MongoDB shell
docker exec -it posdigi-mongodb mongosh

# MongoDB backup
docker exec posdigi-mongodb mongodump --db posdigi_activity_logs --out /backup

# MongoDB restore
docker exec posdigi-mongodb mongorestore --db posdigi_activity_logs /backup/posdigi_activity_logs

# View activity logs count
docker exec -it posdigi-mongodb mongosh --eval 'use posdigi_activity_logs; db.activity_logs.count()'
```

---

## Activity Logging & Monitoring

The system automatically logs all user activities to MongoDB for security auditing and compliance:

### What Gets Logged:
- **Authentication Events** - Login success/failure, token generation, registration
- **User Operations** - User creation, updates, deletions with before/after states
- **Employee Changes** - Profile updates, status changes, department transfers
- **Attendance Actions** - Clock-in/clockout events, modifications
- **System Events** - Health checks, migrations, configuration changes

### Activity Monitoring Examples:

```bash
# Get recent activity
docker exec -it posdigi-mongodb mongosh --eval '
  use posdigi_activity_logs;
  db.activity_logs.find()
    .sort({timestamp: -1})
    .limit(20);
'

# Monitor failed login attempts (security)
docker exec -it posdigi-mongodb mongosh --eval '
  use posdigi_activity_logs;
  db.activity_logs.find({
    action: "LOGIN_FAILED"
  }).sort({timestamp: -1});
'

# Get user activity timeline
docker exec -it posdigi-mongodb mongosh --eval '
  use posdigi_activity_logs;
  db.activity_logs.find({
    user_id: "user-uuid-here"
  }).sort({timestamp: -1});
'

# Generate activity statistics
docker exec -it posdigi-mongodb mongosh --eval '
  use posdigi_activity_logs;
  db.activity_logs.aggregate([
    {$group: {
      _id: "$action",
      count: {$sum: 1}
    }},
    {$sort: {count: -1}}
  ]);
'
```

### Activity Log API (Future Enhancement):
Planned endpoints for querying activity logs:
- `GET /api/v1/activity/logs` - Get all logs (admin only)
- `GET /api/v1/activity/logs/user/:userId` - User activity history
- `GET /api/v1/activity/logs/service/:service` - Service-specific logs
- `GET /api/v1/activity/logs/summary` - Activity statistics

---

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

---

## Troubleshooting

### Common Issues:

**MongoDB Connection Issues:**
```bash
# Check if MongoDB is running
docker-compose ps mongodb

# Check MongoDB logs
docker-compose logs mongodb

# Test MongoDB connection
docker exec -it posdigi-mongodb mongosh --eval "db.adminCommand('ping')"

# Restart MongoDB service
docker-compose restart mongodb
```

**Activity Logging Not Working:**
```bash
# Check MongoDB connection in service logs
docker-compose logs auth-service | grep -i mongo

# Verify MongoDB indexes
docker exec -it posdigi-mongodb mongosh --eval '
  use posdigi_activity_logs;
  db.activity_logs.getIndexes();
'

# Check if activity logs collection exists
docker exec -it posdigi-mongodb mongosh --eval '
  use posdigi_activity_logs;
  db.listCollections();
'
```

**Services Not Starting:**
```bash
# Check service logs
docker-compose logs auth-service

# Verify all dependencies are healthy
docker-compose ps

# Restart specific service
docker-compose restart auth-service

# Rebuild and restart
docker-compose up -d --build auth-service
```

**Database Migration Issues:**
```bash
# Check migration status
make migrate-status

# Fix dirty migration state
make migrate-fix-dirty

# Rollback and reapply
make migrate-redo
```

**Port Conflicts:**
```bash
# Check what's using port 8000
netstat -tulpn | grep :8000

# Change ports in .env file
nano .env
```

### Getting Help:

1. Check service logs: `docker-compose logs -f [service-name]`
2. Verify health endpoints: `curl http://localhost:8000/health`
3. Review [DEPLOYMENT.md](DEPLOYMENT.md) for deployment issues
4. Check [SECURITY_CHECKLIST.md](SECURITY_CHECKLIST.md) for security issues

---
