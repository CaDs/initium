---
name: initium-backend
description: Use when writing or modifying Go backend code in the Initium template — handlers, services, domain entities, persistence, middleware, migrations, config, or backend tests. Triggers on paths under `backend/internal/**`, `backend/cmd/**`, `backend/migrations/**`, or `backend/api/openapi.yaml`. Encodes the hexagonal architecture rules, error-envelope conventions, and contract-first workflow for this specific fork-and-specialize starter template.
---

# initium-backend

You are editing the Go backend of an Initium fork. This template ships auth,
sessions, and observability as scaffolding. Features are added on top without
reinventing the plumbing.

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
2. Run `make gen:openapi` to regenerate Go types (`internal/gen/api/types.gen.go`)
   and TypeScript types (`web/src/lib/api-types.ts`).
3. Implement the handler using the generated request/response types.
4. Register the route in `internal/app/router.go`. The contract test at
   `internal/app/contract_test.go` will fail if the new route is missing from
   the spec or vice versa.
5. If a new schema needs a mobile DTO, add it to `mobile/tool/dto_manifest.yaml`
   and write the hand-coded Dart DTO. `make check:openapi` verifies every
   required field is referenced.

Full workflow: `docs/OPENAPI.md`.

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

## Testing

- Table-driven tests in `*_test.go`, `testify/assert` + `require`.
- `t.Parallel()` where safe.
- 80% coverage floor. Every bug fix gets a regression test.
- Name pattern: `TestServiceName_Method_Scenario_Expected`.
- Handler tests build a test router via `app.NewRouter(app.RouterDeps{...})`
  with stubbed handlers, or construct the handler directly with mocked
  dependencies — see `adapter/handler/mobile_auth_test.go` for the pattern.

## Canonical exemplars (open these when unsure)

- Service: `backend/internal/service/auth.go` — error wrapping, context, testify mocks.
- Handler: `backend/internal/adapter/handler/auth.go` — request parse, service call, error envelope.
- Repo: `backend/internal/adapter/persistence/user_repo.go` — GORM impl with mappers.
- Migration: `backend/migrations/` — sequential numbered .sql files.
- Contract test: `backend/internal/app/contract_test.go` — route↔spec parity.

See also: `patterns/handler.md`, `patterns/service.md`, `patterns/migration.md`, `patterns/test.md`.

## Parity

See [parity.md](../_shared/parity.md). New endpoints typically imply new UI on
both web and mobile. Call out the mirror work in your PR.
