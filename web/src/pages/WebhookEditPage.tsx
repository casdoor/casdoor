import i18next from "i18next";
import {useParams} from "react-router-dom";
import {Input} from "@/components/ui/input";
import {EditableTable} from "@/components/crud/EditableTable";
import {CodeEditor} from "@/components/common/CodeEditor";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {SearchableSelect} from "@/components/common/SearchableSelect";
import {TagsInput} from "@/components/common/TagsInput";
import {TemplateInput, type TemplateVariable} from "@/components/common/TemplateInput";
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

// the variables object/webhook_provider.go replaces in the title and the content
function getTemplateVariables(): TemplateVariable[] {
  return [
    {name: "organization", description: i18next.t("general:Organization")},
    {name: "user", description: i18next.t("webhook:User name")},
    {name: "user.displayName", description: i18next.t("general:Display name")},
    {name: "user.email", description: i18next.t("general:Email")},
    {name: "user.phone", description: i18next.t("general:Phone")},
    {name: "event", description: i18next.t("webhook:Event, e.g. new-user")},
    {name: "ip", description: i18next.t("general:Client IP")},
    {name: "time", description: i18next.t("webhook:Event time")},
  ];
}

// sample values for the preview, taken from the signed-in admin
function renderTemplatePreview(text: string, account: any): string {
  const values: Record<string, string> = {
    "organization": account?.owner ?? "",
    "user": account?.name ?? "",
    "user.displayName": account?.displayName ?? "",
    "user.email": account?.email ?? "",
    "user.phone": account?.phone ?? "",
    "event": "new-user",
    "ip": "203.0.113.10",
    "time": new Date().toISOString(),
  };
  return String(text ?? "").replace(/\$\{([\w.]+)\}/g, (all, name) => (name in values ? values[name] : all));
}

const RECEIVER_PLACEHOLDERS: Record<string, string> = {
  Email: "admin@example.com",
  SMS: "+12025550123",
  Notification: "",
};

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
    {
      type: "custom",
      name: "receivers",
      labelKey: "webhook:Receivers",
      when: isProvider,
      render: (ctx, update) => {
        const receivers: string[] = Array.isArray(ctx.record.receivers) ? ctx.record.receivers : [];
        return (
          <div className="space-y-2">
            <TagsInput
              value={receivers}
              onChange={(value) => update("receivers", value)}
              placeholder={RECEIVER_PLACEHOLDERS[ctx.record.type] ?? ""}
            />
            <div className="flex flex-wrap items-center gap-1.5 text-xs text-muted-foreground">
              <button
                type="button"
                disabled={receivers.includes("$user")}
                className="rounded border bg-muted/50 px-1.5 py-0.5 font-mono text-foreground hover:bg-accent disabled:opacity-50"
                onClick={() => update("receivers", [...receivers, "$user"])}
              >
                + $user
              </button>
              <span>
                {ctx.record.type === "SMS" ? i18next.t("webhook:The phone of the user who triggered the event") :
                  ctx.record.type === "Email" ? i18next.t("webhook:The email of the user who triggered the event") :
                    i18next.t("webhook:Leave empty to send to the channel configured in the provider")}
              </span>
            </div>
          </div>
        );
      },
    },
    {
      type: "custom",
      name: "title",
      labelKey: "general:Title",
      when: (ctx) => ctx.record.type === "Email",
      render: (ctx, update) => (
        <TemplateInput
          value={ctx.record.title}
          onChange={(value) => update("title", value)}
          variables={getTemplateVariables()}
          placeholder={i18next.t("webhook:Type ${ to insert a variable")}
          preview={renderTemplatePreview(ctx.record.title, account)}
        />
      ),
    },
    {
      type: "custom",
      name: "content",
      labelKey: "webhook:Message content",
      when: isProvider,
      render: (ctx, update) => (
        <TemplateInput
          multiline
          rows={6}
          value={ctx.record.content}
          onChange={(value) => update("content", value)}
          variables={getTemplateVariables()}
          placeholder={i18next.t("webhook:Type ${ to insert a variable")}
          preview={renderTemplatePreview(ctx.record.content, account)}
        />
      ),
    },
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
