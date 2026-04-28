BACKEND_DIR := ./backend
WEB_DIR     := ./web
MOBILE_DIR  := ./mobile
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

.PHONY: help \
        setup keygen clobber \
        infra\:up infra\:down infra\:reset logs logs\:db logs\:mail status \
        db\:migrate db\:rollback db\:reset db\:create db\:psql \
        gen gen\:openapi \
        test test\:backend test\:backend\:coverage test\:web test\:web\:coverage test\:mobile test\:mobile\:coverage test\:contract test\:all \
        lint lint\:backend lint\:web lint\:mobile typecheck\:mobile \
        format format\:backend format\:web format\:mobile \
        dev dev\:backend dev\:web dev\:mobile dev\:mobile\:ios dev\:mobile\:android \
        build\:backend build\:web build\:mobile \
        routes docs check\:parity check\:skills check\:staged preflight

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
	  }' $(MAKEFILE_LIST) | sed 's/\\:/:/g'

# ============================================================================
## group: setup
# ============================================================================

setup: infra\:up ## First-time setup: infra, deps, .env files, migrations, JWT keys
	cp -n .env.example .env || true
	cp -n $(BACKEND_DIR)/.env.example $(BACKEND_DIR)/.env || true
	cp -n $(WEB_DIR)/.env.example $(WEB_DIR)/.env.local || true
	cp -n $(MOBILE_DIR)/.env.example $(MOBILE_DIR)/.env || true
	cd $(BACKEND_DIR) && go mod download
	cd $(WEB_DIR) && npm install
	cd $(MOBILE_DIR) && npm install
	@echo "Mobile (Expo): scan the QR from \`make dev:mobile\` with Expo Go on a real device."
	$(MAKE) keygen
	@bash scripts/wait-for-postgres.sh
	$(MAKE) db\:migrate
	@echo ""
	@echo "Setup complete. Edit backend/.env with your Google OAuth credentials."
	@echo "View magic link emails at http://localhost:8025 (Mailpit)"

keygen: ## Generate Ed25519 keypair for JWT signing
	cd $(BACKEND_DIR) && $(OPENSSL) genpkey -algorithm ed25519 -out jwt_private.pem
	cd $(BACKEND_DIR) && $(OPENSSL) pkey -in jwt_private.pem -pubout -out jwt_public.pem
	@echo "JWT keys generated in $(BACKEND_DIR)/."

clobber: ## Nuclear clean — remove build artifacts, node_modules, Expo caches
	rm -rf $(BACKEND_DIR)/bin
	rm -rf $(WEB_DIR)/.next $(WEB_DIR)/node_modules
	rm -rf $(MOBILE_DIR)/node_modules $(MOBILE_DIR)/.expo $(MOBILE_DIR)/dist

# ============================================================================
## group: infra
# ============================================================================

infra\:up: ## Start PostgreSQL and Mailpit
	docker compose up -d postgres mailpit

infra\:down: ## Stop all infrastructure
	docker compose down

infra\:reset: ## Destroy volumes and restart
	docker compose down -v
	docker compose up -d postgres mailpit

status: ## Report which services are reachable (backend, web, mailpit, db)
	@bash scripts/status.sh

logs: ## Tail all docker compose logs
	docker compose logs -f

logs\:db: ## Tail PostgreSQL logs
	docker compose logs -f postgres

logs\:mail: ## Tail Mailpit logs
	docker compose logs -f mailpit

# ============================================================================
## group: db
# ============================================================================

db\:migrate: ## Run pending migrations
	migrate -path $(BACKEND_DIR)/migrations -database "$(DB_URL)" up

db\:rollback: ## Rollback last migration
	migrate -path $(BACKEND_DIR)/migrations -database "$(DB_URL)" down 1

db\:reset: ## Drop, recreate, migrate — DESTROYS ALL DATA
	migrate -path $(BACKEND_DIR)/migrations -database "$(DB_URL)" drop -f
	$(MAKE) db\:migrate

db\:create: ## Create new migration (usage: make db:create NAME=add_orders)
	migrate create -ext sql -dir $(BACKEND_DIR)/migrations -seq $(NAME)

db\:psql: ## Open psql against the dev database
	docker compose exec postgres psql $(DB_URL)

# ============================================================================
## group: gen
# ============================================================================

gen: gen\:openapi ## Run all codegen (currently just gen:openapi)

gen\:openapi: ## Export Huma's in-memory spec to backend/api/openapi.yaml + regen web TS types. Required because preflight runs without a server, and codegen + check:parity read the on-disk file.
	cd $(BACKEND_DIR) && go run ./cmd/gen-openapi
	cd $(WEB_DIR) && npx openapi-typescript ../backend/api/openapi.yaml -o src/lib/api-types.ts

# ============================================================================
## group: test
# ============================================================================

test: ## Fast suite (backend + web + mobile unit tests, parallel).
	@$(MAKE) -j3 test\:backend test\:web test\:mobile

test\:backend: ## Backend Go tests with race detector
	cd $(BACKEND_DIR) && go test ./... -v -race -count=1

