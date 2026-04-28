# Parity Rule

**Every user-facing feature must land on web AND mobile (Expo).**

When you touch one surface, you are accountable for the other. A PR
that adds a profile-editing form on web but not mobile is incomplete;
a mobile-only settings toggle with no web equivalent is incomplete.
"I'll do the other side later" is how drift starts.

## How to apply

- Reading a task: before starting, identify the parity target on the
  other surface. If the feature genuinely doesn't apply to both
  (e.g. a mobile-only push-notification toggle, a web-only keyboard
  shortcut), say so in the PR description with one sentence of
  justification per surface.
- Writing a PR description: include a **Parity** line per surface,
  naming the mirror work: "landed in this PR", "tracked in #X", or
  "N/A because Y".
- Reviewing: if the parity line is missing, request it before approving.

## Scope

Parity applies to **user-facing features**. It does not apply to:

- Backend-only changes (new handlers, middleware, migrations) — but
  adding a new endpoint usually implies new UI on web AND mobile.
- Surface-specific plumbing (Next.js middleware, Expo `app.json`
  config, ESLint configs).
- Internal refactors that don't change behavior.

## How `check:parity` enforces it

`backend/cmd/check-parity/main.go` walks two trees: `web/src/**`
(`.ts`, `.tsx`) and `mobile/**` (`.ts`, `.tsx`). Every `/api/*` spec
path must appear as a literal substring in at least one of them. A
missing consumer fails CI.

The gate enforces *any-surface* coverage, not *both-surface* — a spec
path referenced only by web still passes. The full two-surface rule
is enforced by convention during review. If you added an endpoint
that a surface can't practically consume (e.g. a browser-only
redirect), exclude it in the `excluded` map with a comment explaining
why.

## Why

This template is forked per project. Every fork's first month is spent
on auth, profile, settings, and other mechanics that must look the same
on every surface. Drift between web and mobile at that stage is
expensive to fix later. The contract-first workflow (see
`docs/OPENAPI.md`) keeps the *shape* of API responses in sync; this
rule keeps the *behavior* in sync.
