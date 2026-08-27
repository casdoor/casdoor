import * as React from "react";
import i18next from "i18next";
import {useNavigate, useParams} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Switch} from "@/components/ui/switch";
import {Tabs, TabsContent, TabsList, TabsTrigger} from "@/components/ui/tabs";
import {CodeEditor} from "@/components/common/CodeEditor";
import {Loading} from "@/components/common/Loading";
import {MultiSelect} from "@/components/common/MultiSelect";
import {SearchableSelect} from "@/components/common/SearchableSelect";
import {SelectField} from "@/components/common/SelectField";
import {TagsInput} from "@/components/common/TagsInput";
import {EditPageShell} from "@/components/crud/EditPageShell";
import {EditableTable} from "@/components/crud/EditableTable";
import {FormRow} from "@/components/crud/FormRow";
import {useAccount} from "@/hooks/use-account";
import {useEditRecord} from "@/hooks/use-edit-record";
import {useCertOptions, useOrganizationOptions, useProviderOptions} from "@/hooks/use-options";
import {submitEdit} from "@/lib/crud";
import {SignupTableDefaultCssMap} from "@/lib/signup-css";
import * as ApplicationBackend from "@/backend/ApplicationBackend";
import * as Setting from "@/lib/setting";

const GRANT_TYPES = [
  "authorization_code",
  "password",
  "client_credentials",
  "token",
  "id_token",
  "refresh_token",
  "device_code",
];
const TOKEN_FORMATS = ["JWT", "JWT-Empty", "JWT-Custom", "JWT-Standard"];
const TOKEN_SIGNING_METHODS = ["RS256", "RS512", "ES256", "ES512", "ES384"];
const SIGNIN_RULES = ["All", "Email only", "Phone only", "None"];
const SIGNUP_RULES = ["None", "Normal", "No verification", "Random", "Personal"];
const APPLICATION_TYPES = ["All", "Web", "Native", "SPA"];
const APPLICATION_CATEGORIES = ["Default", "OAuth", "SAML", "CAS"];
const SIGNUP_ITEM_NAMES = [
  "ID", "Username", "Display name", "Affiliation", "ID card", "Country/Region", "Email", "Phone",
  "Email or Phone", "Phone or Email", "Password", "Confirm password", "Invitation code", "Agreement",
  "Signup button", "Providers", "Text 1", "Text 2", "Text 3", "Text 4", "Text 5", "Languages",
];

