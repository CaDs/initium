"use client";

import { useSyncExternalStore } from "react";

type Theme = "light" | "dark" | "system";

const icons: Record<Theme, string> = {
  light: "☀",
  dark: "☾",
  system: "◐",
};

const nextTheme: Record<Theme, Theme> = {
  light: "dark",
  dark: "system",
  system: "light",
};

// Storage events don't fire in the tab that wrote the value — only in
// other tabs. `switchTheme` dispatches a synthetic event so our own
// useSyncExternalStore subscriber re-reads after a user-initiated change.
const THEME_CHANGE_EVENT = "theme-change";

function subscribeTheme(callback: () => void): () => void {
  const onStorage = (e: StorageEvent) => {
    if (e.key === "theme" || e.key === null) callback();
  };
  window.addEventListener("storage", onStorage);
  window.addEventListener(THEME_CHANGE_EVENT, callback);
  return () => {
    window.removeEventListener("storage", onStorage);
    window.removeEventListener(THEME_CHANGE_EVENT, callback);
  };
}

function getThemeSnapshot(): Theme {
  const stored = localStorage.getItem("theme") as Theme | null;
  return stored ?? "system";
}

// React calls this during SSR and again during hydration. Returning the
// same "system" default on both sides avoids a hydration mismatch; after
// hydration completes React naturally transitions to getThemeSnapshot.
function getServerThemeSnapshot(): Theme {
  return "system";
}

function applyTheme(t: Theme) {
  const root = document.documentElement;
  if (t === "system") {
    root.removeAttribute("data-theme");
    root.classList.remove("dark");
    if (window.matchMedia("(prefers-color-scheme: dark)").matches) {
      root.classList.add("dark");
    }
  } else if (t === "dark") {
    root.classList.add("dark");
    root.setAttribute("data-theme", "dark");
  } else {
    root.classList.remove("dark");
    root.setAttribute("data-theme", "light");
  }
}

function switchTheme(t: Theme) {
  localStorage.setItem("theme", t);
  applyTheme(t);
  window.dispatchEvent(new Event(THEME_CHANGE_EVENT));
}

export default function ThemeSwitcher() {
  const theme = useSyncExternalStore(
    subscribeTheme,
    getThemeSnapshot,
    getServerThemeSnapshot,
  );

  return (
    <button
      onClick={() => switchTheme(nextTheme[theme])}
      className="w-8 h-8 flex items-center justify-center rounded-lg text-neutral-500 hover:text-neutral-900 hover:bg-neutral-100 dark:text-neutral-400 dark:hover:text-white dark:hover:bg-neutral-800 transition-colors"
      aria-label={`Switch theme (current: ${theme})`}
      title={`Theme: ${theme}`}
    >
      <span className="text-sm">{icons[theme]}</span>
    </button>
  );
}
