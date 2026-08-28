import {useParams} from "react-router-dom";
import {CasbinEditor} from "@/components/casbin/CasbinEditor";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useOrganizationOptions} from "@/hooks/use-options";
import * as ModelBackend from "@/backend/ModelBackend";
import * as Setting from "@/lib/setting";

export default function ModelEditPage() {
  const {organizationName = "", modelName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();

  const isBuiltIn = (ctx: {record: any}) => Setting.builtInObject(ctx.record);

  const fields: EditField[] = [
    {
      type: "select",
      name: "owner",
      labelKey: "general:Organization",
      options: () => organizations,
      disabled: (ctx) => !Setting.isAdminUser(account) || Setting.builtInObject(ctx.record),
    },
    {type: "text", name: "name", labelKey: "general:Name", disabled: isBuiltIn},
    {type: "text", name: "displayName", labelKey: "general:Display name"},
    {type: "text", name: "description", labelKey: "general:Description"},
    {
      type: "custom",
      name: "modelText",
      labelKey: "model:Model text",
      block: true,
      render: (ctx, update) => (
        <CasbinEditor model={ctx.record} onModelTextChange={(value) => update("modelText", value)} />
      ),
    },
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
