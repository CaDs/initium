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

> **What ships today.** Both apps ship an auth-gated 3-tab shell:
> login screen (magic link working; Google button stubbed) gives way
> to Home / Main / Settings after sign-in. The Home tab renders the
> authenticated user profile (email / name / role / id) mirroring
> `web/src/app/home/page.tsx`. Each app has a secure token store,
> a single-flight refresh interceptor, and deep-link handling for
> `initium://auth/verify?token=...`.
>
> **Still deferred** (do NOT pre-scaffold): Google Sign-In SDKs,
> OpenAPI codegen, theme switcher, locale switcher + i18n,
> Sentry / Firebase Crashlytics. Each lands in its own follow-up PR.

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

## The contract-first workflow (current — hand-written DTOs)

When a new API response needs a mobile client:

1. Edit `backend/api/openapi.yaml` (backend side).
2. Run `make gen:openapi` → regenerates Go + TypeScript types.
3. Update the hand-written DTOs:
   - iOS: `mobile/ios/initium/initium/API/Models.swift` — add / edit
     the `Codable` struct with explicit `CodingKeys` for snake_case
     field names.
   - Android: `mobile/android/app/src/main/java/com/example/initium/api/Models.kt` —
     add / edit the Moshi data class with `@Json(name = "...")` for
     snake_case field names.
4. Add a method on `APIClient` (iOS) / `ApiClient` (Android) that
   consumes or produces the new type.
5. `make check:parity` verifies the spec path is referenced on at
   least one surface.

**OpenAPI codegen is deferred** to a follow-up PR (SPM plugin on iOS,
Gradle plugin on Android). Until it lands, every spec change must be
mirrored manually in both `Models.*` files.

## Platform-specific conventions

See the two per-platform pattern docs, both of which include exemplar
file references checked by `make check:skills`:

- [patterns/ios.md](patterns/ios.md) — SwiftUI, Liquid Glass, `TabView`,
  `@Observable`, Swift Testing.
- [patterns/android.md](patterns/android.md) — Jetpack Compose,
  Material 3, `NavigationSuiteScaffold`, `StateFlow`, Compose UI tests.

## Testing

Both apps follow the **protocol/interface refactor pattern** so the
auth state machine is unit-testable without spinning up the real API
client or token store.

- **iOS**: `APIClientProtocol` + `TokenStorageProtocol` extracted in
  `mobile/ios/initium/initium/API/`. `AuthStore` accepts `any
  APIClientProtocol` and `any TokenStorageProtocol`. Tests use
  `MockAPIClient` + `MockTokenStorage` from
  `initiumTests/MockAPIClient.swift` <!-- expect: MockAPIClient -->.
  Run with `make test:ios` or `make test:ios:coverage`.
- **Android**: `ApiClient` + `TokenStorage` interfaces in
  `mobile/android/app/src/main/java/com/example/initium/api/`. The
  concrete class for the network client is `OkHttpApiClient`; the
  concrete token store remains `TokenStore` (now implements the
  interface). `AuthViewModel` accepts both interfaces. Tests use
  `FakeApiClient` + `FakeTokenStore` from
  `app/src/test/java/com/example/initium/api/FakeApiClient.kt` <!-- expect: FakeApiClient -->.
  ApiClient tests use `MockWebServer` (no Robolectric needed). Run
  with `make test:android` or `make test:android:coverage`
  (Jacoco; 25% line floor, ramps to 80% in follow-ups).
- Coverage gates are mobile-side only (not in `make preflight`)
  because Xcode + Android SDK aren't guaranteed in every dev
  environment. Run them when touching the mobile apps before merging.

## Canonical exemplars

Cross-platform (paths apply to both apps equally):

- Parity rule: `.claude/skills/_shared/parity.md`
- Mobile-wide agent doc: `mobile/AGENTS.md`

Auth state machines (protocol-injection ready):

- iOS: `mobile/ios/initium/initium/Auth/AuthStore.swift` <!-- expect: AuthStore -->
- Android: `mobile/android/app/src/main/java/com/example/initium/auth/AuthViewModel.kt` <!-- expect: AuthViewModel -->

Platform-specific exemplars live in the platform pattern docs.
