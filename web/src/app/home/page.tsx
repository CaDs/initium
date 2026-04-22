import { redirect } from "next/navigation";
import { getTranslations } from "next-intl/server";
import { apiFetch } from "@/lib/api";
import { userSchema } from "@/lib/schemas";
import type { User } from "@/lib/types";

export default async function HomePage() {
  const result = await apiFetch<User>("/api/me", {}, userSchema);

  if (!result.ok) {
    redirect("/login");
  }

  const t = await getTranslations("nav");
  const user = result.data;

  return (
    <div className="mx-auto w-full max-w-2xl px-6 py-10 space-y-6">
      <h1 className="text-2xl font-semibold">Home</h1>
      <dl className="rounded-lg border border-border bg-card p-6 grid grid-cols-[100px_1fr] gap-y-2 text-sm">
        <dt className="text-muted">Email</dt>
        <dd>{user.email}</dd>
        <dt className="text-muted">Name</dt>
        <dd>{user.name || "—"}</dd>
        <dt className="text-muted">Role</dt>
        <dd>{user.role}</dd>
        <dt className="text-muted">ID</dt>
        <dd className="font-mono text-xs break-all">{user.id}</dd>
      </dl>
      <form action="/api/auth/logout" method="POST">
        <button
          type="submit"
          className="inline-flex items-center justify-center rounded-md border border-border px-4 py-2 text-sm font-medium hover:bg-foreground/5 transition-colors"
        >
          {t("logout")}
        </button>
      </form>
    </div>
  );
}
