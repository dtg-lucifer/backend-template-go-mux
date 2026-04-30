# Go + Mux Backend Template

A production-ready Go backend template built with **gorilla/mux**, **SQLC**, and **pgx**. Structured after the [bun-express-backend-template](https://github.com/dtg-lucifer/bun-express-backend-template) and the [Everato](https://github.com/everato-industries/everato) Go project — same conventions, same discipline, different runtime.

## What this template includes

- PostgreSQL access via `pgx/v5` with a connection pool
- Type-safe SQL via **SQLC** — queries live in `internal/db/queries/`, generated code in `internal/db/repository/`
- Database migrations via **golang-migrate** (plain `.up.sql` / `.down.sql` files)
- Explicit handler registry — implement `handlers.Handler`, call `RegisterRoutes()` in `server.go`
- `HttpWriter` fluent response helper — chainable `.Status().JSON()` / `.Error()` / `.Text()`
- `ApiResponse` service-layer envelope — services return `ApiResponse`, handlers call `SendResponse()`
- JWT authentication middleware (access + refresh tokens, cookie + Bearer header)
- Per-request UUID (`X-Request-ID` header, injected into every response body)
- Structured logging via `log/slog` — JSON to `logs/app.log`, text to stdout
- In-memory per-IP rate limiter (swap for Redis-backed in production)
- CORS middleware driven by `CORS_ORIGINS` env var
- Per-request timeout middleware
- Graceful shutdown on `SIGINT` / `SIGTERM`
- Hot reload via **Air** (`make dev`)
- Module rename script (`scripts/rename-module.sh`)

For a full explanation of every component, see [WORKFLOW.md](./WORKFLOW.md).

---

## Quick Start

### 1. Rename the module

```bash
./scripts/rename-module.sh github.com/your-org/your-project
go mod tidy
```

### 2. Install dev tools

```bash
make install
```

### 3. Start PostgreSQL

```bash
make db
```

### 4. Copy and fill in `.env`

```bash
cp .env.example .env
# Edit .env — at minimum set JWT_SECRET and JWT_REFRESH_SECRET
```

### 5. Run migrations

```bash
make migrate-up
```

### 6. Generate SQLC repository code

```bash
make sqlc-gen
```

### 7. Start the dev server (hot reload)

```bash
make dev
```

---

## Scripts

| Command | Description |
|---|---|
| `make dev` | Start with hot reload (Air) |
| `make build` | Compile to `bin/app` |
| `make run` | Build then run |
| `make test` | Run all tests with race detector |
| `make lint` | Run golangci-lint |
| `make fmt` | Format all Go files |
| `make sqlc-gen` | Regenerate `internal/db/repository/` from SQL |
| `make db` | Start PostgreSQL via Docker Compose |
| `make db-stop` | Stop PostgreSQL |
| `make migrate-up` | Apply all pending migrations |
| `make migrate-down` | Roll back one migration |
| `make migrate-status` | Show current migration version |
| `make migrate-new` | Create a new migration file pair |
| `make clean` | Remove `bin/` and `logs/` |

---

## Project Structure

```
.
├── main.go                          Application entry point
├── config.yaml                      Static, non-secret configuration
├── .env.example                     Environment variable template
├── sqlc.yaml                        SQLC code generation config
├── Makefile                         All dev commands
├── .air.toml                        Hot reload config
│
├── config/
│   └── config.go                    Config loader (YAML + env overrides)
│
├── pkg/                             Shared, reusable utilities (no business logic)
│   ├── logger.go                    Dual-output structured logger (file + stdout)
│   ├── jwt.go                       JWT sign / verify + StandardClaims helper
│   ├── password.go                  bcrypt hash + compare
│   └── env.go                       GetEnv helper
│
├── internal/
│   ├── server/
│   │   └── server.go                HTTP server wiring, startup, graceful shutdown
│   │
│   ├── db/
│   │   ├── db.go                    pgxpool connection factory
│   │   ├── migrations/              SQL migration files (*.up.sql / *.down.sql)
│   │   ├── queries/                 SQLC query definitions (*.sql)
│   │   └── repository/              SQLC-generated Go code (do not edit manually)
│   │       └── helpers.go           Hand-written helpers (StringToUUID, etc.)
│   │
│   ├── middlewares/
│   │   ├── requestid_middleware.go  UUID per request → X-Request-ID header
│   │   ├── logger_middleware.go     HTTP access log (stdout + logs/events.log)
│   │   ├── cors_middleware.go       CORS headers from CORS_ORIGINS env var
│   │   ├── timeout_middleware.go    Per-request deadline from config.yaml
│   │   ├── auth_middleware.go       JWT Guard + OptionalGuard + UIDFromContext
│   │   └── ratelimit_middleware.go  In-memory per-IP rate limiter
│   │
│   ├── modules/                     ← Feature modules (mirrors src/modules/ in Express template)
│   │   ├── routes.go                Central registry — mount modules here (mirrors modules/index.ts)
│   │   ├── health/
│   │   │   └── routes.go            GET /health — status, DB ping, memory, uptime
│   │   └── auth/
│   │       ├── schema.go            Input types + Validate() (mirrors auth.schema.ts)
│   │       ├── service.go           Business logic returning ApiResponse (mirrors auth.service.ts)
│   │       └── routes.go            Router + inline controllers (mirrors auth.routes.ts)
│   │
│   └── utils/
│       ├── http_utils.go            ApiResponse, SendResponse, HttpWriter, ParseBody, M
│       └── utils.go                 GetEnv, GetParam, GetIP
│
├── scripts/
│   └── rename-module.sh             Rename the Go module path across the whole project
│
└── docker/
    └── docker-compose.yaml          PostgreSQL service for local development
```

---

## Adding a New Route Module

1. Create the module directory:

```
internal/modules/<name>/
  schema.go   — input types + Validate() methods
  service.go  — business logic returning utils.ApiResponse
  routes.go   — RegisterRoutes() + inline controller functions
```

2. `schema.go` — define and validate inputs:

```go
package thing

type CreateInput struct {
    Name string `json:"name"`
}

func (i *CreateInput) Validate() error {
    if strings.TrimSpace(i.Name) == "" {
        return errors.New("name is required")
    }
    return nil
}
```

3. `service.go` — business logic, returns `utils.ApiResponse`, never touches `http.ResponseWriter`:

```go
func (s *Service) Create(ctx context.Context, input CreateInput) utils.ApiResponse {
    thing, err := s.repo.CreateThing(ctx, repository.CreateThingParams{Name: input.Name})
    if err != nil {
        return utils.ApiError("failed to create thing", err.Error(), 500)
    }
    return utils.ApiSuccess("thing created", map[string]any{"thing": thing}, 201)
}
```

4. `routes.go` — wire routes and controllers:

```go
func RegisterRoutes(r *mux.Router, pool *pgxpool.Pool) {
    svc := newService(pool)
    auth := middlewares.NewAuthMiddleware()

    sub := r.PathPrefix("/things").Subrouter()
    sub.HandleFunc("", list(svc)).Methods(http.MethodGet)
    sub.Handle("", auth.Guard(http.HandlerFunc(create(svc)))).Methods(http.MethodPost)
}

func create(svc *Service) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var input CreateInput
        if err := utils.ParseBody(r, &input); err != nil {
            utils.SendResponse(w, utils.ApiError(err.Error(), nil, http.StatusBadRequest))
            return
        }
        if err := input.Validate(); err != nil {
            utils.SendResponse(w, utils.ApiError(err.Error(), nil, http.StatusUnprocessableEntity))
            return
        }
        utils.SendResponse(w, svc.Create(r.Context(), input))
    }
}
```

5. Register in `internal/modules/routes.go` — one line:

```go
func Register(apiRouter *mux.Router, pool *pgxpool.Pool, startTime time.Time) {
    health.RegisterRoutes(apiRouter, pool, startTime)
    auth.RegisterRoutes(apiRouter, pool)
    thing.RegisterRoutes(apiRouter, pool) // ← add here
}
```

---

## API Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/api/v1/health` | — | Server health, DB ping, memory, uptime |
| `POST` | `/api/v1/auth/register` | — | Register a new user |
| `POST` | `/api/v1/auth/login` | — | Login, returns access + refresh tokens |
| `GET` | `/api/v1/auth/me` | Bearer | Current authenticated user |
| `POST` | `/api/v1/auth/refresh` | — | Exchange refresh token for new access token |

---

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `DB_URL` | *(built from config.yaml)* | Full PostgreSQL DSN — overrides individual DB fields |
| `JWT_SECRET` | `change-me-in-production` | Access token signing secret |
| `JWT_REFRESH_SECRET` | `change-me-refresh-secret` | Refresh token signing secret |
| `CORS_ORIGINS` | `http://localhost:3000,...` | Comma-separated allowed origins |
| `PORT` | `8080` | HTTP listen port — overrides config.yaml |
| `HOST` | `0.0.0.0` | HTTP bind address — overrides config.yaml |

---

## Migrations

```bash
# Create a new migration
make migrate-new
# → prompts for a name, creates 000002_<name>.up.sql and .down.sql

# Apply all pending
make migrate-up

# Roll back one
make migrate-down

# Check current version
make migrate-status
```

Migration file format:

```sql
-- 000002_add_posts.up.sql
CREATE TABLE posts (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  title TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 000002_add_posts.down.sql
DROP TABLE IF EXISTS posts;
```

After adding queries, regenerate the repository:

```bash
make sqlc-gen
```
