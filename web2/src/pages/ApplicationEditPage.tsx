import * as React from "react";
import i18next from "i18next";
import copy from "copy-to-clipboard";
import {Copy, Link as LinkIcon, Upload} from "lucide-react";
import {useNavigate, useParams} from "react-router-dom";
import {UnauthorizedPage} from "@/components/common/UnauthorizedPage";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Switch} from "@/components/ui/switch";
import {RadioGroup, RadioGroupItem} from "@/components/ui/radio-group";
import {Tabs, TabsContent, TabsList, TabsTrigger} from "@/components/ui/tabs";
import {CodeEditor} from "@/components/common/CodeEditor";
import {Loading} from "@/components/common/Loading";
import {MultiSelect} from "@/components/common/MultiSelect";
import {SearchableSelect} from "@/components/common/SearchableSelect";
import {SelectField} from "@/components/common/SelectField";
import {TagsInput} from "@/components/common/TagsInput";
import {ThemeEditor} from "@/components/common/ThemeEditor";
import {ApplicationExportButton, ApplicationImportModal} from "@/components/application/ApplicationImportExport";
import {EditPageShell} from "@/components/crud/EditPageShell";
import {EditableTable} from "@/components/crud/EditableTable";
import {FormRow} from "@/components/crud/FormRow";
import {useAccount} from "@/hooks/use-account";
import {useEditRecord} from "@/hooks/use-edit-record";
import {useCertOptions, useOrganizationOptions, useProviderList, useProviderOptions} from "@/hooks/use-options";
import {getModeTitleKey, submitEdit} from "@/lib/crud";
import {
  enumOptions,
  enumSelectOptions,
  type EnumMap,
  PROVIDER_BINDING_RULES,
  PROVIDER_CAPTCHA_RULES,
  PROVIDER_CODE_RULES,
  PROVIDER_GOOGLE_RULES,
} from "@/lib/enum-labels";
import {SigninTableDefaultCssMap} from "@/lib/signin-css";
import {SignupTableDefaultCssMap} from "@/lib/signup-css";
import * as ApplicationBackend from "@/backend/ApplicationBackend";
import * as ResourceBackend from "@/backend/ResourceBackend";
import * as Setting from "@/lib/setting";

/** the provider kinds a user account can be linked to */
const LINKABLE_PROVIDER_CATEGORIES = ["OAuth", "Web3", "SAML"];

/** The rule options a provider row offers, or null when it has no rule at all. */
function getProviderRuleMap(provider: any) {
  if (provider?.type === "Google") {
    return PROVIDER_GOOGLE_RULES;
  }
  if (provider?.category === "Captcha") {
    return PROVIDER_CAPTCHA_RULES;
  }
  if (provider?.category === "SMS" || provider?.category === "Email") {
    return PROVIDER_CODE_RULES;
  }
  return null;
}

/** antd rewrites the stored "None" to each kind's own default before showing it. */
function normalizeProviderRule(provider: any, rule: string) {
  if (rule !== "None") {
    return rule;
  }
  if (provider?.type === "Google") {
    return "Default";
  }
  if (provider?.category === "SMS" || provider?.category === "Email") {
    return "all";
  }
  return rule;
}

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
/**
 * A signin method's "Rule" also depends on the method, as in the antd
 * SigninMethodTable. Methods not listed have no rule.
 */
const SIGNIN_METHOD_RULES: Record<string, EnumMap> = {
  "Verification code": {
    "All": {i18nKey: "general:All"},
    "Email only": {i18nKey: "general:Email only"},
    "Phone only": {i18nKey: "general:Phone only"},
  },
  "Password": {
    "All": {i18nKey: "general:All"},
    "Non-LDAP": {i18nKey: "general:Non-LDAP"},
    "Hide password": {i18nKey: "general:Hide password"},
  },
  "WeChat": {
    "Tab": {i18nKey: "general:Tab"},
    "Login page": {i18nKey: "general:Login page"},
  },
  "Device login": {
    "Tab": {i18nKey: "general:Tab"},
    "Login page": {i18nKey: "general:Login page"},
  },
};
const TOKEN_ATTRIBUTE_CATEGORIES: EnumMap = {
  "Static Value": {i18nKey: "application:Static Value"},
  "Existing Field": {i18nKey: "application:Existing Field"},
};

/** the user fields an "Existing Field" token attribute may copy */
const TOKEN_ATTRIBUTE_USER_FIELDS = [
  "Owner", "Name", "Id", "DisplayName", "Avatar", "Email", "Phone",
  "Tag", "Roles", "Permissions", "permissionNames", "Groups",
];

const SIGNUP_ITEM_TYPES: EnumMap = {
  "Input": {i18nKey: "application:Input"},
  "Single Choice": {i18nKey: "application:Single Choice"},
  "Multiple Choices": {i18nKey: "application:Multiple Choices"},
};

/**
 * A signup item's "Rule" means something different for each item, so the antd
 * table picks the options from the item name. Items not listed have no rule.
 */
const SIGNUP_ITEM_RULES: Record<string, EnumMap> = {
  "ID": {
    "Random": {i18nKey: "application:Random"},
    "Incremental": {i18nKey: "application:Incremental"},
  },
  "Display name": {
    "None": {i18nKey: "general:None"},
    "Real name": {i18nKey: "application:Real name"},
    "First, last": {i18nKey: "application:First, last"},
  },
  "Email": {
    "Normal": {i18nKey: "application:Normal"},
    "No verification": {i18nKey: "application:No verification"},
  },
  "Phone": {
    "Normal": {i18nKey: "application:Normal"},
    "No verification": {i18nKey: "application:No verification"},
  },
  "Agreement": {
    "None": {i18nKey: "application:Only signup"},
    "Signin": {i18nKey: "application:Signin"},
    "Signin (Default True)": {i18nKey: "application:Signin (Default True)"},
  },
  "Providers": {
    "big": {i18nKey: "application:Big icon"},
    "small": {i18nKey: "application:Small icon"},
  },
  "Languages": {
    "None": {i18nKey: "general:Default"},
    "Label": {i18nKey: "signup:Label"},
  },
};
const APPLICATION_TYPES = ["All", "Web", "Native", "SPA"];
const APPLICATION_CATEGORIES = ["Default", "OAuth", "SAML", "CAS"];
const SSL_MODES = ["", "HTTP", "HTTPS and HTTP", "HTTPS Only"];
const SAML_HASH_ALGORITHMS = ["SHA1", "SHA256", "SHA512"];
const TOKEN_ATTRIBUTE_TYPES = ["String", "Number", "Boolean"];

