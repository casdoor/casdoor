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
import {Link} from "react-router-dom";
import {Badge} from "@/components/ui/badge";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {clientIpColumn, dateColumn, organizationColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useOrganizationFilter} from "@/hooks/use-organization";
import * as MagicLinkBackend from "@/backend/MagicLinkBackend";
import * as Setting from "@/lib/setting";

/** a link can only be revoked while it has neither been used nor expired */
const REVOCABLE_STATUSES = ["created", "sent"];

const STATUS_VARIANTS: Record<string, "default" | "success" | "destructive" | "secondary"> = {
  created: "default",
  sent: "default",
  used: "success",
  failed: "destructive",
  expired: "secondary",
  revoked: "secondary",
};

const Empty = () => <span className="text-muted-foreground">-</span>;

export default function MagicLinkListPage() {
  // GetMagicLinks() filters by "owner", an empty one means every organization
  const owner = useOrganizationFilter();

  const columns: ColumnDef<any>[] = [
    organizationColumn(140, "owner", undefined, "left"),
    textColumn({dataIndex: "name", title: i18next.t("general:Name"), width: 180, searchable: true, mono: true}),
    textColumn({
      dataIndex: "application",
      title: i18next.t("general:Application"),
      width: 160,
      searchable: true,
      link: (value, record: any) => (value ? `/applications/${record.owner}/${value}` : undefined),
    }),
    textColumn({dataIndex: "email", title: i18next.t("general:Email"), width: 200, searchable: true}),
    {
      dataIndex: "requester",
      title: i18next.t("general:User"),
      width: 140,
      sortable: true,
      searchable: true,
      render: (value, record: any) =>
        value ? (
          <Link to={`/users/${record.owner}/${value}`} className="underline-offset-4 hover:underline">
            {value}
          </Link>
        ) : (
          <Empty />
        ),
    },
    {
      dataIndex: "status",
      title: i18next.t("general:Status"),
      width: 110,
      sortable: true,
      searchable: true,
      render: (value) => (value ? <Badge variant={STATUS_VARIANTS[value] ?? "secondary"}>{value}</Badge> : <Empty />),
    },
    textColumn({dataIndex: "authAction", title: i18next.t("magicLink:Auth action"), width: 170, searchable: true}),
    dateColumn(),
    dateColumn("expireTime", i18next.t("magicLink:Expire time")),
    dateColumn("usedTime", i18next.t("magicLink:Used time")),
    clientIpColumn({dataIndex: "remoteAddr"}),
    {
      dataIndex: "lastError",
      title: i18next.t("magicLink:Last error"),
      width: 220,
      sortable: false,
      searchable: true,
      render: (value) => (value ? <span className="text-destructive">{value}</span> : <Empty />),
    },
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Magic Links")}
      columns={columns}
      deps={[owner]}
      rowKey={(record) => `${record.owner}/${record.name}`}
      remove={(record) => MagicLinkBackend.deleteMagicLink(record.owner, record.name)}
      rowActions={(record, _index, {refresh}) => [
        {
          key: "revoke",
          label: i18next.t("magicLink:Revoke"),
          disabled: !REVOCABLE_STATUSES.includes(record.status),
          confirm: {title: `${i18next.t("general:Confirm")} ${i18next.t("magicLink:Revoke")}?`},
          onSelect: () =>
            MagicLinkBackend.revokeMagicLink(record.owner, record.name).then((res: any) => {
              if (res.status === "ok") {
                Setting.showMessage("success", i18next.t("general:Successfully saved"));
                refresh();
              } else {
                Setting.showMessage("error", res.msg);
              }
            }),
        },
      ]}
      fetch={(q) =>
        MagicLinkBackend.getMagicLinks(
          owner,
          q.page,
          q.pageSize,
          q.searchedColumn,
          q.searchText,
          q.sortField,
          q.sortOrder,
        )
      }
    />
  );
}
