import i18next from "i18next";
import {Link} from "react-router-dom";
import {Badge} from "@/components/ui/badge";
import {Button} from "@/components/ui/button";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as WebhookEventBackend from "@/backend/WebhookEventBackend";
import * as Setting from "@/lib/setting";

const stateVariant = (state: string) =>
  state === "Success" ? "success" : state === "Pending" ? "warning" : "destructive";

export default function WebhookEventListPage() {
  const organizationName = useRequestOrganization();

  const columns: ColumnDef<any>[] = [
    textColumn({dataIndex: "name", title: i18next.t("general:Name"), width: 180, searchable: true, mono: true}),
    {
      dataIndex: "webhook",
      title: i18next.t("general:Webhook"),
      width: 180,
      sortable: true,
      searchable: true,
      render: (value) => (
        <Link to={`/webhooks/${value}`} className="underline-offset-4 hover:underline">
          {value}
        </Link>
      ),
    },
    textColumn({dataIndex: "organization", title: i18next.t("general:Organization"), width: 140}),
    dateColumn(),
    textColumn({dataIndex: "attemptCount", title: i18next.t("webhook:Attempt Count"), width: 130}),
    {
      dataIndex: "nextRetryTime",
      title: i18next.t("webhook:Next Retry Time"),
      width: 165,
      sortable: true,
      render: (value) => Setting.getFormattedDate(value),
    },
    {
      dataIndex: "state",
      title: i18next.t("general:State"),
      width: 110,
      sortable: true,
      render: (value) => <Badge variant={stateVariant(value) as any}>{value}</Badge>,
    },
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Webhook Events")}
      columns={columns}
      deps={[organizationName]}
      actionColumnWidth={120}
      fetch={(q) =>
        WebhookEventBackend.getWebhookEvents(
          "admin",
          organizationName,
          q.page,
          q.pageSize,
          "",
          "",
          q.sortField,
          q.sortOrder,
        )
      }
      rowActions={(record) => (
        <Button
          variant="outline"
          size="sm"
          onClick={() => {
            WebhookEventBackend.replayWebhookEvent(record.name).then((res: any) => {
              if (res.status === "ok") {
                Setting.showMessage("success", i18next.t("general:Successfully saved"));
              } else {
                Setting.showMessage("error", res.msg);
              }
            });
          }}
        >
          {i18next.t("webhook:Replay")}
        </Button>
      )}
    />
  );
}
