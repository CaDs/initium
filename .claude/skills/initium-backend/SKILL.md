---
name: initium-backend
description: Use when writing or modifying Go backend code in the Initium template — handlers, services, domain entities, persistence, middleware, migrations, config, or backend tests. Triggers on paths under `backend/internal/**`, `backend/cmd/**`, `backend/migrations/**`, or `backend/api/openapi.yaml`. Encodes the hexagonal architecture rules, error-envelope conventions, and contract-first workflow for this specific fork-and-specialize starter template.
---

# initium-backend

You are editing the Go backend of an Initium fork. This template ships auth,
sessions, and observability as scaffolding. Features are added on top without
reinventing the plumbing.

## Gates that will fail your PR

Prose rules drift; these gates don't. Run `make preflight` before
committing. It fails if any of the following is true:

- A chi route exists without a matching OpenAPI path, or vice versa
  (`backend/internal/app/contract_test.go`).
- A domain error sentinel isn't mapped in `respond.go`
  (`error_envelope_test.go`).
- A `/api/*` spec path has no consumer in web or mobile code
  (`make check:parity`).
- A required schema field is missing from its hand-written mobile DTO
  (`make check:openapi`).
- An exemplar path cited in this skill no longer contains its
  `<!-- expect: symbol -->` annotation (`make check:skills`).
- `git status --porcelain` is non-empty after the run (`make check:staged`).

If you add a convention to this skill, ask whether a gate can enforce it.
If yes, add the gate. If no, the convention will drift.

## Architecture (strict — violations are bugs)

```
internal/domain/          Pure entities, interfaces (port.go), errors.
                          NO imports from service/, adapter/, infra/, or any
                          third-party package except stdlib.
internal/service/         Business logic implementing domain interfaces.
                          Imports domain only.
internal/adapter/handler/ chi HTTP handlers (thin controllers).
                          Imports domain + service + infra types as needed.
internal/adapter/middleware/  auth, CORS, rate limit, request ID.
internal/adapter/persistence/ GORM models + mappers + repo implementations.
                          Hand-written GORM structs in models.go with
                          toDomain()/fromDomain() helpers.
internal/infra/           Config, DB setup, JWT, OAuth, email. Outermost.
internal/app/             Router builder + contract tests. Keeps cmd/server
                          thin so tests can construct a router with stubs.
internal/gen/api/         Generated from backend/api/openapi.yaml — never hand-edit.
cmd/server/               main.go composition root.
cmd/check-dto-drift/      Utility verifying mobile DTOs match the spec.
```

The dependency rule: **inner layers never import outer layers**. Domain imports
nothing project-local. Service imports domain. Adapter imports domain + service.
Infra is the outermost ring.

## Naming

| Kind | Convention | Example |
|------|------------|---------|
| Domain error sentinel | `Err` prefix, `PascalCase` | `domain.ErrUserNotFound` |
| Error code (wire) | `SNAKE_UPPER` | `"USER_NOT_FOUND"` |
| Repo interface | `<Entity>Repository` in domain | `domain.UserRepository` |
| Repo impl | `Gorm<Entity>Repo` in persistence | `persistence.GormUserRepo` |
| Service interface | `<Name>Service` in domain | `domain.AuthService` |
| Service impl | `<Name>Service` struct in service | `service.AuthService` |
| Handler | `<Feature>Handler` in handler | `handler.AuthHandler` |
| Constructor | `New<Type>` | `NewAuthHandler` |

## The contract-first workflow (non-negotiable)

`backend/api/openapi.yaml` is the single source of truth. When adding or
changing any endpoint:

1. Edit `backend/api/openapi.yaml` first — path, request schema, response
   schema, every error response referenced back to `#/components/schemas/ErrorResponse`.
   **Every schema used on the wire MUST declare a `required:` array listing
   every field that is always present.** Missing it produces nullable-everywhere
   types downstream, forcing every consumer to guard fields the backend always
   returns.
2. **List endpoints use envelope schemas**, not bare arrays. A list response is
   `{"resource_name": [...]}` (e.g. `RouteList { routes: [...] }`). Define both
   the item schema and the envelope schema, and mark the array `required`.
   Bare-array responses break every client's Zod mapper.
3. Run `make gen:openapi` to regenerate Go types (`internal/gen/api/types.gen.go`)
   and TypeScript types (`web/src/lib/api-types.ts`).
4. Implement the handler using the generated request/response types from
   `internal/gen/api`. See `patterns/handler.md`. Request decoding uses
   `DisallowUnknownFields()` so unknown keys 400 instead of silently ignoring.
