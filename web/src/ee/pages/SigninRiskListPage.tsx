// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

import * as React from "react";
import i18next from "i18next";
import {Badge} from "@/components/ui/badge";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {clientIpColumn, dateColumn, textColumn, valueFilters} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {EditionNotice} from "@/components/common/EditionNotice";
import {useRequestOrganization} from "@/hooks/use-organization";
import {RiskSignals} from "@/lib/risk-signals";
import * as AdaptiveBackend from "@/ee/backend/AdaptiveBackend";

export function getRiskSignalLabel(signal: string) {
  if (signal === "trusted-network") {
    return i18next.t("organization:Trusted network");
  }
  if (signal === "flow") {
    return i18next.t("flow:Flows");
  }
  const known = RiskSignals.find((item) => item.name === signal);
  return known ? i18next.t(known.labelKey) : signal;
}

export function getRiskDecisionLabel(decision: string) {
  switch (decision) {
  case "Deny":
    return i18next.t("enterprise:Blocked");
  case "MFA":
    return i18next.t("enterprise:MFA required");
  default:
    return i18next.t("enterprise:Allowed");
  }
}

export function RiskDecisionBadge({decision}: {decision: string}) {
  const variant = decision === "Deny" ? "destructive" : decision === "MFA" ? "warning" : "secondary";
  return <Badge variant={variant}>{getRiskDecisionLabel(decision)}</Badge>;
}

function SummaryTile({label, value, tone}: {label: string; value: React.ReactNode; tone?: string}) {
  return (
    <div className="rounded-lg border bg-card px-4 py-3">
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className={`mt-1 text-2xl font-semibold tabular-nums ${tone ?? ""}`}>{value}</div>
    </div>
  );
}

/** The counts of the last 7 days, and what triggered them most. */
function RiskSummary({owner}: {owner: string}) {
  const [summary, setSummary] = React.useState<any>(null);
  React.useEffect(() => {
    if (!owner) {
      return;
    }
    AdaptiveBackend.getSigninRiskSummary(owner, 7).then((res: any) => {
      if (res.status === "ok") {
        setSummary(res.data);
      }
    });
  }, [owner]);
  if (!summary) {
    return null;
  }

  const signals = Object.entries(summary.signals ?? {}).sort((a: any, b: any) => b[1] - a[1]).slice(0, 6);
  return (
    <div className="space-y-3">
      <div className="grid gap-3 sm:grid-cols-4">
        <SummaryTile label={i18next.t("enterprise:Risky sign-ins in 7 days")} value={summary.total} />
        <SummaryTile label={i18next.t("enterprise:MFA required")} value={summary.decisions?.MFA ?? 0} tone="text-warning" />
        <SummaryTile label={i18next.t("enterprise:Blocked")} value={summary.decisions?.Deny ?? 0} tone="text-destructive" />
        <SummaryTile label={i18next.t("enterprise:Allowed")} value={summary.decisions?.Allow ?? 0} />
      </div>
      {signals.length > 0 ? (
        <div className="flex flex-wrap gap-2 text-sm">
          {signals.map(([signal, count]: any) => (
            <span key={signal} className="rounded-md border px-2 py-0.5">
              {getRiskSignalLabel(signal)} <span className="tabular-nums text-muted-foreground">{count}</span>
            </span>
          ))}
        </div>
      ) : null}
    </div>
  );
}

/** The sign-ins that adaptive risk found a signal in, with the score and what it decided. */
export default function SigninRiskListPage() {
  const organizationName = useRequestOrganization();

  const columns: ColumnDef<any>[] = [
    dateColumn("createdTime", i18next.t("general:Time")),
    textColumn({dataIndex: "user", title: i18next.t("general:User"), searchable: true, width: 140}),
    clientIpColumn({dataIndex: "ip"}),
    textColumn({dataIndex: "location", title: i18next.t("user:Location"), width: 180}),
    {
      dataIndex: "score",
      title: i18next.t("enterprise:Risk score"),
      width: 100,
      sortable: true,
      render: (value: number) => (
        <span className={`tabular-nums font-medium ${value >= 70 ? "text-destructive" : value >= 40 ? "text-warning" : ""}`}>{value}</span>
      ),
    },
    {
      dataIndex: "signals",
      title: i18next.t("enterprise:Signals"),
      width: 260,
      render: (value: string[]) => (
        <div className="flex flex-wrap gap-1">
          {(value ?? []).map((signal) => (
            <Badge key={signal} variant="outline">{getRiskSignalLabel(signal)}</Badge>
          ))}
        </div>
      ),
    },
    {
      dataIndex: "decision",
      title: i18next.t("entry:Decision"),
      width: 120,
      filters: valueFilters(["Allow", "MFA", "Deny"]),
      render: (value: string) => <RiskDecisionBadge decision={value} />,
    },
    {
      dataIndex: "result",
      title: i18next.t("enterprise:Signed in"),
      width: 100,
      render: (value: string) => (value === "Succeeded" ? i18next.t("general:Yes") : "-"),
    },
    textColumn({dataIndex: "application", title: i18next.t("general:Application"), width: 140}),
    textColumn({dataIndex: "userAgent", title: i18next.t("general:User agent"), width: 220, className: "truncate max-w-[220px]"}),
  ];

  return (
    <CrudListPage
      title={i18next.t("enterprise:Sign-in risks")}
      description={i18next.t("enterprise:Sign-in risks - Tooltip")}
      header={
        <div>
          <EditionNotice />
          <RiskSummary owner={organizationName} />
        </div>
      }
      columns={columns}
      deps={[organizationName]}
      readOnly
      showActionColumn={false}
      fetch={(q) =>
        AdaptiveBackend.getSigninRiskEvents(organizationName, q.page, q.pageSize, q.searchedColumn, q.searchText, q.sortField, q.sortOrder)
      }
    />
  );
}
