import * as React from "react";
import i18next from "i18next";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {MultiSelect} from "@/components/common/MultiSelect";
import {SelectField} from "@/components/common/SelectField";
import {EditableTable} from "@/components/crud/EditableTable";
import * as RuleBackend from "@/backend/RuleBackend";
import * as Setting from "@/lib/setting";

export interface Expression {
  name: string;
  operator: string;
  value: string;
}

type Rows = Expression[];
type OnChange = (rows: Rows) => void;

const wafDefaults = (): Rows => [
  {
    name: i18next.t("rule:Enable XML request body parser"),
    operator: "match",
    value: "SecRule REQUEST_HEADERS:Content-Type \"^(?:application(?:/soap\\+|/)|text/)xml\" \"id:'200000',phase:1,t:none,t:lowercase,pass,nolog,ctl:requestBodyProcessor=XML\"",
  },
  {
    name: i18next.t("rule:Enable JSON request body parser"),
    operator: "match",
    value: "SecRule REQUEST_HEADERS:Content-Type \"^application/json\" \"id:'200001',phase:1,t:none,t:lowercase,pass,nolog,ctl:requestBodyProcessor=JSON\"",
  },
  {
    name: i18next.t("rule:Verify request body processing"),
    operator: "match",
    value: "SecRule &REQUEST_BODY \"@eq 0\" \"id:'200002',phase:2,t:none,deny,status:400,msg:'Failed to parse request body.'\"",
  },
];

const ipDefaults = (): Rows => [
  {name: i18next.t("rule:Loopback address"), operator: "is in", value: "127.0.0.1"},
  {name: i18next.t("rule:LAN CIDR"), operator: "is in", value: "10.0.0.0/8,192.168.0.0/16"},
];

const ipRateDefaults = (): Rows => [
  {name: i18next.t("rule:Default IP rate"), operator: "100", value: "6000"},
];

const compoundDefaults = (): Rows => [
  {name: i18next.t("rule:begin"), operator: "begin", value: "rule1"},
  {name: i18next.t("rule:and"), operator: "and", value: "rule2"},
];

const uaDefaults = (): Rows => [
  {name: i18next.t("rule:Current User-Agent"), operator: "equals", value: window.navigator.userAgent},
];

const UA_OPERATORS = ["equals", "does not equal", "contains", "does not contain", "match"];

