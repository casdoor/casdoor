import * as React from "react";
import i18next from "i18next";
import * as SystemBackend from "@/backend/SystemInfo";

export const PricingUrl = "https://casdoor.cn/pricing";

const ranks: Record<string, number> = {Community: 0, Commercial: 1, Enterprise: 2};

let editionPromise: Promise<string> | null = null;

function getEdition(): Promise<string> {
  if (editionPromise === null) {
    editionPromise = SystemBackend.getLicense()
      .then((res: any) => (res.status === "ok" ? res.data.edition : "Community"))
      .catch(() => "Community");
  }
  return editionPromise;
}

export function useEdition(): string | null {
  const [edition, setEdition] = React.useState<string | null>(null);

  React.useEffect(() => {
    let stopped = false;
    getEdition().then((value) => {
      if (!stopped) {
        setEdition(value);
      }
    });
    return () => {
      stopped = true;
    };
  }, []);

  return edition;
}

export function getEditionText(edition: string) {
  if (edition === "Enterprise") {
    return i18next.t("system:Enterprise edition");
  }
  if (edition === "Commercial") {
    return i18next.t("system:Commercial edition");
  }
  return i18next.t("system:Community edition");
}

/** Tells that a setting doesn't take effect, when the license gives a lower edition than it needs. */
export function EditionNotice({edition = "Commercial"}: {edition?: string}) {
  const current = useEdition();
  if (current === null || (ranks[current] ?? 0) >= ranks[edition]) {
    return null;
  }

  return (
    <p className="mt-1.5 text-xs text-warning">
      {i18next.t("general:Commercial edition required")}{" "}
      <a href={PricingUrl} target="_blank" rel="noreferrer" className="underline underline-offset-2">
        {i18next.t("general:Learn more")}
      </a>
    </p>
  );
}