test\:backend\:coverage: ## Backend tests + coverage report (fails under 35% — phased ramp toward 80%)
	cd $(BACKEND_DIR) && go test ./... -race -count=1 -coverprofile=coverage.out
	@cd $(BACKEND_DIR) && go tool cover -func coverage.out | awk '/^total:/{ \
		pct = substr($$3, 1, length($$3)-1) + 0; \
		printf "backend coverage: %.1f%%\n", pct; \
		if (pct < 35.0) { print "FAIL: coverage below 35% floor"; exit 1 } \
	}'

test\:web: ## Web Vitest suite
	cd $(WEB_DIR) && npm run test

test\:web\:coverage: ## Web tests with coverage (fails under 25% lines/branches — phased ramp toward 80%)
	cd $(WEB_DIR) && npm run test:coverage

test\:mobile: ## Mobile (Expo) Jest suite
	cd $(MOBILE_DIR) && npm test

test\:mobile\:coverage: ## Mobile tests with coverage (fails under 25% lines — matches web)
	cd $(MOBILE_DIR) && npm test -- --coverage

test\:contract: ## Schemathesis contract tests (requires running backend; heavy — CI-only)
	bash scripts/schemathesis.sh

test\:all: test test\:contract ## Everything: fast suite + contract tests

# ============================================================================
## group: lint
# ============================================================================

lint: ## All linters (backend + web + mobile, parallel).
	@$(MAKE) -j3 lint\:backend lint\:web lint\:mobile

lint\:backend: ## golangci-lint backend
	cd $(BACKEND_DIR) && golangci-lint run ./...

lint\:web: ## ESLint + TypeScript web
	cd $(WEB_DIR) && npm run lint

lint\:mobile: ## ESLint + TypeScript mobile (Expo)
	cd $(MOBILE_DIR) && npm run lint
	cd $(MOBILE_DIR) && npm run typecheck

typecheck\:mobile: ## TypeScript-only check for the Expo app
	cd $(MOBILE_DIR) && npm run typecheck

format: ## Format all code (backend + web + mobile)
	@$(MAKE) -j3 format\:backend format\:web format\:mobile

format\:backend: ## gofmt backend
	cd $(BACKEND_DIR) && gofmt -w .

format\:web: ## prettier web
	cd $(WEB_DIR) && npx prettier --write "src/**/*.{ts,tsx,json,css}"

format\:mobile: ## prettier mobile (Expo)
	cd $(MOBILE_DIR) && npm run format

# ============================================================================
## group: dev
# ============================================================================

dev: infra\:up ## Backend (air) + web (Next.js) — both in one shell
	@echo "Starting backend (8000) + web (3000)... Ctrl-C stops both."
	@bash -c 'trap "kill 0" INT TERM EXIT; \
		(cd $(BACKEND_DIR) && air) & \
		(cd $(WEB_DIR) && npm run dev) & \
		wait'

dev\:backend: ## Backend only (hot reload via air)
	cd $(BACKEND_DIR) && air

dev\:web: ## Web only (Next.js on port 3000)
	cd $(WEB_DIR) && npm run dev

dev\:mobile: ## Mobile (Expo) — interactive Metro bundler with QR for Expo Go
	cd $(MOBILE_DIR) && npx expo start

dev\:mobile\:ios: ## Mobile (Expo) — auto-launch the iOS simulator
	cd $(MOBILE_DIR) && npx expo start --ios

dev\:mobile\:android: ## Mobile (Expo) — auto-launch a connected Android device/emulator
	cd $(MOBILE_DIR) && npx expo start --android

# ============================================================================
## group: build
# ============================================================================

build\:backend: ## Backend binary to backend/bin/server
	cd $(BACKEND_DIR) && go build -o bin/server ./cmd/server

build\:web: ## Web production bundle
	cd $(WEB_DIR) && npm run build

build\:mobile: ## Mobile (Expo) — static export under mobile/dist (preview/web)
	cd $(MOBILE_DIR) && npx expo export

# ============================================================================
## group: ops
# ============================================================================

routes: ## Print HTTP route table from running backend (dev-only endpoint)
	@bash scripts/routes.sh

docs: ## Open the auto-generated API docs (requires `make dev:backend` running)
	@echo "Huma serves rendered API docs at:"
	@echo "  http://localhost:8000/docs"
	@echo "Spec endpoints:"
	@echo "  http://localhost:8000/openapi.yaml"
	@echo "  http://localhost:8000/openapi.json"
	@echo ""
	@echo "Run \`make dev:backend\` first if not already running."

check\:parity: ## Verify every /api/ spec path has a consumer in web/src or mobile
	cd $(BACKEND_DIR) && go run ./cmd/check-parity

check\:skills: ## Verify exemplar paths and <!-- expect: symbol --> annotations in each SKILL.md
	@bash scripts/check-skills.sh

check\:staged: ## Fail if git has untracked or unstaged changes
	@bash scripts/check-staged.sh

preflight: ## Every gate a PR must pass: lint + test + check:parity + check:skills + check:staged
	@bash scripts/preflight.sh
