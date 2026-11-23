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

import i18next from "i18next";
import {Badge} from "@/components/ui/badge";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {clientIpColumn, dateColumn, textColumn, valueFilters} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as DeviceBackend from "@/backend/DeviceBackend";
import * as Setting from "@/lib/setting";

export function getDeviceStatusLabel(status: string) {
  switch (status) {
  case "Trusted":
    return i18next.t("device:Trusted");
  case "Blocked":
    return i18next.t("enterprise:Blocked");
  default:
    return i18next.t("device:Waiting for approval");
  }
}

export function DeviceStatusBadge({device}: {device: any}) {
  const variant = device.status === "Trusted" ? (device.isCompliant ? "success" : "warning") : device.status === "Blocked" ? "destructive" : "info";
  return <Badge variant={variant}>{getDeviceStatusLabel(device.status)}</Badge>;
}

export function getDeviceProblemLabel(problem: string) {
  switch (problem) {
  case "disk not encrypted":
    return i18next.t("device:Disk not encrypted");
  case "firewall off":
    return i18next.t("device:Firewall off");
  case "screen lock off":
    return i18next.t("device:Screen lock off");
  case "no posture reported":
    return i18next.t("device:No posture reported");
  default:
    return problem;
  }
}

export function DevicePostureCell({device}: {device: any}) {
  const posture = device.posture ?? {};
  const item = (ok: boolean, label: string) => (
    <span className={ok ? "text-success" : "text-muted-foreground line-through"}>{label}</span>
  );
  return (
    <div className="flex flex-wrap gap-x-2 text-xs">
      {item(!!posture.diskEncrypted, i18next.t("device:Encrypted"))}
      {item(!!posture.firewallEnabled, i18next.t("device:Firewall"))}
      {item(!!posture.screenLockEnabled, i18next.t("device:Screen lock"))}
      {(device.problems ?? []).length > 0 ? (
        <span className="text-warning">{device.problems.map(getDeviceProblemLabel).join(", ")}</span>
      ) : null}
    </div>
  );
}

/** The computers enrolled with the Casdoor device agent, which an admin trusts or blocks. */
export default function DeviceListPage() {
  const organizationName = useRequestOrganization();

  const setStatus = (device: any, status: string, refresh: () => void) =>
    DeviceBackend.updateDevice({...device, status}).then((res: any) => {
      if (res.status === "ok") {
        Setting.showMessage("success", i18next.t("general:Successfully saved"));
        refresh();
      } else {
        Setting.showMessage("error", res.msg);
      }
    });

  const columns: ColumnDef<any>[] = [
    textColumn({dataIndex: "displayName", title: i18next.t("entry:Device"), searchable: true, width: 180}),
    textColumn({dataIndex: "user", title: i18next.t("general:User"), searchable: true, width: 130}),
    {
      dataIndex: "platform",
      title: i18next.t("device:Platform"),
      width: 170,
      render: (value: string, record: any) => (
        <span>
          {value}
          {record.posture?.osVersion ? <span className="text-muted-foreground"> {record.posture.osVersion}</span> : null}
        </span>
      ),
    },
    {
      dataIndex: "status",
      title: i18next.t("general:Status"),
      width: 140,
      filters: valueFilters(["Pending", "Trusted", "Blocked"]),
      render: (_: string, record: any) => <DeviceStatusBadge device={record} />,
    },
    {
      dataIndex: "posture",
      title: i18next.t("device:Posture"),
      width: 260,
      render: (_: any, record: any) => <DevicePostureCell device={record} />,
    },
    dateColumn("lastSeenTime", i18next.t("device:Last seen")),
    clientIpColumn({dataIndex: "lastIp"}),
    textColumn({dataIndex: "agentVersion", title: i18next.t("device:Agent version"), width: 110}),
    dateColumn(),
  ];

  return (
    <CrudListPage
      title={i18next.t("device:Devices")}
      description={i18next.t("device:Devices - Tooltip")}
      columns={columns}
      deps={[organizationName]}
      fetch={(q) => DeviceBackend.getDevices(organizationName, q.page, q.pageSize, q.searchedColumn, q.searchText, q.sortField, q.sortOrder)}
      remove={(r) => DeviceBackend.deleteDevice(r)}
      rowActions={(record, _index, ctx) => [
        record.status !== "Trusted"
          ? {key: "trust", label: i18next.t("device:Trust"), onSelect: () => setStatus(record, "Trusted", ctx.refresh)}
          : null,
        record.status !== "Blocked"
          ? {key: "block", label: i18next.t("rule:Block"), destructive: true, onSelect: () => setStatus(record, "Blocked", ctx.refresh)}
          : null,
      ]}
    />
  );
}
