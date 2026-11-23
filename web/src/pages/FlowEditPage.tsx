// Copyright 2026 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import * as React from "react";
import i18next from "i18next";
import {useParams} from "react-router-dom";
import {FlaskConical, Sparkles} from "lucide-react";
import {Badge} from "@/components/ui/badge";
import {Button} from "@/components/ui/button";
import {Card, CardContent, CardHeader, CardTitle} from "@/components/ui/card";
import {Input} from "@/components/ui/input";
import {SearchableSelect} from "@/components/common/SearchableSelect";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useApplicationOptions, useGroupNameOptions, useOrganizationOptions, useRoleNameOptions, useUserNameOptions} from "@/hooks/use-options";
import {FlowStepEditor, FlowStepResultTag, FlowTriggers, getFlowTriggerLabel, newFlowStep} from "@/components/flow/FlowStepEditor";
import {RiskSignals} from "@/lib/risk-signals";
import * as FlowBackend from "@/backend/FlowBackend";
import * as Setting from "@/lib/setting";

interface FlowTemplate {
  key: string;
  nameKey: string;
  trigger: string;
  steps: () => any[];
}

const step = (type: "condition" | "action", trigger: string, patch: any) => ({...newFlowStep(type, trigger), ...patch});

/** Ready-made flows for the common cases, a starting point to adjust. */
const FlowTemplates: FlowTemplate[] = [
  {
    key: "admin-mfa-outside-office",
    nameKey: "flow:Ask admins for MFA outside the office network",
    trigger: "pre-signin",
    steps: () => [
      step("condition", "pre-signin", {field: "user.isAdmin", operator: "exists", value: ""}),
      step("condition", "pre-signin", {field: "request.ip", operator: "notInNetwork", value: "10.0.0.0/8, 192.168.0.0/16"}),
      step("action", "pre-signin", {action: "require-mfa"}),
    ],
  },
  {
    key: "block-contractors-off-hours",
    nameKey: "flow:Refuse contractors outside working hours",
    trigger: "pre-signin",
    steps: () => [
      step("condition", "pre-signin", {field: "user.groups", operator: "contains", value: "contractors"}),
      step("condition", "pre-signin", {field: "request.hour", operator: "in", value: "0, 1, 2, 3, 4, 5, 6, 7, 20, 21, 22, 23"}),
      step("action", "pre-signin", {action: "deny", params: {message: "Contractors can sign in from 8:00 to 20:00"}}),
    ],
  },
  {
    key: "alert-risky-signin",
    nameKey: "flow:Email the security team about risky sign-ins",
    trigger: "pre-signin",
    steps: () => [
      step("condition", "pre-signin", {field: "risk.score", operator: "greaterThan", value: "39"}),
      step("action", "pre-signin", {action: "send-email", params: {to: "security@example.com", subject: "Risky sign-in of {{user.name}}", content: "User: {{user.name}}\nIP: {{request.ip}} ({{risk.location}})\nScore: {{risk.score}}\nSignals: {{risk.signals}}"}}),
    ],
  },
  {
    key: "block-free-email-signup",
    nameKey: "flow:Only allow company emails to sign up",
    trigger: "pre-signup",
    steps: () => [
      step("condition", "pre-signup", {field: "user.emailDomain", operator: "notIn", value: "example.com"}),
      step("action", "pre-signup", {action: "deny", params: {message: "Please sign up with your company email"}}),
    ],
  },
  {
    key: "group-by-email-domain",
    nameKey: "flow:Put new users into a group by their email domain",
    trigger: "post-signup",
    steps: () => [
      step("condition", "post-signup", {field: "user.emailDomain", operator: "equals", value: "partner.com"}),
      step("action", "post-signup", {action: "add-to-group", params: {group: "partners"}}),
    ],
  },
  {
    key: "department-claim",
    nameKey: "flow:Add the department to the tokens",
    trigger: "pre-token",
    steps: () => [
      step("action", "pre-token", {action: "add-claim", params: {name: "department", value: "{{user.properties.department}}"}}),
    ],
  },
  {
    key: "external-check",
    nameKey: "flow:Ask your own service before each sign-in",
    trigger: "pre-signin",
    steps: () => [
      step("action", "pre-signin", {action: "http", params: {method: "POST", url: "https://api.example.com/casdoor/pre-signin", onError: "continue"}}),
    ],
  },
];

