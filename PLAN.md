# Employee Attendance Microservice System - Implementation Plan

## Context
Building a microservice backend system for a technical test with **3 calendar days deadline**. The system will handle employee attendance with clock-in/clockout functionality, using Echo framework (Golang), PostgreSQL, MongoDB, and Redis (Upstash).

---

## Requirements Summary (from SOAL_TEST)

### Services to Build
1. **Gateway Service** - Single entry point, rate limiting, routing
2. **Auth Service** - Register, login, JWT validation, roles (user/admin)
3. **User Service** - Employee CRUD, pagination, search
4. **Attendance Service** - Clock-in/clockout, history, work hours calculation

### Databases
- **PostgreSQL** - Users, credentials, roles, attendance records (structured data)
- **MongoDB** - Activity logs, audit trails, user profile metadata (flexible data)
- **Redis (Upstash)** - Caching, rate limiting counters (BONUS feature)

### Key Features
- JWT authentication with bcrypt password hashing
- Rate limiting on gateway (BONUS - Redis-backed)
- Total hours calculation for attendance periods
- Swagger/OpenAPI documentation
- Docker & Docker Compose
- Health check endpoints
- Structured logging with request IDs (BONUS - observability)

---

## Architecture Overview

```
Client / HTTP
      │
      ▼
┌─────────────────────────────────────┐
│        Gateway Service :8000        │
│  Rate limit · Auth middleware · Proxy│
└──────┬────────────┬─────────────────┘
       │            │            │
       ▼            ▼            ▼
┌──────────┐ ┌──────────┐ ┌────────────┐
│   Auth   │ │   User   │ │ Attendance │
│  :8001   │ │  :8002   │ │   :8003    │
└────┬─────┘ └────┬─────┘ └─────┬──────┘
     │            │             │
     ▼            ▼             ▼
┌──────────┐ ┌──────────┐ ┌────────────┐
│PostgreSQL│ │PostgreSQL│ │ PostgreSQL │
│ (users)  │ │(profiles)│ │(attendance)│
└──────────┘ └────┬─────┘ └─────┬──────┘
                  │              │
           ┌──────▼──────┐ ┌────▼──────┐
           │   MongoDB   │ │ Redis/    │
           │ (logs/meta) │ │ Upstash   │
           └─────────────┘ └───────────┘
```

### Services

| Service | Port | Responsibility |
|---|---|---|
| Gateway | 8000 | Single entry point, rate limiting, auth middleware, reverse proxy |
| Auth | 8001 | Register, login, JWT generation, token validation |
| User | 8002 | CRUD employees, pagination, filter, soft delete |
| Attendance | 8003 | Clock in/out, history, work hours calculation |

---

## Project Structure

```
posdigi-microservices/
├── services/
│   ├── gateway/
│   │   ├── main.go
│   │   ├── middleware/
│   │   │   ├── auth.go          # validate token via Auth Service
│   │   │   ├── ratelimit.go     # Redis token bucket per IP
│   │   │   ├── requestid.go     # generate unique request ID
│   │   │   └── logger.go        # structured logging
│   │   ├── proxy/
│   │   │   └── routes.go        # reverse proxy routing table
│   │   ├── Dockerfile
│   │   └── .env.example
│   ├── auth/
│   │   ├── main.go
│   │   ├── handler/
│   │   │   ├── register.go
│   │   │   ├── login.go
│   │   │   └── validate.go
│   │   ├── service/
│   │   │   └── auth_service.go
│   │   ├── repository/
│   │   │   └── postgres_repo.go
│   │   ├── model/
│   │   │   ├── user.go
│   │   │   └── token.go
│   │   ├── migrations/
│   │   │   └── 001_users.up.sql
│   │   ├── Dockerfile
│   │   └── .env.example
│   ├── user/
│   │   ├── main.go
│   │   ├── handler/
│   │   │   └── user_handler.go
│   │   ├── service/
│   │   │   └── user_service.go
│   │   ├── repository/
│   │   │   ├── postgres_repo.go
│   │   │   └── mongo_repo.go    # user_profiles, activity_logs
│   │   ├── model/
│   │   │   ├── employee.go
│   │   │   └── profile.go
│   │   ├── migrations/
│   │   │   └── 002_employees.up.sql
│   │   ├── Dockerfile
│   │   └── .env.example
│   └── attendance/
│       ├── main.go
│       ├── handler/
│       │   ├── clockin.go
│       │   ├── clockout.go
│       │   ├── history.go
│       │   └── summary.go       # total hours endpoints
│       ├── service/
│       │   ├── attendance_service.go
│       │   └── total_hours_service.go
│       ├── repository/
│       │   ├── postgres_repo.go
│       │   ├── mongo_repo.go    # attendance_logs
│       │   └── redis_repo.go    # cache total hours
│       ├── model/
│       │   ├── attendance.go
│       │   └── activity_log.go
│       ├── migrations/
│       │   └── 003_attendance.up.sql
│       ├── Dockerfile
│       └── .env.example
├── docker-compose.yml
├── .env.example
├── Makefile
├── README.md
└── .gitignore
```

