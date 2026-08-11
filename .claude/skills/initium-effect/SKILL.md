---
name: initium-effect
description: Use when modifying packages/core or TypeScript API contracts, Effect schemas, tagged failures, or shared request behavior in Initium.
---

# initium-effect

The shared package is the framework-neutral TypeScript boundary between the
Go API and the Next.js/Expo consumers. It uses stable Effect v3 core. The Go
Huma handlers and their generated OpenAPI document remain canonical.

## Rules

- Decode every external value with `Schema`; never cast JSON to a DTO.
- Model expected failures in the Effect error channel with tagged data errors.
- `Effect.runPromise` belongs only at Next.js, Expo, React, or test boundaries.
- Shared code cannot import React, Next.js, Expo, React Native, Node APIs, or
  `@effect/platform`.
- Keep endpoint path literals in web/mobile adapters so `make check/parity`
  continues to prove actual client consumption.
- Never hand-edit `packages/core/src/generated/api-types.ts`; run
  `make gen/openapi`.
- Add each wire schema to the conformance assertions in
  `packages/core/src/conformance.ts` <!-- expect: ContractsMatchGeneratedOpenAPI -->.

## Canonical exemplars

- Contracts: `packages/core/src/contracts.ts` <!-- expect: UserSchema -->
- Tagged failures: `packages/core/src/errors.ts` <!-- expect: HttpError -->
- Interruptible fetch: `packages/core/src/http.ts` <!-- expect: requestJson -->
- Tests: `packages/core/__tests__/contracts.test.ts` <!-- expect: shared contracts -->

## Gates

Run `make lint/core`, `make test/core/coverage`, `make gen/openapi`, and finally
`make preflight`. Generated drift and schema/type disagreement must fail before
an API change is considered complete.
