import i18next from "i18next";
import {Link} from "react-router-dom";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, linkColumn, organizationColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as EnforcerBackend from "@/backend/EnforcerBackend";
import {newEnforcer} from "@/pages/defaults";

export default function EnforcerListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/enforcers/${r.owner}/${r.name}`}),
    organizationColumn(),
    dateColumn(),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 170}),
    {
      dataIndex: "model",
      title: i18next.t("general:Model"),
      width: 180,
      sortable: true,
      render: (value) =>
        value ? (
          <Link to={`/models/${value}`} className="underline-offset-4 hover:underline">
            {value}
          </Link>
        ) : null,
    },
    {
      dataIndex: "adapter",
      title: i18next.t("general:Adapter"),
      width: 180,
      sortable: true,
      render: (value) =>
        value ? (
          <Link to={`/adapters/${value}`} className="underline-offset-4 hover:underline">
            {value}
          </Link>
        ) : null,
    },
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Enforcers")}
      columns={columns}
      deps={[organizationName]}
      fetch={(q) =>
        EnforcerBackend.getEnforcers(organizationName, q.page, q.pageSize, q.searchedColumn, q.searchText, q.sortField, q.sortOrder)
      }
      newRecord={account ? () => newEnforcer(account) : undefined}
      editUrl={(r) => `/enforcers/${r.owner}/${r.name}`}
      remove={(r) => EnforcerBackend.deleteEnforcer(r)}
    />
  );
}
