// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

import * as React from "react";
import i18next from "i18next";
import {Input} from "@/components/ui/input";
import {Switch} from "@/components/ui/switch";
import {TagsInput} from "@/components/common/TagsInput";
import {MultiSelect} from "@/components/common/MultiSelect";
import {SelectField} from "@/components/common/SelectField";
import * as IntegrationBackend from "@/backend/IntegrationBackend";
import {getActionKind, parseBundle} from "@/lib/integration";
import {FormRow} from "@/components/crud/FormRow";
import {EditionNotice} from "@/components/common/EditionNotice";
import * as Setting from "@/lib/setting";
import {RiskSignals} from "@/lib/risk-signals";


/** The organization's security and risk control settings, which need a commercial edition license. */
export function SecurityPolicyTab({owner, policy, onChange}: {owner: string; policy: any; onChange: (policy: any) => void}) {
  const value = policy ?? {};
  const set = (key: string, v: any) => onChange({...value, [key]: v});

  // the password checks and migrations are actions installed from the Marketplace
  const [actions, setActions] = React.useState<any[]>([]);
  React.useEffect(() => {
    if (!owner) {
      return;
    }
    IntegrationBackend.getIntegrations(owner).then((res: any) => {
      if (res.status === "ok") {
        setActions((res.data ?? []).filter((integration: any) => integration.type === "action"));
      }
    });
  }, [owner]);
  const actionOptions = (kind: string) => actions
    .filter((integration: any) => (integration.kind || getActionKind(parseBundle(integration.bundle))) === kind)
    .map((integration: any) => ({value: integration.name, label: `${integration.displayName} (${integration.name})`}));

  const number = (key: string, labelKey: string, placeholder?: string) => (
    <FormRow labelKey={labelKey}>
      <Input
        type="number"
        min={0}
        placeholder={placeholder ?? "0"}
        value={value[key] || ""}
        onChange={(e) => set(key, Setting.myParseInt(e.target.value))}
      />
    </FormRow>
  );
  const toggle = (key: string, labelKey: string) => (
    <FormRow labelKey={labelKey}>
      <Switch checked={!!value[key]} onCheckedChange={(v) => set(key, v)} />
    </FormRow>
  );
  const section = (titleKey: string) => (
    <h3 className="col-span-full mt-4 border-b pb-1 text-sm font-semibold first:mt-0">{i18next.t(titleKey)}</h3>
  );

  return (
    <>
      <div className="col-span-full">
        <EditionNotice />
      </div>

      {section("organization:Sessions and accounts")}
      {number("inactiveTimeoutMinutes", "organization:Inactive timeout minutes")}
      {number("disableInactiveUserDays", "organization:Disable inactive users after days")}

      {section("organization:Passwords and credential stuffing")}
      <FormRow labelKey="organization:Password check plugins">
        <MultiSelect
          value={value.passwordCheckPlugins ?? []}
          onChange={(v) => set("passwordCheckPlugins", v)}
          options={actionOptions("password-check")}
        />
      </FormRow>
      {number("stuffingFailureThreshold", "organization:Credential stuffing threshold")}
      {number("stuffingBlockMinutes", "organization:Credential stuffing block minutes", "60")}

      {section("organization:Sign-in alerts")}
      {toggle("alertOnNewLocation", "organization:Alert on new location")}
      {toggle("alertOnNewDevice", "organization:Alert on new device")}

      {section("organization:Adaptive risk")}
      {toggle("adaptiveRisk", "organization:Enable adaptive risk")}
      {value.adaptiveRisk ? (
        <>
          {number("riskMfaScore", "organization:Ask for MFA at score", "40")}
          {number("riskDenyScore", "organization:Block at score", "100")}
          {number("impossibleTravelHours", "organization:Impossible travel within hours", "2")}
          <FormRow labelKey="organization:Signal scores" block>
            <div className="grid gap-2 sm:grid-cols-3">
              {RiskSignals.map((signal) => (
                <label key={signal.name} className="flex items-center justify-between gap-2 rounded-md border px-3 py-1.5 text-sm">
                  <span>{i18next.t(signal.labelKey)}</span>
                  <Input
                    type="number"
                    min={0}
                    max={100}
                    className="h-8 w-20"
                    placeholder={String(signal.score)}
                    value={value.riskScores?.[signal.name] ?? ""}
                    onChange={(e) => {
                      const scores = {...(value.riskScores ?? {})};
                      if (e.target.value === "") {
                        delete scores[signal.name];
                      } else {
                        scores[signal.name] = Setting.myParseInt(e.target.value);
                      }
                      set("riskScores", scores);
                    }}
                  />
                </label>
              ))}
            </div>
          </FormRow>
          <FormRow labelKey="organization:Trusted networks" block>
            <TagsInput value={value.riskTrustedNetworks ?? []} onChange={(v) => set("riskTrustedNetworks", v)} placeholder="10.0.0.0/8" />
          </FormRow>
          <FormRow labelKey="organization:Blocked countries" block>
            <TagsInput value={value.riskBlockedCountries ?? []} onChange={(v) => set("riskBlockedCountries", v)} placeholder="KP" />
          </FormRow>
          <FormRow labelKey="organization:Country header">
            <Input placeholder="CF-IPCountry" value={value.riskCountryHeader ?? ""} onChange={(e) => set("riskCountryHeader", e.target.value)} />
          </FormRow>
        </>
      ) : null}

      {section("organization:Personal information")}
      {toggle("maskSensitiveData", "organization:Mask personal information in the console")}
      {toggle("allowDataExport", "organization:Allow users to export their data")}
      {toggle("allowAccountDeletion", "organization:Allow users to delete their accounts")}

      {section("organization:Members and SMS")}
      {toggle("mergeByWechatUnionId", "organization:Merge members by WeChat unionid")}
      {toggle("smsDomesticOnly", "organization:Mainland China numbers only")}
      {number("smsPerPhoneHourly", "organization:SMS per number per hour")}
      {number("smsPerPhoneDaily", "organization:SMS per number per day")}
      {number("smsPerPrefixHourly", "organization:SMS per number segment per hour")}
      {number("smsDailyAlertCount", "organization:SMS daily alert count")}
      {number("smsDailyLimit", "organization:SMS daily limit")}

      {section("organization:Migration")}
      <FormRow labelKey="organization:Password migration plugin">
        <SelectField
          value={value.passwordMigrationPlugin || "-"}
          onChange={(v) => set("passwordMigrationPlugin", v === "-" ? "" : v)}
          options={[{id: "-", name: i18next.t("general:None")}, ...actionOptions("password-migration").map((item) => ({id: item.value, name: item.label}))]}
        />
      </FormRow>
      <FormRow labelKey="organization:Password migration URL">
        <Input
          placeholder="https://example.authing.cn"
          value={value.passwordMigrationUrl ?? ""}
          onChange={(e) => set("passwordMigrationUrl", e.target.value)}
        />
      </FormRow>
      <FormRow labelKey="organization:Password migration secret">
        <Input
          type="password"
          value={value.passwordMigrationSecret ?? ""}
          onChange={(e) => set("passwordMigrationSecret", e.target.value)}
        />
      </FormRow>

      {section("organization:Certificates and agents")}
      {number("certExpireAlertDays", "organization:Cert expire alert days", "30")}
      {toggle("autoRotateCerts", "organization:Rotate signing keys automatically")}
      {toggle("recordMcpToolCalls", "organization:Record MCP tool calls")}

      {section("organization:Alerts")}
      <FormRow labelKey="organization:Alert emails" block>
        <TagsInput value={value.alertEmails ?? []} onChange={(v) => set("alertEmails", v)} />
      </FormRow>
    </>
  );
}
