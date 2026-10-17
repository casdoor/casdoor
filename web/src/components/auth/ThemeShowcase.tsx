import * as React from "react";
import {createPortal} from "react-dom";
import i18next from "i18next";
import {ChevronLeft, ChevronRight, Pause, Play} from "lucide-react";
import * as Conf from "@/Conf";
import {Button} from "@/components/ui/button";
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from "@/components/ui/select";
import {getSafeHtml} from "@/lib/custom-html";
import {applyTheme, getLocalized} from "@/lib/integration";

const ROTATE_MS = 6000;
// the application's own look, listed before the Marketplace themes
const DEFAULT_THEME = "default";

interface ShowcaseTheme {
  id: string;
  name: Record<string, string> | string;
  url: string;
}

interface ThemeBundle {
  theme: any;
  values: Record<string, string>;
}

let themesPromise: Promise<ShowcaseTheme[]> | null = null;
const bundlePromises = new Map<string, Promise<ThemeBundle | null>>();

function getMarketplaceFileUrl(path: string): string {
  return `${Conf.MarketplaceUrl.replace(/\/+$/, "")}/${path.replace(/^\/+/, "")}`;
}

function loadThemes(): Promise<ShowcaseTheme[]> {
  if (!themesPromise) {
    themesPromise = fetch(getMarketplaceFileUrl("index.json"))
      .then((res) => res.json())
      .then((index: any) => {
        const revoked = (index.revoked ?? []) as any[];
        return ((index.integrations ?? []) as any[]).flatMap((item) => {
          const version = (item.versions ?? []).find((v: any) => v.version === item.latest);
          const isRevoked = revoked.some((r) => r.id === item.id && (!r.version || r.version === version?.version));
          return item.type === "theme" && version && !isRevoked ? [{id: item.id, name: item.name, url: version.url}] : [];
        });
      })
      .catch(() => {
        themesPromise = null;
        return [];
      });
  }
  return themesPromise;
}

/** A theme's files, with each variable set to its example, the way the install dialog previews it. */
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

function getInitialTheme(): string {
  return new URLSearchParams(window.location.search).get("theme") ?? DEFAULT_THEME;
}

/**
 * Puts the Marketplace themes on show on the sign-in page of a demo site: they take turns
 * on the application, and the visitor can step through them or pick one. Nothing is saved,
 * the themes only change the copy of the application the page renders. `?theme=<id>` opens
 * the page on one theme, stopped.
 */
