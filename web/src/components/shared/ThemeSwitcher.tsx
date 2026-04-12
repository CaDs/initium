"use client";

import { useEffect, useState } from "react";

type Theme = "light" | "dark" | "system";

export default function ThemeSwitcher() {
  const [theme, setTheme] = useState<Theme>("system");

  useEffect(() => {
    const saved = localStorage.getItem("theme") as Theme | null;
    if (saved) {
      setTheme(saved);
      applyTheme(saved);
    }
  }, []);

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
    setTheme(t);
    localStorage.setItem("theme", t);
    applyTheme(t);
  }

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
