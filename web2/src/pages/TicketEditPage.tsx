import i18next from "i18next";
import {useParams} from "react-router-dom";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {TicketMessages} from "@/components/user/TicketMessages";
import {useAccount} from "@/hooks/use-account";
import {useOrganizationOptions, useUserNameOptions} from "@/hooks/use-options";
import * as TicketBackend from "@/backend/TicketBackend";
import * as Setting from "@/lib/setting";

const STATES = ["Open", "In Progress", "Resolved", "Closed"];

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
    {type: "textarea", name: "content", labelKey: "provider:Content", rows: 8},
    {
      type: "select",
      name: "state",
      labelKey: "general:State",
      options: () => STATES.map((item) => ({value: item, label: i18next.t(`ticket:${item}`)})),
      // once a ticket is closed only an admin can move it back
      disabled: (ctx) => !Setting.isAdminUser(account) && ctx.record.state === "Closed",
    },
    {
      type: "custom",
      name: "messages",
      labelKey: "ticket:Messages",
      block: true,
      // messages are appended to the saved ticket, so not while it is being created
      when: (ctx) => ctx.mode !== "add",
      render: (ctx) => (
        <TicketMessages
          organizationName={organizationName}
          ticketName={ticketName}
          messages={ctx.record.messages ?? []}
          account={account}
          onSent={ctx.reload}
        />
      ),
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
