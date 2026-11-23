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
import {dateColumn, linkColumn, textColumn, valueFilters} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import {FlowTriggers, getFlowTriggerLabel} from "@/components/flow/FlowStepEditor";
import * as FlowBackend from "@/backend/FlowBackend";
import * as Setting from "@/lib/setting";

export function newFlow(organizationName: string) {
  const randomName = Setting.getRandomName();
  return {
    owner: organizationName,
    name: `flow_${randomName}`,
    createdTime: new Date().toISOString(),
    displayName: `New Flow - ${randomName}`,
    description: "",
    trigger: "pre-signin",
    applications: [],
    order: 0,
    isEnabled: false,
    steps: [],
  };
}

/** The no-code flows of the organization, which run at the sign-up, the sign-in and the token issuance. */
export default function FlowListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/flows/${r.owner}/${r.name}`, width: 180}),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 220}),
    {
      dataIndex: "trigger",
      title: i18next.t("flow:Trigger"),
      width: 170,
      filters: valueFilters(FlowTriggers),
      render: (value: string) => getFlowTriggerLabel(value),
    },
    {
      dataIndex: "steps",
      title: i18next.t("flow:Steps"),
      width: 80,
      render: (value: any[]) => (value ?? []).length,
    },
    {
      dataIndex: "isEnabled",
      title: i18next.t("general:Enabled"),
      width: 100,
      render: (value: boolean) => <Badge variant={value ? "success" : "secondary"}>{value ? i18next.t("general:Enabled") : i18next.t("general:Disabled")}</Badge>,
    },
    textColumn({dataIndex: "order", title: i18next.t("application:Order"), width: 80}),
    textColumn({dataIndex: "runCount", title: i18next.t("flow:Runs"), width: 90}),
    dateColumn("lastRunTime", i18next.t("flow:Last run")),
    dateColumn(),
  ];

  return (
    <CrudListPage
      title={i18next.t("flow:Flows")}
      description={i18next.t("flow:Flows - Tooltip")}
      columns={columns}
      deps={[organizationName]}
      fetch={(q) => FlowBackend.getFlows(organizationName, q.page, q.pageSize, q.searchedColumn, q.searchText, q.sortField, q.sortOrder)}
      newRecord={account ? () => newFlow(organizationName) : undefined}
      add={(record) => FlowBackend.addFlow(record)}
      editUrl={(r) => `/flows/${r.owner}/${r.name}`}
      remove={(r) => FlowBackend.deleteFlow(r)}
    />
  );
}
