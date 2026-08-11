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

If `make db/migrate` fails with `no migration found for version N`, the
Postgres volume has a `schema_migrations` row that doesn't match the
files on disk (usually a leftover from a branch that added a migration,
got reverted, but left the volume dirty). Recover with
`make infra/reset && make db/migrate`.

## Read this first, then load your stack skill

```
.claude/skills/initium-backend/   — Go + chi + sqlc + PostgreSQL
.claude/skills/initium-effect/    — shared Effect contracts, failures, HTTP boundary
.claude/skills/initium-web/       — Next.js App Router + Server Actions + Effect
.claude/skills/initium-mobile/    — Expo (React Native + TypeScript) — single codebase, both platforms
.claude/skills/_shared/parity.md  — the "every feature on web AND mobile" rule
```

Load the backend skill for `backend/**`, the Effect skill for
`packages/core/**`, the web skill for `web/**`, and the mobile skill for
`mobile/**`. When web or mobile work touches API contracts, request handling,
typed failures, or auth state, load the Effect skill as well. Read the selected
SKILL.md and its existing `patterns/*.md` before making changes. Everything
stack-specific lives there. Do not infer conventions from training data.

For mobile, also load `mobile/AGENTS.md` for the cross-stack rules.

## Repository and Effect model

- The root is an npm workspace containing `packages/core`, `web`, and
  `mobile`. The root `package-lock.json` is the only lockfile; run dependency
  installation from the repository root.
- `@initium/core` owns shared Effect Schemas for wire DTOs, compile-time
  conformance with generated OpenAPI types, tagged failures, and the
  interruptible HTTP request primitive. Do not put UI, storage, cookies, or
  framework-specific behavior there.
- The Go/Huma backend remains the source of truth for API behavior and does
  not depend on Effect. Never hand-edit
  `packages/core/src/generated/api-types.ts`; regenerate it with
  `make gen/openapi`.
- Expected TypeScript failures belong in Effect's typed error channel. Run
  Effects only at application boundaries: Next.js actions/routes/components,
  Expo action atoms/hooks, and tests.
- Web keeps React and Next.js lifecycle/state primitives. Mobile uses Effect
  Atom for auth state and workflows. Effect is an explicit boundary and
  workflow tool, not a blanket replacement for framework APIs.

## Commands and project-local tools

- Make namespaces use `/`, never `:`: for example `make db/migrate`,
  `make test/web`, and `make dev/mobile/ios`. Colons remain valid inside npm
  script names such as `test:coverage`.
- Run `make help` for the supported command inventory. Remove obsolete or
  no-op targets instead of keeping compatibility aliases.
- `make setup` installs root workspace dependencies, starts infrastructure,
  ensures JWT keys exist, waits for PostgreSQL, and applies migrations.
- Go development tools are pinned and built into ignored `backend/bin/`.
  Targets build their required tool lazily; `make tools` provisions all of
  them. Do not require a globally installed `migrate`, Air, sqlc, or
  golangci-lint.
- `make keygen` is idempotent and must not overwrite an existing JWT keypair.
- Run `make test/contract` only against a backend started with
  `DEV_BYPASS_AUTH=false`; the contract suite intentionally exercises auth
  failures.

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

Which runs, in order: `lint` → coverage → generated-artifact drift →
`check/parity` → `check/skills` → `check/staged`. A green preflight means:
- Every chi route has a corresponding OpenAPI path and vice versa
  (`backend/internal/app/contract_test.go`).
- Every `/api/*` spec path has a consumer in at least one of
  `web/src/**` or `mobile/**` (`backend/cmd/check-parity`).
  Any-surface coverage passes the gate; both-surface parity is
  enforced by review (see `.claude/skills/_shared/parity.md`).
- Every exemplar path cited in a SKILL.md still exists and still contains
  the claimed symbol (`scripts/check-skills.sh`).
- Every domain error is mapped to an HTTP envelope
  (`error_envelope_test.go`).
- No untracked or unstaged files; staged changes are allowed.

Mobile (Expo) tests + lint run as part of `make preflight` in pure
Node — no Xcode and no Android Studio required. The QR-driven dev
loop (`make dev/mobile`) needs a real device with Expo Go installed,
but that's separate from the gate.

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

Edit the typed Huma handler first → run `make gen/openapi` → update the shared
Effect Schema in `packages/core` → add endpoint consumers on web and mobile →
`make preflight`. Generated OpenAPI TypeScript types and compile-time
conformance checks live in the shared package; see `docs/OPENAPI.md`.

## Auth model

- Backend owns session state. Short-lived access tokens (15min) + refresh
  tokens (7d) in the `sessions` table.
- Google OAuth — web: server-side redirect flow. Mobile: the Expo app
  uses `expo-auth-session/providers/google` to get an ID token, then
  POSTs it to `/api/auth/mobile/google`.
- Magic links — web: `/verify` redirects with cookies. Mobile: the deep
  link `initium://auth/verify?token=…` lands at `mobile/app/auth/verify.tsx`,
  which POSTs to `/api/auth/mobile/verify` and stores the returned
  pair in `expo-secure-store`.
- `DEV_BYPASS_AUTH=true` (dev only) injects `dev@initium.local`. Release
  builds hard-fail if the flag is on. The mobile equivalent is
  `EXPO_PUBLIC_DEV_BYPASS_AUTH=true`.

## Observability (shipped vs opt-in)

Shipped: `/healthz`, `/readyz`, `/metrics` (Prometheus default collectors),
slog JSON access log with request IDs.

Opt-in stubs (env vars only, no init code): `SENTRY_DSN` (backend),
`NEXT_PUBLIC_SENTRY_DSN` (web), `OTEL_EXPORTER_OTLP_ENDPOINT`. Mobile
observability (Sentry React Native SDK) is deferred until the Expo app
has real user-facing behavior worth instrumenting. Adding a native SDK
will require an EAS dev build (Expo Go does not bundle Sentry).

## Conventions (cross-cutting)

- **Conventional Commits**: `feat:`, `fix:`, `test:`, `refactor:`,
  `docs:`, `chore:`. Body explains _why_.
- **No secrets in version control**. Use `.env.example` templates.
- **Parameterized queries only**. Never string-interpolate SQL.
- **i18n**: web localizes in en/es/ja. The Expo app renders hardcoded
  English — localization is deferred until the MVP shell becomes a
  real app. Per-stack mechanics live in the skill.
- **Theme**: web supports light/dark/system. The Expo app adopts the
  system appearance via `userInterfaceStyle: "automatic"` in
  `mobile/app.json`. Explicit in-app theme switchers are deferred.
- **Accessibility**: required baseline per stack.
- **Autonomy**: run linters + tests after changes automatically. Never
  push, merge, or open PRs without explicit approval.
