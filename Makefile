BACKEND_DIR := ./backend
WEB_DIR := ./web
MOBILE_DIR := ./mobile
DB_URL := postgres://initium:initium@127.0.0.1:5432/initium_dev?sslmode=disable

.PHONY: help setup dev infra-up infra-down infra-reset \
        db-migrate db-rollback db-create \
        backend-dev backend-run backend-test backend-lint backend-build \
        web-dev web-build web-test web-lint \
        mobile-dev mobile-test mobile-gen mobile-build-apk mobile-build-ios mobile-lint \
        test lint logs clean gen gen-backend gen-web \
        keygen

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# --- Setup ---

setup: infra-up ## First-time project setup
	cp -n $(BACKEND_DIR)/.env.example $(BACKEND_DIR)/.env || true
	cp -n $(WEB_DIR)/.env.example $(WEB_DIR)/.env.local || true
	cp -n $(MOBILE_DIR)/.env.example $(MOBILE_DIR)/.env || true
	cd $(BACKEND_DIR) && go mod download
	cd $(WEB_DIR) && npm install
	cd $(MOBILE_DIR) && flutter pub get
	$(MAKE) keygen
	@echo "Waiting for PostgreSQL..."
	@until docker compose exec -T postgres pg_isready -U initium >/dev/null 2>&1; do sleep 1; done
	$(MAKE) db-migrate
	@echo ""
	@echo "Setup complete. Edit backend/.env with your Google OAuth credentials."
	@echo "View magic link emails at http://localhost:8025 (Mailpit)"

dev: infra-up ## Start backend (air) + web dev server (requires air: go install github.com/air-verse/air@latest)
	@echo "Starting backend (8000) + web (3000)... Press Ctrl-C to stop both."
	@bash -c 'trap "kill 0" INT TERM EXIT; \
		(cd $(BACKEND_DIR) && air) & \
		(cd $(WEB_DIR) && npm run dev) & \
		wait'

# --- Infrastructure ---

infra-up: ## Start PostgreSQL and Mailpit
	docker compose up -d postgres mailpit

infra-down: ## Stop all infrastructure
	docker compose down

infra-reset: ## Destroy volumes and restart
	docker compose down -v
	docker compose up -d postgres mailpit

# --- Database ---

db-migrate: ## Run pending migrations
	migrate -path $(BACKEND_DIR)/migrations -database "$(DB_URL)" up

db-rollback: ## Rollback last migration
	migrate -path $(BACKEND_DIR)/migrations -database "$(DB_URL)" down 1

db-create: ## Create new migration (usage: make db-create NAME=add_orders)
	migrate create -ext sql -dir $(BACKEND_DIR)/migrations -seq $(NAME)

# --- Backend ---

backend-dev: ## Run Go backend with hot reload (requires air)
	cd $(BACKEND_DIR) && air

backend-run: ## Run Go backend directly
	cd $(BACKEND_DIR) && go run ./cmd/server

backend-test: ## Run backend tests
	cd $(BACKEND_DIR) && go test ./... -v -race -count=1

backend-lint: ## Lint backend code
	cd $(BACKEND_DIR) && golangci-lint run ./...

backend-build: ## Build backend binary
	cd $(BACKEND_DIR) && go build -o bin/server ./cmd/server

# --- Web ---

web-dev: ## Run Next.js dev server
	cd $(WEB_DIR) && npm run dev

web-build: ## Build web for production
	cd $(WEB_DIR) && npm run build

web-test: ## Run web tests
	cd $(WEB_DIR) && npm run test

web-lint: ## Lint web code
	cd $(WEB_DIR) && npm run lint

# --- Mobile ---

mobile-dev: _ensure-simulator ## Run Flutter app with env config
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
		if [ -z "$$UDID" ]; then \
			echo "Invalid selection."; exit 1; \
		fi; \
		echo "Booting simulator $$UDID..."; \
		xcrun simctl boot "$$UDID"; \
		open -a Simulator; \
		echo "Waiting for simulator..."; \
		sleep 5; \
	fi

mobile-test: ## Run Flutter tests
	cd $(MOBILE_DIR) && flutter test

mobile-gen: ## Run build_runner (required after DTO changes)
	cd $(MOBILE_DIR) && dart run build_runner build --delete-conflicting-outputs

mobile-build-apk: ## Build Android debug APK
	cd $(MOBILE_DIR) && flutter build apk --debug

mobile-build-ios: ## Build iOS simulator build
	cd $(MOBILE_DIR) && flutter build ios --simulator

mobile-lint: ## Lint mobile Dart code
	cd $(MOBILE_DIR) && dart analyze

# --- Aggregate ---

test: ## Run all tests (backend, web, mobile) — fail fast
	$(MAKE) backend-test
	$(MAKE) web-test
	$(MAKE) mobile-test

lint: ## Lint all code (backend, web, mobile)
	$(MAKE) backend-lint
	$(MAKE) web-lint
	$(MAKE) mobile-lint

logs: ## Tail docker compose logs
	docker compose logs -f

clean: ## Remove build artifacts (bin, .next cache, Flutter build)
	rm -rf $(BACKEND_DIR)/bin
	rm -rf $(WEB_DIR)/.next
	rm -rf $(WEB_DIR)/node_modules/.cache
	rm -rf $(MOBILE_DIR)/build

# --- Codegen ---

gen: gen-backend gen-web ## Regenerate API types from backend/api/openapi.yaml (Go + TS)

gen-backend: ## Regenerate Go types from OpenAPI spec (oapi-codegen, pinned in go.mod tools)
	cd $(BACKEND_DIR) && go tool oapi-codegen -config internal/gen/api/config.yaml api/openapi.yaml

gen-web: ## Regenerate TypeScript types from OpenAPI spec (openapi-typescript)
	cd $(WEB_DIR) && npx openapi-typescript ../backend/api/openapi.yaml -o src/lib/api-types.ts

# --- Utils ---

OPENSSL := $(shell which /opt/homebrew/opt/openssl/bin/openssl 2>/dev/null || which /usr/local/opt/openssl/bin/openssl 2>/dev/null || echo openssl)

keygen: ## Generate Ed25519 keypair for JWT signing
	cd $(BACKEND_DIR) && $(OPENSSL) genpkey -algorithm ed25519 -out jwt_private.pem
	cd $(BACKEND_DIR) && $(OPENSSL) pkey -in jwt_private.pem -pubout -out jwt_public.pem
	@echo "JWT keys generated in $(BACKEND_DIR)/."
