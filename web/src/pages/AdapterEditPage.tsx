import i18next from "i18next";
import {useParams} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {Switch} from "@/components/ui/switch";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useOrganizationOptions} from "@/hooks/use-options";
import * as AdapterBackend from "@/backend/AdapterBackend";
import * as Setting from "@/lib/setting";

export const DATABASE_TYPES = [
  {id: "mysql", name: "MySQL"},
  {id: "postgres", name: "PostgreSQL"},
  {id: "mssql", name: "SQL Server"},
  {id: "oracle", name: "Oracle"},
  {id: "sqlite3", name: "Sqlite 3"},
];

/** the fields the "Use same DB" switch fills in or clears, see web/src/AdapterEditPage.js */
const OWN_DB_DEFAULTS = {
  type: "Database",
  databaseType: "mysql",
  host: "localhost",
  port: 3306,
  user: "root",
  password: "123456",
  database: "dbName",
};

const OWN_DB_CLEARED = {
  type: "",
  databaseType: "",
  host: "",
  port: 0,
  user: "",
  password: "",
  database: "",
};

export default function AdapterEditPage() {
  const {organizationName = "", adapterName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();

  const isBuiltIn = (ctx: {record: any}) => Setting.builtInObject(ctx.record);
  const usesOwnDb = (ctx: {record: any}) => ctx.record.useSameDb !== true && !Setting.builtInObject(ctx.record);

  const fields: EditField[] = [
    {
      type: "select",
      name: "owner",
      labelKey: "general:Organization",
      options: () => organizations,
      disabled: (ctx) => !Setting.isAdminUser(account) || Setting.builtInObject(ctx.record),
    },
    {type: "text", name: "name", labelKey: "general:Name", disabled: isBuiltIn},
    {type: "text", name: "table", labelKey: "syncer:Table", disabled: isBuiltIn},
    {
      type: "custom",
      name: "useSameDb",
      labelKey: "adapter:Use same DB",
      render: (ctx, update) => (
        <Switch
          disabled={Setting.builtInObject(ctx.record)}
          checked={!!ctx.record.useSameDb || Setting.builtInObject(ctx.record)}
          onCheckedChange={(checked) => {
            update("useSameDb", checked);
            const values: Record<string, any> = checked ? OWN_DB_CLEARED : OWN_DB_DEFAULTS;
            Object.entries(values).forEach(([key, value]) => update(key, value));
          }}
        />
      ),
    },
    {
      type: "select",
      name: "type",
      labelKey: "general:Type",
      when: usesOwnDb,
      disabled: isBuiltIn,
      options: () => [{value: "Database", label: "Database"}],
    },
    {
      type: "select",
      name: "databaseType",
      labelKey: "syncer:Database type",
      when: usesOwnDb,
      disabled: isBuiltIn,
      options: () => DATABASE_TYPES.map((item) => ({value: item.id, label: item.name})),
    },
    {type: "text", name: "host", labelKey: "general:Host", when: usesOwnDb},
    {type: "number", name: "port", labelKey: "general:Port", when: usesOwnDb},
    {type: "text", name: "user", labelKey: "general:User", when: usesOwnDb},
    {type: "password", name: "password", labelKey: "general:Password", when: usesOwnDb},
    {type: "text", name: "database", labelKey: "syncer:Database", when: usesOwnDb, disabled: isBuiltIn},
    {
      type: "custom",
      name: "dbTest",
      labelKey: "provider:DB test",
      render: (ctx) => (
        <Button
          // the test goes through the saved adapter record, so it is not available before the adapter is created
          disabled={ctx.mode === "add" || organizationName !== ctx.record.owner}
          onClick={() => {
            AdapterBackend.getPolicies("", "", `${ctx.record.owner}/${ctx.record.name}`)
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
          {i18next.t("syncer:Test DB Connection")}
        </Button>
      ),
    },
  ];

  return (
    <SimpleEditPage
      titleKey="adapter:Edit Adapter"
      backTo="/adapters"
      deps={[organizationName, adapterName]}
      fields={fields}
      fetch={() => AdapterBackend.getAdapter(organizationName, adapterName)}
      add={(record) => AdapterBackend.addAdapter(record)}
      update={(record) => AdapterBackend.updateAdapter(organizationName, adapterName, record)}
      editUrl={(record) => `/adapters/${record.owner}/${record.name}`}
    />
  );
}
