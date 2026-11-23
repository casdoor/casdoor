import * as React from "react";
import i18next from "i18next";
import {Link, useNavigate, useParams, useSearchParams} from "react-router-dom";
import {Badge} from "@/components/ui/badge";
import {Avatar, AvatarFallback, AvatarImage} from "@/components/ui/avatar";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from "@/components/ui/select";
import {Switch} from "@/components/ui/switch";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {XlsxImport} from "@/components/crud/XlsxImport";
import {boolColumn, dateColumn, linkColumn, organizationColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as OrganizationBackend from "@/backend/OrganizationBackend";
import * as UserBackend from "@/backend/UserBackend";
import * as Setting from "@/lib/setting";
import {canDelegated, isDelegatedAdmin} from "@/ee/lib/delegated-admin";
import {newUser} from "@/pages/defaults";
import {cn} from "@/lib/utils";

/**
 * Also embedded by the group tree page, which scopes the list to the selected
 * group through props instead of the route.
 */
export default function UserListPage({
  organizationName: organizationNameProp,
  groupName: groupNameProp,
  formItems,
}: {organizationName?: string; groupName?: string; formItems?: any[]} = {}) {
  const {account} = useAccount();
  const params = useParams();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const embedded = groupNameProp !== undefined;
  const groupName = groupNameProp ?? searchParams.get("groupName") ?? "";
  const scopedOrganizationName = organizationNameProp ?? params.organizationName;
  const organizationName = useRequestOrganization(scopedOrganizationName);

  // antd stops you removing or deleting yourself, or the built-in admin
  const isProtected = (record: any) =>
    (record.owner === account?.owner && record.name === account?.name) ||
    (record.owner === "built-in" && record.name === "admin");

  const [organization, setOrganization] = React.useState<any>({});
  // "enabled" lists the users who can sign in, "disabled" the forbidden or deleted ones
  const [status, setStatus] = React.useState("");
  const [forbidding, setForbidding] = React.useState<any>(null);
  const [saving, setSaving] = React.useState(false);
  const refreshRef = React.useRef<() => void>(() => {});

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

  const isGlobal = account ? Setting.isDefaultOrganizationSelected(account) && !scopedOrganizationName : false;
  // a delegated admin sees the users of their groups, and adds or deletes them only when allowed to manage users
  const delegated = isDelegatedAdmin(account);
  const canManageUsers = !delegated || canDelegated(account, "users");
  const canForbid = (record: any) => canManageUsers && (delegated || Setting.isLocalAdminUser(account)) && !isProtected(record);

  const saveForbidden = async() => {
    setSaving(true);
    try {
      const res = await UserBackend.updateUser(forbidding.owner, forbidding.name, {isForbidden: !forbidding.isForbidden}, ["isForbidden"]);
      if (res.status === "ok") {
        Setting.showMessage("success", i18next.t("general:Successfully saved"));
        setForbidding(null);
        refreshRef.current();
      } else {
        Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
      }
    } catch (error) {
      Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
    } finally {
      setSaving(false);
    }
  };

  const tagLabels = React.useMemo(() => {
    const map: Record<string, string> = {};
    ((organization?.tags ?? []) as string[]).forEach((tag) => {
      const tokens = tag.split("|");
      map[tokens[0]] = Setting.getLanguage() !== "zh" ? tokens[0] : tokens[1] ?? tokens[0];
    });
    return map;
  }, [organization]);

  const nameColumn = linkColumn<any>({dataIndex: "name", to: (r) => `/users/${r.owner}/${r.name}`, width: 150});
  const renderName = nameColumn.render;
  nameColumn.render = (value, record, index) => (
    <span className="inline-flex items-center gap-2">
      {renderName ? renderName(value, record, index) : value}
      {record.pendingApproval ? <Badge variant="outline">{i18next.t("user:Pending approval")}</Badge> : null}
    </span>
  );

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
    nameColumn,
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
    {
      dataIndex: "region",
      title: i18next.t("user:Country/Region"),
      width: 140,
      sortable: true,
      searchable: true,
      // the column stores the ISO code, antd shows the country's own name
      render: (value) =>
        value ? Setting.initCountries().getName(value, Setting.getLanguage(), {select: "official"}) ?? value : null,
    },
    textColumn({dataIndex: "type", title: i18next.t("general:User type"), width: 130, searchable: true}),
    {
      dataIndex: "tag",
      title: i18next.t("general:Tag"),
      width: 110,
      sortable: true,
      searchable: true,
      // an organization stores its tags as "<name>|<zh name>"; show the localized one
      render: (value) => (value ? <Badge variant="secondary">{tagLabels[value] ?? value}</Badge> : null),
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
    {
      dataIndex: "pendingApproval",
      title: i18next.t("user:Pending approval"),
      width: 120,
      sortable: true,
      align: "center",
      render: (value) => (value ? <Badge variant="outline">{i18next.t("user:Pending approval")}</Badge> : null),
    },
    {
      ...boolColumn<any>({dataIndex: "isForbidden", title: i18next.t("user:Is forbidden"), invertColor: true}),
      render: (value, record) => (
        <span className="inline-flex items-center gap-1.5">
          <Switch
            checked={Boolean(value)}
            disabled={!canForbid(record)}
            onCheckedChange={() => setForbidding(record)}
            aria-label={`${i18next.t("user:Is forbidden")}: ${record.name}`}
            className={cn(!canForbid(record) && "opacity-100", "data-[state=checked]:bg-destructive")}
          />
          <span className="text-xs text-muted-foreground">
            {value ? i18next.t("general:ON") : i18next.t("general:OFF")}
          </span>
        </span>
      ),
    },
    boolColumn({dataIndex: "isDeleted", title: i18next.t("user:Is deleted"), invertColor: true}),
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Users")}
      description={groupName ? `${i18next.t("general:Groups")}: ${groupName}` : undefined}
      columns={columns}
      formType="users"
      formItems={formItems}
      toolbar={({refresh, setQuery}) => {
        refreshRef.current = refresh;
        return (
          <>
            <Select
              value={status || "all"}
              onValueChange={(value) => {
                setStatus(value === "all" ? "" : value);
                setQuery({page: 1});
              }}
            >
              <SelectTrigger className="h-8 w-28" aria-label={i18next.t("general:Status")}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">{i18next.t("general:All")}</SelectItem>
                <SelectItem value="enabled">{i18next.t("general:Enabled")}</SelectItem>
                <SelectItem value="disabled">{i18next.t("general:Disabled")}</SelectItem>
              </SelectContent>
            </Select>
            {delegated ? null : (
              <XlsxImport
                columns={Setting.getUserColumns()}
                templateName="import-user.xlsx"
                uploadApi="upload-users"
                successMessage="Users uploaded successfully, refreshing the page"
                onUploaded={refresh}
              />
            )}
            <AlertDialog open={forbidding !== null} onOpenChange={(open) => !open && !saving && setForbidding(null)}>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>
                    {forbidding?.isForbidden ? i18next.t("user:Allow this user to sign in again?") : i18next.t("user:Forbid this user from signing in?")}
                  </AlertDialogTitle>
                  <AlertDialogDescription>{forbidding ? `${forbidding.owner}/${forbidding.name}` : null}</AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel disabled={saving}>{i18next.t("general:Cancel")}</AlertDialogCancel>
                  <AlertDialogAction
                    className={cn(!forbidding?.isForbidden && "bg-destructive text-destructive-foreground hover:bg-destructive/90")}
                    disabled={saving}
                    onClick={(e) => {
                      e.preventDefault();
                      saveForbidden();
                    }}
                  >
                    {i18next.t("general:OK")}
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </>
        );
      }}
      deps={[organizationName, groupName, isGlobal, status]}
      fetch={(q) =>
        isGlobal
          ? UserBackend.getGlobalUsers(q.page, q.pageSize, q.searchedColumn, q.searchText, q.sortField, q.sortOrder, status)
          : UserBackend.getUsers(
            organizationName,
            q.page,
            q.pageSize,
            q.searchedColumn,
            q.searchText,
            q.sortField,
            q.sortOrder,
            groupName,
            status,
          )
      }
      newRecord={account && canManageUsers ? () => newUser(account, organization, organizationName, groupName) : undefined}
      deleteDisabled={isProtected}
      editUrl={(r) => `/users/${r.owner}/${r.name}`}
      remove={canManageUsers ? (r) => UserBackend.deleteUser(r) : undefined}
      rowActions={(record, _index, {refresh}) => [
        record.pendingApproval && canManageUsers
          ? {
            key: "approve",
            label: i18next.t("user:Approve signup"),
            destructive: false,
            confirm: {title: i18next.t("user:Sure to approve"), description: `${record.name ?? ""}`},
            onSelect: () =>
              UserBackend.getUser(record.owner, record.name)
                .then((res: any) => {
                  if (res.status !== "ok" || !res.data) {
                    throw new Error(res.msg);
                  }
                  return UserBackend.updateUser(record.owner, record.name, {...res.data, isForbidden: false});
                })
                .then((res: any) => {
                  if (res.status === "ok") {
                    Setting.showMessage("success", i18next.t("user:Successfully approved"));
                    refresh();
                  } else {
                    Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
                  }
                })
                .catch((error) => Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${error.message ?? error}`)),
          }
          : null,
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
        groupName && canManageUsers
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
      actionColumnWidth={groupName ? 440 : 360}
      // the tree page has one route per group, so pin the column choices to the page
      tableId={embedded ? "/trees" : undefined}
    />
  );
}
