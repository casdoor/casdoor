import i18next from "i18next";
import {Link} from "react-router-dom";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {XlsxImport} from "@/components/crud/XlsxImport";
import {boolColumn, dateColumn, linkColumn, organizationColumn, refsColumn, tagsColumn, textColumn, valueFilters} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {enumColumn, PERMISSION_EFFECTS, PERMISSION_STATES} from "@/lib/enum-labels";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as PermissionBackend from "@/backend/PermissionBackend";
import * as Setting from "@/lib/setting";
import {newPermission} from "@/pages/defaults";

export default function PermissionListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();
  // a normal user may only list the permissions they submitted themselves
  const isAdmin = Setting.isLocalAdminUser(account);
  const isGlobal = account ? Setting.isDefaultOrganizationSelected(account) : false;

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/permissions/${r.owner}/${r.name}`}),
    organizationColumn(),
    dateColumn(),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 160}),
    {
      dataIndex: "model",
      searchable: true,
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
    refsColumn({dataIndex: "users", title: i18next.t("role:Sub users"), urlPrefix: "/users", width: 180, sortable: true, searchable: true}),
    refsColumn({dataIndex: "groups", title: i18next.t("role:Sub groups"), urlPrefix: "/groups", width: 160, sortable: true, searchable: true}),
    refsColumn({dataIndex: "roles", title: i18next.t("role:Sub roles"), urlPrefix: "/roles", width: 160, sortable: true, searchable: true}),
    tagsColumn({dataIndex: "domains", title: i18next.t("role:Sub domains"), width: 140, sortable: true, searchable: true}),
    textColumn({dataIndex: "resourceType", title: i18next.t("permission:Resource type"), width: 130, filters: valueFilters(["Application"])}),
    tagsColumn({dataIndex: "resources", title: i18next.t("general:Resources"), width: 180, sortable: true, searchable: true}),
    tagsColumn({dataIndex: "actions", title: i18next.t("permission:Actions"), width: 140}),
    enumColumn({dataIndex: "effect", title: i18next.t("permission:Effect"), map: PERMISSION_EFFECTS, width: 100, filters: true}),
    boolColumn({dataIndex: "isEnabled", title: i18next.t("general:Is enabled")}),
    textColumn({dataIndex: "submitter", title: i18next.t("permission:Submitter"), width: 130, link: (v, r: any) => `/users/${r.owner}/${encodeURIComponent(v)}`}),
    textColumn({dataIndex: "approver", title: i18next.t("permission:Approver"), width: 130, link: (v, r: any) => `/users/${r.owner}/${encodeURIComponent(v)}`}),
    {
      dataIndex: "approveTime",
      title: i18next.t("permission:Approve time"),
      width: 165,
      sortable: true,
      render: (value) => Setting.getFormattedDate(value),
    },
    enumColumn({dataIndex: "state", title: i18next.t("general:State"), map: PERMISSION_STATES, filters: true}),
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Permissions")}
      columns={columns}
      toolbar={({refresh}) => (
        <XlsxImport
          columns={Setting.getPermissionColumns()}
          templateName="import-permission.xlsx"
          uploadApi="upload-permissions"
          successMessage="Permissions uploaded successfully, refreshing the page"
          onUploaded={refresh}
        />
      )}
      deps={[organizationName, isAdmin, isGlobal]}
      fetch={(q) =>
        (isAdmin ? PermissionBackend.getPermissions : PermissionBackend.getPermissionsBySubmitter)(
          isGlobal ? "" : organizationName,
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
