import * as React from "react";
import i18next from "i18next";
import * as SystemBackend from "@/backend/SystemInfo";
import {PricingUrl} from "@/lib/setting";

export {PricingUrl};

const ranks: Record<string, number> = {Community: 0, Commercial: 1, Enterprise: 2};

interface LicenseInfo {
  edition: string;
  error?: string;
}

let licensePromise: Promise<LicenseInfo> | null = null;

function getLicenseInfo(): Promise<LicenseInfo> {
  if (licensePromise === null) {
    licensePromise = SystemBackend.getLicense()
      .then((res: any) => (res.status === "ok" ? res.data : {edition: "Community"}))
      .catch(() => ({edition: "Community"}));
  }
  return licensePromise;
}

function useLicenseInfo(): LicenseInfo | null {
  const [info, setInfo] = React.useState<LicenseInfo | null>(null);

  React.useEffect(() => {
    let stopped = false;
    getLicenseInfo().then((value) => {
      if (!stopped) {
        setInfo(value);
      }
    });
    return () => {
      stopped = true;
    };
  }, []);

  return info;
}

export function useEdition(): string | null {
  return useLicenseInfo()?.edition ?? null;
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

/** Tells that a setting needs a license, when the license gives a lower edition than it needs. */
export function EditionNotice({edition = "Commercial", message}: {edition?: string; message?: string}) {
  const current = useEdition();
  if (current === null || (ranks[current] ?? 0) >= ranks[edition]) {
    return null;
  }

  return (
    <p className="mt-1.5 text-xs text-warning">
      {message ?? i18next.t("general:Commercial edition required")}{" "}
      <a href={PricingUrl} target="_blank" rel="noreferrer" className="underline underline-offset-2">
        {i18next.t("general:View editions and pricing")}
      </a>
    </p>
  );
}

/** Tells the administrators of the console that the configured license is invalid or has expired, so the paid features are off. */
export function LicenseBanner() {
  const info = useLicenseInfo();
  if (!info?.error) {
    return null;
  }

  return (
    <div className="border-b border-warning/40 bg-warning/10 px-4 py-2 text-sm text-warning md:px-6">
      {i18next.t("system:The license is invalid or has expired, the paid features are turned off")}: {info.error}{" "}
      <a href={PricingUrl} target="_blank" rel="noreferrer" className="font-medium underline underline-offset-2">
        {i18next.t("general:View editions and pricing")}
      </a>
    </div>
  );
}
