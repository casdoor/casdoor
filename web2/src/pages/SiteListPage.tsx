import i18next from "i18next";
import {Link} from "react-router-dom";
import {Badge} from "@/components/ui/badge";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {linkColumn, organizationColumn, textColumn} from "@/components/crud/columns";
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
    textColumn({dataIndex: "tag", title: i18next.t("general:Tag"), width: 110, link: (v, r: any) => (v ? `/nodes/${r.owner}/${v}` : undefined)}),
    linkColumn({dataIndex: "name", to: (r) => `/sites/${r.owner}/${r.name}`}),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 170}),
    textColumn({
      dataIndex: "domain",
      title: i18next.t("provider:Domain"),
      width: 190,
      searchable: true,
      link: (v) => (v ? `https://${v}` : undefined),
      linkExternal: true,
    }),
    {
      dataIndex: "otherDomains",
      title: i18next.t("application:Other domains"),
      width: 190,
      // each one opens, and greys out when the site only redirects to it
      render: (value: string[], record: any) =>
        !value || value.length === 0 ? null : (
          <div className="flex flex-wrap gap-1">
            {value.map((domain) => (
              <a key={domain} href={`https://${domain}`} target="_blank" rel="noreferrer">
                <Badge variant={record.needRedirect ? "secondary" : "info"} className="font-normal">
                  {domain}
                </Badge>
              </a>
            ))}
          </div>
        ),
    },
    {
      dataIndex: "host",
      title: i18next.t("general:Host"),
      width: 140,
      sortable: true,
      // antd shows host:port, and warns when the site is not Active
      render: (_value, record: any) => {
        const host = record.host ? `${record.host}:${record.port}` : record.port;
        if (record.status === "Active") {
          return host;
        }
        return <Badge variant="warning">{host}</Badge>;
      },
    },
    {
      dataIndex: "hosts",
      title: i18next.t("site:Hosts"),
      width: 170,
      render: (value: string[]) =>
        !Array.isArray(value) || value.length === 0 ? null : (
          <div className="flex flex-wrap gap-1">
            {value.map((host) => (
              <Badge key={host} variant="info" className="font-normal">
                {host}
              </Badge>
            ))}
          </div>
        ),
    },
    {
      dataIndex: "nodes",
      title: i18next.t("site:Nodes"),
      width: 150,
      // the rows are node objects, and their colour carries the node's health;
      // a node that reports a version links to its release notes
      render: (value: any[], record: any) =>
        !Array.isArray(value) || value.length === 0 ? null : (
          <div className="flex flex-wrap gap-1">
            {value.map((node: any) => {
              const versionInfo = Setting.getVersionInfo(node.version, record.name);
              let variant = node.message === "" ? "info" : "destructive";
              if (variant === "info" && node.provider !== "") {
                variant = node.version === "" ? "warning" : "success";
              }
              const badge = (
                <Badge variant={variant as any} className="font-normal" title={node.message || undefined}>
                  {versionInfo === null ? node.name : `${node.name} (${versionInfo.text})`}
                </Badge>
              );
              return versionInfo === null ? (
                <span key={node.name}>{badge}</span>
              ) : (
                <a key={node.name} href={versionInfo.link} target="_blank" rel="noreferrer">
                  {badge}
                </a>
              );
            })}
          </div>
        ),
    },
    {
      dataIndex: "rules",
      title: i18next.t("general:Rules"),
      width: 170,
      render: (value: string[]) =>
        !value || value.length === 0 ? null : (
          <div className="flex flex-wrap gap-1">
            {value.map((rule) => (
              <Link key={rule} to={`/rules/${rule}`}>
                <Badge variant="info" className="font-normal">
                  {rule}
                </Badge>
              </Link>
            ))}
          </div>
        ),
    },
    textColumn({dataIndex: "publicIp", title: i18next.t("site:Public IP"), width: 140}),
    textColumn({dataIndex: "node", title: i18next.t("site:Node"), width: 130}),
    textColumn({dataIndex: "sslMode", title: i18next.t("site:Mode"), width: 130}),
    textColumn({dataIndex: "sslCert", title: i18next.t("application:SSL cert"), width: 160, link: (v) => (v ? `/certs/admin/${v}` : undefined)}),
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
