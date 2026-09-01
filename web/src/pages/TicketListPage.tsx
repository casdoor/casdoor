import i18next from "i18next";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, linkColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {enumColumn, TICKET_STATES} from "@/lib/enum-labels";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as TicketBackend from "@/backend/TicketBackend";
import {newTicket} from "@/pages/defaults";

export default function TicketListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/tickets/${r.owner}/${r.name}`, width: 170}),
    dateColumn(),
    dateColumn("updatedTime", i18next.t("general:Updated time")),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 170}),
    textColumn({dataIndex: "title", title: i18next.t("general:Title"), searchable: true, width: 240}),
    textColumn({dataIndex: "user", title: i18next.t("general:User"), width: 130, searchable: true, link: (v) => `/users/${v}`}),
    enumColumn({dataIndex: "state", title: i18next.t("general:State"), map: TICKET_STATES}),
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Tickets")}
      columns={columns}
      deps={[organizationName]}
      fetch={(q) =>
        TicketBackend.getTickets(
          organizationName,
          q.page,
          q.pageSize,
          q.searchedColumn,
          q.searchText,
          q.sortField,
          q.sortOrder,
        )
      }
      newRecord={account ? () => newTicket(account) : undefined}
      editUrl={(r) => `/tickets/${r.owner}/${r.name}`}
      remove={(r) => TicketBackend.deleteTicket(r)}
    />
  );
}
