# Mobile — Flutter (Dart, Riverpod, Dio)

> **For agents: the authoritative guide is `.claude/skills/initium-mobile/SKILL.md`
> + `patterns/*.md`.** Load it before making changes. This file is a human-facing
> quick reference.

## Build & Test

```bash
make dev:mobile             # flutter run with --dart-define-from-file=.env, boots simulator if needed
make test:mobile            # flutter test (widget + unit)
make gen:mobile             # flutter gen-l10n (after editing lib/l10n/*.arb)
make lint:mobile            # dart analyze
make check:openapi          # verify hand-written DTOs match backend/api/openapi.yaml
```

No `build_runner`, no `json_serializable`, no `freezed`. DTOs are hand-written.

## Architecture

```
lib/domain/           # Pure Dart. No Flutter, no package imports. Entities, interfaces, errors.
lib/data/remote/      # Dio client, hand-written DTOs + per-aggregate mappers
lib/data/repository/  # Implements domain interfaces; returns Future<(T?, DomainError?)>
lib/data/local/       # flutter_secure_storage, SessionManager (plain Dart class)
lib/providers/        # Riverpod DI wiring, AuthState sealed class, feature providers
lib/presentation/     # Screens, widgets, go_router
lib/l10n/             # ARB files + generated AppLocalizations
mobile/tool/          # dto_manifest.yaml (registered DTOs for the drift check)
```

## Key Rules

- `domain/` has zero imports from `data/`, `providers/`, `presentation/`, or any package.
- `AuthState` sealed class lives in `providers/auth_provider.dart` — UI concern, NOT a domain entity.
- `SessionManager` in `data/local/` is plain Dart, not a Riverpod Notifier. Riverpod wraps it in `providers/`.
- DTOs are **hand-written JSON maps** in `data/remote/dto/`. No annotations. Shape verified by `make check:openapi`.
- Mappers split per aggregate: `auth_mapper.dart`, `user_mapper.dart`. No single `dto_mapper.dart`.
- Repositories return `Future<(T?, DomainError?)>` positional records — never throw. Map `DioException` via `_mapError`.
- Feature providers (non-auth) live in `providers/<feature>_provider.dart`, flat — never nested under `presentation/`.
- List endpoints use envelope schemas (`{"resource_name": [...]}`). Unwrap before mapping.
- `pubspec.yaml` SDK: `^3.10.1` (Dart 3 sealed classes + records required).

## i18n

ARB files at `lib/l10n/app_{en,es,ja}.arb`.

- Access via `AppLocalizations.of(context)!` (import from `package:mobile/l10n/app_localizations.dart`).
- Add new keys to **all three** ARB files before using.
- Run `make gen:mobile` (or `flutter gen-l10n`) after ARB changes.
- Parameterized messages use `{name}` syntax with matching `@key` metadata.

## Theme

Material 3 via `ColorScheme.fromSeed(Colors.indigo)` in `main.dart`. Light/dark/system switcher persisted in `SharedPreferences`.

- Use `Theme.of(context).colorScheme` / `theme.textTheme`. Never hardcode colors or text styles.
- No custom wrapper widgets (no AppBtn, AppScaffold, etc). Raw Material only.

## Accessibility

- `Semantics` on interactive elements and status messages.
- `semanticsLabel` on non-decorative icons.
- `tooltip` on `IconButton`s and icon-bearing buttons.
- `autofillHints` on text fields only when a category applies (email, password, username, newPassword, oneTimeCode). Omit for free-form fields.
- `liveRegion: true` for dynamic status updates.
- Pair label + value in `Semantics(label: '$label: $value')` for profile rows.

## Auth

- Google Sign-In via `google_sign_in` → ID token → `POST /api/auth/mobile/google`.
- Magic link: enter email → backend sends link → deep link → `/auth/verify?token=...` → `POST /api/auth/mobile/verify`.
- Tokens stored in `flutter_secure_storage` (Keychain iOS / EncryptedSharedPreferences Android).
- `DEV_BYPASS_AUTH`: `--dart-define=DEV_BYPASS_AUTH=true` → stub user. Release-build guard in `main.dart`.

## Security — Non-Obvious

- **iOS keychain persistence**: Keychain survives app uninstall. `token_storage.dart` first-launch checks `shared_preferences` and wipes stale keychain.
- **Android backup**: `AndroidManifest.xml` has `android:allowBackup="false"`.
- **Token refresh race**: Dio interceptor uses `Completer<void>` lock — serializes concurrent 401s. Without it, racing requests cause spurious logouts.
- **Release guard**: `main.dart` asserts `!(kReleaseMode && devBypassAuth)`.

## Platform Setup

See `SETUP.md`:
- `google-services.json` (Android) + `GoogleService-Info.plist` (iOS).
- `Info.plist` URL scheme for Google Sign-In callback.
- SHA-1 fingerprint registration.

App will NOT compile without these.

## Gotchas

- Environment config uses `--dart-define-from-file=.env`, NOT `flutter_dotenv` (compile-time, not runtime).
- Navigation: `context.push` for detail routes (preserves back stack), `context.go` for redirects/top-level swaps. Never `Navigator.push`.
- Screens reading `authProvider` import BOTH `providers/api_provider.dart` (for the provider) AND `providers/auth_provider.dart` (for `AuthState` variants used in pattern matching).
