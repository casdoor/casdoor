import i18next from "i18next";
import {useParams} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Switch} from "@/components/ui/switch";
import {SelectField} from "@/components/common/SelectField";
import {CodeEditor} from "@/components/common/CodeEditor";
import {EditableTable} from "@/components/crud/EditableTable";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useCertOptions, useOrganizationOptions} from "@/hooks/use-options";
import {getSyncerTableColumns} from "@/lib/syncer-columns";
import * as SyncerBackend from "@/backend/SyncerBackend";
import * as Setting from "@/lib/setting";

const TYPES = [
  "Database", "Keycloak", "WeCom", "Azure AD", "Active Directory", "Google Workspace",
  "DingTalk", "Lark", "Okta", "SCIM", "AWS IAM",
];

const DATABASE_TYPES = [
  {id: "mysql", name: "MySQL"},
  {id: "postgres", name: "PostgreSQL"},
  {id: "mssql", name: "SQL Server"},
  {id: "oracle", name: "Oracle"},
  {id: "sqlite3", name: "Sqlite 3"},
];

const SSL_MODES = ["disable", "require", "verify-ca", "verify-full"];

const SYNCER_TABLE_TYPES = ["string", "integer", "boolean"];

/** the types that talk to an API instead of a database, so they have no schema fields */
const API_TYPES = [
  "WeCom", "Azure AD", "Google Workspace", "DingTalk", "Lark", "Okta", "SCIM", "AWS IAM",
];

/** these have no host at all: the SDK knows the endpoint */
const NO_HOST_TYPES = ["WeCom", "DingTalk", "Lark"];

const isApiType = (type: string) => API_TYPES.includes(type);

/** Active Directory is LDAP, so it keeps host/port/base DN but has no database schema */
const hasDatabaseSchema = (type: string) => !isApiType(type) && type !== "Active Directory";

/** SSH tunnelling is only wired up for the databases Casdoor can dial through it */
const needSshFields = (syncer: any) =>
  syncer.type === "Database" && ["mysql", "mssql", "postgres"].includes(syncer.databaseType);

/** "user" carries the client id / app key / bind DN, so the label follows the type */
function userLabelKey(type: string): string {
  switch (type) {
  case "WeCom": return "syncer:Corp ID";
  case "DingTalk": return "provider:App Key";
  case "Lark": return "provider:App ID";
  case "Azure AD": return "provider:Client ID";
  case "Active Directory": return "syncer:Bind DN";
  case "SCIM": return "syncer:Username (optional)";
  case "AWS IAM": return "syncer:AWS Access Key ID";
  default: return "general:User";
  }
}

function passwordLabelKey(type: string): string {
  switch (type) {
  case "WeCom": return "syncer:Corp secret";
  case "DingTalk":
  case "Lark": return "provider:App secret";
  case "Azure AD": return "provider:Client secret";
  case "SCIM": return "syncer:API Token / Password";
  case "AWS IAM": return "syncer:AWS Secret Access Key";
  default: return "general:Password";
  }
}

function hostLabelKey(type: string): string {
  switch (type) {
  case "Azure AD": return "provider:Tenant ID";
  case "Google Workspace": return "syncer:Admin Email";
  case "Active Directory": return "ldap:Server";
  case "SCIM": return "syncer:SCIM Server URL";
  case "AWS IAM": return "syncer:AWS Region";
  default: return "general:Host";
  }
}

