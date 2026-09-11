import * as React from "react";
import i18next from "i18next";
import {Link} from "react-router-dom";
import {X} from "lucide-react";
import {Badge} from "@/components/ui/badge";
import {ConfirmButton} from "@/components/common/ConfirmButton";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useOrganizationFilter} from "@/hooks/use-organization";
import * as SessionBackend from "@/backend/SessionBackend";
import * as Setting from "@/lib/setting";
import {DeviceIcon, parseUserAgent} from "@/lib/user-agent";

export default function SessionListPage() {
  const organizationName = useOrganizationFilter();
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

  const renderSessionId = (record: any, id: string) => {
    // ids recorded before the infos existed have no info
    const info = (record.sessionInfos ?? []).find((i: any) => i?.sessionId === id);
    const ua = info?.userAgent ? parseUserAgent(info.userAgent) : null;
    const expired = info?.expireTime ? new Date(info.expireTime).getTime() < Date.now() : null;

    return (
      <div key={id} className="flex items-start gap-2 rounded-md border px-2 py-1.5 text-xs">
        {ua ? <DeviceIcon device={ua.device} /> : null}
        <div className="min-w-0 flex-1 space-y-0.5">
          {info ? (
            <div className="flex flex-wrap items-center gap-x-2 gap-y-0.5">
              {ua ? <span className="font-medium" title={info.userAgent}>{ua.label}</span> : null}
              {info.ip ? <span className="font-mono text-muted-foreground">{info.ip}</span> : null}
              {expired === null ? null : (
                <Badge variant={expired ? "secondary" : "success"} className="py-0 text-[10px] font-normal">
                  {expired ? i18next.t("general:Expired") : i18next.t("general:Active")}
                </Badge>
              )}
            </div>
          ) : null}
          {info?.lastActiveTime ? (
            <div className="text-muted-foreground">
              {i18next.t("general:Last active")}: {Setting.getFormattedDate(info.lastActiveTime)}
            </div>
          ) : null}
          <div className="break-all font-mono text-[11px] text-muted-foreground">{id}</div>
        </div>
        <ConfirmButton
          variant="ghost"
          size="iconSm"
          className="h-5 w-5 shrink-0"
          title={i18next.t("general:Sure to delete")}
          description={`${i18next.t("general:Session ID")}: ${id}`}
          onConfirm={() => deleteSession(record, id)}
        >
          <X className="h-3 w-3" />
        </ConfirmButton>
      </div>
    );
  };

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
      width: 420,
      // each id can be signed out on its own, which is what the antd tag's
      // close button did; removing the last one deletes the row
      render: (value: string[], record: any) =>
        !value || value.length === 0 ? null : (
          <div className="flex flex-col gap-1.5">
            {value.map((id) => renderSessionId(record, id))}
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