5. Register the route in `internal/app/router.go`. The contract test at
   `internal/app/contract_test.go` will fail if the new route is missing from
   the spec or vice versa. Routes with chi URL params like `/api/notes/{id}`
   must appear in `openapi.yaml` using `{id}` syntax (chi's `{id}` maps directly).
   The walker strips trailing `/*`; only add to `excludedPaths` for non-API
   operational routes (e.g. `/metrics`).
6. Native mobile (iOS / Android) codegen is **deferred** — see `docs/OPENAPI.md`.
   Until it's wired, mobile-bound endpoints can be added to the `excluded`
   list in `backend/cmd/check-parity/main.go` so the parity gate doesn't
   fail on them.

Full workflow: `docs/OPENAPI.md`.

**Cross-stack completeness**: backend endpoints almost always imply
matching web UI. If you add `/api/notes`, a web Zod schema should land
with it. `make check:parity` fails if a spec path has no consumer in
`web/src` — treat that as a real signal, not lint noise. Mobile parity
is paused while the native apps catch up; see
`.claude/skills/_shared/parity.md`.

## Error handling

All handlers route errors through `handler.Error(w, r, err)` in
`internal/adapter/handler/respond.go`. It wraps with the standard envelope:

```json
{ "code": "SNAKE_UPPER", "message": "...", "request_id": "..." }
```

New domain errors go through three places:

1. Sentinel in `internal/domain/errors.go` — `var ErrFooBar = errors.New("foo bar")`
2. Mapping in `respond.go` `mapError` switch — return `("FOO_BAR", http.StatusConflict)`
3. Response documented in `openapi.yaml` for every endpoint that can return it

Unmapped errors become `INTERNAL_ERROR` + 500 with a generic message (never leak
internals). Test `TestError_EnvelopeShape` in
`internal/adapter/handler/error_envelope_test.go` will fail if a new sentinel
is not mapped.

## Context + logging + concurrency

- `context.Context` is always the first parameter on service/repo methods.
- Structured logging via `slog` only — no `fmt.Println`, no `log.Printf`.
- Error wrapping: `fmt.Errorf("loading user: %w", err)`. Never return naked `err`
  when there's meaningful context to add.
- Goroutines: use `errgroup` or explicit cancellation. Never fire-and-forget.
- Graceful shutdown lives in `infra/server.go` via `signal.NotifyContext` —
  new background workers (cron, worker pools) must register their `Close()`
  callback with `infra.ServeHTTP` so SIGINT drains cleanly.

## Auth endpoints reference

All routes mounted under `/api`. Authentication column reflects runtime gate,
not spec visibility.

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| GET | `/api/auth/google` | No | Redirect to Google consent |
| GET | `/api/auth/google/callback` | No | Exchange code, set cookies |
| POST | `/api/auth/magic-link` | No | Send magic link email (rate limited) |
| GET | `/api/auth/verify` | No | Verify magic link token (browser) |
| POST | `/api/auth/mobile/google` | No | Verify mobile ID token |
| POST | `/api/auth/mobile/verify` | No | Verify mobile magic-link token (JSON) |
| POST | `/api/auth/refresh` | refresh cookie or body | Issue new token pair |
| POST | `/api/auth/logout` | Yes | Revoke current session |
| POST | `/api/auth/logout-all` | Yes | Revoke all sessions for user |
| GET | `/api/me` | Yes | Current user profile |
| PATCH | `/api/me` | Yes | Update profile |
| GET | `/api/admin/ping` | Yes + admin role | Admin-role liveness check |
| GET | `/healthz` | No | Process liveness |
| GET | `/readyz` | No | DB-reachable readiness |
| GET | `/metrics` | No | Prometheus scrape |
| GET | `/_debug/routes` | No | Dev-only route table (omitted when `APP_ENV=production`) |

## Testing

- Table-driven tests in `*_test.go`, `testify/assert` + `require`.
- `t.Parallel()` where safe.
- 80% coverage floor. Every bug fix gets a regression test.
- Name pattern: `TestServiceName_Method_Scenario_Expected`.
- Handler tests build a test router via `app.NewRouter(app.RouterDeps{...})`
  with stubbed handlers, or construct the handler directly with mocked
  dependencies. For unauthenticated endpoints see
  `adapter/handler/mobile_auth_test.go`; for authenticated endpoints see the
  `withUser()` helper in `patterns/test.md`.

## Canonical exemplars (open these when unsure)

- Service: `backend/internal/service/auth.go` <!-- expect: AuthService --> — error wrapping, context, testify mocks.
- Handler (unauth, request parse with generated type): `backend/internal/adapter/handler/mobile_auth.go` <!-- expect: api.MobileVerifyRequest -->
- Handler (auth'd, user_id from context): `backend/internal/adapter/handler/user.go` <!-- expect: middleware.GetUserID -->
- Handler (response type via `api.User` conversion): `backend/internal/adapter/handler/user.go` <!-- expect: writeUser -->
- Repo: `backend/internal/adapter/persistence/user_repo.go` <!-- expect: GormUserRepo --> — GORM impl with mappers.
- Migration: `backend/migrations/` — sequential numbered .sql files.
- Contract test: `backend/internal/app/contract_test.go` <!-- expect: TestRouter_MatchesOpenAPISpec --> — route↔spec parity + `/*` stripping.
- Error envelope test: `backend/internal/adapter/handler/error_envelope_test.go` <!-- expect: TestError_EnvelopeShape -->

See also: `patterns/handler.md`, `patterns/service.md`, `patterns/migration.md`, `patterns/test.md`, `patterns/feature-crud.md`.

## Parity

See [parity.md](../_shared/parity.md). New endpoints typically imply new UI on
both web and mobile. Call out the mirror work in your PR.
