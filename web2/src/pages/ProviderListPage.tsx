import i18next from "i18next";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, linkColumn, organizationColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as ProviderBackend from "@/backend/ProviderBackend";
import * as Setting from "@/lib/setting";
import {newProvider} from "@/pages/defaults";

export default function ProviderListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();
  const isGlobal = account ? Setting.isDefaultOrganizationSelected(account) : false;

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/providers/${r.owner}/${r.name}`, width: 160}),
    organizationColumn(),
    dateColumn(),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 170}),
    textColumn({dataIndex: "category", title: i18next.t("general:Category"), width: 120}),
    {
      dataIndex: "type",
      title: i18next.t("general:Type"),
      width: 150,
      sortable: true,
      render: (_value, record) => (
        <span className="flex items-center gap-2">
          {Setting.getProviderLogo(record)}
          <span>{record.type}</span>
        </span>
      ),
    },
    textColumn({dataIndex: "clientId", title: i18next.t("provider:Client ID"), width: 180, mono: true}),
    {
      dataIndex: "providerUrl",
      title: i18next.t("provider:Provider URL"),
      width: 220,
      render: (value) =>
        value ? (
          <a href={value} target="_blank" rel="noreferrer" className="underline-offset-4 hover:underline">
            {Setting.getShortText(value)}
          </a>
        ) : null,
    },
  ];

  return (
    <CrudListPage
      title={i18next.t("application:Providers")}
      columns={columns}
      deps={[organizationName, isGlobal]}
      fetch={(q) =>
        isGlobal
          ? ProviderBackend.getGlobalProviders(q.page, q.pageSize, q.searchedColumn, q.searchText, q.sortField, q.sortOrder)
          : ProviderBackend.getProviders(
            organizationName,
            q.page,
            q.pageSize,
            q.searchedColumn,
            q.searchText,
            q.sortField,
            q.sortOrder,
          )
      }
      newRecord={account ? () => newProvider(account, organizationName) : undefined}
      editUrl={(r) => `/providers/${r.owner}/${r.name}`}
      remove={(r) => ProviderBackend.deleteProvider(r)}
    />
  );
}
