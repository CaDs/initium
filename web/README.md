# Initium — Web Frontend

Next.js 16 App Router frontend for the Initium full-stack starter template.

## Quickstart

From the repo root (starts backend + web together):
```bash
make dev
```

Or from this directory with the backend already running:
```bash
npm run dev   # http://localhost:3000
```

## Commands

| Command | Description |
|---------|-------------|
| `npm run dev` | Dev server on port 3000 |
| `npm run build` | Production build |
| `npm run test` | Vitest unit tests |
| `npm run lint` | ESLint |

## Auth Flow

1. Google Sign-In redirects to the backend `/api/auth/google`; the backend handles OAuth and sets httpOnly cookies, then redirects back to `/home`.
2. Magic link: user submits email, backend sends a link; clicking it hits `/api/auth/verify` which redirects to the backend to validate and set cookies.
3. `middleware.ts` guards `/home` — on missing access token it attempts a silent refresh, then redirects to `/login` on failure.

## Environment

Copy `.env.example` to `.env.local` and set:

```
API_URL=http://localhost:8000          # server-side backend base URL
NEXT_PUBLIC_API_URL=http://localhost:8000  # client-side (optional)
```

`API_URL` is required in production and will throw at startup if missing.

## Full-stack context

See the root [`README.md`](/Users/eridia/Projects/initium/README.md) (or `../README.md` from here) for backend, mobile, and infrastructure setup.
