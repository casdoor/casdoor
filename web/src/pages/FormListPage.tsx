import i18next from "i18next";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, linkColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import * as FormBackend from "@/backend/FormBackend";
import {newForm} from "@/pages/defaults";

export default function FormListPage() {
  const {account} = useAccount();

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/forms/${r.name}`, width: 200}),
    dateColumn(),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 200}),
    textColumn({dataIndex: "type", title: i18next.t("general:Type"), width: 150}),
    {
      dataIndex: "formItems",
      title: i18next.t("form:Form items"),
      render: (value: any[]) => (value ? `${value.length}` : "0"),
    },
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Forms")}
      columns={columns}
      deps={[account?.owner]}
      fetch={(q) =>
        FormBackend.getForms(
          account?.owner ?? "admin",
          q.page,
          q.pageSize,
          q.searchedColumn,
          q.searchText,
          q.sortField,
          q.sortOrder,
        )
      }
      newRecord={account ? () => newForm(account) : undefined}
      editUrl={(r) => `/forms/${r.name}`}
      remove={(r) => FormBackend.deleteForm(r)}
    />
  );
}
