---
name: initium-backend
description: Use when writing or modifying Go backend code in the Initium template — handlers, services, domain entities, persistence, middleware, migrations, config, or backend tests. Triggers on paths under `backend/internal/**`, `backend/cmd/**`, `backend/migrations/**`, or `backend/api/openapi.yaml`. Encodes the hexagonal architecture rules, error-envelope conventions, and Huma code-first workflow for this specific fork-and-specialize starter template.
---

# initium-backend

You are editing the Go backend of an Initium fork. This template ships auth,
sessions, observability, and a Huma-driven OpenAPI contract as scaffolding.
Features are added on top without reinventing the plumbing.

## Gates that will fail your PR

Prose rules drift; these gates don't. Run `make preflight` before
committing. It fails if any of the following is true:

- A domain error sentinel isn't mapped in `respond.go`
  (`error_envelope_test.go`).
- `make gen:openapi` produces a diff against `backend/api/openapi.yaml`
  or `web/src/lib/api-types.ts` — i.e. someone added a Huma operation
  but forgot to regenerate downstream artifacts. The drift step in
  `scripts/preflight.sh` runs gen and diffs the tracked files.
- A `/api/*` spec path has no consumer in web or mobile code
  (`make check:parity`).
- Backend coverage drops below 35% (`make test:backend:coverage`).
  Phased ramp toward 80% in follow-up PRs as coverage grows.
- An exemplar path cited in this skill no longer contains its
  `<!-- expect: symbol -->` annotation (`make check:skills`).
- `git status --porcelain` is non-empty after the run
  (`make check:staged`).

If you add a convention to this skill, ask whether a gate can enforce
it. If yes, add the gate. If no, the convention will drift.

## Architecture (strict — violations are bugs)

```
internal/domain/          Pure entities, interfaces (port.go), errors.
                          NO imports from service/, adapter/, infra/, or any
                          third-party package except stdlib.
internal/service/         Business logic implementing domain interfaces.
                          Imports domain only.
internal/adapter/handler/ Huma handlers + chi-native handlers.
                          Imports domain + service + infra types as needed.
internal/adapter/middleware/   auth, request ID, access log (chi style).
internal/adapter/persistence/  GORM models + mappers + repo implementations.
                               Hand-written GORM structs in models.go with
                               toDomain()/fromDomain() helpers.
internal/infra/           Config, DB setup, JWT, OAuth, email. Outermost.
internal/app/             Router builder + Huma API constructor. Keeps
                          cmd/server thin so cmd/gen-openapi can build the
                          same API graph with stub deps.
cmd/server/               main.go composition root.
cmd/gen-openapi/          Builds the Huma API in-process and writes
                          backend/api/openapi.yaml. Run via `make gen:openapi`.
cmd/check-parity/         Verifies every /api/ spec path has a client consumer.
```

The dependency rule: **inner layers never import outer layers**. Domain
imports nothing project-local. Service imports domain. Adapter imports
domain + service. Infra is the outermost ring.

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
| Huma operation ID | `kebab-case` | `"request-magic-link"` |
| Huma input/output types | `lowercase` (unexported) struct types per handler | `magicLinkInput`, `messageOutput` |
| Wire DTO type | `PascalCase` in `handler/types.go` | `User`, `TokenPair` |

## The code-first workflow (Huma)

The Go code is the source of truth for the OpenAPI spec. `backend/api/openapi.yaml`
is a generated artifact — never hand-edit. When adding or changing any endpoint:

1. Define the request + response shapes as Go structs. Wire shapes
   (visible to clients) live in `internal/adapter/handler/types.go` —
   add new ones there with Huma struct tags (`required:"true"`,
   `format:"email"`, `minLength:"1"`, `enum:"a,b"`, `doc:"..."`).
   Per-handler input/output structs (often `xInput` / `xOutput`)
   live in the handler file itself.
2. Write the handler as `func (ctx context.Context, in *Input) (*Output, error)`.
   On error, `return nil, MapDomainErr(ctx, err)` — never write to
   the response writer manually. The MapDomainErr helper produces an
   `APIError` (`code`, `message`, `request_id`) that Huma serializes
   with the correct status.
3. Add a `RegisterX` function on the handler that calls `huma.Register(api, op, h.X)`.
   Operation metadata (`OperationID`, `Method`, `Path`, `Summary`,
   `Tags`, `Security`, `Middlewares`) goes in the `huma.Operation`
   struct alongside the registration. Auth gates use the
   `huma.Middlewares` slot — see `handler/auth_huma.go` for the
   `HumaAuthMiddleware` and `HumaRequireRole` helpers.
