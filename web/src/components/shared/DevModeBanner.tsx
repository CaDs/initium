import { useTranslations } from "next-intl";

export default function DevModeBanner() {
  if (process.env.NEXT_PUBLIC_DEV_BYPASS_AUTH !== "true") return null;

  return <DevBannerContent />;
}

function DevBannerContent() {
  const t = useTranslations("devBanner");

  return (
    <div
      className="bg-warning-bg border-b border-warning-text/20 px-4 py-2 text-center text-sm text-warning-text"
      role="status"
    >
      {t("message")}
    </div>
  );
}
