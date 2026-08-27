import i18next from "i18next";
import {Link} from "react-router-dom";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {boolColumn, dateColumn, linkColumn, organizationColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as PlanBackend from "@/backend/PlanBackend";
import * as Setting from "@/lib/setting";
import {newPlan} from "@/pages/defaults";

export default function PlanListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/plans/${r.owner}/${r.name}`}),
    organizationColumn(),
    dateColumn(),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 170}),
    {
      dataIndex: "price",
      title: i18next.t("order:Price"),
      width: 120,
      sortable: true,
      render: (value, record) => Setting.getPriceDisplay(value, record.currency),
    },
    textColumn({dataIndex: "period", title: i18next.t("plan:Period"), width: 120}),
    {
      dataIndex: "role",
      title: i18next.t("general:Role"),
      width: 150,
      sortable: true,
      render: (value, record) =>
        value ? (
          <Link to={`/roles/${record.owner}/${value}`} className="underline-offset-4 hover:underline">
            {value}
          </Link>
        ) : null,
    },
    {
      dataIndex: "product",
      title: i18next.t("plan:Related product"),
      width: 170,
      render: (value, record) =>
        value ? (
          <Link to={`/products/${record.owner}/${value}`} className="underline-offset-4 hover:underline">
            {value}
          </Link>
        ) : null,
    },
    boolColumn({dataIndex: "isEnabled", title: i18next.t("general:Is enabled")}),
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Plans")}
      columns={columns}
      deps={[organizationName]}
      fetch={(q) =>
        PlanBackend.getPlans(organizationName, q.page, q.pageSize, q.searchedColumn, q.searchText, q.sortField, q.sortOrder)
      }
      newRecord={account ? () => newPlan(account) : undefined}
      editUrl={(r) => `/plans/${r.owner}/${r.name}`}
      remove={(r) => PlanBackend.deletePlan(r)}
    />
  );
}
