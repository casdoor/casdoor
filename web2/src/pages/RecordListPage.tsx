import i18next from "i18next";
import {Link} from "react-router-dom";
import {Badge} from "@/components/ui/badge";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {boolColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as RecordBackend from "@/backend/RecordBackend";
import * as Setting from "@/lib/setting";

export default function RecordListPage() {
  const organizationName = useRequestOrganization();

  const columns: ColumnDef<any>[] = [
    textColumn({dataIndex: "name", title: i18next.t("general:Name"), width: 140, searchable: true}),
    textColumn({dataIndex: "id", title: i18next.t("general:ID"), width: 90}),
    textColumn({dataIndex: "clientIp", title: i18next.t("general:Client IP"), width: 140, searchable: true}),
    {
      dataIndex: "createdTime",
      title: i18next.t("general:Timestamp"),
      width: 165,
      sortable: true,
      render: (value) => (
        <span className="whitespace-nowrap tabular-nums text-muted-foreground">{Setting.getFormattedDate(value)}</span>
      ),
    },
    {
      dataIndex: "organization",
      title: i18next.t("general:Organization"),
      width: 130,
      sortable: true,
      searchable: true,
      render: (value) => (
        <Link to={`/organizations/${value}`} className="underline-offset-4 hover:underline">
          {value}
        </Link>
      ),
    },
    {
      dataIndex: "user",
      title: i18next.t("general:User"),
      width: 130,
      sortable: true,
      searchable: true,
      render: (value, record) => (
        <Link to={`/users/${record.organization}/${value}`} className="underline-offset-4 hover:underline">
          {value}
        </Link>
      ),
    },
    textColumn({dataIndex: "method", title: i18next.t("general:Method"), width: 100}),
    textColumn({dataIndex: "requestUri", title: i18next.t("general:Request URI"), width: 240, searchable: true}),
    textColumn({dataIndex: "language", title: i18next.t("user:Language"), width: 100}),
    {
      dataIndex: "statusCode",
      title: i18next.t("rule:Status code"),
      width: 110,
      sortable: true,
      render: (value) =>
        value ? (
          <Badge variant={String(value).startsWith("2") ? "success" : "destructive"}>{value}</Badge>
        ) : null,
    },
    textColumn({dataIndex: "response", title: i18next.t("record:Response"), width: 200}),
    {
      dataIndex: "object",
      title: i18next.t("record:Object"),
      render: (value) =>
        value ? <pre className="max-h-24 max-w-xl overflow-auto rounded-md bg-muted p-2 text-xs">{value}</pre> : null,
    },
    textColumn({dataIndex: "action", title: i18next.t("general:Action"), width: 140, searchable: true}),
    boolColumn({dataIndex: "isTriggered", title: i18next.t("record:Is triggered")}),
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Records")}
      columns={columns}
      deps={[organizationName]}
      showActionColumn={false}
      fetch={(q) =>
        RecordBackend.getRecords(
          organizationName,
          q.page,
          q.pageSize,
          q.searchedColumn,
          q.searchText,
          q.sortField,
          q.sortOrder,
        )
      }
      rowKey={(row, index) => `${row.id ?? index}`}
    />
  );
}
