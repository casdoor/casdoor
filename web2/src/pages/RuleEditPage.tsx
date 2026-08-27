import i18next from "i18next";
import {useParams} from "react-router-dom";
import {Input} from "@/components/ui/input";
import {SelectField} from "@/components/common/SelectField";
import {EditableTable} from "@/components/crud/EditableTable";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useOrganizationOptions} from "@/hooks/use-options";
import * as RuleBackend from "@/backend/RuleBackend";
import * as Setting from "@/lib/setting";

const TYPES = ["User-Agent", "IP", "IP Rate", "WAF", "URL"];
const ACTIONS = ["Block", "Allow", "Captcha", "Verify"];
const OPERATORS = ["Contains", "Equals", "Starts with", "Ends with", "Matches"];

export default function RuleEditPage() {
  const {organizationName = "", ruleName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();

  const fields: EditField[] = [
    {
      type: "select",
      name: "owner",
      labelKey: "general:Organization",
      options: () => organizations,
      disabled: () => !Setting.isAdminUser(account),
    },
    {type: "text", name: "name", labelKey: "general:Name"},
    {
      type: "select",
      name: "type",
      labelKey: "general:Type",
      options: () => TYPES.map((item) => ({value: item, label: item})),
    },
    {
      type: "custom",
      name: "expressions",
      labelKey: "rule:Expressions",
      block: true,
      render: (ctx, update) => (
        <EditableTable
          rows={ctx.record.expressions ?? []}
          onChange={(rows) => update("expressions", rows)}
          newRow={() => ({operator: "Contains", value: ""})}
          columns={[
            {
              key: "operator",
              title: i18next.t("rule:Operator"),
              width: 180,
              render: (row: any, _i, patch) => (
                <SelectField
                  value={row.operator}
                  onChange={(v) => patch({operator: v})}
                  options={OPERATORS.map((item) => ({id: item, name: item}))}
                />
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
      ),
    },
    {
      type: "select",
      name: "action",
      labelKey: "general:Action",
      options: () => ACTIONS.map((item) => ({value: item, label: item})),
    },
    {type: "number", name: "statusCode", labelKey: "rule:Status code"},
    {type: "text", name: "reason", labelKey: "rule:Reason"},
    {type: "switch", name: "isVerbose", labelKey: "rule:Is verbose"},
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
