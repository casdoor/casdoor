import i18next from "i18next";
import {useParams} from "react-router-dom";
import {SearchableSelect} from "@/components/common/SearchableSelect";
import {RuleExpressionEditor} from "@/components/rule/RuleExpressionEditor";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useOrganizationOptions} from "@/hooks/use-options";
import * as RuleBackend from "@/backend/RuleBackend";
import * as Setting from "@/lib/setting";

export default function RuleEditPage() {
  const {organizationName = "", ruleName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();

  const notWaf = (ctx: {record: any}) => ctx.record.type !== "WAF";

  const fields: EditField[] = [
    {
      type: "select",
      name: "owner",
      labelKey: "general:Organization",
      options: () => organizations,
      disabled: () => !Setting.isAdminUser(account),
    },
    {type: "text", name: "name", labelKey: "general:Name", required: true},
    {
      type: "custom",
      name: "type",
      labelKey: "general:Type",
      render: (ctx, update) => (
        <SearchableSelect
          value={ctx.record.type ?? ""}
          onChange={(value) => {
            // each type has its own expression shape, so the old rows cannot carry over
            update("type", value);
            update("expressions", []);
          }}
          options={[
            {value: "WAF", label: "WAF"},
            {value: "IP", label: "IP"},
            {value: "User-Agent", label: "User-Agent"},
            {value: "IP Rate Limiting", label: i18next.t("rule:IP Rate Limiting")},
            {value: "Compound", label: i18next.t("rule:Compound")},
          ]}
        />
      ),
    },
    {
      type: "custom",
      name: "expressions",
      labelKey: "rule:Expressions",
      block: true,
      render: (ctx, update) => (
        <RuleExpressionEditor
          rule={ctx.record}
          owner={ctx.record.owner ?? organizationName}
          onChange={(rows) => update("expressions", rows)}
        />
      ),
    },
    {
      type: "select",
      name: "action",
      labelKey: "general:Action",
      when: notWaf,
      options: () => [
        {value: "Allow", label: i18next.t("permission:Allow")},
        {value: "Block", label: i18next.t("rule:Block")},
      ],
    },
    {
      type: "number",
      name: "statusCode",
      labelKey: "rule:Status code",
      when: (ctx) => notWaf(ctx) && ["Allow", "Block"].includes(ctx.record.action),
    },
    {type: "text", name: "reason", labelKey: "rule:Reason"},
    {type: "switch", name: "isVerbose", labelKey: "rule:Verbose mode"},
  ];

  return (
    <SimpleEditPage
      titleKey="rule:Edit Rule"
      backTo="/rules"
      deps={[organizationName, ruleName]}
      fields={fields}
      fetch={() => RuleBackend.getRule(organizationName, ruleName)}
      add={(record) => RuleBackend.addRule(record)}
      update={(record) => RuleBackend.updateRule(organizationName, ruleName, record)}
      editUrl={(record) => `/rules/${record.owner}/${record.name}`}
    />
  );
}
