import i18next from "i18next";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {linkColumn, organizationColumn, tagsColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as SiteBackend from "@/backend/SiteBackend";
import * as Setting from "@/lib/setting";
import {newSite} from "@/pages/defaults";

export default function SiteListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();
  // with the default organization selected an admin sees every organization's sites
  const isGlobal = account ? Setting.isDefaultOrganizationSelected(account) : false;

  const columns: ColumnDef<any>[] = [
    organizationColumn(110, "owner", i18next.t("general:Owner")),
    textColumn({dataIndex: "tag", title: i18next.t("general:Tag"), width: 110}),
    linkColumn({dataIndex: "name", to: (r) => `/sites/${r.owner}/${r.name}`}),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 170}),
    textColumn({dataIndex: "domain", title: i18next.t("provider:Domain"), width: 190, searchable: true}),
    tagsColumn({dataIndex: "otherDomains", title: i18next.t("application:Other domains"), width: 190}),
    textColumn({dataIndex: "host", title: i18next.t("general:Host"), width: 140}),
    tagsColumn({dataIndex: "hosts", title: i18next.t("site:Hosts"), width: 170}),
    tagsColumn({dataIndex: "nodes", title: i18next.t("site:Nodes"), width: 150}),
    tagsColumn({dataIndex: "rules", title: i18next.t("general:Rules"), width: 170}),
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
      deps={[organizationName, isGlobal]}
      fetch={(q) =>
        isGlobal
          ? SiteBackend.getGlobalSites()
          : SiteBackend.getSites(organizationName, q.page, q.pageSize, q.searchedColumn, q.searchText, q.sortField, q.sortOrder)
      }
      newRecord={account ? () => newSite(account) : undefined}
      editUrl={(r) => `/sites/${r.owner}/${r.name}`}
      remove={(r) => SiteBackend.deleteSite(r)}
    />
  );
}