/** The blocks the sign-in page is built from, see web/src/table/SigninTable.js. */
const SIGNIN_ITEM_NAMES: {name: string; labelKey: string}[] = [
  {name: "Signin methods", labelKey: "application:Signin methods"},
  {name: "Logo", labelKey: "general:Logo"},
  {name: "Back button", labelKey: "login:Back button"},
  {name: "Languages", labelKey: "general:Languages"},
  {name: "Username", labelKey: "signup:Username"},
  {name: "Password", labelKey: "general:Password"},
  {name: "Verification code", labelKey: "login:Verification code"},
  {name: "Providers", labelKey: "application:Providers"},
  {name: "Agreement", labelKey: "signup:Agreement"},
  {name: "Forgot password?", labelKey: "login:Forgot password?"},
  {name: "Login button", labelKey: "login:Signin button"},
  {name: "Signup link", labelKey: "general:Signup link"},
  {name: "Captcha", labelKey: "general:Captcha"},
  {name: "Auto sign in", labelKey: "login:Auto sign in"},
  {name: "Select organization", labelKey: "login:Select organization"},
];

/** Only a few signin items take a rule, and each has its own option set. */
function getSigninItemRuleOptions(name: string) {
  switch (name) {
  case "Providers":
    return [
      {id: "big", name: i18next.t("application:Big icon")},
      {id: "small", name: i18next.t("application:Small icon")},
    ];
  case "Captcha":
    return [
      {id: "pop up", name: i18next.t("application:Pop up")},
      {id: "inline", name: i18next.t("application:Inline")},
    ];
  case "Forgot password?":
    return [
      {id: "None", name: `${i18next.t("login:Auto sign in")} - ${i18next.t("general:True")}`},
      {id: "Auto sign in - False", name: `${i18next.t("login:Auto sign in")} - ${i18next.t("general:False")}`},
    ];
  case "Languages":
    return [
      {id: "None", name: i18next.t("general:Default")},
      {id: "Label", name: i18next.t("signup:Label")},
    ];
  default:
    return [];
  }
}

const SIGNUP_ITEM_NAMES = [
  "ID", "Username", "Display name", "Affiliation", "ID card", "Country/Region", "Email", "Phone",
  "Email or Phone", "Phone or Email", "Password", "Confirm password", "Invitation code", "Agreement",
  "Signup button", "Providers", "Text 1", "Text 2", "Text 3", "Text 4", "Text 5", "Languages",
];

/** how the edit page lays its tabs out: across the top, or down the left side */
type MenuMode = "horizontal" | "vertical";

/** the tabs, in the antd page's order; also what the URL hash may name */
const TAB_KEYS = [
  "basic",
  "authentication",
  "oidc-oauth",
  "saml",
  "providers",
  "ui-customization",
  "security",
  "reverse-proxy",
];

