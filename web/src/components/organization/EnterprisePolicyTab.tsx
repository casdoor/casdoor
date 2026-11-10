import * as React from "react";
import i18next from "i18next";
import {Plus, Trash2} from "lucide-react";
import {Input} from "@/components/ui/input";
import {Switch} from "@/components/ui/switch";
import {Button} from "@/components/ui/button";
import {MultiSelect} from "@/components/common/MultiSelect";
import {FormRow} from "@/components/crud/FormRow";
import {EditionNotice} from "@/components/common/EditionNotice";
import {useRoleOptions, useUserNameOptions} from "@/hooks/use-options";

const ENCRYPTABLE_FIELDS = ["idCard", "realName", "birthday", "location", "address"];

function getFieldLabel(field: string) {
  switch (field) {
  case "idCard":
    return i18next.t("user:ID card");
  case "realName":
    return i18next.t("application:Real name");
  case "birthday":
    return i18next.t("user:Birthday");
  case "location":
    return i18next.t("user:Location");
  default:
    return i18next.t("user:Address");
  }
}

/** The organization's compliance and governance settings, which need an enterprise edition license. */
export function EnterprisePolicyTab({owner, policy, onChange}: {owner: string; policy: any; onChange: (policy: any) => void}) {
  const value = policy ?? {};
  const set = (key: string, v: any) => onChange({...value, [key]: v});
  const userOptions = useUserNameOptions(owner);
  const roleOptions = useRoleOptions(owner);
  const rules: any[] = value.sodRules ?? [];

  const section = (titleKey: string, descriptionKey?: string) => (
    <div className="col-span-full mt-4 border-b pb-1 first:mt-0">
      <h3 className="text-sm font-semibold">{i18next.t(titleKey)}</h3>
      {descriptionKey ? <p className="mt-0.5 text-xs text-muted-foreground">{i18next.t(descriptionKey)}</p> : null}
    </div>
  );
  const toggle = (key: string, labelKey: string) => (
    <FormRow labelKey={labelKey}>
      <Switch checked={!!value[key]} onCheckedChange={(v) => set(key, v)} />
    </FormRow>
  );
  const admins = (key: string, labelKey: string) => (
    <FormRow labelKey={labelKey}>
      <MultiSelect value={value[key] ?? []} onChange={(v) => set(key, v)} options={userOptions} />
    </FormRow>
  );
  const setRule = (index: number, rule: any) => set("sodRules", rules.map((item, i) => (i === index ? rule : item)));

  return (
    <>
      <div className="col-span-full">
        <EditionNotice edition="Enterprise" />
      </div>

      {section("organization:Separation of admins", "organization:Separation of admins - Tooltip")}
      {toggle("separationOfDuties", "organization:Separate the admins")}
      {value.separationOfDuties ? (
        <>
          {admins("systemAdmins", "organization:System admins")}
          {admins("securityAdmins", "organization:Security admins")}
          {admins("auditAdmins", "organization:Audit admins")}
        </>
      ) : null}

      {section("organization:Audit records", "organization:Tamper-proof records - Tooltip")}
      {toggle("tamperProofRecords", "organization:Tamper-proof records")}

      {section("organization:Personal information", "organization:Encrypted fields - Tooltip")}
      <FormRow labelKey="organization:Encrypted fields">
        <MultiSelect
          value={value.encryptedFields ?? []}
          onChange={(v) => set("encryptedFields", v)}
          options={ENCRYPTABLE_FIELDS.map((field) => ({value: field, label: getFieldLabel(field)}))}
        />
      </FormRow>
      {toggle("recordConsents", "organization:Record consents")}

      {section("organization:Separation of duties", "organization:Separation of duties - Tooltip")}
      <div className="col-span-full space-y-2">
        {rules.map((rule, index) => (
          <div key={index} className="flex flex-wrap items-start gap-2">
            <Input
              className="w-48"
              placeholder={i18next.t("general:Name")}
              value={rule.name ?? ""}
              onChange={(e) => setRule(index, {...rule, name: e.target.value})}
            />
            <MultiSelect
              className="min-w-64 flex-1"
              value={rule.roles ?? []}
              onChange={(v) => setRule(index, {...rule, roles: v})}
              options={roleOptions}
              placeholder={i18next.t("organization:Roles that one user cannot hold together")}
            />
            <Button variant="ghost" size="icon" onClick={() => set("sodRules", rules.filter((_, i) => i !== index))}>
              <Trash2 className="size-4" />
            </Button>
          </div>
        ))}
        <Button variant="outline" size="sm" onClick={() => set("sodRules", [...rules, {name: "", roles: []}])}>
          <Plus className="size-4" />
          {i18next.t("organization:Add rule")}
        </Button>
      </div>
    </>
  );
}
