import i18next from "i18next";
import {Link} from "react-router-dom";
import {Badge} from "@/components/ui/badge";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {boolColumn, dateColumn, linkColumn, organizationColumn, refsColumn, tagsColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as PermissionBackend from "@/backend/PermissionBackend";
import * as Setting from "@/lib/setting";
import {newPermission} from "@/pages/defaults";

const stateVariant = (state: string) =>
  state === "Approved" ? "success" : state === "Pending" ? "warning" : "destructive";

export default function PermissionListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/permissions/${r.owner}/${r.name}`}),
    organizationColumn(),
    dateColumn(),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 160}),
    {
      dataIndex: "model",
      title: i18next.t("general:Model"),
      width: 140,
      sortable: true,
      render: (value, record) =>
        value ? (
          <Link to={`/models/${record.owner}/${value}`} className="underline-offset-4 hover:underline">
            {value}
          </Link>
        ) : null,
    },
    refsColumn({dataIndex: "users", title: i18next.t("role:Sub users"), urlPrefix: "/users", width: 180}),
    refsColumn({dataIndex: "groups", title: i18next.t("role:Sub groups"), urlPrefix: "/groups", width: 160}),
    refsColumn({dataIndex: "roles", title: i18next.t("role:Sub roles"), urlPrefix: "/roles", width: 160}),
    tagsColumn({dataIndex: "domains", title: i18next.t("role:Sub domains"), width: 140}),
    textColumn({dataIndex: "resourceType", title: i18next.t("permission:Resource type"), width: 130}),
    tagsColumn({dataIndex: "resources", title: i18next.t("general:Resources"), width: 180}),
    tagsColumn({dataIndex: "actions", title: i18next.t("permission:Actions"), width: 140}),
    textColumn({dataIndex: "effect", title: i18next.t("permission:Effect"), width: 100}),
    boolColumn({dataIndex: "isEnabled", title: i18next.t("general:Is enabled")}),
    textColumn({dataIndex: "submitter", title: i18next.t("permission:Submitter"), width: 130}),
    textColumn({dataIndex: "approver", title: i18next.t("permission:Approver"), width: 130}),
    {
      dataIndex: "approveTime",
      title: i18next.t("permission:Approve time"),
      width: 165,
      sortable: true,
      render: (value) => Setting.getFormattedDate(value),
    },
    {
      dataIndex: "state",
      title: i18next.t("general:State"),
      width: 110,
      sortable: true,
      render: (value) => <Badge variant={stateVariant(value) as any}>{value}</Badge>,
    },
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Permissions")}
      columns={columns}
      deps={[organizationName]}
      fetch={(q) =>
        PermissionBackend.getPermissions(
          organizationName,
          q.page,
          q.pageSize,
          q.searchedColumn,
          q.searchText,
          q.sortField,
          q.sortOrder,
        )
      }
      newRecord={account ? () => newPermission(account) : undefined}
      editUrl={(r) => `/permissions/${r.owner}/${r.name}`}
      remove={(r) => PermissionBackend.deletePermission(r)}
    />
  );
}
