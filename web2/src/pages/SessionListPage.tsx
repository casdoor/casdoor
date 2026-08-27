import i18next from "i18next";
import {Link} from "react-router-dom";
import {Badge} from "@/components/ui/badge";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as SessionBackend from "@/backend/SessionBackend";

export default function SessionListPage() {
  const organizationName = useRequestOrganization();

  const columns: ColumnDef<any>[] = [
    {
      dataIndex: "name",
      title: i18next.t("general:Name"),
      width: 180,
      sortable: true,
      searchable: true,
      render: (value, record) => (
        <Link to={`/users/${record.owner}/${value}`} className="font-medium underline-offset-4 hover:underline">
          {value}
        </Link>
      ),
    },
    {
      dataIndex: "owner",
      title: i18next.t("general:Organization"),
      width: 150,
      sortable: true,
      searchable: true,
      render: (value) => (
        <Link to={`/organizations/${value}`} className="underline-offset-4 hover:underline">
          {value}
        </Link>
      ),
    },
    dateColumn(),
    {
      dataIndex: "sessionId",
      title: i18next.t("general:Session ID"),
      render: (value: string[]) =>
        !value || value.length === 0 ? null : (
          <div className="flex flex-wrap gap-1">
            {value.map((id) => (
              <Badge key={id} variant="secondary" className="font-mono text-[11px] font-normal">
                {id}
              </Badge>
            ))}
          </div>
        ),
    },
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Sessions")}
      columns={columns}
      deps={[organizationName]}
      fetch={(q) =>
        SessionBackend.getSessions(
          organizationName,
          q.page,
          q.pageSize,
          q.searchedColumn,
          q.searchText,
          q.sortField,
          q.sortOrder,
        )
      }
      remove={(r) => SessionBackend.deleteSession(r)}
      actionColumnWidth={120}
    />
  );
}
