import {useParams} from "react-router-dom";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useGroupOptions, useOrganizationOptions, useRoleOptions, useUserOptions} from "@/hooks/use-options";
import * as RoleBackend from "@/backend/RoleBackend";
import * as Setting from "@/lib/setting";

export default function RoleEditPage() {
  const {organizationName = "", roleName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();
  const users = useUserOptions(organizationName);
  const groups = useGroupOptions(organizationName);
  const roles = useRoleOptions(organizationName, `${organizationName}/${roleName}`);

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
    {type: "switch", name: "isEnabled", labelKey: "general:Is enabled"},
  ];

  return (
    <SimpleEditPage
      titleKey="role:Edit Role"
      backTo="/roles"
      deps={[organizationName, roleName]}
      fields={fields}
      fetch={() => RoleBackend.getRole(organizationName, roleName)}
      add={(record) => RoleBackend.addRole(record)}
      update={(record) => RoleBackend.updateRole(organizationName, roleName, record)}
      editUrl={(record) => `/roles/${record.owner}/${record.name}`}
    />
  );
}
