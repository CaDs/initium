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
make backend-test # Go tests with race detector
make web-test     # Vitest
make mobile-test  # Flutter tests
make mobile-gen   # Required after DTO changes (build_runner)
```

## Auth Model

- Backend owns session state (single source of truth)
- Short-lived access tokens (15min) + refresh tokens (7d) in sessions table
- Google OAuth (web: server-side flow, mobile: ID token via `/auth/mobile/google`)
- Magic links (web: redirect flow, mobile: `/auth/mobile/verify` returns JSON)
- `DEV_BYPASS_AUTH=true` (dev only): skips auth, injects `dev@initium.local`

## API Contract

`backend/api/openapi.yaml` is the canonical spec. When changing API responses:
1. Update `openapi.yaml` first
2. Update backend Go structs
3. Update `web/src/lib/schemas.ts` (Zod)
4. Update `mobile/lib/data/remote/dto/` + run `make mobile-gen`

## Conventions

- Conventional Commits: `feat:`, `fix:`, `test:`, `refactor:`, `docs:`, `chore:`
- No secrets in version control. Use `.env.example` templates.
- Parameterized queries only. No string interpolation for SQL.
- i18n: all user-facing strings localized in en/es/ja (details in platform CLAUDE.md)
- Theme: three modes — light/dark/system (details in platform CLAUDE.md)
- Accessibility: required baseline (details in platform CLAUDE.md)
- Run linters and tests after changes automatically. Never push or open PRs without approval.
