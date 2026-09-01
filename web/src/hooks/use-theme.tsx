import * as React from "react";

type Theme = "light" | "dark" | "system";

interface ThemeContextValue {
  theme: Theme;
  resolvedTheme: "light" | "dark";
  setTheme: (theme: Theme) => void;
}

const ThemeContext = React.createContext<ThemeContextValue>({
  theme: "system",
  resolvedTheme: "light",
  setTheme: () => undefined,
});

const STORAGE_KEY = "themeAlgorithm";

function readStoredTheme(): Theme {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) {
      return "system";
    }
    // Stay compatible with the legacy antd frontend, which stored an array
    // such as ["dark"] or ["default", "compact"].
    if (raw.startsWith("[")) {
      const parsed = JSON.parse(raw);
      return Array.isArray(parsed) && parsed.includes("dark") ? "dark" : "light";
    }
    if (raw === "dark" || raw === "light" || raw === "system") {
      return raw;
    }
  } catch {
    // fall through
  }
  return "system";
}

function systemPrefersDark(): boolean {
  return typeof window !== "undefined" && window.matchMedia("(prefers-color-scheme: dark)").matches;
}

export function ThemeProvider({children}: {children: React.ReactNode}) {
  const [theme, setThemeState] = React.useState<Theme>(() => readStoredTheme());
  const [systemDark, setSystemDark] = React.useState<boolean>(() => systemPrefersDark());

  React.useEffect(() => {
    const mq = window.matchMedia("(prefers-color-scheme: dark)");
    const onChange = (e: MediaQueryListEvent) => setSystemDark(e.matches);
    mq.addEventListener("change", onChange);
    return () => mq.removeEventListener("change", onChange);
  }, []);

  const resolvedTheme: "light" | "dark" = theme === "system" ? (systemDark ? "dark" : "light") : theme;

  React.useEffect(() => {
    const root = document.documentElement;
    root.classList.toggle("dark", resolvedTheme === "dark");
    root.style.colorScheme = resolvedTheme;
  }, [resolvedTheme]);

  const setTheme = React.useCallback((next: Theme) => {
    setThemeState(next);
    try {
      localStorage.setItem(STORAGE_KEY, next === "dark" ? JSON.stringify(["dark"]) : JSON.stringify(["default"]));
    } catch {
      // ignore quota / private-mode failures
    }
  }, []);

  const value = React.useMemo(() => ({theme, resolvedTheme, setTheme}), [theme, resolvedTheme, setTheme]);

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>;
}

export function useTheme() {
  return React.useContext(ThemeContext);
}

/**
 * Whether the dark palette is actually painted right now. That is not always the
 * viewer's own preference: an organization or application whose `themeData` sets
 * `themeType: "dark"` forces the class on too (see `useThemeData`), and anything
 * that swaps a light asset for a dark one has to follow what is on screen.
 */
export function useIsDark() {
  const {resolvedTheme} = useTheme();
  const [forced, setForced] = React.useState(() =>
    typeof document !== "undefined" && document.documentElement.classList.contains("dark"),
  );

  React.useEffect(() => {
    const root = document.documentElement;
    const sync = () => setForced(root.classList.contains("dark"));
    sync();
    const observer = new MutationObserver(sync);
    observer.observe(root, {attributes: true, attributeFilter: ["class"]});
    return () => observer.disconnect();
  }, []);

  return resolvedTheme === "dark" || forced;
}
