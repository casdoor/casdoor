import {useParams} from "react-router-dom";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useOrganizationOptions} from "@/hooks/use-options";
import * as ModelBackend from "@/backend/ModelBackend";
import * as Setting from "@/lib/setting";

export default function ModelEditPage() {
  const {organizationName = "", modelName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();

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
    {type: "code", name: "modelText", labelKey: "model:Model text", height: 360},
  ];

  return (
    <SimpleEditPage
      titleKey="model:Edit Model"
      backTo="/models"
      deps={[organizationName, modelName]}
      fields={fields}
      fetch={() => ModelBackend.getModel(organizationName, modelName)}
      add={(record) => ModelBackend.addModel(record)}
      update={(record) => ModelBackend.updateModel(organizationName, modelName, record)}
      editUrl={(record) => `/models/${record.owner}/${record.name}`}
    />
  );
}
