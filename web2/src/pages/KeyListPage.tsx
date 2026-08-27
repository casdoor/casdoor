import i18next from "i18next";
import {Badge} from "@/components/ui/badge";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, linkColumn, organizationColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as KeyBackend from "@/backend/KeyBackend";
import * as Setting from "@/lib/setting";
import {newKey} from "@/pages/defaults";

export default function KeyListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();
  const isGlobal = account ? Setting.isDefaultOrganizationSelected(account) : false;

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/keys/${r.owner}/${r.name}`}),
    organizationColumn(),
    dateColumn(),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 170}),
    textColumn({dataIndex: "type", title: i18next.t("general:Type"), width: 130}),
    {
      dataIndex: "accessKey",
      title: i18next.t("general:Access key"),
      width: 200,
      render: (value) => (value ? Setting.getClickable(Setting.getShortText(value)) : null),
    },
    {
      dataIndex: "expireTime",
      title: i18next.t("general:Expire time"),
      width: 165,
      sortable: true,
      render: (value) => Setting.getFormattedDate(value),
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
      title={i18next.t("general:Keys")}
      columns={columns}
      deps={[organizationName, isGlobal]}
      fetch={(q) =>
        isGlobal
          ? KeyBackend.getGlobalKeys(q.page, q.pageSize, q.searchedColumn, q.searchText, q.sortField, q.sortOrder)
          : KeyBackend.getKeys(organizationName, q.page, q.pageSize, q.searchedColumn, q.searchText, q.sortField, q.sortOrder)
      }
      newRecord={account ? () => newKey(account) : undefined}
      editUrl={(r) => `/keys/${r.owner}/${r.name}`}
      remove={(r) => KeyBackend.deleteKey(r)}
    />
  );
}
