import i18next from "i18next";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {boolColumn, dateColumn, linkColumn, organizationColumn, textColumn, valueFilters} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as AdapterBackend from "@/backend/AdapterBackend";
import {newAdapter} from "@/pages/defaults";

export default function AdapterListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/adapters/${r.owner}/${r.name}`}),
    organizationColumn(),
    dateColumn(),
    textColumn({dataIndex: "table", title: i18next.t("syncer:Table"), width: 140, searchable: true}),
    boolColumn({dataIndex: "useSameDb", title: i18next.t("adapter:Use same DB"), width: 130}),
    textColumn({dataIndex: "type", title: i18next.t("general:Type"), width: 110, filters: valueFilters(["Database"])}),
    textColumn({dataIndex: "databaseType", title: i18next.t("syncer:Database type"), width: 130}),
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
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Adapters")}
      columns={columns}
      deps={[organizationName]}
      fetch={(q) =>
        AdapterBackend.getAdapters(organizationName, q.page, q.pageSize, q.searchedColumn, q.searchText, q.sortField, q.sortOrder)
      }
      newRecord={account ? () => newAdapter(account) : undefined}
      editUrl={(r) => `/adapters/${r.owner}/${r.name}`}
      remove={(r) => AdapterBackend.deleteAdapter(r)}
    />
  );
}
