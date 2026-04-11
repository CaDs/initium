# Initium

Opinionated POC starter template. Fork and specialize per project.

## Stack

| Component | Technology |
|-----------|-----------|
| Backend | Go (chi + GORM + PostgreSQL) |
| Frontend | Next.js (App Router, TypeScript, Tailwind) |
| Mobile | Flutter (Dart, Riverpod, Dio) |
| Infra | Docker Compose (PostgreSQL + Mailpit) |

## Quick Start

```bash
# Prerequisites: Go 1.23+, Node 20+, Flutter 3.22+, Docker, Task (taskfile.dev)
# Install golang-migrate: brew install golang-migrate

make setup    # Starts infra, installs deps, runs migrations, generates JWT keys
make dev      # Starts backend (8000) + web (3000)
```

Open http://localhost:3000 to see the landing page.

### Dev Bypass Auth

For immediate access to protected pages without configuring OAuth:

1. Set `DEV_BYPASS_AUTH=true` in `backend/.env`
2. Restart the backend
3. All protected endpoints return data for `dev@initium.local`

## Project Structure

```
initium/
├── backend/          # Go API server
├── web/              # Next.js frontend
├── mobile/           # Flutter app (iOS + Android)
├── docker-compose.yml
├── Taskfile.yml      # All dev commands
└── CLAUDE.md         # AI assistant conventions
```

## Development Paths

### Backend Developer

```bash
make infra-up         # Start PostgreSQL + Mailpit
make backend-dev      # Hot reload via air
make backend-test     # Run tests
make backend-lint     # Lint
make db-migrate       # Run migrations
make db-create NAME=add_orders  # Create new migration
```

API spec: `backend/api/openapi.yaml`

### Frontend Developer

```bash
make infra-up         # Start PostgreSQL + Mailpit
make backend-run      # Start backend (needed for API)
make web-dev          # Next.js dev server on port 3000
make web-test         # Run tests
make web-lint         # Lint
```

### Mobile Developer

```bash
make infra-up         # Start PostgreSQL + Mailpit
make backend-run      # Start backend (needed for API)
make mobile-dev       # Flutter run with env config
make mobile-test      # Run tests
make mobile-gen       # Required after DTO changes
```

See `mobile/SETUP.md` for Google Sign-In platform configuration.

## Auth Model

- **No passwords** — Google OAuth + magic links only
- Backend owns session state (JWTs + refresh tokens in PostgreSQL)
- Web: httpOnly cookies set by backend
- Mobile: tokens stored in Keychain (iOS) / EncryptedSharedPreferences (Android)

## Architecture

Ports & Adapters at infrastructure boundaries:

- `domain/` — Pure entities, interfaces, errors. Zero framework imports.
- `service/` (backend) — Business logic implementing domain interfaces.
- `adapter/` — HTTP handlers, persistence, middleware.
- `infra/` — Config, DB, JWT, OAuth, email. Outermost ring.

See `CLAUDE.md` for detailed conventions.

## Useful URLs (Local Dev)

| URL | Service |
|-----|---------|
| http://localhost:3000 | Next.js frontend |
| http://localhost:8000 | Go backend API |
| http://localhost:8025 | Mailpit (view magic link emails) |
