import i18next from "i18next";
import {Badge} from "@/components/ui/badge";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, linkColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import * as RadiusClientBackend from "@/backend/RadiusClientBackend";
import {newRadiusClient} from "@/pages/defaults";

export default function RadiusClientListPage() {
  const {account} = useAccount();

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/radius-clients/${r.owner}/${r.name}`}),
    dateColumn(),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 170}),
    textColumn({dataIndex: "organization", title: i18next.t("general:Organization"), searchable: true, width: 150}),
    textColumn({dataIndex: "ipRanges", title: i18next.t("radiusClient:IP ranges"), searchable: true, width: 260}),
    {
      dataIndex: "state",
      searchable: true,
      title: i18next.t("general:State"),
      width: 110,
      sortable: true,
      render: (value) => <Badge variant={value === "Active" ? "success" : "secondary"}>{value}</Badge>,
    },
  ];

  return (
    <CrudListPage
      title={i18next.t("general:RADIUS Clients")}
      columns={columns}
      fetch={(q) => RadiusClientBackend.getRadiusClients(q.page, q.pageSize, q.searchedColumn, q.searchText, q.sortField, q.sortOrder)}
      newRecord={account ? () => newRadiusClient() : undefined}
      editUrl={(r) => `/radius-clients/${r.owner}/${r.name}`}
      remove={(r) => RadiusClientBackend.deleteRadiusClient(r)}
    />
  );
}
