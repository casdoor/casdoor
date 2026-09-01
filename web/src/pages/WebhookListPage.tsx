import i18next from "i18next";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {boolColumn, dateColumn, linkColumn, tagsColumn, textColumn, urlColumn, valueFilters} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as WebhookBackend from "@/backend/WebhookBackend";
import {newWebhook} from "@/pages/defaults";

export default function WebhookListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/webhooks/${r.name}`, width: 180}),
    textColumn({dataIndex: "organization", title: i18next.t("general:Organization"), width: 140, searchable: true, link: (v) => `/organizations/${v}`}),
    dateColumn(),
    urlColumn({dataIndex: "url", title: i18next.t("general:URL"), width: 240}),
    textColumn({dataIndex: "method", title: i18next.t("general:Method"), width: 100, searchable: true}),
    textColumn({
      dataIndex: "contentType",
      title: i18next.t("webhook:Content type"),
      width: 150,
      filters: valueFilters(["application/json", "application/x-www-form-urlencoded"]),
    }),
    tagsColumn({dataIndex: "events", title: i18next.t("webhook:Events"), width: 220, sortable: true, searchable: true}),
    boolColumn({dataIndex: "isUserExtended", title: i18next.t("webhook:Is user extended")}),
    boolColumn({dataIndex: "singleOrgOnly", title: i18next.t("webhook:Single org only")}),
    {...boolColumn({dataIndex: "isEnabled", title: i18next.t("general:Is enabled")}), fixed: "right" as const},
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Webhooks")}
      columns={columns}
      deps={[organizationName]}
      fetch={(q) =>
        WebhookBackend.getWebhooks(
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
      newRecord={account ? () => newWebhook(account) : undefined}
      editUrl={(r) => `/webhooks/${r.name}`}
      remove={(r) => WebhookBackend.deleteWebhook(r)}
    />
  );
}
