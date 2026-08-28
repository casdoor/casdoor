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
import * as Setting from "@/lib/setting";

const RESOURCE_TYPES = ["Application", "TreeNode", "Custom"];
const ACTIONS = ["Read", "Write", "Admin"];
const EFFECTS = ["Allow", "Deny"];
const STATES = ["Approved", "Pending", "Rejected"];

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
    {type: "text", name: "name", labelKey: "general:Name"},
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
      options: () => RESOURCE_TYPES.map((item) => ({value: item, label: item})),
    },
    {
      type: "multiselect",
      name: "resources",
      labelKey: "general:Resources",
      creatable: true,
      options: (ctx) => (ctx.record.resourceType === "Application" ? applications : []),
    },
    {
      type: "multiselect",
      name: "actions",
      labelKey: "permission:Actions",
      options: () => ACTIONS.map((item) => ({value: item, label: i18next.t(`permission:${item}`)})),
    },
    {
      type: "select",
      name: "effect",
      labelKey: "permission:Effect",
      options: () => EFFECTS.map((item) => ({value: item, label: i18next.t(`permission:${item}`)})),
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
      options: () => STATES.map((item) => ({value: item, label: i18next.t(`permission:${item}`)})),
      disabled: () => !Setting.isLocalAdminUser(account),
    },
  ];

  return (
    <SimpleEditPage
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