4. Wire the new `RegisterX` call from `internal/app/router.go` (so
   `cmd/server` runs it) AND from `cmd/gen-openapi/main.go` (so the
   on-disk spec includes it).
5. Run `make gen:openapi` to regenerate `backend/api/openapi.yaml`
   and `web/src/lib/api-types.ts`. Commit the regenerated files.

**No more "edit spec first" cadence.** Validation tags on the input
struct generate the right schema constraints automatically; manual
`if req.Email == ""` checks are gone.

**Cross-stack completeness**: backend endpoints almost always imply
matching web UI. If you add `/api/notes`, a web Zod schema should
land with it. `make check:parity` fails if a spec path has no
consumer in `web/src` — treat that as a real signal, not lint
noise. Mobile parity is paused while the native apps catch up; see
`.claude/skills/_shared/parity.md`.

## Routes that stay chi-native

Not every route fits the typed-API model. The following stay as
`http.HandlerFunc` registered directly on the chi router (their
spec entries are absent from `openapi.yaml` by design):

| Route | Why chi-native |
|-------|---------------|
| `GET /api/auth/google` | 307 redirect + `Set-Cookie`; browser flow, not REST |
| `GET /api/auth/google/callback` | Same — redirect after Google consent |
| `GET /api/auth/verify` | Same — redirect after magic-link tap |
| `GET /healthz`, `/readyz` | LB probes; tiny payloads, no benefit from typed I/O |
| `GET /_debug/routes` | Dev-only chi route introspection |
| `Handle /metrics` | Prometheus streaming text — Huma can't wrap a streaming handler cleanly |

Rule of thumb: *"Is this a REST endpoint a typed client should
consume?"* → yes → Huma. → no → chi-native.

## Error handling

Huma handlers `return nil, MapDomainErr(ctx, err)` for any domain
error. `MapDomainErr` (in `internal/adapter/handler/errs.go`) looks
up the wire code + HTTP status via the existing `mapError` switch
and returns an `*APIError` (`code`, `message`, `request_id`) that
implements `huma.StatusError`. Huma serializes it as the response body.

Errors Huma synthesizes itself (validation failures, malformed body)
also use the `APIError` envelope thanks to the `huma.NewError`
override installed by `InstallErrorEnvelope()`.

Adding a new domain error:

1. Sentinel in `internal/domain/errors.go` — `var ErrFooBar = errors.New("foo bar")`.
2. Mapping in `respond.go` `mapError` switch — return `("FOO_BAR", http.StatusConflict)`.
3. Test `TestError_EnvelopeShape` (in `error_envelope_test.go`) will
   fail if a new sentinel is not mapped.

The chi-native redirect + ops handlers use the small `writeChiError`
helper in `auth.go` for consistency — same envelope shape, no Huma
machinery needed.

## Streaming + SSE (future)

Huma's `huma/v2/sse` package supports typed Server-Sent Events
out of the box. Use `sse.Register(api, op, eventTypeMap, handler)`
when adding chatbot / LLM-token / live-notification endpoints. Raw
byte streaming via `huma.StreamResponse` works too. See
`https://huma.rocks/features/response-streaming/` for the patterns.

## Context + logging + concurrency

- `context.Context` is always the first parameter on service / repo
  / Huma handler methods.
- Structured logging via `slog` only — no `fmt.Println`, no `log.Printf`.
- Error wrapping: `fmt.Errorf("loading user: %w", err)`. Never return
  naked `err` when there's meaningful context to add.
- Goroutines: use `errgroup` or explicit cancellation. Never
  fire-and-forget.
- Graceful shutdown lives in `infra/server.go` via `signal.NotifyContext` —
  new background workers (cron, worker pools) must register their
  `Close()` callback with `infra.ServeHTTP` so SIGINT drains cleanly.

## Auth endpoints reference

