import * as React from "react";
import i18next from "i18next";
import {Link, useNavigate} from "react-router-dom";
import {Badge} from "@/components/ui/badge";
import {Button} from "@/components/ui/button";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, linkColumn, organizationColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as TransactionBackend from "@/backend/TransactionBackend";
import * as Setting from "@/lib/setting";
import {newRechargeTransaction, newTransaction} from "@/pages/defaults";

export default function TransactionListPage() {
  const {account} = useAccount();
  const navigate = useNavigate();
  const organizationName = useRequestOrganization();
  // the antd list pages let a non-admin look but not touch these
  const readOnly = !Setting.isLocalAdminUser(account);
  const [recharging, setRecharging] = React.useState(false);

  /**
   * "Recharge" adds the transaction before the edit page opens, unlike "Add" —
   * the backend names the row and returns that name, and the edit page then
   * shows it in recharge mode.
   */
  const recharge = async(refresh: () => void) => {
    if (!account) {
      return;
    }
    const record = newRechargeTransaction(account);
    setRecharging(true);
    try {
      const res: any = await TransactionBackend.addTransaction(record);
      if (res.status === "ok") {
        Setting.showMessage("success", i18next.t("general:Successfully added"));
        navigate(`/transactions/${record.owner}/${res.data}`, {state: {recharge: true}});
      } else {
        Setting.showMessage("error", `${i18next.t("general:Failed to add")}: ${res.msg}`);
        refresh();
      }
    } catch (error: any) {
      Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
    } finally {
      setRecharging(false);
    }
  };

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
      readOnly={readOnly}
      toolbar={({refresh}) => (
        <Button disabled={readOnly} loading={recharging} onClick={() => recharge(refresh)}>
          {i18next.t("transaction:Recharge")}
        </Button>
      )}
      editUrl={(r) => `/transactions/${r.owner}/${r.name}`}
      remove={(r) => TransactionBackend.deleteTransaction(r)}
    />
  );
}
