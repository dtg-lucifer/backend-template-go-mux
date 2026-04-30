# Application Workflow & Architecture

This document explains every module, component, and subsystem in this template — what it is, why it exists, and how it connects to everything else.

---

## Table of Contents

1. [High-Level Architecture](#1-high-level-architecture)
2. [Startup Sequence](#2-startup-sequence)
3. [Request Lifecycle](#3-request-lifecycle)
4. [PostgreSQL & SQLC Query Layer](#4-postgresql--sqlc-query-layer)
5. [Authentication & JWT](#5-authentication--jwt)
6. [Middleware Stack](#6-middleware-stack)
7. [Handler & Module Structure](#7-handler--module-structure)
8. [Service Layer](#8-service-layer)
9. [Configuration System](#9-configuration-system)
10. [Logging System](#10-logging-system)
11. [API Response Convention](#11-api-response-convention)
12. [Graceful Shutdown](#12-graceful-shutdown)
13. [Adding Features](#13-adding-features)

---

## 1. High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         CLIENT                              │
│                    (HTTP REST)                              │
└──────────────────────────┬──────────────────────────────────┘
                           │ HTTP
                           ▼
┌──────────────────────────────────────────────────────────────┐
│                  gorilla/mux Router                          │
│                                                              │
│  Global middleware chain:                                    │
│  RequestID → CORS → RateLimit → Logger → Timeout             │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐    │
│  │ GET /health  │  │ POST /auth/* │  │  (your modules)  │    │
│  └──────────────┘  └──────┬───────┘  └──────────────────┘    │
└─────────────────────────── │ ────────────────────────────────┘
                             │ service calls
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                     Service Layer                           │
│                                                             │
│  AuthService ──► repository.Queries ──► PostgreSQL (pgx)    │
│                                                             │
│  Services return ApiResponse — never touch ResponseWriter   │
└─────────────────────────────────────────────────────────────┘
```

---

## 2. Startup Sequence

**Entry point:** `main.go`

```
main.go
  │
  ├─ godotenv.Load()              load .env into process.env
  ├─ pkg.NewLogger()              initialise dual-output logger
  ├─ config.NewConfig(path)       parse & validate config.yaml
  └─ server.New(cfg, logger).Start()
       │
       ├─ db.Connect()            open pgxpool, ping database
       ├─ mux.NewRouter()         create the gorilla/mux router
       ├─ router.Use(...)         attach global middleware stack
       ├─ handler.RegisterRoutes  mount all HTTP routes
       └─ httpSrv.ListenAndServe  start accepting connections
```

---

## 3. Request Lifecycle

Every HTTP request passes through this pipeline before reaching a handler:

```
Incoming Request
      │
      ▼
  RequestIDMiddleware    generate UUID → X-Request-ID header
      │
      ▼
  CORSMiddleware         validate Origin, set Access-Control-* headers
      │
      ▼
  RateLimiter.Middleware per-IP sliding window (200 req/min default)
      │
      ▼
  LoggerMiddleware       log method, path, status, duration, IP
      │
      ▼
  TimeoutMiddleware      wrap context with deadline from config.yaml
      │
      ▼
  AuthMiddleware.Guard   (protected routes only) verify Bearer JWT
      │
      ▼
  Route Handler          call service, get ApiResponse, SendResponse()
      │
      ▼
  SendResponse()         inject requestId, w.WriteHeader(), json.Encode()
```

---

## 4. PostgreSQL & SQLC Query Layer

**Files:** `internal/db/queries/`, `internal/db/migrations/`, `internal/db/repository/`

**Why SQLC:** SQL stays in `.sql` files — readable, reviewable, and version-controlled. SQLC generates fully type-safe Go functions from those queries. No ORM magic, no runtime reflection.

**How it works:**

```
internal/db/queries/users.sql
  │
  └─ sqlc generate (make sqlc-gen)
       │
       └─► internal/db/repository/
               ├─ db.go          DBTX interface + New()
               ├─ models.go      Go structs mirroring DB tables
               └─ users.sql.go   Generated query functions
```

**Adding a query:**

1. Write SQL in `internal/db/queries/<domain>.sql` with a `-- name: FunctionName :one/:many/:exec` annotation
2. Run `make sqlc-gen`
3. Call the generated function from your service via `s.repo.FunctionName(ctx, params)`

**Migrations** use `golang-migrate` with sequential numbered files:

```
internal/db/migrations/
  000001_init_auth.up.sql
  000001_init_auth.down.sql
  000002_add_posts.up.sql     ← created by: make migrate-new
  000002_add_posts.down.sql
```

**Connection pool** is configured in `config.yaml → database`:

- `pool_size` — max open connections
- `MaxConnLifetime: 30m` — connections are recycled after 30 minutes
- `MaxConnIdleTime: 5m` — idle connections are closed after 5 minutes

---

## 5. Authentication & JWT

**Files:** `pkg/jwt.go`, `internal/middlewares/auth_middleware.go`, `internal/services/auth/`

**Token flow:**

```
POST /auth/register
  └─► AuthService.Register()
        ├─ check email uniqueness
        ├─ pkg.HashPassword()     bcrypt, DefaultCost
        └─ repo.CreateUser()

POST /auth/login
  └─► AuthService.Login()
        ├─ repo.GetUserByEmail()
        ├─ pkg.ComparePassword()  bcrypt timing-safe compare
        ├─ pkg.NewTokenSigner(JWT_SECRET).Sign()         24h access token
        └─ pkg.NewTokenSigner(JWT_REFRESH_SECRET).Sign() 30d refresh token

GET /auth/me  (protected)
  └─► AuthMiddleware.Guard
        ├─ extract token from "jwt" cookie or Authorization: Bearer header
        ├─ pkg.NewTokenSigner(JWT_SECRET).Verify()
        └─ inject uid into request context via context.WithValue
  └─► AuthService.Me(uid)
        └─ repo.GetUserByID()

POST /auth/refresh
  └─► AuthService.RefreshToken(refreshToken)
        ├─ pkg.NewTokenSigner(JWT_REFRESH_SECRET).Verify()
        └─ issue new access token
```

**`OptionalGuard`**: A variant of `Guard` that does not reject the request if no token is present — useful for endpoints that behave differently for authenticated vs anonymous users.

**Reading the UID in a handler:**

```go
uid := middlewares.UIDFromContext(r.Context())
```

---

## 6. Middleware Stack

**Files:** `internal/middlewares/`

| File | What it does |
|---|---|
| `requestid_middleware.go` | Generates a UUID per request, sets `X-Request-ID` response header |
| `logger_middleware.go` | Logs method, path, status, duration, IP to stdout + `logs/events.log` |
| `cors_middleware.go` | Sets `Access-Control-*` headers from `CORS_ORIGINS` env var |
| `timeout_middleware.go` | Wraps each request in a context with a deadline from `config.yaml` |
| `auth_middleware.go` | `Guard` (required auth), `OptionalGuard`, `UIDFromContext` |
| `ratelimit_middleware.go` | In-memory per-IP sliding-window rate limiter |

**Registration order in `server.go`** (matters):

```
RequestID → CORS → RateLimit → Logger → Timeout
```

RequestID runs first so every subsequent middleware and log entry has the ID available.

---

## 7. Handler & Module Structure

**Files:** `internal/handlers/`

Every feature module is a struct that implements `handlers.Handler`:

```go
type Handler interface {
    RegisterRoutes(router *mux.Router)
}
```

**Pattern inside a handler:**

```go
func (h *ThingHandler) RegisterRoutes(router *mux.Router) {
    sub := router.PathPrefix("/things").Subrouter()
    sub.HandleFunc("", h.list).Methods(http.MethodGet)
    // Protected route — wrap with Guard:
    sub.Handle("", h.authMW.Guard(http.HandlerFunc(h.create))).Methods(http.MethodPost)
}

func (h *ThingHandler) create(w http.ResponseWriter, r *http.Request) {
    wr := utils.NewHttpWriter(w, r)

    var body struct { Name string `json:"name"` }
    if err := wr.ParseBody(&body); err != nil {
        wr.Status(http.StatusBadRequest).JSON(utils.M{"success": false, "message": err.Error()})
        return
    }

    resp := h.service.Create(r.Context(), body.Name)
    apiresponse.SendResponse(w, resp)
}
```

**Registering a new handler** — add one line to `internal/server/server.go`:

```go
handlers := []interface{ RegisterRoutes(*mux.Router) }{
    apihandlers.NewHealthHandler(pool, s.startTime),
    apihandlers.NewAuthHandler(pool),
    apihandlers.NewThingHandler(pool), // ← here
}
```

---

## 8. Service Layer

**Files:** `internal/services/`

Services contain all business logic. They:

- Accept typed input structs
- Call `repository.Queries` methods for DB access
- Return `apiresponse.ApiResponse` — **never** touch `http.ResponseWriter`

```go
func (s *Service) Create(ctx context.Context, input CreateInput) apiresponse.ApiResponse {
    thing, err := s.repo.CreateThing(ctx, repository.CreateThingParams{
        Name: input.Name,
    })
    if err != nil {
        return apiresponse.Error("failed to create thing", err, 500)
    }
    return apiresponse.Success("thing created", map[string]any{"thing": thing}, 201)
}
```

This separation means services are trivially unit-testable without an HTTP layer.

---

## 9. Configuration System

**Files:** `config/config.go`, `config.yaml`, `.env`

**Two-layer config:**

```
config.yaml     static, committed, non-secret settings
.env            secrets and environment-specific overrides
```

**Environment overrides** (take precedence over config.yaml):

| Env var | Overrides |
|---|---|
| `DB_URL` | Entire database DSN |
| `PORT` | `server.port` |
| `HOST` | `server.host` |
| `JWT_SECRET` | *(no yaml equivalent — env only)* |
| `JWT_REFRESH_SECRET` | *(no yaml equivalent — env only)* |
| `CORS_ORIGINS` | *(no yaml equivalent — env only)* |

---

## 10. Logging System

**Files:** `pkg/logger.go`, `internal/middlewares/logger_middleware.go`

```
pkg.Logger
  │
  ├─► FileLogger   (slog.JSONHandler → logs/app.log)    level: DEBUG (captures everything)
  └─► StdoutLogger (slog.TextHandler → os.Stdout)       level: INFO  (human-readable)

LoggerMiddleware
  │
  ├─► os.Stdout                                          HTTP access log (text)
  └─► logs/events.log                                    HTTP access log (text, append)
```

**Usage in handlers and services:**

```go
logger := pkg.NewLogger()
defer logger.Close()
logger.Info("user registered", "user_id", user.ID)
logger.Error("db query failed", "error", err)
```

---

## 11. API Response Convention

**Files:** `internal/utils/apiresponse/apiresponse.go`

All responses follow a single JSON shape:

```json
// Success
{
  "success": true,
  "message": "user registered successfully",
  "data": { "user": { ... } },
  "status_code": 201,
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}

// Error
{
  "success": false,
  "message": "email already registered",
  "errors": null,
  "status_code": 409,
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Services** build responses:

```go
return apiresponse.Success("ok", data, 200)
return apiresponse.Error("not found", nil, 404)
```

**Handlers** send them:

```go
resp := h.service.DoThing(r.Context(), input)
apiresponse.SendResponse(w, resp)
```

`SendResponse` injects the `X-Request-ID` header value into `request_id` before encoding.

---

## 12. Graceful Shutdown

**File:** `internal/server/server.go`

```
SIGINT / SIGTERM received
  │
  └─► httpSrv.Shutdown(ctx with 30s timeout)
        ├─ stop accepting new connections
        ├─ wait for in-flight requests to complete
        └─ pool.Close()   ← deferred, closes pgxpool
```

---

## 13. Adding Features

### New route module checklist

1. **SQL queries** → `internal/db/queries/<name>.sql`
2. **Regenerate** → `make sqlc-gen`
3. **Service** → `internal/services/<name>/<name>_service.go`
   - Accept typed input, return `apiresponse.ApiResponse`
4. **Handler** → `internal/handlers/v1/api/<name>_handler.go`
   - Implement `handlers.Handler`, parse body, call service, call `SendResponse`
5. **Register** → add to the handlers slice in `internal/server/server.go`

### New migration checklist

1. `make migrate-new` → enter a name
2. Write SQL in the generated `.up.sql` and `.down.sql` files
3. `make migrate-up`
4. If the migration adds/changes tables with queries, update `.sql` files and `make sqlc-gen`
