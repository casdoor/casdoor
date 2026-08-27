import i18next from "i18next";
import {Link} from "react-router-dom";
import {Badge} from "@/components/ui/badge";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, linkColumn, organizationColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as SubscriptionBackend from "@/backend/SubscriptionBackend";
import * as Setting from "@/lib/setting";
import {newSubscription} from "@/pages/defaults";

export default function SubscriptionListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/subscriptions/${r.owner}/${r.name}`}),
    organizationColumn(),
    dateColumn(),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 170}),
    textColumn({dataIndex: "period", title: i18next.t("plan:Period"), width: 120}),
    {
      dataIndex: "startTime",
      title: i18next.t("subscription:Start time"),
      width: 165,
      sortable: true,
      render: (value) => Setting.getFormattedDate(value),
    },
    {
      dataIndex: "endTime",
      title: i18next.t("subscription:End time"),
      width: 165,
      sortable: true,
      render: (value) => Setting.getFormattedDate(value),
    },
    {
      dataIndex: "plan",
      title: i18next.t("general:Plan"),
      width: 150,
      render: (value, record) =>
        value ? (
          <Link to={`/plans/${record.owner}/${value}`} className="underline-offset-4 hover:underline">
            {value}
          </Link>
        ) : null,
    },
    {
      dataIndex: "user",
      title: i18next.t("general:User"),
      width: 130,
      searchable: true,
      render: (value, record) =>
        value ? (
          <Link to={`/users/${record.owner}/${value}`} className="underline-offset-4 hover:underline">
            {value}
          </Link>
        ) : null,
    },
    {
      dataIndex: "payment",
      title: i18next.t("general:Payment"),
      width: 170,
      render: (value, record) =>
        value ? (
          <Link to={`/payments/${record.owner}/${value}`} className="underline-offset-4 hover:underline">
            {value}
          </Link>
        ) : null,
    },
    {
      dataIndex: "state",
      title: i18next.t("general:State"),
      width: 110,
      sortable: true,
      render: (value) => <Badge variant={value === "Active" ? "success" : "secondary"}>{value}</Badge>,
    },
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Subscriptions")}
      columns={columns}
      deps={[organizationName]}
      fetch={(q) =>
        SubscriptionBackend.getSubscriptions(
          organizationName,
          q.page,
          q.pageSize,
          q.searchedColumn,
          q.searchText,
          q.sortField,
          q.sortOrder,
        )
      }
      newRecord={account ? () => newSubscription(account) : undefined}
      editUrl={(r) => `/subscriptions/${r.owner}/${r.name}`}
      remove={(r) => SubscriptionBackend.deleteSubscription(r)}
    />
  );
}
