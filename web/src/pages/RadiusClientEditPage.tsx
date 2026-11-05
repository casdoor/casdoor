import {useParams} from "react-router-dom";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useOrganizationOptions} from "@/hooks/use-options";
import * as RadiusClientBackend from "@/backend/RadiusClientBackend";
import * as Setting from "@/lib/setting";

export default function RadiusClientEditPage() {
  const {organizationName = "", radiusClientName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();

  const fields: EditField[] = [
    {
      type: "select",
      name: "owner",
      labelKey: "general:Organization",
      options: () => organizations,
      disabled: (ctx) => !Setting.isAdminUser(account) || Setting.builtInObject(ctx.record),
    },
    {type: "text", name: "name", labelKey: "general:Name", required: true},
    {type: "text", name: "displayName", labelKey: "general:Display name"},
    {
      type: "password",
      name: "secret",
      labelKey: "general:Secret",
      required: true,
    },
    {
      type: "textarea",
      name: "ipCidr",
      labelKey: "general:IP CIDR",
      rows: 3,
      placeholder: "192.168.20.50, 192.168.20.0/24",
    },
    {type: "text", name: "description", labelKey: "general:Description"},
    {
      type: "select",
      name: "state",
      labelKey: "general:State",
      options: () => [
        {label: "Active", value: "Active"},
        {label: "Inactive", value: "Inactive"},
      ],
    },
  ];

  return (
    <SimpleEditPage
      titleKey="general:Edit Radius Client"
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
