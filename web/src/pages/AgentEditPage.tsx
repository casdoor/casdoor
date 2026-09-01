import {useParams} from "react-router-dom";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useApplicationOptions, useOrganizationOptions} from "@/hooks/use-options";
import * as AgentBackend from "@/backend/AgentBackend";
import * as Setting from "@/lib/setting";

export default function AgentEditPage() {
  const {organizationName = "", agentName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();
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
    {type: "text", name: "url", labelKey: "general:Listening URL"},
    {type: "text", name: "token", labelKey: "token:Access token"},
    {type: "select", name: "application", labelKey: "general:Application", options: () => applications},
  ];

  return (
    <SimpleEditPage
      titleKey="agent:Edit Agent"
      backTo="/agents"
      deps={[organizationName, agentName]}
      fields={fields}
      fetch={() => AgentBackend.getAgent(organizationName, agentName)}
      add={(record) => AgentBackend.addAgent(record)}
      update={(record) => AgentBackend.updateAgent(organizationName, agentName, record)}
      editUrl={(record) => `/agents/${record.owner}/${record.name}`}
    />
  );
}
