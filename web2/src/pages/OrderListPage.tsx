import i18next from "i18next";
import {Link} from "react-router-dom";
import {Badge} from "@/components/ui/badge";
import {Button} from "@/components/ui/button";
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
      render: (value, record) => Setting.getPriceDisplay(value, record.currency),
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
      editUrl={(r) => `/orders/${r.owner}/${r.name}`}
      remove={(r) => OrderBackend.deleteOrder(r)}
      rowActions={(record) =>
        record.state === "Created" ? (
          <Button variant="ghost" size="sm" asChild>
            <Link to={`/orders/${record.owner}/${record.name}/pay`}>{i18next.t("order:Pay")}</Link>
          </Button>
        ) : null
      }
      actionColumnWidth={240}
    />
  );
}
