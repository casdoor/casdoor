import i18next from "i18next";
import {Link} from "react-router-dom";
import {Badge} from "@/components/ui/badge";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, linkColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as ApplicationBackend from "@/backend/ApplicationBackend";
import * as Setting from "@/lib/setting";
import {newApplication} from "@/pages/defaults";

export default function ApplicationListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();
  const isGlobal = account ? Setting.isDefaultOrganizationSelected(account) : false;

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/applications/${r.organization}/${r.name}`, width: 170}),
    dateColumn(),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 180}),
    textColumn({dataIndex: "category", title: i18next.t("general:Category"), width: 120}),
    textColumn({dataIndex: "type", title: i18next.t("general:Type"), width: 110}),
    {
      dataIndex: "logo",
      title: "Logo",
      width: 170,
      render: (value) =>
        value ? (
          <a href={value} target="_blank" rel="noreferrer">
            <img src={value} alt="logo" className="h-8 max-w-[150px] object-contain" />
          </a>
        ) : null,
    },
    {
      dataIndex: "organization",
      title: i18next.t("general:Organization"),
      width: 140,
      sortable: true,
      searchable: true,
      render: (value) => (
        <Link to={`/organizations/${value}`} className="underline-offset-4 hover:underline">
          {value}
        </Link>
      ),
    },
    {
      dataIndex: "providers",
      title: i18next.t("application:Providers"),
      render: (value: any[], record) => {
        if (!value || value.length === 0) {
          return null;
        }
        return (
          <div className="flex flex-wrap gap-1">
            {value.map((item: any) => (
              <Badge key={item.name} variant="secondary" className="font-normal">
                <Link to={`/providers/${record.organization}/${item.name}`} className="underline-offset-2 hover:underline">
                  {item.name}
                </Link>
              </Badge>
            ))}
          </div>
        );
      },
    },
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Applications")}
      columns={columns}
      deps={[organizationName, isGlobal]}
      fetch={(q) =>
        isGlobal
          ? ApplicationBackend.getApplications("admin", q.page, q.pageSize, q.searchedColumn, q.searchText, q.sortField, q.sortOrder)
          : ApplicationBackend.getApplicationsByOrganization(
            "admin",
            organizationName,
            q.page,
            q.pageSize,
            q.searchedColumn,
            q.searchText,
            q.sortField,
            q.sortOrder,
          )
      }
      newRecord={account ? () => newApplication(account) : undefined}
      editUrl={(r) => `/applications/${r.organization}/${r.name}`}
      remove={(r) => ApplicationBackend.deleteApplication(r)}
    />
  );
}
