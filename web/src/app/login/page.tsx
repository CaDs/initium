import { redirect } from "next/navigation";
import { useTranslations } from "next-intl";
import GoogleSignInButton from "@/components/auth/GoogleSignInButton";
import MagicLinkForm from "@/components/auth/MagicLinkForm";
import { hasSession } from "@/lib/session";

export default async function LoginPage() {
  if (await hasSession()) {
    redirect("/home");
  }

  return <LoginContent />;
}

function LoginContent() {
  const t = useTranslations("login");

  return (
    <div className="flex flex-col items-center justify-center min-h-[80vh] px-6">
      <div className="w-full max-w-sm space-y-6">
        <div className="text-center">
          <h1 className="text-2xl font-bold">{t("title")}</h1>
          <p className="text-muted mt-1">{t("subtitle")}</p>
        </div>

        <GoogleSignInButton />

        <div className="relative" role="separator">
          <div className="absolute inset-0 flex items-center">
            <div className="w-full border-t border-border" />
          </div>
          <div className="relative flex justify-center text-sm">
            <span className="bg-background px-4 text-muted">{t("divider")}</span>
          </div>
        </div>

        <MagicLinkForm />
      </div>
    </div>
  );
}
