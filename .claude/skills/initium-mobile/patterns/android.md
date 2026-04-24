# Android pattern (Jetpack Compose + Material 3)

You are editing the Compose app under `mobile/android/`. This file covers
the conventions that apply to the Android side specifically. For the
cross-platform rules see the parent `SKILL.md`.

## Toolchain

- **Android Studio** (latest stable) for interactive work.
- **AGP 9.x**, **Kotlin 2.2.x**, **Compose BOM** pinned in
  `gradle/libs.versions.toml`. Versions update via the version catalog,
  not scattered `implementation("...:x.y.z")` calls.
- **minSdk 24 / targetSdk 36 / compileSdk 36** (matches the current
  scaffold — adjust only with a clear reason).
- **Material 3**, not Material 2. `androidx.compose.material3:*` only —
  do not add `androidx.compose.material:material` (the M2 package).
  The exception is `androidx.compose.material:material-icons-core`,
  which is shared between M2 and M3 and ships the default `Icons.Filled.*`
  set.
- **Jetpack Compose only.** Do not introduce XML layouts or view
  bindings for new code.

## Project layout

```
mobile/android/                                       ← Gradle project root
├── settings.gradle.kts                               ← includes :app
├── build.gradle.kts                                  ← root plugins
├── gradle/libs.versions.toml                         ← version catalog
├── gradle.properties
├── gradlew, gradlew.bat
└── app/
    ├── build.gradle.kts                              ← module config
    └── src/
        ├── main/
        │   ├── AndroidManifest.xml
        │   └── java/com/example/initium/
        │       ├── MainActivity.kt                   ← setContent { InitiumTheme { InitiumApp() } }
        │       └── ui/theme/                         ← M3 theme, colors, typography
        ├── test/java/com/example/initium/            ← JUnit unit tests
        └── androidTest/java/com/example/initium/     ← Compose UI tests
```

Package name is `com.example.initium` (Android Studio template default).
Forks should rename to their own reverse-DNS identifier in
`build.gradle.kts` (`namespace`, `applicationId`) AND move the Kotlin
files into the new directory structure. Until that rename happens, do
NOT add new files under other package paths.

## Navigation

The root composable `InitiumApp()` in
`mobile/android/app/src/main/java/com/example/initium/MainActivity.kt`
<!-- expect: fun InitiumApp -->
uses `NavigationSuiteScaffold` from
`androidx.compose.material3:material3-adaptive-navigation-suite`. This
is the adaptive nav container — it renders a bottom `NavigationBar` on
phones and a `NavigationRail` on larger/foldable screens automatically.

Rules:

- Tab state lives in `rememberSaveable { mutableStateOf(AppTab.HOME) }` so
  rotation / process death preserve selection without external persistence.
- The `AppTab` enum owns both `label` and `icon` (an `ImageVector`). Add
  a new tab by extending the enum; the `forEach` loop picks it up.
- Each tab's content renders through a single `TabContent` composable
  that switches on the selected tab. When a tab needs real screen
  structure (toolbars, back-stack), graduate it to its own composable
  file under `ui/screens/<tab>/` and compose it from `TabContent`.

## State management (future state)

Current code is stateless beyond the tab selector. When real state lands:

- **ViewModel** per screen, scoped via `hiltViewModel()` (once Hilt is
  added) or `viewModel()` factory otherwise.
- **`StateFlow<UiState>`** exposed from the ViewModel; the composable
  uses `val state by vm.state.collectAsStateWithLifecycle()`.
- **One-shot events** via `SharedFlow<Event>` consumed with
  `LaunchedEffect`.
- **No `LiveData`.**

## Tests

- **Unit tests** (host-side JVM): `app/src/test/java/...` — see
  `mobile/android/app/src/test/java/com/example/initium/ExampleUnitTest.kt`
  <!-- expect: AppTabUnitTest --> for the current enum-property test.
  Run with `make test:android` (or `./gradlew test`).
- **Compose UI tests** (instrumented, needs emulator/device):
  `app/src/androidTest/java/...` — see
  `mobile/android/app/src/androidTest/java/com/example/initium/ExampleInstrumentedTest.kt`
  <!-- expect: InitiumAppTabsTest -->. Run with
  `make test:android:instrumented` (or `./gradlew connectedAndroidTest`).

Use `Modifier.testTag("descriptive-id")` on composables you want to
target from UI tests, then `onNodeWithTag("descriptive-id")` in the test.
Test tags are stripped in release builds.

## Theme

The M3 theme is defined in
`mobile/android/app/src/main/java/com/example/initium/ui/theme/Theme.kt`
<!-- expect: InitiumTheme -->
and supports dynamic color on Android 12+. The initial color seed is
Material's default purple — forks should rebrand this as the first UI
change.

## Build

- Debug APK: `make build:android` or `./gradlew assembleDebug`.
- Install + launch on a running emulator/device: `make dev:android`.
- Release + signing: not yet wired (fork concern).

## Deferred (do NOT scaffold yet)

- Hilt (or any DI container).
- ktlint / detekt plugin wiring.
- Retrofit / OkHttp / kotlinx.serialization — wait for the first
  networked feature.
- EncryptedSharedPreferences / DataStore with Crypto.
- Google Sign-In SDK.
- Multi-module split (`:core`, `:data`, `:feature-*`). A single
  `:app` module is correct for a 3-tab MVP.
- Firebase / Play Services.
