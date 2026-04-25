# Parity Rule

**Every user-facing feature must land on web, iOS, AND Android.**

When you touch one platform, you are accountable for the other two. A PR
that adds a profile-editing form on web but not iOS or Android is
incomplete; an iOS-only settings toggle with no Android equivalent is
incomplete. "I'll do the other sides later" is how drift starts.

## How to apply

- Reading a task: before starting, identify the parity target on the
  other two platforms. If the feature genuinely doesn't apply to all
  surfaces (e.g. an iOS-only share sheet using `UIActivityViewController`,
  a web-only keyboard shortcut, an Android-only intent handler), say so
  in the PR description with one sentence of justification per surface.
- Writing a PR description: include a **Parity** line per surface,
  naming the mirror work: "landed in this PR", "tracked in #X", or
  "N/A because Y".
- Reviewing: if the parity line is missing, request it before approving.

## Scope

Parity applies to **user-facing features**. It does not apply to:

- Backend-only changes (new handlers, middleware, migrations) — but
  adding a new endpoint usually implies new UI on web, iOS, AND Android.
- Platform-specific plumbing (Android manifest entries, iOS Info.plist,
  web middleware, Gradle / Xcode config, ktlint / detekt / SwiftLint
  config).
- Internal refactors that don't change behavior.

## How `check:parity` enforces it

`backend/cmd/check-parity/main.go` walks three trees: `web/src/**`
(`.ts`, `.tsx`), `mobile/ios/**` (`.swift`), `mobile/android/**`
(`.kt`). Every `/api/*` spec path must appear as a literal substring
in at least one of them. A missing consumer fails CI.

The gate enforces *any-surface* coverage, not *all-surface* — a spec
path referenced only by web still passes. The full three-surface rule
is enforced by convention during review. If you added an endpoint
that a platform can't practically consume (e.g. a browser-only
redirect), exclude it in the `excluded` map with a comment explaining
why.

## Why

This template is forked per project. Every fork's first month is spent
on auth, profile, settings, and other mechanics that must look the same
on every surface. Drift between web, iOS, and Android at that stage is
expensive to fix later. The contract-first workflow (see
`docs/OPENAPI.md`) will enforce shape parity once mobile codegen lands;
this rule enforces behavior parity today.
