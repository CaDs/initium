# OpenAPI Contract Workflow

The Go handlers are the source of truth for the API contract. Huma
(https://huma.rocks/) generates `backend/api/openapi.yaml` from the
typed handler signatures. Downstream tools (web TypeScript codegen,
mobile codegen once it lands) consume the generated spec — never
hand-edited.

## Generators

| Client            | Approach                       | Output                                  | Source of truth for... |
|-------------------|--------------------------------|-----------------------------------------|------------------------|
| Backend           | Code-first via Huma            | `backend/api/openapi.yaml` (generated)  | API contract — single source of truth |
| Web               | `openapi-typescript`           | `web/src/lib/api-types.ts`              | TypeScript types (Zod still runtime-validates) |
| Mobile (Expo)     | hand-written                   | `mobile/src/api/models.ts`              | TypeScript DTOs; codegen via `openapi-typescript` is deferred until a feature proves it worth wiring |

Backend domain entities in `backend/internal/domain/` remain
**hand-written**. Handler-level wire types live in
`backend/internal/adapter/handler/types.go`. Per-handler input/output
structs live in the handler files themselves with Huma struct tags.

### Why mobile codegen is deferred

The Expo app at `mobile/` consumes the spec via hand-written DTOs in
`mobile/src/api/models.ts`. Wiring `openapi-typescript` would add a
codegen step + generated-code diffs in every API PR; for a starter
template that's net-negative until there's a real surface area to
cover. Pair the first networked feature where parity drift becomes
painful with the `openapi-typescript --output mobile/src/api/spec.ts`
plumbing.

## Workflow

```bash
# 1. Edit the handler types + huma.Operation in Go.
$EDITOR backend/internal/adapter/handler/foo.go

# 2. Regenerate the on-disk spec + web TypeScript types.
make gen:openapi

# 3. Verify everything is in sync (also runs in `make preflight`).
make check:parity

# 4. Commit the regenerated artifacts alongside your handler.
git add backend/api/openapi.yaml \
        web/src/lib/api-types.ts \
        backend/internal/adapter/handler/foo.go
git commit -m "feat(api): <your change>"
```

If you forget step 2, the gen-drift step in `scripts/preflight.sh`
fails the PR with a "run `make gen:openapi`" hint.

## Rules

- **Required fields are explicit.** Use Huma struct tags
  (`required:"true"`) on fields that the response always returns.
  Optional fields use a pointer type or `omitempty`. Missing
  declarations generate optional-everywhere downstream types,
  forcing every consumer to guard fields the backend always returns.
- **Every error response uses the shared envelope.** The `APIError`
  type in `handler/errs.go` matches the existing `code` + `message` +
  `request_id` shape. Huma's `huma.NewError` is overridden in
  `InstallErrorEnvelope()` so synthesized errors (validation failures,
  malformed body) use the same envelope.
- **List endpoints use envelope schemas**, not bare arrays. A list
  response is `{"resource_name": [...]}`. Bare-array responses break
  every client's Zod mapper.
- **Non-JSON responses stay chi-native.** OAuth 307 redirects + magic
  link verify are not in `openapi.yaml`. Their handlers live in
  `handler/auth.go` as `http.HandlerFunc`s and register directly on
  the chi router (see `internal/app/router.go`).

## Error code conventions

Error codes are SNAKE_UPPER. The canonical list lives in
`backend/internal/adapter/handler/respond.go` (`mapError`). When
adding a new domain error:

1. Add the sentinel to `backend/internal/domain/errors.go`.
2. Map it in `respond.go`'s `mapError` switch.
3. Huma handlers `return nil, MapDomainErr(ctx, err)` — no per-endpoint
   wiring needed.

## Live spec + docs endpoints

The running backend serves Huma's auto-generated documentation:

- `http://localhost:8000/docs` — rendered API docs UI.
- `http://localhost:8000/openapi.yaml` / `.json` — runtime spec
  (the same content as the on-disk `backend/api/openapi.yaml`).
- `http://localhost:8000/openapi-3.0.{yaml,json}` — OpenAPI 3.0
  variants for tools that don't support 3.1.

`make docs` prints these URLs as a reminder; no Docker swagger-ui
required.

## Route discovery

In development, the running backend exposes `GET /_debug/routes`
(mounted only when `APP_ENV != "production"`). This endpoint returns
the live chi route table; `make routes` curls and pretty-prints it.

## Adding a new endpoint (end-to-end)

1. Define request + response Go types. Wire-shape DTOs land in
   `backend/internal/adapter/handler/types.go`. Per-handler I/O
   structs (often unexported `xInput` / `xOutput`) live in the
   handler file. Use Huma struct tags for validation
   (`required:"true"`, `format:"email"`, `minLength:"1"`,
   `enum:"a,b"`, `doc:"..."`).
2. Write the handler:
   `func (h *FooHandler) bar(ctx context.Context, in *barInput) (*barOutput, error)`.
   On error, `return nil, MapDomainErr(ctx, err)`.
3. Add a `RegisterFoo(api huma.API, ...)` method that calls
   `huma.Register(api, op, h.bar)` with full operation metadata
   (`OperationID`, `Method`, `Path`, `Summary`, `Tags`, `Security`,
   `Middlewares`).
4. Wire `RegisterFoo` from `internal/app/router.go` AND from
   `cmd/gen-openapi/main.go` so the operation appears in the
   generated spec.
5. `make gen:openapi` to update `backend/api/openapi.yaml` +
   `web/src/lib/api-types.ts`.
6. Add a web (Zod) consumer or a mobile consumer (mirror the DTO in
   `mobile/src/api/models.ts` and add the endpoint helper in
   `mobile/src/api/endpoints.ts`); `make check:parity` fails otherwise.
7. Commit the handler + regenerated artifacts together.
