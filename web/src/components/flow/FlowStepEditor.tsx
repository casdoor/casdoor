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
import {ArrowDown, ArrowUp, Ban, Filter, Plus, Trash2, Zap} from "lucide-react";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Switch} from "@/components/ui/switch";
import {Textarea} from "@/components/ui/textarea";
import {SelectField} from "@/components/common/SelectField";
import {SearchableSelect, type SearchableOption} from "@/components/common/SearchableSelect";
import {cn} from "@/lib/utils";

export const FlowTriggers = ["pre-signup", "post-signup", "pre-signin", "post-signin", "pre-token"];

export function getFlowTriggerLabel(trigger: string) {
  switch (trigger) {
  case "pre-signup":
    return i18next.t("flow:Before sign-up");
  case "post-signup":
    return i18next.t("flow:After sign-up");
  case "pre-signin":
    return i18next.t("flow:Before sign-in");
  case "post-signin":
    return i18next.t("flow:After sign-in");
  case "pre-token":
    return i18next.t("flow:Before a token is issued");
  default:
    return trigger;
  }
}

export function getFlowTriggerDescription(trigger: string) {
  switch (trigger) {
  case "pre-signup":
    return i18next.t("flow:The user is not saved yet, a flow can refuse the sign-up or fill in fields");
  case "post-signup":
    return i18next.t("flow:The user has been added, a flow runs in the background");
  case "pre-signin":
    return i18next.t("flow:The credentials are verified, a flow can refuse the sign-in or ask for MFA, and sees the risk of adaptive risk");
  case "post-signin":
    return i18next.t("flow:The user has signed in, a flow runs in the background");
  case "pre-token":
    return i18next.t("flow:A flow can add claims to the access and ID tokens, or refuse the token");
  default:
    return "";
  }
}

const TriggerActions: Record<string, string[]> = {
  "pre-signup": ["deny", "set-user-field", "add-to-group", "http", "stop"],
  "post-signup": ["set-user-field", "add-to-group", "remove-from-group", "add-to-role", "remove-from-role", "http", "send-email", "stop"],
  "pre-signin": ["deny", "require-mfa", "set-user-field", "add-to-group", "remove-from-group", "add-to-role", "remove-from-role", "http", "send-email", "stop"],
  "post-signin": ["set-user-field", "add-to-group", "remove-from-group", "add-to-role", "remove-from-role", "http", "send-email", "stop"],
  "pre-token": ["deny", "add-claim", "http", "stop"],
};

export function getFlowActionLabel(action: string) {
  switch (action) {
  case "deny":
    return i18next.t("flow:Refuse");
  case "require-mfa":
    return i18next.t("flow:Ask for MFA");
  case "add-claim":
    return i18next.t("flow:Add a token claim");
  case "set-user-field":
    return i18next.t("flow:Set a user field");
  case "add-to-group":
    return i18next.t("flow:Add to group");
  case "remove-from-group":
    return i18next.t("flow:Remove from group");
  case "add-to-role":
    return i18next.t("flow:Grant role");
  case "remove-from-role":
    return i18next.t("flow:Revoke role");
  case "http":
    return i18next.t("flow:Call an HTTP endpoint");
  case "send-email":
    return i18next.t("flow:Send an email");
  case "stop":
    return i18next.t("flow:End the flow");
  default:
    return action;
  }
}

type FieldKind = "text" | "number" | "bool" | "list" | "ip";

interface FlowField {
  value: string;
  labelKey: string;
  kind: FieldKind;
  triggers?: string[];
}

const SigninTriggers = ["pre-signin", "post-signin"];

