BACKEND_DIR := ./backend
WEB_DIR     := ./web
MOBILE_DIR  := ./mobile
CORE_DIR    := ./packages/core
BACKEND_URL := http://localhost:8000
DOCS_PORT   := 8088

# Load the root .env (if present) so docker-compose variables are available
# to every recipe + subprocess. Fresh clones without a .env fall through to
# the defaults below.
-include .env
export

POSTGRES_USER     ?= initium
POSTGRES_PASSWORD ?= initium
POSTGRES_DB       ?= initium_dev
POSTGRES_PORT     ?= 5432

# Cascade Postgres credentials into the DB_* names the Go backend reads
# via godotenv. Exporting means `cd $(BACKEND_DIR) && go run ./cmd/server`
# picks these up even though they're not in backend/.env. godotenv.Load
# doesn't overwrite already-set env, so backend/.env's hardcoded values
# are overridden by these when invoked via make.
DB_HOST     ?= 127.0.0.1
DB_PORT     ?= $(POSTGRES_PORT)
DB_USER     ?= $(POSTGRES_USER)
DB_PASSWORD ?= $(POSTGRES_PASSWORD)
DB_NAME     ?= $(POSTGRES_DB)
export DB_HOST DB_PORT DB_USER DB_PASSWORD DB_NAME

DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

OPENSSL := $(shell which /opt/homebrew/opt/openssl/bin/openssl 2>/dev/null || which /usr/local/opt/openssl/bin/openssl 2>/dev/null || echo openssl)
TOOLS_DIR := $(BACKEND_DIR)/bin
MIGRATE_BIN := $(TOOLS_DIR)/migrate
AIR_BIN := $(TOOLS_DIR)/air
SQLC_BIN := $(TOOLS_DIR)/sqlc
GOLANGCI_LINT_BIN := $(TOOLS_DIR)/golangci-lint
JWT_PRIVATE_KEY := $(BACKEND_DIR)/jwt_private.pem
JWT_PUBLIC_KEY := $(BACKEND_DIR)/jwt_public.pem

AIR_VERSION := v1.65.3
SQLC_VERSION := v1.31.1
GOLANGCI_LINT_VERSION := v2.12.2

.PHONY: help \
        setup tools keygen clobber \
        infra/up infra/down infra/reset logs logs/db logs/mail status \
        db/migrate db/rollback db/reset db/create db/psql \
        gen gen/openapi gen/sqlc \
        test test/backend test/backend/coverage test/core test/core/coverage test/web test/web/coverage test/mobile test/mobile/coverage test/contract \
        lint lint/backend lint/core lint/web lint/mobile typecheck/mobile \
        format format/backend format/core format/web format/mobile \
        dev dev/backend dev/web dev/mobile dev/mobile/ios dev/mobile/android \
        build/backend build/web build/mobile \
        routes docs check/parity check/skills check/staged preflight

help: ## Show this help (grouped by namespace)
	@awk 'BEGIN { FS = " ## "; group = ""; prev_group = ""; } \
	  /^## group:/ { \
	    group = $$0; sub(/^## group: */, "", group); next; \
	  } \
	  / ## / && /^[a-zA-Z]/ { \
	    split($$1, parts, " "); tgt = parts[1]; \
	    sub(/:$$/, "", tgt); \
	    if (group != prev_group) { printf "\n\033[1;33m%s\033[0m\n", toupper(group); prev_group = group; } \
	    printf "  \033[36m%-24s\033[0m %s\n", tgt, $$2; \
	  }' $(MAKEFILE_LIST)

# ============================================================================
## group: setup
# ============================================================================

setup: infra/up ## First-time setup: infra, deps, migrations, JWT keys
	cp -n .env.example .env || true
	cp -n $(BACKEND_DIR)/.env.example $(BACKEND_DIR)/.env || true
	cp -n $(WEB_DIR)/.env.example $(WEB_DIR)/.env.local || true
	cp -n $(MOBILE_DIR)/.env.example $(MOBILE_DIR)/.env || true
	cd $(BACKEND_DIR) && go mod download
	npm install
	@echo "Mobile (Expo): scan the QR from \`make dev/mobile\` with Expo Go on a real device."
	$(MAKE) keygen
	@bash scripts/wait-for-postgres.sh
	$(MAKE) db/migrate
	@echo ""
	@echo "Setup complete. Edit backend/.env with your Google OAuth credentials."
	@echo "View magic link emails at http://localhost:8025 (Mailpit)"

tools: $(MIGRATE_BIN) $(AIR_BIN) $(SQLC_BIN) $(GOLANGCI_LINT_BIN) ## Build pinned project-local CLI tools

$(MIGRATE_BIN): Makefile $(BACKEND_DIR)/go.mod $(BACKEND_DIR)/go.sum
	cd $(BACKEND_DIR) && go build -tags postgres -o bin/migrate github.com/golang-migrate/migrate/v4/cmd/migrate

