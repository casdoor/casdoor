import i18next from "i18next";
import {Link} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, linkColumn, organizationColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as EntryBackend from "@/backend/EntryBackend";
import {newEntry} from "@/pages/defaults";

export default function EntryListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();

  const columns: ColumnDef<any>[] = [
    organizationColumn(),
    linkColumn({dataIndex: "name", to: (r) => `/entries/${r.owner}/${r.name}`, width: 180}),
    dateColumn(),
    textColumn({dataIndex: "provider", title: i18next.t("general:Provider"), width: 150, searchable: true}),
    textColumn({dataIndex: "type", title: i18next.t("general:Type"), width: 120}),
    textColumn({dataIndex: "clientIp", title: i18next.t("general:Client IP"), width: 140}),
    textColumn({dataIndex: "userAgent", title: i18next.t("general:User agent"), width: 200}),
    textColumn({dataIndex: "message", title: i18next.t("payment:Message"), width: 220}),
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Entries")}
      columns={columns}
      deps={[organizationName]}
      fetch={(q) =>
        EntryBackend.getEntries(organizationName, q.page, q.pageSize, q.searchedColumn, q.searchText, q.sortField, q.sortOrder)
      }
      newRecord={account ? () => newEntry(account) : undefined}
      editUrl={(r) => `/entries/${r.owner}/${r.name}`}
      remove={(r) => EntryBackend.deleteEntry(r)}
      rowActions={(record) => (
        <Button variant="ghost" size="sm" asChild>
          <Link to={`/entries/${record.owner}/${record.name}/transcript`}>{i18next.t("entry:Transcript")}</Link>
        </Button>
      )}
      actionColumnWidth={260}
    />
  );
}
