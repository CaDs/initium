# Mobile — native iOS + Android

This folder holds **two independent native apps**, replacing the
previous Flutter codebase (dropped on branch `feat/dropping_flutter`).

- **`ios/initium/`** — SwiftUI app, deployment target iOS 17.0+. Liquid
  Glass is supported as an opt-in treatment that falls back gracefully
  on iOS < 26. Xcode 26+ required to build.
- **`android/`** — Jetpack Compose + Material 3 app, minSdk 24,
  targetSdk 36, Kotlin 2.2.x, Gradle KTS + version catalog.

Both currently ship a minimal 3-tab shell (Home / Main / Settings) with
hardcoded text per tab. Nothing is networked, no auth, no secure
storage. Everything that used to be in the Flutter app (Google Sign-In,
magic-link verify, token refresh, theme/locale switchers, en/es/ja
localization) is **deferred** until a feature needs it — see the
`.claude/skills/initium-mobile/SKILL.md` "Deferred" section of each
pattern doc.

## Before you edit anything in this folder

Load the skill: `.claude/skills/initium-mobile/SKILL.md` (cross-platform
overview) plus the per-platform pattern file:

- `.claude/skills/initium-mobile/patterns/ios.md`
- `.claude/skills/initium-mobile/patterns/android.md`

Skills are the authoritative source for conventions. This AGENTS.md is
a pointer + the cross-platform rules that DON'T live in the skill:

## Cross-platform rules

1. **Parity is non-negotiable.** Every user-facing feature must land on
   web, iOS, AND Android — see `.claude/skills/_shared/parity.md`. If
   one platform genuinely can't support the feature, say so in the PR
   description with one sentence of justification per surface.

2. **Bundle / application IDs are not yet synchronized.** iOS uses
   `cads.initium`, Android uses `com.example.initium`. Forks must
   rename both to match the product. Don't change just one.

3. **When adding a shared feature, commit both sides together** — even
   if the two halves land as separate commits, they should be in the
   same PR. A PR that adds the iOS side without the Android side is
   incomplete.

4. **OpenAPI codegen is not wired yet.** When the first networked
   feature lands, use `swift-openapi-generator` on iOS and the Kotlin
   target of `openapi-generator` on Android. Do NOT hand-write DTOs —
   wait for the codegen plumbing.

## Quick start

```sh
# iOS (simulator)
make dev:ios              # boots simulator if needed, builds + launches

# Android (emulator or device must already be running)
make dev:android          # installs + launches the debug APK

# Tests
make test:ios             # Swift Testing (xcodebuild test)
make test:android         # JUnit unit tests (./gradlew test)
make test:android:instrumented  # Compose UI tests (needs running device)
```

`make preflight` does NOT run native mobile tests or linters — Xcode and
a simulator aren't guaranteed in every environment. Run the mobile
targets explicitly when touching this folder.

## What a "good" mobile PR looks like

- Touches both `ios/initium/` and `android/` for the same feature.
- Includes a **Parity** line in the description naming both mirrors.
- Runs `make test:ios` and `make test:android` locally and posts the
  result in the PR body.
- Does NOT add new dependencies (Firebase, Hilt, Koin, Sentry, etc.)
  without a separate "wire up X" PR first.
- Does NOT introduce SwiftUI view models before they're needed, Hilt
  before there's something to inject, or Retrofit/OkHttp before the
  first API call.

Keep PRs small, match the MVP's minimalism until real features force
scaling up.
