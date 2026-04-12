import { redirect } from "next/navigation";
import { useTranslations } from "next-intl";
import { apiFetch } from "@/lib/api";
import { userSchema } from "@/lib/schemas";
import type { User } from "@/lib/types";

export default async function HomePage() {
  const result = await apiFetch<User>("/api/me", {}, userSchema);

  if (!result.ok) {
    redirect("/login");
  }

  return <HomeContent user={result.data} />;
}

function HomeContent({ user }: { user: User }) {
  const t = useTranslations("home");

  return (
    <div className="max-w-2xl mx-auto px-6 py-12">
      <h1 className="text-3xl font-bold mb-2">
        {user.name ? t("welcomeUser", { name: user.name }) : t("welcome")}
      </h1>
      <p className="text-muted mb-8">
        {t("subtitle")}
      </p>

      <div className="bg-card border border-border rounded-lg p-6 space-y-3">
        <h2 className="text-lg font-semibold">{t("profile")}</h2>
        <dl className="grid grid-cols-[100px_1fr] gap-2 text-sm">
          <dt className="text-muted">{t("email")}</dt>
          <dd>{user.email}</dd>
          <dt className="text-muted">{t("name")}</dt>
          <dd>{user.name || "—"}</dd>
          <dt className="text-muted">{t("userId")}</dt>
          <dd className="font-mono text-xs break-all">{user.id}</dd>
        </dl>
      </div>
    </div>
  );
}
