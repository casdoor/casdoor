import i18next from "i18next";
import {Link} from "react-router-dom";
import {Badge} from "@/components/ui/badge";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, linkColumn, organizationColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as TransactionBackend from "@/backend/TransactionBackend";
import * as Setting from "@/lib/setting";
import {newTransaction} from "@/pages/defaults";

export default function TransactionListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/transactions/${r.owner}/${r.name}`, width: 190}),
    organizationColumn(),
    dateColumn(),
    textColumn({dataIndex: "application", title: i18next.t("general:Application"), width: 150}),
    textColumn({dataIndex: "provider", title: i18next.t("general:Provider"), width: 160}),
    textColumn({dataIndex: "type", title: i18next.t("general:Type"), width: 120}),
    textColumn({dataIndex: "tag", title: i18next.t("general:Tag"), width: 130}),
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
      dataIndex: "amount",
      title: i18next.t("product:Amount"),
      width: 130,
      sortable: true,
      render: (value, record) => Setting.getPriceDisplay(value, record.currency),
    },
    {
      dataIndex: "state",
      title: i18next.t("general:State"),
      width: 110,
      sortable: true,
      render: (value) => <Badge variant={value === "Paid" ? "success" : "secondary"}>{value}</Badge>,
    },
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Transactions")}
      columns={columns}
      deps={[organizationName]}
      fetch={(q) =>
        TransactionBackend.getTransactions(
          organizationName,
          q.page,
          q.pageSize,
          q.searchedColumn,
          q.searchText,
          q.sortField,
          q.sortOrder,
        )
      }
      newRecord={account ? () => newTransaction(account) : undefined}
      editUrl={(r) => `/transactions/${r.owner}/${r.name}`}
      remove={(r) => TransactionBackend.deleteTransaction(r)}
    />
  );
}
