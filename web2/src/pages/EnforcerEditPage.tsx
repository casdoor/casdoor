import {useParams} from "react-router-dom";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useAdapterOptions, useModelOptions, useOrganizationOptions} from "@/hooks/use-options";
import * as EnforcerBackend from "@/backend/EnforcerBackend";
import * as Setting from "@/lib/setting";

export default function EnforcerEditPage() {
  const {organizationName = "", enforcerName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();
  const models = useModelOptions(organizationName);
  const adapters = useAdapterOptions(organizationName);

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
    {type: "select", name: "model", labelKey: "general:Model", options: () => models},
    {type: "select", name: "adapter", labelKey: "general:Adapter", options: () => adapters},
  ];

  return (
    <SimpleEditPage
      titleKey="enforcer:Edit Enforcer"
      backTo="/enforcers"
      deps={[organizationName, enforcerName]}
      fields={fields}
      fetch={() => EnforcerBackend.getEnforcer(organizationName, enforcerName)}
      add={(record) => EnforcerBackend.addEnforcer(record)}
      update={(record) => EnforcerBackend.updateEnforcer(organizationName, enforcerName, record)}
      editUrl={(record) => `/enforcers/${record.owner}/${record.name}`}
    />
  );
}
