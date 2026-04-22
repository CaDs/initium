# Initium Mobile

Flutter app for the Initium starter template. Targets iOS and Android.

## Quickstart

```bash
# From repo root — starts infra, backend (air), and web together
make dev

# Mobile only (from repo root)
make dev:mobile
# Equivalent, from mobile/ directly:
flutter run --dart-define-from-file=.env
```

> `make dev:mobile` boots the iOS Simulator automatically if none is running.

## Common commands

| Command | What it does |
|---------|--------------|
| `make test:mobile` | Run all Flutter unit + widget tests |
| `make gen:mobile` | Regenerate Flutter localizations from lib/l10n/*.arb |
| `make lint:mobile` | Static analysis via `dart analyze` |
| `make format:mobile` | `dart format` the mobile tree |
| `make build:mobile:apk` | Debug APK for Android |
| `make build:mobile:ios` | iOS simulator build |

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

Detailed architectural rules and conventions live in the
[`initium-mobile` skill](../.claude/skills/initium-mobile/) — agents load
it automatically when editing `mobile/lib/**`.
