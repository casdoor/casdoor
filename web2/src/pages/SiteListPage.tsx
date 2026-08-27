import i18next from "i18next";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, linkColumn, organizationColumn, tagsColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as SiteBackend from "@/backend/SiteBackend";
import {newSite} from "@/pages/defaults";

export default function SiteListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();

  const columns: ColumnDef<any>[] = [
    organizationColumn(),
    textColumn({dataIndex: "tag", title: i18next.t("general:Tag"), width: 110}),
    linkColumn({dataIndex: "name", to: (r) => `/sites/${r.owner}/${r.name}`}),
    dateColumn(),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 170}),
    textColumn({dataIndex: "domain", title: i18next.t("provider:Domain"), width: 190, searchable: true}),
    tagsColumn({dataIndex: "otherDomains", title: i18next.t("application:Other domains"), width: 190}),
    textColumn({dataIndex: "host", title: i18next.t("general:Host"), width: 140}),
    textColumn({dataIndex: "publicIp", title: i18next.t("site:Public IP"), width: 140}),
    textColumn({dataIndex: "node", title: i18next.t("site:Node"), width: 130}),
    textColumn({dataIndex: "sslMode", title: i18next.t("site:Mode"), width: 130}),
    textColumn({dataIndex: "sslCert", title: i18next.t("application:SSL cert"), width: 160}),
    textColumn({dataIndex: "casdoorApplication", title: i18next.t("site:Casdoor app"), width: 170}),
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Sites")}
      columns={columns}
      deps={[organizationName]}
      fetch={(q) =>
        SiteBackend.getSites(organizationName, q.page, q.pageSize, q.searchedColumn, q.searchText, q.sortField, q.sortOrder)
      }
      newRecord={account ? () => newSite(account) : undefined}
      editUrl={(r) => `/sites/${r.owner}/${r.name}`}
      remove={(r) => SiteBackend.deleteSite(r)}
    />
  );
}
