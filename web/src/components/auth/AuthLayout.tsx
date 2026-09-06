import * as React from "react";
import i18next from "i18next";
import {AlertCircle, ArrowLeft} from "lucide-react";
import {PoweredBy} from "@/components/layout/PoweredBy";
import {Button} from "@/components/ui/button";
import {LanguageSelect} from "@/components/common/LanguageSelect";
import {ThemeToggle} from "@/components/common/ThemeToggle";
import {CustomHtml, CustomStyle} from "@/components/common/CustomHtml";
import {
  getOrganizationCookieChrome,
  useApplicationHelmet,
  useApplicationTheme,
  useCustomHead,
} from "@/hooks/use-application-chrome";
import {useIsMobile} from "@/hooks/use-mobile";
import {useIsDark} from "@/hooks/use-theme";
import * as Setting from "@/lib/setting";
import {cn} from "@/lib/utils";

const PHONE_QUERY = "(max-width: 639.98px)";

interface AuthLayoutProps {
  application?: any;
  children: React.ReactNode;
  className?: string;
  /** widen the card for the signup form */
  wide?: boolean;
  /** the sign-in page drives these two slots from its own signinItems */
  hideLogo?: boolean;
  hideLanguages?: boolean;
  /** told which language the visitor picked, so the signup page can save it */
  onLanguageChange?: (key: string) => void;
  /** a back arrow above the form, for a form that has steps */
  onBack?: () => void;
  /** rendered under the panel, outside it: the "no account? sign up" line */
  footer?: React.ReactNode;
  /**
   * Rendered inside the application editor's preview: the application's theme,
   * title, favicon and custom head belong to the visitor's page, not to the
   * console the preview is embedded in.
   */
  preview?: boolean;
}

/** The "we cannot sign you in" panel, port of auth/Util.js renderMessageLarge(). */
function BlockedMessage({message}: {message: string}) {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4 px-4 text-center">
      <AlertCircle className="h-12 w-12 text-destructive" />
      <h1 className="text-xl font-semibold">{i18next.t("general:There was a problem signing you in..")}</h1>
      <p className="max-w-lg text-sm text-muted-foreground">{message}</p>
      <Button onClick={() => window.history.go(-2)}>{i18next.t("general:Back")}</Button>
    </div>
  );
}

/**
 * The rule between the credential form and the other ways in. It draws its own
 * line instead of covering one, so it needs to know nothing about the colour an
 * application painted its panel.
 */
export function AuthDivider({label}: {label?: string}) {
  return (
    <div className="flex items-center gap-3">
      <span className="h-px flex-1 bg-border" />
      <span className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
        {label || i18next.t("general:Or")}
      </span>
      <span className="h-px flex-1 bg-border" />
    </div>
  );
}

const NAMED_LIGHT_COLORS = new Set(["white", "whitesmoke", "snow", "ivory", "azure", "aliceblue", "ghostwhite", "floralwhite", "mintcream", "seashell", "linen"]);

/**
 * The `sm` breakpoint, so the panel dissolves into the page exactly where the
 * phone rules in index.css take over. `useIsMobile` answers a wider question —
 * which of an application's two sets of assets to use — and is 128px off.
 */
function useIsPhone() {
  return React.useSyncExternalStore(
    (onChange) => {
      const mql = window.matchMedia(PHONE_QUERY);
      mql.addEventListener("change", onChange);
      return () => mql.removeEventListener("change", onChange);
    },
    () => window.matchMedia(PHONE_QUERY).matches,
    () => false,
  );
}

/** Whether a CSS colour is a light one, ignoring a see-through value. */
function isLightColor(color: string): boolean {
  const value = color.trim().toLowerCase();
  if (NAMED_LIGHT_COLORS.has(value)) {
    return true;
  }

  let rgb: number[] | null = null;
  const hex = /^#([\da-f]{3}|[\da-f]{4}|[\da-f]{6}|[\da-f]{8})$/i.exec(value);
  if (hex) {
    let digits = hex[1];
    if (digits.length <= 4) {
      digits = digits.split("").map((c) => c + c).join("");
    }
    if (digits.length === 8 && parseInt(digits.slice(6, 8), 16) < 128) {
      return false;
    }
    rgb = [0, 2, 4].map((at) => parseInt(digits.slice(at, at + 2), 16));
  } else {
    const fn = /^rgba?\(([^)]+)\)$/.exec(value);
    if (fn) {
      const parts = fn[1].split(/[\s,/]+/).filter(Boolean).map(Number);
      if (parts.length >= 4 && parts[3] < 0.5) {
        return false;
      }
      rgb = parts.slice(0, 3);
    }
  }

  if (!rgb || rgb.some((channel) => !Number.isFinite(channel))) {
    return false;
  }
  return (0.2126 * rgb[0] + 0.7152 * rgb[1] + 0.0722 * rgb[2]) / 255 > 0.55;
}