---

## Database Design

### PostgreSQL (structured / relational)

**`users` table** — Auth Service
```sql
CREATE TABLE users (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name          VARCHAR(100) NOT NULL,
  email         VARCHAR(150) NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  role          VARCHAR(20) NOT NULL DEFAULT 'employee', -- employee | admin
  department    VARCHAR(100),
  is_deleted    BOOLEAN NOT NULL DEFAULT false,
  created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_department ON users(department);
```

**`employees` table** — User Service (optional, can use users table directly)
```sql
CREATE TABLE employees (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  phone         VARCHAR(20),
  position      VARCHAR(100),
  is_active     BOOLEAN NOT NULL DEFAULT true,
  created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMP NOT NULL DEFAULT NOW()
);
```

**`attendance` table** — Attendance Service
```sql
CREATE TABLE attendance (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id               UUID NOT NULL REFERENCES users(id),
  clock_in              TIMESTAMP NOT NULL,
  clock_out             TIMESTAMP,
  work_duration_minutes INT,     -- computed on clock-out
  date                  DATE NOT NULL,
  notes                 TEXT,
  created_at            TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_attendance_user_id ON attendance(user_id);
CREATE INDEX idx_attendance_date ON attendance(date);
CREATE INDEX idx_attendance_clock_in ON attendance(clock_in);
```

### MongoDB (logs & flexible data)

| Collection | Used By | Purpose |
|---|---|---|
| `activity_logs` | All services | API request logs (user_id, action, ip, timestamp) |
| `attendance_logs` | Attendance | Raw clock in/out events before aggregation |
| `user_profiles` | User | Extended profile (avatar, address, emergency contact) |

**MongoDB Document Example:**
```javascript
// activity_logs collection
{
  "_id": ObjectId("..."),
  "user_id": "uuid",
  "action": "clock_in",
  "resource": "attendance",
  "ip_address": "192.168.1.1",
  "user_agent": "Mozilla/5.0...",
  "timestamp": ISODate("2025-01-15T08:00:00Z"),
  "details": {
    "attendance_id": "uuid",
    "notes": "Regular shift"
  }
}
```

---

## Environment Variables

**.env.example**
```env
# Ports
GATEWAY_PORT=8000
AUTH_PORT=8001
USER_PORT=8002
ATTENDANCE_PORT=8003

# Upstream URLs (used by Gateway)
AUTH_SERVICE_URL=http://auth:8001
USER_SERVICE_URL=http://user:8002
ATTENDANCE_SERVICE_URL=http://attendance:8003

# PostgreSQL
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=attendance_db
DATABASE_URL=postgres://postgres:postgres@postgres:5432/attendance_db?sslmode=disable

# MongoDB
MONGO_URI=mongodb://mongo:27017
MONGO_DB=attendance_logs

# JWT
JWT_SECRET=your-super-secret-key-here
JWT_EXPIRY_HOURS=24

# Redis (Upstash) - BONUS feature
REDIS_URL=rediss://:your-upstash-token@your-upstash-host:6380

# Rate Limiting - BONUS feature
RATE_LIMIT_REQUESTS=60
RATE_LIMIT_WINDOW_SECONDS=60
```

