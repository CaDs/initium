---
name: initium-mobile
description: Use when writing or modifying either native mobile app in the Initium template — the SwiftUI iOS app (`mobile/ios/initium/`) or the Jetpack Compose Android app (`mobile/android/`). Triggers on paths under `mobile/ios/**` or `mobile/android/**`. Encodes the two-app structure, the parity-with-web rule, and the per-platform conventions for this fork-and-specialize starter template.
---

# initium-mobile

You are editing one of **two native mobile apps** in this Initium fork. The
Flutter codebase was removed on branch `feat/dropping_flutter`; the mobile
surface is now:

- **iOS**: SwiftUI, targeting iOS 17.0+, built with Xcode 26+. Liquid Glass
  is supported as an *opt-in* per-surface treatment that falls back
  gracefully on iOS < 26 via `.regularMaterial`. See
  `patterns/ios.md` for conventions and the `.liquidGlassCard()` modifier.
- **Android**: Jetpack Compose + Material 3, Kotlin 2.x, minSdk 24 /
  targetSdk 36. Uses `NavigationSuiteScaffold` for adaptive nav (bottom bar
  on phones, nav rail on larger screens). See `patterns/android.md`.

> **Scope of the current MVP is tiny on purpose.** Both apps ship a 3-tab
> shell (Home / Main / Settings) that renders text per tab — nothing
> networked, no auth, no secure storage, no i18n. Everything that used to
> be in the Flutter app (Google sign-in, magic link, token refresh, theme
> switcher, en/es/ja localization) is **deferred** until a feature needs
> it. Do not pre-scaffold auth or API clients; add them when the first
> feature blocks on them.

## Gates that will fail your PR

Run `make preflight` before committing. It fails if any of the following
is true:

- Every `/api/*` spec path that should have a web consumer doesn't
  (`make check:parity`). Mobile-side parity is **paused** while the
  native apps catch up — `/api/auth/mobile/*` is in the checker's
  exclusion list.
- An exemplar path cited in this skill no longer resolves, or an
  `<!-- expect: symbol -->` annotation has gone stale
  (`make check:skills`).
- `git status --porcelain` is non-empty after the run
  (`make check:staged`).

Native mobile tests and linters are **not** part of `make preflight`
(Xcode and a simulator aren't guaranteed in every environment). Run them
explicitly: `make test:ios`, `make test:android`, `make lint:ios`,
`make lint:android`.

## The two-app structure

```
mobile/
├── AGENTS.md          — cross-platform rules (what applies to both apps)
├── CLAUDE.md          — symlink → AGENTS.md
├── README.md          — one-page index
├── ios/
│   └── initium/       — Xcode project root (`initium.xcodeproj` lives here)
│       ├── initium/           — app sources
│       ├── initiumTests/      — Swift Testing unit tests
│       ├── initiumUITests/    — XCUITest UI tests
│       ├── AGENTS.md
│       └── CLAUDE.md → AGENTS.md
└── android/
    ├── app/src/main/java/com/example/initium/   — Kotlin sources
    ├── gradle/libs.versions.toml                — version catalog
    ├── AGENTS.md
    └── CLAUDE.md → AGENTS.md
```

Shared product identity (bundle ID, app name, icons, marketing strings)
is NOT yet extracted to a single source of truth. Each app carries its
own copy. When the template matures, a `mobile/shared/` with a single
`brand.yaml` can be added and code-generated out.

## Cross-cutting rules

- **Parity**: see [../_shared/parity.md](../_shared/parity.md). The rule
  now reads "every user-facing feature on web, iOS, AND Android". If
  you ship a feature on one but not the other two, the PR description
  must explain why.
- **No OpenAPI codegen for mobile yet.** When API calls land, iOS will
  use `swift-openapi-generator`, Android will use the Kotlin target of
  `openapi-generator`. Do not hand-write DTOs — wait for the codegen
  wiring instead.
- **Bundle / package identifiers (current scaffolding state)**:
  - iOS: `cads.initium` (Xcode template default — rename per-fork)
  - Android: `com.example.initium` (Android Studio template default —
    rename per-fork)
  Forks should rename both to match the product. A mismatch between
  the two is fine during development; converging them is a `feat` PR
  when the product is named.
- **Commit hygiene**: when adding a new feature, include BOTH
  `mobile/ios/**` and `mobile/android/**` changes in the same commit
  (or two coordinated commits landing together). Half-feature PRs
  break parity and are harder to review.

## The contract-first workflow (future state)

When a new API response needs a mobile client:

1. Edit `backend/api/openapi.yaml` (backend side).
2. Run `make gen:openapi` → regenerates Go + TypeScript types.
3. **iOS**: (TBD — once `swift-openapi-generator` is wired, a codegen
   target will produce `Sources/InitiumAPI/`.)
4. **Android**: (TBD — once `openapi-generator` is wired, a Gradle
   task will produce `app/build/generated/openapi/`.)
5. Implement repository / data-source code that wraps the generated
   client.

Until that wiring is in place, do NOT add a feature to the mobile apps
that requires talking to the backend — the parity gate allows it
(mobile is paused), but it bypasses the contract enforcement. Prefer
pairing the codegen wiring with the first mobile feature that needs it.

## Platform-specific conventions

See the two per-platform pattern docs, both of which include exemplar
file references checked by `make check:skills`:

- [patterns/ios.md](patterns/ios.md) — SwiftUI, Liquid Glass, `TabView`,
  `@Observable`, Swift Testing.
- [patterns/android.md](patterns/android.md) — Jetpack Compose,
  Material 3, `NavigationSuiteScaffold`, `StateFlow`, Compose UI tests.

## Canonical exemplars

Cross-platform (paths apply to both apps equally):

- Parity rule: `.claude/skills/_shared/parity.md`
- Mobile-wide agent doc: `mobile/AGENTS.md`

Platform-specific exemplars live in the platform pattern docs.
