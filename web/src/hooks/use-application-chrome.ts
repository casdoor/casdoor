import * as React from "react";
import * as Cookie from "cookie";
import * as Conf from "@/Conf";
import {useTheme} from "@/hooks/use-theme";
import * as Setting from "@/lib/setting";

interface ThemeData {
  themeType?: string;
  colorPrimary?: string;
  borderRadius?: number;
  isCompact?: boolean;
}

interface Hsl {
  h: number;
  s: number;
  l: number;
}

/** "#262626" -> {h: 0, s: 0, l: 0.149}. */
function hexToHsl(hex: string): Hsl | null {
  const match = /^#?([\da-f]{3}|[\da-f]{6})$/i.exec(hex.trim());
  if (!match) {
    return null;
  }
  let value = match[1];
  if (value.length === 3) {
    value = value.split("").map((c) => c + c).join("");
  }
  const r = parseInt(value.slice(0, 2), 16) / 255;
  const g = parseInt(value.slice(2, 4), 16) / 255;
  const b = parseInt(value.slice(4, 6), 16) / 255;

  const max = Math.max(r, g, b);
  const min = Math.min(r, g, b);
  const l = (max + min) / 2;
  const d = max - min;

  let h = 0;
  let s = 0;
  if (d !== 0) {
    s = l > 0.5 ? d / (2 - max - min) : d / (max + min);
    if (max === r) {
      h = ((g - b) / d + (g < b ? 6 : 0)) / 6;
    } else if (max === g) {
      h = ((b - r) / d + 2) / 6;
    } else {
      h = ((r - g) / d + 4) / 6;
    }
  }

  return {h: h * 360, s, l};
}

/** The triplet shadcn's CSS variables expect: "0 0% 14.9%". */
function formatHsl({h, s, l}: Hsl): string {
  const round = (n: number) => Math.round(n * 10) / 10;
  return `${round(h)} ${round(s * 100)}% ${round(l * 100)}%`;
}

/** Black or white, whichever stays readable on top of that lightness. */
function foregroundFor(lightness: number): string {
  return lightness > 0.6 ? "0 0% 9%" : "0 0% 98%";
}

/**
 * A brand colour picked against a white page is usually far too dark to sit on a
 * dark one — Casdoor's own default, #262626, disappears into the panel it is
 * painted on. antd handled this in its `darkAlgorithm`, which derived a lighter
 * primary from the same hue; this keeps to the one step that matters: a colourful
 * brand is lifted to a usable lightness, and a neutral one (a black or a grey)
 * hands over to the palette's own near-white primary.
 */
function adaptPrimaryForDark(color: Hsl): Hsl {
  if (color.s < 0.15) {
    return {...color, l: 0.98};
  }
  return {...color, l: Math.max(color.l, 0.62)};
}

/** Pushes a `themeData` object into the shadcn CSS variables for as long as it is mounted. */
export function useThemeData(themeData: ThemeData | undefined | null) {
  // the provider owns the `dark` class, so re-assert a forced dark theme whenever
  // it writes, otherwise the two would fight over the root element
  const {resolvedTheme} = useTheme();

  React.useEffect(() => {
    const root = document.documentElement;
    const primary = themeData?.colorPrimary ? hexToHsl(themeData.colorPrimary) : null;
    const touched: string[] = [];
    // `themeType: "dark"` forces the dark palette on below, so the brand colour
    // has to be adapted for it even when the visitor asked for the light one
    const dark = themeData?.themeType === "dark" || resolvedTheme === "dark";

    if (primary) {
      // the un-adapted colour stays available to a subtree painted light inside a
      // dark page, see `.theme-scope-light` in index.css
      root.style.setProperty("--primary-base", formatHsl(primary));
      root.style.setProperty("--primary-base-foreground", foregroundFor(primary.l));

      const painted = dark ? adaptPrimaryForDark(primary) : primary;
      root.style.setProperty("--primary", formatHsl(painted));
      // keep the label readable whichever end of the scale the brand colour sits on
      root.style.setProperty("--primary-foreground", foregroundFor(painted.l));
      root.style.setProperty("--ring", formatHsl(painted));
      touched.push("--primary-base", "--primary-base-foreground", "--primary", "--primary-foreground", "--ring");
    }
    if (typeof themeData?.borderRadius === "number") {
      root.style.setProperty("--radius", `${themeData.borderRadius / 16}rem`);
      touched.push("--radius");
    }

    // "dark" is antd's darkAlgorithm and "compact" its compactAlgorithm
    const forceDark = themeData?.themeType === "dark" && resolvedTheme !== "dark";
    if (forceDark) {
      root.classList.add("dark");
    }
    const compact = themeData?.isCompact === true;
    if (compact) {
      root.classList.add("compact");
    }

    return () => {
      touched.forEach((name) => root.style.removeProperty(name));
      if (forceDark) {
        root.classList.remove("dark");
      }
      if (compact) {
        root.classList.remove("compact");
      }
    };
  }, [themeData, resolvedTheme]);
}

