# Android — Jetpack Compose + Material 3

This is the native Android app, replacing the Flutter app dropped on
branch `feat/dropping_flutter`.

## Toolchain

- **Android Studio** latest stable for interactive work.
- **AGP 9.x**, **Kotlin 2.2.x**, **Compose BOM**, **minSdk 24**,
  **targetSdk 36** — all pinned in `gradle/libs.versions.toml`.
- **Material 3** (`androidx.compose.material3:*`). Material 2 is
  forbidden (`androidx.compose.material:material`), with one exception:
  `material-icons-core` which ships the `Icons.Filled.*` set.
- **Jetpack Compose only.** No XML layouts, no view bindings for new
  code.

## Conventions

Load `.claude/skills/initium-mobile/patterns/android.md` before editing.
Highlights:

- Navigation uses `NavigationSuiteScaffold` (adaptive — bottom bar on
  phones, rail on tablets) driven by the `AppTab` enum in
  `app/src/main/java/com/example/initium/MainActivity.kt`.
- Tab state lives in `rememberSaveable` so rotation + process death
  preserve the selection.
- Use `Modifier.testTag("descriptive-id")` for anything a UI test
  needs to target; reference with `onNodeWithTag("descriptive-id")`.
- Version bumps go through `gradle/libs.versions.toml`, not scattered
  through individual `build.gradle.kts` files.

## Quick start

```sh
# Install + launch on a running emulator/device
make dev:android

# Tests
make test:android                      # JVM unit tests (./gradlew test)
make test:android:instrumented         # Compose UI tests (needs device)

# Lint
make lint:android                      # Android Lint (./gradlew lint)
```

First-time setup: open the folder in Android Studio, let Gradle sync,
then use `make dev:android` or the IDE run button.

## What NOT to do

- Don't add Hilt until there's something to inject. Manual DI is fine
  for three stateless screens.
- Don't add Retrofit/OkHttp/kotlinx.serialization until the first
  networked feature.
- Don't add Firebase / Play Services for the MVP.
- Don't break out a multi-module structure (`:core`, `:data`,
  `:feature-*`). `:app` is correct for now.
- Don't import `androidx.compose.material:material` (the M2 package).
  Use `androidx.compose.material3:*` plus `material-icons-core`.

## Files worth knowing

- `app/src/main/java/com/example/initium/MainActivity.kt` — `InitiumApp`
  + `AppTab` + `TabContent` all live here until they grow out of one
  file.
- `app/src/main/java/com/example/initium/ui/theme/Theme.kt` —
  `InitiumTheme` composable with dynamic-color support.
- `gradle/libs.versions.toml` — the only place versions live.
- `app/src/test/java/com/example/initium/ExampleUnitTest.kt` — unit
  tests.
- `app/src/androidTest/java/com/example/initium/ExampleInstrumentedTest.kt`
  — Compose UI tests.

## Application ID

`com.example.initium` (Android Studio default). Forks should:

1. Rename `namespace` and `applicationId` in `app/build.gradle.kts`.
2. Move Kotlin sources from `java/com/example/initium/` to
   `java/<new>/<package>/<path>/`.
3. Update the package declaration in every `.kt` file.
4. Update the manifest's `android:name=".MainActivity"` if you renamed
   the activity class.
