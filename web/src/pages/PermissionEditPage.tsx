import i18next from "i18next";
import {useParams} from "react-router-dom";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {
  useAdapterOptions,
  useApplicationOptions,
  useGroupOptions,
  useModelOptions,
  useOrganizationOptions,
  useRoleOptions,
  useUserOptions,
} from "@/hooks/use-options";
import * as PermissionBackend from "@/backend/PermissionBackend";
import {
  enumOptions,
  PERMISSION_ACTIONS,
  PERMISSION_API_ACTIONS,
  PERMISSION_EFFECTS,
  PERMISSION_RESOURCE_TYPES,
  PERMISSION_STATES,
} from "@/lib/enum-labels";
import * as Setting from "@/lib/setting";


export default function PermissionEditPage() {
  const {organizationName = "", permissionName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();
  const users = useUserOptions(organizationName);
  const groups = useGroupOptions(organizationName);
  const roles = useRoleOptions(organizationName);
  const models = useModelOptions(organizationName);
  const adapters = useAdapterOptions(organizationName);
  const applications = useApplicationOptions(organizationName);

  const fields: EditField[] = [
    {
      type: "select",
      name: "owner",
      labelKey: "general:Organization",
      options: () => organizations,
      disabled: () => !Setting.isAdminUser(account),
    },
    {type: "text", name: "name", labelKey: "general:Name", required: true},
    {type: "text", name: "displayName", labelKey: "general:Display name"},
    {type: "text", name: "description", labelKey: "general:Description"},
    {type: "multiselect", name: "users", labelKey: "role:Sub users", options: () => users},
    {type: "multiselect", name: "groups", labelKey: "role:Sub groups", options: () => groups},
    {type: "multiselect", name: "roles", labelKey: "role:Sub roles", options: () => roles},
    {type: "tags", name: "domains", labelKey: "role:Sub domains"},
    {type: "select", name: "model", labelKey: "general:Model", options: () => models},
    {type: "select", name: "adapter", labelKey: "general:Adapter", options: () => adapters},
    {
      type: "select",
      name: "resourceType",
      labelKey: "permission:Resource type",
      options: () => enumOptions(PERMISSION_RESOURCE_TYPES),
    },
    {
      type: "multiselect",
      name: "resources",
      labelKey: "general:Resources",
      creatable: true,
      // an API permission is scoped to backend paths; everything else to applications
      options: (ctx) =>
        ctx.record.resourceType === "API"
          ? Setting.getApiPaths().map((path: string) => ({value: path, label: path}))
          : [{value: "*", label: i18next.t("general:All")}, ...applications],
    },
    {
      type: "multiselect",
      name: "actions",
      labelKey: "permission:Actions",
      options: (ctx) =>
        enumOptions(ctx.record.resourceType === "API" ? PERMISSION_API_ACTIONS : PERMISSION_ACTIONS),
    },
    {
      type: "select",
      name: "effect",
      labelKey: "permission:Effect",
      options: () => enumOptions(PERMISSION_EFFECTS),
    },
    {type: "switch", name: "isEnabled", labelKey: "general:Is enabled"},
    {type: "text", name: "submitter", labelKey: "permission:Submitter", disabled: () => true},
    {type: "text", name: "approver", labelKey: "permission:Approver", disabled: () => true},
    {type: "text", name: "approveTime", labelKey: "permission:Approve time", disabled: () => true},
    {type: "text", name: "expireTime", labelKey: "general:Expire time"},
    {
      type: "select",
      name: "state",
      labelKey: "general:State",
      options: () => enumOptions(PERMISSION_STATES),
      disabled: () => !Setting.isLocalAdminUser(account),
    },
  ];

  return (
    <SimpleEditPage
      beforeSave={(permission) => {
        // the same checks the antd page runs before it POSTs
        if ((permission.users?.length ?? 0) === 0 && (permission.roles?.length ?? 0) === 0) {
          Setting.showMessage("error", i18next.t("general:The users and roles cannot be empty at the same time"));
          return null;
        }
        if ((permission.resources?.length ?? 0) === 0) {
          Setting.showMessage("error", i18next.t("general:The resources cannot be empty"));
          return null;
        }
        if ((permission.actions?.length ?? 0) === 0) {
          Setting.showMessage("error", i18next.t("general:The actions cannot be empty"));
          return null;
        }
        if (!Setting.isLocalAdminUser(account) && permission.submitter !== account?.name) {
          Setting.showMessage("error", i18next.t("general:A normal user can only modify the permission submitted by itself"));
          return null;
        }
        return permission;
      }}
      titleKey="permission:Edit Permission"
      backTo="/permissions"
      deps={[organizationName, permissionName]}
      fields={fields}
      fetch={() => PermissionBackend.getPermission(organizationName, permissionName)}
      add={(record) => PermissionBackend.addPermission(record)}
      update={(record) => PermissionBackend.updatePermission(organizationName, permissionName, record)}
      editUrl={(record) => `/permissions/${record.owner}/${record.name}`}
    />
  );
}
