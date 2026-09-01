import {Link, useParams} from "react-router-dom";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {EntryMessageViewer} from "@/components/entry/EntryMessageViewer";
import {useAccount} from "@/hooks/use-account";
import {useOrganizationOptions} from "@/hooks/use-options";
import * as EntryBackend from "@/backend/EntryBackend";
import * as Setting from "@/lib/setting";

/** antd's `isNonEmptyEntryField`: the read-only rows are hidden when the entry has no value. */
function isNonEmpty(value: any) {
  if (value === undefined || value === null) {
    return false;
  }
  return String(value).trim() !== "";
}

export default function EntryEditPage() {
  const {organizationName = "", entryName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();

  const fields: EditField[] = [
    {
      type: "select",
      name: "owner",
      labelKey: "general:Organization",
      options: () => organizations,
      disabled: () => !Setting.isAdminUser(account),
    },
    {type: "text", name: "name", labelKey: "general:Name", required: true},
    {type: "text", name: "displayName", labelKey: "general:Display name"},
    {
      type: "custom",
      name: "provider",
      labelKey: "general:Provider",
      when: ({record}) => isNonEmpty(record.provider),
      render: ({record}) => (
        <Link
          to={`/providers/${record.owner}/${record.provider}`}
          className="text-primary underline-offset-4 hover:underline"
        >
          {record.provider}
        </Link>
      ),
    },
    {
      type: "custom",
      name: "application",
      labelKey: "general:Application",
      when: ({record}) => isNonEmpty(record.application),
      render: ({record}) => (
        <Link
          to={`/applications/${record.owner}/${record.application}`}
          className="text-primary underline-offset-4 hover:underline"
        >
          {record.application}
        </Link>
      ),
    },
    {
      type: "text",
      name: "type",
      labelKey: "general:Type",
      when: ({record}) => isNonEmpty(record.type),
      disabled: () => true,
    },
    {
      type: "text",
      name: "clientIp",
      labelKey: "general:Client IP",
      when: ({record}) => isNonEmpty(record.clientIp),
      disabled: () => true,
    },
    {
      type: "text",
      name: "userAgent",
      labelKey: "general:User agent",
      when: ({record}) => isNonEmpty(record.userAgent),
      disabled: () => true,
    },
  ];

  return (
    <SimpleEditPage
      titleKey="entry:Edit Entry"
      backTo="/entries"
      deps={[organizationName, entryName]}
      fields={fields}
      fetch={() => EntryBackend.getEntry(organizationName, entryName)}
      add={(record) => EntryBackend.addEntry(record)}
      update={(record) => EntryBackend.updateEntry(organizationName, entryName, record)}
      editUrl={(record) => `/entries/${record.owner}/${record.name}`}
    >
      {({record}) => <EntryMessageViewer entry={record} />}
    </SimpleEditPage>
  );
}
