---
name: initium-web
description: Use when writing or modifying the Next.js web app in the Initium template — pages, Server Actions, Server Components, middleware, API route handlers, Effect schemas, i18n, or web tests. Triggers on paths under `web/src/**` or `web/messages/**`.
---

# initium-web

You are editing the Next.js web app of an Initium fork. This template ships
auth (Google + magic link), session cookies, i18n (en/es/ja), theme (light/dark/
system), and an Effect Schema-validated API client — all minimal, unstyled, ready to be
skinned by the fork.

> **This is NOT the Next.js you know.** App Router semantics, Server Actions,
> and `useActionState` have moved since Next 13. If a pattern you remember
> compiles but feels off, check `node_modules/next/dist/docs/` or the
> exemplars below before shipping.

## Gates that will fail your PR

Run `make preflight` before committing. It fails if any of the following
is true:

- A `/api/*` spec path has no consumer in this codebase
  (`make check/parity`).
- An exemplar path cited in this skill no longer contains its
  `<!-- expect: symbol -->` annotation (`make check/skills`).
- `make lint/web`, `make test/web/coverage`, or `make build/web` fails.
- Work remains untracked or unstaged after the run (`make check/staged`).

Web is an npm workspace of the repo root. Install dependencies from the root
(`npm install` / `make setup`); the root `package-lock.json` is the only
lockfile — a per-package lockfile under `web/` must not come back.

The Effect Schema boundary is shared in `packages/core`. `apiFetch()` requires
a schema and maps typed Effect failures to the UI result envelope; do not reach
around it or cast JSON.

## Architecture (framework-native, no extra abstractions)

```
src/app/              Routes. Server Components for data, Client Components for interaction.
  layout.tsx          Root layout: NextIntlClientProvider, DevModeBanner, Nav, Toaster.
  page.tsx            Session-aware redirect to /home or /login.
  login/page.tsx      Plain auth form (Google link + MagicLinkForm).
  home/page.tsx       Protected; fetches /api/me via apiFetch; on failure redirects.
  api/auth/*/route.ts Route handlers for OAuth callbacks + logout.
src/actions/          Server Actions for form mutations (e.g., requestMagicLink).
src/components/       Reusable components (shared/ + auth/).
src/lib/
  api.ts              apiFetch() — server-side Effect boundary: base URL,
                      cookie forwarding and shared Schema decoding.
  session.ts          httpOnly cookie helpers; hasSession(), getAccessToken().
middleware.ts         Cookie-existence guard + refresh on access_token miss.
src/i18n/request.ts   next-intl setup; reads cookie locale.
messages/{en,es,ja}.json  Translations. Add keys to ALL three.
```

## Rules

- **Server Actions** for user mutations (forms). Never POST from client.
- **Route Handlers** (`app/api/`) only for system redirects like OAuth callbacks.
- Server Components check auth by calling the backend (`/api/me`). Never parse
  JWTs client-side.
- Effect Schema-decode every API response using `@initium/core` contracts.
- `middleware.ts` does cookie-presence fast path only. Real validation lives in
  Server Components.
- Cookie flags: `httpOnly` + `Secure` (prod) + `SameSite=Lax`. Set in `lib/session.ts`.
- TypeScript strict. No `any`. Use `unknown` + type guards.
- Run Effects at Next.js boundaries only; shared helpers return Effects and
  rendering stays framework-native.
- Use semantic Tailwind tokens: `bg-background`, `text-foreground`, `bg-card`,
  `text-muted`, `border-border`, `bg-accent`, `bg-accent-foreground`,
  `text-error`. Never hardcoded `text-gray-600`.
- **When adding a new authenticated route**: update `middleware.ts`
  `PROTECTED_PATHS` + the `config.matcher`, AND add a nav link in
  `components/shared/Nav.tsx` with a `nav.*` translation key in all three
  locale files. Without the nav link, users can't reach the page without
  typing the URL.
- **Staging new files in a PR**: `git add -A` catches new files in `actions/`,
  `components/`, ARB keys, and schemas. A stale `git diff` that misses
  untracked files is the most common "forgot half the feature" failure mode.

## The contract-first workflow

When an API change lands on the backend:

1. Edit the typed Huma handler; generated OpenAPI is never hand-edited.
2. Run `make gen/openapi` to regenerate OpenAPI and shared TypeScript types.
3. Add or update the Effect Schema and conformance assertion in `packages/core`.
4. **List endpoints use envelope objects** (e.g. `NoteList { notes: Note[] }`).
   Effect Schema and `apiFetch()` call must match the envelope, not a bare array.
5. Server Components / Server Actions call the endpoint through `apiFetch()`.

Never hand-edit `packages/core/src/generated/api-types.ts`. Full workflow: `docs/OPENAPI.md`.

**Cross-stack completeness**: if your feature requires a new endpoint,
the backend handler + service + migration must exist before web code
calls it. A web-only PR that `apiFetch`es a nonexistent endpoint will
404 silently — `make check/parity` catches the spec side, but not the
handler side. Pair with the backend change, or explicitly defer.

## Auth flow (browser)

1. User clicks Google or submits magic-link email.
2. Google: `<a href="${API_URL}/api/auth/google">` — the backend handles the
   full OAuth dance and redirects to `/home` with cookies set.
3. Magic link: Server Action `requestMagicLink` POSTs to backend; user clicks
   email link; backend sets cookies and redirects to `/home`.
4. `middleware.ts` allows through on cookie presence, attempts refresh on
   access_token miss, redirects to `/login` on failure.
5. Server Component in `home/page.tsx` calls `/api/me`; on failure, redirects.

