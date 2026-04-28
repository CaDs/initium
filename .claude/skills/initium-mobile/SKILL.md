---
name: initium-mobile
description: Use when writing or modifying the Expo (React Native + TypeScript) mobile app in the Initium template. Triggers on paths under `mobile/**`. Encodes the single-codebase Expo conventions, the parity-with-web rule, and the auth/deep-link contract for this fork-and-specialize starter template.
---

# initium-mobile

You are editing the **single Expo app** that ships with this Initium fork.
Native iOS (SwiftUI) and Android (Jetpack Compose) were dropped because
forking POCs against two native toolchains slowed iteration. The mobile
surface is now one TypeScript codebase that targets both platforms via
Expo Go (dev) and EAS Build (release).

## Stack at a glance

- **Runtime**: Expo SDK 54, React Native 0.81, TypeScript strict.
- **Router**: Expo Router v4, file-based, typed routes (`experiments.typedRoutes: true`).
- **Styling**: NativeWind v4 — Tailwind class strings with `className=`.
  Same tokens as `web/src/app/globals.css` so designs port directly.
- **State**: Zustand. One auth store at `mobile/src/auth/store.ts`. Do
  not introduce Redux, MobX, Recoil, or React Context for app-wide
  state.
- **Secure storage**: `expo-secure-store` (Keychain on iOS,
  EncryptedSharedPreferences on Android).
- **Google Sign-In**: `expo-auth-session/providers/google` —
  Expo-Go-compatible. No native module dependency.
- **Deep linking**: scheme `initium://`, declared in `mobile/app.json`.
  Magic links land at `mobile/app/auth/verify.tsx`.
- **Testing**: Jest (`jest-expo` preset) + `@testing-library/react-native`.
  Coverage floor 25% lines (enforced in `mobile/jest.config.js`).
- **Lint**: `eslint-config-expo` + Prettier.
- **Distribution**: Expo Go for dev (scan QR from `make dev:mobile`).
  `mobile/eas.json` is scaffolded with `development`/`preview`/`production`
  profiles for when a fork wants standalone builds.

## Gates that will fail your PR

Run `make preflight` before committing. It fails if any of the following
is true:

- Mobile lint or typecheck fails (`make lint:mobile` runs ESLint + `tsc --noEmit`).
- Mobile tests or coverage fail (`make test:mobile:coverage`, 25% line floor).
- Every `/api/*` spec path is consumed by either `web/src/**` or
  `mobile/**` (`make check:parity`). The Expo app is a real consumer
  now — `/api/auth/mobile/*` no longer needs an exclusion.
- Exemplar paths cited in this skill no longer resolve, or an
  `<!-- expect: symbol -->` annotation has gone stale
  (`make check:skills`).
- `git status --porcelain` is non-empty after the run
  (`make check:staged`).

Mobile gates run in pure Node — no Xcode, no Android Studio. Devices
are only needed for the QR-driven Expo Go dev loop.

## Layout

```
mobile/
├── app/                    Expo Router routes (file-based)
│   ├── _layout.tsx         root layout, auth bootstrap, splash
│   ├── index.tsx           bootstrap redirect (auth gate)
│   ├── login.tsx           magic-link form + Google button
│   ├── auth/verify.tsx     magic-link deep-link target
│   └── (tabs)/             3-tab authenticated shell
│       ├── _layout.tsx     Tabs navigator + auth guard
│       ├── home.tsx        /api/me profile card
│       ├── main.tsx        placeholder
│       └── settings.tsx    logout
├── src/
│   ├── api/
│   │   ├── client.ts       APIClient — single-flight 401 refresh
│   │   ├── tokens.ts       SecureStore wrapper
│   │   ├── models.ts       hand-written DTOs
│   │   ├── endpoints.ts    one method per /api/* path
│   │   └── deeplink.ts     parseDeepLink (pure)
│   ├── auth/
│   │   ├── store.ts        Zustand auth store
│   │   └── useAuth.ts      selector hooks
│   ├── ui/LiquidCard.tsx   expo-blur card
│   └── config.ts           env-driven config (API URL, Google IDs)
├── __tests__/              Jest specs
└── app.json | eas.json | metro.config.js | tailwind.config.ts | …
```

## Cross-cutting rules

- **Parity**: see [../_shared/parity.md](../_shared/parity.md). The rule
  reads "every user-facing feature on web AND mobile". One Expo app
  satisfies both store targets, so a feature lands twice (web + mobile),
  not three times.
- **Hand-written DTOs.** Mobile uses `mobile/src/api/models.ts` —
  `openapi-typescript` codegen is deferred until a feature proves it
  worth wiring. When the spec changes, update the matching DTO
  manually.
