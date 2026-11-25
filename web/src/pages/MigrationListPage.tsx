import i18next from "i18next";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, linkColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useOrganizationFilter} from "@/hooks/use-organization";
import * as MigrationBackend from "@/backend/MigrationBackend";
import {newMigration} from "@/pages/defaults";

export default function MigrationListPage() {
  const {account} = useAccount();
  const organizationName = useOrganizationFilter();

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/migrations/${r.organization}/${r.name}`, width: 170}),
    textColumn({dataIndex: "organization", title: i18next.t("general:Organization"), width: 140, searchable: true, link: (v) => `/organizations/${v}`}),
    dateColumn(),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), width: 180, searchable: true}),
    textColumn({dataIndex: "plugin", title: i18next.t("general:Action"), width: 180, searchable: true}),
    textColumn({dataIndex: "host", title: i18next.t("general:Host"), width: 200, searchable: true}),
    textColumn({dataIndex: "state", title: i18next.t("general:State"), width: 120}),
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Migrations")}
      description={i18next.t("migration:Migrations - Tooltip")}
      columns={columns}
      deps={[organizationName]}
      fetch={(q) =>
        MigrationBackend.getMigrations(
          "admin",
          organizationName,
          q.page,
          q.pageSize,
          q.searchedColumn,
          q.searchText,
          q.sortField,
          q.sortOrder,
        )
      }
      newRecord={account ? () => newMigration(account) : undefined}
      add={(r) => MigrationBackend.addMigration(r)}
      editUrl={(r) => `/migrations/${r.organization}/${r.name}`}
      remove={(r) => MigrationBackend.deleteMigration(r)}
    />
  );
}
