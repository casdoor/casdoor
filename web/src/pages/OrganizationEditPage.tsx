import * as React from "react";
import i18next from "i18next";
import {Link, useNavigate, useParams} from "react-router-dom";
import {Link as LinkIcon} from "lucide-react";
import {UnauthorizedPage} from "@/components/common/UnauthorizedPage";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Switch} from "@/components/ui/switch";
import {Tabs, TabsContent, TabsList, TabsTrigger} from "@/components/ui/tabs";
import {Loading} from "@/components/common/Loading";
import {MultiSelect} from "@/components/common/MultiSelect";
import {NavItemTree, WidgetItemTree} from "@/components/common/NavItemTree";
import {SearchableSelect} from "@/components/common/SearchableSelect";
import {SelectField} from "@/components/common/SelectField";
import {TagsInput} from "@/components/common/TagsInput";
import {ThemeEditor} from "@/components/common/ThemeEditor";
import {EditPageShell} from "@/components/crud/EditPageShell";
import {EditableTable} from "@/components/crud/EditableTable";
import {FormRow, formGridClass} from "@/components/crud/FormRow";
import {useAccount} from "@/hooks/use-account";
import {useEditRecord} from "@/hooks/use-edit-record";
import {getModeTitleKey, submitEdit} from "@/lib/crud";
import * as Obfuscator from "@/auth/Obfuscator";
import * as ApplicationBackend from "@/backend/ApplicationBackend";
import * as LdapBackend from "@/backend/LdapBackend";
import {ConfirmButton} from "@/components/common/ConfirmButton";
import * as OrganizationBackend from "@/backend/OrganizationBackend";
import * as Setting from "@/lib/setting";

const PASSWORD_TYPES = ["plain", "salt", "sha512-salt", "md5-salt", "bcrypt", "pbkdf2-salt", "argon2id", "pbkdf2-django"];
const TOKEN_FORMATS = ["JWT", "JWT-Empty", "JWT-Custom", "JWT-Standard"];
const OBFUSCATOR_TYPES = ["Plain", "AES", "DES"];
const VIEW_RULES = ["Public", "Self", "Admin"];
const MODIFY_RULES = ["Self", "Admin", "Immutable"];
/** an item only an admin may see is not one the user can be allowed to modify */
const ADMIN_MODIFY_RULES = ["Admin", "Immutable"];
/** `general:Optional` and friends do not exist; the antd table uses these. */
const MFA_RULES: Record<string, string> = {
  "Optional": "organization:Optional",
  "Prompted": "organization:Prompt",
  "Required": "organization:Required",
};

function passwordOptions() {
  return [
    {value: "AtLeast6", label: i18next.t("user:The password must have at least 6 characters")},
    {value: "AtLeast8", label: i18next.t("user:The password must have at least 8 characters")},
    {
      value: "Aa123",
      label: i18next.t(
        "user:The password must contain at least one uppercase letter, one lowercase letter and one digit",
      ),
    },
    {value: "SpecialChar", label: i18next.t("user:The password must contain at least one special character")},
    {value: "NoRepeat", label: i18next.t("user:The password must not contain any repeated characters")},
  ];
}