- **Bundle ID + scheme.** Both platforms use `com.initium.app` with
  scheme `initium://`. Forks rename via `mobile/app.json` (`ios.bundleIdentifier`,
  `android.package`, `scheme`).
- **One-PR-per-feature.** When you ship a feature, the PR touches both
  `web/src/**` and `mobile/**` (or names the parity mirror in the PR
  description with one sentence of justification per surface).

## What NOT to do

- **Do not introduce React Navigation.** Routing belongs to Expo Router.
- **Do not hand-roll `fetch`.** Always go through `APIClient.send` so
  the Bearer header + single-flight refresh stay consistent.
- **Do not add native modules without a dev build.** Expo Go bundles
  a fixed set of native modules; anything else (e.g.
  `@react-native-google-signin/google-signin`, Sentry SDKs, Firebase)
  requires `eas build --profile development` first.
- **Do not pre-scaffold deferred plumbing**: theme switcher, i18n
  (en/es/ja), Sentry/Crashlytics, OpenAPI codegen. Each lands in its
  own follow-up PR with the first feature that needs it.
- **Do not commit the `ios/` or `android/` directories** that
  `expo prebuild` creates — they're in `.gitignore` because Expo's
  managed workflow regenerates them.

## Auth flow (single-flight 401 refresh)

The shape of `mobile/src/api/client.ts` <!-- expect: refreshInFlight --> mirrors the deleted SwiftUI `APIClient.swift`:

1. `send()` reads the stored access token, attaches a `Bearer` header,
   fires the request via the injected `fetchImpl`.
2. On `401` (and `skipAuth` false), it calls `runRefresh()`. If a
   refresh is already in flight, callers `await` the same `Promise`
   instead of making a second `/api/auth/refresh` call (single-flight).
3. After the refresh resolves, the original request is retried once.
4. If the second attempt also returns `401`, `onUnauthorized()` fires
   and the auth store flips to `unauthenticated`.

Token persistence: `secureTokenStorage` in `mobile/src/api/tokens.ts`.
Reads + writes are async — there is no synchronous accessor.

## Magic-link deep linking

1. User taps the magic link in their email (Mailpit during dev).
2. The OS hands `initium://auth/verify?token=…` to the Expo app.
3. Expo Router maps the path to `mobile/app/auth/verify.tsx` <!-- expect: useLocalSearchParams -->,
   which reads `?token=` via `useLocalSearchParams`.
4. The component calls `useAuth().verifyMagicLink(token)`, which POSTs
   `{token}` to `/api/auth/mobile/verify` and stores the returned
   `TokenPair` via `secureTokenStorage`.
5. On success the store transitions to `authenticated` and the
   `<Redirect>` in the verify screen sends the user to `/(tabs)/home`.

## Google Sign-In

`mobile/app/login.tsx` uses `Google.useAuthRequest()` from
`expo-auth-session/providers/google`. Three OAuth client IDs are
required (`EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID`,
`EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID`,
`EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID`) — provision them in Google Cloud
Console. When unset, the button shows a "configure GOOGLE_*_CLIENT_ID
to enable" hint. On success, the `id_token` POSTs to
`/api/auth/mobile/google`.

## Testing convention

`mobile/__tests__/` holds Jest specs. The auth store at
`mobile/src/auth/store.ts` <!-- expect: create --> is testable by
calling `useAuthStore.getState()` directly and mocking `global.fetch`.
The `expo-secure-store` mock in `mobile/jest.setup.ts` keeps tokens in
process memory across the test run, so writes from one test do not
leak into the next (call `tokens.clear()` in `beforeEach`).

For network-shape assertions on the API client, inject `fetchImpl` and
a memory-backed `TokenStorage` directly into the `APIClient` constructor
— see `mobile/__tests__/client.test.ts` for the pattern. Do not mock
`expo-secure-store` for client-level tests; pass an in-memory storage
implementation instead.

## Canonical exemplars

- API client (single-flight refresh): `mobile/src/api/client.ts`
- Token storage wrapper: `mobile/src/api/tokens.ts`
- Auth store: `mobile/src/auth/store.ts`
- Magic-link deep-link target: `mobile/app/auth/verify.tsx`
- Login screen (Google + magic link): `mobile/app/login.tsx`
- Tabbed shell: `mobile/app/(tabs)/_layout.tsx`
- Profile screen (mirrors web `/home`): `mobile/app/(tabs)/home.tsx`
- Parity rule: `.claude/skills/_shared/parity.md`
- Mobile-wide agent doc: `mobile/AGENTS.md`

Platform-specific patterns (Expo Router, NativeWind, SecureStore,
Jest setup) live in [patterns/expo.md](patterns/expo.md).
