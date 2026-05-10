# PosDigi Microservices

A microservice backend system for employee attendance with clock-in/clockout functionality. Built with **Golang** using the Echo framework, following clean architecture principles.

## 🚀 Live Deployment

**Production URL:** [http://43.156.158.174:8000](http://43.156.158.174:8000)

### Access Points
- **Gateway Service:** http://43.156.158.174:8000
- **Auth Service:** http://43.156.158.174:8001
- **User Service:** http://43.156.158.174:8002
- **Attendance Service:** http://43.156.158.174:8003

### API Documentation
- **Auth Service:** http://43.156.158.174:8001/docs/index.html
- **User Service:** http://43.156.158.174:8002/docs/index.html
- **Attendance Service:** http://43.156.158.174:8003/docs/index.html

## 🏗️ Architecture

```
Client ──► Gateway (:8000) ──┬──► Auth Service (:8001)
                             │
                             ├──► User Service (:8002)
                             │
                             └──► Attendance Service (:8003)
                             └──► Shared Library (Activity Logging)
```

### Service Overview

- **Gateway**: Single entry point, rate limiting (in-memory), authentication middleware, reverse proxy routing
- **Auth Service**: User registration, login, JWT generation/validation, bcrypt password hashing
- **User Service**: User CRUD operations with pagination, filtering; comprehensive employee profile management
- **Attendance Service**: Clock-in/clockout, attendance history, work hours calculation
- **Shared Library**: MongoDB-based activity logging for audit trails across all services

## 🛠️ Tech Stack

- **Language:** Go 1.25.0
- **Framework:** Echo v4.15.2 (high-performance Go web framework)
- **ORM:** GORM v1.31.1 for PostgreSQL with automated migrations
- **Authentication:** JWT v5.3.1 with configurable expiry, bcrypt password hashing
- **Validation:** go-playground/validator v10.30.2 with custom error formatting
- **Documentation:** Swagger/OpenAPI with echo-swagger v1.5.2
- **Databases:**
  - PostgreSQL 15 (users, employees, attendance records)
  - MongoDB 7.0 (activity logs and audit trails)
- **Migration Tool:** golang-migrate with SQL files
- **Containerization:** Multi-stage Docker builds + Docker Compose orchestration
- **Go Workspace:** Multi-module management with `go.work`
- **Logging:** logrus v1.9.4 with structured logging

## 📋 Prerequisites

- Go 1.25.0 or higher
- Docker and Docker Compose
- PostgreSQL 15
- MongoDB 7.0
- make (optional, for Makefile commands)

## 🚦 Quick Start

### Using Docker (Recommended)

```bash
# Clone the repository
git clone <repository-url>
cd posdigi-microservices

# Copy environment file
cp .env.example .env

# Update environment variables if needed
# Start all services (including databases)
make docker-up

# Run database migrations
make migrate-up

# Access the application
# Gateway: http://localhost:8000
# Swagger docs: http://localhost:8001/docs/index.html
```

### Local Development

```bash
# Install dependencies
make install-deps

# Copy environment file
cp .env.example .env

# Start databases with Docker
docker-compose up -d postgres mongodb

# Run database migrations
make migrate-up

# Run individual services
make run-auth       # Auth service on :8001
make run-user       # User service on :8002
make run-attendance # Attendance service on :8003
make run-gateway    # Gateway service on :8000

# Or run all services at once
make run-all
```

## 📚 Common Commands

### Development

```bash
make run-auth      # Auth service on :8001
make run-user      # User service on :8002
make run-attendance # Attendance service on :8003
make run-gateway   # Gateway service on :8000
make run-all       # All services in parallel
make install-deps  # Install dependencies
make tidy          # Tidy Go modules
```

### Docker Operations

```bash
make docker-up      # Start all services
make docker-down    # Stop all services
make docker-rebuild # Rebuild images (no cache)
make docker-logs          # View all logs
make docker-logs-auth     # View auth service logs
make docker-logs-user     # View user service logs
make docker-logs-gateway  # View gateway logs
```

### Database Migrations

```bash
make migrate-up        # Run all migrations
make migrate-down      # Rollback all migrations
make migrate-down-one  # Rollback last migration
make migrate-redo      # Redo migrations (down + up)
make migrate-status    # Check migration status
make migrate-create NAME=add_table  # Create new migration
make migrate-fix-dirty # Fix dirty database state
```

### Full Setup

```bash
make setup  # Complete setup (copies .env, starts docker, runs migrations)
```

## 📁 Project Structure

```
posdigi-microservices/
├── services/
│   ├── auth/           # Authentication service
│   ├── user/           # User & employee management
│   ├── attendance/     # Attendance tracking
│   ├── gateway/        # API Gateway
│   └── shared/         # Shared libraries
│       ├── activitylogger/  # MongoDB activity logging
│       └── mongodb/         # MongoDB client utilities
├── migrations/         # Database migrations
├── docker-compose.yml  # Container orchestration
├── go.work            # Go workspace configuration
└── Makefile           # Build automation
```

Each service follows clean architecture:

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

## ⚙️ Configuration

Services use environment variables (see `.env.example`):

```bash
# PostgreSQL Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=posdigi_microservices
DB_USER=postgres
DB_PASSWORD=postgres

# MongoDB (Activity Logs)
MONGODB_HOST=localhost
MONGODB_PORT=27017
MONGODB_DATABASE=posdigi_activity_logs
MONGODB_USERNAME=
MONGODB_PASSWORD=
MONGODB_AUTH_DB=admin

# Service Ports
AUTH_PORT=8001
USER_PORT=8002
ATTENDANCE_PORT=8003
GATEWAY_PORT=8000

# JWT Configuration
JWT_SECRET=your-secret-key-change-in-production
JWT_EXPIRY=24

# Internal Service Communication
INTERNAL_SERVICE_KEY=internal-service-key-change-in-production

# Gateway Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m

# Logging
LOG_LEVEL=INFO
ENVIRONMENT=development
```

## 🔌 API Endpoints

### Gateway Service (:8000)
- `GET /health` - Aggregated health check
- `POST /api/v1/auth/*` - Public auth routes (no JWT required)
- `ANY /api/v1/users/*` - Protected user routes (JWT required)
- `ANY /api/v1/employees/*` - Protected employee routes (JWT required)
- `ANY /api/v1/attendance/*` - Protected attendance routes (JWT required)

### Auth Service (:8001)
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User authentication with JWT
- `POST /api/v1/auth/validate` - Token validation (POST)
- `GET /api/v1/auth/validate` - Token validation (GET)
- `GET /health` - Service health check
- `GET /docs/*` - Swagger documentation

### User Service (:8002)

**User Endpoints:**
- `POST /api/v1/users` - Create user
- `GET /api/v1/users` - List users (paginated, searchable)
- `GET /api/v1/users/:id` - Get user by ID
- `GET /api/v1/users/email/:email` - Get user by email (internal)
- `POST /api/v1/users/authenticate` - Validate credentials (internal)
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Soft delete user

**Employee Endpoints:**
- `POST /api/v1/employees` - Create employee profile
- `GET /api/v1/employees` - List employees (paginated, filterable)
- `GET /api/v1/employees/:id` - Get employee by ID
- `GET /api/v1/employees/:id/profile` - Get complete employee profile
- `PUT /api/v1/employees/:id` - Update employee
- `DELETE /api/v1/employees/:id` - Soft delete employee
- `PATCH /api/v1/employees/:id/status` - Update employment status
- `GET /api/v1/employees/active` - Get all active employees
- `GET /api/v1/employees/user/:userId` - Get employee by user ID
- `GET /api/v1/employees/code/:code` - Get employee by code
- `GET /api/v1/employees/department/:department` - Get by department
- `GET /api/v1/employees/manager/:managerId/subordinates` - Get subordinates

### Attendance Service (:8003)
- `POST /api/v1/attendance/clock-in` - Record clock-in
- `POST /api/v1/attendance/clock-out` - Record clock-out
- `GET /api/v1/attendance/history/:userId` - Get attendance history (paginated)
- `GET /api/v1/attendance/summary/:userId` - Get attendance summary by date range
- `GET /health` - Service health check
- `GET /docs/*` - Swagger documentation

## 🗄️ Database Schema

### Users Table
- `id` (UUID, primary key)
- `email` (VARCHAR, unique, not null)
- `password` (VARCHAR, bcrypt hashed, not null)
- `role` (VARCHAR, default 'user')
- `created_at`, `updated_at` (TIMESTAMP)
- `deleted_at` (TIMESTAMP, nullable, for soft delete)

### Employees Table
- `id` (UUID, primary key)
- `user_id` (UUID, foreign key to users)
- `employee_code` (VARCHAR, unique)
- `full_name`, `phone`, `department`, `position`
- `salary` (DECIMAL)
- `hire_date` (DATE)
- `employment_status` (VARCHAR: active/terminated/on_leave/suspended)
- `manager_id` (UUID, self-referencing for hierarchy)
- `emergency_contact`, `emergency_phone`, `address`
- `profile_image` (VARCHAR)
- `metadata` (JSONB for flexible data)
- `created_at`, `updated_at`, `deleted_at`

### Attendances Table
- `id` (UUID, primary key)
- `user_id` (UUID, foreign key to users)
- `clock_in` (TIMESTAMP)
- `clock_out` (TIMESTAMP, nullable)
- `notes` (TEXT)
- `created_at`, `updated_at` (TIMESTAMP)
- `deleted_at` (TIMESTAMP, nullable)

### MongoDB Activity Logs
- Automatic activity logging across all services
- Audit trails for user actions
- Request/response tracking for compliance

## 🔐 Security Features

- **JWT Authentication**: 24-hour token expiry with configurable secret
- **Password Security**: Bcrypt hashing for secure credential storage
- **Rate Limiting**: 100 requests/minute at Gateway level
- **Internal Service Auth**: Shared secret for inter-service communication
- **CORS Configuration**: Cross-origin request support
- **Input Validation**: Comprehensive request validation with clear error messages
- **Request ID Tracking**: Unique identifiers for audit trails
- **Activity Logging**: MongoDB-based audit logs for all service operations

## 🔄 Development Workflow

1. **Local Development**: Use `make run-{service}` to start services individually
2. **Testing**: All external requests go through Gateway (`:8000`)
3. **Documentation**: Access Swagger UI at `http://localhost:{port}/docs/*`
4. **Health Checks**: Each service has a `/health` endpoint
5. **Authentication Flow**:
   - Public routes (register/login) bypass Gateway auth
   - Protected routes require JWT in `Authorization: Bearer <token>` header
   - Inter-service communication uses `X-Service-Auth` header

## 🚧 Known Limitations

- No unit or integration tests currently implemented
- Rate limiting uses in-memory storage (lost on restart)
- Limited error recovery mechanisms
- No distributed tracing or advanced monitoring
- No HTTPS/TLS configuration (requires reverse proxy)

## 🚀 Production Deployment

### Before deploying to production:

1. **Change Security Keys:**
   ```bash
   JWT_SECRET=CHANGE_THIS_MINIMUM_32_CHAR_SECRET
   INTERNAL_SERVICE_KEY=CHANGE_THIS_MINIMUM_24_CHAR_SECRET
   ```

2. **Use Production Environment File:**
   ```bash
   cp .env.production.example .env
   # Update all production values
   ```

3. **Database Configuration:**
   - Configure PostgreSQL connection pooling
   - Set up MongoDB replica sets for high availability
   - Use strong database passwords

4. **Infrastructure:**
   - Use HTTPS/TLS termination (nginx, traefik, etc.)
   - Configure proper firewall rules
   - Set up monitoring and alerting
   - Implement log aggregation (ELK, Loki, etc.)
   - Consider Redis for distributed rate limiting

5. **Health Monitoring:**
   - Monitor `/health` endpoints
   - Set up alerts for service failures
   - Track MongoDB and PostgreSQL performance

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👥 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.