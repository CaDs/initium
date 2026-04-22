# OpenAPI Contract Workflow

`backend/api/openapi.yaml` is the single source of truth for every client contract. Edit the spec, regenerate, commit the generated files — never hand-edit generated code.

## Generators

| Client  | Approach            | Output / Check                            | Source of truth for... |
|---------|---------------------|-------------------------------------------|------------------------|
| Backend | `oapi-codegen`      | `backend/internal/gen/api/types.gen.go`   | wire DTOs (requests, responses) |
| Web     | `openapi-typescript`| `web/src/lib/api-types.ts`                | TypeScript types (Zod still runtime-validates) |
| Mobile  | Hand-written + drift check | `mobile/lib/data/remote/dto/*.dart` | Dart DTOs (checked by `make check-openapi`) |

Backend domain entities in `backend/internal/domain/` and mobile domain entities in `mobile/lib/domain/` remain **hand-written**. Generated types cover the wire contract only; domain stays authoritative.

### Why no Dart codegen

A B1 spike (openapi-generator-cli `dart` and `dart-dio` templates) showed that producing clean Dart DTOs would force 4 new mobile dependencies (`json_annotation`, `json_serializable`, `build_runner`, `copy_with_extension`) plus a build_runner step after every regen — heavy for ~15 small DTOs in a starter template.

Instead, the drift checker (`backend/cmd/check-dto-drift`) catches the regression class that motivated contract-first in the first place (the `UserDto.role` field going missing from hand-written Dart). It reads `mobile/tool/dto_manifest.yaml`, loads each schema from `backend/api/openapi.yaml`, and asserts every required field is referenced by its snake_case JSON key in the corresponding Dart class. Zero mobile dependencies, ~150 LOC of Go, no build step. Revisit full codegen when the spec grows 3-5x.

## Workflow

```bash
# 1. Edit the spec
$EDITOR backend/api/openapi.yaml

# 2. Regenerate web + backend types
make gen

# 3. If a mobile DTO field changed, update the corresponding Dart DTO by hand.
#    If a new schema needs a Dart DTO, add the mapping in mobile/tool/dto_manifest.yaml.

# 4. Verify everything is in sync
make check-openapi

# 5. Commit the spec AND the regenerated files AND any Dart DTO changes
git add backend/api/openapi.yaml backend/internal/gen/api/ web/src/lib/api-types.ts \
        mobile/lib/data/remote/dto/ mobile/tool/dto_manifest.yaml
git commit -m "feat(api): <your change>"
```

If you forget step 4, CI's contract check will fail the PR.

## Rules

- Every schema used on the wire must declare a `required:` array. Missing `required` generates optional fields, which forces every consumer to guard fields the backend always returns.
- Every 4xx and 5xx response must reference `#/components/schemas/ErrorResponse` (the shared envelope: `code` + `message` + `request_id`).
- Non-JSON responses (OAuth 307 redirects, magic-link browser verify) must be documented with a description; do not invent a JSON body.
- New endpoints must land in the spec **before** the handler. The contract test (`backend/internal/app/contract_test.go`, B2) asserts every chi route has a spec path and vice versa.

## Error code conventions

Error codes are SNAKE_UPPER. The canonical list lives in `backend/internal/adapter/handler/respond.go` (`mapError`). When adding a new domain error:

1. Add the sentinel to `backend/internal/domain/errors.go`.
2. Map it in `respond.go`'s `mapError` switch.
3. Document the HTTP status + code in the endpoint's OpenAPI response.

## Route discovery

In development, the running backend exposes `GET /_debug/routes` (mounted only when `APP_ENV != "production"`). This endpoint is documented in the spec and returns the live chi route table. `make routes` curls and pretty-prints it.

## Adding a new endpoint (end-to-end)

1. Edit `backend/api/openapi.yaml`: add path, request schema, response schema, error responses.
2. `make gen:openapi` — regenerates Go + TypeScript + Dart types.
3. Implement the handler in `backend/internal/adapter/handler/` using the generated request/response types from `internal/gen/api`.
4. Wire the route in `backend/cmd/server/main.go`.
5. Use the generated client types from the web (`web/src/lib/api-types.ts`) and mobile (`mobile/lib/data/gen/`) sides; map to hand-written domain entities via per-aggregate mappers.
6. Contract test (B2) automatically covers route↔spec parity.