| Method | Path | Auth | Style | Purpose |
|--------|------|------|-------|---------|
| GET | `/api/auth/google` | No | chi | Redirect to Google consent |
| GET | `/api/auth/google/callback` | No | chi | Exchange code, set cookies |
| POST | `/api/auth/magic-link` | No | Huma | Send magic link email (rate limited) |
| GET | `/api/auth/verify` | No | chi | Verify magic link token (browser, redirects) |
| POST | `/api/auth/mobile/google` | No | Huma | Verify mobile ID token |
| POST | `/api/auth/mobile/verify` | No | Huma | Verify mobile magic-link token (JSON) |
| POST | `/api/auth/refresh` | refresh cookie OR body | Huma | Issue new token pair |
| POST | `/api/auth/logout` | Yes | Huma | Revoke current session |
| POST | `/api/auth/logout-all` | Yes | Huma | Revoke all sessions for user |
| GET | `/api/me` | Yes | Huma | Current user profile |
| PATCH | `/api/me` | Yes | Huma | Update profile |
| GET | `/api/admin/ping` | Yes + admin | Huma | Admin-role liveness check |
| GET | `/api/landing` | No | Huma | Public landing payload |
| GET | `/healthz` | No | chi | Process liveness |
| GET | `/readyz` | No | chi | DB-reachable readiness |
| Handle | `/metrics` | No | chi | Prometheus scrape |
| GET | `/_debug/routes` | No | chi | Dev-only route table (omitted when `APP_ENV=production`) |

## Testing

- Table-driven tests in `*_test.go`, `testify/assert` + `require`.
- `t.Parallel()` where safe.
- Phased coverage gate at 35% lines (current: 44.7%); ramping to
  80% as coverage grows in follow-up PRs. Run with
  `make test:backend:coverage`. Every bug fix gets a regression test.
- Name pattern: `TestServiceName_Method_Scenario_Expected`.
- **Mocks live in `internal/testutil/`** as canonical, function-injection
  fakes (one `Fn` field per interface method). Compile-time
  conformance assertions at the bottom of `mocks.go` mean adding a
  method to a domain interface fails the build there first. Don't
  hand-roll a mock locally in a `*_test.go` file — extend the
  testutil one. See `mocks.go` <!-- expect: MockAuthService --> +
  `fixtures.go` <!-- expect: RegularUser --> +
  `decode.go` <!-- expect: MustDecodeJSON -->.
- Huma handler tests use `humatest.New(t)` to build a typed test
  API; register the handler via its `RegisterX` method, then call
  `api.Get(...)`, `api.Post(...)`, etc. Returns an `*httptest.ResponseRecorder`
  via `resp.Code` / `resp.Body`. See
  `adapter/handler/mobile_auth_test.go` for the pattern.
- chi-bridged middleware (httprate, etc.) tests CANNOT use humatest
  because it's built on humaflow, not humachi — and `HumaFromHTTP`
  calls `humachi.Unwrap`. Use a real chi router + `humachi.New(r, cfg)`
  with `httptest.NewRecorder()` instead. See
  `adapter/handler/middleware_bridge_test.go` <!-- expect: humachi.New --> for the harness.

## Canonical exemplars (open these when unsure)

- Service: `backend/internal/service/auth.go` <!-- expect: AuthService --> — error wrapping, context, testify mocks.
- Huma handler (input + output struct, validation tags): `backend/internal/adapter/handler/mobile_auth.go` <!-- expect: RegisterMobileAuth -->
- Huma handler (auth + role middleware via huma.Operation.Middlewares): `backend/internal/adapter/handler/admin.go` <!-- expect: RegisterAdmin -->
- Huma handler (cookie input + Set-Cookie output): `backend/internal/adapter/handler/auth.go` <!-- expect: tokenPairOutput -->
- chi-native handler (browser redirect + Set-Cookie): `backend/internal/adapter/handler/auth.go` <!-- expect: GoogleRedirect -->
- Huma auth middleware bridge: `backend/internal/adapter/handler/auth_huma.go` <!-- expect: HumaAuthMiddleware -->
- chi → Huma middleware bridge (rate limit): `backend/internal/adapter/handler/middleware_bridge.go` <!-- expect: HumaFromHTTP -->
- Error envelope override: `backend/internal/adapter/handler/errs.go` <!-- expect: InstallErrorEnvelope -->
- Error envelope test: `backend/internal/adapter/handler/error_envelope_test.go` <!-- expect: TestError_EnvelopeShape -->
- Handler test (humatest): `backend/internal/adapter/handler/mobile_auth_test.go` <!-- expect: humatest.New -->
- Repo: `backend/internal/adapter/persistence/user_repo.go` <!-- expect: GormUserRepo --> — GORM impl with mappers.
- Migration: `backend/migrations/` — sequential numbered .sql files.
- Spec generator: `backend/cmd/gen-openapi/main.go` <!-- expect: api.OpenAPI() -->

See also: `patterns/handler.md`, `patterns/service.md`, `patterns/migration.md`, `patterns/test.md`, `patterns/feature-crud.md`.

## Parity

See [parity.md](../_shared/parity.md). New endpoints typically imply
new UI on both web and mobile. Call out the mirror work in your PR.
