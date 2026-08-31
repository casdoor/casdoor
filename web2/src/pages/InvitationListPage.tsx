import i18next from "i18next";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, linkColumn, organizationColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {enumColumn, INVITATION_STATES} from "@/lib/enum-labels";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as InvitationBackend from "@/backend/InvitationBackend";
import {newInvitation} from "@/pages/defaults";

export default function InvitationListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/invitations/${r.owner}/${r.name}`}),
    organizationColumn(),
    dateColumn(),
    dateColumn("updatedTime", i18next.t("general:Updated time")),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 170}),
    textColumn({dataIndex: "code", title: i18next.t("invitation:Code"), width: 150, mono: true, searchable: true}),
    textColumn({dataIndex: "quota", title: i18next.t("invitation:Quota"), width: 100, searchable: true}),
    textColumn({dataIndex: "usedCount", title: i18next.t("invitation:Used count"), width: 120, searchable: true}),
    textColumn({dataIndex: "application", title: i18next.t("general:Application"), width: 150, searchable: true}),
    textColumn({dataIndex: "email", title: i18next.t("general:Email"), width: 170, searchable: true}),
    textColumn({dataIndex: "phone", title: i18next.t("general:Phone"), width: 130, searchable: true}),
    enumColumn({dataIndex: "state", title: i18next.t("general:State"), map: INVITATION_STATES}),
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Invitations")}
      columns={columns}
      deps={[organizationName]}
      fetch={(q) =>
        InvitationBackend.getInvitations(
          organizationName,
          q.page,
          q.pageSize,
          q.searchedColumn,
          q.searchText,
          q.sortField,
          q.sortOrder,
        )
      }
      newRecord={account ? () => newInvitation(account) : undefined}
      editUrl={(r) => `/invitations/${r.owner}/${r.name}`}
      remove={(r) => InvitationBackend.deleteInvitation(r)}
    />
  );
}
