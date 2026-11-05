import i18next from "i18next";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, linkColumn, organizationColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useOrganizationFilter} from "@/hooks/use-organization";
import * as RadiusClientBackend from "@/backend/RadiusClientBackend";
import {newRadiusClient} from "@/pages/defaults";

export default function RadiusClientListPage() {
  const {account} = useAccount();
  const organizationName = useOrganizationFilter();

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/radius-clients/${r.owner}/${r.name}`}),
    organizationColumn(),
    dateColumn(),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 180}),
    {
      dataIndex: "secret",
      title: i18next.t("general:Secret"),
      width: 120,
      render: (value) => (value ? "••••••••" : ""),
    },
    {
      dataIndex: "ipCidr",
      title: i18next.t("general:IP CIDR"),
      width: 180,
      render: (value) => (
        <span className="font-mono text-xs">{value}</span>
      ),
    },
    textColumn({dataIndex: "state", title: i18next.t("general:State"), width: 100}),
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Radius Clients")}
      columns={columns}
      deps={[organizationName]}
      fetch={(q) =>
        RadiusClientBackend.getRadiusClients(organizationName, q.page, q.pageSize, q.searchedColumn, q.searchText, q.sortField, q.sortOrder)
      }
      newRecord={account ? () => newRadiusClient(account) : undefined}
      editUrl={(r) => `/radius-clients/${r.owner}/${r.name}`}
      remove={(r) => RadiusClientBackend.deleteRadiusClient(r)}
    />
  );
}