export default function ApplicationEditPage() {
  const {organizationName = "", applicationName = ""} = useParams();
  const navigate = useNavigate();
  const {account} = useAccount();
  const [saving, setSaving] = React.useState(false);

  const organizations = useOrganizationOptions();
  const certs = useCertOptions(organizationName);
  const providers = useProviderOptions(organizationName);

  const {record: application, updateField, loading, mode, setMode} = useEditRecord<any>({
    fetch: () => ApplicationBackend.getApplication("admin", applicationName),
    deps: [applicationName],
  });

  if (loading || application === null) {
    return <Loading />;
  }

  const save = async(exitAfterSave: boolean) => {
    setSaving(true);
    await submitEdit({
      mode,
      record: Setting.deepCopy(application),
      add: (record) => ApplicationBackend.addApplication(record),
      update: (record) => ApplicationBackend.updateApplication("admin", applicationName, record),
      onSaved: () => {
        setMode("edit");
        if (exitAfterSave) {
          navigate("/applications");
        } else if (application.name !== applicationName) {
          navigate(`/applications/${application.organization}/${application.name}`, {replace: true});
        }
      },
    });
    setSaving(false);
  };

  return (
    <EditPageShell
      title={`${i18next.t("application:Edit Application")} - ${application.displayName || application.name}`}
      mode={mode}
      backTo="/applications"
      onSave={save}
      saving={saving}
      extraActions={
        <Button variant="outline" onClick={() => Setting.goToLink(Setting.getLoginLink(application))}>
          {i18next.t("general:Preview")}
        </Button>
      }
    >
      <Tabs defaultValue="basic">
        <TabsList className="mb-2 flex-wrap">
          <TabsTrigger value="basic">{i18next.t("general:Basic info")}</TabsTrigger>
          <TabsTrigger value="signin">{i18next.t("application:Signin methods")}</TabsTrigger>
          <TabsTrigger value="signup">{i18next.t("application:Signup items")}</TabsTrigger>
          <TabsTrigger value="providers">{i18next.t("application:Providers")}</TabsTrigger>
          <TabsTrigger value="oauth">OAuth</TabsTrigger>
          <TabsTrigger value="saml">SAML</TabsTrigger>
          <TabsTrigger value="appearance">{i18next.t("theme:Customize theme")}</TabsTrigger>
        </TabsList>

        <TabsContent value="basic">
          <FormRow labelKey="general:Organization">
            <SearchableSelect
              value={application.organization}
              disabled={!Setting.isAdminUser(account)}
              onChange={(v) => updateField("organization", v)}
              options={organizations}
            />
          </FormRow>
          <FormRow labelKey="general:Name">
            <Input value={application.name ?? ""} onChange={(e) => updateField("name", e.target.value)} />
          </FormRow>
          <FormRow labelKey="general:Display name">
            <Input value={application.displayName ?? ""} onChange={(e) => updateField("displayName", e.target.value)} />
          </FormRow>
          <FormRow labelKey="general:Description">
            <Input value={application.description ?? ""} onChange={(e) => updateField("description", e.target.value)} />
          </FormRow>
          <FormRow labelKey="general:Category">
            <SelectField
              value={application.category ?? "Default"}
              onChange={(v) => updateField("category", v)}
              options={APPLICATION_CATEGORIES.map((item) => ({id: item, name: item}))}
            />
          </FormRow>
          <FormRow labelKey="general:Type">
            <SelectField
              value={application.type ?? "All"}
              onChange={(v) => updateField("type", v)}
              options={APPLICATION_TYPES.map((item) => ({id: item, name: item}))}
            />
          </FormRow>
          <FormRow labelKey="general:Logo">
            <div className="space-y-2">
              <Input value={application.logo ?? ""} onChange={(e) => updateField("logo", e.target.value)} />
              {application.logo ? (
                <img src={application.logo} alt="logo" className="h-16 rounded bg-white object-contain p-1" />
              ) : null}
            </div>
          </FormRow>
          <FormRow labelKey="general:Favicon">
            <Input value={application.favicon ?? ""} onChange={(e) => updateField("favicon", e.target.value)} />
          </FormRow>
          <FormRow labelKey="application:Home">
            <Input value={application.homepageUrl ?? ""} onChange={(e) => updateField("homepageUrl", e.target.value)} />
          </FormRow>
          <FormRow labelKey="general:Tags">
            <TagsInput value={application.tags ?? []} onChange={(v) => updateField("tags", v)} />
          </FormRow>
          <FormRow labelKey="application:Is shared">
            <Switch checked={!!application.isShared} onCheckedChange={(v) => updateField("isShared", v)} />
          </FormRow>
          <FormRow labelKey="application:Disable signin">
            <Switch checked={!!application.disableSignin} onCheckedChange={(v) => updateField("disableSignin", v)} />
          </FormRow>
          <FormRow labelKey="application:Enable signup">
            <Switch checked={!!application.enableSignUp} onCheckedChange={(v) => updateField("enableSignUp", v)} />
          </FormRow>
          <FormRow labelKey="general:IP whitelist">
            <TagsInput value={application.ipWhitelist ?? []} onChange={(v) => updateField("ipWhitelist", v)} />
          </FormRow>
        </TabsContent>

        <TabsContent value="signin">
          <FormRow labelKey="application:Signin methods" block>
            <EditableTable
              rows={application.signinMethods ?? []}
              onChange={(rows) => updateField("signinMethods", rows)}
              newRow={() => ({name: "Password", displayName: "Password", rule: "All"})}
              columns={[
                {
                  key: "name",
                  title: i18next.t("general:Name"),
                  width: 200,
                  render: (row: any, _i, patch) => (
                    <SelectField
                      value={row.name}
                      onChange={(v) => patch({name: v, displayName: v})}
                      options={["Password", "Verification code", "WebAuthn", "LDAP", "Face ID", "Device login"].map(
                        (item) => ({id: item, name: item}),
                      )}
                    />
                  ),
                },
                {
                  key: "displayName",
                  title: i18next.t("general:Display name"),
                  width: 200,
                  render: (row: any, _i, patch) => (
                    <Input value={row.displayName ?? ""} onChange={(e) => patch({displayName: e.target.value})} />
                  ),
                },
                {
                  key: "rule",
                  title: i18next.t("application:Rule"),
                  width: 180,
                  render: (row: any, _i, patch) => (
                    <SelectField
                      value={row.rule}
                      onChange={(v) => patch({rule: v})}
                      options={[...SIGNIN_RULES, "Hide password"].map((item) => ({id: item, name: item}))}
                    />
                  ),
                },
              ]}
            />
          </FormRow>
          <FormRow labelKey="application:Enable code signin">
            <Switch
              checked={!!application.enableCodeSignin}
              onCheckedChange={(v) => updateField("enableCodeSignin", v)}
            />
          </FormRow>
          <FormRow labelKey="application:Enable signin session">
            <Switch
              checked={!!application.enableSigninSession}
              onCheckedChange={(v) => updateField("enableSigninSession", v)}
            />
          </FormRow>
          <FormRow labelKey="application:Enable auto signin">
            <Switch
              checked={!!application.enableAutoSignin}
              onCheckedChange={(v) => updateField("enableAutoSignin", v)}
            />
          </FormRow>
          <FormRow labelKey="application:Failed signin limit">
            <Input
              type="number"
              value={application.failedSigninLimit ?? 5}
              onChange={(e) => updateField("failedSigninLimit", Setting.myParseInt(e.target.value))}
            />
          </FormRow>
          <FormRow labelKey="application:Failed signin frozen time">
            <Input
              type="number"
              value={application.failedSigninFrozenTime ?? 15}
              onChange={(e) => updateField("failedSigninFrozenTime", Setting.myParseInt(e.target.value))}
            />
          </FormRow>
          <FormRow labelKey="application:Signin URL">
            <Input value={application.signinUrl ?? ""} onChange={(e) => updateField("signinUrl", e.target.value)} />
          </FormRow>
          <FormRow labelKey="application:Signup URL">
            <Input value={application.signupUrl ?? ""} onChange={(e) => updateField("signupUrl", e.target.value)} />
          </FormRow>
          <FormRow labelKey="application:Forget URL">
            <Input value={application.forgetUrl ?? ""} onChange={(e) => updateField("forgetUrl", e.target.value)} />
          </FormRow>
        </TabsContent>

        <TabsContent value="signup">
          <FormRow labelKey="application:Signup items" block>
            <EditableTable
              rows={application.signupItems ?? []}
              onChange={(rows) => updateField("signupItems", rows)}
              newRow={() => ({name: "Username", visible: true, required: true, rule: "None"})}
              columns={[
                {
                  key: "name",
                  title: i18next.t("general:Name"),
                  width: 190,
                  render: (row: any, _i, patch) => (
                    <SelectField
                      value={row.name}
                      onChange={(v) => patch({name: v, customCss: SignupTableDefaultCssMap[v] ?? ""})}
                      options={SIGNUP_ITEM_NAMES.map((item) => ({id: item, name: item}))}
                    />
                  ),
                },
                {
                  key: "visible",
                  title: i18next.t("organization:Visible"),
                  width: 90,
                  render: (row: any, _i, patch) => (
                    <Switch checked={!!row.visible} onCheckedChange={(v) => patch({visible: v})} />
                  ),
                },
                {
                  key: "required",
                  title: i18next.t("organization:Required"),
                  width: 90,
                  render: (row: any, _i, patch) => (
                    <Switch
                      checked={!!row.required}
                      disabled={!row.visible}
                      onCheckedChange={(v) => patch({required: v})}
                    />
                  ),
                },
                {
                  key: "prompted",
                  title: i18next.t("provider:Prompted"),
                  width: 90,
                  render: (row: any, _i, patch) => (
                    <Switch
                      checked={!!row.prompted}
                      disabled={row.visible}
                      onCheckedChange={(v) => patch({prompted: v})}
                    />
                  ),
                },
                {
                  key: "rule",
                  title: i18next.t("application:Rule"),
                  width: 160,
                  render: (row: any, _i, patch) => (
                    <SelectField
                      value={row.rule}
                      onChange={(v) => patch({rule: v})}
                      options={SIGNUP_RULES.map((item) => ({id: item, name: item}))}
                    />
                  ),
                },
                {
                  key: "label",
                  title: i18next.t("signup:Label"),
                  width: 180,
                  render: (row: any, _i, patch) => (
                    <Input value={row.label ?? ""} onChange={(e) => patch({label: e.target.value})} />
                  ),
                },
                {
                  key: "placeholder",
                  title: i18next.t("signup:Placeholder"),
                  width: 180,
                  render: (row: any, _i, patch) => (
                    <Input value={row.placeholder ?? ""} onChange={(e) => patch({placeholder: e.target.value})} />
                  ),
                },
                {
                  key: "regex",
                  title: i18next.t("signup:Regex"),
                  width: 180,
                  render: (row: any, _i, patch) => (
                    <Input value={row.regex ?? ""} onChange={(e) => patch({regex: e.target.value})} />
                  ),
                },
              ]}
            />
          </FormRow>
          <FormRow labelKey="application:Terms of Use">
            <Input value={application.termsOfUse ?? ""} onChange={(e) => updateField("termsOfUse", e.target.value)} />
          </FormRow>
          <FormRow labelKey="application:Default group">
            <Input value={application.defaultGroup ?? ""} onChange={(e) => updateField("defaultGroup", e.target.value)} />
          </FormRow>
          <FormRow labelKey="application:Default tag">
            <Input value={application.defaultTag ?? ""} onChange={(e) => updateField("defaultTag", e.target.value)} />
          </FormRow>
          <FormRow labelKey="application:Enable link with email">
            <Switch
              checked={!!application.enableLinkWithEmail}
              onCheckedChange={(v) => updateField("enableLinkWithEmail", v)}
            />
          </FormRow>
        </TabsContent>

        <TabsContent value="providers">
          <FormRow labelKey="application:Providers" block>
            <EditableTable
              rows={application.providers ?? []}
              onChange={(rows) => updateField("providers", rows)}
              newRow={() => ({
                name: providers[0]?.value ?? "",
                canSignUp: true,
                canSignIn: true,
                canUnlink: true,
                prompted: false,
                signupGroup: "",
                rule: "None",
              })}
              columns={[
                {
                  key: "name",
                  title: i18next.t("general:Name"),
                  width: 240,
                  render: (row: any, _i, patch) => (
                    <SelectField
                      value={row.name}
                      onChange={(v) => patch({name: v})}
                      options={providers.map((option) => ({id: option.value, name: option.label as string}))}
                    />
                  ),
                },
                {
                  key: "canSignUp",
                  title: i18next.t("provider:Can signup"),
                  width: 110,
                  render: (row: any, _i, patch) => (
                    <Switch checked={!!row.canSignUp} onCheckedChange={(v) => patch({canSignUp: v})} />
                  ),
                },
                {
                  key: "canSignIn",
                  title: i18next.t("provider:Can signin"),
                  width: 110,
                  render: (row: any, _i, patch) => (
                    <Switch checked={!!row.canSignIn} onCheckedChange={(v) => patch({canSignIn: v})} />
                  ),
                },
                {
                  key: "canUnlink",
                  title: i18next.t("provider:Can unlink"),
                  width: 110,
                  render: (row: any, _i, patch) => (
                    <Switch checked={!!row.canUnlink} onCheckedChange={(v) => patch({canUnlink: v})} />
                  ),
                },
                {
                  key: "prompted",
                  title: i18next.t("provider:Prompted"),
                  width: 110,
                  render: (row: any, _i, patch) => (
                    <Switch checked={!!row.prompted} onCheckedChange={(v) => patch({prompted: v})} />
                  ),
                },
              ]}
            />
          </FormRow>
        </TabsContent>

        <TabsContent value="oauth">
          <FormRow labelKey="application:Client ID">
            <Input value={application.clientId ?? ""} disabled />
          </FormRow>
          <FormRow labelKey="application:Client secret">
            <Input value={application.clientSecret ?? ""} disabled />
          </FormRow>
          <FormRow labelKey="application:Redirect URLs">
            <TagsInput value={application.redirectUris ?? []} onChange={(v) => updateField("redirectUris", v)} />
          </FormRow>
          <FormRow labelKey="application:Grant types">
            <MultiSelect
              value={application.grantTypes ?? []}
              onChange={(v) => updateField("grantTypes", v)}
              options={GRANT_TYPES.map((item) => ({value: item, label: item}))}
            />
          </FormRow>
          <FormRow labelKey="general:Cert">
            <SearchableSelect
              value={application.cert ?? ""}
              onChange={(v) => updateField("cert", v)}
              options={certs}
            />
          </FormRow>
          <FormRow labelKey="application:Token format">
            <SelectField
              value={application.tokenFormat ?? "JWT"}
              onChange={(v) => updateField("tokenFormat", v)}
              options={TOKEN_FORMATS.map((item) => ({id: item, name: item}))}
            />
          </FormRow>
          <FormRow labelKey="application:Token signing method">
            <SelectField
              value={application.tokenSigningMethod ?? "RS256"}
              onChange={(v) => updateField("tokenSigningMethod", v)}
              options={TOKEN_SIGNING_METHODS.map((item) => ({id: item, name: item}))}
            />
          </FormRow>
          <FormRow labelKey="application:Token fields">
            <MultiSelect
              value={application.tokenFields ?? []}
              onChange={(v) => updateField("tokenFields", v)}
              creatable
              options={Setting.UserFields.map((item: string) => ({value: item, label: item}))}
            />
          </FormRow>
          <FormRow labelKey="application:Token expire">
            <Input
              type="number"
              value={application.expireInHours ?? 168}
              onChange={(e) => updateField("expireInHours", Setting.myParseInt(e.target.value))}
            />
          </FormRow>
          <FormRow labelKey="application:Refresh token expire">
            <Input
              type="number"
              value={application.refreshExpireInHours ?? 168}
              onChange={(e) => updateField("refreshExpireInHours", Setting.myParseInt(e.target.value))}
            />
          </FormRow>
          <FormRow labelKey="application:Cookie expire">
            <Input
              type="number"
              value={application.cookieExpireInHours ?? 720}
              onChange={(e) => updateField("cookieExpireInHours", Setting.myParseInt(e.target.value))}
            />
          </FormRow>
          <FormRow labelKey="application:Scopes">
            <TagsInput value={application.scopes ?? []} onChange={(v) => updateField("scopes", v)} />
          </FormRow>
        </TabsContent>

        <TabsContent value="saml">
          <FormRow labelKey="application:SAML reply URL">
            <Input value={application.samlReplyUrl ?? ""} onChange={(e) => updateField("samlReplyUrl", e.target.value)} />
          </FormRow>
          <FormRow labelKey="application:Enable SAML compress">
            <Switch
              checked={!!application.enableSamlCompress}
              onCheckedChange={(v) => updateField("enableSamlCompress", v)}
            />
          </FormRow>
          <FormRow labelKey="application:Enable SAML POST binding">
            <Switch
              checked={!!application.enableSamlPostBinding}
              onCheckedChange={(v) => updateField("enableSamlPostBinding", v)}
            />
          </FormRow>
          <FormRow labelKey="application:Enable SAML assertion signature">
            <Switch
              checked={!!application.enableSamlAssertionSignature}
              onCheckedChange={(v) => updateField("enableSamlAssertionSignature", v)}
            />
          </FormRow>
          <FormRow labelKey="application:Use Email as NameID">
            <Switch
              checked={!!application.useEmailAsSamlNameId}
              onCheckedChange={(v) => updateField("useEmailAsSamlNameId", v)}
            />
          </FormRow>
          <FormRow labelKey="application:Disable SAML attributes">
            <Switch
              checked={!!application.disableSamlAttributes}
              onCheckedChange={(v) => updateField("disableSamlAttributes", v)}
            />
          </FormRow>
          <FormRow labelKey="application:SAML attributes" block>
            <EditableTable
              rows={application.samlAttributes ?? []}
              onChange={(rows) => updateField("samlAttributes", rows)}
              newRow={() => ({name: "", nameFormat: "", value: ""})}
              reorderable={false}
              columns={[
                {
                  key: "name",
                  title: i18next.t("general:Name"),
                  width: 220,
                  render: (row: any, _i, patch) => (
                    <Input value={row.name ?? ""} onChange={(e) => patch({name: e.target.value})} />
                  ),
                },
                {
                  key: "nameFormat",
                  title: i18next.t("general:Name format"),
                  width: 220,
                  render: (row: any, _i, patch) => (
                    <Input value={row.nameFormat ?? ""} onChange={(e) => patch({nameFormat: e.target.value})} />
                  ),
                },
                {
                  key: "value",
                  title: i18next.t("webhook:Value"),
                  render: (row: any, _i, patch) => (
                    <Input value={row.value ?? ""} onChange={(e) => patch({value: e.target.value})} />
                  ),
                },
              ]}
            />
          </FormRow>
        </TabsContent>

        <TabsContent value="appearance">
          <FormRow labelKey="application:Form CSS" block>
            <CodeEditor
              language="css"
              value={application.formCss ?? ""}
              onChange={(v) => updateField("formCss", v)}
            />
          </FormRow>
          <FormRow labelKey="application:Form CSS Mobile" block>
            <CodeEditor
              language="css"
              value={application.formCssMobile ?? ""}
              onChange={(v) => updateField("formCssMobile", v)}
            />
          </FormRow>
          <FormRow labelKey="application:Background URL">
            <Input
              value={application.formBackgroundUrl ?? ""}
              onChange={(e) => updateField("formBackgroundUrl", e.target.value)}
            />
          </FormRow>
          <FormRow labelKey="application:Header HTML" block>
            <CodeEditor
              language="html"
              value={application.headerHtml ?? ""}
              onChange={(v) => updateField("headerHtml", v)}
            />
          </FormRow>
          <FormRow labelKey="application:Footer HTML" block>
            <CodeEditor
              language="html"
              value={application.footerHtml ?? ""}
              onChange={(v) => updateField("footerHtml", v)}
            />
          </FormRow>
          <FormRow labelKey="application:Side panel HTML" block>
            <CodeEditor
              language="html"
              value={application.formSideHtml ?? ""}
              onChange={(v) => updateField("formSideHtml", v)}
            />
          </FormRow>
        </TabsContent>
      </Tabs>
    </EditPageShell>
  );
}
