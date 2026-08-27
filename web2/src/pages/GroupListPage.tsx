import i18next from "i18next";
import {Link} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, linkColumn, organizationColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as GroupBackend from "@/backend/GroupBackend";
import {newGroup} from "@/pages/defaults";

export default function GroupListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/groups/${r.owner}/${r.name}`}),
    organizationColumn(),
    dateColumn(),
    dateColumn("updatedTime", i18next.t("general:Updated time")),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 170}),
    textColumn({dataIndex: "type", title: i18next.t("general:Type"), width: 110}),
    textColumn({dataIndex: "parentId", title: i18next.t("group:Parent group"), width: 150}),
    {
      dataIndex: "users",
      title: i18next.t("general:Users"),
      width: 110,
      align: "center",
      render: (_value, record) => (
        <Button variant="outline" size="sm" asChild>
          <Link to={`/organizations/${record.owner}/users?groupName=${record.name}`}>
            {i18next.t("general:Users")}
          </Link>
        </Button>
      ),
    },
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Groups")}
      columns={columns}
      deps={[organizationName]}
      toolbar={
        <Button variant="outline" asChild>
          <Link to={`/trees/${organizationName}`}>{i18next.t("group:Tree")}</Link>
        </Button>
      }
      fetch={(q) =>
        GroupBackend.getGroups(
          organizationName,
          false,
          q.page,
          q.pageSize,
          q.searchedColumn,
          q.searchText,
          q.sortField,
          q.sortOrder,
        )
      }
      newRecord={account ? () => newGroup(account) : undefined}
      editUrl={(r) => `/groups/${r.owner}/${r.name}`}
      remove={(r) => GroupBackend.deleteGroup(r)}
    />
  );
}
