# Backend — Go (chi + GORM + PostgreSQL)

## Build & Test

```bash
go run ./cmd/server          # Start server (port 8000)
go test ./... -v -race       # Tests with race detector
golangci-lint run ./...      # Lint
```

Hot reload: `air` (config in `.air.toml`).

## Architecture

```
cmd/server/main.go           # Composition root: wire deps, register routes
internal/domain/             # Entities, interfaces (port.go), errors — NO imports from outer layers
internal/service/            # Business logic implementing domain interfaces
internal/adapter/handler/    # chi HTTP handlers (thin controllers)
internal/adapter/middleware/  # auth, CORS, rate limit, request ID
internal/adapter/persistence/ # GORM repo implementations + models.go (GORM tags + mappers)
internal/infra/              # Config, DB setup, JWT, OAuth verifier, email sender
migrations/                  # golang-migrate SQL files
api/openapi.yaml             # Canonical API spec
```

## Key Rules

- `domain/` must have zero imports from `service/`, `adapter/`, `infra/`, or any third-party package
- GORM struct tags go in `adapter/persistence/models.go` with `toDomain()`/`fromDomain()` mappers
- All handlers use `adapter/handler/respond.go` for standardized JSON responses
- Error format: `{"code": "SNAKE_UPPER", "message": "...", "request_id": "..."}`
- `context.Context` is always the first parameter
- Errors wrap with `fmt.Errorf("context: %w", err)` — never naked returns
- Structured logging via `slog` — no `fmt.Println` or `log.Printf`
- Graceful shutdown in `infra/server.go` via `signal.NotifyContext`

## Migrations

```bash
task db:migrate              # Run pending migrations
task db:rollback             # Rollback last migration
task db:create NAME=xxx      # Create new migration pair
```

Never use `gorm.AutoMigrate`. Schema changes go through versioned SQL files in `migrations/`.

## Auth Endpoints

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| GET | `/auth/google` | No | Redirect to Google consent |
| GET | `/auth/google/callback` | No | Exchange code, set cookies |
| POST | `/auth/magic-link` | No | Send magic link email (rate limited) |
| GET | `/auth/verify` | No | Verify magic link token |
| POST | `/auth/mobile/google` | No | Verify mobile ID token |
| POST | `/auth/refresh` | Refresh token | Issue new access token |
| POST | `/auth/logout` | Yes | Revoke session |
| GET | `/me` | Yes | Get current user profile |
| PATCH | `/me` | Yes | Update profile |

## Testing

- Table-driven tests in `*_test.go`, `testify` for assertions
- `t.Parallel()` where safe
- Name pattern: `TestServiceName_Method_Scenario_ExpectedResult`