export default function OrganizationEditPage() {
  const {organizationName = ""} = useParams();
  const navigate = useNavigate();
  const {account} = useAccount();
  const [saving, setSaving] = React.useState(false);
  const [applications, setApplications] = React.useState<any[]>([]);
  const [ldaps, setLdaps] = React.useState<any[] | null>(null);

  const {record: organization, setRecord, updateField, loading, denied, mode, setMode} = useEditRecord<any>({
    fetch: () => OrganizationBackend.getOrganization("admin", organizationName),
    transform: (org) => ({...org, enableDarkLogo: !!org.logoDark}),
    deps: [organizationName],
  });

  const loadRelated = React.useCallback(() => {
    ApplicationBackend.getApplicationsByOrganization("admin", organizationName).then((res: any) => {
      if (res.status === "ok") {
        setApplications(res.data ?? []);
      }
    });
    LdapBackend.getLdaps(organizationName).then((res: any) => {
      setLdaps(res.status === "ok" ? res.data ?? [] : []);
    });
  }, [organizationName]);

  React.useEffect(() => {
    if (mode === "add") {
      setLdaps([]);
      return;
    }
    loadRelated();
  }, [mode, loadRelated]);

  const deleteLdap = (ldap: any) =>
    LdapBackend.deleteLdap(ldap)
      .then((res: any) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully deleted"));
          setLdaps((prev) => (prev ?? []).filter((item) => item.id !== ldap.id));
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to delete")}: ${res.msg}`);
        }
      })
      .catch((error) => Setting.showMessage("error", `${i18next.t("general:Failed to delete")}: ${error}`));

  if (denied) {
    return <UnauthorizedPage />;
  }

  if (loading || organization === null) {
    return <Loading />;
  }

  const update = updateField;

  const updatePasswordObfuscator = (key: "type" | "key", value: string) => {
    if (key === "type") {
      // a new type needs a key of its own length, so antd generates one for it
      setRecord((prev: any) => ({
        ...prev,
        passwordObfuscatorType: value,
        passwordObfuscatorKey: Obfuscator.getRandomKeyForObfuscator(value),
      }));
    } else {
      update("passwordObfuscatorKey", value);
    }
  };

  const save = async(exitAfterSave: boolean) => {
    const payload = Setting.deepCopy(organization);
    payload.accountItems = payload.accountItems?.filter(
      (item: any) => item.name !== "Please select an account item",
    );

    // a key that does not match its obfuscator would break every password sign-in
    // of the organization, and the backend does not re-check it
    const obfuscatorError = Obfuscator.checkPasswordObfuscator(
      payload.passwordObfuscatorType,
      payload.passwordObfuscatorKey,
    );
    if (obfuscatorError.length > 0) {
      Setting.showMessage("error", obfuscatorError);
      return;
    }

    setSaving(true);
    await submitEdit({
      mode,
      record: payload,
      add: (record) => OrganizationBackend.addOrganization(record),
      update: (record) => OrganizationBackend.updateOrganization(organization.owner, organizationName, record),
      onSaved: () => {
        window.dispatchEvent(new Event("storageOrganizationsChanged"));
        setMode("edit");
        if (exitAfterSave) {
          navigate("/organizations");
        } else if (organization.name !== organizationName) {
          navigate(`/organizations/${organization.name}`, {replace: true});
        }
      },
      onFailed: () => {
        if (mode !== "add") {
          update("name", organizationName);
        }
      },
    });
    setSaving(false);
  };

  return (
    <EditPageShell
      title={`${i18next.t(getModeTitleKey("organization:Edit Organization", mode))} - ${organization.displayName || organization.name}`}
      mode={mode}
      backTo="/organizations"
      onSave={save}
      saving={saving}
    >
      <Tabs defaultValue="basic">
        <TabsList className="mb-2 flex-wrap">
          <TabsTrigger value="basic">{i18next.t("application:Basic")}</TabsTrigger>
          <TabsTrigger value="password">{i18next.t("general:Password type")}</TabsTrigger>
          <TabsTrigger value="account">{i18next.t("organization:Account items")}</TabsTrigger>
          <TabsTrigger value="advanced">{i18next.t("provider:Advanced")}</TabsTrigger>
        </TabsList>

        <TabsContent value="basic" className={formGridClass}>
          <FormRow labelKey="general:Name">
            <Input
              value={organization.name ?? ""}
              disabled={organization.name === "built-in"}
              onChange={(e) => update("name", e.target.value)}
            />
          </FormRow>
          <FormRow labelKey="general:Display name">
            <Input value={organization.displayName ?? ""} onChange={(e) => update("displayName", e.target.value)} />
          </FormRow>
          <FormRow labelKey="general:Logo">
            <div className="space-y-2">
              <div className="relative">
                <LinkIcon className="pointer-events-none absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                  className="pl-8"
                  value={organization.logo ?? ""}
                  onChange={(e) => update("logo", e.target.value)}
                />
              </div>
              {organization.logo ? (
                <a href={organization.logo} target="_blank" rel="noreferrer">
                  <img src={organization.logo} alt="logo" className="h-14 max-w-[240px] rounded-md border bg-white object-contain p-1.5" />
                </a>
              ) : null}
            </div>
          </FormRow>
          <FormRow labelKey="general:Enable dark logo">
            <Switch
              checked={!!organization.enableDarkLogo}
              onCheckedChange={(checked) => {
                update("enableDarkLogo", checked);
                if (!checked) {
                  update("logoDark", "");
                }
              }}
            />
          </FormRow>
          {organization.enableDarkLogo ? (
            <FormRow labelKey="general:Logo dark">
              <div className="space-y-2">
                <Input value={organization.logoDark ?? ""} onChange={(e) => update("logoDark", e.target.value)} />
                {organization.logoDark ? (
                  <a href={organization.logoDark} target="_blank" rel="noreferrer">
                    <img
                      src={organization.logoDark}
                      alt="logo dark"
                      className="h-14 max-w-[240px] rounded-md border bg-black object-contain p-1.5"
                    />
                  </a>
                ) : null}
              </div>
            </FormRow>
          ) : null}
          <FormRow labelKey="general:Favicon">
            <div className="space-y-2">
              <Input value={organization.favicon ?? ""} onChange={(e) => update("favicon", e.target.value)} />
              {organization.favicon ? (
                <a href={organization.favicon} target="_blank" rel="noreferrer">
                  <img src={organization.favicon} alt="favicon" className="h-10 w-10 object-contain" />
                </a>
              ) : null}
            </div>
          </FormRow>
          <FormRow labelKey="organization:Website URL">
            <Input value={organization.websiteUrl ?? ""} onChange={(e) => update("websiteUrl", e.target.value)} />
          </FormRow>
          <FormRow labelKey="general:Default avatar">
            <div className="space-y-2">
              <Input
                value={organization.defaultAvatar ?? ""}
                onChange={(e) => update("defaultAvatar", e.target.value)}
              />
              {organization.defaultAvatar ? (
                <a href={organization.defaultAvatar} target="_blank" rel="noreferrer">
                  <img src={organization.defaultAvatar} alt="default avatar" className="h-20 w-20 rounded-full object-cover" />
                </a>
              ) : null}
            </div>
          </FormRow>
          <FormRow labelKey="general:Default application">
            <SearchableSelect
              value={organization.defaultApplication ?? ""}
              onChange={(value) => update("defaultApplication", value)}
              options={applications.map((app) => ({value: app.name, label: app.displayName || app.name}))}
            />
          </FormRow>
          <FormRow labelKey="general:Languages">
            <MultiSelect
              value={organization.languages ?? []}
              onChange={(value) => update("languages", value)}
              options={(Setting.Countries as any[]).map((country) => ({
                value: country.key,
                label: country.label,
                keywords: country.alt,
              }))}
            />
          </FormRow>
          <FormRow labelKey="general:Supported country codes">
            <MultiSelect
              value={organization.countryCodes ?? []}
              onChange={(value) => update("countryCodes", value)}
              options={[
                {value: "All", label: i18next.t("general:All")},
                ...Setting.getCountryCodeData().map((country: any) => ({
                  value: country.code,
                  label: `${country.name} (+${country.phone})`,
                  keywords: `${country.name} ${country.code} ${country.phone}`,
                })),
              ]}
            />
          </FormRow>
          <FormRow labelKey="organization:User types">
            <TagsInput value={organization.userTypes ?? []} onChange={(value) => update("userTypes", value)} />
          </FormRow>
          <FormRow labelKey="organization:Tags">
            <TagsInput value={organization.tags ?? []} onChange={(value) => update("tags", value)} />
          </FormRow>
          <FormRow labelKey="general:Enable tour">
            <Switch checked={!!organization.enableTour} onCheckedChange={(v) => update("enableTour", v)} />
          </FormRow>
          <FormRow labelKey="application:Disable signin">
            <Switch checked={!!organization.disableSignin} onCheckedChange={(v) => update("disableSignin", v)} />
          </FormRow>
          <FormRow labelKey="organization:Disable console">
            <Switch checked={!!organization.disableConsole} onCheckedChange={(v) => update("disableConsole", v)} />
          </FormRow>
        </TabsContent>

        <TabsContent value="password" className={formGridClass}>
          <FormRow labelKey="general:Password type">
            <SelectField
              value={organization.passwordType}
              onChange={(value) => update("passwordType", value)}
              options={PASSWORD_TYPES.map((item) => ({id: item, name: item}))}
            />
          </FormRow>
          <FormRow labelKey="general:Password salt">
            <Input value={organization.passwordSalt ?? ""} onChange={(e) => update("passwordSalt", e.target.value)} />
          </FormRow>
          <FormRow labelKey="general:Password complexity options">
            <MultiSelect
              value={organization.passwordOptions ?? []}
              onChange={(value) => update("passwordOptions", value)}
              options={passwordOptions()}
            />
          </FormRow>
          <FormRow labelKey="general:Password obfuscator">
            <SelectField
              value={organization.passwordObfuscatorType || "Plain"}
              onChange={(value) => updatePasswordObfuscator("type", value)}
              options={OBFUSCATOR_TYPES.map((item) => ({id: item, name: item}))}
            />
          </FormRow>
          {organization.passwordObfuscatorType && organization.passwordObfuscatorType !== "Plain" ? (
            <FormRow labelKey="general:Password obf key">
              <Input
                value={organization.passwordObfuscatorKey ?? ""}
                onChange={(e) => updatePasswordObfuscator("key", e.target.value)}
              />
            </FormRow>
          ) : null}
          <FormRow labelKey="general:Master password">
            <Input
              type="password"
              value={organization.masterPassword ?? ""}
              onChange={(e) => update("masterPassword", e.target.value)}
            />
          </FormRow>
          <FormRow labelKey="general:Default password">
            <Input
              value={organization.defaultPassword ?? ""}
              onChange={(e) => update("defaultPassword", e.target.value)}
            />
          </FormRow>
          <FormRow block labelKey="general:Master verification code">
            <Input
              value={organization.masterVerificationCode ?? ""}
              onChange={(e) => update("masterVerificationCode", e.target.value)}
            />
          </FormRow>
          <FormRow block labelKey="organization:Password expire days">
            <Input
              type="number"
              value={organization.passwordExpireDays ?? 0}
              onChange={(e) => update("passwordExpireDays", Setting.myParseInt(e.target.value))}
            />
          </FormRow>
          <FormRow block labelKey="application:MFA remember time">
            <Input
              type="number"
              value={organization.mfaRememberInHours ?? 12}
              onChange={(e) => update("mfaRememberInHours", Setting.myParseInt(e.target.value))}
            />
          </FormRow>
          <FormRow block labelKey="general:MFA items">
            <EditableTable
              rows={organization.mfaItems ?? []}
              onChange={(rows) => update("mfaItems", rows)}
              newRow={() => ({name: "Email", rule: "Optional"})}
              columns={[
                {
                  key: "name",
                  title: i18next.t("general:Name"),
                  width: 220,
                  render: (row: any, _index, patch) => (
                    <SelectField
                      value={row.name}
                      onChange={(value) => patch({name: value})}
                      options={["Email", "SMS", "TOTP"].map((item) => ({id: item, name: item}))}
                    />
                  ),
                },
                {
                  key: "rule",
                  title: i18next.t("application:Rule"),
                  width: 220,
                  render: (row: any, _index, patch) => (
                    <SelectField
                      value={row.rule}
                      onChange={(value) => {
                        // exactly one factor may be mandatory, as the antd table enforces
                        const required = (organization.mfaItems ?? []).filter(
                          (item: any) => item.rule === "Required",
                        ).length;
                        if (value === "Required" && required >= 1 && row.rule !== "Required") {
                          Setting.showMessage("error", i18next.t("general:Only 1 MFA method can be required"));
                          return;
                        }
                        patch({rule: value});
                      }}
                      options={Object.entries(MFA_RULES).map(([id, key]) => ({id, name: i18next.t(key)}))}
                    />
                  ),
                },
              ]}
            />
          </FormRow>
        </TabsContent>

        <TabsContent value="account" className={formGridClass}>
          <FormRow labelKey="organization:Account items" block>
            <EditableTable
              rows={organization.accountItems ?? []}
              onChange={(rows) => update("accountItems", rows)}
              newRow={() => ({name: "Please select an account item", visible: true, viewRule: "Public", modifyRule: "Self"})}
              rowKey={(row: any, index) => `${row.name}-${index}`}
              columns={[
                {
                  key: "name",
                  title: i18next.t("general:Name"),
                  width: 260,
                  render: (row: any) => <span className="text-sm">{row.name}</span>,
                },
                {
                  key: "visible",
                  title: i18next.t("organization:Visible"),
                  width: 100,
                  render: (row: any, _index, patch) => (
                    <Switch checked={!!row.visible} onCheckedChange={(v) => patch({visible: v})} />
                  ),
                },
                {
                  key: "viewRule",
                  title: i18next.t("organization:View rule"),
                  width: 160,
                  render: (row: any, _index, patch) => (
                    <SelectField
                      value={row.viewRule}
                      disabled={!row.visible}
                      onChange={(value) => patch({viewRule: value})}
                      options={VIEW_RULES.map((item) => ({id: item, name: item}))}
                    />
                  ),
                },
                {
                  key: "modifyRule",
                  title: i18next.t("organization:Modify rule"),
                  width: 160,
                  render: (row: any, _index, patch) => (
                    <SelectField
                      value={row.modifyRule}
                      disabled={!row.visible}
                      onChange={(value) => patch({modifyRule: value})}
                      options={(row.viewRule === "Admin" || row.name === "Is admin" ? ADMIN_MODIFY_RULES : MODIFY_RULES)
                        .map((item) => ({id: item, name: item}))}
                    />
                  ),
                },
              ]}
            />
          </FormRow>
        </TabsContent>

        <TabsContent value="advanced" className={formGridClass}>
          <FormRow labelKey="organization:Default token format">
            <SelectField
              value={organization.defaultTokenFormat || "JWT"}
              onChange={(value) => update("defaultTokenFormat", value)}
              options={TOKEN_FORMATS.map((item) => ({id: item, name: item}))}
            />
          </FormRow>
          <FormRow labelKey="organization:Is profile public">
            <Switch checked={!!organization.isProfilePublic} onCheckedChange={(v) => update("isProfilePublic", v)} />
          </FormRow>
          <FormRow labelKey="organization:Use Email as username">
            <Switch
              checked={!!organization.useEmailAsUsername}
              onCheckedChange={(v) => update("useEmailAsUsername", v)}
            />
          </FormRow>
          <FormRow labelKey="organization:Use permanent avatar">
            <Switch
              checked={!!organization.usePermanentAvatar}
              onCheckedChange={(v) => update("usePermanentAvatar", v)}
            />
          </FormRow>
          <FormRow labelKey="organization:Soft deletion">
            <Switch
              checked={!!organization.enableSoftDeletion}
              onCheckedChange={(v) => update("enableSoftDeletion", v)}
            />
          </FormRow>
          <FormRow labelKey="organization:Has privilege consent">
            {/* granting it is the dangerous direction, so only that one asks */}
            {organization.hasPrivilegeConsent ? (
              <Switch checked onCheckedChange={(v) => update("hasPrivilegeConsent", v)} />
            ) : (
              <ConfirmButton
                variant="ghost"
                size="iconSm"
                destructive={false}
                title={i18next.t("organization:Has privilege consent warning")}
                onConfirm={() => update("hasPrivilegeConsent", true)}
              >
                <Switch checked={false} className="pointer-events-none" />
              </ConfirmButton>
            )}
          </FormRow>
          <FormRow labelKey="general:IP whitelist">
            <Input
              value={organization.ipWhitelist ?? ""}
              onChange={(e) => update("ipWhitelist", e.target.value)}
            />
          </FormRow>
          <FormRow labelKey="organization:Init score">
            <Input
              type="number"
              value={organization.initScore ?? 0}
              onChange={(e) => update("initScore", Setting.myParseInt(e.target.value))}
            />
          </FormRow>
          <FormRow labelKey="organization:Balance currency">
            <SelectField
              value={organization.balanceCurrency || "USD"}
              onChange={(value) => update("balanceCurrency", value)}
              options={(Setting.CurrencyOptions as any[]).map((item) => ({id: item.id, name: item.name}))}
            />
          </FormRow>
          <FormRow labelKey="organization:Balance credit">
            <Input
              type="number"
              value={organization.balanceCredit ?? 0}
              onChange={(e) => update("balanceCredit", Number(e.target.value))}
            />
          </FormRow>
          <FormRow labelKey="organization:Record retention days">
            <Input
              type="number"
              value={organization.recordRetentionDays ?? 0}
              onChange={(e) => update("recordRetentionDays", Setting.myParseInt(e.target.value))}
            />
          </FormRow>
          <FormRow labelKey="organization:Token retention days">
            <Input
              type="number"
              value={organization.tokenRetentionDays ?? 0}
              onChange={(e) => update("tokenRetentionDays", Setting.myParseInt(e.target.value))}
            />
          </FormRow>
          <FormRow labelKey="organization:Org balance">
            <Input
              type="number"
              value={organization.orgBalance ?? 0}
              onChange={(e) => update("orgBalance", Number(e.target.value))}
            />
          </FormRow>
          <FormRow block labelKey="organization:User balance">
            {/* maintained by the backend, the antd page shows it read-only too */}
            <Input type="number" value={organization.userBalance ?? 0} disabled />
          </FormRow>
          <FormRow block labelKey="organization:Account menu">
            <SelectField
              value={organization.accountMenu || "Horizontal"}
              onChange={(value) => update("accountMenu", value)}
              options={[
                {id: "Horizontal", name: i18next.t("application:Horizontal")},
                {id: "Vertical", name: i18next.t("application:Vertical")},
              ]}
            />
          </FormRow>
          <FormRow block labelKey="organization:LDAP attributes">
            <MultiSelect
              creatable
              value={organization.ldapAttributes ?? []}
              onChange={(value) => update("ldapAttributes", value)}
              options={Setting.getUserCommonFields().map((item: string) => ({value: item, label: item}))}
            />
          </FormRow>
          <FormRow labelKey="organization:Admin navbar items" block>
            <NavItemTree
              disabled={!Setting.isAdminUser(account)}
              checkedKeys={organization.navItems ?? ["all"]}
              onCheck={(value) => update("navItems", value)}
            />
          </FormRow>
          <FormRow labelKey="organization:User navbar items" block>
            <NavItemTree
              disabled={!Setting.isAdminUser(account)}
              checkedKeys={organization.userNavItems ?? []}
              onCheck={(value) => update("userNavItems", value)}
            />
          </FormRow>
          <FormRow labelKey="organization:Widget items" block>
            <WidgetItemTree
              disabled={!Setting.isAdminUser(account)}
              checkedKeys={organization.widgetItems ?? ["all"]}
              onCheck={(value) => update("widgetItems", value)}
            />
          </FormRow>
          <FormRow block labelKey="organization:Kerberos realm">
            <Input
              value={organization.kerberosRealm ?? ""}
              onChange={(e) => update("kerberosRealm", e.target.value)}
            />
          </FormRow>
          <FormRow block labelKey="organization:Kerberos KDC host">
            <Input
              value={organization.kerberosKdcHost ?? ""}
              onChange={(e) => update("kerberosKdcHost", e.target.value)}
            />
          </FormRow>
          <FormRow block labelKey="organization:Kerberos service name">
            <Input
              value={organization.kerberosServiceName ?? ""}
              onChange={(e) => update("kerberosServiceName", e.target.value)}
            />
          </FormRow>
          <FormRow block labelKey="organization:Kerberos keytab">
            <Input
              value={organization.kerberosKeytab ?? ""}
              onChange={(e) => update("kerberosKeytab", e.target.value)}
            />
          </FormRow>
          <FormRow labelKey="theme:Customize theme" block>
            <ThemeEditor themeData={organization.themeData} onChange={(next) => update("themeData", next)} />
          </FormRow>
          <FormRow labelKey="general:LDAPs" block>
            <div className="space-y-2">
              <div className="rounded-lg border">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b text-xs uppercase text-muted-foreground">
                      <th className="px-3 py-2 text-left">{i18next.t("ldap:Server name")}</th>
                      <th className="px-3 py-2 text-left">{i18next.t("ldap:Server")}</th>
                      <th className="px-3 py-2 text-left">{i18next.t("ldap:Base DN")}</th>
                      <th className="px-3 py-2 text-left">{i18next.t("ldap:Auto Sync")}</th>
                      <th className="px-3 py-2 text-left">{i18next.t("ldap:Last Sync")}</th>
                      <th className="px-3 py-2 text-left">{i18next.t("general:Action")}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {(ldaps ?? []).length === 0 ? (
                      <tr>
                        <td colSpan={6} className="px-3 py-6 text-center text-muted-foreground">
                          {i18next.t("general:No data")}
                        </td>
                      </tr>
                    ) : (
                      (ldaps ?? []).map((ldap: any) => (
                        <tr key={ldap.id} className="border-b last:border-0">
                          <td className="px-3 py-2">
                            <Link
                              to={`/ldap/${organization.name}/${ldap.id}`}
                              className="underline-offset-4 hover:underline"
                            >
                              {ldap.serverName}
                            </Link>
                          </td>
                          <td className="px-3 py-2">
                            {ldap.host}:{ldap.port}
                          </td>
                          <td className="px-3 py-2">{ldap.baseDn}</td>
                          <td className="px-3 py-2">
                            {ldap.autoSync === 0 ? (
                              <span className="text-warning">{i18next.t("general:Disable")}</span>
                            ) : (
                              <span className="text-success">{ldap.autoSync} mins</span>
                            )}
                          </td>
                          <td className="px-3 py-2">{ldap.lastSync}</td>
                          <td className="px-3 py-2">
                            <div className="flex flex-wrap gap-1">
                              <Button variant="outline" size="sm" asChild>
                                <Link to={`/ldap/sync/${organization.name}/${ldap.id}`}>{i18next.t("general:Sync")}</Link>
                              </Button>
                              <Button variant="outline" size="sm" asChild>
                                <Link to={`/ldap/${organization.name}/${ldap.id}`}>{i18next.t("general:Edit")}</Link>
                              </Button>
                              <ConfirmButton
                                variant="destructive"
                                size="sm"
                                description={`${ldap.serverName ?? ""}`}
                                onConfirm={() => deleteLdap(ldap)}
                              >
                                {i18next.t("general:Delete")}
                              </ConfirmButton>
                            </div>
                          </td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>
              {mode !== "add" ? (
                <Button variant="outline" size="sm" onClick={() => navigate(`/ldap/${organization.name}/new`)}>
                  {i18next.t("general:Add")}
                </Button>
              ) : null}
            </div>
          </FormRow>
        </TabsContent>
      </Tabs>
    </EditPageShell>
  );
}
