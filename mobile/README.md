# mobile

A single Expo app that targets iOS and Android from one TypeScript
codebase. Replaces the prior native SwiftUI + Jetpack Compose apps so
forking POCs no longer pays a two-toolchain tax.

## Quick start

```sh
cd mobile && npm install
cp .env.example .env

make dev:mobile             # Metro + QR for Expo Go on a real device
make test:mobile            # Jest suite
make lint:mobile            # ESLint + tsc --noEmit
```

`make preflight` includes the mobile lint, typecheck, and test suite —
no Xcode or Android Studio needed.

## Where things live

- `app/` — Expo Router routes (file-based)
- `src/api/` — `APIClient`, token storage, models, endpoint helpers
- `src/auth/` — Zustand auth store + selectors
- `src/ui/` — shared UI primitives
- `__tests__/` — Jest specs

See `AGENTS.md` for cross-stack rules and the
`.claude/skills/initium-mobile/` skill for conventions.
