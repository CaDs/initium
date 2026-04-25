BACKEND_DIR := ./backend
WEB_DIR     := ./web
IOS_DIR     := ./mobile/ios/initium
ANDROID_DIR := ./mobile/android
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

# iOS simulator used by test:ios / dev:ios. Override with `make dev:ios IOS_SIM='iPhone 17'`.
IOS_SIM ?= iPhone 17 Pro

# Android toolchain — auto-detect Android Studio's bundled JDK and the SDK
# in the standard macOS location. `which java` from a normal shell often
# resolves to a stub on macOS, so we set JAVA_HOME explicitly for every
# Gradle invocation. ANDROID_HOME is required by AGP. ADB is invoked by
# absolute path because it's rarely on $PATH outside Android Studio.
ANDROID_SDK ?= $(HOME)/Library/Android/sdk
ANDROID_STUDIO_JBR ?= /Applications/Android Studio.app/Contents/jbr/Contents/Home
ADB := $(ANDROID_SDK)/platform-tools/adb
GRADLE_ENV := JAVA_HOME="$(ANDROID_STUDIO_JBR)" ANDROID_HOME="$(ANDROID_SDK)"

.PHONY: help \
        setup keygen clobber \
        infra\:up infra\:down infra\:reset logs logs\:db logs\:mail status \
        db\:migrate db\:rollback db\:reset db\:seed db\:create db\:psql \
        gen gen\:openapi \
        test test\:backend test\:web test\:ios test\:android test\:contract test\:all \
        lint lint\:backend lint\:web lint\:ios lint\:android \
        format format\:backend format\:web format\:ios format\:android \
        dev dev\:backend dev\:web dev\:ios dev\:android \
        build\:backend build\:web build\:ios build\:android \
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
	cd $(BACKEND_DIR) && go mod download
	cd $(WEB_DIR) && npm install
	@echo "iOS: open $(IOS_DIR)/initium.xcodeproj in Xcode 26+ (simulator download handled on first run)."
	@echo "Android: open $(ANDROID_DIR) in Android Studio (Gradle syncs on first import)."
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

clobber: ## Nuclear clean — remove build artifacts, node_modules, Xcode DerivedData, Gradle caches
	rm -rf $(BACKEND_DIR)/bin
	rm -rf $(WEB_DIR)/.next $(WEB_DIR)/node_modules
	rm -rf $(IOS_DIR)/build $(IOS_DIR)/.swiftpm $(IOS_DIR)/DerivedData
	rm -rf $(ANDROID_DIR)/build $(ANDROID_DIR)/.gradle $(ANDROID_DIR)/app/build

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

gen\:openapi: ## Generate openapi.yaml from Huma + regenerate web TypeScript types
	cd $(BACKEND_DIR) && go run ./cmd/gen-openapi
	cd $(WEB_DIR) && npx openapi-typescript ../backend/api/openapi.yaml -o src/lib/api-types.ts

# ============================================================================
## group: test
# ============================================================================

test: ## Fast suite (backend + web unit tests, parallel). Native mobile runs separately.
	@$(MAKE) -j2 test\:backend test\:web

test\:backend: ## Backend Go tests with race detector
	cd $(BACKEND_DIR) && go test ./... -v -race -count=1

test\:web: ## Web Vitest suite
	cd $(WEB_DIR) && npm run test

test\:ios: ## iOS Swift Testing on simulator (requires Xcode 26+)
	cd $(IOS_DIR) && xcodebuild test \
		-project initium.xcodeproj \
		-scheme initium \
		-destination 'platform=iOS Simulator,name=$(IOS_SIM)' \
		-quiet

test\:android: _ensure-android ## Android unit tests (./gradlew test)
	cd $(ANDROID_DIR) && $(GRADLE_ENV) ./gradlew test

test\:android\:instrumented: _ensure-android ## Android Compose UI tests (requires running emulator or device)
	cd $(ANDROID_DIR) && $(GRADLE_ENV) ./gradlew connectedAndroidTest

test\:contract: ## Schemathesis contract tests (requires running backend; heavy — CI-only)
	bash scripts/schemathesis.sh

test\:all: test test\:contract ## Everything: fast suite + contract tests

# ============================================================================
## group: lint
# ============================================================================

lint: ## All linters (backend + web, parallel). Native iOS/Android linters run separately.
	@$(MAKE) -j2 lint\:backend lint\:web

lint\:backend: ## golangci-lint backend
	cd $(BACKEND_DIR) && golangci-lint run ./...

lint\:web: ## ESLint + TypeScript web
	cd $(WEB_DIR) && npm run lint

lint\:ios: ## iOS static analysis (xcodebuild with -quiet; SwiftLint/SwiftFormat not yet wired)
	cd $(IOS_DIR) && xcodebuild -project initium.xcodeproj -scheme initium -destination 'platform=iOS Simulator,name=$(IOS_SIM)' -quiet analyze

