import { useCallback, useEffect, useState } from "react";

export type PublicTheme = "light" | "dark";

const STORAGE_KEY = "vyntrio.public-theme";

function readInitialTheme(): PublicTheme {
  if (typeof window === "undefined") {
    return "light";
  }
  const stored = window.localStorage.getItem(STORAGE_KEY);
  if (stored === "light" || stored === "dark") {
    return stored;
  }
  return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
}

export function usePublicTheme() {
  const [theme, setThemeState] = useState<PublicTheme>(readInitialTheme);

  useEffect(() => {
    window.localStorage.setItem(STORAGE_KEY, theme);
  }, [theme]);

  const setTheme = useCallback((next: PublicTheme) => {
    setThemeState(next);
  }, []);

  const toggleTheme = useCallback(() => {
    setThemeState((prev) => (prev === "light" ? "dark" : "light"));
  }, []);

  return { theme, setTheme, toggleTheme };
}
