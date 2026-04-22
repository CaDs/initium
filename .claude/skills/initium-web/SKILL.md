---
name: initium-web
description: Use when writing or modifying the Next.js web app in the Initium template — pages, Server Actions, Server Components, middleware, API route handlers, Zod schemas, i18n, or web tests. Triggers on paths under `web/src/**` or `web/messages/**`. Encodes the App Router + Server Actions + Zod conventions, cookie-based session model, and the contract-first workflow for this specific fork-and-specialize starter template.
---

# initium-web

You are editing the Next.js web app of an Initium fork. This template ships
auth (Google + magic link), session cookies, i18n (en/es/ja), theme (light/dark/
system), and a Zod-validated API client — all minimal, unstyled, ready to be
skinned by the fork.

> **This is NOT the Next.js you know.** App Router semantics, Server Actions,
> and `useActionState` have moved since Next 13. If a pattern you remember
> compiles but feels off, check `node_modules/next/dist/docs/` or the
> exemplars below before shipping.

## Gates that will fail your PR

Run `make preflight` before committing. It fails if any of the following
is true:

- A `/api/*` spec path has no consumer in this codebase
  (`make check:parity`).
- An exemplar path cited in this skill no longer contains its
  `<!-- expect: symbol -->` annotation (`make check:skills`).
- `npm run lint`, `npm run test`, or `npm run build` fails.
- `git status --porcelain` is non-empty after the run (`make check:staged`).

The Zod-at-boundary rule (`lib/schemas.ts`) is enforced at runtime, not in
CI — a page that forgets its Zod schema will 500 the first time the
backend returns an unexpected shape. The `apiFetch<T>()` wrapper demands
a Zod schema argument; do not reach around it.

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
  api.ts              apiFetch<T>() — server-side fetch wrapper: base URL,
                      cookie forwarding, cache: no-store, Zod validation.
  session.ts          httpOnly cookie helpers; hasSession(), getAccessToken().
  schemas.ts          Zod schemas for API response validation.
  api-types.ts        Generated from openapi.yaml — NEVER hand-edit.
  types.ts            Shared TypeScript types.
middleware.ts         Cookie-existence guard + refresh on access_token miss.
src/i18n/request.ts   next-intl setup; reads cookie locale.
messages/{en,es,ja}.json  Translations. Add keys to ALL three.
```

## Rules

- **Server Actions** for user mutations (forms). Never POST from client.
- **Route Handlers** (`app/api/`) only for system redirects like OAuth callbacks.
- Server Components check auth by calling the backend (`/api/me`). Never parse
  JWTs client-side.
- Zod validate at every API response boundary — use schemas in `lib/schemas.ts`.
- `middleware.ts` does cookie-presence fast path only. Real validation lives in
  Server Components.
- Cookie flags: `httpOnly` + `Secure` (prod) + `SameSite=Lax`. Set in `lib/session.ts`.
- TypeScript strict. No `any`. Use `unknown` + type guards.
- **Zod v4 is pinned.** Use `.issues` (not `.errors`), `z.email()` (not
  `z.string().email()`), `z.string().datetime()` for timestamp fields.
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

1. Someone edits `backend/api/openapi.yaml`.
2. Someone runs `make gen:openapi` → regenerates `web/src/lib/api-types.ts`.
3. Update `web/src/lib/schemas.ts` to add a Zod schema for the new response
   (or extend an existing one). Zod remains the runtime guard; generated
   types are the compile-time check. Also update `web/src/lib/types.ts` if
   you use a hand-written TypeScript alias alongside the generated one.
4. **List endpoints use envelope objects** (e.g. `NoteList { notes: Note[] }`).
   Zod schema and `apiFetch<T>` generic must match the envelope, not a bare
   array — see `patterns/api-fetch.md` for the pattern.
5. Server Components / Server Actions call the endpoint via `apiFetch<T>()`
   with the Zod schema.

Never hand-edit `api-types.ts`. Full workflow: `docs/OPENAPI.md`.

**Cross-stack completeness**: if your feature requires a new endpoint,
the backend handler + service + migration must exist before web code
calls it. A web-only PR that `apiFetch`es a nonexistent endpoint will
404 silently — `make check:parity` catches the spec side, but not the
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
- Mock `next-intl`'s `useTranslations`, Server Actions, and `sonner` (see
  `MagicLinkForm.test.tsx` for the pattern).
- `useActionState` needs a small shim — mock it to return the idle/initial state.
- No e2e tests at the template level. Forks can add Playwright or similar.

## Canonical exemplars (open these when unsure)

- Server Action: `web/src/actions/auth.ts` <!-- expect: requestMagicLink --> — form state, Zod validation, toast-friendly returns.
- API fetcher: `web/src/lib/api.ts` <!-- expect: apiFetch --> — Zod-validated response, ApiError object on failure.
- Session guard: `web/src/lib/session.ts` <!-- expect: hasSession -->, `web/src/middleware.ts` <!-- expect: PROTECTED_PATHS -->
- Protected Server Component: `web/src/app/home/page.tsx` <!-- expect: getTranslations -->
- Form component: `web/src/components/auth/MagicLinkForm.tsx` <!-- expect: useActionState --> — useActionState + toast + a11y.
- Unit test: `web/src/__tests__/MagicLinkForm.test.tsx` <!-- expect: MagicLinkForm -->

See also: `patterns/server-action.md`, `patterns/api-fetch.md`, `patterns/component.md`.

## Gotchas

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
