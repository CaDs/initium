# Initium

*From the Latin word for "beginning."*

An opinionated full-stack starter template for experimenting with new ideas without burning tokens writing boilerplate code. Fork it, specialize it, ship your POC.

## Stack

| Component | Technology |
|-----------|-----------|
| Backend | Go (chi + GORM + PostgreSQL) |
| Frontend | Next.js (App Router, TypeScript, Tailwind) |
| Mobile | Flutter (Dart, Riverpod, Dio) |
| Infra | Docker Compose (PostgreSQL + Mailpit) |

## Quick Start

```bash
# Prerequisites: Go 1.23+, Node 20+, Flutter 3.22+, Docker
# Install golang-migrate: brew install golang-migrate

make setup    # Starts infra, installs deps, runs migrations, generates JWT keys
make dev      # Starts backend (8000) + web (3000)
```

Open http://localhost:3000 to see the landing page.

### Dev Bypass Auth

For immediate access to protected pages without configuring OAuth:

1. Set `DEV_BYPASS_AUTH=true` in `backend/.env` (enabled by default)
2. Start the backend
3. All protected endpoints return data for `dev@initium.local`

Configure Google OAuth when you're ready — not before.

## What's Included

Every POC needs the same boring foundation. Initium provides it so you can focus on your idea:

- **Passwordless auth** — Google OAuth + magic links, no password flows to build or secure
- **Landing page** — public page with CTA
- **Authenticated home** — protected dashboard, ready to customize
- **Session management** — short-lived JWTs + refresh token rotation, backend-owned
- **API spec** — OpenAPI 3.1 as the canonical contract
- **Dev tools** — hot reload, Docker infra, Mailpit for email testing, Makefile for everything

## Project Structure

```
initium/
├── backend/           # Go API server
├── web/               # Next.js frontend
├── mobile/            # Flutter app (iOS + Android)
├── docker-compose.yml
├── Makefile           # All dev commands (make help)
└── CLAUDE.md          # AI assistant conventions
```

## Development Paths

### Backend

```bash
make infra-up                   # Start PostgreSQL + Mailpit
make backend-dev                # Hot reload via air
make backend-test               # Run tests
make backend-lint               # Lint
make db-migrate                 # Run migrations
make db-create NAME=add_orders  # Create new migration
```

API spec: `backend/api/openapi.yaml`

### Frontend

```bash
make infra-up         # Start PostgreSQL + Mailpit
make backend-run      # Start backend (needed for API)
make web-dev          # Next.js dev server on port 3000
make web-test         # Run tests
make web-lint         # Lint
```

### Mobile

```bash
make infra-up         # Start PostgreSQL + Mailpit
make backend-run      # Start backend (needed for API)
make mobile-dev       # Flutter run with env config
make mobile-test      # Run tests
make mobile-gen       # Required after DTO changes
```

See `mobile/SETUP.md` for Google Sign-In platform configuration.

## Architecture

Ports & Adapters (hexagonal) at infrastructure boundaries:

```
domain/    Pure entities, interfaces, errors. Zero framework imports.
service/   Business logic implementing domain interfaces.
adapter/   HTTP handlers, persistence, middleware, DTOs.
infra/     Config, DB, JWT, OAuth, email. Outermost ring.
```

The dependency rule: inner layers never import outer layers.

See `CLAUDE.md` for detailed conventions per component.

## Auth Model

- **No passwords** — Google OAuth + magic links only
- Backend owns session state (JWTs + refresh tokens in PostgreSQL)
- Web: httpOnly cookies set by backend
- Mobile: tokens stored in Keychain (iOS) / EncryptedSharedPreferences (Android)
- Magic links are single-use (enforced via DB)
- Refresh token rotation on every refresh

## Local Dev URLs

| URL | Service |
|-----|---------|
| http://localhost:3000 | Next.js frontend |
| http://localhost:8000 | Go backend API |
| http://localhost:8025 | Mailpit (view magic link emails) |

## All Commands

Run `make help` to see the full list.
