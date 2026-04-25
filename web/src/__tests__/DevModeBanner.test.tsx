import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { NextIntlClientProvider } from "next-intl";

import DevModeBanner from "@/components/shared/DevModeBanner";

// DevModeBanner is gated on DEV_BYPASS_AUTH=true. It's deliberately a
// server-rendered banner so a misconfigured production deploy can never
// silently keep dev-bypass enabled — the banner is visually loud.

const messages = {
  devBanner: { message: "Dev mode: auth bypassed" },
};

beforeEach(() => {
  delete process.env.DEV_BYPASS_AUTH;
  vi.resetModules();
});
afterEach(() => {
  delete process.env.DEV_BYPASS_AUTH;
});

const renderBanner = async () => {
  const Component = (await import("@/components/shared/DevModeBanner")).default;
  return render(
    <NextIntlClientProvider locale="en" messages={messages}>
      {/* @ts-expect-error — Server Component rendered in test */}
      <Component />
    </NextIntlClientProvider>,
  );
};

describe("DevModeBanner", () => {
  it("renders nothing when DEV_BYPASS_AUTH is unset", async () => {
    const { container } = await renderBanner();
    expect(container).toBeEmptyDOMElement();
  });

  it("renders nothing when DEV_BYPASS_AUTH is something other than 'true'", async () => {
    process.env.DEV_BYPASS_AUTH = "false";
    const { container } = await renderBanner();
    expect(container).toBeEmptyDOMElement();
  });

  it("renders the banner when DEV_BYPASS_AUTH is 'true'", async () => {
    process.env.DEV_BYPASS_AUTH = "true";
    await renderBanner();
    expect(screen.getByText("Dev mode: auth bypassed")).toBeInTheDocument();
  });
});

// Suppress the unused-import warning if `DevModeBanner` ever becomes used
// statically here — we import dynamically inside renderBanner so resetModules
// can re-evaluate the env var.
void DevModeBanner;
