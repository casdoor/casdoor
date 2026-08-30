import i18next from "i18next";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, linkColumn, organizationColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {ScanServerDialog} from "@/components/server/ScanServerDialog";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as ServerBackend from "@/backend/ServerBackend";
import {newServer} from "@/pages/defaults";

export default function ServerListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/servers/${r.owner}/${r.name}`}),
    organizationColumn(),
    dateColumn(),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 180}),
    textColumn({dataIndex: "url", title: i18next.t("general:URL"), width: 260}),
    textColumn({dataIndex: "application", title: i18next.t("general:Application"), width: 160}),
  ];

  return (
    <CrudListPage
      title={i18next.t("general:MCP Servers")}
      columns={columns}
      formType="servers"
      deps={[organizationName]}
      fetch={(q) =>
        ServerBackend.getServers(organizationName, q.page, q.pageSize, q.searchedColumn, q.searchText, q.sortField, q.sortOrder)
      }
      toolbar={({refresh}) => <ScanServerDialog organizationName={organizationName} onAdded={refresh} />}
      newRecord={account ? () => newServer(account) : undefined}
      editUrl={(r) => `/servers/${r.owner}/${r.name}`}
      remove={(r) => ServerBackend.deleteServer(r)}
    />
  );
}