export const FlowFields: FlowField[] = [
  {value: "user.name", labelKey: "general:Name", kind: "text"},
  {value: "user.displayName", labelKey: "general:Display name", kind: "text"},
  {value: "user.email", labelKey: "general:Email", kind: "text"},
  {value: "user.emailDomain", labelKey: "flow:Email domain", kind: "text"},
  {value: "user.emailVerified", labelKey: "user:Email verified", kind: "bool"},
  {value: "user.phone", labelKey: "general:Phone", kind: "text"},
  {value: "user.countryCode", labelKey: "user:Country code", kind: "text"},
  {value: "user.type", labelKey: "general:Type", kind: "text"},
  {value: "user.tag", labelKey: "general:Tag", kind: "text"},
  {value: "user.region", labelKey: "provider:Region", kind: "text"},
  {value: "user.affiliation", labelKey: "user:Affiliation", kind: "text"},
  {value: "user.title", labelKey: "general:Title", kind: "text"},
  {value: "user.groups", labelKey: "general:Groups", kind: "list"},
  {value: "user.roles", labelKey: "general:Roles", kind: "list"},
  {value: "user.isAdmin", labelKey: "user:Is admin", kind: "bool"},
  {value: "user.isMfaEnabled", labelKey: "flow:MFA is set up", kind: "bool"},
  {value: "user.signupApplication", labelKey: "general:Signup application", kind: "text"},
  {value: "user.accountAgeDays", labelKey: "flow:Account age (days)", kind: "number"},
  {value: "user.properties.", labelKey: "user:Properties", kind: "text"},
  {value: "application.name", labelKey: "general:Application", kind: "text"},
  {value: "request.ip", labelKey: "general:Client IP", kind: "ip"},
  {value: "request.country", labelKey: "flow:Country", kind: "text"},
  {value: "request.userAgent", labelKey: "general:User agent", kind: "text"},
  {value: "request.hour", labelKey: "flow:Hour of the day", kind: "number"},
  {value: "request.weekday", labelKey: "flow:Day of the week", kind: "text"},
  {value: "risk.score", labelKey: "enterprise:Risk score", kind: "number", triggers: SigninTriggers},
  {value: "risk.signals", labelKey: "enterprise:Signals", kind: "list", triggers: SigninTriggers},
  {value: "risk.decision", labelKey: "entry:Decision", kind: "text", triggers: SigninTriggers},
  {value: "claims.", labelKey: "flow:Token claim", kind: "text", triggers: ["pre-token"]},
];

const OperatorsByKind: Record<FieldKind, string[]> = {
  text: ["equals", "notEquals", "contains", "notContains", "startsWith", "endsWith", "in", "notIn", "matches", "exists", "notExists"],
  number: ["equals", "notEquals", "greaterThan", "lessThan", "in"],
  bool: ["exists", "notExists"],
  list: ["contains", "notContains", "in", "notIn", "exists", "notExists"],
  ip: ["inNetwork", "notInNetwork", "equals", "in", "startsWith"],
};

export function getFlowOperatorLabel(operator: string) {
  switch (operator) {
  case "equals":
    return i18next.t("flow:is");
  case "notEquals":
    return i18next.t("flow:is not");
  case "contains":
    return i18next.t("rule:contains");
  case "notContains":
    return i18next.t("flow:doesn't contain");
  case "startsWith":
    return i18next.t("flow:starts with");
  case "endsWith":
    return i18next.t("flow:ends with");
  case "in":
    return i18next.t("flow:is one of");
  case "notIn":
    return i18next.t("flow:is none of");
  case "matches":
    return i18next.t("flow:matches the regex");
  case "greaterThan":
    return i18next.t("flow:is greater than");
  case "lessThan":
    return i18next.t("flow:is less than");
  case "exists":
    return i18next.t("flow:is set (or true)");
  case "notExists":
    return i18next.t("flow:is empty (or false)");
  case "inNetwork":
    return i18next.t("flow:is in the networks");
  case "notInNetwork":
    return i18next.t("flow:is not in the networks");
  default:
    return operator;
  }
}

/** The field of the catalog that a condition's field belongs to, "user.properties.dept" belongs to "user.properties.". */
function findFlowField(field: string): FlowField | undefined {
  return FlowFields.find((item) => item.value === field) ?? FlowFields.find((item) => item.value.endsWith(".") && field.startsWith(item.value));
}

