import i18next from "i18next";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, linkColumn, organizationColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as ModelBackend from "@/backend/ModelBackend";
import {newModel} from "@/pages/defaults";

export default function ModelListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/models/${r.owner}/${r.name}`}),
    organizationColumn(),
    dateColumn(),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 180}),
    {
      dataIndex: "modelText",
      sortable: true,
      title: i18next.t("model:Model text"),
      render: (value) => (
        <pre className="max-h-24 max-w-xl overflow-auto rounded-md bg-muted p-2 text-xs">{value}</pre>
      ),
    },
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Models")}
      columns={columns}
      deps={[organizationName]}
      fetch={(q) =>
        ModelBackend.getModels(organizationName, q.page, q.pageSize, q.searchedColumn, q.searchText, q.sortField, q.sortOrder)
      }
      newRecord={account ? () => newModel(account) : undefined}
      editUrl={(r) => `/models/${r.owner}/${r.name}`}
      remove={(r) => ModelBackend.deleteModel(r)}
    />
  );
}
