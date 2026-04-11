# Frontend — Next.js (App Router, TypeScript, Tailwind)

## Build & Test

```bash
npm run dev       # Dev server (port 3000)
npm run build     # Production build
npm run test      # Vitest
npm run lint      # ESLint + tsc --noEmit
```

## Architecture

Uses Next.js framework-native patterns — no Clean Architecture overlay.

```
src/app/              # Routes (Server Components = data, Client Components = interaction)
src/actions/          # Server Actions for form mutations (magic link request, profile update)
src/components/       # Reusable UI components
src/lib/api.ts        # Server-side fetch wrapper (base URL, cookie forwarding, cache: no-store)
src/lib/session.ts    # Cookie helpers (httpOnly, Secure, SameSite=Lax)
src/lib/schemas.ts    # Zod schemas for API response validation
src/lib/types.ts      # Shared TypeScript types
middleware.ts         # Cookie-existence auth guard with explicit route matcher
```

## Key Rules

- **Server Actions** for user-initiated mutations (forms)
- **Route Handlers** (`app/api/`) for system redirects (OAuth callbacks only)
- Server Components validate auth by calling backend `GET /auth/me` — never parse JWTs client-side
- Zod validation at every API response boundary (`lib/schemas.ts`)
- TypeScript strict mode, no `any` — use `unknown` + type guards
- `middleware.ts` checks cookie existence only (fast path). Real validation happens in Server Components.
- Cookie flags: `httpOnly`, `Secure` (prod), `SameSite=Lax` — enforced in `lib/session.ts`

## Auth Flow

1. User clicks Google Sign-In → redirects to backend `/auth/google`
2. Backend handles OAuth, sets httpOnly cookies, redirects to `/home`
3. `middleware.ts` checks cookie exists → allows through
4. Server Component in `home/page.tsx` calls `/auth/me` → if invalid, redirect to `/login`

## Gotchas

- `middleware.ts` fast-path may briefly show page shell before Server Component redirect fires. Server Components that fetch user data must treat failed `/auth/me` as a redirect, not an error boundary.
- OAuth callback state parameter validated server-side in `app/api/auth/callback/route.ts`
- Environment variables validated at build time in `next.config.ts`
- Backend API base URL comes from `NEXT_PUBLIC_API_URL` (client) or `API_URL` (server)
