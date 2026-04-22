---
name: initium-mobile
description: Use when writing or modifying the Flutter mobile app in the Initium template — screens, Riverpod providers, Dio client, DTOs, mappers, go_router, i18n ARBs, or Flutter tests. Triggers on paths under `mobile/lib/**` or `mobile/test/**`. Encodes the layered data architecture, Riverpod `AuthState` sealed class, secure token storage, and the hand-written DTO + drift-check workflow for this specific fork-and-specialize starter template.
---

# initium-mobile

You are editing the Flutter app of an Initium fork. This template ships auth
(Google + magic link), session management (Keychain / EncryptedSharedPreferences),
Dio + refresh-token serialization, i18n (en/es/ja), and a Material 3 theme with
light/dark/system switching — all minimal, ready to be skinned.

> **This skill is authoritative.** If `mobile/CLAUDE.md` or the root `CLAUDE.md`
> contradicts anything here (they may still mention `@JsonSerializable`,
> `build_runner`, `freezed`, or old `mobile-gen`/`mobile-test` target names),
> the skill wins. Mobile does NOT use freezed or json_serializable; DTOs are
> hand-written. Make targets are namespaced: `make gen:mobile`,
> `make test:mobile`, `make check:openapi`, `make lint:mobile`.

## Architecture (layered, strict)

```
lib/domain/
  entity/         Pure Dart entities (User, Session). No Flutter, no package imports.
  repository/     Interfaces (UserRepository, AuthRepository).
  error/          Domain-level errors.
lib/data/
  remote/
    api_client.dart     Dio client wired with refresh interceptor.
    dto/                Hand-written DTOs (UserDto, AuthResponseDto, MessageResponseDto).
    mapper/             Per-aggregate DTO ↔ domain mappers (user_mapper.dart, etc).
  local/
    session_manager.dart   Plain Dart class — NOT a Riverpod Notifier.
    token_storage.dart     flutter_secure_storage wrapper with first-launch wipe.
  repository/     Repo implementations that compose remote + local.
lib/providers/
  api_provider.dart       tokenStorageProvider, apiClientProvider, authProvider (StateNotifier).
  auth_provider.dart      AuthState sealed class (Loading/Authenticated/Unauthenticated/Error).
  theme_provider.dart     ThemeMode with SharedPreferences persistence.
  locale_provider.dart    Locale with SharedPreferences persistence.
  <feature>_provider.dart Feature-specific StateNotifierProvider + repository Provider. Flat.
lib/presentation/
  router/         go_router config with Riverpod-driven redirects.
  login/          Login screen + Google button + magic link form.
  home/           Protected home screen.
  auth/           Magic link verify screen.
  shared/         DevModeBanner, ThemeSwitcher, LocaleSwitcher.
lib/l10n/         ARB files (app_en.arb, app_es.arb, app_ja.arb). Generated app_localizations*.dart.
```

## Rules

- `domain/` has **zero imports** from `data/`, `providers/`, `presentation/`, or any package.
- `AuthState` sealed class is a **UI concern**, not a domain entity. It lives in
  `providers/`, not `domain/`.
- `SessionManager` is plain Dart. Riverpod wraps it in `providers/api_provider.dart`.
- DTOs are hand-written JSON maps. They live only in `data/remote/dto/`. Their
  shape is verified by `make check:openapi` (the drift check in
  `backend/cmd/check-dto-drift` compares each mapped DTO against the OpenAPI spec).
- Mappers split per aggregate (`auth_mapper.dart`, `user_mapper.dart`). No
  single `dto_mapper.dart` kitchen sink.
- Use `Theme.of(context).colorScheme` and `theme.textTheme`. Never hardcode
  `Colors.grey[600]` or `TextStyle(fontSize: ...)`.
- No custom design system wrappers (no AppBtn, AppHeader, AppScaffold). Use
  raw `ElevatedButton`, `AppBar`, `Scaffold`. Forks add wrappers if they want them.
- No flutter_animate, no motion libraries, no parallax. Forks add them where
  specific screens need polish.
- **Feature providers (non-auth) live in `providers/<feature>_provider.dart`** —
  flat, never nested under `presentation/`. `presentation/` holds only widgets.
- **Repositories return `Future<(T?, DomainError?)>` positional records.** Never
  throw from repos; map `DioException` via a private `_mapError`. See
  `patterns/dio-client.md` and `user_repository_impl.dart` for the canonical shape.
- **Before registering a new DTO in `dto_manifest.yaml`**, verify the OpenAPI
  schema exists in `backend/api/openapi.yaml`. Registering a manifest entry for
  a nonexistent schema makes `make check:openapi` fail immediately.
- **When committing a new feature**: `git add -A` so untracked new files
  (entity, DTO, mapper, repo, provider, screen) all land. A `git diff` that
  misses untracked files is the most common "forgot half the feature" mode.

## The contract-first workflow

When a new API response (or new required field) needs a mobile DTO:

1. Someone edits `backend/api/openapi.yaml` (backend side).
2. Run `make gen:openapi` (regenerates Go + TypeScript types; leaves mobile alone).
3. If the schema is already in `mobile/tool/dto_manifest.yaml`, update the
   corresponding Dart DTO's `fromJson()` to reference the new field.
