import * as React from "react";
import i18next from "i18next";
import {Link} from "react-router-dom";
import {Badge} from "@/components/ui/badge";
import {Sheet, SheetContent, SheetHeader, SheetTitle} from "@/components/ui/sheet";
import {Switch} from "@/components/ui/switch";
import {CodeEditor} from "@/components/common/CodeEditor";
import {DescriptionList} from "@/components/common/DescriptionList";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {clientIpColumn, textColumn, valueFilters} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as RecordBackend from "@/backend/RecordBackend";
import * as Setting from "@/lib/setting";

/** the actions for which the backend keeps an `isTriggered` flag worth showing */
const TRIGGERABLE_ACTIONS = ["signup", "login", "logout", "update-user", "new-user"];

/** pretty-prints the stored object, leaving it alone when it is not JSON */
function jsonStrFormatter(value: any): string {
  if (!value) {
    return "";
  }
  try {
    return JSON.stringify(JSON.parse(String(value)), null, 2);
  } catch {
    return String(value);
  }
}

/**
 * The antd list opens each record in a right-hand drawer, because the row is far
 * too wide to read in the table — the response and the object are whole JSON
 * documents. Same contents here, in a sheet.
 */
function RecordDetailSheet({record, onClose}: {record: any; onClose: () => void}) {
  const field = (key: string) => record?.[key] ?? "";

  const items = [
    {label: i18next.t("general:ID"), children: field("id")},
    {label: i18next.t("general:Client IP"), children: field("clientIp")},
    {label: i18next.t("general:Timestamp"), children: Setting.getFormattedDate(field("createdTime"))},
    {
      label: i18next.t("general:Organization"),
      children: field("organization") ? (
        <Link to={`/organizations/${field("organization")}`} className="underline-offset-4 hover:underline">
          {field("organization")}
        </Link>
      ) : null,
    },
    {
      label: i18next.t("general:User"),
      children: field("user") ? (
        <Link to={`/users/${field("organization")}/${field("user")}`} className="underline-offset-4 hover:underline">
          {field("user")}
        </Link>
      ) : null,
    },
    {label: i18next.t("general:Method"), children: field("method")},
    {label: i18next.t("general:Request URI"), children: field("requestUri")},
    {label: i18next.t("user:Language"), children: field("language")},
    {label: i18next.t("rule:Status code"), children: field("statusCode")},
    {label: i18next.t("general:Action"), children: field("action")},
    {
      label: i18next.t("record:Response"),
      children: <CodeEditor value={String(field("response"))} onChange={() => undefined} height={120} readOnly />,
    },
    {
      label: i18next.t("record:Object"),
      children: (
        <CodeEditor value={jsonStrFormatter(field("object"))} onChange={() => undefined} language="json" height={260} readOnly />
      ),
    },
  ];

  return (
    <Sheet open={record !== null} onOpenChange={(open) => !open && onClose()}>
      <SheetContent side="right" className="flex w-full flex-col gap-0 overflow-y-auto p-0 sm:max-w-[720px]">
        <SheetHeader className="border-b p-4 pr-12 text-left">
          <SheetTitle className="text-base">{i18next.t("general:Detail")}</SheetTitle>
        </SheetHeader>
        <div className="p-4">{record ? <DescriptionList items={items} /> : null}</div>
      </SheetContent>
    </Sheet>
  );
}

export default function RecordListPage() {
  const organizationName = useRequestOrganization();
  const [detail, setDetail] = React.useState<any>(null);

  const columns: ColumnDef<any>[] = [
    textColumn({dataIndex: "name", title: i18next.t("general:Name"), width: 140, searchable: true}),
    textColumn({dataIndex: "id", title: i18next.t("general:ID"), width: 90, searchable: true}),
    clientIpColumn(),
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
    textColumn({
      dataIndex: "method",
      title: i18next.t("general:Method"),
      width: 100,
      filters: valueFilters(["GET", "HEAD", "POST", "PUT", "DELETE", "CONNECT", "OPTIONS", "TRACE", "PATCH"]),
    }),
    textColumn({dataIndex: "requestUri", title: i18next.t("general:Request URI"), width: 240, searchable: true}),
    textColumn({dataIndex: "language", title: i18next.t("user:Language"), width: 100, searchable: true}),
    {
      dataIndex: "statusCode",
      searchable: true,
      title: i18next.t("rule:Status code"),
      width: 110,
      sortable: true,
      render: (value) =>
        value ? (
          <Badge variant={String(value).startsWith("2") ? "success" : "destructive"}>{value}</Badge>
        ) : null,
    },
    textColumn({dataIndex: "response", title: i18next.t("record:Response"), width: 200, searchable: true}),
    {
      dataIndex: "object",
      sortable: true,
      searchable: true,
      title: i18next.t("record:Object"),
      render: (value) =>
        value ? <pre className="max-h-24 max-w-xl overflow-auto rounded-md bg-muted p-2 text-xs">{value}</pre> : null,
    },
    textColumn({dataIndex: "action", title: i18next.t("general:Action"), width: 140, searchable: true}),
    {
      dataIndex: "isTriggered",
      title: i18next.t("record:Is triggered"),
      width: 110,
      fixed: "right",
      sortable: true,
      align: "center",
      // only the account lifecycle actions carry the flag; the rest leave the cell empty
      render: (value, record) =>
        TRIGGERABLE_ACTIONS.includes(record.action) ? (
          <span className="inline-flex items-center gap-1.5">
            <Switch checked={Boolean(value)} disabled aria-readonly className="opacity-100" />
            <span className="text-xs text-muted-foreground">
              {value ? i18next.t("general:ON") : i18next.t("general:OFF")}
            </span>
          </span>
        ) : null,
    },
  ];

  return (
    <>
      <CrudListPage
        title={i18next.t("general:Records")}
        columns={columns}
        deps={[organizationName]}
        actionColumnWidth={110}
        rowActions={(record) => [
          {key: "view", label: i18next.t("general:View"), onSelect: () => setDetail(record)},
        ]}
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
      <RecordDetailSheet record={detail} onClose={() => setDetail(null)} />
    </>
  );
}