/**
 * The colour an application's own form CSS paints the panel, if it paints one.
 *
 * This is read out of the stylesheet rather than measured on the element: the
 * panel has to know which palette it is on before its first paint, and Chrome
 * does not reliably repaint a subtree whose colours only moved because a custom
 * property above it changed.
 */
function getPanelBackground(css: string | undefined | null): string | null {
  const inner = Setting.getStyleInnerCss(css);
  if (!inner) {
    return null;
  }
  let found: string | null = null;
  // the last rule wins, the way the cascade would settle it
  const rules = /(^|[},])([^{}]*\.login-panel[^{}]*)\{([^}]*)\}/g;
  let rule: RegExpExecArray | null;
  while ((rule = rules.exec(inner)) !== null) {
    const declarations = /(?:^|;)\s*background(?:-color)?\s*:\s*([^;!]+)/g;
    let declaration: RegExpExecArray | null;
    while ((declaration = declarations.exec(rule[3])) !== null) {
      found = declaration[1].trim();
    }
  }
  return found;
}

/**
 * Centred panel shared by the sign-in, sign-up, password-reset and result pages.
 *
 * The panel is a single column: logo, heading, form, and a line under the card
 * for whatever leads away from it. Everything an application can restyle keeps
 * the class name the antd frontend gave it — `login-panel`, `login-form`,
 * `login-logo-box` — so form CSS written against the old pages still lands.
 */
