# mobile

Two independent native apps:

- **`ios/initium/`** — SwiftUI, iOS 17.0+, Liquid Glass opt-in, Xcode 26+.
- **`android/`** — Jetpack Compose + Material 3, minSdk 24, Kotlin 2.2.x.

Both ship the same 3-tab shell (Home / Main / Settings) as starter
scaffolding. See `AGENTS.md` in this folder for cross-platform rules
and the two per-platform `AGENTS.md` files for platform-specific
conventions.

## Quick start

```sh
make dev:ios        # build + run the iOS app on a simulator
make dev:android    # install + launch the Android app on a device/emulator

make test:ios       # Swift Testing
make test:android   # Android JVM unit tests
```

The Flutter app that used to live here was dropped on branch
`feat/dropping_flutter`. See that branch's history for context on what
was removed.
