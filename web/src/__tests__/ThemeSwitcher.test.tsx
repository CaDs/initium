import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import ThemeSwitcher from "@/components/shared/ThemeSwitcher";

// happy-dom's localStorage is file-based and ships without clear/removeItem
// by default, so we install a plain in-memory shim for these tests.
const storage = new Map<string, string>();
vi.stubGlobal("localStorage", {
  getItem: (k: string) => storage.get(k) ?? null,
  setItem: (k: string, v: string) => storage.set(k, v),
  removeItem: (k: string) => storage.delete(k),
  clear: () => storage.clear(),
  key: (i: number) => [...storage.keys()][i] ?? null,
  get length() {
    return storage.size;
  },
});

// Regression guard for the hydration-mismatch bug that shipped with the
// original ThemeSwitcher: reading localStorage in useState's initializer
// produced "dark" on the client and "system" on the server, crashing React
// with a hydration mismatch on first render.
//
// These tests don't exercise real SSR (Vitest is jsdom), but they do assert
// that the useSyncExternalStore server snapshot returns "system" regardless
// of what localStorage contains — which is the contract React relies on to
// keep server and initial-client renders aligned.
describe("ThemeSwitcher initial render", () => {
  beforeEach(() => {
    localStorage.removeItem("theme");
  });

  it("defaults to 'system' when localStorage has no theme (matches server snapshot)", () => {
    render(<ThemeSwitcher />);
    expect(
      screen.getByRole("button", { name: "Switch theme (current: system)" }),
    ).toBeInTheDocument();
  });

  it("reflects stored theme on the client after mount", () => {
    localStorage.setItem("theme", "light");
    render(<ThemeSwitcher />);
    expect(
      screen.getByRole("button", { name: "Switch theme (current: light)" }),
    ).toBeInTheDocument();
  });
});
