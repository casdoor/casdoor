import * as React from "react";
import i18next from "i18next";
import {Plus, Trash2} from "lucide-react";
import {Input} from "@/components/ui/input";
import {Textarea} from "@/components/ui/textarea";
import {Switch} from "@/components/ui/switch";
import {Button} from "@/components/ui/button";
import {MultiSelect} from "@/components/common/MultiSelect";
import {SearchableSelect} from "@/components/common/SearchableSelect";
import {SelectField} from "@/components/common/SelectField";
import {ActionSelect} from "@/components/common/ActionSelect";
import {FormRow} from "@/components/crud/FormRow";
import {EditionNotice} from "@/components/common/EditionNotice";
import {useApplicationOptions, useGroupNameOptions, useRoleOptions, useUserNameOptions} from "@/hooks/use-options";
import * as ServerBackend from "@/backend/ServerBackend";

const ENCRYPTABLE_FIELDS = ["idCard", "realName", "birthday", "location", "address", "managedAccounts"];

const DELEGATED_ACTIONS = ["users", "groups", "applications", "passwords"];

function getDelegatedActionLabel(action: string) {
  switch (action) {
  case "users":
    return i18next.t("organization:Manage users");
  case "groups":
    return i18next.t("organization:Manage groups");
  case "applications":
    return i18next.t("organization:Manage application settings");
  default:
    return i18next.t("organization:Reset passwords");
  }
}

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
  case "managedAccounts":
    return i18next.t("user:Managed account passwords");
  default:
    return i18next.t("user:Address");
  }
}

const CERT_FIELDS = [
  {id: "CN", key: "organization:Subject common name (CN)"},
  {id: "SERIALNUMBER", key: "organization:Subject serial number (SERIALNUMBER)"},
  {id: "Email", key: "organization:Email address in the certificate"},
  {id: "UID", key: "organization:Subject user ID (UID)"},
  {id: "Fingerprint", key: "organization:Certificate SHA-256 fingerprint"},
];

const CERT_USER_FIELDS = [
  {id: "name", key: "general:Name"},
  {id: "idCard", key: "user:ID card"},
  {id: "email", key: "general:Email"},
  {id: "phone", key: "general:Phone"},
  {id: "certFingerprint", key: "organization:Bound certificate fingerprint"},
];

/** Signing in with a certificate in a UKey: the trusted CAs and how a certificate is matched to a user. */
function CertLoginSection({owner, value, onChange}: {owner: string; value: any; onChange: (value: any) => void}) {
  const certLogin = value ?? {};
  const set = (key: string, v: any) => onChange({certField: "CN", userField: "name", ...certLogin, [key]: v});
  return (
    <>
      <FormRow labelKey="organization:Enable certificate login">
        <Switch checked={!!certLogin.enabled} onCheckedChange={(v) => set("enabled", v)} />
      </FormRow>
      {certLogin.enabled ? (
        <>
          <FormRow labelKey="organization:Trusted CA certificates" block>
            <Textarea
              className="min-h-32 font-mono text-xs"
              placeholder="-----BEGIN CERTIFICATE-----"
              value={certLogin.trustedCas ?? ""}
              onChange={(e) => set("trustedCas", e.target.value)}
            />
          </FormRow>
          <FormRow labelKey="organization:Certificate field">
            <SelectField
              value={certLogin.certField ?? "CN"}
              onChange={(v) => set("certField", v)}
              options={CERT_FIELDS.map((item) => ({id: item.id, name: i18next.t(item.key)}))}
            />
          </FormRow>
          <FormRow labelKey="organization:Matched user attribute">
            <SelectField
              value={certLogin.userField ?? "name"}
              onChange={(v) => set("userField", v)}
              options={CERT_USER_FIELDS.map((item) => ({id: item.id, name: i18next.t(item.key)}))}
            />
          </FormRow>
          <FormRow labelKey="organization:CRL URL">
            <Input value={certLogin.crlUrl ?? ""} placeholder="http://ca.example.com/crl/ca.crl" onChange={(e) => set("crlUrl", e.target.value)} />
          </FormRow>
          <FormRow labelKey="organization:Certificate login URL">
            <Input value={certLogin.loginUrl ?? ""} placeholder="https://door.example.com:18444/api/login/cert" onChange={(e) => set("loginUrl", e.target.value)} />
          </FormRow>
          <FormRow labelKey="organization:Certificate verifier">
            <ActionSelect owner={owner} kind="cert" allowNone value={certLogin.verifier ?? ""} onChange={(v) => set("verifier", v)} />
          </FormRow>
          <div className="col-span-full space-y-1 rounded-md border bg-muted/40 p-3 text-xs text-muted-foreground">
            <p>{i18next.t("organization:Certificate login - Help")}</p>
            <p>{i18next.t("organization:UKey plug-in sign-in - Help")}</p>
          </div>
        </>
      ) : null}
    </>
  );
}

