import i18next from "i18next";
import {Link} from "react-router-dom";
import {Badge} from "@/components/ui/badge";
import {Button} from "@/components/ui/button";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, linkColumn, organizationColumn, tagsColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as PaymentBackend from "@/backend/PaymentBackend";
import * as Setting from "@/lib/setting";
import {newPayment} from "@/pages/defaults";

const stateVariant = (state: string) =>
  state === "Paid" ? "success" : state === "Created" ? "warning" : state === "Error" ? "destructive" : "secondary";

export default function PaymentListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();
  // the antd list pages let a non-admin look but not touch these
  const readOnly = !Setting.isLocalAdminUser(account);

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/payments/${r.owner}/${r.name}`, width: 180}),
    organizationColumn(),
    textColumn({dataIndex: "provider", title: i18next.t("general:Provider"), width: 170, searchable: true, link: (v, r: any) => `/providers/${r.owner}/${v}`}),
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
    dateColumn(),
    {
      ...textColumn({
        dataIndex: "type",
        title: i18next.t("general:Type"),
        width: 140,
        filters: Setting.getProviderTypeOptions("Payment").map((option: any) => ({value: option.name, label: option.id})),
      }),
      // antd shows the payment provider's logo here; the type stays next to it
      render: (value: string, record: any) =>
        value ? (
          <span className="flex items-center gap-2">
            {Setting.getProviderLogo({...record, category: "Payment"})}
            <span>{value}</span>
          </span>
        ) : null,
    },
    tagsColumn({dataIndex: "products", title: i18next.t("general:Products"), width: 200}),
    {
      dataIndex: "price",
      searchable: true,
      title: i18next.t("order:Price"),
      width: 120,
      sortable: true,
      render: (value, record) => Setting.getPriceDisplay(value, record.currency),
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
      title={i18next.t("general:Payments")}
      columns={columns}
      deps={[organizationName]}
      fetch={(q) =>
        PaymentBackend.getPayments(organizationName, q.page, q.pageSize, q.searchedColumn, q.searchText, q.sortField, q.sortOrder)
      }
      newRecord={account ? () => newPayment(account) : undefined}
      readOnly={readOnly}
      actionColumnWidth={240}
      rowActions={(record) => (
        <Button variant="outline" size="sm" asChild>
          <Link to={`/payments/${record.owner}/${record.name}/result`}>{i18next.t("payment:Result")}</Link>
        </Button>
      )}
      editUrl={(r) => `/payments/${r.owner}/${r.name}`}
      remove={(r) => PaymentBackend.deletePayment(r)}
    />
  );
}
