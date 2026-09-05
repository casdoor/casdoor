
import i18next from "i18next";
import {Input} from "@/components/ui/input";
import {Switch} from "@/components/ui/switch";
import {MultiSelect} from "@/components/common/MultiSelect";
import {SelectField} from "@/components/common/SelectField";
import type {SearchableOption} from "@/components/common/SearchableSelect";
import {EditableTable} from "@/components/crud/EditableTable";
import {FormRow} from "@/components/crud/FormRow";
import type {ApplicationTabProps} from "@/components/application/types";
import {
  enumLabel,
  enumOptions,
  enumSelectOptions,
  PROVIDER_BINDING_RULES,
  PROVIDER_CAPTCHA_RULES,
  PROVIDER_CODE_RULES,
  PROVIDER_GOOGLE_RULES,
} from "@/lib/enum-labels";
import * as Setting from "@/lib/setting";

/** the provider kinds a user account can be linked to */
const LINKABLE_PROVIDER_CATEGORIES = ["OAuth", "Web3", "SAML"];

/** the methods an Email or SMS provider row can be picked for, "All" being the absence of one */
const CODE_PROVIDER_METHODS = Object.keys(PROVIDER_CODE_RULES).filter((rule) => rule !== "all");

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

/** every spelling the backend's GetProviderByCategoryAndRule() reads as "no method in particular" */
function getCodeProviderRule(row: any) {
  const rule = row?.rule;
  return !rule || rule === "all" || rule === "All" || rule === "None" ? "all" : rule;
}

/** an SMS row is only outranked by one serving every country it serves; an Email row has no countries */
function outranksCountryCodes(row: any, other: any, category: string) {
  if (category !== "SMS") {
    return true;
  }

  const codesOf = (item: any) => (item?.countryCodes?.length ? item.countryCodes : ["All"]);
  const codes = codesOf(row);
  if (codes.some((code: string) => code === "" || code === "All" || code === "all")) {
    return true;
  }

  return codesOf(other).every((code: string) => codes.includes(code));
}

/**
 * Which methods an Email or SMS row is actually used for. The backend takes the
 * first row whose rule names the method and otherwise the first "All" row, so a
 * later row carrying a rule an earlier one already carries is never reached.
 */
function getCodeProviderCoverage(rows: any[], index: number, categoryOf: (row: any) => string | undefined) {
  const row = rows[index];
  const category = categoryOf(row);
  if (category !== "Email" && category !== "SMS") {
    return null;
  }

  const rule = getCodeProviderRule(row);
  const outranks = (other: any) => categoryOf(other) === category && outranksCountryCodes(other, row, category);

  if (rows.some((other, i) => i < index && getCodeProviderRule(other) === rule && outranks(other))) {
    return {methods: [], shadowed: true};
  }

  if (rule !== "all") {
    return {methods: [rule], shadowed: false};
  }

  const named = new Set(rows.filter(outranks).map(getCodeProviderRule));
  return {methods: CODE_PROVIDER_METHODS.filter((method) => !named.has(method)), shadowed: false};
}

interface ApplicationProvidersTabProps extends ApplicationTabProps {
  providers: SearchableOption[];
  providerObjs: any[];
}

/** The "Providers" tab: which OAuth, Email, SMS, Captcha and Storage providers this application uses. */
export function ApplicationProvidersTab({
  application,
  updateField,
  providers,
  providerObjs,
}: ApplicationProvidersTabProps) {
  /**
   * The provider a row points at. The backend embeds it on the row, but a row
   * the user just added only has a name, so fall back to the fetched list.
   */
  const resolveProvider = (row: any) =>
    row?.provider ?? providerObjs.find((item: any) => item.name === row?.name);
  return (
    <>
      <FormRow labelKey="application:Providers" block>
        <EditableTable
          rows={application.providers ?? []}
          onChange={(rows) => updateField("providers", rows)}
          newRow={() => {
            const taken = new Set((application.providers ?? []).map((item: any) => item.name));
            return {
              name: providers.find((option) => !taken.has(option.value))?.value ?? "",
              canSignUp: true,
              canSignIn: true,
              canUnlink: true,
              prompted: false,
              signupGroup: "",
              rule: "None",
            };
          }}
          columns={[
            {
              key: "name",
              title: i18next.t("general:Name"),
              width: 240,
              // the same provider twice would only ever be reached once, so it cannot be picked twice
              render: (row: any, index, patch) => {
                const taken = new Set(
                  (application.providers ?? [])
                    .filter((_: any, i: number) => i !== index)
                    .map((item: any) => item.name),
                );
                return (
                  <SelectField
                    value={row.name}
                    onChange={(v) => patch({name: v})}
                    options={providers
                      .filter((option) => !taken.has(option.value))
                      .map((option) => ({id: option.value, name: option.label as string}))}
                  />
                );
              },
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
                        keywords: `${country.name} ${country.code} ${country.phone}`,
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
            {
              key: "coverage",
              title: i18next.t("application:Used for"),
              width: 220,
              // an Email or SMS row can be shadowed by an earlier one, which is invisible otherwise
              render: (_row: any, index) => {
                const coverage = getCodeProviderCoverage(
                  application.providers ?? [],
                  index,
                  (item: any) => resolveProvider(item)?.category,
                );
                if (!coverage) {
                  return null;
                }
                if (coverage.shadowed) {
                  return (
                    <div className="text-xs text-amber-600 dark:text-amber-500">
                      {i18next.t("application:Never used, an earlier row already carries this rule")}
                    </div>
                  );
                }
                return (
                  <div className="flex flex-wrap gap-1">
                    {coverage.methods.map((method) => (
                      <span key={method} className="rounded bg-muted px-1.5 py-0.5 text-xs text-muted-foreground">
                        {enumLabel(PROVIDER_CODE_RULES, method)}
                      </span>
                    ))}
                  </div>
                );
              },
            },
          ]}
        />
      </FormRow>
    </>
  );
}
