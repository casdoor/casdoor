// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

import i18next from "i18next";
import {useAccount} from "@/hooks/use-account";
import {getDelegatedAdmin} from "@/ee/lib/delegated-admin";
import * as Conf from "@/Conf";
import {getPricingUrl} from "@/lib/setting";

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
      {Conf.HideUpgrades ? null : (
        <a href={getPricingUrl("delegated-admin", "Enterprise")} target="_blank" rel="noreferrer" className="font-medium underline underline-offset-2">
          {i18next.t("general:View editions and pricing")}
        </a>
      )}
    </div>
  );
}