---

## API Endpoints

### Gateway
| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/health` | No | Health check |
| GET | `/docs` | No | Swagger UI |

### Auth Service (proxied via Gateway at `/auth/*`)
| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/auth/register` | No | Register new employee |
| POST | `/auth/login` | No | Login, returns JWT |
| GET | `/auth/validate` | Bearer | Validate token, returns payload |

### User Service (proxied via Gateway at `/users/*`)
| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/users` | Bearer | List users with pagination & filter |
| GET | `/users/:id` | Bearer | Get user detail |
| POST | `/users` | Admin | Create user/profile |
| PUT | `/users/:id` | Admin/Self | Update user/profile |
| DELETE | `/users/:id` | Admin | Soft delete user |

**Query params for `GET /users`:** `page`, `limit`, `name`, `email`, `role`, `department`

### Attendance Service (proxied via Gateway at `/attendance/*`)
| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/attendance/clock-in` | Bearer | Clock in (fails if session already open) |
| POST | `/attendance/clock-out` | Bearer | Clock out, computes work duration |
| GET | `/attendance/history` | Bearer | Paginated attendance history |
| GET | `/attendance/total-hours` | Bearer | Total hours for date range (cached) |
| GET | `/attendance/today` | Bearer | Today's worked hours |
| GET | `/attendance/weekly` | Bearer | This week's total hours |
| GET | `/attendance/monthly` | Bearer | This month's total hours |

---

## Standard Response Format

**Success**
```json
{
  "success": true,
  "message": "ok",
  "data": { ... }
}
```

**Error**
```json
{
  "success": false,
  "message": "unauthorized",
  "code": 401
}
```

---

## Day-by-Day Implementation Plan

### Day 1 — Foundation (~7–8 hours)

#### Morning (3h): Scaffold + Infra
- [ ] Init monorepo, one Go module per service (`go mod init`)
- [ ] Write `docker-compose.yml` — all 4 services + Postgres + MongoDB + Redis
- [ ] Write `.env.example`
- [ ] Set up Echo in all services with `/health` endpoint
- [ ] Add basic request logging middleware to each service
- [ ] Install `swaggo/echo-swagger` and set up `swag init`

#### Afternoon (4h): Auth Service
- [ ] PostgreSQL connection + `golang-migrate` for `users` table
- [ ] `POST /auth/register` — validate input, bcrypt hash, save
- [ ] `POST /auth/login` — verify password, sign JWT (claims: user_id, role, email)
- [ ] `GET /auth/validate` — parse token, return claims
- [ ] Seed: 1 admin user, 2 employee users
- [ ] Swagger annotations on all auth endpoints

---

### Day 2 — Core Services (~8–9 hours)

#### Morning (4h): User Service
- [ ] GORM for PostgreSQL users CRUD
- [ ] MongoDB connection for `user_profiles`
- [ ] `GET /users` — pagination + filter by name/email/role/department
- [ ] `GET /users/:id`
- [ ] `POST /users` — admin only, create user
- [ ] `PUT /users/:id` — admin or self
- [ ] `DELETE /users/:id` — soft delete (`is_deleted = true`)
- [ ] Swagger annotations

#### Afternoon (4h): Attendance Service
- [ ] PostgreSQL migration for `attendance` table
- [ ] MongoDB connection for `attendance_logs`
- [ ] `POST /attendance/clock-in` — check no open session, insert with clock_in timestamp
- [ ] `POST /attendance/clock-out` — find open session, set clock_out, compute `work_duration_minutes`
- [ ] `GET /attendance/history` — paginated, filter by date range
- [ ] `GET /attendance/total-hours` — with Redis cache
- [ ] Log every event to MongoDB `attendance_logs`
- [ ] Swagger annotations

---

### Day 3 — Gateway, Polish, Docs (~6–7 hours)

#### Morning (3h): Gateway Service
- [ ] Echo reverse proxy to each upstream service
- [ ] Auth middleware: extract Bearer → call Auth Service `/auth/validate` → attach `X-User-ID` and `X-User-Role` headers
- [ ] Rate limiting via `go-redis` + Upstash (token bucket, 60 req/min per IP)
- [ ] Route table:
  - `/auth/*` → Auth Service (public)
  - `/users/*` → User Service (protected)
  - `/attendance/*` → Attendance Service (protected)
- [ ] Consistent error wrapper for upstream errors

#### Afternoon (3h): Finalization
- [ ] Run `swag init` in each service, commit `docs/` folder
- [ ] End-to-end test via docker compose (register → login → clock-in → clock-out → history)
- [ ] Write README (architecture, env setup, `docker compose up`, Swagger URLs)
- [ ] Verify all Dockerfiles build cleanly
- [ ] Check no secrets committed to repo
- [ ] Optional: VPS deployment prep

---

## Effort Summary

| Area | Est. Hours |
|---|---|
| Scaffold + docker-compose | 2h |
| Auth Service | 3h |
| User Service | 3h |
| Attendance Service | 4h |
| Gateway + rate limit | 3h |
| Swagger + README | 2h |
| End-to-end testing | 1h |
| Buffer / VPS deploy | 2h |
| **Total** | **~20 hours** |

Fits comfortably in **3 calendar days**, leaving buffer for polish.

---

## Key Implementation Notes

### JWT Flow
- **Auth Service** signs tokens with `JWT_SECRET`
- **Gateway** calls Auth Service to validate — does NOT re-implement validation
- **User/Attendance Services** trust `X-User-ID` and `X-User-Role` headers forwarded by Gateway
- Services do NOT need `JWT_SECRET` in their env

### Work Hours Calculation
```go
// On clock-out
duration := clockOut.Sub(clockIn)
minutes := int(duration.Minutes())
// Store in attendance.work_duration_minutes
// For summary: SUM(work_duration_minutes) / 60.0 = total hours
```

### Clock-in Guard
```go
// Before inserting clock-in, check for open session
var open Attendance
err := db.Where("user_id = ? AND clock_out IS NULL", userID).First(&open).Error
if err == nil {
    return echo.NewHTTPError(400, "already clocked in")
}
```

### Rate Limiting with Upstash
```go
// go-redis client works directly with Upstash REST URL
rdb := redis.NewClient(&redis.Options{
    Addr: os.Getenv("REDIS_URL"), // Upstash endpoint
})
// Token bucket: INCR key with TTL per IP
```

### Migrations
Use `golang-migrate` — drop SQL files in `migrations/` and run on startup:
```go
m, _ := migrate.New("file://migrations", os.Getenv("DATABASE_URL"))
m.Up()
```
Much cleaner than GORM AutoMigrate for reviewer inspection.

### Shared Response Package
Create a small `response/` package inside each service:
```go
type Response struct {
    Success bool        `json:"success"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
    Code    int         `json:"code,omitempty"`
}

func OK(c echo.Context, data interface{}) error {
    return c.JSON(200, Response{Success: true, Message: "ok", Data: data})
}

func Err(c echo.Context, code int, msg string) error {
    return c.JSON(code, Response{Success: false, Message: msg, Code: code})
}
```

### Redis Caching for Total Hours
```go
// Cache key format: "total_hours:{user_id}:{period}:{start_date}:{end_date}"
// TTL: 5 minutes
// Invalidate on: clock-in, clock-out

// Check cache first
if cached, err := cache.Get(ctx, key); err == nil {
    return cached, nil
}

// Calculate from PostgreSQL
result := calculateFromDB(ctx, req)

// Store in cache
cache.Set(ctx, key, result, 5*time.Minute)
```

---

## Dependencies

```bash
# Web Framework
go get github.com/labstack/echo/v4
go get github.com/labstack/echo/v4/middleware

# Database ORM
go get gorm.io/gorm
go get gorm.io/driver/postgres

# Database Drivers
go get github.com/lib/pq
go get go.mongodb.org/mongo-driver/mongo

# JWT & Security
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto/bcrypt

# Validation
go get github.com/go-playground/validator/v10

# Database Migration
go get github.com/golang-migrate/migrate/v4

# Swagger
go get github.com/swaggo/echo-swagger
go get github.com/swaggo/swag/cmd/swag

# Logging
go get go.uber.org/zap

# Env Config
go get github.com/joho/godotenv

# Redis (Upstash) - BONUS
go get github.com/redis/go-redis/v9
```

---

## Scoring Checklist (Aligned with Test Requirements)

| Criteria | Weight | What to nail |
|---|---|---|
| Microservice Architecture | 20% | Clear service separation, Gateway as single entry point |
| Feature Implementation | 20% | All endpoints working, pagination, role checks, clock-in/out |
| Database Design | 15% | Postgres for structured data, Mongo for logs/profiles, migrations included |
| Code Quality | 15% | Modular layout, consistent error handling, logging |
| Basic Security | 10% | Bcrypt, JWT via env, no secrets committed, input validation |
| Documentation | 10% | README complete, Swagger on `/docs` per service |
| Docker/Deployment | 10% | `docker compose up` works from scratch |

### Bonus Points (already planned)
- [x] Rate limiting on Gateway (Upstash Redis)
- [x] Role-based access control (admin/employee)
- [x] Redis caching (Upstash)
- [x] Observability (Request ID, structured logging)
- [ ] Refresh token flow *(skip to save time)*
- [ ] CI/CD via GitHub Actions *(optional, add if time permits)*

---

## Verification Checklist

- [ ] All services start with `docker-compose up`
- [ ] Can register new user via POST /auth/register
- [ ] Can login via POST /auth/login and receive JWT
- [ ] Can validate token via GET /auth/validate
- [ ] Can list users via GET /users (with pagination & filter)
- [ ] Can create user via POST /users (admin only)
- [ ] Can update user via PUT /users/:id
- [ ] Can soft delete user via DELETE /users/:id
- [ ] Can clock in via POST /attendance/clock-in
- [ ] Can clock out via POST /attendance/clock-out
- [ ] Clock out calculates and stores work_duration_minutes correctly
- [ ] Can get attendance history via GET /attendance/history
- [ ] Can get total hours via GET /attendance/total-hours
- [ ] Total hours cache works (Redis/Upstash)
- [ ] Rate limiting works on gateway (60 req/min per IP)
- [ ] Swagger UI accessible at http://localhost:8000/docs
- [ ] Health checks return 200 OK for all services
- [ ] MongoDB activity logs are created
- [ ] No secrets in repository
- [ ] README is complete with architecture and setup instructions
- [ ] (Optional) Deployed on VPS

---

## Submission Template

```
Nama Kandidat  : [Your Name]
Posisi         : Backend Engineer
Link Repository: https://github.com/yourname/posdigi-microservices
Link Deployment: https://your-vps-ip:8000 (Gateway) - if available
Swagger URLs   :
  - Gateway     : http://localhost:8000/docs
  - Auth        : http://localhost:8000/docs (proxied)
  - User        : http://localhost:8000/docs (proxied)
  - Attendance  : http://localhost:8000/docs (proxied)
Catatan        :
  - PostgreSQL digunakan untuk data transaksional (users, attendance)
  - MongoDB digunakan untuk activity logs dan user profile metadata
  - Redis (Upstash) digunakan untuk rate limiting di Gateway dan caching total hours
  - Attendance Service ditambahkan sebagai domain utama tema aplikasi (employee attendance)
  - Bonus points: Rate limiting (Redis), Caching (Redis), RBAC, Observability (Request ID, structured logging)
```

---

*Good luck with your technical test! 🚀*
