import * as React from "react";
import * as Conf from "@/Conf";
import * as Setting from "@/lib/setting";
import {getSafeHtml} from "@/lib/custom-html";
import {applyTheme} from "@/lib/integration";

/**
 * The Marketplace themes on show on a demo site: they take turns on the sign-in, sign-up and
 * other entry pages and on the console, and the visitor can step through them or pick one.
 * Nothing is saved on the server, the choice only lives in the tab. `?theme=<id>` opens the
 * site on one theme, stopped, also when the page is embedded in another site.
 */

export const ROTATE_MS = 6000;
// the site's own look, listed before the Marketplace themes
export const DEFAULT_THEME = "default";
const STORAGE_KEY = "themeShowcase";

export interface ShowcaseTheme {
  id: string;
  name: Record<string, string> | string;
  url: string;
}

export interface ThemeBundle {
  theme: any;
  values: Record<string, string>;
}

interface ShowcaseState {
  themes: ShowcaseTheme[];
  selected: string;
  // the theme on screen stays until the selected one has loaded
  shown: string;
  playing: boolean;
  bundles: Record<string, ThemeBundle>;
}

const listeners = new Set<() => void>();
let state: ShowcaseState | null = null;
let timer: number | undefined;
let timerKey = "";
const bundlePromises = new Map<string, Promise<ThemeBundle | null>>();

// a page embedded in another site shows a theme only when the host asks for one
const isFramed = Setting.inIframe();
const themeFromUrl = new URLSearchParams(window.location.search).get("theme");

export function isThemeShowcaseEnabled(): boolean {
  return Conf.ThemeShowcase && (!isFramed || !!themeFromUrl);
}

/** Embedded with `?theme=<id>`: the theme is fixed by the host page, with no switcher and nothing saved. */
export function isFramedShowcase(): boolean {
  return isFramed && !!themeFromUrl;
}

function getMarketplaceFileUrl(path: string): string {
  return `${Conf.MarketplaceUrl.replace(/\/+$/, "")}/${path.replace(/^\/+/, "")}`;
}

export function getThemeIds(themes: ShowcaseTheme[]): string[] {
  return [DEFAULT_THEME, ...themes.map((theme) => theme.id)];
}

function setState(patch: Partial<ShowcaseState>) {
  if (!state) {
    return;
  }
  state = {...state, ...patch};
  // an embedded page shares the tab's storage with the site itself, which keeps its own choice
  if (!isFramed) {
    try {
      sessionStorage.setItem(STORAGE_KEY, JSON.stringify({selected: state.selected, playing: state.playing}));
    } catch {
      // the choice is only kept for this page then
    }
  }
  schedule();
  listeners.forEach((listener) => listener());
}

function schedule() {
  if (!state) {
    return;
  }
  const ids = getThemeIds(state.themes);
  const key = state.playing && ids.length > 1 ? `${state.selected}|${state.shown}|${ids.length}` : "";
  if (key === timerKey) {
    return;
  }
  timerKey = key;
  window.clearTimeout(timer);
  if (key) {
    const next = ids[(Math.max(0, ids.indexOf(state.selected)) + 1) % ids.length];
    timer = window.setTimeout(() => select(next, true), ROTATE_MS);
  }
}

function loadBundle(theme: ShowcaseTheme): Promise<ThemeBundle | null> {
  let promise = bundlePromises.get(theme.id);
  if (!promise) {
    promise = fetch(getMarketplaceFileUrl(theme.url))
      .then((res) => res.json())
      .then((bundle: any) => {
        const files = bundle?.files?.["theme.json"];
        if (!files) {
          return null;
        }
        // each variable is set to its example, the way the install dialog previews the theme
        const values = Object.fromEntries(((bundle.manifest?.variables ?? []) as any[]).map((variable) => [variable.name, variable.example ?? ""]));
        // the background image is fetched ahead, so the theme does not show up before its picture
        const backgroundUrl = applyTheme({}, files, values).formBackgroundUrl;
        if (backgroundUrl) {
          new Image().src = backgroundUrl;
        }
        return {theme: files, values};
      })
      .catch(() => {
        bundlePromises.delete(theme.id);
        return null;
      });
    bundlePromises.set(theme.id, promise);
  }
  return promise;
}

function ensureBundle(id: string) {
  const theme = state?.themes.find((item) => item.id === id);
  if (!theme || state?.bundles[id]) {
    return;
  }
  loadBundle(theme).then((bundle) => {
    if (!state || !bundle) {
      return;
    }
    const bundles = {...state.bundles, [id]: bundle};
    setState(state.selected === id ? {bundles, shown: id} : {bundles});
  });
}