const WritableUserFields = ["displayName", "firstName", "lastName", "avatar", "tag", "region", "location", "affiliation", "title", "homepage", "bio", "language", "gender", "education"];

let stepCounter = 0;
export function newFlowStep(type: "condition" | "action", trigger: string): any {
  stepCounter++;
  const id = `${Date.now().toString(36)}${stepCounter}`;
  if (type === "condition") {
    return {id, type, field: "user.groups", operator: "contains", value: "", params: {}};
  }
  const action = (TriggerActions[trigger] ?? ["stop"])[0];
  return {id, type, action, params: {}};
}

export interface FlowOptions {
  groups: SearchableOption[];
  roles: SearchableOption[];
}

function ParamRow({label, children, hint}: {label: string; children: React.ReactNode; hint?: string}) {
  return (
    <label className="grid gap-1 text-sm sm:grid-cols-[140px_1fr] sm:items-start sm:gap-3">
      <span className="pt-2 text-muted-foreground">{label}</span>
      <span className="min-w-0">
        {children}
        {hint ? <span className="mt-1 block text-xs text-muted-foreground">{hint}</span> : null}
      </span>
    </label>
  );
}

const templateHint = () => `${i18next.t("flow:Fields can be put in the text, such as")} {{user.email}}, {{user.properties.department}}, {{request.ip}}`;

