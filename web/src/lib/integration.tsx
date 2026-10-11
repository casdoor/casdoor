import * as React from "react";
import i18next from "i18next";

/** The public Marketplace site, where every integration has a page with its full guide. */
export const MarketplaceSiteUrl = "https://marketplace.casdoor.cn";

export const IntegrationTypes = ["app", "provider", "theme"] as const;

export function getIntegrationTypeLabel(type: string): string {
  switch (type) {
  case "app":
    return i18next.t("integration:App");
  case "provider":
    return i18next.t("general:Provider");
  case "theme":
    return i18next.t("theme:Theme");
  case "action":
    return i18next.t("general:Action");
  default:
    return type;
  }
}

/** Marketplace texts are maps of language to text, with English always present. */
export function getLocalized(text: Record<string, string> | string | undefined | null): string {
  if (!text) {
    return "";
  }
  if (typeof text === "string") {
    return text;
  }
  const language = i18next.language?.split("-")[0];
  return text[i18next.language] ?? text[language] ?? text.en ?? Object.values(text)[0] ?? "";
}

export function fillVariables(text: string, values: Record<string, string>): string {
  return text.replace(/\{\{\s*([a-zA-Z][a-zA-Z0-9]*)\s*\}\}/g, (match, name) => values[name] || match);
}

/** Guide texts use `code` and **bold**. */
export function renderInlineText(text: string): React.ReactNode[] {
  return text.split(/(`[^`]+`|\*\*[^*]+\*\*)/g).filter(Boolean).map((part, index) => {
    if (part.startsWith("`") && part.endsWith("`")) {
      return <code key={index} className="rounded bg-muted px-1 py-0.5 font-mono text-[0.85em]">{part.slice(1, -1)}</code>;
    }
    if (part.startsWith("**") && part.endsWith("**")) {
      return <strong key={index}>{part.slice(2, -2)}</strong>;
    }
    return <React.Fragment key={index}>{part}</React.Fragment>;
  });
}

export function parseBundle(bundle: string | undefined | null): any {
  if (!bundle) {
    return null;
  }
  try {
    return JSON.parse(bundle);
  } catch {
    return null;
  }
}

export function IntegrationLogo({logo, name, className = "size-10"}: {logo?: string; name: string; className?: string}) {
  const [failed, setFailed] = React.useState(false);
  if (logo && !failed) {
    return <img src={logo} alt="" className={`${className} shrink-0 object-contain`} onError={() => setFailed(true)} />;
  }
  return (
    <div className={`${className} flex shrink-0 items-center justify-center rounded-md bg-primary/10 font-semibold text-primary`}>
      {(name || "?").slice(0, 1).toUpperCase()}
    </div>
  );
}