function TemplatePicker({onPick}: {onPick: (template: FlowTemplate) => void}) {
  return (
    <div className="mb-4 rounded-lg border border-dashed p-4">
      <div className="mb-3 flex items-center gap-2 text-sm font-medium">
        <Sparkles className="size-4 text-primary" />
        {i18next.t("flow:Start from a template")}
      </div>
      <div className="grid gap-2 sm:grid-cols-2">
        {FlowTemplates.map((template) => (
          <button
            key={template.key}
            type="button"
            className="rounded-md border px-3 py-2 text-left text-sm transition-colors hover:border-primary hover:bg-primary/5"
            onClick={() => onPick(template)}
          >
            <div>{i18next.t(template.nameKey)}</div>
            <div className="text-xs text-muted-foreground">{getFlowTriggerLabel(template.trigger)}</div>
          </button>
        ))}
      </div>
    </div>
  );
}

/** Runs the flow as it is in the editor on a chosen user, without changing anything. */
function TestPanel({record, onResults}: {record: any; onResults: (results: Record<string, any> | undefined) => void}) {
  const users = useUserNameOptions(record.owner);
  const applications = useApplicationOptions(record.owner);
  const [user, setUser] = React.useState("");
  const [application, setApplication] = React.useState("");
  const [ip, setIp] = React.useState("");
  const [riskScore, setRiskScore] = React.useState("");
  const [riskSignals, setRiskSignals] = React.useState<string[]>([]);
  const [running, setRunning] = React.useState(false);
  const [outcome, setOutcome] = React.useState<any>(null);
  const isSignin = record.trigger === "pre-signin" || record.trigger === "post-signin";

  const run = () => {
    if (!user) {
      Setting.showMessage("error", i18next.t("flow:Choose a user to test with"));
      return;
    }
    setRunning(true);
    FlowBackend.testFlow({flow: record, user, application, ip, riskScore: Setting.myParseInt(riskScore || "0"), riskSignals})
      .then((res: any) => {
        if (res.status === "ok") {
          setOutcome({run: res.data, result: res.data2});
          const results: Record<string, any> = {};
          (res.data?.steps ?? []).forEach((item: any) => {
            results[item.id] = item;
          });
          onResults(results);
        } else {
          Setting.showMessage("error", res.msg);
        }
      })
      .catch((error) => Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`))
      .finally(() => setRunning(false));
  };

  const result = outcome?.result;
  return (
    <Card className="mt-6">
      <CardHeader className="flex flex-row items-center gap-2 space-y-0">
        <FlaskConical className="size-4 text-primary" />
        <CardTitle className="text-base">{i18next.t("flow:Test")}</CardTitle>
        <span className="text-xs text-muted-foreground">{i18next.t("flow:The test runs the flow as edited, nothing is changed or sent")}</span>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
          <label className="space-y-1 text-sm">
            <span className="text-muted-foreground">{i18next.t("general:User")}</span>
            <SearchableSelect value={user} onChange={setUser} options={users} />
          </label>
          <label className="space-y-1 text-sm">
            <span className="text-muted-foreground">{i18next.t("general:Application")}</span>
            <SearchableSelect value={application} onChange={setApplication} options={applications} />
          </label>
          <label className="space-y-1 text-sm">
            <span className="text-muted-foreground">{i18next.t("general:Client IP")}</span>
            <Input placeholder="203.0.113.7" value={ip} onChange={(e) => setIp(e.target.value)} />
          </label>
          {isSignin ? (
            <label className="space-y-1 text-sm">
              <span className="text-muted-foreground">{i18next.t("enterprise:Risk score")}</span>
              <Input type="number" min={0} max={100} placeholder="0" value={riskScore} onChange={(e) => setRiskScore(e.target.value)} />
            </label>
          ) : null}
        </div>
        {isSignin ? (
          <div className="flex flex-wrap gap-1.5">
            {RiskSignals.map((signal) => {
              const active = riskSignals.includes(signal.name);
              return (
                <button
                  key={signal.name}
                  type="button"
                  className={`rounded-full border px-2.5 py-0.5 text-xs ${active ? "border-primary bg-primary/10 text-primary" : "text-muted-foreground"}`}
                  onClick={() => setRiskSignals(active ? riskSignals.filter((item) => item !== signal.name) : [...riskSignals, signal.name])}
                >
                  {i18next.t(signal.labelKey)}
                </button>
              );
            })}
          </div>
        ) : null}
        <div className="flex items-center gap-3">
          <Button type="button" onClick={run} loading={running}>
            {i18next.t("flow:Run test")}
          </Button>
          {outcome ? (
            <Button type="button" variant="ghost" size="sm" onClick={() => {
              setOutcome(null);
              onResults(undefined);
            }}>
              {i18next.t("general:Clear")}
            </Button>
          ) : null}
        </div>
        {outcome ? (
          <div className="rounded-md border bg-muted/30 p-3 text-sm">
            <div className="mb-2 flex flex-wrap items-center gap-2">
              <span className="font-medium">{i18next.t("payment:Result")}</span>
              <FlowRunStatusBadge status={outcome.run.status} />
              {result?.deny ? <Badge variant="destructive">{i18next.t("flow:Refused")}: {result.deny}</Badge> : null}
              {result?.requireMfa ? <Badge variant="warning">{i18next.t("flow:Ask for MFA")}</Badge> : null}
            </div>
            {result?.claims && Object.keys(result.claims).length > 0 ? (
              <pre className="overflow-x-auto rounded bg-background p-2 text-xs">{JSON.stringify(result.claims, null, 2)}</pre>
            ) : null}
            <p className="mt-1 text-xs text-muted-foreground">{i18next.t("flow:The result of each step is shown under it in the editor above")}</p>
          </div>
        ) : null}
      </CardContent>
    </Card>
  );
}

export function getFlowRunStatusLabel(status: string) {
  switch (status) {
  case "Completed":
    return i18next.t("enterprise:Completed");
  case "Stopped":
    return i18next.t("flow:Ended by a condition");
  case "Denied":
    return i18next.t("flow:Refused");
  case "Failed":
    return i18next.t("webhook:Failed");
  default:
    return status;
  }
}

function FlowRunStatusBadge({status}: {status: string}) {
  const variant = status === "Completed" ? "success" : status === "Denied" ? "warning" : status === "Failed" ? "destructive" : "secondary";
  return <Badge variant={variant}>{getFlowRunStatusLabel(status)}</Badge>;
}

/** The latest runs of the flow, each with how its steps went. */
function RunsPanel({record}: {record: any}) {
  const [runs, setRuns] = React.useState<any[]>([]);
  const [open, setOpen] = React.useState<string>("");
  const [version, setVersion] = React.useState(0);
  React.useEffect(() => {
    FlowBackend.getFlowRuns(record.owner, record.name, 1, 20).then((res: any) => {
      if (res.status === "ok") {
        setRuns(res.data ?? []);
      }
    });
  }, [record.owner, record.name, version]);

  return (
    <Card className="mt-6">
      <CardHeader className="flex flex-row items-center gap-2 space-y-0">
        <CardTitle className="text-base">{i18next.t("flow:Recent runs")}</CardTitle>
        <Button type="button" variant="ghost" size="sm" className="ml-auto" onClick={() => setVersion((v) => v + 1)}>
          {i18next.t("general:Refresh")}
        </Button>
      </CardHeader>
      <CardContent>
        {runs.length === 0 ? (
          <p className="text-sm text-muted-foreground">{i18next.t("flow:The flow hasn't run yet")}</p>
        ) : (
          <div className="divide-y rounded-md border">
            {runs.map((run) => (
              <div key={run.name} className="text-sm">
                <button type="button" className="flex w-full flex-wrap items-center gap-3 px-3 py-2 text-left hover:bg-muted/40" onClick={() => setOpen(open === run.name ? "" : run.name)}>
                  <span className="tabular-nums text-muted-foreground">{Setting.getFormattedDate(run.createdTime)}</span>
                  <span className="font-medium">{run.user}</span>
                  {run.application ? <span className="text-muted-foreground">{run.application}</span> : null}
                  <FlowRunStatusBadge status={run.status} />
                  <span className="ml-auto tabular-nums text-xs text-muted-foreground">{run.durationMs} ms</span>
                </button>
                {open === run.name ? (
                  <div className="space-y-1 bg-muted/20 px-3 pb-3">
                    {(run.steps ?? []).map((item: any, index: number) => (
                      <div key={`${item.id}-${index}`}>
                        <div className="pt-2 text-xs text-muted-foreground">{index + 1}. {item.summary}</div>
                        <FlowStepResultTag result={item} />
                      </div>
                    ))}
                    {run.error ? <p className="pt-2 text-xs text-destructive">{run.error}</p> : null}
                  </div>
                ) : null}
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

/** A flow: when it runs, and its conditions and actions in a visual chain, with a test panel. */
export default function FlowEditPage() {
  const {organizationName = "", flowName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();
  const applications = useApplicationOptions(organizationName);
  const groups = useGroupNameOptions(organizationName);
  const roles = useRoleNameOptions(organizationName);
  const [results, setResults] = React.useState<Record<string, any> | undefined>(undefined);

  const fields: EditField[] = [
    {type: "select", name: "owner", labelKey: "general:Organization", options: () => organizations, disabled: () => !Setting.isAdminUser(account)},
    {type: "text", name: "name", labelKey: "general:Name", required: true},
    {type: "text", name: "displayName", labelKey: "general:Display name"},
    {type: "textarea", name: "description", labelKey: "general:Description"},
    {
      type: "select",
      name: "trigger",
      labelKey: "flow:Trigger",
      required: true,
      options: () => FlowTriggers.map((item) => ({value: item, label: getFlowTriggerLabel(item)})),
    },
    {type: "multiselect", name: "applications", labelKey: "flow:Only for these applications", options: () => applications.map((item) => ({value: item.value, label: String(item.value)}))},
    {type: "number", name: "order", labelKey: "application:Order", min: 0},
    {type: "switch", name: "isEnabled", labelKey: "general:Enabled"},
    {
      type: "custom",
      name: "steps",
      labelKey: "flow:Steps",
      block: true,
      render: (ctx, update) => (
        <div>
          {(ctx.record.steps ?? []).length === 0 && ctx.mode !== "view" ? (
            <TemplatePicker
              onPick={(template) => {
                update("trigger", template.trigger);
                update("steps", template.steps());
                if (!ctx.record.displayName || ctx.record.displayName.startsWith("New Flow")) {
                  update("displayName", i18next.t(template.nameKey));
                }
              }}
            />
          ) : null}
          <FlowStepEditor
            trigger={ctx.record.trigger}
            steps={ctx.record.steps ?? []}
            onChange={(steps) => {
              setResults(undefined);
              update("steps", steps);
            }}
            options={{groups, roles}}
            disabled={ctx.mode === "view"}
            results={results}
          />
        </div>
      ),
    },
  ];

  return (
    <SimpleEditPage
      titleKey="flow:Edit Flow"
      backTo="/flows"
      deps={[organizationName, flowName]}
      fields={fields}
      fetch={() => FlowBackend.getFlow(organizationName, flowName)}
      add={(record) => FlowBackend.addFlow(record)}
      update={(record) => FlowBackend.updateFlow(organizationName, flowName, record)}
      editUrl={(record) => `/flows/${record.owner}/${record.name}`}
    >
      {(ctx) => (
        <>
          {ctx.mode === "view" ? null : <TestPanel record={ctx.record} onResults={setResults} />}
          {ctx.mode === "add" ? null : <RunsPanel record={ctx.record} />}
        </>
      )}
    </SimpleEditPage>
  );
}