export default function ApplicationEditPage() {
  const {organizationName = "", applicationName = ""} = useParams();
  const navigate = useNavigate();
  const {account} = useAccount();
  const [saving, setSaving] = React.useState(false);

  const organizations = useOrganizationOptions();
  const certs = useCertOptions(organizationName);
  const providers = useProviderOptions(organizationName);
  const providerObjs = useProviderList(organizationName);

  const {record: application, updateField, setRecord, loading, denied, mode, setMode} = useEditRecord<any>({
    fetch: () => ApplicationBackend.getApplication("admin", applicationName),
    deps: [applicationName],
  });

  // antd keeps the open tab in the URL hash and lets the user lay the tabs out
  // across the top or down the side; both are page state, not application fields
  const [activeTab, setActiveTab] = React.useState(() => {
    const hash = window.location.hash.replace("#", "");
    return TAB_KEYS.includes(hash) ? hash : "basic";
  });
  const [menuMode, setMenuMode] = React.useState<MenuMode>("horizontal");
  const selectTab = (key: string) => {
    setActiveTab(key);
    window.location.hash = key;
  };
  const [samlMetadata, setSamlMetadata] = React.useState("");
  const [uploadingTerms, setUploadingTerms] = React.useState(false);
  const termsFileRef = React.useRef<HTMLInputElement>(null);
  const enableSamlPostBinding = !!application?.enableSamlPostBinding;

  React.useEffect(() => {
    // the metadata is generated from the saved application, so it needs it to exist
    if (mode === "add" || !applicationName) {
      return;
    }
    ApplicationBackend.getSamlMetadata("admin", applicationName, enableSamlPostBinding).then((data: any) => {
      setSamlMetadata(data?.toString() ?? "");
    });
  }, [mode, applicationName, enableSamlPostBinding]);

  if (denied) {
    return <UnauthorizedPage />;
  }

  if (loading || application === null) {
    return <Loading />;
  }

  /** the application name ends up in URLs, so these characters are rejected outright */
  const updateName = (value: string) => {
    if (/[/?:@#&%=+;]/.test(value)) {
      Setting.showMessage(
        "error",
        `${i18next.t("application:Invalid characters in application name")}: / ? : @ # & % = + ;`,
      );
      return;
    }
    updateField("name", value);
  };

  const uploadTermsOfUse = (file: File | undefined) => {
    if (termsFileRef.current) {
      termsFileRef.current.value = "";
    }
    if (!file) {
      return;
    }
    if (file.type !== "text/html") {
      Setting.showMessage("error", i18next.t("application:Please select a HTML file"));
      return;
    }
    setUploadingTerms(true);
    const fullFilePath = `termsOfUse/${application.owner}/${application.name}.html`;
    ResourceBackend.uploadResource(account?.owner, account?.name, "termsOfUse", "ApplicationEditPage", fullFilePath, file)
      .then((res: any) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("application:File uploaded successfully"));
          updateField("termsOfUse", res.data);
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
        }
      })
      .finally(() => setUploadingTerms(false));
  };

  /**
   * The provider a row points at. The backend embeds it on the row, but a row
   * the user just added only has a name, so fall back to the fetched list.
   */
  const resolveProvider = (row: any) =>
    row?.provider ?? providerObjs.find((item: any) => item.name === row?.name);

  const copyToClipboard = (text: string) => {
    copy(text);
    Setting.showMessage("success", i18next.t("general:Copied to clipboard successfully"));
  };

  const idpInitiatedSsoUrl =
    `${window.location.origin}/login/saml/authorize/${application.owner}/${encodeURIComponent(applicationName)}`;
  const samlMetadataUrl =
    `${window.location.origin}/api/saml/metadata?application=admin/${encodeURIComponent(applicationName)}` +
    `&enablePostBinding=${enableSamlPostBinding}`;

  // the sign-in link antd shows next to its login preview
  const redirectUri = application.redirectUris?.length > 0
    ? application.redirectUris[0]
    : "\"ERROR: You must specify at least one Redirect URL in 'Redirect URLs'\"";
  const clientId = application.isShared ? `${application.clientId}-org-${account?.owner}` : application.clientId;
  const signInUrl =
    `/login/oauth/authorize?client_id=${clientId}&response_type=code&redirect_uri=${redirectUri}&scope=read&state=casdoor`;
  const signUpUrl = Setting.isPasswordEnabled(application)
    ? `/signup/${application.name}`
    : signInUrl.replace("/login/oauth/authorize", "/signup/oauth/authorize");
  const promptUrl = `/prompt/${application.name}`;

  const save = async(exitAfterSave: boolean) => {
    // antd refuses to save a custom scope without a name, and drops the empty rows
    const customScopes = (application.customScopes ?? []).filter(
      (scope: any) => scope && Object.values(scope).some((value) => value !== "" && value !== undefined),
    );
    if (customScopes.some((scope: any) => !scope.name)) {
      Setting.showMessage("error", `${i18next.t("general:Name")}: ${i18next.t("provider:This field is required")}`);
      return;
    }

    setSaving(true);
    await submitEdit({
      mode,
      record: {...Setting.deepCopy(application), customScopes},
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
      title={`${i18next.t(getModeTitleKey("application:Edit Application", mode))} - ${application.displayName || application.name}`}
      mode={mode}
      backTo="/applications"
      onSave={save}
      saving={saving}
      extraActions={
        <>
          <Button variant="outline" onClick={() => Setting.goToLink(Setting.getLoginLink(application))}>
            {i18next.t("general:Preview")}
          </Button>
          {mode !== "add" ? (
            <>
              <ApplicationExportButton application={application} />
              <ApplicationImportModal
                application={application}
                onImport={(updates) => setRecord((prev: any) => (prev === null ? prev : {...prev, ...updates}))}
              />
            </>
          ) : null}
        </>
      }
    >
      <Tabs
        value={activeTab}
        onValueChange={selectTab}
        orientation={menuMode === "vertical" ? "vertical" : "horizontal"}
        className={menuMode === "vertical" ? "flex items-start gap-4" : undefined}
      >
        <TabsList className={menuMode === "vertical" ? "sticky top-0 h-auto flex-col items-stretch" : "mb-2 flex-wrap"}>
          <TabsTrigger value="basic">{i18next.t("application:Basic")}</TabsTrigger>
          <TabsTrigger value="authentication">{i18next.t("application:Authentication")}</TabsTrigger>
          <TabsTrigger value="oidc-oauth">OIDC/OAuth</TabsTrigger>
          <TabsTrigger value="saml">SAML</TabsTrigger>
          <TabsTrigger value="providers">{i18next.t("application:Providers")}</TabsTrigger>
          <TabsTrigger value="ui-customization">{i18next.t("application:UI Customization")}</TabsTrigger>
          <TabsTrigger value="security">{i18next.t("application:Security")}</TabsTrigger>
          <TabsTrigger value="reverse-proxy">{i18next.t("application:Reverse Proxy")}</TabsTrigger>
        </TabsList>

        <TabsContent value="basic" className={menuMode === "vertical" ? "mt-0 min-w-0 flex-1" : undefined}>
          {/* antd puts these next to its live sign-in previews; this frontend has no
              previews, so the links live on their own row at the top of the tab */}
          {mode === "add" ? null : (
            <FormRow labelKey="general:Login page" block>
              <div className="flex flex-wrap gap-2">
                <Button variant="outline" size="sm" onClick={() => copyToClipboard(`${window.location.origin}${signInUrl}`)}>
                  <Copy />
                  {i18next.t("application:Copy signin page URL")}
                </Button>
                <Button variant="outline" size="sm" onClick={() => copyToClipboard(`${window.location.origin}${signUpUrl}`)}>
                  <Copy />
                  {i18next.t("application:Copy signup page URL")}
                </Button>
                {Setting.hasPromptPage(application) ? (
                  <Button variant="outline" size="sm" onClick={() => copyToClipboard(`${window.location.origin}${promptUrl}`)}>
                    <Copy />
                    {i18next.t("application:Copy prompt page URL")}
                  </Button>
                ) : null}
              </div>
            </FormRow>
          )}
          <FormRow labelKey="general:Organization">
            <SearchableSelect
              value={application.organization}
              disabled={!Setting.isAdminUser(account)}
              onChange={(v) => updateField("organization", v)}
              options={organizations}
            />
          </FormRow>
          <FormRow labelKey="general:Name">
            <Input value={application.name ?? ""} onChange={(e) => updateName(e.target.value)} />
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
          <FormRow labelKey="general:Is shared">
            <Switch checked={!!application.isShared} onCheckedChange={(v) => updateField("isShared", v)} />
          </FormRow>
          <FormRow labelKey="general:Logo">
            <div className="space-y-2">
              <Input value={application.logo ?? ""} onChange={(e) => updateField("logo", e.target.value)} />
              {application.logo ? (
                <a href={application.logo} target="_blank" rel="noreferrer">
                  <img src={application.logo} alt="logo" className="h-16 rounded bg-white object-contain p-1" />
                </a>
              ) : null}
            </div>
          </FormRow>
          <FormRow labelKey="general:Title">
            <Input value={application.title ?? ""} onChange={(e) => updateField("title", e.target.value)} />
          </FormRow>
          <FormRow labelKey="general:Favicon">
            <div className="space-y-2">
              <Input value={application.favicon ?? ""} onChange={(e) => updateField("favicon", e.target.value)} />
              {application.favicon ? (
                <a href={application.favicon} target="_blank" rel="noreferrer">
                  <img src={application.favicon} alt="favicon" className="h-10 w-10 object-contain" />
                </a>
              ) : null}
            </div>
          </FormRow>
          <FormRow labelKey="general:Home">
            <Input value={application.homepageUrl ?? ""} onChange={(e) => updateField("homepageUrl", e.target.value)} />
          </FormRow>
          <FormRow labelKey="organization:Tags">
            <TagsInput value={application.tags ?? []} onChange={(v) => updateField("tags", v)} />
          </FormRow>
          <FormRow labelKey="application:Default tag">
            <Input value={application.defaultTag ?? ""} onChange={(e) => updateField("defaultTag", e.target.value)} />
          </FormRow>
          <FormRow labelKey="application:Order">
            <Input
              type="number"
              min={0}
              step={1}
              value={application.order ?? 0}
              onChange={(e) => updateField("order", Setting.myParseInt(e.target.value))}
            />
          </FormRow>
          <FormRow labelKey="application:Menu mode">
            <RadioGroup
              className="flex gap-4"
              value={menuMode}
              onValueChange={(value) => setMenuMode(value as MenuMode)}
            >
              {(["horizontal", "vertical"] as MenuMode[]).map((value) => (
                <div key={value} className="flex items-center gap-2">
                  <RadioGroupItem value={value} id={`menu-mode-${value}`} />
                  <label htmlFor={`menu-mode-${value}`} className="text-sm">
                    {i18next.t(value === "horizontal" ? "application:Horizontal" : "application:Vertical")}
                  </label>
                </div>
              ))}
            </RadioGroup>
          </FormRow>
        </TabsContent>

        <TabsContent value="authentication" className={menuMode === "vertical" ? "mt-0 min-w-0 flex-1" : undefined}>
          <FormRow labelKey="application:Cookie expire">
            <Input
              type="number"
              value={application.cookieExpireInHours ?? 720}
              onChange={(e) => updateField("cookieExpireInHours", Setting.myParseInt(e.target.value))}
            />
          </FormRow>
          <FormRow labelKey="ldap:Default group">
            <Input value={application.defaultGroup ?? ""} onChange={(e) => updateField("defaultGroup", e.target.value)} />
          </FormRow>
          <FormRow labelKey="application:Enable signup">
            <Switch checked={!!application.enableSignUp} onCheckedChange={(v) => updateField("enableSignUp", v)} />
          </FormRow>
          <FormRow labelKey="application:Disable signin">
            <Switch checked={!!application.disableSignin} onCheckedChange={(v) => updateField("disableSignin", v)} />
          </FormRow>
          <FormRow labelKey="application:Enable guest signin">
            <Switch
              checked={!!application.enableGuestSignin}
              onCheckedChange={(v) => updateField("enableGuestSignin", v)}
            />
          </FormRow>
          <FormRow labelKey="application:Enable exclusive signin">
            <Switch
              checked={!!application.enableExclusiveSignin}
              onCheckedChange={(v) => updateField("enableExclusiveSignin", v)}
            />
          </FormRow>
          <FormRow labelKey="application:Signin session">
            <Switch
              checked={!!application.enableSigninSession}
              onCheckedChange={(v) => updateField("enableSigninSession", v)}
            />
          </FormRow>
          <FormRow labelKey="application:Auto signin">
            <Switch
              checked={!!application.enableAutoSignin}
              onCheckedChange={(v) => {
                // auto signin reuses the Casdoor session, so it needs one to exist
                if (v && !application.enableSigninSession) {
                  Setting.showMessage(
                    "error",
                    i18next.t("application:Please enable \"Signin session\" first before enabling \"Auto signin\""),
                  );
                  return;
                }
                updateField("enableAutoSignin", v);
              }}
            />
          </FormRow>
          <FormRow labelKey="application:Enable Email linking">
            <Switch
              checked={!!application.enableLinkWithEmail}
              onCheckedChange={(v) => updateField("enableLinkWithEmail", v)}
            />
          </FormRow>
          <FormRow labelKey="general:Signup URL">
            <Input value={application.signupUrl ?? ""} onChange={(e) => updateField("signupUrl", e.target.value)} />
          </FormRow>
          <FormRow labelKey="general:Signin URL">
            <Input value={application.signinUrl ?? ""} onChange={(e) => updateField("signinUrl", e.target.value)} />
          </FormRow>
          <FormRow labelKey="general:Forget URL">
            <Input value={application.forgetUrl ?? ""} onChange={(e) => updateField("forgetUrl", e.target.value)} />
          </FormRow>
          <FormRow labelKey="general:Affiliation URL">
            <Input
              value={application.affiliationUrl ?? ""}
              onChange={(e) => updateField("affiliationUrl", e.target.value)}
            />
          </FormRow>
        </TabsContent>

        <TabsContent value="oidc-oauth" className={menuMode === "vertical" ? "mt-0 min-w-0 flex-1" : undefined}>
          {/* both are editable so that an admin can rotate the pair, as in the antd page */}
          <FormRow labelKey="provider:Client ID">
            <Input value={application.clientId ?? ""} onChange={(e) => updateField("clientId", e.target.value)} />
          </FormRow>
          <FormRow labelKey="provider:Client secret">
            <Input
              value={application.clientSecret ?? ""}
              onChange={(e) => updateField("clientSecret", e.target.value)}
            />
          </FormRow>
          <FormRow labelKey="application:Redirect URLs">
            <TagsInput value={application.redirectUris ?? []} onChange={(v) => updateField("redirectUris", v)} />
          </FormRow>
          <FormRow labelKey="application:Forced redirect origin">
            <Input
              value={application.forcedRedirectOrigin ?? ""}
              onChange={(e) => updateField("forcedRedirectOrigin", e.target.value)}
            />
          </FormRow>
          <FormRow labelKey="application:Backchannel logout URL">
            <Input
              value={application.backchannelLogoutUri ?? ""}
              onChange={(e) => updateField("backchannelLogoutUri", e.target.value)}
            />
          </FormRow>
          <FormRow labelKey="application:Grant types">
            <MultiSelect
              value={application.grantTypes ?? []}
              onChange={(v) => updateField("grantTypes", v)}
              options={GRANT_TYPES.map((item) => ({value: item, label: item}))}
            />
          </FormRow>
          {/* scopes are ScopeItem objects (name / displayName / description), the
              same three columns antd's ScopeTable edits */}
          <FormRow labelKey="general:Scopes" block>
            <EditableTable
              rows={application.scopes ?? []}
              onChange={(rows) => updateField("scopes", rows)}
              newRow={() => ({name: "", displayName: "", description: ""})}
              columns={[
                {
                  key: "name",
                  title: i18next.t("general:Name"),
                  width: "25%",
                  render: (row: any, index, update) => (
                    <Input
                      value={row.name ?? ""}
                      placeholder="e.g., files:read"
                      onChange={(e) => update({name: e.target.value})}
                    />
                  ),
                },
                {
                  key: "displayName",
                  title: i18next.t("general:Display name"),
                  width: "25%",
                  render: (row: any, index, update) => (
                    <Input
                      value={row.displayName ?? ""}
                      placeholder="e.g., Read Files"
                      onChange={(e) => update({displayName: e.target.value})}
                    />
                  ),
                },
                {
                  key: "description",
                  title: i18next.t("general:Description"),
                  width: "40%",
                  render: (row: any, index, update) => (
                    <Input
                      value={row.description ?? ""}
                      placeholder="e.g., Allow reading your files and documents"
                      onChange={(e) => update({description: e.target.value})}
                    />
                  ),
                },
              ]}
            />
          </FormRow>
          <FormRow labelKey="general:Custom scopes" block>
            <EditableTable
              rows={application.customScopes ?? []}
              onChange={(rows) => updateField("customScopes", rows)}
              newRow={() => ({scope: "", displayName: "", description: ""})}
              columns={[
                {
                  key: "scope",
                  title: i18next.t("general:Name"),
                  width: 200,
                  render: (row: any, _i, patch) => (
                    <Input value={row.scope ?? ""} onChange={(e) => patch({scope: e.target.value})} />
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
                  key: "description",
                  title: i18next.t("general:Description"),
                  render: (row: any, _i, patch) => (
                    <Input value={row.description ?? ""} onChange={(e) => patch({description: e.target.value})} />
                  ),
                },
              ]}
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
          {application.tokenFormat === "JWT-Custom" ? (
            <FormRow labelKey="general:Token attributes" block>
              <EditableTable
                rows={application.tokenAttributes ?? []}
                onChange={(rows) => updateField("tokenAttributes", rows)}
                newRow={() => ({name: "", category: "Static Value", value: "", type: "Array"})}
                columns={[
                  {
                    key: "name",
                    title: i18next.t("general:Name"),
                    width: 180,
                    render: (row: any, _i, patch) => (
                      <Input value={row.name ?? ""} onChange={(e) => patch({name: e.target.value})} />
                    ),
                  },
                  {
                    key: "category",
                    title: i18next.t("general:Category"),
                    width: 160,
                    render: (row: any, _i, patch) => (
                      <SelectField
                        value={row.category ?? "Static Value"}
                        onChange={(v) => patch({category: v, value: ""})}
                        options={enumSelectOptions(TOKEN_ATTRIBUTE_CATEGORIES)}
                      />
                    ),
                  },
                  {
                    key: "value",
                    title: i18next.t("webhook:Value"),
                    // an "Existing Field" attribute copies one of the user's own
                    // fields, so the value is picked rather than typed
                    render: (row: any, _i, patch) =>
                      row.category === "Existing Field" ? (
                        <SelectField
                          value={row.value ?? ""}
                          onChange={(v) => patch({value: v})}
                          options={TOKEN_ATTRIBUTE_USER_FIELDS.map((field) => ({id: field, name: field}))}
                        />
                      ) : (
                        <Input value={row.value ?? ""} onChange={(e) => patch({value: e.target.value})} />
                      ),
                  },
                  {
                    key: "type",
                    title: i18next.t("general:Type"),
                    width: 140,
                    render: (row: any, _i, patch) => (
                      <SelectField
                        value={row.type}
                        onChange={(v) => patch({type: v})}
                        options={TOKEN_ATTRIBUTE_TYPES.map((item) => ({id: item, name: item}))}
                      />
                    ),
                  },
                ]}
              />
            </FormRow>
          ) : null}
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
        </TabsContent>

        <TabsContent value="saml" className={menuMode === "vertical" ? "mt-0 min-w-0 flex-1" : undefined}>
          <FormRow labelKey="application:SAML reply URL">
            <Input value={application.samlReplyUrl ?? ""} onChange={(e) => updateField("samlReplyUrl", e.target.value)} />
          </FormRow>
          <FormRow labelKey="application:Enable SAML compression">
            <Switch
              checked={!!application.enableSamlCompress}
              onCheckedChange={(v) => updateField("enableSamlCompress", v)}
            />
          </FormRow>
          <FormRow labelKey="application:Enable SAML C14N10">
            <Switch
              checked={!!application.enableSamlC14n10}
              onCheckedChange={(v) => updateField("enableSamlC14n10", v)}
            />
          </FormRow>
          {application.enableSamlC14n10 ? (
            <FormRow labelKey="application:SAML C14N10 prefix">
              <Input
                value={application.samlC14nPrefix ?? ""}
                placeholder="xs"
                onChange={(e) => updateField("samlC14nPrefix", e.target.value)}
              />
            </FormRow>
          ) : null}
          <FormRow labelKey="application:Use Email as NameID">
            <Switch
              checked={!!application.useEmailAsSamlNameId}
              onCheckedChange={(v) => updateField("useEmailAsSamlNameId", v)}
            />
          </FormRow>
          <FormRow labelKey="application:Enable SAML POST binding">
            <Switch
              checked={!!application.enableSamlPostBinding}
              onCheckedChange={(v) => updateField("enableSamlPostBinding", v)}
            />
          </FormRow>
          <FormRow labelKey="application:SAML hash algorithm">
            <SelectField
              value={application.samlHashAlgorithm ?? "SHA256"}
              onChange={(v) => updateField("samlHashAlgorithm", v)}
              options={SAML_HASH_ALGORITHMS.map((item) => ({id: item, name: item}))}
            />
          </FormRow>
          <FormRow labelKey="application:Disable SAML attributes">
            <Switch
              checked={!!application.disableSamlAttributes}
              onCheckedChange={(v) => updateField("disableSamlAttributes", v)}
            />
          </FormRow>
          <FormRow labelKey="application:Enable SAML assertion signature">
            <Switch
              checked={!!application.enableSamlAssertionSignature}
              onCheckedChange={(v) => updateField("enableSamlAssertionSignature", v)}
            />
          </FormRow>
          <FormRow labelKey="general:SAML attributes" block>
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
          {/* both are generated from the saved application, so only after it exists */}
          {mode === "add" ? null : (
            <>
              <FormRow labelKey="application:SAML metadata" block>
                <div className="space-y-2">
                  <CodeEditor language="xml" value={samlMetadata} readOnly onChange={() => undefined} />
                  <Button variant="outline" size="sm" onClick={() => copyToClipboard(samlMetadataUrl)}>
                    <Copy />
                    {i18next.t("application:Copy SAML metadata URL")}
                  </Button>
                </div>
              </FormRow>
              <FormRow labelKey="application:IdP-initiated SSO URL" block>
                <div className="space-y-2">
                  <div className="flex items-center gap-2">
                    <LinkIcon className="h-4 w-4 shrink-0 text-muted-foreground" />
                    <Input value={idpInitiatedSsoUrl} readOnly />
                  </div>
                  <Button variant="outline" size="sm" onClick={() => copyToClipboard(idpInitiatedSsoUrl)}>
                    <Copy />
                    {i18next.t("application:Copy IdP-initiated SSO URL")}
                  </Button>
                </div>
              </FormRow>
            </>
          )}
        </TabsContent>

        <TabsContent value="providers" className={menuMode === "vertical" ? "mt-0 min-w-0 flex-1" : undefined}>
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
                  // only an identity provider can prompt for a link
                  render: (row: any, _i, patch) =>
                    LINKABLE_PROVIDER_CATEGORIES.includes(resolveProvider(row)?.category) ? (
                      <Switch checked={!!row.prompted} onCheckedChange={(v) => patch({prompted: v})} />
                    ) : null,
                },
                {
                  key: "category",
                  title: i18next.t("general:Category"),
                  width: 110,
                  render: (row: any) => {
                    const provider = resolveProvider(row);
                    if (!provider?.category) {
                      return null;
                    }
                    return (
                      <a
                        href={`/providers/${provider.owner}/${provider.name}`}
                        target="_blank"
                        rel="noreferrer"
                        className="underline-offset-4 hover:underline"
                      >
                        {provider.category}
                      </a>
                    );
                  },
                },
                {
                  key: "type",
                  title: i18next.t("general:Type"),
                  width: 110,
                  render: (row: any) => resolveProvider(row)?.type ?? null,
                },
                {
                  key: "countryCodes",
                  title: i18next.t("user:Country/Region"),
                  width: 200,
                  // an SMS provider can be limited to the countries it serves
                  render: (row: any, _i, patch) =>
                    resolveProvider(row)?.category === "SMS" ? (
                      <MultiSelect
                        value={row.countryCodes ?? ["All"]}
                        onChange={(v) => patch({countryCodes: v})}
                        options={[
                          {value: "All", label: i18next.t("general:All")},
                          ...Setting.getCountryCodeData(application.organizationObj?.countryCodes).map((country: any) => ({
                            value: country.code,
                            label: `${country.name} (+${country.phone})`,
                          })),
                        ]}
                      />
                    ) : null,
                },
                {
                  key: "bindingRule",
                  title: i18next.t("provider:Binding rule"),
                  width: 200,
                  render: (row: any, _i, patch) =>
                    LINKABLE_PROVIDER_CATEGORIES.includes(resolveProvider(row)?.category) ? (
                      <MultiSelect
                        value={row.bindingRule?.length ? row.bindingRule : ["Email", "Phone", "Name"]}
                        onChange={(v) => patch({bindingRule: v})}
                        options={enumOptions(PROVIDER_BINDING_RULES)}
                      />
                    ) : null,
                },
                {
                  key: "signupGroup",
                  title: i18next.t("provider:Signup group"),
                  width: 150,
                  render: (row: any, _i, patch) =>
                    ["OAuth", "Web3"].includes(resolveProvider(row)?.category) ? (
                      <Input value={row.signupGroup ?? ""} onChange={(e) => patch({signupGroup: e.target.value})} />
                    ) : null,
                },
                {
                  key: "rule",
                  title: i18next.t("application:Rule"),
                  width: 170,
                  // the rule means something different for each kind of provider
                  render: (row: any, _i, patch) => {
                    const map = getProviderRuleMap(resolveProvider(row));
                    if (!map) {
                      return null;
                    }
                    return (
                      <SelectField
                        value={normalizeProviderRule(resolveProvider(row), row.rule)}
                        onChange={(v) => patch({rule: v})}
                        options={enumSelectOptions(map)}
                      />
                    );
                  },
                },
              ]}
            />
          </FormRow>
        </TabsContent>

        <TabsContent value="ui-customization" className={menuMode === "vertical" ? "mt-0 min-w-0 flex-1" : undefined}>
          <FormRow labelKey="application:Org choice mode">
            <SelectField
              value={application.orgChoiceMode ?? "None"}
              onChange={(v) => updateField("orgChoiceMode", v)}
              options={[
                {id: "None", name: i18next.t("general:None")},
                {id: "Select", name: i18next.t("application:Select")},
                {id: "Input", name: i18next.t("application:Input")},
              ]}
            />
          </FormRow>
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
                      options={enumSelectOptions(SIGNIN_METHOD_RULES[row.name] ?? {})}
                    />
                  ),
                },
              ]}
            />
          </FormRow>
          <FormRow labelKey="provider:Signup HTML" block>
            <CodeEditor
              language="html"
              value={application.signupHtml ?? ""}
              onChange={(v) => updateField("signupHtml", v)}
            />
          </FormRow>
          <FormRow labelKey="provider:Signin HTML" block>
            <CodeEditor
              language="html"
              value={application.signinHtml ?? ""}
              onChange={(v) => updateField("signinHtml", v)}
            />
          </FormRow>
          <FormRow labelKey="application:Signin items" block>
            <EditableTable
              title={
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() =>
                    updateField("signinItems", [
                      ...(application.signinItems ?? []),
                      // a custom item is a free-form HTML block, named so it stays unique
                      {name: `Text ${Date.now()}`, visible: true, isCustom: true, label: "", placeholder: "", rule: "None"},
                    ])
                  }
                >
                  {i18next.t("general:Add custom item")}
                </Button>
              }
              rows={application.signinItems ?? []}
              onChange={(rows) => updateField("signinItems", rows)}
              newRow={() => ({
                name: "Logo",
                visible: true,
                label: "",
                customCss: SigninTableDefaultCssMap["Logo"],
                placeholder: "",
                rule: "None",
              })}
              columns={[
                {
                  key: "name",
                  title: i18next.t("general:Name"),
                  width: 190,
                  render: (row: any, _i, patch) => (
                    <SelectField
                      value={row.name}
                      onChange={(v) => patch({name: v, customCss: SigninTableDefaultCssMap[v] ?? ""})}
                      options={SIGNIN_ITEM_NAMES.map((item) => ({id: item.name, name: i18next.t(item.labelKey)}))}
                    />
                  ),
                },
                {
                  key: "visible",
                  title: i18next.t("organization:Visible"),
                  width: 90,
                  render: (row: any, _i, patch) => (
                    <Switch checked={!!row.visible} onCheckedChange={(v) => patch({visible: v, required: v})} />
                  ),
                },
                {
                  key: "label",
                  title: i18next.t("signup:Label"),
                  width: 170,
                  render: (row: any, _i, patch) => (
                    <Input value={row.label ?? ""} onChange={(e) => patch({label: e.target.value})} />
                  ),
                },
                {
                  key: "placeholder",
                  title: i18next.t("signup:Placeholder"),
                  width: 170,
                  render: (row: any, _i, patch) => (
                    <Input value={row.placeholder ?? ""} onChange={(e) => patch({placeholder: e.target.value})} />
                  ),
                },
                {
                  key: "customCss",
                  title: i18next.t("application:Custom CSS"),
                  width: 200,
                  render: (row: any, _i, patch) => (
                    <Input
                      value={row.customCss ?? SigninTableDefaultCssMap[row.name] ?? ""}
                      onChange={(e) => patch({customCss: e.target.value || SigninTableDefaultCssMap[row.name]})}
                    />
                  ),
                },
                {
                  key: "rule",
                  title: i18next.t("application:Rule"),
                  width: 170,
                  render: (row: any, _i, patch) => {
                    const options = getSigninItemRuleOptions(row.name);
                    if (options.length === 0) {
                      return null;
                    }
                    return <SelectField value={row.rule} onChange={(v) => patch({rule: v})} options={options} />;
                  },
                },
              ]}
            />
          </FormRow>
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
                  key: "type",
                  title: i18next.t("general:Type"),
                  width: 160,
                  render: (row: any, _i, patch) => (
                    <SelectField
                      value={row.type ?? "Input"}
                      onChange={(v) => patch({type: v})}
                      options={enumSelectOptions(SIGNUP_ITEM_TYPES)}
                    />
                  ),
                },
                {
                  key: "rule",
                  title: i18next.t("application:Rule"),
                  width: 160,
                  render: (row: any, _i, patch) => {
                    const map = SIGNUP_ITEM_RULES[row.name];
                    if (!map) {
                      return null;
                    }
                    return (
                      <SelectField
                        value={row.rule}
                        onChange={(v) => patch({rule: v})}
                        options={enumSelectOptions(map)}
                      />
                    );
                  },
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
                  key: "customCss",
                  title: i18next.t("application:Custom CSS"),
                  width: 200,
                  render: (row: any, _i, patch) => (
                    <Input
                      value={row.customCss ?? SignupTableDefaultCssMap[row.name] ?? ""}
                      onChange={(e) => patch({customCss: e.target.value || SignupTableDefaultCssMap[row.name]})}
                    />
                  ),
                },
                {
                  key: "options",
                  title: i18next.t("signup:Options"),
                  width: 200,
                  // only a choice item has options to offer
                  render: (row: any, _i, patch) =>
                    row.type === "Single Choice" || row.type === "Multiple Choices" ? (
                      <TagsInput value={row.options ?? []} onChange={(v) => patch({options: v})} />
                    ) : null,
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
          <FormRow labelKey="application:Background URL">
            <Input
              value={application.formBackgroundUrl ?? ""}
              onChange={(e) => updateField("formBackgroundUrl", e.target.value)}
            />
          </FormRow>
          <FormRow labelKey="application:Background URL Mobile">
            <Input
              value={application.formBackgroundUrlMobile ?? ""}
              onChange={(e) => updateField("formBackgroundUrlMobile", e.target.value)}
            />
          </FormRow>
          <FormRow labelKey="application:Custom CSS" block>
            <CodeEditor
              language="css"
              value={application.formCss ?? ""}
              onChange={(v) => updateField("formCss", v)}
            />
          </FormRow>
          <FormRow labelKey="application:Custom CSS Mobile" block>
            <CodeEditor
              language="css"
              value={application.formCssMobile ?? ""}
              onChange={(v) => updateField("formCssMobile", v)}
            />
          </FormRow>
          <FormRow labelKey="application:Form position">
            <div className="flex flex-wrap gap-2">
              {[
                {value: 1, labelKey: "application:Left"},
                {value: 2, labelKey: "application:Center"},
                {value: 3, labelKey: "application:Right"},
                {value: 4, labelKey: "application:Enable side panel"},
              ].map((item) => (
                <Button
                  key={item.value}
                  type="button"
                  size="sm"
                  variant={application.formOffset === item.value ? "default" : "outline"}
                  onClick={() => updateField("formOffset", item.value)}
                >
                  {i18next.t(item.labelKey)}
                </Button>
              ))}
            </div>
          </FormRow>
          {/* formOffset 4 is "Enable side panel", the only position that shows it */}
          {application.formOffset === 4 ? (
            <FormRow labelKey="application:Side panel HTML" block>
              <CodeEditor
                language="html"
                value={application.formSideHtml ?? ""}
                onChange={(v) => updateField("formSideHtml", v)}
              />
            </FormRow>
          ) : null}
          <FormRow labelKey="theme:Customize theme" block>
            <ThemeEditor
              themeData={application.themeData}
              onChange={(next) => updateField("themeData", next)}
              followLabelKey="application:Follow organization theme"
            />
          </FormRow>
          <FormRow labelKey="application:Header HTML" block>
            <CodeEditor
              language="html"
              value={application.headerHtml ?? ""}
              onChange={(v) => updateField("headerHtml", v)}
            />
          </FormRow>
          <FormRow labelKey="application:Page HTML" block>
            <CodeEditor
              language="html"
              value={application.pageHtml ?? ""}
              onChange={(v) => updateField("pageHtml", v)}
            />
          </FormRow>
          <FormRow labelKey="application:Footer HTML" block>
            <CodeEditor
              language="html"
              value={application.footerHtml ?? ""}
              onChange={(v) => updateField("footerHtml", v)}
            />
            <div className="mt-2 flex flex-wrap gap-2">
              <Button variant="outline" size="sm" onClick={() => updateField("footerHtml", Setting.getDefaultFooterContent())}>
                {i18next.t("general:Reset to Default")}
              </Button>
              <Button variant="outline" size="sm" onClick={() => updateField("footerHtml", Setting.getEmptyFooterContent())}>
                {i18next.t("application:Reset to Empty")}
              </Button>
            </div>
          </FormRow>
        </TabsContent>

        <TabsContent value="security" className={menuMode === "vertical" ? "mt-0 min-w-0 flex-1" : undefined}>
          <FormRow labelKey="application:Token cert">
            <SearchableSelect
              value={application.cert ?? ""}
              onChange={(v) => updateField("cert", v)}
              options={certs}
            />
          </FormRow>
          <FormRow labelKey="application:Client cert">
            <SearchableSelect
              value={application.clientCert ?? ""}
              onChange={(v) => updateField("clientCert", v)}
              options={certs}
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
          <FormRow labelKey="application:Code resend timeout">
            <Input
              type="number"
              value={application.codeResendTimeout ?? 60}
              onChange={(e) => updateField("codeResendTimeout", Setting.myParseInt(e.target.value))}
            />
          </FormRow>
          <FormRow labelKey="general:IP whitelist">
            <Input
              placeholder={application.organizationObj?.ipWhitelist}
              value={application.ipWhitelist ?? ""}
              onChange={(e) => updateField("ipWhitelist", e.target.value)}
            />
          </FormRow>
          <FormRow labelKey="signup:Terms of Use">
            <div className="flex flex-wrap items-center gap-2">
              <Input
                className="min-w-[16rem] flex-1"
                value={application.termsOfUse ?? ""}
                onChange={(e) => updateField("termsOfUse", e.target.value)}
              />
              {/* the upload writes the stored URL back onto the application, so it
                  can only be done once the application exists */}
              {mode === "add" ? null : (
                <>
                  <input
                    ref={termsFileRef}
                    type="file"
                    accept=".html"
                    className="hidden"
                    onChange={(e) => uploadTermsOfUse(e.target.files?.[0])}
                  />
                  <Button
                    variant="outline"
                    loading={uploadingTerms}
                    onClick={() => termsFileRef.current?.click()}
                  >
                    <Upload />
                    {i18next.t("general:Click to Upload")}
                  </Button>
                </>
              )}
            </div>
          </FormRow>
        </TabsContent>

        <TabsContent value="reverse-proxy" className={menuMode === "vertical" ? "mt-0 min-w-0 flex-1" : undefined}>
          <FormRow labelKey="provider:Domain">
            <Input
              value={application.domain ?? ""}
              placeholder="e.g., blog.example.com"
              onChange={(e) => updateField("domain", e.target.value)}
            />
          </FormRow>
          <FormRow labelKey="application:Other domains">
            <TagsInput value={application.otherDomains ?? []} onChange={(v) => updateField("otherDomains", v)} />
          </FormRow>
          <FormRow labelKey="application:Upstream host">
            <Input
              value={application.upstreamHost ?? ""}
              placeholder="e.g., localhost:8080 or 192.168.1.100:3000"
              onChange={(e) => updateField("upstreamHost", e.target.value)}
            />
          </FormRow>
          <FormRow labelKey="provider:SSL mode">
            <SelectField
              value={application.sslMode ?? ""}
              onChange={(v) => updateField("sslMode", v)}
              options={SSL_MODES.map((item) => ({id: item, name: item === "" ? i18next.t("general:None") : item}))}
            />
          </FormRow>
          <FormRow labelKey="application:SSL cert">
            <SearchableSelect
              value={application.sslCert ?? ""}
              onChange={(v) => updateField("sslCert", v)}
              options={[{value: "", label: i18next.t("general:None")}, ...certs]}
            />
          </FormRow>
        </TabsContent>
      </Tabs>
    </EditPageShell>
  );
}
