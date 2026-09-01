import * as React from "react";
import i18next from "i18next";
import {Badge} from "@/components/ui/badge";
import {DescriptionList, type DescriptionItem} from "@/components/common/DescriptionList";
import {FormRow} from "@/components/crud/FormRow";

function escapeRegExp(value: string) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

function extractValue(message: string, key: string) {
  const escapedKey = escapeRegExp(key);
  const quotedMatch = message.match(new RegExp(`(?:^|\\s)${escapedKey}="([^"]*)"`, "i"));
  if (quotedMatch) {
    return quotedMatch[1];
  }

  const plainMatch = message.match(new RegExp(`(?:^|\\s)${escapedKey}=([^\\s]+)`, "i"));
  return plainMatch ? plainMatch[1] : "";
}

/** Pulls the audit fields out of an `[severity] type=AVC msg=audit(...)` line. */
export function parseSELinuxMessage(rawMessage: any) {
  const message = `${rawMessage ?? ""}`.trim();
  const severityMatch = message.match(/^\[([^\]]+)\]\s*/);
  const severity = severityMatch ? severityMatch[1] : "";
  const body = severityMatch ? message.slice(severityMatch[0].length) : message;

  return {
    severity,
    auditType: extractValue(body, "type"),
    auditStamp: (body.match(/msg=audit\(([^)]+)\)/) || [])[1] || "",
    decision: (body.match(/\bavc:\s+([a-z_]+)/i) || [])[1] || "",
    permission: (body.match(/\{\s*([^}]+?)\s*\}/) || [])[1] || "",
    pid: extractValue(body, "pid"),
    command: extractValue(body, "comm"),
    executable: extractValue(body, "exe"),
    path: extractValue(body, "path"),
    device: extractValue(body, "dev"),
    inode: extractValue(body, "ino"),
    sourceContext: extractValue(body, "scontext"),
    targetContext: extractValue(body, "tcontext"),
    targetClass: extractValue(body, "tclass"),
    permissive: extractValue(body, "permissive"),
    rawBody: body,
  };
}

function getSeverityVariant(severity: string) {
  switch ((severity || "").toLowerCase()) {
  case "warning":
    return "warning" as const;
  case "error":
    return "destructive" as const;
  case "info":
    return "default" as const;
  default:
    return "secondary" as const;
  }
}

/** Port of `web/src/SELinuxEntryViewer.js`. */
export function SELinuxEntryViewer({entry}: {entry: any}) {
  const details = parseSELinuxMessage(entry?.message);

  const value = (text: string, render?: (value: string) => React.ReactNode) => {
    if (!text) {
      return "-";
    }
    return render ? render(text) : text;
  };

  const items: DescriptionItem[] = [
    {
      label: i18next.t("general:Severity"),
      children: value(details.severity, (v) => <Badge variant={getSeverityVariant(v)}>{v}</Badge>),
    },
    {label: i18next.t("general:Type"), children: value(details.auditType)},
    {label: i18next.t("entry:Decision"), children: value(details.decision)},
    {label: i18next.t("entry:Permission"), children: value(details.permission)},
    {label: i18next.t("entry:Audit stamp"), children: value(details.auditStamp)},
    {label: i18next.t("entry:Permissive"), children: value(details.permissive)},
    {label: i18next.t("entry:Process ID"), children: value(details.pid)},
    {label: i18next.t("entry:Command"), children: value(details.command)},
    {label: i18next.t("entry:Executable"), children: value(details.executable)},
    {label: i18next.t("entry:Target class"), children: value(details.targetClass)},
    {label: i18next.t("general:Path"), children: value(details.path)},
    {label: i18next.t("entry:Device"), children: value(details.device)},
    {label: i18next.t("entry:Inode"), children: value(details.inode)},
    {label: i18next.t("entry:Source context"), span: 2, children: value(details.sourceContext)},
    {label: i18next.t("entry:Target context"), span: 2, children: value(details.targetContext)},
  ];

  return (
    <FormRow label={`${i18next.t("entry:SELinux event")}:`} block>
      <DescriptionList items={items} columns={2} />
    </FormRow>
  );
}

export default SELinuxEntryViewer;
