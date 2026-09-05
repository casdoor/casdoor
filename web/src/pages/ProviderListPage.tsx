import i18next from "i18next";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, linkColumn, organizationColumn, textColumn, valueFilters} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as ProviderBackend from "@/backend/ProviderBackend";
import * as Setting from "@/lib/setting";
import {newProvider} from "@/pages/defaults";

/** the provider categories the antd list offers as filters, in its order */
const PROVIDER_CATEGORIES = [
  "Captcha", "Email", "Face ID", "ID Verification", "Log", "MFA", "Notification",
  "OAuth", "Payment", "SAML", "Scan", "SMS", "Storage", "Web3",
];

export default function ProviderListPage({formItems}: {formItems?: any[]} = {}) {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();
  const isGlobal = account ? Setting.isDefaultOrganizationSelected(account) : false;

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/providers/${r.owner}/${r.name}`, width: 160}),
    organizationColumn(),
    dateColumn(),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 170}),
    textColumn({dataIndex: "category", title: i18next.t("general:Category"), width: 120, filters: valueFilters(PROVIDER_CATEGORIES)}),
    {
      dataIndex: "type",
      title: i18next.t("general:Type"),
      width: 150,
      sortable: true,
      // antd's two-level type filter: the provider types grouped by their category
      filters: PROVIDER_CATEGORIES.map((category) => ({
        value: category,
        label: category,
        children: Setting.getProviderTypeOptions(category).map((option: any) => ({value: option.name, label: option.id})),
      })),
      render: (_value, record) => (
        <span className="flex items-center gap-2">
          {Setting.getProviderLogo(record)}
          <span>{record.type}</span>
        </span>
      ),
    },
    {
      ...textColumn({dataIndex: "clientId", title: i18next.t("provider:Client ID"), width: 180, mono: true, searchable: true}),
      render: (value: string) => (value ? Setting.getShortText(value) : null),
    },
    {
      dataIndex: "providerUrl",
      sortable: true,
      searchable: true,
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
      formType="providers"
      formItems={formItems}
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
