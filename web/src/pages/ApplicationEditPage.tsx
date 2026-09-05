import * as React from "react";
import i18next from "i18next";
import {useNavigate, useParams} from "react-router-dom";
import {UnauthorizedPage} from "@/components/common/UnauthorizedPage";
import {Button} from "@/components/ui/button";
import {Tabs, TabsContent, TabsList, TabsTrigger} from "@/components/ui/tabs";
import {Loading} from "@/components/common/Loading";
import {ApplicationExportButton, ApplicationImportModal} from "@/components/application/ApplicationImportExport";
import {ApplicationAuthenticationTab} from "@/components/application/ApplicationAuthenticationTab";
import {ApplicationBasicTab} from "@/components/application/ApplicationBasicTab";
import {ApplicationOidcOauthTab} from "@/components/application/ApplicationOidcOauthTab";
import {ApplicationProvidersTab} from "@/components/application/ApplicationProvidersTab";
import {ApplicationReverseProxyTab} from "@/components/application/ApplicationReverseProxyTab";
import {ApplicationSamlTab} from "@/components/application/ApplicationSamlTab";
import {ApplicationSecurityTab} from "@/components/application/ApplicationSecurityTab";
import {ApplicationUiCustomizationTab} from "@/components/application/ApplicationUiCustomizationTab";
import type {MenuMode} from "@/components/application/types";
import {EditPageShell} from "@/components/crud/EditPageShell";
import {formGridClass} from "@/components/crud/FormRow";
import {useAccount} from "@/hooks/use-account";
import {useEditRecord} from "@/hooks/use-edit-record";
import {useCertOptions, useOrganizationOptions, useProviderList, useProviderOptions} from "@/hooks/use-options";
import {getModeTitleKey, submitEdit} from "@/lib/crud";
import * as ApplicationBackend from "@/backend/ApplicationBackend";
import * as Setting from "@/lib/setting";
import {cn} from "@/lib/utils";

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

  const {record: application, updateField, setRecord, loading, denied, mode, setMode, reload} = useEditRecord<any>({
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

  const idpInitiatedSsoUrl =
    `${window.location.origin}/login/saml/authorize/${application.owner}/${encodeURIComponent(applicationName)}`;
  const samlMetadataUrl =
    `${window.location.origin}/api/saml/metadata?application=admin/${encodeURIComponent(applicationName)}` +
    `&enablePostBinding=${enableSamlPostBinding}`;

  const save = async(exitAfterSave: boolean) => {
    // a provider row whose provider has since been deleted, or a sign-in method
    // this Casdoor does not know, would be posted back as a dangling reference
    const knownProviders = providerObjs.map((item: any) => item.name);
    const applicationProviders = (application.providers ?? []).filter((item: any) =>
      knownProviders.includes(item.name),
    );
    const signinMethods = (application.signinMethods ?? []).filter((item: any) =>
      ["Password", "Verification code", "WebAuthn", "LDAP", "Face ID", "Device login", "WeChat"].includes(item.name),
    );

    // antd trims every custom scope and refuses to save one without a scope name,
    // which is also what the backend's validateCustomScopes() enforces
    const customScopes = (application.customScopes ?? []).map((item: any) => ({
      ...item,
      scope: (item?.scope ?? "").trim(),
      displayName: (item?.displayName ?? "").trim(),
      description: (item?.description ?? "").trim(),
    }));
    if (customScopes.some((scope: any) => scope.scope === "")) {
      Setting.showMessage("error", `${i18next.t("general:Name")}: ${i18next.t("general:This field is required")}`);
      return;
    }

    const isAdd = mode === "add";
    setSaving(true);
    await submitEdit({
      mode,
      record: {...Setting.deepCopy(application), providers: applicationProviders, signinMethods, customScopes},
      add: (record) => ApplicationBackend.addApplication(record),
      update: (record) => ApplicationBackend.updateApplication("admin", applicationName, record),
      onSaved: () => {
        setMode("edit");
        if (exitAfterSave) {
          navigate("/applications");
        } else if (application.name !== applicationName) {
          navigate(`/applications/${application.organization}/${application.name}`, {replace: true});
        } else if (isAdd) {
          // the client ID, the client secret and the token format are generated by the
          // server, so the application has to be reloaded to pick them up
          reload();
        }
      },
    });
    setSaving(false);
  };

  const tabClassName = cn(formGridClass, menuMode === "vertical" && "mt-0 min-w-0 flex-1");

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

        <TabsContent value="basic" className={tabClassName}>
          <ApplicationBasicTab
            application={application}
            updateField={updateField}
            mode={mode}
            account={account}
            organizations={organizations}
            menuMode={menuMode}
            setMenuMode={setMenuMode}
          />
        </TabsContent>

        <TabsContent value="authentication" className={tabClassName}>
          <ApplicationAuthenticationTab application={application} updateField={updateField} />
        </TabsContent>

        <TabsContent value="oidc-oauth" className={tabClassName}>
          <ApplicationOidcOauthTab application={application} updateField={updateField} />
        </TabsContent>

        <TabsContent value="saml" className={tabClassName}>
          <ApplicationSamlTab
            application={application}
            updateField={updateField}
            mode={mode}
            samlMetadata={samlMetadata}
            samlMetadataUrl={samlMetadataUrl}
            idpInitiatedSsoUrl={idpInitiatedSsoUrl}
          />
        </TabsContent>

        <TabsContent value="providers" className={tabClassName}>
          <ApplicationProvidersTab
            application={application}
            updateField={updateField}
            providers={providers}
            providerObjs={providerObjs}
          />
        </TabsContent>

        <TabsContent value="ui-customization" className={tabClassName}>
          <ApplicationUiCustomizationTab application={application} updateField={updateField} />
        </TabsContent>

        <TabsContent value="security" className={tabClassName}>
          <ApplicationSecurityTab
            application={application}
            updateField={updateField}
            mode={mode}
            account={account}
            certs={certs}
          />
        </TabsContent>

        <TabsContent value="reverse-proxy" className={tabClassName}>
          <ApplicationReverseProxyTab application={application} updateField={updateField} certs={certs} />
        </TabsContent>
      </Tabs>
    </EditPageShell>
  );
}
