import i18next from "i18next";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {boolColumn, dateColumn, linkColumn, organizationColumn, refsColumn, tagsColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as RoleBackend from "@/backend/RoleBackend";
import {newRole} from "@/pages/defaults";

export default function RoleListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/roles/${r.owner}/${r.name}`}),
    organizationColumn(),
    dateColumn(),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 160}),
    refsColumn({dataIndex: "users", title: i18next.t("role:Sub users"), urlPrefix: "/users", width: 200}),
    refsColumn({dataIndex: "groups", title: i18next.t("role:Sub groups"), urlPrefix: "/groups", width: 180}),
    refsColumn({dataIndex: "roles", title: i18next.t("role:Sub roles"), urlPrefix: "/roles", width: 180}),
    tagsColumn({dataIndex: "domains", title: i18next.t("role:Sub domains"), width: 160}),
    boolColumn({dataIndex: "isEnabled", title: i18next.t("general:Is enabled")}),
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Roles")}
      columns={columns}
      deps={[organizationName]}
      fetch={(q) =>
        RoleBackend.getRoles(organizationName, q.page, q.pageSize, q.searchedColumn, q.searchText, q.sortField, q.sortOrder)
      }
      newRecord={account ? () => newRole(account) : undefined}
      editUrl={(r) => `/roles/${r.owner}/${r.name}`}
      remove={(r) => RoleBackend.deleteRole(r)}
    />
  );
}
