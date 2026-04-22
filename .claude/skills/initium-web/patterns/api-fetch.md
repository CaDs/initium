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
  // handle error — result.error is string
  redirect("/login");
}

// result.data is User (validated)
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