$(AIR_BIN): Makefile
	GOBIN=$(abspath $(TOOLS_DIR)) go install github.com/air-verse/air@$(AIR_VERSION)

$(SQLC_BIN): Makefile
	GOBIN=$(abspath $(TOOLS_DIR)) go install github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION)

$(GOLANGCI_LINT_BIN): Makefile
	GOBIN=$(abspath $(TOOLS_DIR)) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

keygen: $(JWT_PRIVATE_KEY) $(JWT_PUBLIC_KEY) ## Ensure an Ed25519 JWT keypair exists
	@echo "JWT keys ready in $(BACKEND_DIR)/."

$(JWT_PRIVATE_KEY):
	cd $(BACKEND_DIR) && $(OPENSSL) genpkey -algorithm ed25519 -out jwt_private.pem

$(JWT_PUBLIC_KEY): $(JWT_PRIVATE_KEY)
	cd $(BACKEND_DIR) && $(OPENSSL) pkey -in jwt_private.pem -pubout -out jwt_public.pem

clobber: ## Nuclear clean — remove build artifacts, node_modules, Expo caches
	rm -rf $(BACKEND_DIR)/bin
	rm -rf ./node_modules $(WEB_DIR)/.next
	rm -rf $(MOBILE_DIR)/.expo $(MOBILE_DIR)/dist

# ============================================================================
## group: infra
# ============================================================================

infra/up: ## Start PostgreSQL and Mailpit
	docker compose up -d postgres mailpit

infra/down: ## Stop all infrastructure
	docker compose down

infra/reset: ## Destroy volumes and restart
	docker compose down -v
	docker compose up -d postgres mailpit

status: ## Report which services are reachable (backend, web, mailpit, db)
	@bash scripts/status.sh

logs: ## Tail all docker compose logs
	docker compose logs -f

logs/db: ## Tail PostgreSQL logs
	docker compose logs -f postgres

logs/mail: ## Tail Mailpit logs
	docker compose logs -f mailpit

# ============================================================================
## group: db
# ============================================================================

db/migrate: $(MIGRATE_BIN) ## Run pending migrations
	$(MIGRATE_BIN) -path $(BACKEND_DIR)/migrations -database "$(DB_URL)" up

db/rollback: $(MIGRATE_BIN) ## Roll back one migration
	$(MIGRATE_BIN) -path $(BACKEND_DIR)/migrations -database "$(DB_URL)" down 1

db/reset: $(MIGRATE_BIN) ## Drop and remigrate the dev database — DESTROYS ALL DATA
	$(MIGRATE_BIN) -path $(BACKEND_DIR)/migrations -database "$(DB_URL)" drop -f
	$(MAKE) db/migrate

db/create: $(MIGRATE_BIN) ## Create a migration pair (usage: make db/create NAME=add_orders)
	@test -n "$(NAME)" || (echo "Usage: make db/create NAME=add_orders" >&2; exit 2)
	$(MIGRATE_BIN) create -ext sql -dir $(BACKEND_DIR)/migrations -seq $(NAME)

db/psql: ## Open psql inside the PostgreSQL container
	docker compose exec postgres psql -U "$(DB_USER)" -d "$(DB_NAME)"

# ============================================================================
## group: gen
# ============================================================================

gen: gen/openapi gen/sqlc ## Run all codegen (OpenAPI spec + sqlc queries)

gen/openapi: ## Export Huma's spec + regenerate shared TypeScript contract types
	cd $(BACKEND_DIR) && go run ./cmd/gen-openapi
	npx openapi-typescript $(BACKEND_DIR)/api/openapi.yaml -o $(CORE_DIR)/src/generated/api-types.ts

gen/sqlc: $(SQLC_BIN) ## Regenerate typed query code from SQL queries
	cd $(BACKEND_DIR) && $(abspath $(SQLC_BIN)) generate

# ============================================================================
## group: test
# ============================================================================

test: ## Fast suite (backend + shared core + web + mobile unit tests, parallel).
	@$(MAKE) -j4 test/backend test/core test/web test/mobile

test/backend: ## Backend Go tests with race detector
	cd $(BACKEND_DIR) && go test ./... -v -race -count=1

test/backend/coverage: ## Backend tests + coverage report (35% floor)
	cd $(BACKEND_DIR) && go test ./... -race -count=1 -coverprofile=coverage.out
	@cd $(BACKEND_DIR) && go tool cover -func coverage.out | awk '/^total:/{ \
		pct = substr($$3, 1, length($$3)-1) + 0; \
		printf "backend coverage: %.1f%%\n", pct; \
		if (pct < 35.0) { print "FAIL: coverage below 35% floor"; exit 1 } \
	}'

test/core: ## Shared Effect core Vitest suite
	npm run test --workspace @initium/core

