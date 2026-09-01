import i18next from "i18next";
import {Link} from "react-router-dom";
import {Badge} from "@/components/ui/badge";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, linkColumn, organizationColumn, tagsColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as OrderBackend from "@/backend/OrderBackend";
import * as Setting from "@/lib/setting";
import {newOrder} from "@/pages/defaults";

const stateVariant = (state: string) =>
  state === "Paid" ? "success" : state === "Created" ? "warning" : "secondary";

export default function OrderListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();
  // the antd list pages let a non-admin look but not touch these
  const readOnly = !Setting.isLocalAdminUser(account);

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/orders/${r.owner}/${r.name}`, width: 180}),
    organizationColumn(),
    dateColumn(),
    tagsColumn({dataIndex: "products", title: i18next.t("general:Products"), width: 220}),
    {
      dataIndex: "price",
      title: i18next.t("order:Price"),
      width: 120,
      sortable: true,
      // a paid order links to the payment that settled it
      render: (value, record) => {
        const price = Setting.getPriceDisplay(value, record.currency);
        return record.payment ? (
          <Link to={`/payments/${record.owner}/${record.payment}`} className="underline-offset-4 hover:underline">
            {price}
          </Link>
        ) : (
          price
        );
      },
    },
    {
      dataIndex: "user",
      title: i18next.t("general:User"),
      width: 130,
      sortable: true,
      searchable: true,
      render: (value, record) =>
        value ? (
          <Link to={`/users/${record.owner}/${value}`} className="underline-offset-4 hover:underline">
            {value}
          </Link>
        ) : null,
    },
    {
      dataIndex: "state",
      searchable: true,
      title: i18next.t("general:State"),
      width: 110,
      sortable: true,
      render: (value) => <Badge variant={stateVariant(value) as any}>{value}</Badge>,
    },
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Orders")}
      columns={columns}
      deps={[organizationName]}
      fetch={(q) =>
        OrderBackend.getOrders(organizationName, q.page, q.pageSize, q.searchedColumn, q.searchText, q.sortField, q.sortOrder)
      }
      newRecord={account ? () => newOrder(account) : undefined}
      readOnly={readOnly}
      editUrl={(r) => `/orders/${r.owner}/${r.name}`}
      remove={(r) => OrderBackend.deleteOrder(r)}
      rowActions={(record, _index, {refresh}) => [
        // the same page pays an unpaid order and shows a paid one
        {
          key: "pay",
          label: record.state === "Created" || record.state === "Failed"
            ? i18next.t("order:Pay")
            : i18next.t("general:Detail"),
          href: `/orders/${record.owner}/${record.name}/pay`,
        },
        // only an admin may cancel, and only an order nobody has paid for yet
        record.state === "Created" && Setting.isLocalAdminUser(account)
          ? {
            key: "cancel",
            label: i18next.t("general:Cancel"),
            destructive: true,
            confirm: {description: `${record.name ?? ""}`},
            onSelect: () =>
              OrderBackend.cancelOrder(record.owner, record.name)
                .then((res: any) => {
                  if (res.status === "ok") {
                    Setting.showMessage("success", i18next.t("general:Successfully canceled"));
                    refresh();
                  } else {
                    Setting.showMessage("error", `${i18next.t("general:Failed to cancel")}: ${res.msg}`);
                  }
                })
                .catch((error) =>
                  Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`),
                ),
          }
          : null,
      ]}
      actionColumnWidth={300}
    />
  );
}
