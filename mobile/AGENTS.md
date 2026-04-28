# Mobile — Expo (React Native + TypeScript)

This folder is one Expo app that targets iOS and Android from a single
TypeScript codebase. It replaces the prior native SwiftUI + Jetpack
Compose apps because forking POCs against two native toolchains slowed
iteration. The web Next.js app continues to ship in `web/`.

## Stack

- **Expo SDK 54** + React Native 0.81, TypeScript strict.
- **Expo Router** — file-based routing under `app/`, scheme `initium://`.
- **NativeWind v4** — Tailwind classes via `className=`. Tokens parity
  with `web/`.
- **Zustand** — single auth store at `src/auth/store.ts`.
- **`expo-secure-store`** — Keychain on iOS, EncryptedSharedPreferences
  on Android.
- **`expo-auth-session/providers/google`** — Google Sign-In, no native
  module dependency, works in Expo Go.

## Before you edit anything in this folder

Load the skill: `.claude/skills/initium-mobile/SKILL.md` plus
`.claude/skills/initium-mobile/patterns/expo.md`. Skills are the
authoritative source for conventions. This AGENTS.md is a pointer +
the cross-stack rules that don't live in the skill.

## Cross-stack rules

1. **Parity is non-negotiable.** Every user-facing feature must land on
   web AND mobile — see `.claude/skills/_shared/parity.md`. If one
   surface genuinely can't support the feature, justify it in the PR
   description with one sentence per surface.
2. **Bundle ID + scheme**. iOS `ios.bundleIdentifier` and Android
   `android.package` in `app.json` are both `com.initium.app`. Scheme
   is `initium://`. Forks rename via `app.json` only — there are no
   per-platform Xcode/Gradle files until `expo prebuild` runs (and
   those outputs are gitignored).
3. **No native modules without a dev build.** Expo Go bundles a fixed
   set. Anything else (Sentry SDKs, Firebase, native Google Sign-In)
   needs `eas build --profile development`. Don't add the dependency
   without the corresponding plumbing PR first.
4. **Hand-written DTOs for now.** `src/api/models.ts` mirrors
   `backend/api/openapi.yaml`. `openapi-typescript` codegen for mobile
   is deferred until a feature proves it worth wiring.

## Quick start

```sh
# 1. Install deps (only needed once or after package.json changes)
cd mobile && npm install

# 2. Copy env defaults
cp .env.example .env

# 3. Start Metro + the QR loop. Scan the QR with Expo Go on a real
#    device, or tap "open on iOS simulator" / "open on Android emulator"
#    if Xcode / Android Studio is installed.
make dev:mobile

# 4. Tests, lint, typecheck — all run in pure Node, no Xcode required.
make test:mobile             # Jest
make test:mobile:coverage    # Jest with 25% line floor
make lint:mobile             # ESLint + tsc --noEmit
```

`make preflight` runs `lint:mobile` and `test:mobile:coverage` as part
of the standard gate. No Xcode, no Android Studio.

## Deep-link contract

The backend mints magic links pointing at `initium://auth/verify?token=…`.
`app/auth/verify.tsx` reads `?token=` via `useLocalSearchParams`, calls
`useAuth().verifyMagicLink(token)`, and `<Redirect>`s to `(tabs)/home`
on success or `/login` on failure. The scheme is declared in
`app.json`'s `scheme` field — change it there if a fork wants a
different scheme.

## What a "good" mobile PR looks like

- Mirrors the equivalent web change in the same PR (or names the mirror
  in the description).
- Goes through `APIClient.send` — never hand-rolls `fetch`.
- Adds Jest specs under `__tests__/` for new logic; covers the API
  client + auth state machine when those change.
- Doesn't bring in React Navigation, Redux, or hand-written native
  modules. Doesn't pre-scaffold deferred plumbing (theme switcher,
  i18n, Sentry).

Keep PRs small. Match the MVP's minimalism until real features force
scaling up.