test/core/coverage: ## Shared Effect core tests with coverage
	npm run test:coverage --workspace @initium/core

test/web: ## Web Vitest suite
	cd $(WEB_DIR) && npm run test

test/web/coverage: ## Web tests with coverage (25% lines/branches floor)
	cd $(WEB_DIR) && npm run test:coverage

test/mobile: ## Mobile (Expo) Jest suite
	cd $(MOBILE_DIR) && npm test

test/mobile/coverage: ## Mobile tests with coverage (25% lines floor)
	cd $(MOBILE_DIR) && npm test -- --coverage

test/contract: ## Schemathesis smoke test (run backend with DEV_BYPASS_AUTH=false)
	bash scripts/schemathesis.sh

# ============================================================================
## group: lint
# ============================================================================

lint: ## All linters/typechecks (backend + shared core + web + mobile, parallel).
	@$(MAKE) -j4 lint/backend lint/core lint/web lint/mobile

lint/backend: $(GOLANGCI_LINT_BIN) ## golangci-lint backend
	cd $(BACKEND_DIR) && $(abspath $(GOLANGCI_LINT_BIN)) run ./...

lint/core: ## Typecheck the shared Effect core
	npm run typecheck --workspace @initium/core

lint/web: ## ESLint + TypeScript web
	cd $(WEB_DIR) && npm run lint

lint/mobile: ## ESLint + TypeScript mobile (Expo)
	cd $(MOBILE_DIR) && npm run lint
	cd $(MOBILE_DIR) && npm run typecheck

typecheck/mobile: ## TypeScript-only check for the Expo app
	cd $(MOBILE_DIR) && npm run typecheck

format: ## Format all code (backend + shared core + web + mobile)
	@$(MAKE) -j4 format/backend format/core format/web format/mobile

format/backend: ## Format backend Go packages
	cd $(BACKEND_DIR) && go fmt ./...

format/core: ## Prettier shared Effect core
	npx prettier --write "$(CORE_DIR)/src/*.ts" "$(CORE_DIR)/__tests__/*.ts"

format/web: ## Prettier web
	cd $(WEB_DIR) && npx prettier --write "src/**/*.{ts,tsx,json,css}"

format/mobile: ## Prettier mobile (Expo)
	cd $(MOBILE_DIR) && npm run format

# ============================================================================
## group: dev
# ============================================================================

dev: infra/up $(AIR_BIN) ## Backend (Air) + web (Next.js) in one shell
	@echo "Starting backend (8000) + web (3000)... Ctrl-C stops both."
	@bash -c 'trap "kill 0" INT TERM EXIT; \
		(cd $(BACKEND_DIR) && $(abspath $(AIR_BIN))) & \
		(cd $(WEB_DIR) && npm run dev) & \
		wait'

dev/backend: $(AIR_BIN) ## Backend only (hot reload via Air)
	cd $(BACKEND_DIR) && $(abspath $(AIR_BIN))

dev/web: ## Web only (Next.js on port 3000)
	cd $(WEB_DIR) && npm run dev

dev/mobile: ## Mobile (Expo) — Metro bundler with QR for Expo Go
	cd $(MOBILE_DIR) && npx expo start

dev/mobile/ios: ## Mobile (Expo) — launch the iOS simulator
	cd $(MOBILE_DIR) && npx expo start --ios

dev/mobile/android: ## Mobile (Expo) — launch an Android device/emulator
	cd $(MOBILE_DIR) && npx expo start --android

# ============================================================================
## group: build
# ============================================================================

build/backend: ## Backend binary to backend/bin/server
	cd $(BACKEND_DIR) && go build -o bin/server ./cmd/server

build/web: ## Web production bundle
	cd $(WEB_DIR) && npm run build

build/mobile: ## Expo export for web, Android, and iOS under mobile/dist
	cd $(MOBILE_DIR) && npx expo export

# ============================================================================
## group: ops
# ============================================================================

routes: ## Print HTTP route table from running backend (dev-only endpoint)
	@bash scripts/routes.sh

docs: ## Print API documentation URLs
	@echo "Huma serves rendered API docs at:"
	@echo "  http://localhost:8000/docs"
	@echo "Spec endpoints:"
	@echo "  http://localhost:8000/openapi.yaml"
	@echo "  http://localhost:8000/openapi.json"
	@echo ""
	@echo "Run \`make dev/backend\` first if not already running."

check/parity: ## Verify every API spec path has a web or mobile consumer
	cd $(BACKEND_DIR) && go run ./cmd/check-parity

check/skills: ## Verify skill exemplar paths and expected symbols
	@bash scripts/check-skills.sh

check/staged: ## Fail if work is untracked or not staged
	@bash scripts/check-staged.sh

preflight: ## Every PR gate: lint, coverage, generation, parity, skills, staged work
	@bash scripts/preflight.sh
