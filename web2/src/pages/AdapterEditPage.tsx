import {useParams} from "react-router-dom";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useOrganizationOptions} from "@/hooks/use-options";
import * as AdapterBackend from "@/backend/AdapterBackend";
import * as Setting from "@/lib/setting";

const DATABASE_TYPES = ["mysql", "postgres", "mssql", "oracle", "sqlite3", "tidb"];

export default function AdapterEditPage() {
  const {organizationName = "", adapterName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();

  const notSameDb = (ctx: {record: any}) => ctx.record.useSameDb !== true;

  const fields: EditField[] = [
    {
      type: "select",
      name: "owner",
      labelKey: "general:Organization",
      options: () => organizations,
      disabled: () => !Setting.isAdminUser(account),
    },
    {type: "text", name: "name", labelKey: "general:Name"},
    {type: "switch", name: "useSameDb", labelKey: "adapter:Use same DB"},
    {
      type: "select",
      name: "databaseType",
      labelKey: "syncer:Database type",
      when: notSameDb,
      options: () => DATABASE_TYPES.map((item) => ({value: item, label: item})),
    },
    {type: "text", name: "host", labelKey: "general:Host", when: notSameDb},
    {type: "number", name: "port", labelKey: "general:Port", when: notSameDb},
    {type: "text", name: "user", labelKey: "general:User", when: notSameDb},
    {type: "password", name: "password", labelKey: "general:Password", when: notSameDb},
    {type: "text", name: "database", labelKey: "syncer:Database", when: notSameDb},
    {type: "text", name: "table", labelKey: "syncer:Table"},
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
