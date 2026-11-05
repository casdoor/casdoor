import i18next from "i18next";
import {useParams} from "react-router-dom";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useOrganizationOptions} from "@/hooks/use-options";
import * as RadiusClientBackend from "@/backend/RadiusClientBackend";
import {enumOptions, KEY_STATES} from "@/lib/enum-labels";

export default function RadiusClientEditPage() {
  const {organizationName = "", radiusClientName = ""} = useParams();
  const organizations = useOrganizationOptions();

  const fields: EditField[] = [
    {type: "text", name: "name", labelKey: "general:Name", required: true},
    {type: "text", name: "displayName", labelKey: "general:Display name"},
    {
      type: "select",
      name: "organization",
      labelKey: "radiusClient:Organization",
      options: () => [{value: "", label: i18next.t("general:None")}, ...organizations],
    },
    {type: "text", name: "ipRanges", labelKey: "radiusClient:IP ranges", required: true},
    {type: "password", name: "secret", labelKey: "radiusClient:Shared secret", required: true},
    {
      type: "select",
      name: "state",
      labelKey: "general:State",
      options: () => enumOptions(KEY_STATES),
    },
  ];

  return (
    <SimpleEditPage
      titleKey="radiusClient:Edit RADIUS Client"
      backTo="/radius-clients"
      deps={[organizationName, radiusClientName]}
      fields={fields}
      fetch={() => RadiusClientBackend.getRadiusClient(organizationName, radiusClientName)}
      add={(record) => RadiusClientBackend.addRadiusClient(record)}
      update={(record) => RadiusClientBackend.updateRadiusClient(organizationName, radiusClientName, record)}
      editUrl={(record) => `/radius-clients/${record.owner}/${record.name}`}
    />
  );
}