4. If the schema is new:
   - (a) **Confirm the schema exists in `backend/api/openapi.yaml` first.**
     Registering a manifest entry for a nonexistent schema breaks
     `make check:openapi` immediately.
   - (b) Hand-write the Dart DTO in `lib/data/remote/dto/`.
   - (c) Add a mapper in `lib/data/remote/mapper/`. Parse date-time strings
     into `DateTime` here; domain entities never carry wire types.
   - (d) Register **every** wire schema in `mobile/tool/dto_manifest.yaml` —
     response envelopes (`XxxList`), request bodies (`CreateXxxRequest`), AND
     the item schemas — each needs its own DTO class + manifest entry.
5. Run `make check:openapi` — it verifies every required schema field has a
   matching `json['snake_case_name']` reference in the target Dart class.
6. **List endpoints use envelope schemas** (`{"resource_name": [...]}`). Unwrap
   via `response.data['resource_name']` in the repo; do not treat
   `response.data` as a bare `List`. See `patterns/dio-client.md`.

Full workflow: `docs/OPENAPI.md`. Why no full Dart codegen: `docs/OPENAPI.md#why-no-dart-codegen`.

## Auth flow

- Google Sign-In: `google_sign_in` → ID token → `POST /api/auth/mobile/google` →
  backend returns `TokenPair`; tokens stored via `TokenStorage` (Keychain on iOS,
  EncryptedSharedPreferences on Android).
- Magic link: user enters email → `POST /api/auth/magic-link` → email deep link
  (initium://auth/verify?token=...) → `VerifyScreen` calls
  `POST /api/auth/mobile/verify` → tokens stored → `go('/home')`.
- Refresh: Dio interceptor uses a `Completer<void>` lock to serialize concurrent
  refresh attempts. Without the lock, simultaneous 401s cause spurious logouts.
- `DEV_BYPASS_AUTH`: injected via `--dart-define=DEV_BYPASS_AUTH=true`; emits
  authenticated state with a stub user. `main.dart` asserts it cannot be true
  in release builds.

## Security (non-obvious)

- **iOS keychain persistence**: Keychain items survive app uninstall. `TokenStorage`
  does a first-launch check via `shared_preferences` and wipes stale keychain
  data on reinstall.
- **Android backup**: `AndroidManifest.xml` sets `android:allowBackup="false"`
  to prevent `EncryptedSharedPreferences` leaking via Google backup.
- **Release guard**: `main.dart` asserts `!(kReleaseMode && devBypassAuth)`.

## i18n

- ARB files in `lib/l10n/app_{en,es,ja}.arb`. Add new keys to **all three**
  before referencing them.
- `flutter gen-l10n` (or `make gen:mobile`) regenerates `app_localizations*.dart`
  after ARB edits.
- Parameterized messages use `{name}` syntax with matching `@key` metadata.
- Access: `AppLocalizations.of(context)!`.

## Testing

- `flutter test` (or `make test:mobile`) runs widget + unit tests.
- Widget smoke test at `test/widget_test.dart`.
- Additional tests go under `test/data/`, `test/domain/`, `test/presentation/`
  (scaffolded empty).

## Canonical exemplars (open these when unsure)

- Entity: `mobile/lib/domain/entity/user.dart` <!-- expect: class User --> — pure Dart, no imports.
- DTO: `mobile/lib/data/remote/dto/user_dto.dart` <!-- expect: UserDto --> — hand-written `fromJson`.
- Mapper: `mobile/lib/data/remote/mapper/user_mapper.dart` <!-- expect: UserDtoMapper --> — extension on DTO.
- **Repository: `mobile/lib/data/repository/user_repository_impl.dart`** <!-- expect: _mapError --> —
  returns `Future<(T?, DomainError?)>` positional records; maps
  `DioException` via private `_mapError`. First file to copy for any new CRUD feature.
- Provider: `mobile/lib/providers/api_provider.dart` <!-- expect: authProvider --> — `tokenStorageProvider`,
  `apiClientProvider`, `authProvider` (StateNotifierProvider) wiring.
- Auth state: `mobile/lib/providers/auth_provider.dart` <!-- expect: AuthAuthenticated --> — sealed `AuthState`
  class + `AuthNotifier`. Screens reading `authProvider` must import BOTH files:
  `api_provider.dart` for the provider and `auth_provider.dart` for the
  `AuthState` variants used in pattern matching.
- Screen: `mobile/lib/presentation/login/login_screen.dart` <!-- expect: LoginScreen --> — raw Material,
  Riverpod `ref.watch`, localized strings.
- Home-to-sub-screen navigation: `mobile/lib/presentation/home/home_screen.dart` <!-- expect: HomeScreen -->
  uses `context.push('/path')` for detail routes (preserves back-stack).
- Router: `mobile/lib/presentation/router/app_router.dart` <!-- expect: routerProvider --> — go_router with
  Riverpod-driven redirects.
- DTO manifest: `mobile/tool/dto_manifest.yaml` <!-- expect: mappings --> — registering new DTOs for drift check.

See also: `patterns/riverpod-auth.md`, `patterns/dio-client.md`, `patterns/screen.md`, `patterns/feature-crud.md`.

## Parity

See [parity.md](../_shared/parity.md). If you add a screen here, the mirror
screen belongs on web. Call it out in the PR.
