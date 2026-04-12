# Initium

Opinionated POC starter template. Fork and specialize per project.

## Stack

- **Backend**: Go (chi + GORM + PostgreSQL)
- **Frontend**: Next.js (App Router, TypeScript, Tailwind)
- **Mobile**: Flutter (Dart, Riverpod, Dio)

## Architecture

Ports & Adapters (hexagonal) at infrastructure boundaries. Four layers per component:

| Layer | Rule |
|-------|------|
| `domain/` | Zero framework imports. Pure entities, interfaces, errors. |
| `service/` (backend) or `usecase/` (mobile) | Imports domain only. Business logic. |
| `adapter/` | Handlers, persistence, middleware, DTOs. Imports domain + service. |
| `infra/` or framework layer | Config, DB, external services. Outermost ring. |

**Dependency rule**: inner layers never import outer layers. Violations are bugs.

## Build & Run

```bash
make setup        # First-time: infra, deps, .env, migrations, JWT keys
make dev          # Backend (8000) + web (3000) concurrently
make backend-test # Go tests with race detector
make web-test     # Vitest
make mobile-test  # Flutter tests
make mobile-gen   # Required after DTO changes (build_runner)
```

## Auth Model

- Backend owns session state (single source of truth)
- Short-lived access tokens (15min) + refresh tokens (7d) in sessions table
- Google OAuth (web: server-side flow, mobile: ID token verification via `/auth/mobile/google`)
- Magic links (single-use, stored with hash in `magic_link_tokens` table)
- `DEV_BYPASS_AUTH=true` (dev only): skips auth, injects test user `dev@initium.local`

## API Contract

`backend/api/openapi.yaml` is the canonical spec. When changing API responses:
1. Update `openapi.yaml` first
2. Update backend Go structs
3. Update `web/src/lib/schemas.ts` (Zod)
4. Update `mobile/lib/data/remote/dto/` + run `make mobile-gen`

## Internationalization (i18n)

Supported locales: **en**, **es**, **ja**. All user-facing strings must be localized.

- **Web**: `next-intl` — translations in `web/messages/{en,es,ja}.json`. Use `useTranslations('namespace')` in components.
- **Mobile**: Flutter `intl` + ARB files — translations in `mobile/lib/l10n/app_{en,es,ja}.arb`. Run `flutter gen-l10n` after changes. Use `AppLocalizations.of(context)!`.
- **Backend**: API responses return data (not UI strings). The backend does NOT localize — clients handle display. If a future feature needs server-side localization (e.g., email templates), use Go `embed` with per-locale template directories.

When adding new UI text: add the key to ALL locale files (en, es, ja) before using it.

## Theme Support

Three modes: **light**, **dark**, **system default**.

- **Web**: CSS variables in `globals.css` swap via `.dark` class. Tailwind uses semantic color names (`bg-card`, `text-muted`, `border-border`). Theme toggle persisted in `localStorage`.
- **Mobile**: `ThemeData` with `ColorScheme.fromSeed()` for both `theme` and `darkTheme`. `ThemeMode.system` by default. Use `Theme.of(context).colorScheme` — never hardcode colors.

## Accessibility

- **Web**: Skip-to-main link, `aria-label` on all interactive elements, `aria-live` on dynamic content, `role` attributes, visible focus indicators via `:focus-visible`, semantic HTML (`<nav>`, `<main>`, `<dl>`), form labels via `<label htmlFor>`.
- **Mobile**: `Semantics` widgets on interactive elements, `semanticLabel` on icons/images, `tooltip` on `IconButton`, `autofillHints` on text fields, `liveRegion` on status messages.

## Conventions

- Conventional Commits: `feat:`, `fix:`, `test:`, `refactor:`, `docs:`, `chore:`
- No secrets in version control. Use `.env.example` templates.
- Parameterized queries only. No string interpolation for SQL.
- Run linters and tests after code changes automatically.
- Never push or create PRs without explicit approval.

## Common Gotchas

- GORM tags live ONLY in `adapter/persistence/models.go`, never in `domain/`
- Next.js Server Components validate auth via `/auth/me` call (not cookie parsing)
- Flutter: `dart run build_runner build` required after any `@JsonSerializable` or `@freezed` change
- Flutter: iOS keychain items persist across app uninstall — `token_storage.dart` handles this
- `DEV_BYPASS_AUTH` is guarded: fails fast at startup if enabled outside `APP_ENV=development`
