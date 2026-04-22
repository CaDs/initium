# Server Action pattern

Server Actions handle form mutations. They run on the server, validate input
with Zod, call the backend via `apiFetch`, and return a typed state consumable
by `useActionState`.

## Skeleton

```ts
// src/actions/orders.ts
"use server";

import { z } from "zod";
import { revalidatePath } from "next/cache";
import { apiFetch } from "@/lib/api";
import { orderSchema } from "@/lib/schemas";

const createOrderSchema = z.object({
  email: z.email(),
});

export type CreateOrderState = { ok: boolean; message: string };

export async function createOrder(
  _prev: CreateOrderState,
  formData: FormData,
): Promise<CreateOrderState> {
  const parsed = createOrderSchema.safeParse({
    email: formData.get("email"),
  });
  if (!parsed.success) {
    return { ok: false, message: parsed.error.issues[0]?.message ?? "Invalid input" };
  }

  const result = await apiFetch("/api/orders", {
    method: "POST",
    body: JSON.stringify(parsed.data),
  }, orderSchema);

  if (!result.ok) {
    return { ok: false, message: result.error.message };
  }

  // If a Server Component on the same route renders the post-mutation data,
  // invalidate its cache so a fresh fetch runs on the next render.
  revalidatePath("/orders");
  return { ok: true, message: "Created" };
}
```

## Consumption

```tsx
"use client";
import { useActionState, useEffect, useRef } from "react";
import { createOrder, type CreateOrderState } from "@/actions/orders";

const initial: CreateOrderState = { ok: false, message: "" };

export default function CreateOrderForm() {
  const [state, action, isPending] = useActionState(createOrder, initial);
  const formRef = useRef<HTMLFormElement>(null);

  // Reset the form after a successful mutation so the input doesn't linger.
  useEffect(() => {
    if (state.ok) formRef.current?.reset();
  }, [state.ok]);

  return (
    <form ref={formRef} action={action}>
      {/* ...inputs + submit; surface state.message for feedback */}
    </form>
  );
}
```

## Rules

- **Zod v4 is pinned.** Use `parsed.error.issues` (not `.errors` — renamed in
  v4), `z.email()` (not `z.string().email()` — deprecated), and
  `z.string().datetime()` for timestamp validation.
- **`apiFetch` returns `{ ok: false, error: ApiError }`** on failure. `error`
  is an object `{ code, message, request_id? }` — not a string. Use
  `result.error.message` when surfacing to the user.
- Always export the `State` type alongside the action — the component imports it.
- Validate input with Zod at the top of the action.
- Route errors back through the state object; don't throw. `useActionState`
  surfaces the return value to the client.
- Never bypass `apiFetch` — it handles base URL, cookies, and Zod validation.
- After a mutation whose result is rendered by a Server Component on the same
  route, call `revalidatePath("/route")` before returning success. Without it,
  the cached RSC payload serves stale data until a full reload.
- For CRUD forms that stay mounted after success, reset the form element via
  `useRef` + `useEffect` (shown above). The auth flow exemplar
  (`MagicLinkForm.tsx`) swaps the form for a success message so reset isn't
  needed — that's a different pattern.
