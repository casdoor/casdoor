import {useParams} from "react-router-dom";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useApplicationOptions, useOrganizationOptions, useUserNameOptions} from "@/hooks/use-options";
import * as KeyBackend from "@/backend/KeyBackend";
import * as Setting from "@/lib/setting";

const TYPES = ["Organization", "Application", "User"];
const STATES = ["Active", "Suspended"];

export default function KeyEditPage() {
  const {organizationName = "", keyName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();
  const applications = useApplicationOptions(organizationName);
  const users = useUserNameOptions(organizationName);

  const fields: EditField[] = [
    {
      type: "select",
      name: "owner",
      labelKey: "general:Organization",
      options: () => organizations,
      disabled: () => !Setting.isAdminUser(account),
      // the key stores the organization twice, and the backend reads `organization`
      onChange: (value, _ctx, updateFields) => updateFields({owner: value, organization: value}),
    },
    {type: "text", name: "name", labelKey: "general:Name"},
    {type: "text", name: "displayName", labelKey: "general:Display name"},
    {
      type: "select",
      name: "type",
      labelKey: "general:Type",
      options: () => TYPES.map((item) => ({value: item, label: item})),
    },
    {
      type: "select",
      name: "application",
      labelKey: "general:Application",
      when: (ctx) => ctx.record.type === "Application",
      options: () => applications,
    },
    {
      type: "select",
      name: "user",
      labelKey: "general:User",
      when: (ctx) => ctx.record.type === "User",
      options: () => users,
    },
    {type: "text", name: "accessKey", labelKey: "general:Access key", disabled: () => true},
    {type: "text", name: "accessSecret", labelKey: "cert:Access secret", disabled: () => true},
    {type: "text", name: "expireTime", labelKey: "general:Expire time"},
    {
      type: "select",
      name: "state",
      labelKey: "general:State",
      options: () => STATES.map((item) => ({value: item, label: item})),
    },
  ];

  return (
    <SimpleEditPage
      titleKey="key:Edit Key"
      backTo="/keys"
      deps={[organizationName, keyName]}
      fields={fields}
      fetch={() => KeyBackend.getKey(organizationName, keyName)}
      add={(record) => KeyBackend.addKey(record)}
      update={(record) => KeyBackend.updateKey(organizationName, keyName, record)}
      editUrl={(record) => `/keys/${record.owner}/${record.name}`}
    />
  );
}
