import * as React from "react";
import i18next from "i18next";
import {Link} from "react-router-dom";
import {X} from "lucide-react";
import {Badge} from "@/components/ui/badge";
import {ConfirmButton} from "@/components/common/ConfirmButton";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as SessionBackend from "@/backend/SessionBackend";
import * as Setting from "@/lib/setting";

export default function SessionListPage() {
  const organizationName = useRequestOrganization();
  // signing one id out changes a row rather than the page, so the list is
  // re-fetched by bumping a dep
  const [nonce, setNonce] = React.useState(0);

  const deleteSession = (record: any, sessionId: string) =>
    SessionBackend.deleteSession(record, sessionId)
      .then((res: any) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully deleted"));
          setNonce((n) => n + 1);
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to delete")}: ${res.msg}`);
        }
      })
      .catch((error: any) =>
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`),
      );

  const columns: ColumnDef<any>[] = [
    {
      dataIndex: "name",
      title: i18next.t("general:Name"),
      width: 180,
      fixed: "left",
      sortable: true,
      searchable: true,
      render: (value, record) => (
        <Link to={`/users/${record.owner}/${value}`} className="font-medium underline-offset-4 hover:underline">
          {value}
        </Link>
      ),
    },
    {
      dataIndex: "owner",
      title: i18next.t("general:Organization"),
      width: 150,
      sortable: true,
      searchable: true,
      render: (value) => (
        <Link to={`/organizations/${value}`} className="underline-offset-4 hover:underline">
          {value}
        </Link>
      ),
    },
    dateColumn(),
    {
      dataIndex: "sessionId",
      title: i18next.t("general:Session ID"),
      // each id can be signed out on its own, which is what the antd tag's
      // close button did; removing the last one deletes the row
      render: (value: string[], record: any) =>
        !value || value.length === 0 ? null : (
          <div className="flex flex-wrap gap-1">
            {value.map((id) => (
              <Badge key={id} variant="secondary" className="gap-1 py-0.5 font-mono text-[11px] font-normal">
                {id}
                <ConfirmButton
                  variant="ghost"
                  size="iconSm"
                  className="h-4 w-4 shrink-0"
                  title={i18next.t("general:Sure to delete")}
                  description={`${i18next.t("general:Session ID")}: ${id}`}
                  onConfirm={() => deleteSession(record, id)}
                >
                  <X className="h-3 w-3" />
                </ConfirmButton>
              </Badge>
            ))}
          </div>
        ),
    },
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Sessions")}
      columns={columns}
      deps={[organizationName, nonce]}
      fetch={(q) =>
        SessionBackend.getSessions(
          organizationName,
          q.page,
          q.pageSize,
          q.searchedColumn,
          q.searchText,
          q.sortField,
          q.sortOrder,
        )
      }
      remove={(r) => SessionBackend.deleteSession(r)}
      actionColumnWidth={120}
    />
  );
}
