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
