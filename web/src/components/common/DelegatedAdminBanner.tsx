import i18next from "i18next";
import {useAccount} from "@/hooks/use-account";
import {getDelegatedAdmin} from "@/lib/delegated-admin";
import {PricingUrl} from "@/lib/setting";

/** Tells a delegated admin why they can't administer anything: the license no longer covers delegated administration. */
export function DelegatedAdminBanner() {
  const {account} = useAccount();
  const info = getDelegatedAdmin(account);
  if (!info || info.active) {
    return null;
  }

  return (
    <div className="border-b border-warning/40 bg-warning/10 px-4 py-2 text-sm text-warning md:px-6">
      {i18next.t("enterprise:You are a delegated admin of this organization, but delegated administration needs an enterprise edition license, so it is turned off")}{" "}
      <a href={PricingUrl} target="_blank" rel="noreferrer" className="font-medium underline underline-offset-2">
        {i18next.t("general:View editions and pricing")}
      </a>
    </div>
  );
}