/**
 * Applies the organization/application `themeData` to the shadcn CSS variables.
 * Before the application has loaded, the backend's `organizationTheme` cookie
 * (set by routers/theme_filter.go) supplies the same values, which is what keeps
 * a branded sign-in page from flashing the default palette first.
 *
 * `enabled` is false for the application editor's preview: the previewed
 * application must not restyle the console it is embedded in.
 */
export function useApplicationTheme(application: any, enabled = true) {
  const themeData: ThemeData = React.useMemo(() => {
    if (application) {
      return Setting.getThemeData(application.organizationObj, application);
    }
    const cookieTheme = readCookie("organizationTheme");
    if (cookieTheme && cookieTheme !== "null") {
      try {
        return JSON.parse(cookieTheme);
      } catch {
        // fall through to the default
      }
    }
    return Conf.ThemeDefault;
  }, [application]);

  useThemeData(enabled ? themeData : undefined);
  return themeData;
}

function readCookie(name: string): string | undefined {
  try {
    return Cookie.parse(document.cookie)[name];
  } catch {
    return undefined;
  }
}

/** The logo/footer the backend put in a cookie, used until the application loads. */
export function getOrganizationCookieChrome() {
  const logo = readCookie("organizationLogo");
  const footerHtml = readCookie("organizationFootHtml");
  return {
    logo: logo ? logo : undefined,
    footerHtml: footerHtml ? footerHtml : undefined,
  };
}

/**
 * Sets the tab title and favicon from the application, falling back to the
 * organization. Port of Setting.renderHelmet(), which used react-helmet.
 */
export function useApplicationHelmet(application: any) {
  React.useEffect(() => {
    const organization = application?.organizationObj;
    if (!application || !organization) {
      return;
    }

    // the application's own title/favicon win over the organization's
    const title = application.title || organization.displayName;
    const favicon = application.favicon || organization.favicon;

    const previousTitle = document.title;
    if (title) {
      document.title = title;
    }

    let link = document.querySelector<HTMLLinkElement>("link[rel='icon']");
    const previousHref = link?.getAttribute("href") ?? null;
    if (favicon) {
      if (!link) {
        link = document.createElement("link");
        link.rel = "icon";
        document.head.appendChild(link);
      }
      link.href = favicon;
    }

    return () => {
      document.title = previousTitle;
      if (favicon && link) {
        if (previousHref === null) {
          link.remove();
        } else {
          link.href = previousHref;
        }
      }
    };
  }, [application]);
}

/** what the tab shows before an organization has been loaded */
const DEFAULT_FAVICON = "https://cdn.casdoor.com/static/favicon.png";

/**
 * The console's tab title and favicon, which follow the signed-in user's
 * organization. Signed out, the favicon falls back to Casdoor's own — the same
 * two <Helmet> blocks web/src/App.js renders.
 */
export function useAccountHelmet(account: any) {
  const organization = account?.organization;
  const title = organization?.displayName;
  const favicon = organization?.favicon || DEFAULT_FAVICON;

  React.useEffect(() => {
    if (title) {
      document.title = title;
    }

    let link = document.querySelector<HTMLLinkElement>("link[rel='icon']");
    if (!link) {
      link = document.createElement("link");
      link.rel = "icon";
      document.head.appendChild(link);
    }
    link.href = favicon;
  }, [title, favicon]);
}

const customHeadLoadedIds = new Set<string>();

/**
 * Appends an application's custom HTML (`headerHtml` / `pageHtml`) to <head>,
 * re-creating <script> nodes so the browser actually executes them. Port of
 * web/src/basic/CustomHead.js — like it, each id is injected only once per page
 * load, since the injected markup is not tracked for removal.
 */
export function useCustomHead(html: string | undefined | null, id = "default") {
  React.useEffect(() => {
    if (!html || customHeadLoadedIds.has(id)) {
      return;
    }

    const container = document.createElement("div");
    container.innerHTML = html;

    Array.from(container.childNodes).forEach((node) => {
      if (node.nodeType !== Node.ELEMENT_NODE) {
        return;
      }
      let element = node as Element;
      if (element.localName === "script") {
        const script = document.createElement("script");
        Array.from(element.attributes).forEach((attr) => script.setAttribute(attr.name, attr.value));
        script.text = element.textContent ?? "";
        element = script;
      }
      element.setAttribute("app-custom-head", id);
      document.head.appendChild(element);
    });

    customHeadLoadedIds.add(id);
  }, [html, id]);
}
