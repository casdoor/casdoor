import i18next from "i18next";
import {useParams} from "react-router-dom";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useOrganizationOptions, useUserNameOptions} from "@/hooks/use-options";
import * as TicketBackend from "@/backend/TicketBackend";
import * as Setting from "@/lib/setting";

const STATES = ["Open", "Closed"];

export default function TicketEditPage() {
  const {organizationName = "", ticketName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();
  const users = useUserNameOptions(organizationName);

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
    {type: "select", name: "user", labelKey: "general:User", options: () => users},
    {type: "text", name: "title", labelKey: "general:Title"},
    {type: "textarea", name: "content", labelKey: "ticket:Content", rows: 8},
    {
      type: "select",
      name: "state",
      labelKey: "general:State",
      options: () => STATES.map((item) => ({value: item, label: i18next.t(`ticket:${item}`)})),
    },
  ];

  return (
    <SimpleEditPage
      titleKey="ticket:Edit Ticket"
      backTo="/tickets"
      deps={[organizationName, ticketName]}
      fields={fields}
      fetch={() => TicketBackend.getTicket(organizationName, ticketName)}
      add={(record) => TicketBackend.addTicket(record)}
      update={(record) => TicketBackend.updateTicket(organizationName, ticketName, record)}
      editUrl={(record) => `/tickets/${record.owner}/${record.name}`}
    />
  );
}
