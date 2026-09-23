
import i18next from "i18next";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Switch} from "@/components/ui/switch";
import {CodeEditor} from "@/components/common/CodeEditor";
import {SelectField} from "@/components/common/SelectField";
import {TagsInput} from "@/components/common/TagsInput";
import {ThemeEditor} from "@/components/common/ThemeEditor";
import {EditableTable} from "@/components/crud/EditableTable";
import {ApplicationPromptPreview, ApplicationSignupSigninPreview} from "@/components/application/ApplicationPreview";
import {FormRow} from "@/components/crud/FormRow";
import type {ApplicationTabProps} from "@/components/application/types";
import {enumSelectOptions, type EnumMap} from "@/lib/enum-labels";
import {SigninTableDefaultCssMap} from "@/lib/signin-css";
import {SignupTableDefaultCssMap} from "@/lib/signup-css";
import * as Setting from "@/lib/setting";

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
  "Magic link": {
    "Sign in only": {i18nKey: "application:Sign in only"},
    "Sign in or sign up": {i18nKey: "application:Sign in or sign up"},
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

// a rule left over from another method (e.g. "All" on WebAuthn) hides the method on the login page
function getDefaultSigninMethodRule(name: string): string {
  const rules = Object.keys(SIGNIN_METHOD_RULES[name] ?? {});
  return rules.length > 0 ? rules[0] : "None";
}

const SIGNIN_METHOD_NAMES: {name: string; labelKey: string}[] = [
  {name: "Password", labelKey: "general:Password"},
  {name: "Verification code", labelKey: "login:Verification code"},
  {name: "Magic link", labelKey: "login:Magic link"},
  {name: "WebAuthn", labelKey: "login:WebAuthn"},
  {name: "LDAP", labelKey: "login:LDAP"},
  {name: "Face ID", labelKey: "login:Face ID"},
  {name: "Device login", labelKey: "login:Device login"},
  {name: "WeChat", labelKey: "login:WeChat"},
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

const SIGNUP_ITEM_NAMES: {name: string; labelKey: string}[] = [
  {name: "Username", labelKey: "signup:Username"},
  {name: "ID", labelKey: "general:ID"},
  {name: "Display name", labelKey: "general:Display name"},
  {name: "First name", labelKey: "general:First name"},
  {name: "Last name", labelKey: "general:Last name"},
  {name: "Affiliation", labelKey: "user:Affiliation"},
  {name: "Gender", labelKey: "user:Gender"},
  {name: "Bio", labelKey: "user:Bio"},
  {name: "Tag", labelKey: "general:Tag"},
  {name: "Education", labelKey: "user:Education"},
  {name: "Country/Region", labelKey: "user:Country/Region"},
  {name: "ID card", labelKey: "user:ID card"},
  {name: "Password", labelKey: "general:Password"},
  {name: "Confirm password", labelKey: "general:Confirm"},
  {name: "Email", labelKey: "general:Email"},
  {name: "Phone", labelKey: "general:Phone"},
  {name: "Email or Phone", labelKey: "general:Email or Phone"},
  {name: "Phone or Email", labelKey: "general:Phone or Email"},
  {name: "Invitation code", labelKey: "application:Invitation code"},
  {name: "Agreement", labelKey: "signup:Agreement"},
  {name: "Signup button", labelKey: "signup:Signup button"},
  {name: "Providers", labelKey: "application:Providers"},
  {name: "Languages", labelKey: "general:Languages"},
  {name: "Text 1", labelKey: "signup:Text 1"},
  {name: "Text 2", labelKey: "signup:Text 2"},
  {name: "Text 3", labelKey: "signup:Text 3"},
  {name: "Text 4", labelKey: "signup:Text 4"},
  {name: "Text 5", labelKey: "signup:Text 5"},
];

/** the signin items whose text the antd table lets you relabel */
const LABELED_SIGNIN_ITEMS = ["Username", "Password", "Verification code", "Signup link", "Forgot password?", "Login button"];
/** signup items that are page furniture rather than fields, so they cannot be required or prompted */
const NON_FIELD_SIGNUP_ITEMS = ["Signup button", "Providers", "Languages"];
const NO_REGEX_SIGNUP_ITEMS = ["Password", "Confirm password", "Signup button", "Provider", "Providers", "Languages"];

/** a picked item cannot be picked again by another row */
function getUnusedItemOptions(names: {name: string; labelKey: string}[], rows: any[], index: number) {
  return names
    .filter((item) => !rows.some((other: any, i: number) => i !== index && other.name === item.name))
    .map((item) => ({id: item.name, name: i18next.t(item.labelKey)}));
}

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

/** The "UI Customization" tab: the blocks the sign-in and sign-up pages are built from. */
export function ApplicationUiCustomizationTab({application, updateField}: ApplicationTabProps) {
  return (
    <>
      <FormRow block labelKey="application:Org choice mode">
        <SelectField
          value={application.orgChoiceMode || "None"}
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
          newRow={(application.signinMethods ?? []).length < SIGNIN_METHOD_NAMES.length ? () => ({
            name: Setting.getNewRowNameForTable(application.signinMethods ?? [], "Please select a signin method"),
            displayName: "",
            rule: "None",
          }) : undefined}
          canDelete={() => (application.signinMethods ?? []).length > 1}
          columns={[
            {
              key: "name",
              title: i18next.t("general:Name"),
              width: 200,
              render: (row: any, index, patch) => (
                <SelectField
                  value={row.name}
                  onChange={(v) => patch({name: v, displayName: v, rule: getDefaultSigninMethodRule(v)})}
                  options={SIGNIN_METHOD_NAMES
                    .filter((item) => !(application.signinMethods ?? []).some((other: any, i: number) => i !== index && other.name === item.name))
                    .map((item) => ({id: item.name, name: i18next.t(item.labelKey)}))}
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
              render: (row: any, _i, patch) =>
                SIGNIN_METHOD_RULES[row.name] ? (
                  <SelectField
                    value={row.rule}
                    onChange={(v) => patch({rule: v})}
                    options={enumSelectOptions(SIGNIN_METHOD_RULES[row.name])}
                  />
                ) : null,
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
            name: Setting.getNewRowNameForTable(application.signinItems ?? [], "Please select a signin item"),
            visible: true,
            required: true,
            rule: "None",
          })}
          columns={[
            {
              key: "name",
              title: i18next.t("general:Name"),
              width: 190,
              render: (row: any, index, patch) =>
                Setting.isCustomFormItem(row) ? (
                  <Input value={row.name ?? ""} disabled />
                ) : (
                  <SelectField
                    value={row.name}
                    onChange={(v) => patch({name: v, customCss: SigninTableDefaultCssMap[v] ?? ""})}
                    options={getUnusedItemOptions(SIGNIN_ITEM_NAMES, application.signinItems ?? [], index)}
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
              // a custom item's HTML lives in customCss, which is what the login page renders
              render: (row: any, _i, patch) => {
                if (Setting.isCustomFormItem(row)) {
                  return <Input value={row.customCss ?? ""} onChange={(e) => patch({customCss: e.target.value})} />;
                }
                if (!LABELED_SIGNIN_ITEMS.includes(row.name)) {
                  return null;
                }
                return <Input value={row.label ?? ""} onChange={(e) => patch({label: e.target.value})} />;
              },
            },
            {
              key: "placeholder",
              title: i18next.t("signup:Placeholder"),
              width: 170,
              render: (row: any, _i, patch) =>
                row.name === "Username" || row.name === "Password" ? (
                  <Input value={row.placeholder ?? ""} onChange={(e) => patch({placeholder: e.target.value})} />
                ) : null,
            },
            {
              key: "customCss",
              title: i18next.t("application:Custom CSS"),
              width: 200,
              render: (row: any, _i, patch) =>
                Setting.isCustomFormItem(row) ? null : (
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
          newRow={() => ({
            name: Setting.getNewRowNameForTable(application.signupItems ?? [], "Please select a signup item"),
            visible: true,
            required: true,
            options: [],
            rule: "None",
            customCss: "",
          })}
          canDelete={(row: any) => row.name !== "Signup button"}
          columns={[
            {
              key: "name",
              title: i18next.t("general:Name"),
              width: 190,
              render: (row: any, index, patch) => (
                <SelectField
                  value={row.name}
                  onChange={(v) => patch({name: v, customCss: SignupTableDefaultCssMap[v] ?? ""})}
                  options={getUnusedItemOptions(SIGNUP_ITEM_NAMES, application.signupItems ?? [], index)}
                />
              ),
            },
            {
              key: "visible",
              title: i18next.t("organization:Visible"),
              width: 90,
              render: (row: any, _i, patch) =>
                row.name === "ID" ? null : (
                  <Switch checked={!!row.visible} onCheckedChange={(v) => patch({visible: v, required: v})} />
                ),
            },
            {
              key: "required",
              title: i18next.t("organization:Required"),
              width: 90,
              render: (row: any, _i, patch) =>
                !row.visible || NON_FIELD_SIGNUP_ITEMS.includes(row.name) ? null : (
                  <Switch
                    checked={!!row.required}
                    disabled={row.name === "Password"}
                    onCheckedChange={(v) => patch({required: v})}
                  />
                ),
            },
            {
              key: "prompted",
              title: i18next.t("provider:Prompted"),
              width: 90,
              // a hidden item can be asked for after signup; Country/Region even when shown
              render: (row: any, _i, patch) => {
                if (row.name === "ID" || NON_FIELD_SIGNUP_ITEMS.includes(row.name)) {
                  return null;
                }
                if (row.visible && row.name !== "Country/Region") {
                  return null;
                }
                return <Switch checked={!!row.prompted} onCheckedChange={(v) => patch({prompted: v})} />;
              },
            },
            {
              key: "type",
              title: i18next.t("general:Type"),
              width: 160,
              render: (row: any, _i, patch) => (
                <SelectField
                  value={row.type || "Input"}
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
              render: (row: any, _i, patch) =>
                row.name?.startsWith("Text ") ? null : (
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
              render: (row: any, _i, patch) =>
                row.name?.startsWith("Text ") || NO_REGEX_SIGNUP_ITEMS.includes(row.name) ? null : (
                  <Input value={row.regex ?? ""} onChange={(e) => patch({regex: e.target.value})} />
                ),
            },
          ]}
        />
      </FormRow>
      <FormRow labelKey="general:Preview" block>
        <ApplicationSignupSigninPreview application={application} />
      </FormRow>
      <FormRow block labelKey="application:Background URL">
        <Input
          value={application.formBackgroundUrl ?? ""}
          onChange={(e) => updateField("formBackgroundUrl", e.target.value)}
        />
      </FormRow>
      <FormRow block labelKey="application:Background URL Mobile">
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
      <FormRow block labelKey="application:Form position">
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
      {Setting.hasPromptPage(application) ? (
        <FormRow labelKey="general:Preview" block>
          <ApplicationPromptPreview application={application} />
        </FormRow>
      ) : null}
    </>
  );
}