function select(id: string, playing: boolean) {
  if (!state) {
    return;
  }
  const ids = getThemeIds(state.themes);
  const ready = id === DEFAULT_THEME || !!state.bundles[id];
  setState(ready ? {selected: id, shown: id, playing} : {selected: id, playing});
  ensureBundle(id);
  // the one after it is fetched while this one is on screen
  ensureBundle(ids[(Math.max(0, ids.indexOf(id)) + 1) % ids.length]);
}

function init() {
  let saved: any = null;
  try {
    saved = JSON.parse(sessionStorage.getItem(STORAGE_KEY) ?? "null");
  } catch {
    saved = null;
  }
  const fromUrl = themeFromUrl;
  state = {
    themes: [],
    selected: fromUrl ?? saved?.selected ?? DEFAULT_THEME,
    shown: DEFAULT_THEME,
    playing: fromUrl ? false : saved?.playing ?? true,
    bundles: {},
  };

  fetch(getMarketplaceFileUrl("index.json"))
    .then((res) => res.json())
    .then((index: any) => {
      const revoked = (index.revoked ?? []) as any[];
      const themes = ((index.integrations ?? []) as any[]).flatMap((item) => {
        const version = (item.versions ?? []).find((v: any) => v.version === item.latest);
        const isRevoked = revoked.some((r) => r.id === item.id && (!r.version || r.version === version?.version));
        return item.type === "theme" && version && !isRevoked ? [{id: item.id, name: item.name, url: version.url}] : [];
      });
      setState({themes});
      if (state) {
        select(state.selected, state.playing);
      }
    })
    .catch(() => undefined);

  // typing stops the themes, the page must not change under the visitor's hands
  document.addEventListener("input", (event) => {
    if (state?.playing && event.isTrusted) {
      setState({playing: false});
    }
  }, true);
}

function subscribe(listener: () => void) {
  if (!state && isThemeShowcaseEnabled()) {
    init();
  }
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

/** The showcase's state, null when the site has no showcase. */
export function useThemeShowcase(): ShowcaseState | null {
  return React.useSyncExternalStore(subscribe, () => state, () => null);
}

export const showcaseActions = {
  pick: (id: string) => select(id, false),
  setPlaying: (playing: boolean) => setState({playing}),
};

/** The theme on screen, null for the site's own look or when the site has no showcase. */
export function useShowcaseBundle(enabled = true): ThemeBundle | null {
  const showcase = useThemeShowcase();
  if (!enabled || !showcase || showcase.shown === DEFAULT_THEME) {
    return null;
  }
  return showcase.bundles[showcase.shown] ?? null;
}

/** A copy of the application with the theme on screen, its HTML sanitized: the theme comes from the Marketplace, not from the organization. */
export function useShowcaseApplication(application: any, enabled = true): any {
  const bundle = useShowcaseBundle(enabled);
  return React.useMemo(() => {
    if (!bundle) {
      return application;
    }
    const themed = applyTheme(application ?? {}, bundle.theme, bundle.values);
    return {...themed, formSideHtml: getSafeHtml(themed.formSideHtml, false), footerHtml: getSafeHtml(themed.footerHtml, false)};
  }, [application, bundle]);
}

/**
 * The console version of a sign-in theme: the palette, color scheme and font its form CSS
 * sets on `.login-background`, put on the whole page, with the sidebar following the card.
 */
export function getConsoleCss(formCss: string | undefined | null): string {
  const inner = Setting.getStyleInnerCss(formCss) ?? "";
  const block = /\.login-background\s*\{([^}]*)\}/.exec(inner);
  if (!block) {
    return "";
  }
  const imports = inner.match(/@import\s+url\([^)]*\)[^;]*;/g) ?? [];
  const declarations = block[1].split(";").map((item) => item.trim()).filter((item) => /^(--[\w-]+|color-scheme)\s*:/.test(item));
  const font = block[1].split(";").map((item) => item.trim()).find((item) => /^font-family\s*:/.test(item));
  if (declarations.some((item) => item.startsWith("--card:"))) {
    declarations.push("--sidebar-background: var(--card)", "--sidebar-foreground: var(--card-foreground)");
  }
  if (declarations.some((item) => item.startsWith("--accent:"))) {
    declarations.push("--sidebar-accent: var(--accent)", "--sidebar-accent-foreground: var(--accent-foreground)");
  }
  if (declarations.some((item) => item.startsWith("--border:"))) {
    declarations.push("--sidebar-border: var(--border)");
  }
  return `${imports.join("\n")}\n:root { ${declarations.join("; ")} }\n${font ? `body { ${font} }` : ""}`;
}
