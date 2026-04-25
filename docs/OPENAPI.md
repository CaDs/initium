# OpenAPI Contract Workflow

`backend/api/openapi.yaml` is the single source of truth for every client contract. Edit the spec, regenerate, commit the generated files — never hand-edit generated code.

## Generators

| Client            | Approach                 | Output / Check                            | Source of truth for... |
|-------------------|--------------------------|-------------------------------------------|------------------------|
| Backend           | `oapi-codegen`           | `backend/internal/gen/api/types.gen.go`   | wire DTOs (requests, responses) |
| Web               | `openapi-typescript`     | `web/src/lib/api-types.ts`                | TypeScript types (Zod still runtime-validates) |
| Mobile — iOS      | **deferred**             | will use `swift-openapi-generator`        | will produce typed Swift clients when wired |
| Mobile — Android  | **deferred**             | will use `openapi-generator` (kotlin target) | will produce Retrofit + kotlinx.serialization clients when wired |

Backend domain entities in `backend/internal/domain/` remain
**hand-written**. Generated types cover the wire contract only; domain
stays authoritative.

### Why native mobile codegen is deferred

The Flutter app that used to live under `mobile/` was removed on branch
`feat/dropping_flutter`. The replacement — two native apps (SwiftUI +
Jetpack Compose) — starts as a 3-tab UI shell with no backend calls.
Codegen wiring is valuable only once the apps actually *talk* to the
backend; pre-scaffolding it now would add:

- Build-time complexity (SPM `plugin:` or Gradle `openapi-generator`
  task running on every build).
- Generated-code diffs landing in PRs that don't touch API behavior.
- Apple / Kotlin idiom decisions frozen before there's a need.

The plan: pair the first networked feature with the codegen plumbing.
Either team can go first — they don't need to be simultaneous.

Until that lands, the `check:parity` gate scans only `web/src/**` and
`/api/auth/mobile/*` paths are temporarily excluded from the orphan
check (see `backend/cmd/check-parity/main.go`).

## Workflow

```bash
# 1. Edit the spec
$EDITOR backend/api/openapi.yaml

# 2. Regenerate web + backend types
make gen

# 3. Verify everything is in sync
make check:parity

# 4. Commit the spec AND the regenerated files
git add backend/api/openapi.yaml backend/internal/gen/api/ web/src/lib/api-types.ts
git commit -m "feat(api): <your change>"
```

If you forget step 3, CI will fail the PR.

## Rules

- Every schema used on the wire must declare a `required:` array. Missing `required` generates optional fields, which forces every consumer to guard fields the backend always returns.
- Every 4xx and 5xx response must reference `#/components/schemas/ErrorResponse` (the shared envelope: `code` + `message` + `request_id`).
- Non-JSON responses (OAuth 307 redirects, magic-link browser verify) must be documented with a description; do not invent a JSON body.
- New endpoints must land in the spec **before** the handler. The contract test (`backend/internal/app/contract_test.go`) asserts every chi route has a spec path and vice versa.

## Error code conventions

Error codes are SNAKE_UPPER. The canonical list lives in `backend/internal/adapter/handler/respond.go` (`mapError`). When adding a new domain error:

1. Add the sentinel to `backend/internal/domain/errors.go`.
2. Map it in `respond.go`'s `mapError` switch.
3. Document the HTTP status + code in the endpoint's OpenAPI response.

## Route discovery

In development, the running backend exposes `GET /_debug/routes` (mounted only when `APP_ENV != "production"`). This endpoint is documented in the spec and returns the live chi route table. `make routes` curls and pretty-prints it.

## Adding a new endpoint (end-to-end)

1. Edit `backend/api/openapi.yaml`: add path, request schema, response schema, error responses.
2. `make gen:openapi` — regenerates Go + TypeScript types.
3. Implement the handler in `backend/internal/adapter/handler/` using the generated request/response types from `internal/gen/api`.
4. Wire the route in `backend/cmd/server/main.go`.
5. Use the generated client types from the web (`web/src/lib/api-types.ts`) side; map to hand-written domain entities via per-aggregate mappers.
6. If the endpoint is mobile-bound, add a parity TODO: once native mobile codegen lands, the new iOS / Android clients should consume this path. Until then, add it to the `excluded` list in `backend/cmd/check-parity/main.go` or the parity gate will fail.
7. Contract test automatically covers route↔spec parity.
