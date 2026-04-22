# apiFetch pattern

`apiFetch<T>()` in `src/lib/api.ts` is the only way to call the backend from the
web app. It handles base URL, cookie forwarding, `cache: no-store`, and Zod
validation of the response.

## Skeleton

```ts
import { apiFetch } from "@/lib/api";
import { userSchema } from "@/lib/schemas";
import type { User } from "@/lib/types";

const result = await apiFetch<User>("/api/me", {}, userSchema);

if (!result.ok) {
  // result.error is an ApiError object: { code, message, request_id? }.
  // For auth-guarded fetches, redirect on UNAUTHORIZED; surface other codes
  // via an error boundary or inline error UI rather than bouncing to /login.
  if (result.error.code === "UNAUTHORIZED") redirect("/login");
  throw new Error(result.error.message);
}

// result.data is User (validated)
```

## List endpoints use envelopes

Initium list endpoints wrap arrays in an envelope object (see `NoteList` /
`RouteList` in `backend/api/openapi.yaml`). The Zod schema and `apiFetch<T>`
type must match the envelope, not a bare array:

```ts
// correct
const listSchema = z.object({ notes: z.array(noteSchema) });
const result = await apiFetch<{ notes: Note[] }>("/api/notes", {}, listSchema);

// wrong — will Zod-fail at runtime and land you on /login forever
const listSchema = z.array(noteSchema);
```

## Rules

- Every response must pass through a Zod schema. The generated
  `api-types.ts` gives you compile-time shape; Zod gives runtime safety.
- Treat `{ ok: false }` as a redirect condition for auth-guarded pages,
  not an error boundary throw.
- `apiFetch` automatically forwards cookies. Don't re-implement.
- Don't call `fetch` directly from Server Components — you'll skip the
  Zod gate and lose cookie forwarding.

## When the API shape changes

1. Backend: edit `openapi.yaml`, run `make gen:openapi`.
2. Web: `api-types.ts` regenerates. TypeScript will flag call sites.
3. Update the corresponding Zod schema in `lib/schemas.ts` — add/rename
   fields to match.
4. CI's contract test + `make check:openapi` enforce server-side shape.
   Web-side, the Zod failure at runtime is your drift detector.
