import * as React from "react";
import i18next from "i18next";
import {Link} from "react-router-dom";
import {Switch} from "@/components/ui/switch";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, linkColumn, organizationColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {AUTO_RENEW_PAYMENT_TYPES, enumColumn, EnumBadge, RENEWAL_STATES, SUBSCRIPTION_STATES} from "@/lib/enum-labels";
import {useAccount} from "@/hooks/use-account";
import {useOrganizationFilter} from "@/hooks/use-organization";
import * as SubscriptionBackend from "@/backend/SubscriptionBackend";
import * as Setting from "@/lib/setting";
import {newSubscription} from "@/pages/defaults";

/** only a subscription paid with a provider that can charge again shows the switch */
function AutoRenewCell({record, editable}: {record: any; editable: boolean}) {
  const [checked, setChecked] = React.useState(Boolean(record.autoRenew));
  const [saving, setSaving] = React.useState(false);

  if (!AUTO_RENEW_PAYMENT_TYPES.includes(record.paymentType)) {
    return null;
  }

  const onChange = (value: boolean) => {
    setSaving(true);
    SubscriptionBackend.setSubscriptionAutoRenew(record.owner, record.name, value)
      .then((res: any) => {
        if (res.status === "ok") {
          setChecked(value);
          Setting.showMessage("success", i18next.t("general:Successfully saved"));
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
        }
      })
      .catch((error: any) => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      })
      .finally(() => setSaving(false));
  };

  const canToggle = editable && ["Active", "Upcoming"].includes(record.state);
  return (
    <span className="inline-flex items-center gap-1.5">
      <Switch checked={checked} disabled={!canToggle || saving} onCheckedChange={onChange} />
      {record.renewalState ? (
        <span title={record.renewalMessage || undefined}>
          <EnumBadge map={RENEWAL_STATES} value={record.renewalState} />
        </span>
      ) : null}
    </span>
  );
}

export default function SubscriptionListPage() {
  const {account} = useAccount();
  const organizationName = useOrganizationFilter();
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
      dataIndex: "group",
      title: i18next.t("subscription:Group"),
      width: 130,
      searchable: true,
      render: (value, record) =>
        value ? (
          <Link to={`/groups/${record.owner}/${value}`} className="underline-offset-4 hover:underline">
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
    {
      dataIndex: "autoRenew",
      title: i18next.t("subscription:Auto renew"),
      width: 150,
      defaultHidden: false,
      render: (_, record) => (
        <AutoRenewCell
          key={`${record.owner}/${record.name}/${record.autoRenew}`}
          record={record}
          editable={!readOnly || (account?.owner === record.owner && account?.name === record.user)}
        />
      ),
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
      readOnly={readOnly}
      editUrl={(r) => `/subscriptions/${r.owner}/${r.name}`}
      remove={(r) => SubscriptionBackend.deleteSubscription(r)}
    />
  );
}
