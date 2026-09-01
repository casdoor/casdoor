import i18next from "i18next";
import {useParams} from "react-router-dom";
import {Input} from "@/components/ui/input";
import {EditableTable} from "@/components/crud/EditableTable";
import {CodeEditor} from "@/components/common/CodeEditor";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useOrganizationOptions} from "@/hooks/use-options";
import * as WebhookBackend from "@/backend/WebhookBackend";
import {buildWebhookPreview} from "@/lib/webhook-preview";
import * as Setting from "@/lib/setting";

const METHODS = ["POST", "GET", "PUT", "DELETE"];
const CONTENT_TYPES = ["application/json", "application/x-www-form-urlencoded"];

export default function WebhookEditPage() {
  const {webhookName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();

  const fields: EditField[] = [
    {
      type: "select",
      name: "organization",
      labelKey: "general:Organization",
      options: () => organizations,
      disabled: () => !Setting.isAdminUser(account),
    },
    {type: "text", name: "name", labelKey: "general:Name"},
    {type: "text", name: "url", labelKey: "general:URL"},
    {
      type: "select",
      name: "method",
      labelKey: "general:Method",
      options: () => METHODS.map((item) => ({value: item, label: item})),
    },
    {
      type: "select",
      name: "contentType",
      labelKey: "webhook:Content type",
      options: () => CONTENT_TYPES.map((item) => ({value: item, label: item})),
    },
    {
      type: "multiselect",
      name: "events",
      labelKey: "webhook:Events",
      creatable: true,
      options: () => Setting.getApiPaths().map((path: string) => ({value: path, label: path})),
    },
    {
      type: "custom",
      name: "headers",
      labelKey: "webhook:Headers",
      block: true,
      render: (ctx, update) => (
        <EditableTable
          rows={ctx.record.headers ?? []}
          onChange={(rows) => update("headers", rows)}
          newRow={() => ({name: "", value: ""})}
          reorderable={false}
          columns={[
            {
              key: "name",
              title: i18next.t("general:Name"),
              width: 240,
              render: (row: any, _i, patch) => (
                <Input value={row.name ?? ""} onChange={(e) => patch({name: e.target.value})} />
              ),
            },
            {
              key: "value",
              title: i18next.t("webhook:Value"),
              render: (row: any, _i, patch) => (
                <Input value={row.value ?? ""} onChange={(e) => patch({value: e.target.value})} />
              ),
            },
          ]}
        />
      ),
    },
    {
      type: "multiselect",
      name: "objectFields",
      labelKey: "webhook:Object fields",
      creatable: true,
      // antd offers "All" above the field list as a catch-all
      options: () => [
        {value: "All", label: i18next.t("general:All")},
        ...Setting.getUserCommonFields().map((item: string) => ({value: item, label: item})),
      ],
    },
    {
      type: "multiselect",
      name: "tokenFields",
      labelKey: "webhook:Extended user fields",
      creatable: true,
      options: () => Setting.getUserCommonFields().map((item: string) => ({value: item, label: item})),
    },
    {
      type: "custom",
      name: "preview",
      labelKey: "general:Preview",
      block: true,
      render: (ctx) => (
        <CodeEditor
          language="json"
          readOnly
          height={300}
          value={buildWebhookPreview(ctx.record)}
          onChange={() => {}}
        />
      ),
    },
    {type: "switch", name: "isUserExtended", labelKey: "webhook:Is user extended"},
    {type: "switch", name: "singleOrgOnly", labelKey: "webhook:Single org only"},
    {type: "switch", name: "isEnabled", labelKey: "general:Is enabled"},
  ];

  return (
    <SimpleEditPage
      titleKey="webhook:Edit Webhook"
      backTo="/webhooks"
      deps={[webhookName]}
      fields={fields}
      fetch={() => WebhookBackend.getWebhook("admin", webhookName)}
      add={(record) => WebhookBackend.addWebhook(record)}
      update={(record) => WebhookBackend.updateWebhook("admin", webhookName, record)}
      editUrl={(record) => `/webhooks/${record.name}`}
    />
  );
}
