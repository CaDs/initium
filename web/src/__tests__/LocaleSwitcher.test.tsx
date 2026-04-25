import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";

import LocaleSwitcher from "@/components/shared/LocaleSwitcher";

// LocaleSwitcher writes a `locale=<code>` cookie + calls router.refresh().
// Both must happen on selection or i18n stays on the previous locale.

const refreshSpy = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({ refresh: refreshSpy }),
}));

beforeEach(() => {
  refreshSpy.mockClear();
  // happy-dom's document.cookie is read/write but we want a clean slate.
  document.cookie = "locale=; path=/; max-age=0";
});
afterEach(() => {
  document.cookie = "locale=; path=/; max-age=0";
});

describe("LocaleSwitcher", () => {
  it("renders the current locale as the selected option", () => {
    render(<LocaleSwitcher current="es" />);

    const select = screen.getByRole("combobox", { name: /language/i });
    expect((select as HTMLSelectElement).value).toBe("es");
  });

  it("writes a locale cookie and calls router.refresh on change", () => {
    render(<LocaleSwitcher current="en" />);

    const select = screen.getByRole("combobox", { name: /language/i });
    fireEvent.change(select, { target: { value: "ja" } });

    expect(document.cookie).toContain("locale=ja");
    expect(refreshSpy).toHaveBeenCalledTimes(1);
  });
});
