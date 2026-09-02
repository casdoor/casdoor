
import i18next from "i18next";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Switch} from "@/components/ui/switch";
import {CodeEditor} from "@/components/common/CodeEditor";
import {SelectField} from "@/components/common/SelectField";
import {TagsInput} from "@/components/common/TagsInput";
import {ThemeEditor} from "@/components/common/ThemeEditor";
import {EditableTable} from "@/components/crud/EditableTable";
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
  "WeChat": {
    "Tab": {i18nKey: "general:Tab"},
    "Login page": {i18nKey: "general:Login page"},
  },
  "Device login": {
    "Tab": {i18nKey: "general:Tab"},
    "Login page": {i18nKey: "general:Login page"},
  },
};

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

const SIGNUP_ITEM_NAMES = [
  "ID", "Username", "Display name", "Affiliation", "ID card", "Country/Region", "Email", "Phone",
  "Email or Phone", "Phone or Email", "Password", "Confirm password", "Invitation code", "Agreement",
  "Signup button", "Providers", "Text 1", "Text 2", "Text 3", "Text 4", "Text 5", "Languages",
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

/** The "UI Customization" tab: the blocks the sign-in and sign-up pages are built from. */
export function ApplicationUiCustomizationTab({application, updateField}: ApplicationTabProps) {
  return (
    <>
      <FormRow block labelKey="application:Org choice mode">
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
    </>
  );
}
