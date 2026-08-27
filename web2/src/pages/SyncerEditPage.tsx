import i18next from "i18next";
import {useParams} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Switch} from "@/components/ui/switch";
import {SelectField} from "@/components/common/SelectField";
import {EditableTable} from "@/components/crud/EditableTable";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useOrganizationOptions} from "@/hooks/use-options";
import * as SyncerBackend from "@/backend/SyncerBackend";
import * as Setting from "@/lib/setting";

const TYPES = ["Database", "LDAP", "Keycloak", "Casdoor", "WeCom", "Lark", "DingTalk"];
const DATABASE_TYPES = ["mysql", "postgres", "mssql", "oracle", "sqlite3", "tidb"];
const SSL_MODES = ["Default", "SSL", "TLS", "SSH"];
const SYNCER_TABLE_TYPES = ["string", "integer", "boolean"];

export default function SyncerEditPage() {
  const {organizationName = "", syncerName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();

  const isDatabase = (ctx: {record: any}) => ctx.record.type === "Database";

  const fields: EditField[] = [
    {
      type: "select",
      name: "organization",
      labelKey: "general:Organization",
      options: () => organizations,
      disabled: () => !Setting.isAdminUser(account),
    },
    {type: "text", name: "name", labelKey: "general:Name"},
    {
      type: "select",
      name: "type",
      labelKey: "general:Type",
      options: () => TYPES.map((item) => ({value: item, label: item})),
    },
    {
      type: "select",
      name: "databaseType",
      labelKey: "syncer:Database type",
      when: isDatabase,
      options: () => DATABASE_TYPES.map((item) => ({value: item, label: item})),
    },
    {type: "text", name: "host", labelKey: "general:Host"},
    {type: "number", name: "port", labelKey: "general:Port"},
    {type: "text", name: "user", labelKey: "general:User"},
    {type: "password", name: "password", labelKey: "general:Password"},
    {type: "text", name: "database", labelKey: "syncer:Database", when: isDatabase},
    {type: "text", name: "table", labelKey: "syncer:Table", when: isDatabase},
    {
      type: "select",
      name: "sslMode",
      labelKey: "syncer:SSL mode",
      options: () => SSL_MODES.map((item) => ({value: item, label: item})),
    },
    {type: "text", name: "sshHost", labelKey: "syncer:SSH host", when: (ctx) => ctx.record.sslMode === "SSH"},
    {type: "number", name: "sshPort", labelKey: "syncer:SSH port", when: (ctx) => ctx.record.sslMode === "SSH"},
    {type: "text", name: "sshUser", labelKey: "syncer:SSH user", when: (ctx) => ctx.record.sslMode === "SSH"},
    {type: "text", name: "affiliationTable", labelKey: "syncer:Affiliation table", when: isDatabase},
    {type: "text", name: "avatarBaseUrl", labelKey: "syncer:Avatar base URL"},
    {type: "number", name: "syncInterval", labelKey: "syncer:Sync interval"},
    {type: "switch", name: "isReadOnly", labelKey: "syncer:Is read-only"},
    {type: "switch", name: "isEnabled", labelKey: "general:Is enabled"},
    {
      type: "custom",
      name: "tableColumns",
      labelKey: "syncer:Table columns",
      block: true,
      render: (ctx, update) => (
        <EditableTable
          rows={ctx.record.tableColumns ?? []}
          onChange={(rows) => update("tableColumns", rows)}
          newRow={() => ({name: "", type: "string", casdoorName: "id", isHashed: true, values: []})}
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
              SyncerBackend.testSyncerDb(ctx.record).then((res: any) => {
                if (res.status === "ok") {
                  Setting.showMessage("success", i18next.t("general:Connection successful"));
                } else {
                  Setting.showMessage("error", res.msg);
                }
              });
            }}
          >
            {i18next.t("syncer:Test DB Connection")}
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