## i18n

- **Server Components**: `await getTranslations('namespace')` from
  `next-intl/server`. Async. This is what `Nav.tsx` and `home/page.tsx` use.
- **Client Components**: `useTranslations('namespace')` from `next-intl`.
  Sync. Works via `NextIntlClientProvider` in the root layout.
- Locale stored in `locale` cookie; switched via `LocaleSwitcher`.
- **Add new keys to ALL three locale files (en, es, ja) before using.**
  Otherwise missing-key hydration warnings on non-en locales.

## Accessibility baseline

- Skip-to-main link in root layout.
- `aria-label` on any button/link without visible text.
- `aria-live="polite"` on dynamic status (form success/error).
- `aria-invalid` + `aria-describedby` on form fields with errors.
- `<label>` associated with every form input (visible or `sr-only`).

## Testing

- Vitest + `@testing-library/react`. Tests in `src/__tests__/`.
- Phased coverage gate: 25% lines/branches (current 75.97%/66.21%);
  ramps to 80% in follow-up PRs. Run with `make test/web/coverage`.
- **MSW v2 intercepts every backend call.** Default handlers in
  `src/__tests__/msw/handlers.ts` cover the 9 backend endpoints; per-test
  overrides go through `server.use(http.post(...))`. Setup hooks in
  `src/__tests__/setup.ts` (`server.listen`, `resetHandlers`, `server.close`)
  fail tests that hit unmatched URLs — preventing accidental real-API calls.
- Middleware tests construct `NextRequest` directly + call the exported
  `middleware()` function. The DEV_BYPASS const snapshots env at module
  load, so tests that flip it must `vi.resetModules()` then re-import.
  See `__tests__/middleware.test.ts` <!-- expect: middleware -->.
- ESLint test-only rules (vitest + testing-library plugins) are scoped
  to `src/__tests__/**` in `eslint.config.mjs`. Production code stays
  out of test-rule territory.
- Mock `next-intl`'s `useTranslations`, Server Actions, and `sonner` (see
  `MagicLinkForm.test.tsx` for the pattern).
- `useActionState` needs a small shim — mock it to return the idle/initial state.
- No e2e tests at the template level. Forks can add Playwright or similar.

## Canonical exemplars (open these when unsure)

- Server Action: `web/src/actions/auth.ts` <!-- expect: requestMagicLink --> — form state, Effect validation, toast-friendly returns.
- API fetcher: `web/src/lib/api.ts` <!-- expect: apiFetch --> — Effect-decoded response and typed failure mapping.
- Session guard: `web/src/lib/session.ts` <!-- expect: hasSession -->, `web/src/middleware.ts` <!-- expect: PROTECTED_PATHS -->
- Protected Server Component: `web/src/app/home/page.tsx` <!-- expect: getTranslations -->
- Form component: `web/src/components/auth/MagicLinkForm.tsx` <!-- expect: useActionState --> — useActionState + toast + a11y.
- Unit test: `web/src/__tests__/MagicLinkForm.test.tsx` <!-- expect: MagicLinkForm -->
- Middleware test (cookie + refresh paths): `web/src/__tests__/middleware.test.ts` <!-- expect: middleware -->
- MSW handlers: `web/src/__tests__/msw/handlers.ts` <!-- expect: defaultHandlers -->
- apiFetch test (Bearer + Schema + non-JSON fallback): `web/src/__tests__/lib/api.test.ts` <!-- expect: apiFetch -->

## Gotchas

- **Hydration mismatches**: any Client Component that reads `localStorage`,
  `document`, `window.matchMedia`, or `Date.now()` in a `useState`
  initializer will render differently on server vs client and trigger a
  "Hydration failed" error. For anything sourced from an external store
  that changes per-user or per-device, use `useSyncExternalStore` with a
  stable `getServerSnapshot` (see `ThemeSwitcher.tsx`). React 19's
  `react-hooks/set-state-in-effect` ESLint rule also bans the "read in
  useEffect and setState" fallback — `useSyncExternalStore` is the
  idiomatic answer.
- **Inline scripts in the root layout**: raw `<script
  dangerouslySetInnerHTML>` in JSX triggers a Next 15 warning ("Scripts
  inside React components are never executed when rendering on the
  client"). Use `<Script strategy="beforeInteractive">` from `next/script`
  instead — see `app/layout.tsx` theme bootstrap. `beforeInteractive`
  runs before hydration, so FOUC prevention still works.
- **Middleware fast-path flash**: `middleware.ts` checks cookie presence only
  (fast). Server Components that fetch user data must treat a failed
  `/api/me` as `redirect("/login")`, not an error boundary — otherwise the
  page shell briefly shows before the redirect fires.
- **OAuth callback is backend-only**: the browser calls `/api/auth/google`
  on the backend; the backend handles the full OAuth handshake + state
  validation and sets cookies on the final redirect. **There is no
  `/api/auth/google/callback` Route Handler in `web/src/app/api/`** — don't
  add one.
- **`DEV_BYPASS_AUTH` is server-only.** No `NEXT_PUBLIC_` prefix. Read it
  only in middleware / Server Components. Release builds must never have
  it true — it's gated in `backend/internal/infra/config/`.
- **Base URL split**: `NEXT_PUBLIC_API_URL` for client components,
  `API_URL` for server components. `lib/api.ts` picks the right one.
  Production builds require one or the other — `npm run build` fails
  silently during static page generation otherwise.

## Parity

See [parity.md](../_shared/parity.md). If you add a screen here, the mirror
screen belongs on mobile. Call it out in the PR.