function ActionParams({step, onChange, options, disabled}: {step: any; onChange: (params: any) => void; options: FlowOptions; disabled?: boolean}) {
  const params = step.params ?? {};
  const set = (key: string, value: string) => onChange({...params, [key]: value});
  const text = (key: string, label: string, placeholder = "", hint?: string) => (
    <ParamRow label={label} hint={hint}>
      <Input disabled={disabled} placeholder={placeholder} value={params[key] ?? ""} onChange={(e) => set(key, e.target.value)} />
    </ParamRow>
  );

  switch (step.action) {
  case "deny":
    return text("message", i18next.t("payment:Message"), i18next.t("flow:Sign-in from outside the office is not allowed"), templateHint());
  case "add-claim":
    return (
      <>
        {text("name", i18next.t("flow:Claim name"), "department")}
        {text("value", i18next.t("webhook:Value"), "{{user.properties.department}}", templateHint())}
        <ParamRow label={i18next.t("flow:Value type")}>
          <SelectField
            disabled={disabled}
            value={params.valueType || "text"}
            onChange={(v) => set("valueType", v === "text" ? "" : v)}
            options={[{id: "text", name: i18next.t("entry:Text")}, {id: "json", name: "JSON"}]}
          />
        </ParamRow>
      </>
    );
  case "set-user-field": {
    const field = params.field ?? "";
    const isProperty = field.startsWith("properties.");
    return (
      <>
        <ParamRow label={i18next.t("flow:Field")}>
          <div className="flex gap-2">
            <SelectField
              disabled={disabled}
              className="flex-1"
              value={isProperty ? "properties." : field}
              onChange={(v) => set("field", v)}
              options={[...WritableUserFields.map((item) => ({id: item, name: item})), {id: "properties.", name: i18next.t("user:Properties")}]}
            />
            {isProperty ? (
              <Input
                disabled={disabled}
                className="flex-1"
                placeholder="department"
                value={field.slice("properties.".length)}
                onChange={(e) => set("field", `properties.${e.target.value}`)}
              />
            ) : null}
          </div>
        </ParamRow>
        {text("value", i18next.t("webhook:Value"), "", templateHint())}
      </>
    );
  }
  case "add-to-group":
  case "remove-from-group":
    return (
      <ParamRow label={i18next.t("subscription:Group")}>
        <SearchableSelect disabled={disabled} value={params.group ?? ""} onChange={(v) => set("group", v)} options={options.groups} />
      </ParamRow>
    );
  case "add-to-role":
  case "remove-from-role":
    return (
      <ParamRow label={i18next.t("general:Role")}>
        <SearchableSelect disabled={disabled} value={params.role ?? ""} onChange={(v) => set("role", v)} options={options.roles} />
      </ParamRow>
    );
  case "http":
    return (
      <>
        <ParamRow label={i18next.t("general:URL")}>
          <div className="flex gap-2">
            <SelectField
              disabled={disabled}
              className="w-28"
              value={params.method || "POST"}
              onChange={(v) => set("method", v)}
              options={["POST", "PUT", "GET"].map((item) => ({id: item, name: item}))}
            />
            <Input disabled={disabled} className="flex-1" placeholder="https://api.example.com/casdoor-hook" value={params.url ?? ""} onChange={(e) => set("url", e.target.value)} />
          </div>
        </ParamRow>
        <ParamRow label={i18next.t("flow:Signing secret")} hint={i18next.t("flow:The body is signed with HMAC-SHA256 in the header X-Casdoor-Signature")}>
          <Input disabled={disabled} type="password" value={params.secret ?? ""} onChange={(e) => set("secret", e.target.value)} />
        </ParamRow>
        <ParamRow label={i18next.t("webhook:Headers")} hint={i18next.t("flow:One header per line, as Name: value")}>
          <Textarea disabled={disabled} rows={2} placeholder="Authorization: Bearer xxx" value={params.headers ?? ""} onChange={(e) => set("headers", e.target.value)} />
        </ParamRow>
        <ParamRow label={i18next.t("flow:Timeout (seconds)")}>
          <Input disabled={disabled} type="number" min={1} max={30} placeholder="5" value={params.timeout ?? ""} onChange={(e) => set("timeout", e.target.value)} />
        </ParamRow>
        <ParamRow label={i18next.t("flow:When the call fails")}>
          <SelectField
            disabled={disabled}
            value={params.onError || "fail"}
            onChange={(v) => set("onError", v === "fail" ? "" : v)}
            options={[
              {id: "fail", name: i18next.t("flow:End the flow with an error")},
              {id: "continue", name: i18next.t("flow:Ignore and continue")},
              {id: "deny", name: i18next.t("flow:Refuse")},
            ]}
          />
        </ParamRow>
        <ParamRow label={i18next.t("flow:Call it in tests")}>
          <Switch disabled={disabled} checked={params.callInTest === "true"} onCheckedChange={(v) => set("callInTest", v ? "true" : "")} />
        </ParamRow>
        <p className="text-xs text-muted-foreground">
          {i18next.t("flow:The endpoint gets the user, the request and the risk as JSON, and can answer {\"deny\": \"...\", \"requireMfa\": true, \"claims\": {...}, \"user\": {...}, \"stop\": true}")}
        </p>
      </>
    );
  case "send-email":
    return (
      <>
        {text("to", i18next.t("flow:To"), i18next.t("flow:The user's email by default"), templateHint())}
        {text("subject", i18next.t("provider:Subject"), "{{user.name}}")}
        <ParamRow label={i18next.t("provider:Content")} hint={templateHint()}>
          <Textarea disabled={disabled} rows={3} value={params.content ?? ""} onChange={(e) => set("content", e.target.value)} />
        </ParamRow>
      </>
    );
  default:
    return null;
  }
}

