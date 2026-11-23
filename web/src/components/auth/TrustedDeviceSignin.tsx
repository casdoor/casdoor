// Copyright 2026 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import * as React from "react";
import i18next from "i18next";
import {Laptop} from "lucide-react";
import {Button} from "@/components/ui/button";
import {probeDeviceAgent, type DeviceAgentAccount} from "@/lib/device-agent";

/**
 * Offers to sign in without a password as the user that this computer is enrolled to, when the Casdoor device agent
 * runs here and the organization lets a trusted device sign in.
 */
export function TrustedDeviceSignin({application, onSignin}: {application: any; onSignin: (account: DeviceAgentAccount) => void}) {
  const [account, setAccount] = React.useState<DeviceAgentAccount | null>(null);
  const policy = application?.organizationObj?.deviceTrust;
  const isEnabled = !!policy?.mode && !!policy?.passwordless;

  React.useEffect(() => {
    if (!isEnabled) {
      return;
    }
    probeDeviceAgent().then((accounts) => {
      const found = accounts.find((item) => item.organization === application.organization && item.status === "Trusted" && item.user);
      setAccount(found ?? null);
    });
  }, [isEnabled, application?.organization]);

  if (!account) {
    return null;
  }
  return (
    <div className="trusted-device-signin mb-4 rounded-lg border bg-muted/30 p-3">
      <div className="mb-2 flex items-center gap-2 text-sm">
        <Laptop className="size-4 text-primary" />
        <span className="text-muted-foreground">{i18next.t("login:This computer is a trusted device of")}</span>
        <span className="font-medium">{account.displayName || account.user}</span>
      </div>
      <Button type="button" className="w-full" onClick={() => onSignin(account)}>
        {i18next.t("login:Sign in with this device")}
      </Button>
    </div>
  );
}
