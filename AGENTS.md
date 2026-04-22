# Initium

Opinionated POC starter template. Fork and specialize per project.

Platform-specific details live in `backend/CLAUDE.md`, `web/CLAUDE.md`, and `mobile/CLAUDE.md`.
This file covers cross-cutting principles and contracts.

# CLAUDE AND AGENTS Principles

## 1. Think Before Coding — don't assume, don't hide confusion

- State assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them — don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.

## 2. Simplicity First — minimum code that solves the problem

- No features beyond what was asked. No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- Ask: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

## 3. Surgical Changes — touch only what you must

- Don't improve adjacent code, comments, or formatting.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it — don't delete it.
- Remove imports/variables/functions that YOUR changes made unused.
- Every changed line should trace directly to the user's request.

## 4. Goal-Driven Execution — define success criteria, loop until verified

- "Add validation" → write tests for invalid inputs, then make them pass
- "Fix the bug" → write a test that reproduces it, then make it pass
- "Refactor X" → ensure tests pass before and after

---

## Stack

- **Backend**: Go (chi + GORM + PostgreSQL) — see `backend/CLAUDE.md`
- **Frontend**: Next.js (App Router, TypeScript, Tailwind) — see `web/CLAUDE.md`
- **Mobile**: Flutter (Dart, Riverpod, Dio) — see `mobile/CLAUDE.md`

## Architecture

Ports & Adapters (hexagonal). Inner layers never import outer layers — violations are bugs.

| Layer | Rule |
|-------|------|
| `domain/` | Zero framework imports. Pure entities, interfaces, errors. |
| `service/` or `usecase/` | Imports domain only. Business logic. |
| `adapter/` | Handlers, persistence, DTOs. Imports domain + service. |
| `infra/` | Config, DB, external services. Outermost ring. |

## Build & Run

```bash
make setup        # First-time: infra, deps, .env, migrations, JWT keys
make dev          # Backend (8000) + web (3000)
make test:backend # Go tests with race detector
make test:web     # Vitest
make test:mobile  # Flutter tests
make gen:mobile   # Required after editing lib/l10n/*.arb (flutter gen-l10n)
make check:openapi # Verify mobile DTOs stay in sync with OpenAPI spec
make preflight    # Every gate a PR must pass (lint + test + check:openapi)
```

## Auth Model

- Backend owns session state (single source of truth)
- Short-lived access tokens (15min) + refresh tokens (7d) in sessions table
- Google OAuth (web: server-side flow, mobile: ID token via `/auth/mobile/google`)
- Magic links (web: redirect flow, mobile: `/auth/mobile/verify` returns JSON)
- `DEV_BYPASS_AUTH=true` (dev only): skips auth, injects `dev@initium.local`

## API Contract

`backend/api/openapi.yaml` is the canonical spec. After editing it, run `make gen:openapi`:

- `backend/internal/gen/api/types.gen.go` — Go types via `oapi-codegen` (pinned in `backend/go.mod` as a tool dependency; invoked via `go tool oapi-codegen`). **Every handler uses these generated types for request + response**; see `backend/internal/adapter/handler/user.go` for the canonical shape. Domain entities in `internal/domain/` remain hand-written.
- `web/src/lib/api-types.ts` — TypeScript types via `openapi-typescript`. Existing Zod schemas in `lib/schemas.ts` remain the runtime validator; cross-check against generated types during review.
- Mobile DTOs in `mobile/lib/data/remote/dto/` stay **hand-written** — the template does NOT use `build_runner` / `json_serializable` / `freezed`. `make check:openapi` verifies every required schema field is referenced in the corresponding Dart class (manifest at `mobile/tool/dto_manifest.yaml`). Also run `make gen:mobile` (= `flutter gen-l10n`) after editing `lib/l10n/*.arb`.

Every OpenAPI schema used on the wire MUST have a `required:` array — otherwise codegen produces optional fields everywhere and consumers have to guard fields that the backend always returns. **List endpoints use envelope schemas** (`{"resource_name": [...]}`), never bare arrays.

## Observability

The template ships opinionated-but-light defaults and leaves vendor choices to the fork author.

**What's wired out of the box:**
- `GET /healthz` — liveness (no deps), returns `{"status":"ok"}`.
- `GET /readyz` — readiness, `db.Ping()` with 2s timeout, returns 503 on failure.
- `GET /metrics` — Prometheus endpoint via `prometheus/client_golang` with default Go runtime + process collectors. Point a Prometheus scraper at it; register your own `prometheus.Counter` / `Histogram` against the default registry to track app metrics.
- Structured access-log middleware (slog JSON: method, path, status, duration_ms, request_id, remote_ip).

**What's NOT wired (env placeholders + pointers only):**
- **Sentry** — DSN env vars exist (`SENTRY_DSN` backend+mobile, `NEXT_PUBLIC_SENTRY_DSN` web) but no init code. To enable:
  - Backend: `go get github.com/getsentry/sentry-go`, call `sentry.Init(sentry.ClientOptions{Dsn: cfg.SentryDSN, Environment: cfg.AppEnv})` in `main.go` before router wiring; add `sentryhttp.New(...).Handle` as a middleware before Recoverer.
  - Web: `npm i @sentry/nextjs`, follow `npx @sentry/wizard@latest -i nextjs`.
  - Mobile: `flutter pub add sentry_flutter`, wrap `SentryFlutter.init((opts) { opts.dsn = ...; }, appRunner: () => runApp(...))` in `main.dart`.
- **OpenTelemetry** — `OTEL_EXPORTER_OTLP_ENDPOINT` env placeholder only. To enable tracing on chi, install `go.opentelemetry.io/contrib/instrumentation/github.com/go-chi/chi/otelchi` and wire at the top of global middleware. Keep it off by default — OTEL config is tedious and vendor-specific.

Pick Sentry OR OTEL, not both, unless you genuinely need span export + error grouping.

## Conventions

- Conventional Commits: `feat:`, `fix:`, `test:`, `refactor:`, `docs:`, `chore:`
- No secrets in version control. Use `.env.example` templates.
- Parameterized queries only. No string interpolation for SQL.
- i18n: all user-facing strings localized in en/es/ja (details in platform CLAUDE.md)
- Theme: three modes — light/dark/system (details in platform CLAUDE.md)
- Accessibility: required baseline (details in platform CLAUDE.md)
- Run linters and tests after changes automatically. Never push or open PRs without approval.
