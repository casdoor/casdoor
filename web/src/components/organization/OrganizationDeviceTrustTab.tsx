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

import i18next from "i18next";
import {Switch} from "@/components/ui/switch";
import {MultiSelect} from "@/components/common/MultiSelect";
import {SelectField} from "@/components/common/SelectField";
import {FormRow} from "@/components/crud/FormRow";
import {useApplicationOptions} from "@/hooks/use-options";

/** How the organization trusts the computers enrolled with the Casdoor device agent. */
export function OrganizationDeviceTrustTab({owner, policy, onChange}: {owner: string; policy: any; onChange: (policy: any) => void}) {
  const value = policy ?? {};
  const set = (key: string, v: any) => onChange({...value, [key]: v});
  const applications = useApplicationOptions(owner);
  const toggle = (key: string, labelKey: string) => (
    <FormRow labelKey={labelKey}>
      <Switch checked={!!value[key]} onCheckedChange={(v) => set(key, v)} />
    </FormRow>
  );

  return (
    <>
      <FormRow labelKey="organization:Device trust">
        <SelectField
          value={value.mode || "-"}
          onChange={(v) => set("mode", v === "-" ? "" : v)}
          options={[
            {id: "-", name: i18next.t("general:None")},
            {id: "Signal", name: i18next.t("organization:Untrusted devices add risk")},
            {id: "Required", name: i18next.t("organization:Only trusted devices can sign in")},
          ]}
        />
      </FormRow>
      {value.mode ? (
        <>
          <FormRow labelKey="organization:Device trust applications">
            <MultiSelect
              value={value.applications ?? []}
              onChange={(v) => set("applications", v)}
              options={applications.map((item) => ({value: item.value, label: String(item.value)}))}
            />
          </FormRow>
          {toggle("requireEncrypted", "organization:Require disk encryption")}
          {toggle("requireFirewall", "organization:Require firewall")}
          {toggle("requireLock", "organization:Require screen lock")}
          {toggle("autoApprove", "organization:Trust compliant devices on enrollment")}
          {toggle("passwordless", "organization:Sign in from a trusted device without a password")}
          {toggle("satisfiesMfa", "organization:A trusted device counts as the second factor")}
        </>
      ) : null}
    </>
  );
}