/** The organization's compliance and governance settings, which need an enterprise edition license. */
export function EnterprisePolicyTab({owner, policy, onChange}: {owner: string; policy: any; onChange: (policy: any) => void}) {
  const value = policy ?? {};
  const set = (key: string, v: any) => onChange({...value, [key]: v});
  const userOptions = useUserNameOptions(owner);
  const roleOptions = useRoleOptions(owner);
  const rules: any[] = value.sodRules ?? [];
  const groupOptions = useGroupNameOptions(owner);
  const applicationOptions = useApplicationOptions(owner);
  const delegatedAdmins: any[] = value.delegatedAdmins ?? [];
  const setDelegatedAdmin = (index: number, item: any) =>
    set("delegatedAdmins", delegatedAdmins.map((current, i) => (i === index ? item : current)));

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
  const [toolOptions, setToolOptions] = React.useState<{value: string; label: string}[]>([]);
  React.useEffect(() => {
    if (!owner) {
      return;
    }
    ServerBackend.getServers(owner, 1, 1000).then((res: any) => {
      if (res.status === "ok") {
        setToolOptions((res.data ?? []).flatMap((server: any) =>
          (server.tools ?? []).map((tool: any) => ({value: `${server.name}/${tool.name}`, label: `${server.name}/${tool.name}`}))));
      }
    });
  }, [owner]);

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

      {section("organization:Certificate and UKey login", "organization:Certificate and UKey login - Tooltip")}
      <CertLoginSection owner={owner} value={value.certLogin} onChange={(v) => set("certLogin", v)} />

      {section("organization:Separation of duties", "organization:Separation of duties - Tooltip")}
      <div className="col-span-full space-y-2" data-tour="sod-rules">
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
            <MultiSelect
              className="min-w-64 flex-1"
              value={rule.tools ?? []}
              onChange={(v) => setRule(index, {...rule, tools: v})}
              options={toolOptions}
              placeholder={i18next.t("organization:MCP tools that one agent cannot be allowed together")}
            />
            <Button variant="ghost" size="icon" onClick={() => set("sodRules", rules.filter((_, i) => i !== index))}>
              <Trash2 className="size-4" />
            </Button>
          </div>
        ))}
        <Button variant="outline" size="sm" onClick={() => set("sodRules", [...rules, {name: "", roles: [], tools: []}])}>
          <Plus className="size-4" />
          {i18next.t("organization:Add rule")}
        </Button>
      </div>

      {section("edition:Delegated administration", "organization:Delegated administration - Tooltip")}
      <div className="col-span-full space-y-3" data-tour="delegated-admins">
        {delegatedAdmins.map((item, index) => (
          <div key={index} className="flex flex-wrap items-start gap-2 rounded-md border p-2">
            <SearchableSelect
              className="w-48"
              value={item.user ?? ""}
              onChange={(v) => setDelegatedAdmin(index, {...item, user: v})}
              options={userOptions}
              placeholder={i18next.t("general:User")}
            />
            <MultiSelect
              className="min-w-56 flex-1"
              value={item.groups ?? []}
              onChange={(v) => setDelegatedAdmin(index, {...item, groups: v})}
              options={groupOptions}
              placeholder={i18next.t("organization:Groups they administer, with subgroups")}
            />
            <MultiSelect
              className="min-w-48 flex-1"
              value={item.applications ?? []}
              onChange={(v) => setDelegatedAdmin(index, {...item, applications: v})}
              options={applicationOptions}
              placeholder={i18next.t("organization:Applications they administer")}
            />
            <MultiSelect
              className="min-w-48 flex-1"
              value={item.actions ?? []}
              onChange={(v) => setDelegatedAdmin(index, {...item, actions: v})}
              options={DELEGATED_ACTIONS.map((action) => ({value: action, label: getDelegatedActionLabel(action)}))}
              placeholder={i18next.t("organization:What they may do")}
            />
            <Button variant="ghost" size="icon" onClick={() => set("delegatedAdmins", delegatedAdmins.filter((_, i) => i !== index))}>
              <Trash2 className="size-4" />
            </Button>
          </div>
        ))}
        <Button
          variant="outline"
          size="sm"
          onClick={() => set("delegatedAdmins", [...delegatedAdmins, {user: "", groups: [], applications: [], actions: ["users", "groups"]}])}
        >
          <Plus className="size-4" />
          {i18next.t("organization:Add delegated admin")}
        </Button>
      </div>
    </>
  );
}
