# Server Action pattern

Server Actions handle form mutations. They run on the server, validate input
with Zod, call the backend via `apiFetch`, and return a typed state consumable
by `useActionState`.

## Skeleton

```ts
// src/actions/orders.ts
"use server";

import { z } from "zod";
import { apiFetch } from "@/lib/api";
import { orderSchema } from "@/lib/schemas";

const createOrderSchema = z.object({
  email: z.string().email(),
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
    return { ok: false, message: parsed.error.errors[0]?.message ?? "Invalid input" };
  }

  const result = await apiFetch("/api/orders", {
    method: "POST",
    body: JSON.stringify(parsed.data),
  }, orderSchema);

  if (!result.ok) {
    return { ok: false, message: result.error };
  }
  return { ok: true, message: "Created" };
}
```

## Consumption

```tsx
"use client";
import { useActionState } from "react";
import { createOrder, type CreateOrderState } from "@/actions/orders";

const initial: CreateOrderState = { ok: false, message: "" };

export default function CreateOrderForm() {
  const [state, action, isPending] = useActionState(createOrder, initial);
  // ...form rendering + state.message feedback
}
```

## Rules

- Always export the `State` type alongside the action — the component imports it.
- Validate input with Zod at the top of the action.
- Route errors back through the state object; don't throw. `useActionState`
  surfaces the return value to the client.
- Never bypass `apiFetch` — it handles base URL, cookies, and Zod validation.
