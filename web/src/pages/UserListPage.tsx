import * as React from "react";
import i18next from "i18next";
import {Link, useNavigate, useParams, useSearchParams} from "react-router-dom";
import {Badge} from "@/components/ui/badge";
import {Avatar, AvatarFallback, AvatarImage} from "@/components/ui/avatar";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {XlsxImport} from "@/components/crud/XlsxImport";
import {boolColumn, dateColumn, linkColumn, organizationColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as OrganizationBackend from "@/backend/OrganizationBackend";
import * as UserBackend from "@/backend/UserBackend";
import * as Setting from "@/lib/setting";
import {newUser} from "@/pages/defaults";

export default function UserListPage() {
  const {account} = useAccount();
  const params = useParams();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const groupName = searchParams.get("groupName") ?? "";
  const organizationName = useRequestOrganization(params.organizationName);

  // antd stops you removing or deleting yourself, or the built-in admin
  const isProtected = (record: any) =>
    (record.owner === account?.owner && record.name === account?.name) ||
    (record.owner === "built-in" && record.name === "admin");

  const [organization, setOrganization] = React.useState<any>({});

  React.useEffect(() => {
    if (!organizationName) {
      return;
    }
    OrganizationBackend.getOrganization("admin", organizationName).then((res: any) => {
      if (res.status === "ok") {
        setOrganization(res.data ?? {});
      }
    });
  }, [organizationName]);

  const isGlobal = account ? Setting.isDefaultOrganizationSelected(account) && !params.organizationName : false;

  const columns: ColumnDef<any>[] = [
    organizationColumn(140, "owner", undefined, "left"),
    {
      dataIndex: "signupApplication",
      title: i18next.t("general:Application"),
      width: 140,
      fixed: "left",
      sortable: true,
      searchable: true,
      render: (value, record) =>
        value ? (
          <Link to={`/applications/${record.owner}/${value}`} className="underline-offset-4 hover:underline">
            {value}
          </Link>
        ) : null,
    },
    linkColumn({dataIndex: "name", to: (r) => `/users/${r.owner}/${r.name}`, width: 150}),
    dateColumn(),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 150}),
    {
      dataIndex: "avatar",
      title: i18next.t("general:Avatar"),
      width: 80,
      align: "center",
      render: (_value, record) => {
        const url = Setting.getEffectiveAvatarUrl(record);
        // antd links the avatar to the image itself
        const avatar = (
          <Avatar className="mx-auto h-9 w-9">
            {url ? <AvatarImage src={url} alt={record.name} /> : null}
            <AvatarFallback style={{backgroundColor: Setting.getAvatarColor(record.name), color: "#fff"}}>
              {(record.name || "?").charAt(0).toUpperCase()}
            </AvatarFallback>
          </Avatar>
        );
        return url ? (
          <a href={url} target="_blank" rel="noreferrer">
            {avatar}
          </a>
        ) : (
          avatar
        );
      },
    },
    {
      dataIndex: "email",
      title: i18next.t("general:Email"),
      width: 180,
      sortable: true,
      searchable: true,
      render: (value) =>
        value ? (
          <a href={`mailto:${value}`} className="underline-offset-4 hover:underline">
            {value}
          </a>
        ) : null,
    },
    textColumn({dataIndex: "phone", title: i18next.t("general:Phone"), width: 130, searchable: true}),
    textColumn({dataIndex: "affiliation", title: i18next.t("user:Affiliation"), width: 140, searchable: true}),
    textColumn({dataIndex: "realName", title: i18next.t("application:Real name"), width: 130, searchable: true}),
    boolColumn({dataIndex: "isVerified", title: i18next.t("user:Is verified")}),
    textColumn({dataIndex: "region", title: i18next.t("user:Country/Region"), width: 120, searchable: true}),
    textColumn({dataIndex: "type", title: i18next.t("general:User type"), width: 130, searchable: true}),
    {
      dataIndex: "tag",
      title: i18next.t("general:Tag"),
      width: 110,
      sortable: true,
      searchable: true,
      render: (value) => (value ? <Badge variant="secondary">{value}</Badge> : null),
    },
    textColumn({dataIndex: "registerType", title: i18next.t("user:Register type"), width: 130, searchable: true}),
    textColumn({dataIndex: "registerSource", title: i18next.t("user:Register source"), width: 160, searchable: true}),
    {
      dataIndex: "balance",
      title: i18next.t("user:Balance"),
      width: 120,
      sortable: true,
      render: (value, record) => Setting.getPriceDisplay(value ?? 0, record.balanceCurrency),
    },
    {
      dataIndex: "balanceCredit",
      title: i18next.t("organization:Balance credit"),
      width: 130,
      sortable: true,
      render: (value, record) => Setting.getPriceDisplay(value ?? 0, record.balanceCurrency),
    },
    textColumn({dataIndex: "balanceCurrency", title: i18next.t("organization:Balance currency"), width: 130}),
    boolColumn({dataIndex: "isAdmin", title: i18next.t("user:Is admin")}),
    boolColumn({dataIndex: "isForbidden", title: i18next.t("user:Is forbidden"), invertColor: true}),
    boolColumn({dataIndex: "isDeleted", title: i18next.t("user:Is deleted"), invertColor: true}),
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Users")}
      description={groupName ? `${i18next.t("general:Groups")}: ${groupName}` : undefined}
      columns={columns}
      formType="users"
      toolbar={({refresh}) => (
        <XlsxImport
          columns={Setting.getUserColumns()}
          templateName="import-user.xlsx"
          uploadApi="upload-users"
          successMessage="Users uploaded successfully, refreshing the page"
          onUploaded={refresh}
        />
      )}
      deps={[organizationName, groupName, isGlobal]}
      fetch={(q) =>
        isGlobal
          ? UserBackend.getGlobalUsers(q.page, q.pageSize, q.searchedColumn, q.searchText, q.sortField, q.sortOrder)
          : UserBackend.getUsers(
            organizationName,
            q.page,
            q.pageSize,
            q.searchedColumn,
            q.searchText,
            q.sortField,
            q.sortOrder,
            groupName,
          )
      }
      newRecord={account ? () => newUser(account, organization, organizationName, groupName) : undefined}
      deleteDisabled={isProtected}
      editUrl={(r) => `/users/${r.owner}/${r.name}`}
      remove={(r) => UserBackend.deleteUser(r)}
      rowActions={(record, _index, {refresh}) => [
        Setting.isLocalAdminUser(account) && record.name !== account?.name
          ? {
            key: "impersonate",
            label: i18next.t("general:Impersonate"),
            onSelect: () => {
              UserBackend.impersonateUser(record.owner, record.name).then((res: any) => {
                if (res.status === "ok") {
                  navigate("/");
                  window.location.reload();
                } else {
                  Setting.showMessage("error", res.msg);
                }
              });
            },
          }
          : null,
        // only offered while the list is scoped to a group, as on the group tree page
        groupName
          ? {
            key: "remove",
            label: i18next.t("general:remove"),
            disabled: isProtected(record),
            // "remove from group", not "delete", so it asks its own question
            confirm: {title: i18next.t("general:Sure to remove"), description: `${record.name ?? ""}`},
            onSelect: () =>
              UserBackend.removeUserFromGroup({groupName, owner: record.owner, name: record.name})
                .then((res: any) => {
                  if (res.status === "ok") {
                    Setting.showMessage("success", i18next.t("general:Successfully removed"));
                    refresh();
                  } else {
                    Setting.showMessage("error", `${i18next.t("general:Failed to remove")}: ${res.msg}`);
                  }
                })
                .catch((error) =>
                  Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`),
                ),
          }
          : null,
      ]}
      actionColumnWidth={groupName ? 340 : 260}
    />
  );
}
