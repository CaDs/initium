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

## Current state — mobile parity is paused

The Flutter app was removed on branch `feat/dropping_flutter` and
replaced by two native apps (`mobile/ios/` SwiftUI, `mobile/android/`
Jetpack Compose). The native apps start from a 3-tab shell with no auth
and no API calls. That means:

- The `check:parity` gate scans only `web/src/**` right now. Spec paths
  with no web consumer still fail CI.
- Mobile-specific paths (`/api/auth/mobile/*`) are in `check:parity`'s
  exclusion list — see `backend/cmd/check-parity/main.go`.
- Mobile's half of parity is enforced by convention, not CI. Reviewers
  must flag feature PRs that skip the native mirrors.

When the native apps are caught up, teach `check:parity` to also scan
`mobile/ios/**/*.swift` and `mobile/android/**/*.kt`, remove the
`/api/auth/mobile/*` exclusions, and delete this "current state"
section.

## Why

This template is forked per project. Every fork's first month is spent
on auth, profile, settings, and other mechanics that must look the same
on every surface. Drift between web, iOS, and Android at that stage is
expensive to fix later. The contract-first workflow (see
`docs/OPENAPI.md`) will enforce shape parity once mobile codegen lands;
this rule enforces behavior parity today.
