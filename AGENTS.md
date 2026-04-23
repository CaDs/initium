# Initium

Agent-first POC starter template. Fork and specialize per project.

This file is always loaded into the agent's context. It covers invariants
that hold across every stack and points you at the stack-specific skill
that owns everything else.

## Running multiple forks in parallel

`docker-compose.yml` reads port + credential + project-name vars from a
root `.env` (copied from `.env.example` by `make setup`). Defaults keep
fresh clones working out of the box. If you run a second fork next to
this one and hit `port is already allocated`, edit `.env` and change
`POSTGRES_PORT` / `MAILPIT_SMTP_PORT` / `MAILPIT_HTTP_PORT` to free
values, then re-run `make setup`.

If `make db:migrate` fails with `no migration found for version N`, the
Postgres volume has a `schema_migrations` row that doesn't match the
files on disk (usually a leftover from a branch that added a migration,
got reverted, but left the volume dirty). Recover with
`make infra:reset && make db:migrate`.

## Read this first, then load your stack skill

```
.claude/skills/initium-backend/   — Go + chi + GORM + PostgreSQL
.claude/skills/initium-web/       — Next.js App Router + Server Actions + Zod
.claude/skills/initium-mobile/    — Flutter + Riverpod + Dio + hand-written DTOs
.claude/skills/_shared/parity.md  — the "every feature on web AND mobile" rule
```

If you're editing `backend/**`, `web/**`, or `mobile/**`, load the matching
SKILL.md + its `patterns/*.md` before making changes. Everything
stack-specific lives there. Do not infer conventions from training data.

## Agent-first principles

1. **Think before coding.** State assumptions; ask when uncertain; surface
   alternatives rather than pick silently.
2. **Simplicity first.** Minimum code that solves the problem. No
   speculative flexibility. No error handling for impossible scenarios.
3. **Surgical changes.** Don't improve adjacent code. Remove only imports
   your changes made unused. Every changed line traces to the request.
4. **Goal-driven.** Define success criteria (usually a failing test), then
   loop until the gates are green.

## Gates — the actual binding mechanism

Prose rules are suggestions. The rules that bind are the ones that fail
CI. Run the full gate before committing:

```bash
make preflight
```

Which runs, in order: `lint` → `test` → `check:openapi` → `check:parity`
→ `check:skills` → `check:staged`. A green preflight means:
- Every chi route has a corresponding OpenAPI path and vice versa
  (`backend/internal/app/contract_test.go`).
- Every required schema field is referenced in its mobile DTO
  (`backend/cmd/check-dto-drift`).
- Every JSON-returning spec path has a web Zod schema + mobile DTO mapping
  (`backend/cmd/check-parity`).
- Every exemplar path cited in a SKILL.md still exists and still contains
  the claimed symbol (`scripts/check-skills.sh`).
- Every domain error is mapped to an HTTP envelope
  (`error_envelope_test.go`).
- No untracked files — `git status --porcelain` is empty.

If you add a rule to a skill, consider whether a gate can enforce it. If
yes, add the gate. Prose-only rules drift.

## Architecture invariants (hexagonal)

| Layer | Rule |
|-------|------|
| `domain/` | Zero framework imports. Pure entities, interfaces, errors. |
| `service/` | Imports domain only. Business logic. |
| `adapter/` | Handlers, persistence, DTOs. Imports domain + service. |
| `infra/` | Config, DB, external services. Outermost ring. |

Inner layers never import outer layers. This is enforced by convention,
caught in review.

## API contract workflow (one sentence)

Edit `backend/api/openapi.yaml` first → run `make gen:openapi` → implement
handler using generated types → update web Zod schema + mobile DTO (if
needed) → `make preflight`. The skills cover the per-stack mechanics.

## Auth model

- Backend owns session state. Short-lived access tokens (15min) + refresh
  tokens (7d) in the `sessions` table.
- Google OAuth — web: server-side redirect flow; mobile: ID token POSTed
  to `/api/auth/mobile/google`.
- Magic links — web: `/verify` redirects with cookies; mobile:
  `/api/auth/mobile/verify` returns JSON.
- `DEV_BYPASS_AUTH=true` (dev only) injects `dev@initium.local`. Release
  builds hard-fail if the flag is on.

## Observability (shipped vs opt-in)

Shipped: `/healthz`, `/readyz`, `/metrics` (Prometheus default collectors),
slog JSON access log with request IDs.

Opt-in stubs (env vars only, no init code): `SENTRY_DSN` (backend +
mobile), `NEXT_PUBLIC_SENTRY_DSN` (web), `OTEL_EXPORTER_OTLP_ENDPOINT`.
Pick one, wire it in the relevant stack's `main.go` / root layout.

## Conventions (cross-cutting)

- **Conventional Commits**: `feat:`, `fix:`, `test:`, `refactor:`,
  `docs:`, `chore:`. Body explains _why_.
- **No secrets in version control**. Use `.env.example` templates.
- **Parameterized queries only**. Never string-interpolate SQL.
- **i18n**: every user-facing string localized in en/es/ja. Per-stack
  mechanics live in the skill.
- **Theme**: three modes — light/dark/system.
- **Accessibility**: required baseline per stack.
- **Autonomy**: run linters + tests after changes automatically. Never
  push, merge, or open PRs without explicit approval.
