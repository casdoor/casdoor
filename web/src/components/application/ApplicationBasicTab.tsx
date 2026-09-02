
import i18next from "i18next";
import {Copy} from "lucide-react";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Switch} from "@/components/ui/switch";
import {RadioGroup, RadioGroupItem} from "@/components/ui/radio-group";
import {SearchableSelect, type SearchableOption} from "@/components/common/SearchableSelect";
import {SelectField} from "@/components/common/SelectField";
import {TagsInput} from "@/components/common/TagsInput";
import {FormRow} from "@/components/crud/FormRow";
import type {ApplicationTabProps, MenuMode} from "@/components/application/types";
import * as Setting from "@/lib/setting";

const APPLICATION_TYPES = ["All", "Web", "Native", "SPA"];
const APPLICATION_CATEGORIES = ["Default", "OAuth", "SAML", "CAS"];

interface ApplicationBasicTabProps extends ApplicationTabProps {
  mode: string;
  account: any;
  organizations: SearchableOption[];
  menuMode: MenuMode;
  setMenuMode: (value: MenuMode) => void;
}

/** The "Basic" tab: what the application is called, how it looks, and where its pages live. */
export function ApplicationBasicTab({
  application,
  updateField,
  mode,
  account,
  organizations,
  menuMode,
  setMenuMode,
}: ApplicationBasicTabProps) {
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
  return (
    <>
      {/* antd puts these next to its live sign-in previews; this frontend has no
          previews, so the links live on their own row at the top of the tab */}
      {mode === "add" ? null : (
        <FormRow labelKey="general:Login page" block>
          <div className="flex flex-wrap gap-2">
            <Button variant="outline" size="sm" onClick={() => Setting.copyToClipboard(`${window.location.origin}${signInUrl}`)}>
              <Copy />
              {i18next.t("application:Copy signin page URL")}
            </Button>
            <Button variant="outline" size="sm" onClick={() => Setting.copyToClipboard(`${window.location.origin}${signUpUrl}`)}>
              <Copy />
              {i18next.t("application:Copy signup page URL")}
            </Button>
            {Setting.hasPromptPage(application) ? (
              <Button variant="outline" size="sm" onClick={() => Setting.copyToClipboard(`${window.location.origin}${promptUrl}`)}>
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
              <img src={application.logo} alt="logo" className="h-14 max-w-[240px] rounded-md border bg-white object-contain p-1.5" />
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
    </>
  );
}
