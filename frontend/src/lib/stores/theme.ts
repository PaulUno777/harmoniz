import { writable } from "svelte/store";

type Theme = "dark" | "light" | "system";

// Helper function to call Wails window theme functions safely
function setWindowTheme(theme: "dark" | "light" | "system") {
  if (typeof window === "undefined") return;

  try {
    // Try to use Wails runtime functions
    // @ts-ignore - Wails runtime may not be available in dev mode
    const runtime = window.runtime;
    if (runtime) {
      console.log(`[Theme] Setting window title bar theme to: ${theme}`);
      if (theme === "system") {
        runtime.WindowSetSystemDefaultTheme?.();
        console.log(`[Theme] Called WindowSetSystemDefaultTheme()`);
      } else if (theme === "dark") {
        runtime.WindowSetDarkTheme?.();
        console.log(`[Theme] Called WindowSetDarkTheme()`);
      } else {
        runtime.WindowSetLightTheme?.();
        console.log(`[Theme] Called WindowSetLightTheme()`);
      }
    } else {
      console.warn(
        `[Theme] Wails runtime not available - window title bar theme won't change`,
      );
    }
  } catch (e) {
    // Wails runtime not available (e.g., in dev mode or browser)
    console.warn(`[Theme] Error setting window theme:`, e);
  }
}

function createThemeStore() {
  // Load initial theme from localStorage or default to system
  const savedTheme = (
    typeof localStorage !== "undefined"
      ? localStorage.getItem("theme")
      : "system"
  ) as Theme;
  const { subscribe, set } = writable<Theme>(savedTheme || "system");

  return {
    subscribe,
    set: (value: Theme) => {
      if (typeof localStorage !== "undefined") {
        localStorage.setItem("theme", value);
      }
      set(value);
      applyTheme(value);
    },
    init: () => {
      applyTheme(savedTheme);

      // Watch for system theme changes
      if (typeof window !== "undefined") {
        window
          .matchMedia("(prefers-color-scheme: dark)")
          .addEventListener("change", () => {
            const current = localStorage.getItem("theme") as Theme;
            if (current === "system") {
              applyTheme("system");
            }
          });
      }
    },
  };
}

function applyTheme(theme: Theme) {
  if (typeof document === "undefined") return;

  const root = document.documentElement;
  let effectiveTheme = theme;

  if (theme === "system") {
    effectiveTheme = window.matchMedia("(prefers-color-scheme: dark)").matches
      ? "dark"
      : "light";
  }

  // Apply CSS theme
  if (effectiveTheme === "light") {
    root.classList.add("light");
    root.classList.remove("dark");
    root.style.colorScheme = "light";
  } else {
    root.classList.add("dark");
    root.classList.remove("light");
    root.style.colorScheme = "dark";
  }

  // Apply window title bar theme (Wails) - use effective theme for system
  setWindowTheme(theme === "system" ? effectiveTheme : theme);
}

export const theme = createThemeStore();