/** the antd tables fill themselves with the defaults when the rule has no expressions yet */
function useDefaults(rows: Rows, onChange: OnChange, defaults: () => Rows) {
  const onChangeRef = React.useRef(onChange);
  onChangeRef.current = onChange;
  const defaultsRef = React.useRef(defaults);
  defaultsRef.current = defaults;

  React.useEffect(() => {
    if ((rows ?? []).length === 0) {
      onChangeRef.current(defaultsRef.current());
    }
    // only on mount, exactly like the antd constructors
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);
}

function TableTitle({title, onAdd, onRestore}: {title: React.ReactNode; onAdd?: () => void; onRestore: () => void}) {
  return (
    <div className="flex flex-wrap items-center gap-2">
      <span>{title}</span>
      {onAdd ? <Button size="sm" onClick={onAdd}>{i18next.t("general:Add")}</Button> : null}
      <Button variant="outline" size="sm" onClick={onRestore}>{i18next.t("general:Restore")}</Button>
    </div>
  );
}

function nameColumn() {
  return {
    key: "name",
    title: i18next.t("general:Name"),
    width: 180,
    render: (row: Expression, _index: number, patch: (p: any) => void) => (
      <Input value={row.name ?? ""} onChange={(e) => patch({name: e.target.value})} />
    ),
  };
}

function WafRuleTable({rows, onChange}: {rows: Rows; onChange: OnChange}) {
  useDefaults(rows, onChange, wafDefaults);

  return (
    <EditableTable<Expression>
      title={
        <TableTitle
          title="Seclang"
          onAdd={() => onChange([...(rows ?? []), {name: `${i18next.t("rule:New WAF rule")} - ${(rows ?? []).length}`, operator: "match", value: ""}])}
          onRestore={() => onChange(wafDefaults())}
        />
      }
      rows={rows}
      onChange={onChange}
      columns={[
        nameColumn(),
        {
          key: "value",
          title: i18next.t("rule:Expression"),
          render: (row, _index, patch) => (
            <Input value={row.value ?? ""} onChange={(e) => patch({value: e.target.value})} />
          ),
        },
      ]}
    />
  );
}

function IpRuleTable({rows, onChange}: {rows: Rows; onChange: OnChange}) {
  useDefaults(rows, onChange, ipDefaults);

  return (
    <EditableTable<Expression>
      title={
        <TableTitle
          title={i18next.t("rule:IP addresses")}
          onAdd={() => onChange([...(rows ?? []), {name: `${i18next.t("rule:New IP rule")} - ${(rows ?? []).length}`, operator: "is in", value: "127.0.0.1"}])}
          onRestore={() => onChange(ipDefaults())}
        />
      }
      rows={rows}
      onChange={onChange}
      columns={[
        nameColumn(),
        {
          key: "operator",
          title: i18next.t("rule:Operator"),
          width: 180,
          render: (row, _index, patch) => (
            <SelectField
              value={row.operator}
              onChange={(value) => patch({operator: value})}
              options={[
                {id: "is in", name: i18next.t("rule:is in")},
                {id: "is not in", name: i18next.t("rule:is not in")},
              ]}
            />
          ),
        },
        {
          key: "value",
          title: i18next.t("rule:IP List"),
          render: (row, _index, patch) => {
            const values = row.value ? row.value.split(",") : [];
            return (
              <MultiSelect
                creatable
                placeholder={i18next.t("rule:Input IP Addresses")}
                value={values}
                onChange={(next) => patch({value: next.map((item) => item.trim()).join(",")})}
                options={values.map((item) => ({value: item, label: item}))}
              />
            );
          },
        },
      ]}
    />
  );
}

function UaRuleTable({rows, onChange}: {rows: Rows; onChange: OnChange}) {
  useDefaults(rows, onChange, uaDefaults);

  return (
    <EditableTable<Expression>
      title={
        <TableTitle
          title="User-Agents"
          onAdd={() => onChange([...(rows ?? []), {name: `${i18next.t("rule:New User-Agent rule")} - ${(rows ?? []).length}`, operator: "equals", value: ""}])}
          onRestore={() => onChange(uaDefaults())}
        />
      }
      rows={rows}
      onChange={onChange}
      columns={[
        nameColumn(),
        {
          key: "operator",
          title: i18next.t("rule:Operator"),
          width: 180,
          render: (row, _index, patch) => (
            <SelectField
              value={row.operator}
              onChange={(value) => patch({operator: value})}
              options={UA_OPERATORS.map((item) => ({
                id: item,
                name: item === "match" ? i18next.t("rule:regex match") : i18next.t(`rule:${item}`),
              }))}
            />
          ),
        },
        {
          key: "value",
          title: i18next.t("webhook:Value"),
          render: (row, _index, patch) => (
            <Input
              value={row.value ?? ""}
              onChange={(e) => patch({value: e.target.value})}
              onBlur={(e) => patch({value: e.target.value.replace(/\s+/g, " ").trim()})}
            />
          ),
        },
      ]}
    />
  );
}

function IpRateRuleTable({rows, onChange}: {rows: Rows; onChange: OnChange}) {
  useDefaults(rows, onChange, ipRateDefaults);

  return (
    <EditableTable<Expression>
      title={<TableTitle title={i18next.t("rule:IP Rate Limiting")} onRestore={() => onChange(ipRateDefaults())} />}
      rows={rows}
      onChange={onChange}
      reorderable={false}
      disabled
      columns={[
        nameColumn(),
        {
          key: "operator",
          title: i18next.t("rule:Rate"),
          width: "40%",
          render: (row, _index, patch) => (
            <div className="flex items-center gap-2">
              <Input
                type="number"
                value={row.operator ?? ""}
                onChange={(e) => patch({operator: String(e.target.value)})}
              />
              <span className="whitespace-nowrap text-xs text-muted-foreground">{i18next.t("rule:Requests per IP per second")}</span>
            </div>
          ),
        },
        {
          key: "value",
          title: i18next.t("rule:Block Duration"),
          render: (row, _index, patch) => (
            <div className="flex items-center gap-2">
              <Input
                type="number"
                value={row.value ?? ""}
                onChange={(e) => patch({value: String(e.target.value)})}
              />
              <span className="whitespace-nowrap text-xs text-muted-foreground">{i18next.t("usage:seconds")}</span>
            </div>
          ),
        },
      ]}
    />
  );
}

function CompoundRuleTable({rows, onChange, owner, ruleName}: {rows: Rows; onChange: OnChange; owner: string; ruleName: string}) {
  const [rules, setRules] = React.useState<string[]>([]);

  useDefaults(rows, onChange, compoundDefaults);

  React.useEffect(() => {
    RuleBackend.getRules(owner).then((res: any) => {
      // a compound rule cannot reference itself
      setRules((res.data ?? [])
        .map((item: any) => Setting.getItemId(item))
        .filter((id: string) => id !== `${owner}/${ruleName}`));
    });
  }, [owner, ruleName]);

  return (
    <EditableTable<Expression>
      title={
        <TableTitle
          title={i18next.t("rule:Compound")}
          onAdd={() => onChange([...(rows ?? []), {name: `${i18next.t("rule:New compound item")} - ${(rows ?? []).length}`, operator: "and", value: ""}])}
          onRestore={() => onChange(compoundDefaults())}
        />
      }
      rows={rows}
      onChange={onChange}
      columns={[
        {
          key: "operator",
          title: i18next.t("rule:Logic"),
          width: 180,
          render: (row, index, patch) => (
            <SelectField
              value={row.operator}
              onChange={(value) => patch({operator: value})}
              options={index === 0
                ? [{id: "begin", name: i18next.t("rule:begin")}]
                : [{id: "and", name: i18next.t("rule:and")}, {id: "or", name: i18next.t("rule:or")}]}
            />
          ),
        },
        {
          key: "value",
          title: i18next.t("application:Rule"),
          render: (row, _index, patch) => (
            <SelectField
              value={row.value}
              onChange={(value) => patch({value})}
              options={rules.map((item) => ({id: item, name: item}))}
            />
          ),
        },
      ]}
    />
  );
}

/** Picks the expression table the rule's type calls for, as web/src/RuleEditPage.js does. */
export function RuleExpressionEditor({rule, owner, onChange}: {rule: any; owner: string; onChange: OnChange}) {
  const rows = rule.expressions ?? [];

  switch (rule.type) {
  case "WAF":
    return <WafRuleTable key="WAF" rows={rows} onChange={onChange} />;
  case "IP":
    return <IpRuleTable key="IP" rows={rows} onChange={onChange} />;
  case "User-Agent":
    return <UaRuleTable key="User-Agent" rows={rows} onChange={onChange} />;
  case "IP Rate Limiting":
    return <IpRateRuleTable key="IP Rate Limiting" rows={rows} onChange={onChange} />;
  case "Compound":
    return <CompoundRuleTable key="Compound" rows={rows} onChange={onChange} owner={owner} ruleName={rule.name} />;
  default:
    return null;
  }
}