function ConditionEditor({step, trigger, onChange, disabled}: {step: any; trigger: string; onChange: (step: any) => void; disabled?: boolean}) {
  const fields = FlowFields.filter((item) => !item.triggers || item.triggers.includes(trigger));
  const field = findFlowField(step.field);
  const kind = field?.kind ?? "text";
  const operators = OperatorsByKind[kind];
  const isPrefix = field?.value.endsWith(".");
  const needsValue = step.operator !== "exists" && step.operator !== "notExists";

  return (
    <div className="flex flex-wrap items-center gap-2">
      <SelectField
        disabled={disabled}
        className="w-64"
        value={field?.value ?? step.field}
        onChange={(v) => {
          const next = findFlowField(v);
          const nextOperators = OperatorsByKind[next?.kind ?? "text"];
          onChange({...step, field: v, operator: nextOperators.includes(step.operator) ? step.operator : nextOperators[0]});
        }}
        options={fields.map((item) => ({id: item.value, name: `${i18next.t(item.labelKey)} (${item.value.replace(/\.$/, ".*")})`}))}
      />
      {isPrefix ? (
        <Input
          disabled={disabled}
          className="w-36"
          placeholder={field?.value === "claims." ? "scope" : "department"}
          value={step.field.slice(field!.value.length)}
          onChange={(e) => onChange({...step, field: `${field!.value}${e.target.value}`})}
        />
      ) : null}
      <SelectField
        disabled={disabled}
        className="w-44"
        value={step.operator}
        onChange={(v) => onChange({...step, operator: v})}
        options={operators.map((item) => ({id: item, name: getFlowOperatorLabel(item)}))}
      />
      {needsValue ? (
        <Input
          disabled={disabled}
          className="min-w-[180px] flex-1"
          placeholder={
            kind === "ip" ? "10.0.0.0/8, 192.168.1.0/24" : step.operator === "in" || step.operator === "notIn" ? "a, b, c" : kind === "number" ? "50" : ""
          }
          value={step.value ?? ""}
          onChange={(e) => onChange({...step, value: e.target.value})}
        />
      ) : null}
    </div>
  );
}

function StepCard({
  step,
  index,
  count,
  trigger,
  options,
  disabled,
  onChange,
  onMove,
  onRemove,
}: {
  step: any;
  index: number;
  count: number;
  trigger: string;
  options: FlowOptions;
  disabled?: boolean;
  onChange: (step: any) => void;
  onMove: (delta: number) => void;
  onRemove: () => void;
}) {
  const isCondition = step.type === "condition";
  const actions = TriggerActions[trigger] ?? [];
  const isAllowed = isCondition || actions.includes(step.action);
  const Icon = isCondition ? Filter : step.action === "deny" ? Ban : Zap;

  return (
    <div className={cn("relative rounded-lg border bg-card", step.disabled && "opacity-60", !isAllowed && "border-destructive")}>
      <div className="flex items-center gap-2 border-b px-3 py-2">
        <span
          className={cn(
            "flex size-6 shrink-0 items-center justify-center rounded-full text-xs",
            isCondition ? "bg-blue-500/10 text-blue-600 dark:text-blue-400" : "bg-primary/10 text-primary",
          )}
        >
          <Icon className="size-3.5" />
        </span>
        <span className="text-xs font-medium text-muted-foreground">
          {index + 1}. {isCondition ? i18next.t("flow:Only if") : i18next.t("flow:Then")}
        </span>
        {!isCondition ? (
          <SelectField
            disabled={disabled}
            className="h-8 w-56"
            value={step.action}
            onChange={(v) => onChange({...step, action: v, params: {}})}
            options={actions.map((item) => ({id: item, name: getFlowActionLabel(item)}))}
          />
        ) : null}
        <div className="ml-auto flex items-center gap-1">
          <span className="mr-1 flex items-center gap-1.5 text-xs text-muted-foreground">
            <Switch disabled={disabled} checked={!step.disabled} onCheckedChange={(v) => onChange({...step, disabled: !v})} />
          </span>
          <Button type="button" variant="ghost" size="iconSm" disabled={disabled || index === 0} onClick={() => onMove(-1)} aria-label="Up">
            <ArrowUp />
          </Button>
          <Button type="button" variant="ghost" size="iconSm" disabled={disabled || index === count - 1} onClick={() => onMove(1)} aria-label="Down">
            <ArrowDown />
          </Button>
          <Button type="button" variant="ghost" size="iconSm" disabled={disabled} onClick={onRemove} aria-label="Delete">
            <Trash2 />
          </Button>
        </div>
      </div>
      <div className="space-y-2 px-3 py-3">
        {isCondition ? (
          <>
            <ConditionEditor step={step} trigger={trigger} onChange={onChange} disabled={disabled} />
            <p className="text-xs text-muted-foreground">{i18next.t("flow:Otherwise the flow ends here")}</p>
          </>
        ) : (
          <ActionParams step={step} options={options} disabled={disabled} onChange={(params) => onChange({...step, params})} />
        )}
        {!isAllowed ? <p className="text-xs text-destructive">{i18next.t("flow:This action can't run at the chosen trigger")}</p> : null}
      </div>
    </div>
  );
}

