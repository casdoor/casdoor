import i18next from "i18next";
import {Link} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {XlsxImport} from "@/components/crud/XlsxImport";
import {dateColumn, linkColumn, organizationColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as GroupBackend from "@/backend/GroupBackend";
import * as Setting from "@/lib/setting";
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
    textColumn({
      dataIndex: "type",
      title: i18next.t("general:Type"),
      width: 110,
      filters: [
        {value: "Virtual", label: i18next.t("group:Virtual")},
        {value: "Physical", label: i18next.t("group:Physical")},
      ],
    }),
    textColumn({dataIndex: "parentId", title: i18next.t("group:Parent group"), width: 150, searchable: true}),
    {
      dataIndex: "users",
      sortable: true,
      searchable: true,
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
      toolbar={({refresh}) => (
        <>
          <Button variant="outline" asChild>
            <Link to={`/trees/${organizationName}`}>{i18next.t("general:Groups")}</Link>
          </Button>
          <XlsxImport
            columns={Setting.getGroupColumns()}
            templateName="import-group.xlsx"
            uploadApi="upload-groups"
            successMessage="Groups uploaded successfully, refreshing the page"
            onUploaded={refresh}
          />
        </>
      )}
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
      deleteDisabled={(r) =>
        r.haveChildren &&
        i18next.t(
          "group:You need to delete all subgroups first. You can view the subgroups in the left group tree of the [Organizations] -> [Groups] page",
        )
      }
    />
  );
}
