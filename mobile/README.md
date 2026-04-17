# Initium Mobile

Flutter app for the Initium starter template. Targets iOS and Android.

## Quickstart

```bash
# From repo root — starts infra, backend (air), and web together
make dev

# Mobile only (from repo root)
make mobile-dev
# Equivalent, from mobile/ directly:
flutter run --dart-define-from-file=.env
```

> `make mobile-dev` boots the iOS Simulator automatically if none is running.

## Common commands

| Command | What it does |
|---------|--------------|
| `make mobile-test` | Run all Flutter unit + widget tests |
| `make mobile-gen` | Re-run build_runner (required after DTO/freezed changes) |
| `make mobile-lint` | Static analysis via `dart analyze` |
| `make mobile-build-apk` | Debug APK for Android |
| `make mobile-build-ios` | iOS simulator build |

## Platform setup

Before the app will compile you need platform-specific credentials (Google Sign-In, Firebase, URL schemes).
See **[mobile/SETUP.md](SETUP.md)** for step-by-step instructions.

## Environment config

Runtime config is injected at compile time via `--dart-define-from-file=.env`.
Copy the example and fill in your values:

```bash
cp mobile/.env.example mobile/.env
```

Do **not** commit `mobile/.env` — it is gitignored.

## Architecture overview

```
lib/domain/        # Pure Dart entities and interfaces
lib/data/          # DTOs, Dio client, repositories, secure storage
lib/providers/     # Riverpod DI wiring and AuthState
lib/presentation/  # Screens, widgets, go_router navigation
```

See [mobile/CLAUDE.md](CLAUDE.md) for detailed architectural rules.