function AddStepButtons({onAdd, disabled}: {onAdd: (type: "condition" | "action") => void; disabled?: boolean}) {
  return (
    <div className="flex justify-center gap-2">
      <Button type="button" variant="outline" size="sm" disabled={disabled} onClick={() => onAdd("condition")}>
        <Plus />
        {i18next.t("flow:Condition")}
      </Button>
      <Button type="button" variant="outline" size="sm" disabled={disabled} onClick={() => onAdd("action")}>
        <Plus />
        {i18next.t("general:Action")}
      </Button>
    </div>
  );
}

/** The steps of a flow as a vertical chain under its trigger: conditions gate the steps after them. */
export function FlowStepEditor({
  trigger,
  steps,
  onChange,
  options,
  disabled,
  results,
}: {
  trigger: string;
  steps: any[];
  onChange: (steps: any[]) => void;
  options: FlowOptions;
  disabled?: boolean;
  results?: Record<string, any>;
}) {
  const list = steps ?? [];
  const update = (index: number, step: any) => onChange(list.map((item, i) => (i === index ? step : item)));
  const move = (index: number, delta: number) => {
    const next = [...list];
    const [item] = next.splice(index, 1);
    next.splice(index + delta, 0, item);
    onChange(next);
  };
  const insert = (index: number, type: "condition" | "action") => {
    const next = [...list];
    next.splice(index, 0, newFlowStep(type, trigger));
    onChange(next);
  };

  const connector = <div className="mx-auto h-4 w-px bg-border" />;

  return (
    <div className="flow-step-editor space-y-0">
      <div className="rounded-lg border border-primary/40 bg-primary/5 px-4 py-3">
        <div className="text-xs font-medium uppercase tracking-wide text-primary">{i18next.t("flow:When")}</div>
        <div className="font-medium">{getFlowTriggerLabel(trigger)}</div>
        <div className="text-xs text-muted-foreground">{getFlowTriggerDescription(trigger)}</div>
      </div>
      {list.map((step, index) => {
        const result = results?.[step.id];
        return (
          <React.Fragment key={step.id ?? index}>
            {connector}
            <div className="relative">
              <StepCard
                step={step}
                index={index}
                count={list.length}
                trigger={trigger}
                options={options}
                disabled={disabled}
                onChange={(next) => update(index, next)}
                onMove={(delta) => move(index, delta)}
                onRemove={() => onChange(list.filter((_, i) => i !== index))}
              />
              {result ? <FlowStepResultTag result={result} /> : null}
            </div>
          </React.Fragment>
        );
      })}
      {connector}
      <AddStepButtons disabled={disabled} onAdd={(type) => insert(list.length, type)} />
    </div>
  );
}

export function getFlowStepResultLabel(result: string) {
  switch (result) {
  case "passed":
    return i18next.t("flow:Passed");
  case "failed":
    return i18next.t("flow:Not met");
  case "done":
    return i18next.t("edition:Done");
  case "error":
    return i18next.t("general:Error");
  case "skipped":
    return i18next.t("enterprise:Skipped");
  default:
    return result;
  }
}

export function FlowStepResultTag({result}: {result: any}) {
  const tone =
    result.result === "passed" || result.result === "done"
      ? "border-success/40 bg-success/10 text-success"
      : result.result === "error" || result.result === "failed"
        ? "border-destructive/40 bg-destructive/10 text-destructive"
        : "border-border bg-muted text-muted-foreground";
  return (
    <div className={cn("mt-1 rounded-md border px-3 py-1.5 text-xs", tone)}>
      <span className="font-medium">{getFlowStepResultLabel(result.result)}</span>
      {result.message ? <span className="ml-2 break-all">{result.message}</span> : null}
    </div>
  );
}