lint\:android: _ensure-android ## Android Lint + (future) ktlint/detekt
	cd $(ANDROID_DIR) && $(GRADLE_ENV) ./gradlew lint

format: ## Format all code (backend + web)
	@$(MAKE) -j2 format\:backend format\:web

format\:backend: ## gofmt backend
	cd $(BACKEND_DIR) && gofmt -w .

format\:web: ## prettier web
	cd $(WEB_DIR) && npx prettier --write "src/**/*.{ts,tsx,json,css}"

format\:ios: ## Swift format (uses Xcode's bundled swift-format)
	cd $(IOS_DIR) && xcrun swift-format -i -r initium initiumTests initiumUITests

format\:android: ## Kotlin format (relies on Android Studio's formatter; ktlint not yet wired)
	@echo "format:android is a placeholder; ktlint/detekt not yet wired. Use Android Studio's Reformat Code for now."

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

dev\:ios: _ensure-simulator ## iOS — boots simulator if needed, builds + runs via Xcode
	cd $(IOS_DIR) && xcodebuild \
		-project initium.xcodeproj \
		-scheme initium \
		-destination 'platform=iOS Simulator,name=$(IOS_SIM)' \
		-configuration Debug \
		-derivedDataPath build \
		build
	@APP_PATH="$(IOS_DIR)/build/Build/Products/Debug-iphonesimulator/initium.app"; \
	UDID=$$(xcrun simctl list devices booted 2>/dev/null | grep -oE '[A-F0-9-]{36}' | head -1); \
	if [ ! -d "$$APP_PATH" ]; then \
		echo "Build succeeded but $$APP_PATH not found." >&2; exit 1; \
	fi; \
	if [ -z "$$UDID" ]; then \
		echo "No simulator booted." >&2; exit 1; \
	fi; \
	echo "Installing $$APP_PATH on $$UDID"; \
	xcrun simctl install "$$UDID" "$$APP_PATH"; \
	BUNDLE_ID=$$(/usr/libexec/PlistBuddy -c 'Print :CFBundleIdentifier' "$$APP_PATH/Info.plist"); \
	echo "Launching $$BUNDLE_ID"; \
	xcrun simctl launch "$$UDID" "$$BUNDLE_ID"

dev\:android: _ensure-android ## Android — installs + launches the debug APK on a running emulator/device
	@if ! $(ADB) get-state >/dev/null 2>&1; then \
		echo "No Android device or emulator detected." >&2; \
		echo "Start one in Android Studio (Tools → Device Manager) and re-run." >&2; \
		exit 1; \
	fi
	cd $(ANDROID_DIR) && $(GRADLE_ENV) ./gradlew installDebug
	@APP_ID=$$(grep -E '^[[:space:]]*applicationId' $(ANDROID_DIR)/app/build.gradle.kts | head -1 | sed -E 's/.*"([^"]+)".*/\1/'); \
	echo "Launching $$APP_ID/.MainActivity"; \
	$(ADB) shell am start -n "$$APP_ID/.MainActivity"

_ensure-android:
	@if [ ! -d "$(ANDROID_STUDIO_JBR)" ]; then \
		echo "Android Studio's bundled JDK not found at $(ANDROID_STUDIO_JBR)." >&2; \
		echo "Install Android Studio from https://developer.android.com/studio" >&2; \
		echo "Or override with: make dev:android ANDROID_STUDIO_JBR=/path/to/jdk" >&2; \
		exit 1; \
	fi
	@if [ ! -d "$(ANDROID_SDK)" ]; then \
		echo "Android SDK not found at $(ANDROID_SDK)." >&2; \
		echo "Open Android Studio at least once to install the SDK," >&2; \
		echo "or override with: make dev:android ANDROID_SDK=/path/to/sdk" >&2; \
		exit 1; \
	fi
	@if [ ! -x "$(ADB)" ]; then \
		echo "adb not found at $(ADB)." >&2; \
		echo "Install Android SDK platform-tools via Android Studio's SDK Manager." >&2; \
		exit 1; \
	fi

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

build\:ios: ## iOS archive-style build against a simulator destination
	cd $(IOS_DIR) && xcodebuild \
		-project initium.xcodeproj \
		-scheme initium \
		-destination 'platform=iOS Simulator,name=$(IOS_SIM)' \
		-configuration Debug \
		build

build\:android: _ensure-android ## Android debug APK
	cd $(ANDROID_DIR) && $(GRADLE_ENV) ./gradlew assembleDebug

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

check\:parity: ## Verify every /api/ spec path has a web consumer (mobile is paused — see mobile/AGENTS.md)
	cd $(BACKEND_DIR) && go run ./cmd/check-parity

check\:skills: ## Verify exemplar paths and <!-- expect: symbol --> annotations in each SKILL.md
	@bash scripts/check-skills.sh

check\:staged: ## Fail if git has untracked or unstaged changes
	@bash scripts/check-staged.sh

preflight: ## Every gate a PR must pass: lint + test + check:parity + check:skills + check:staged
	@bash scripts/preflight.sh
