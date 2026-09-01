import i18next from "i18next";
import {Button} from "@/components/ui/button";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {boolColumn, dateColumn, linkColumn, textColumn, valueFilters} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as SyncerBackend from "@/backend/SyncerBackend";
import * as Setting from "@/lib/setting";
import {newSyncer} from "@/pages/defaults";

export default function SyncerListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/syncers/${r.organization}/${r.name}`, width: 170}),
    textColumn({dataIndex: "organization", title: i18next.t("general:Organization"), width: 140, searchable: true, link: (v) => `/organizations/${v}`}),
    dateColumn(),
    textColumn({dataIndex: "type", title: i18next.t("general:Type"), width: 120, filters: valueFilters(["Database", "LDAP"])}),
    textColumn({dataIndex: "databaseType", title: i18next.t("syncer:Database type"), width: 140}),
    textColumn({dataIndex: "host", title: i18next.t("general:Host"), width: 130, searchable: true}),
    textColumn({dataIndex: "port", title: i18next.t("general:Port"), width: 90, searchable: true}),
    textColumn({dataIndex: "user", title: i18next.t("general:User"), width: 110, searchable: true}),
    {
      dataIndex: "password",
      sortable: true,
      searchable: true,
      title: i18next.t("general:Password"),
      width: 110,
      render: (value) => (value ? "••••••" : null),
    },
    textColumn({dataIndex: "database", title: i18next.t("syncer:Database"), width: 130}),
    textColumn({dataIndex: "table", title: i18next.t("syncer:Table"), width: 130}),
    textColumn({dataIndex: "syncInterval", title: i18next.t("syncer:Sync interval"), width: 130, searchable: true}),
    boolColumn({dataIndex: "isEnabled", title: i18next.t("general:Is enabled")}),
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Syncers")}
      // Sync is the primary action here, not Edit
      columns={columns}
      deps={[organizationName]}
      fetch={(q) =>
        SyncerBackend.getSyncers(
          "admin",
          organizationName,
          q.page,
          q.pageSize,
          q.searchedColumn,
          q.searchText,
          q.sortField,
          q.sortOrder,
        )
      }
      newRecord={account ? () => newSyncer(account) : undefined}
      editUrl={(r) => `/syncers/${r.organization}/${r.name}`}
      remove={(r) => SyncerBackend.deleteSyncer(r)}
      rowActions={(record) => (
        <Button
          size="sm"
          onClick={() => {
            SyncerBackend.runSyncer("admin", record.name, record.organization).then((res: any) => {
              if (res.status === "ok") {
                Setting.showMessage("success", i18next.t("general:Successfully synced"));
              } else {
                Setting.showMessage("error", res.msg);
              }
            });
          }}
        >
          {i18next.t("general:Sync")}
        </Button>
      )}
      actionColumnWidth={250}
    />
  );
}