export default function SyncerEditPage() {
  const {organizationName = "", syncerName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();

  const fields: EditField[] = [
    {
      type: "select",
      name: "organization",
      labelKey: "general:Organization",
      options: () => organizations,
      disabled: () => !Setting.isAdminUser(account),
      // the certs are per-organization, so the SSH cert cannot survive the move
      onChange: (value, ctx, updateFields) =>
        updateFields(value === ctx.record.organization ? {organization: value} : {organization: value, cert: ""}),
    },
    {type: "text", name: "name", labelKey: "general:Name"},
    {
      type: "select",
      name: "type",
      labelKey: "general:Type",
      options: () => TYPES.map((item) => ({value: item, label: item})),
      onChange: (value, ctx, updateFields) =>
        updateFields({
          type: value,
          tableColumns: getSyncerTableColumns(value),
          table: value === "Keycloak" ? "user_entity" : ctx.record.table,
          host: isApiType(value) ? "" : ctx.record.host,
        }),
    },
    {
      type: "select",
      name: "databaseType",
      labelKey: "syncer:Database type",
      when: (ctx) => hasDatabaseSchema(ctx.record.type),
      options: () => DATABASE_TYPES.map((item) => ({value: item.id, label: item.name})),
      // only PostgreSQL takes an sslmode, and it must not be left over on the others
      onChange: (value, _ctx, updateFields) =>
        updateFields({databaseType: value, sslMode: value === "postgres" ? "disable" : ""}),
    },
    {
      type: "select",
      name: "sslMode",
      labelKey: "provider:SSL mode",
      when: (ctx) => ctx.record.databaseType === "postgres",
      options: () => SSL_MODES.map((item) => ({value: item, label: item})),
    },
    {
      type: "text",
      name: "host",
      labelKey: (ctx) => hostLabelKey(ctx.record.type),
      when: (ctx) => !NO_HOST_TYPES.includes(ctx.record.type),
    },
    {
      type: "number",
      name: "port",
      labelKey: (ctx) => (ctx.record.type === "Active Directory" ? "provider:LDAP port" : "general:Port"),
      when: (ctx) => !isApiType(ctx.record.type),
    },
    {
      type: "text",
      name: "user",
      labelKey: (ctx) => userLabelKey(ctx.record.type),
      when: (ctx) => ctx.record.type !== "Google Workspace",
    },
    {
      // the Google Workspace credential is a whole service account JSON key
      type: "textarea",
      name: "password",
      labelKey: "syncer:Service account key",
      rows: 6,
      placeholder: i18next.t("syncer:Paste your Google Workspace service account JSON key here"),
      when: (ctx) => ctx.record.type === "Google Workspace",
    },
    {
      type: "password",
      name: "password",
      labelKey: (ctx) => passwordLabelKey(ctx.record.type),
      when: (ctx) => ctx.record.type !== "Google Workspace",
    },
    {
      type: "text",
      name: "database",
      labelKey: (ctx) => (ctx.record.type === "Active Directory" ? "ldap:Base DN" : "syncer:Database"),
      when: (ctx) => !isApiType(ctx.record.type),
    },
    {
      type: "select",
      name: "sshType",
      labelKey: "general:SSH type",
      when: (ctx) => needSshFields(ctx.record),
      options: () => [
        {value: "", label: i18next.t("general:None")},
        {value: "password", label: i18next.t("general:Password")},
        {value: "cert", label: i18next.t("general:Cert")},
      ],
    },
    {
      type: "text",
      name: "sshHost",
      labelKey: "syncer:SSH host",
      when: (ctx) => needSshFields(ctx.record) && !!ctx.record.sshType,
    },
    {
      type: "number",
      name: "sshPort",
      labelKey: "syncer:SSH port",
      when: (ctx) => needSshFields(ctx.record) && !!ctx.record.sshType,
    },
    {
      type: "text",
      name: "sshUser",
      labelKey: "syncer:SSH user",
      when: (ctx) => needSshFields(ctx.record) && !!ctx.record.sshType,
    },
    {
      type: "password",
      name: "sshPassword",
      labelKey: "syncer:SSH password",
      when: (ctx) => needSshFields(ctx.record) && ctx.record.sshType === "password",
    },
    {
      type: "custom",
      name: "cert",
      labelKey: "general:SSH cert",
      when: (ctx) => needSshFields(ctx.record) && ctx.record.sshType === "cert",
      render: (ctx, update) => (
        <SshCertSelect record={ctx.record} onChange={(value) => update("cert", value)} />
      ),
    },
    {
      type: "text",
      name: "table",
      labelKey: "syncer:Table",
      when: (ctx) => !isApiType(ctx.record.type),
    },
    {
      type: "custom",
      name: "tableColumns",
      labelKey: "syncer:Table columns",
      block: true,
      render: (ctx, update) => (
        <EditableTable
          rows={ctx.record.tableColumns ?? []}
          onChange={(rows) => update("tableColumns", rows)}
          newRow={() => ({name: "", type: "string", casdoorName: "id", isKey: false, isHashed: true, values: []})}
          columns={[
            {
              key: "name",
              title: i18next.t("syncer:Column name"),
              width: 200,
              render: (row: any, _i, patch) => (
                <Input value={row.name ?? ""} onChange={(e) => patch({name: e.target.value})} />
              ),
            },
            {
              key: "type",
              title: i18next.t("syncer:Column type"),
              width: 160,
              render: (row: any, _i, patch) => (
                <SelectField
                  value={row.type}
                  onChange={(v) => patch({type: v})}
                  options={SYNCER_TABLE_TYPES.map((item) => ({id: item, name: item}))}
                />
              ),
            },
            {
              key: "casdoorName",
              title: i18next.t("syncer:Casdoor column"),
              width: 200,
              render: (row: any, _i, patch) => (
                <SelectField
                  value={row.casdoorName}
                  onChange={(v) => patch({casdoorName: v})}
                  options={Setting.getUserCommonFields().map((item: string) => ({id: item, name: item}))}
                />
              ),
            },
            {
              key: "isKey",
              title: i18next.t("syncer:Is key"),
              width: 90,
              render: (row: any, _i, patch) => (
                <Switch checked={!!row.isKey} onCheckedChange={(v) => patch({isKey: v})} />
              ),
            },
            {
              key: "isHashed",
              title: i18next.t("syncer:Is hashed"),
              width: 110,
              render: (row: any, _i, patch) => (
                <Switch checked={!!row.isHashed} onCheckedChange={(v) => patch({isHashed: v})} />
              ),
            },
          ]}
        />
      ),
    },
    {
      type: "text",
      name: "affiliationTable",
      labelKey: "syncer:Affiliation table",
      when: (ctx) => !isApiType(ctx.record.type),
    },
    {type: "text", name: "avatarBaseUrl", labelKey: "syncer:Avatar base URL"},
    {type: "number", name: "syncInterval", labelKey: "syncer:Sync interval"},
    {
      type: "custom",
      name: "errorText",
      labelKey: "syncer:Error text",
      block: true,
      render: (ctx) => (
        <CodeEditor
          value={ctx.record.errorText ?? ""}
          onChange={() => undefined}
          readOnly
          language="javascript"
          height={300}
        />
      ),
    },
    {type: "switch", name: "isReadOnly", labelKey: "syncer:Is read-only"},
    {type: "switch", name: "isEnabled", labelKey: "general:Is enabled"},
  ];

  return (
    <SimpleEditPage
      titleKey="syncer:Edit Syncer"
      backTo="/syncers"
      deps={[organizationName, syncerName]}
      fields={fields}
      fetch={() => SyncerBackend.getSyncer("admin", syncerName, organizationName)}
      add={(record) => SyncerBackend.addSyncer(record)}
      update={(record) => SyncerBackend.updateSyncer("admin", syncerName, record)}
      editUrl={(record) => `/syncers/${record.organization}/${record.name}`}
      extraActions={(ctx) => (
        <>
          <Button
            variant="outline"
            onClick={() => {
              SyncerBackend.testSyncerDb(ctx.record)
                .then((res: any) => {
                  if (res.status === "ok") {
                    Setting.showMessage("success", i18next.t("syncer:Connect successfully"));
                  } else {
                    Setting.showMessage("error", `${i18next.t("syncer:Failed to connect")}: ${res.msg}`);
                  }
                })
                .catch((error: any) => {
                  Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
                });
            }}
          >
            {i18next.t("syncer:Test Connection")}
          </Button>
          <Button
            variant="outline"
            onClick={() => {
              SyncerBackend.runSyncer("admin", ctx.record.name, ctx.record.organization).then((res: any) => {
                if (res.status === "ok") {
                  Setting.showMessage("success", i18next.t("general:Successfully synced"));
                } else {
                  Setting.showMessage("error", res.msg);
                }
              });
            }}
          >
            {i18next.t("general:Sync")}
          </Button>
        </>
      )}
    />
  );
}

/** the SSH cert list follows the syncer's organization, so it is its own component */
function SshCertSelect({record, onChange}: {record: any; onChange: (value: string) => void}) {
  const certs = useCertOptions(record.organization ?? "");
  return (
    <SelectField
      value={record.cert}
      onChange={onChange}
      options={certs.map((item) => ({id: item.value, name: item.label}))}
    />
  );
}
