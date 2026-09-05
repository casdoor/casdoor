import * as React from "react";
import i18next from "i18next";
import {useParams} from "react-router-dom";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {
  useApplicationOptions,
  useGroupOptions,
  useOrganizationOptions,
  useRoleOptions,
  useUserOptions,
} from "@/hooks/use-options";
import * as AdapterBackend from "@/backend/AdapterBackend";
import * as ModelBackend from "@/backend/ModelBackend";
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
  const applications = useApplicationOptions(organizationName);
  // the backend resolves both through GetModel()/GetAdapter(), which take "owner/name"
  const [models, setModels] = React.useState<any[]>([]);
  const [adapters, setAdapters] = React.useState<any[]>([]);

  React.useEffect(() => {
    if (!organizationName) {
      return;
    }
    ModelBackend.getModels(organizationName, 1, 1000).then((res: any) => {
      if (res.status === "ok") {
        setModels(res.data ?? []);
      }
    });
    AdapterBackend.getAdapters(organizationName, 1, 1000).then((res: any) => {
      if (res.status === "ok") {
        setAdapters(res.data ?? []);
      }
    });
  }, [organizationName]);

  const toIdOptions = (items: any[]) =>
    items.map((item) => ({value: `${item.owner}/${item.name}`, label: `${item.owner}/${item.name}`}));

  /** the Casbin model decides whether roles and domains mean anything for this permission */
  const modelTextOf = (modelId: string) =>
    models.find((item) => `${item.owner}/${item.name}` === modelId)?.modelText ?? "";
  const hasRoleDefinition = (modelId: string) => modelTextOf(modelId).includes("role_definition");
  const hasDomainDefinition = (modelId: string) => {
    const match = modelTextOf(modelId).match(/request_definition\s*\]\s*r\s*=\s*([^\r\n]+)/);
    return match ? match[1].split(",").map((token: string) => token.trim()).includes("dom") : false;
  };

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
    {
      type: "multiselect",
      name: "roles",
      labelKey: "role:Sub roles",
      options: () => roles,
      disabled: (ctx) => !hasRoleDefinition(ctx.record.model),
    },
    {
      type: "tags",
      name: "domains",
      labelKey: "role:Sub domains",
      disabled: (ctx) => !hasDomainDefinition(ctx.record.model),
    },
    {type: "select", name: "model", labelKey: "general:Model", options: () => toIdOptions(models)},
    {type: "select", name: "adapter", labelKey: "general:Adapter", options: () => toIdOptions(adapters)},
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
      // antd only lets a "Custom" permission invent its own resource names
      creatable: (ctx) => ctx.record.resourceType === "Custom",
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
      creatable: (ctx) => ctx.record.resourceType === "Custom",
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
