import i18next from "i18next";
import {Link} from "react-router-dom";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, linkColumn, organizationColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {enumColumn, SUBSCRIPTION_STATES} from "@/lib/enum-labels";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as SubscriptionBackend from "@/backend/SubscriptionBackend";
import * as Setting from "@/lib/setting";
import {newSubscription} from "@/pages/defaults";

export default function SubscriptionListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();
  // the antd list pages let a non-admin look but not touch these
  const readOnly = !Setting.isLocalAdminUser(account);

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/subscriptions/${r.owner}/${r.name}`}),
    organizationColumn(),
    dateColumn(),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 170}),
    textColumn({dataIndex: "period", title: i18next.t("plan:Period"), width: 120, searchable: true}),
    {
      dataIndex: "startTime",
      searchable: true,
      title: i18next.t("subscription:Start time"),
      width: 165,
      sortable: true,
      render: (value) => Setting.getFormattedDate(value),
    },
    {
      dataIndex: "endTime",
      searchable: true,
      title: i18next.t("subscription:End time"),
      width: 165,
      sortable: true,
      render: (value) => Setting.getFormattedDate(value),
    },
    {
      dataIndex: "plan",
      searchable: true,
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
      searchable: true,
      title: i18next.t("general:Payment"),
      width: 170,
      render: (value, record) =>
        value ? (
          <Link to={`/payments/${record.owner}/${value}`} className="underline-offset-4 hover:underline">
            {value}
          </Link>
        ) : null,
    },
    enumColumn({dataIndex: "state", title: i18next.t("general:State"), map: SUBSCRIPTION_STATES}),
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
      readOnly={readOnly}
      editUrl={(r) => `/subscriptions/${r.owner}/${r.name}`}
      remove={(r) => SubscriptionBackend.deleteSubscription(r)}
    />
  );
}
