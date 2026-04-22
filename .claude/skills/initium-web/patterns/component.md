# Component pattern

Components live in `src/components/`. Default to Server Components; mark
`"use client"` only when the component needs state, effects, or event handlers.

## Server Component

```tsx
// src/components/shared/Example.tsx
import { getTranslations } from "next-intl/server";

export default async function Example() {
  const t = await getTranslations("example");
  return <p>{t("greeting")}</p>;
}
```

## Client Component

```tsx
"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";

export default function Counter() {
  const t = useTranslations("counter");
  const [n, setN] = useState(0);
  return (
    <button type="button" onClick={() => setN(n + 1)} aria-label={t("increment")}>
      {t("label", { count: n })}
    </button>
  );
}
```

## Rules

- No extra abstractions. Use plain JSX + Tailwind semantic tokens.
- i18n at every user-visible string. Add the key to all three locale files.
- a11y: `aria-label` on non-text buttons, `<label>` on inputs, visible focus
  rings (global via `globals.css`).
- Don't reach for component libraries (shadcn, Radix, etc.). The template's
  house style is minimal raw HTML + Tailwind. Forks add what they need.
- Component state stays in the component. No global state library (Zustand,
  Redux). Use URL params or Server Actions for cross-component data.
