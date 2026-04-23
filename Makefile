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
        db\:migrate db\:rollback db\:reset db\:seed db\:create db\:psql \
        gen gen\:openapi gen\:mobile \
        test test\:backend test\:web test\:mobile test\:contract test\:all \
        lint lint\:backend lint\:web lint\:mobile \
        format format\:backend format\:web format\:mobile \
        dev dev\:backend dev\:web dev\:mobile \
        build\:backend build\:web build\:mobile\:apk build\:mobile\:ios \
        routes docs check\:openapi check\:parity check\:staged preflight

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
	cd $(MOBILE_DIR) && flutter pub get
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

clobber: ## Nuclear clean — remove build artifacts, node_modules, flutter build, dart_tool
	rm -rf $(BACKEND_DIR)/bin
	rm -rf $(WEB_DIR)/.next $(WEB_DIR)/node_modules
	rm -rf $(MOBILE_DIR)/build $(MOBILE_DIR)/.dart_tool

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

db\:seed: ## Seed dev data (idempotent; no-op until backend/cmd/seed exists)
	@if [ -d $(BACKEND_DIR)/cmd/seed ]; then \
		cd $(BACKEND_DIR) && go run ./cmd/seed; \
	else \
		echo "No seed binary yet. Create backend/cmd/seed/main.go to populate dev data."; \
	fi

db\:create: ## Create new migration (usage: make db:create NAME=add_orders)
	migrate create -ext sql -dir $(BACKEND_DIR)/migrations -seq $(NAME)

db\:psql: ## Open psql against the dev database
	docker compose exec postgres psql $(DB_URL)

# ============================================================================
## group: gen
# ============================================================================

gen: gen\:openapi ## Regenerate all types from the OpenAPI spec (alias for gen:openapi)

gen\:openapi: ## Regenerate Go + TypeScript types from backend/api/openapi.yaml
	cd $(BACKEND_DIR) && go tool oapi-codegen -config internal/gen/api/config.yaml api/openapi.yaml
	cd $(WEB_DIR) && npx openapi-typescript ../backend/api/openapi.yaml -o src/lib/api-types.ts

gen\:mobile: ## Flutter localizations (run after editing lib/l10n/*.arb)
	cd $(MOBILE_DIR) && flutter gen-l10n

# ============================================================================
## group: test
# ============================================================================

test: ## Fast suite (backend + web + mobile unit tests, parallel)
	@$(MAKE) -j3 test\:backend test\:web test\:mobile

test\:backend: ## Backend Go tests with race detector
	cd $(BACKEND_DIR) && go test ./... -v -race -count=1

test\:web: ## Web Vitest suite
	cd $(WEB_DIR) && npm run test

test\:mobile: ## Flutter unit + widget tests
	cd $(MOBILE_DIR) && flutter test

test\:contract: ## Schemathesis contract tests (requires running backend; heavy — CI-only)
	bash scripts/schemathesis.sh

test\:all: test test\:contract ## Everything: fast suite + contract tests

# ============================================================================
## group: lint
# ============================================================================

lint: ## All linters (backend + web + mobile, parallel)
	@$(MAKE) -j3 lint\:backend lint\:web lint\:mobile

lint\:backend: ## golangci-lint backend
	cd $(BACKEND_DIR) && golangci-lint run ./...

lint\:web: ## ESLint + TypeScript web
	cd $(WEB_DIR) && npm run lint

lint\:mobile: ## dart analyze mobile
	cd $(MOBILE_DIR) && dart analyze

format: ## Format all code (backend + web + mobile)
	@$(MAKE) -j3 format\:backend format\:web format\:mobile

format\:backend: ## gofmt backend
	cd $(BACKEND_DIR) && gofmt -w .

format\:web: ## prettier web
	cd $(WEB_DIR) && npx prettier --write "src/**/*.{ts,tsx,json,css}"

format\:mobile: ## dart format mobile
	cd $(MOBILE_DIR) && dart format .

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

dev\:mobile: _ensure-simulator ## Mobile — boots simulator if needed, runs flutter run
	cd $(MOBILE_DIR) && flutter run --dart-define-from-file=.env

_ensure-simulator:
	@if ! xcrun simctl list devices booted 2>/dev/null | grep -q "Booted"; then \
		echo "No simulator running. Available devices:"; \
		echo ""; \
		xcrun simctl list devices available | grep -E "iPhone|iPad" | cat -n; \
		echo ""; \
		read -p "Enter number to boot (or press Enter for first iPhone): " choice; \
		if [ -z "$$choice" ]; then \
			UDID=$$(xcrun simctl list devices available | grep "iPhone" | head -1 | grep -oE '[A-F0-9-]{36}'); \
		else \
			UDID=$$(xcrun simctl list devices available | grep -E "iPhone|iPad" | sed -n "$${choice}p" | grep -oE '[A-F0-9-]{36}'); \
		fi; \
		if [ -z "$$UDID" ]; then echo "Invalid selection."; exit 1; fi; \
		echo "Booting $$UDID..."; xcrun simctl boot "$$UDID"; \
		open -a Simulator; sleep 5; \
	fi

# ============================================================================
## group: build
# ============================================================================

build\:backend: ## Backend binary to backend/bin/server
	cd $(BACKEND_DIR) && go build -o bin/server ./cmd/server

build\:web: ## Web production bundle
	cd $(WEB_DIR) && npm run build

build\:mobile\:apk: ## Android debug APK
	cd $(MOBILE_DIR) && flutter build apk --debug

build\:mobile\:ios: ## iOS simulator build
	cd $(MOBILE_DIR) && flutter build ios --simulator

# ============================================================================
## group: ops
# ============================================================================

routes: ## Print HTTP route table from running backend (dev-only endpoint)
	@bash scripts/routes.sh

docs: ## Serve Swagger UI for backend/api/openapi.yaml on :$(DOCS_PORT)
	@echo "Swagger UI: http://localhost:$(DOCS_PORT)"
	@echo "Ctrl-C to stop."
	docker run --rm -p $(DOCS_PORT):8080 \
		-e SWAGGER_JSON=/spec/openapi.yaml \
		-v $(PWD)/backend/api:/spec:ro \
		swaggerapi/swagger-ui

check\:openapi: ## Verify Dart DTOs stay in sync with OpenAPI spec
	cd $(BACKEND_DIR) && go run ./cmd/check-dto-drift

check\:parity: ## Verify every /api/ spec path has a web or mobile consumer
	cd $(BACKEND_DIR) && go run ./cmd/check-parity

check\:staged: ## Fail if git has untracked or unstaged changes
	@bash scripts/check-staged.sh

preflight: ## Every gate a PR must pass: lint + test + check:openapi + check:parity + check:skills + check:staged
	@bash scripts/preflight.sh
