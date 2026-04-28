# Expo conventions

Patterns the Initium Expo app relies on. Treat these as defaults — only
deviate with a one-line PR-description justification.

## Routing — Expo Router

- Files under `mobile/app/` are routes. `_layout.tsx` files declare
  navigators (Stack, Tabs).
- `(group)` directories organize routes without affecting URLs.
- Typed routes are on (`experiments.typedRoutes` in `mobile/app.json`),
  so `<Link href="/(tabs)/home" />` is checked at compile time.
- Auth gating: the route's component returns `<Redirect href="/login" />`
  when the user is not authenticated; do NOT push from `useEffect`. Two
  navigators must not declare the same screen.
- Read deep-link query params with
  `useLocalSearchParams<{ token?: string }>()`. Do not parse `window.location`
  on native.

## Styling — NativeWind v4

- Tailwind class strings via `className=`. Type definitions come from
  `nativewind-env.d.ts`.
- Configure tokens once in `mobile/tailwind.config.ts`. Don't fork the
  config per screen.
- Dark mode: use `dark:` variants. The system appearance flows in via
  `userInterfaceStyle: "automatic"` in `mobile/app.json`; do not add a
  manual theme switcher until product wants one.
- Don't mix `StyleSheet.create()` with NativeWind in the same component
  unless interpolating a runtime value Tailwind can't express. Pick
  one approach per file.

## Secure storage — `expo-secure-store`

- All access goes through `mobile/src/api/tokens.ts`. The bare
  `SecureStore` API is fine inside `tokens.ts`; everywhere else, use
  the `TokenStorage` interface so tests can inject an in-memory double.
- Keys are namespaced (`initium.access_token`, `initium.refresh_token`)
  to avoid collisions with libraries that also write to the keychain.
- All methods are async — `getItemAsync`, `setItemAsync`,
  `deleteItemAsync`. There is no synchronous accessor.

## Auth state — Zustand

- One store: `mobile/src/auth/store.ts`. Selectors live in
  `mobile/src/auth/useAuth.ts`.
- Mutations are methods on the store (`bootstrap`, `verifyMagicLink`,
  `verifyGoogle`, `logout`). State changes happen via `set(...)` calls
  inside those methods.
- Don't read tokens via `useAuth().tokens.read()` in screens — call the
  store method that needs the token (`bootstrap` already encapsulates
  the read).
- Tests can call `useAuthStore.setState(...)` to seed state and
  `useAuthStore.getState().tokens.clear()` to reset between cases.

## API calls

- Always go through `APIClient.send` (`mobile/src/api/client.ts`).
  Endpoint definitions live in `mobile/src/api/endpoints.ts` —
  one function per `/api/*` path.
- Pass `skipAuth: true` for public routes (`/api/landing`,
  `/api/auth/magic-link`, `/api/auth/mobile/*`) so the Bearer header
  isn't attached.
- Errors raise `APIError` (with the backend `code`/`message`/`request_id`
  envelope) or `UnauthorizedError` after a failed refresh. Catch
  selectively; don't `try/catch` and swallow.

## Jest setup

- Preset is `jest-expo` (`mobile/jest.config.js`). The 25 % line
  coverage floor is enforced by `coverageThreshold.global.lines`.
- `mobile/jest.setup.ts` mocks `expo-secure-store` (in-process Map),
  `expo-constants`, and `expo-linking`. Reset state in `beforeEach`.
- Coverage scope: `src/**/*.{ts,tsx}` and `app/**/*.{ts,tsx}`. `__tests__/`
  files live alongside the app, but the matcher only runs files matching
  `*.test.ts(x)`.
- For pure logic (URL parsing, state machines), prefer plain Jest cases
  over RNTL renders. `@testing-library/react-native` is wired in
  `package.json`; reach for it when a screen has interactive behavior
  worth asserting.

## What runs where (preflight)

| Gate                     | Tool                            | Runtime needed |
|--------------------------|---------------------------------|----------------|
| `make lint:mobile`       | ESLint + `tsc --noEmit`         | Node only      |
| `make test:mobile`       | Jest (`jest-expo`)              | Node only      |
| `make test:mobile:coverage` | Jest + `--coverage` (25 % floor) | Node only      |
| `make dev:mobile`        | `npx expo start` (Metro + QR)   | Real device with Expo Go |
| `make build:mobile`      | `npx expo export`               | Node only      |
| EAS Build                | `eas build` (via Expo cloud)    | EAS account    |

`make preflight` runs the first three. The other rows are out-of-band.

## Adding a feature — checklist

1. Edit `backend/api/openapi.yaml` (or the Huma handler — code-first).
   Run `make gen:openapi`.
2. Add the matching DTO in `mobile/src/api/models.ts`.
3. Add the endpoint function in `mobile/src/api/endpoints.ts`.
4. Wire the auth-store method or screen-local hook that calls it.
5. Build the screen under `mobile/app/`. Use NativeWind classes.
6. Add a Jest spec under `mobile/__tests__/`.
7. Mirror the feature on web (`web/src/**`) — see `_shared/parity.md`.
8. `make preflight`.