export function AuthLayout({
  application,
  children,
  className,
  wide,
  hideLogo,
  hideLanguages,
  onLanguageChange,
  onBack,
  footer,
  preview,
}: AuthLayoutProps) {
  useApplicationTheme(application, !preview);
  // an application can force the dark palette regardless of the visitor's own
  // preference, and the logo has to follow the palette that is actually painted
  const isDark = useIsDark();
  const isMobile = useIsMobile();
  const isPhone = useIsPhone();
  useApplicationHelmet(preview ? null : application);
  // headerHtml is the organization/application chrome, pageHtml is the per-page one
  useCustomHead(preview ? undefined : application?.headerHtml, "header");
  useCustomHead(preview ? undefined : application?.pageHtml, "page");

  // the backend hands us the organization's branding in cookies so the first
  // paint is already branded, before /api/get-application has come back
  const cookieChrome = React.useMemo(() => getOrganizationCookieChrome(), []);

  // an application whose own form CSS paints a light panel would otherwise show
  // white text on white once the visitor is on the dark palette
  const formCss = isMobile ? application?.formCssMobile : application?.formCss;
  const panelBackground = React.useMemo(() => getPanelBackground(formCss), [formCss]);
  const panelIsLight = isDark && !!panelBackground && isLightColor(panelBackground);

  // an IP restriction on either the application or its organization blocks the
  // whole sign-in surface, the same way EntryPage does in the antd frontend
  const ipRestriction = application?.ipRestriction || application?.organizationObj?.ipRestriction;
  if (ipRestriction) {
    return <BlockedMessage message={ipRestriction} />;
  }

  const logo = Setting.getThemedLogo(
    application?.logo || cookieChrome.logo,
    application?.logoDark || application?.organizationObj?.logoDark,
    [isDark && !panelIsLight ? "dark" : "light"],
  );
  const footerHtml = application?.footerHtml || cookieChrome.footerHtml;

  // the customized form styling is meant for the standalone page, an embedded
  // one keeps the host page's own look, as the antd pages did
  const embedded = Setting.inIframe();
  const backgroundUrl = embedded
    ? undefined
    : (isMobile ? application?.formBackgroundUrlMobile : application?.formBackgroundUrl);
  // formOffset: 1 left, 2 center, 3 right, 4 center with the side panel
  const offset = embedded || isMobile ? 2 : (application?.formOffset ?? 2);
  const sidePanel = offset === 4 && application?.formSideHtml;
  // on a phone the panel *is* the page, so its border and shadow only add noise
  // — unless it has to hold itself off a background image
  const bleed = isPhone && !backgroundUrl;

  return (
    <div
      className={cn(
        "login-background relative flex min-h-screen flex-col",
        // the tint is what lifts the card off the page; a panel that has
        // dissolved into the page has nothing to be lifted off
        bleed ? "bg-background" : "bg-muted/40 dark:bg-background",
        // a phone browser sizes a fixed background against the whole document,
        // not the viewport, so it only stays attached on the desktop
        backgroundUrl && cn("bg-cover bg-center", isMobile ? "bg-scroll" : "bg-fixed"),
      )}
      style={backgroundUrl ? {backgroundImage: `url(${backgroundUrl})`} : undefined}
    >
      {embedded ? null : <CustomStyle css={formCss} />}

      <div className="flex items-center justify-end gap-1 p-3 sm:p-4">
        {hideLanguages ? null : (
          <LanguageSelect
            className="rounded-full text-muted-foreground hover:text-foreground"
            languages={application?.organizationObj?.languages}
            onLanguageChange={onLanguageChange}
          />
        )}
        <ThemeToggle className="rounded-full text-muted-foreground hover:text-foreground" />
      </div>

      <div
        className={cn(
          "login-content flex flex-1 items-start pb-10 pt-2 sm:px-4 sm:pb-16",
          // a full-bleed panel brings its own phone padding, see index.css
          bleed ? "px-0" : "px-4",
          offset === 1 ? "justify-start sm:pl-[10%]" : offset === 3 ? "justify-end sm:pr-[10%]" : "justify-center",
        )}
      >
        {/* my-auto centres the column rather than items-center, which would clip
            the top of a panel taller than the viewport; on a phone the form
            starts near the top instead, where a thumb can reach it */}
        <div
          className={cn(
            "mb-auto w-full max-sm:mt-2 sm:mt-auto",
            sidePanel ? "max-w-4xl" : wide ? "max-w-xl" : "max-w-md",
          )}
        >
          {/* one box: an application's own `.login-panel` styles the card itself */}
          <div
            className={cn(
              "login-panel flex w-full items-center overflow-hidden text-card-foreground",
              panelIsLight && "theme-scope-light",
              bleed ? "bg-transparent" : "rounded-xl border bg-card shadow-sm",
              // the padding sits on the element an application restyles, so its own
              // `.login-panel { padding }` replaces this one instead of adding to it
              sidePanel ? null : bleed ? "px-1 pt-4" : "px-6 pt-6 sm:px-8 sm:pt-8",
            )}
          >
            {sidePanel ? (
              <CustomHtml html={application.formSideHtml} className="side-image hidden w-[420px] shrink-0 self-stretch lg:block" />
            ) : null}
            <div
              className={cn(
                "login-form w-full min-w-0",
                sidePanel ? (bleed ? "px-1 py-4" : "p-6 sm:p-8") : bleed ? "pb-4" : "pb-6 sm:pb-8",
                className,
              )}
            >
              {hideLogo ? null : (
                <div className="login-logo-box mb-6 flex justify-center">
                  {application?.homepageUrl ? (
                    <a href={application.homepageUrl} target="_blank" rel="noreferrer">
                      <img src={logo} alt={application?.displayName ?? "Casdoor"} className="h-10 max-w-full object-contain" />
                    </a>
                  ) : (
                    <img src={logo} alt={application?.displayName ?? "Casdoor"} className="h-10 max-w-full object-contain" />
                  )}
                </div>
              )}

              {onBack ? (
                <Button
                  type="button"
                  variant="ghost"
                  size="iconSm"
                  className="-ml-2 mb-2 text-muted-foreground hover:text-foreground"
                  aria-label={i18next.t("general:Back")}
                  onClick={onBack}
                >
                  <ArrowLeft />
                </Button>
              ) : null}

              {children}
            </div>
          </div>

          {footer ? <div className="mt-6 px-4 text-center text-sm text-muted-foreground">{footer}</div> : null}
        </div>
      </div>

      {/* below the centred card and at the bottom of the viewport, as antd's
          Layout.Footer is a sibling of the Content it follows */}
      {footerHtml ? (
        <footer
          id="footer"
          className="shrink-0 py-6 text-center text-xs text-muted-foreground"
          dangerouslySetInnerHTML={{__html: footerHtml}}
        />
      ) : (
        <footer id="footer" className="shrink-0 py-6 text-center text-xs text-muted-foreground">
          <PoweredBy />
        </footer>
      )}
    </div>
  );
}
