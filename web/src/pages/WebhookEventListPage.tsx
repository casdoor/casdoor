import * as React from "react";
import i18next from "i18next";
import {Link} from "react-router-dom";
import {Badge} from "@/components/ui/badge";
import {Dialog, DialogContent, DialogHeader, DialogTitle} from "@/components/ui/dialog";
import {CodeEditor} from "@/components/common/CodeEditor";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as WebhookEventBackend from "@/backend/WebhookEventBackend";
import * as Setting from "@/lib/setting";

/** the four states the backend puts on a delivery attempt, plus an unknown fallback */
const STATE_VARIANTS: Record<string, string> = {
  Pending: "warning",
  Success: "success",
  Failed: "destructive",
  Retrying: "secondary",
};

function StateBadge({state}: {state: string}) {
  const variant = STATE_VARIANTS[state] ?? "outline";
  const label = STATE_VARIANTS[state] ? i18next.t("webhook:" + state) : state || i18next.t("webhook:Unknown");
  return <Badge variant={variant as any}>{label}</Badge>;
}

/** pretty-print the stored payload, falling back to the raw string when it is not JSON */
function formatJson(text: any): string {
  if (!text) {
    return "";
  }
  try {
    return JSON.stringify(JSON.parse(String(text)), null, 2);
  } catch (error) {
    return String(text);
  }
}

export default function WebhookEventListPage() {
  const organizationName = useRequestOrganization();
  const [detail, setDetail] = React.useState<any>(null);
  const [replayingId, setReplayingId] = React.useState("");

  const replay = (record: any, refresh: () => void) => {
    const eventId = `${record.owner}/${record.name}`;
    setReplayingId(eventId);
    WebhookEventBackend.replayWebhookEvent(eventId)
      .then((res: any) => {
        if (res.status === "ok") {
          Setting.showMessage(
            "success",
            typeof res.data === "string" ? res.data : i18next.t("webhook:Webhook event replay triggered"),
          );
          refresh();
        } else {
          Setting.showMessage("error", `${i18next.t("webhook:Failed to replay webhook event")}: ${res.msg}`);
        }
      })
      .catch((error) =>
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`),
      )
      .finally(() => setReplayingId(""));
  };

  const columns: ColumnDef<any>[] = [
    textColumn({dataIndex: "name", title: i18next.t("general:Name"), width: 180, searchable: true, mono: true, fixed: "left"}),
    {
      dataIndex: "webhook",
      title: i18next.t("general:Webhook"),
      width: 180,
      sortable: true,
      searchable: true,
      render: (value) =>
        value ? (
          <Link to={`/webhooks/${Setting.getShortName(value)}`} className="underline-offset-4 hover:underline">
            {Setting.getShortName(value)}
          </Link>
        ) : (
          "-"
        ),
    },
    textColumn({dataIndex: "organization", title: i18next.t("general:Organization"), width: 140, link: (v) => `/organizations/${v}`}),
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
      filters: Object.keys(STATE_VARIANTS).map((state) => ({value: state, label: i18next.t("webhook:" + state)})),
      render: (value) => <StateBadge state={value} />,
    },
  ];

  return (
    <>
      <CrudListPage
        title={i18next.t("general:Webhook Events")}
        columns={columns}
        deps={[organizationName]}
        actionColumnWidth={200}
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
        rowActions={(record, _index, {refresh}) => [
          {key: "view", label: i18next.t("general:View"), onSelect: () => setDetail(record)},
          // a delivery that already succeeded has nothing to replay
          record.state !== "Success"
            ? {
              key: "replay",
              label: i18next.t("webhook:Replay"),
              loading: replayingId === `${record.owner}/${record.name}`,
              onSelect: () => replay(record, refresh),
            }
            : null,
        ]}
      />

      <Dialog open={detail !== null} onOpenChange={(open) => (open ? undefined : setDetail(null))}>
        <DialogContent className="sm:max-w-[760px]">
          <DialogHeader>
            <DialogTitle>{i18next.t("general:Detail")}</DialogTitle>
          </DialogHeader>
          {detail ? (
            <div className="max-h-[70vh] space-y-3 overflow-y-auto">
              <dl className="grid grid-cols-[minmax(120px,180px)_1fr] gap-x-4 gap-y-2 text-sm">
                <dt className="text-muted-foreground">{i18next.t("general:Name")}</dt>
                <dd className="font-mono text-xs">{detail.name || "-"}</dd>
                <dt className="text-muted-foreground">{i18next.t("general:Webhook")}</dt>
                <dd>
                  {detail.webhook ? (
                    <Link
                      to={`/webhooks/${Setting.getShortName(detail.webhook)}`}
                      className="underline-offset-4 hover:underline"
                    >
                      {Setting.getShortName(detail.webhook)}
                    </Link>
                  ) : (
                    "-"
                  )}
                </dd>
                <dt className="text-muted-foreground">{i18next.t("general:Organization")}</dt>
                <dd>
                  {detail.organization ? (
                    <Link to={`/organizations/${detail.organization}`} className="underline-offset-4 hover:underline">
                      {detail.organization}
                    </Link>
                  ) : (
                    "-"
                  )}
                </dd>
                <dt className="text-muted-foreground">{i18next.t("general:Created time")}</dt>
                <dd>{detail.createdTime ? Setting.getFormattedDate(detail.createdTime) : "-"}</dd>
                <dt className="text-muted-foreground">{i18next.t("general:State")}</dt>
                <dd><StateBadge state={detail.state} /></dd>
                <dt className="text-muted-foreground">{i18next.t("webhook:Attempt Count")}</dt>
                <dd>{detail.attemptCount || 0}</dd>
                <dt className="text-muted-foreground">{i18next.t("webhook:Next Retry Time")}</dt>
                <dd>{detail.nextRetryTime ? Setting.getFormattedDate(detail.nextRetryTime) : "-"}</dd>
              </dl>
              <div className="space-y-1">
                <p className="text-sm text-muted-foreground">{i18next.t("webhook:Payload")}</p>
                <CodeEditor language="json" value={formatJson(detail.payload)} readOnly onChange={() => undefined} />
              </div>
              <div className="space-y-1">
                <p className="text-sm text-muted-foreground">{i18next.t("webhook:Last Error")}</p>
                <CodeEditor height={120} value={detail.lastError || "-"} readOnly onChange={() => undefined} />
              </div>
            </div>
          ) : null}
        </DialogContent>
      </Dialog>
    </>
  );
}
