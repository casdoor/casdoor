import i18next from "i18next";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, linkColumn, organizationColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as RuleBackend from "@/backend/RuleBackend";
import {newRule} from "@/pages/defaults";

export default function RuleListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();

  const columns: ColumnDef<any>[] = [
    organizationColumn(150, "owner", i18next.t("general:Owner")),
    linkColumn({dataIndex: "name", to: (r) => `/rules/${r.owner}/${r.name}`}),
    dateColumn("createdTime", i18next.t("general:Create time")),
    dateColumn("updatedTime", i18next.t("general:Update time")),
    textColumn({dataIndex: "type", title: i18next.t("general:Type"), width: 130}),
    {
      dataIndex: "expressions",
      title: i18next.t("rule:Expressions"),
      render: (value: any[]) => (value ? `${value.length}` : "0"),
    },
    textColumn({dataIndex: "action", title: i18next.t("general:Action"), width: 110}),
    textColumn({dataIndex: "statusCode", title: i18next.t("rule:Status code"), width: 120}),
    textColumn({dataIndex: "reason", title: i18next.t("rule:Reason"), width: 220}),
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Rules")}
      columns={columns}
      deps={[organizationName]}
      // the rules API only supports sorting, not per-column filtering
      fetch={(q) => RuleBackend.getRules(organizationName, q.page, q.pageSize, q.sortField, q.sortOrder)}
      newRecord={account ? () => newRule(account) : undefined}
      editUrl={(r) => `/rules/${r.owner}/${r.name}`}
      remove={(r) => RuleBackend.deleteRule(r)}
    />
  );
}
