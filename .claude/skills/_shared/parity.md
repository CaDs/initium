# Parity Rule

**Every user-facing feature must land on web AND mobile.**

When you touch one platform, you are accountable for the other. A PR that adds a
profile-editing form on web but not mobile is incomplete; a PR that adds an
admin screen on mobile but not web is incomplete. "I'll do the other side later"
is how drift starts.

## How to apply

- Reading a task: before starting, identify the parity target on the other
  platform. If the feature genuinely doesn't apply to one platform (e.g. an
  iOS-only share sheet, a desktop-only keyboard shortcut), say so in the PR
  description with one sentence of justification.
- Writing a PR description: include a **Parity** line that names the mirror
  work, either as "landed in this PR", "tracked in #X", or "N/A because Y".
- Reviewing: if the parity line is missing, request it before approving.

## Scope

Parity applies to **user-facing features**. It does not apply to:

- Backend-only changes (new handlers, middleware, migrations) — but adding a
  new endpoint usually implies new UI on both clients.
- Platform-specific plumbing (Android manifest entries, iOS Info.plist, web
  middleware).
- Internal refactors that don't change behavior.

## Why

This template is forked per project. Every fork's first month is spent on
auth, profile, settings, and other mechanics that must look the same on every
surface. Drift between web and mobile at that stage is expensive to fix later.
The contract-first workflow (see `docs/OPENAPI.md`) enforces shape parity; the
parity rule enforces behavior parity.