export function useThemeShowcase(application: any, enabled: boolean): {application: any; showcase: React.ReactNode} {
  const [themes, setThemes] = React.useState<ShowcaseTheme[]>([]);
  const [selected, setSelected] = React.useState(getInitialTheme);
  const [playing, setPlaying] = React.useState(() => getInitialTheme() === DEFAULT_THEME);
  const [bundles, setBundles] = React.useState<Record<string, ThemeBundle>>({});
  // the theme on screen stays until the selected one has loaded
  const [shown, setShown] = React.useState(DEFAULT_THEME);

  React.useEffect(() => {
    if (enabled) {
      loadThemes().then(setThemes);
    }
  }, [enabled]);

  const ids = React.useMemo(() => [DEFAULT_THEME, ...themes.map((theme) => theme.id)], [themes]);
  const index = Math.max(0, ids.indexOf(selected));
  const nextId = ids[(index + 1) % ids.length];

  React.useEffect(() => {
    if (!enabled) {
      return;
    }
    let cancelled = false;
    for (const id of [selected, nextId]) {
      const theme = themes.find((item) => item.id === id);
      if (theme && !bundles[id]) {
        loadBundle(theme).then((bundle) => {
          if (!cancelled && bundle) {
            setBundles((prev) => ({...prev, [id]: bundle}));
          }
        });
      }
    }
    return () => {
      cancelled = true;
    };
  }, [enabled, themes, selected, nextId, bundles]);

  React.useEffect(() => {
    if (selected === DEFAULT_THEME || bundles[selected]) {
      setShown(selected);
    }
  }, [selected, bundles]);

  React.useEffect(() => {
    if (!enabled || !playing || ids.length < 2) {
      return;
    }
    const timer = window.setTimeout(() => setSelected(nextId), ROTATE_MS);
    return () => window.clearTimeout(timer);
  }, [enabled, playing, ids.length, nextId, shown]);

  // typing into the form stops the themes, the page must not change under the visitor's hands
  React.useEffect(() => {
    if (!enabled) {
      return;
    }
    const onFocusIn = (event: FocusEvent) => {
      const target = event.target as HTMLElement | null;
      if (target?.closest(".login-form") && target.matches("input, textarea, select")) {
        setPlaying(false);
      }
    };
    document.addEventListener("focusin", onFocusIn);
    return () => document.removeEventListener("focusin", onFocusIn);
  }, [enabled]);

  const themedApplication = React.useMemo(() => {
    const bundle = bundles[shown];
    if (!enabled || !application || shown === DEFAULT_THEME || !bundle) {
      return application;
    }
    const themed = applyTheme(application, bundle.theme, bundle.values);
    // the theme comes from the Marketplace, not from the organization, so its HTML is sanitized
    return {...themed, formSideHtml: getSafeHtml(themed.formSideHtml, false), footerHtml: getSafeHtml(themed.footerHtml, false)};
  }, [enabled, application, shown, bundles]);

  if (!enabled || !application) {
    return {application, showcase: null};
  }

  const pick = (id: string) => {
    setSelected(id);
    setPlaying(false);
  };
  const getName = (id: string) => {
    const theme = themes.find((item) => item.id === id);
    return theme ? getLocalized(theme.name) : i18next.t("general:Default");
  };

  const showcase = createPortal(
    <div className="fixed bottom-3 left-3 z-50 sm:bottom-4 sm:left-4">
      <style>{"@keyframes casdoor-theme-showcase{from{transform:scaleX(0)}to{transform:scaleX(1)}}"}</style>
      <div className="relative flex items-center gap-0.5 overflow-hidden rounded-full border bg-background/90 p-1 text-foreground shadow-lg backdrop-blur">
        <Button
          type="button"
          variant="ghost"
          size="iconSm"
          className="rounded-full"
          aria-label={i18next.t("theme:Previous theme")}
          title={i18next.t("theme:Previous theme")}
          onClick={() => pick(ids[(index - 1 + ids.length) % ids.length])}
        >
          <ChevronLeft />
        </Button>
        <Select value={ids[index]} onValueChange={pick} onOpenChange={(open) => open && setPlaying(false)}>
          <SelectTrigger className="h-8 w-36 rounded-full border-0 bg-transparent shadow-none focus:ring-0" aria-label={i18next.t("theme:Theme")}>
            <SelectValue />
          </SelectTrigger>
          <SelectContent side="top" className="max-h-80">
            {ids.map((id) => (
              <SelectItem key={id} value={id}>{getName(id)}</SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Button
          type="button"
          variant="ghost"
          size="iconSm"
          className="rounded-full"
          aria-label={i18next.t("theme:Next theme")}
          title={i18next.t("theme:Next theme")}
          onClick={() => pick(nextId)}
        >
          <ChevronRight />
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="iconSm"
          className="rounded-full"
          aria-label={playing ? i18next.t("theme:Pause themes") : i18next.t("theme:Play themes")}
          title={playing ? i18next.t("theme:Pause themes") : i18next.t("theme:Play themes")}
          onClick={() => setPlaying(!playing)}
        >
          {playing ? <Pause /> : <Play />}
        </Button>
        {playing ? (
          <span
            key={`${shown}-${selected}`}
            className="pointer-events-none absolute inset-x-4 bottom-0 h-0.5 origin-left bg-primary"
            style={{animation: `casdoor-theme-showcase ${ROTATE_MS}ms linear`}}
          />
        ) : null}
      </div>
    </div>,
    document.body,
  );

  return {application: themedApplication, showcase};
}
