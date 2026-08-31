import i18next from "i18next";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, linkColumn, organizationColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as AgentBackend from "@/backend/AgentBackend";
import {newAgent} from "@/pages/defaults";

export default function AgentListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/agents/${r.owner}/${r.name}`}),
    organizationColumn(),
    dateColumn(),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 180}),
    textColumn({dataIndex: "url", title: i18next.t("general:Listening URL"), width: 240, searchable: true}),
    textColumn({dataIndex: "application", title: i18next.t("general:Application"), width: 160, searchable: true}),
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Agents")}
      columns={columns}
      formType="agents"
      deps={[organizationName]}
      fetch={(q) =>
        AgentBackend.getAgents(organizationName, q.page, q.pageSize, q.searchedColumn, q.searchText, q.sortField, q.sortOrder)
      }
      newRecord={account ? () => newAgent(account) : undefined}
      editUrl={(r) => `/agents/${r.owner}/${r.name}`}
      remove={(r) => AgentBackend.deleteAgent(r)}
    />
  );
}
