# Mobile — Flutter (Dart, Riverpod, Dio)

## Build & Test

```bash
flutter run --dart-define-from-file=.env    # Run with env config
flutter test                                 # Unit + widget tests
flutter test integration_test/               # Integration tests
dart run build_runner build                  # REQUIRED after DTO/freezed changes
dart analyze                                 # Static analysis
```

## Architecture

```
lib/domain/           # Pure Dart. No Flutter, no package imports. Entities, interfaces, errors.
lib/data/remote/      # Dio client, @JsonSerializable DTOs, per-aggregate mappers
lib/data/repository/  # Implements domain interfaces using remote + local
lib/data/local/       # flutter_secure_storage, SessionManager (plain Dart class)
lib/providers/        # Riverpod providers — DI wiring + AuthState sealed class
lib/presentation/     # Screens, widgets, go_router
```

## Key Rules

- `domain/` has zero imports from `data/`, `providers/`, `presentation/`, or any package
- `AuthState` sealed class lives in `providers/auth_provider.dart` — it is a UI concern, NOT a domain entity
- `SessionManager` in `data/local/` is a plain Dart class, NOT a Riverpod Notifier. Riverpod wraps it in `providers/`.
- DTOs use `@JsonSerializable`/`@freezed` — annotations live in `data/remote/dto/`, never in `domain/`
- Mappers split per aggregate: `auth_mapper.dart`, `user_mapper.dart` — no single `dto_mapper.dart`
- Riverpod triple must be version-matched in `pubspec.yaml` (flutter_riverpod, riverpod_annotation, riverpod_generator)

## Auth

- Google Sign-In via `google_sign_in` package → ID token → `POST /auth/mobile/google`
- Magic link: enter email → backend sends link → deep link back to app → verify token
- Tokens stored via `flutter_secure_storage` (Keychain on iOS, EncryptedSharedPreferences on Android)
- `DEV_BYPASS_AUTH`: injected via `--dart-define=DEV_BYPASS_AUTH=true`, emits authenticated state with stub user

## Security — Non-Obvious

- **iOS keychain persistence**: Keychain items survive app uninstall. `token_storage.dart` does a first-launch check using `shared_preferences` and wipes stale keychain data on reinstall.
- **Android backup**: `AndroidManifest.xml` sets `android:allowBackup="false"` to prevent EncryptedSharedPreferences from leaking via Google backup.
- **Token refresh race**: Dio interceptor uses a `Completer<void>` lock to serialize concurrent refresh attempts. Without this, simultaneous 401s cause spurious logouts.
- **Release guard**: `main.dart` asserts `!(kReleaseMode && devBypassAuth)` — DEV_BYPASS_AUTH cannot be enabled in release builds.

## Platform Setup

See `SETUP.md` for:
- `google-services.json` (Android) — from Firebase/GCP console
- `GoogleService-Info.plist` (iOS) — from Firebase/GCP console
- `Info.plist` URL scheme for Google Sign-In callback
- SHA-1 fingerprint registration

App will NOT compile without these files configured.

## Gotchas

- `dart run build_runner build` after ANY change to `@JsonSerializable`, `@freezed`, or `@riverpod` annotated code
- Environment config uses `--dart-define-from-file=.env`, NOT `flutter_dotenv` (compile-time, not runtime)
- `pubspec.yaml` SDK constraint: `>=3.3.0 <4.0.0` (Dart 3 sealed classes required)
