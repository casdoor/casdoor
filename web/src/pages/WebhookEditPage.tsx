import i18next from "i18next";
import {useParams} from "react-router-dom";
import {Input} from "@/components/ui/input";
import {EditableTable} from "@/components/crud/EditableTable";
import {CodeEditor} from "@/components/common/CodeEditor";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {SearchableSelect} from "@/components/common/SearchableSelect";
import {useOrganizationOptions, useProviderList} from "@/hooks/use-options";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as WebhookBackend from "@/backend/WebhookBackend";
import {buildWebhookPreview} from "@/lib/webhook-preview";
import * as Setting from "@/lib/setting";

const METHODS = ["POST", "GET", "PUT", "DELETE"];
const CONTENT_TYPES = ["application/json", "application/x-www-form-urlencoded"];
const TYPES = ["HTTP", "Email", "SMS", "Notification"];

const isHttp = (ctx: {record: any}) => !TYPES.slice(1).includes(ctx.record.type);
const isProvider = (ctx: {record: any}) => !isHttp(ctx);

export default function WebhookEditPage() {
  const {webhookName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();
  // GetWebhook() looks the webhook up inside this organization for a non-global admin
  const organizationName = useRequestOrganization();

  const fields: EditField[] = [
    {
      type: "select",
      name: "organization",
      labelKey: "general:Organization",
      options: () => organizations,
      disabled: () => !Setting.isAdminUser(account),
    },
    {type: "text", name: "name", labelKey: "general:Name", required: true},
    {
      type: "custom",
      name: "type",
      labelKey: "general:Type",
      render: (ctx, update) => (
        <SearchableSelect
          value={ctx.record.type || "HTTP"}
          onChange={(value) => {
            update("type", value);
            update("provider", "");
          }}
          options={TYPES.map((item) => ({value: item, label: item}))}
        />
      ),
    },
    {
      type: "custom",
      name: "provider",
      labelKey: "general:Provider",
      when: isProvider,
      render: (ctx, update) => (
        <WebhookProviderSelect
          organization={ctx.record.organization}
          category={ctx.record.type}
          value={ctx.record.provider}
          onChange={(value) => update("provider", value)}
        />
      ),
    },
    {type: "tags", name: "receivers", labelKey: "webhook:Receivers", when: isProvider, placeholder: "$user"},
    {type: "text", name: "title", labelKey: "general:Title", when: (ctx) => ctx.record.type === "Email"},
    {type: "textarea", name: "content", labelKey: "webhook:Message content", when: isProvider, rows: 6},
    {type: "text", name: "url", labelKey: "general:URL", when: isHttp},
    {
      type: "select",
      name: "method",
      labelKey: "general:Method",
      when: isHttp,
      options: () => METHODS.map((item) => ({value: item, label: item})),
    },
    {
      type: "select",
      name: "contentType",
      labelKey: "webhook:Content type",
      when: isHttp,
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
      when: isHttp,
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
      when: isHttp,
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
      when: isHttp,
      creatable: true,
      options: () => Setting.getUserCommonFields().map((item: string) => ({value: item, label: item})),
    },
    {
      type: "custom",
      name: "preview",
      labelKey: "general:Preview",
      when: isHttp,
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
    {type: "switch", name: "isUserExtended", labelKey: "webhook:Is user extended", when: isHttp},
    {type: "switch", name: "singleOrgOnly", labelKey: "webhook:Single org only"},
    {type: "switch", name: "isEnabled", labelKey: "general:Is enabled"},
  ];

  return (
    <SimpleEditPage
      titleKey="webhook:Edit Webhook"
      backTo="/webhooks"
      deps={[webhookName, organizationName]}
      fields={fields}
      fetch={() => WebhookBackend.getWebhook("admin", webhookName, organizationName)}
      add={(record) => WebhookBackend.addWebhook(record)}
      update={(record) => WebhookBackend.updateWebhook("admin", webhookName, record)}
      editUrl={(record) => `/webhooks/${record.name}`}
    />
  );
}

/** the backend only sends with the organization's own providers, and with the shared ones for "built-in" */
function WebhookProviderSelect({organization, category, value, onChange}: {organization: string; category: string; value: string; onChange: (value: string) => void}) {
  const providers = useProviderList(organization);
  const options = providers
    .filter((item: any) => item.category === category)
    .filter((item: any) => item.owner === organization || (organization === "built-in" && item.owner === "admin"))
    .map((item: any) => ({value: item.name, label: item.displayName ? `${item.displayName} (${item.name})` : item.name}));

  return <SearchableSelect value={value ?? ""} onChange={onChange} options={options} />;
}
