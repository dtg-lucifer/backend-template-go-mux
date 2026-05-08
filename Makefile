## ── Project metadata ──────────────────────────────────────────────────────────
APP_NAME    := app
CMD_PATH    := .
BIN_DIR     := ./bin
BIN_FILE    := $(BIN_DIR)/$(APP_NAME)
LOGS_DIR    := ./logs

## ── Tools ─────────────────────────────────────────────────────────────────────
GO              := go
SQLC            := sqlc
GOLANGCI_LINT   := golangci-lint
MIGRATE         := migrate
AIR             := air

## ── DB config (override via env or .env) ──────────────────────────────────────
DB_URL          ?= postgres://postgres:postgres@localhost:5432/app_db?sslmode=disable
MIGRATIONS_DIR  ?= internal/db/migrations

## ── Flags ─────────────────────────────────────────────────────────────────────
GO_FILES := $(shell find . -type f -name '*.go' -not -path "./vendor/*")

.DEFAULT_GOAL := build

# ── Help ───────────────────────────────────────────────────────────────────────
.PHONY: help
help: ## Show this help.
	@echo "Usage: make <target>"
	@awk 'BEGIN {FS = ":.*##"; printf "\nTargets:\n"} /^[a-zA-Z0-9_.-]+:.*##/ { printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# ── Install dev tools ──────────────────────────────────────────────────────────
.PHONY: install
install: install-sqlc install-air install-migrate install-lint ## Install all dev tools.
	@echo ">> All tools installed."

.PHONY: install-sqlc
install-sqlc: ## Install sqlc.
	@echo ">> Installing sqlc…"
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

.PHONY: install-air
install-air: ## Install air (hot reload).
	@echo ">> Installing air (hot reload)…"
	@go install github.com/air-verse/air@latest

.PHONY: install-migrate
install-migrate: ## Install golang-migrate.
	@echo ">> Installing golang-migrate…"
	@go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

.PHONY: install-lint
install-lint: ## Install golangci-lint.
	@echo ">> Installing golangci-lint…"
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh \
		| sh -s -- -b $$(go env GOPATH)/bin v2.11.4

# ── Build ──────────────────────────────────────────────────────────────────────
.PHONY: build
build: clean ## Build the binary.
	@echo ">> Building binary…"
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_FILE) $(CMD_PATH)
	@echo ">> Binary: $(BIN_FILE)"

# ── Run ────────────────────────────────────────────────────────────────────────
.PHONY: run
run: build ## Build and run the binary.
	@echo ">> Running $(APP_NAME)…"
	$(BIN_FILE)

# ── Dev (hot reload via Air) ───────────────────────────────────────────────────
.PHONY: dev
dev: ## Run with hot reload (Air).
	@$(AIR)

# ── Test ───────────────────────────────────────────────────────────────────────
.PHONY: test
test: ## Run tests with race detection and coverage.
	@echo ">> Running tests…"
	$(GO) test ./... -v -race -cover

# ── Lint & Format ──────────────────────────────────────────────────────────────
.PHONY: lint
lint: ## Run golangci-lint.
	@echo ">> Linting…"
	-$(GOLANGCI_LINT) run ./... || true

.PHONY: fmt
fmt: ## Format Go files.
	@echo ">> Formatting…"
	$(GO) fmt ./...
	@echo ">> Running GFMT in hidden mode"
	@gofmt -s -w $(GO_FILES)

# ── SQLC code generation ───────────────────────────────────────────────────────
.PHONY: sqlc-gen
sqlc-gen: ## Generate code with sqlc.
	@echo ">> Generating repository code from SQL…"
	$(SQLC) generate

# ── Database ───────────────────────────────────────────────────────────────────
.PHONY: db
db: ## Start PostgreSQL via Docker Compose.
	@echo ">> Starting PostgreSQL…"
	@docker compose -f docker/docker-compose.yaml up -d postgres

.PHONY: rabbitmq
rabbitmq: ## Start RabbitMQ via Docker Compose.
	@echo ">> Starting RabbitMQ…"
	@docker compose -f docker/docker-compose.yaml up -d rabbitmq

.PHONY: redis
redis: ## Start Redis via Docker Compose.
	@echo ">> Starting Redis…"
	@docker compose -f docker/docker-compose.yaml up -d redis

.PHONY: infra
infra: ## Start PostgreSQL, RabbitMQ, and Redis.
	@echo ">> Starting all infrastructure (PostgreSQL + RabbitMQ + Redis)…"
	@docker compose -f docker/docker-compose.yaml up -d

.PHONY: db-stop
db-stop: ## Stop all infrastructure services.
	@echo ">> Stopping all infrastructure…"
	@docker compose -f docker/docker-compose.yaml down

# ── Migrations ─────────────────────────────────────────────────────────────────
.PHONY: migrate-up
migrate-up: ## Apply all pending migrations.
	@echo ">> Applying all pending migrations…"
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

.PHONY: migrate-down
migrate-down: ## Roll back the last migration.
	@echo ">> Rolling back last migration…"
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1

.PHONY: migrate-status
migrate-status: ## Show current migration version.
	@echo ">> Migration status…"
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" version

.PHONY: migrate-drop
migrate-drop: ## Drop all migrations (destructive).
	@echo ">> Dropping all migrations (DESTRUCTIVE)…"
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" drop -f

.PHONY: migrate-new
migrate-new: ## Create a new migration (prompts for name).
	@read -p "Migration name: " name; \
		$(MIGRATE) create -ext sql -dir $(MIGRATIONS_DIR) -seq "$$name"; \
		echo ">> Created migration: $$name"

# ── Clean ──────────────────────────────────────────────────────────────────────
.PHONY: clean
clean: ## Remove build artifacts and logs.
	@echo ">> Cleaning build artifacts…"
	@rm -rf $(BIN_DIR)
	@echo ">> Removing log files…"
	@rm -rf $(LOGS_DIR)
