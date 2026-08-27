import {useParams} from "react-router-dom";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useOrganizationOptions} from "@/hooks/use-options";
import * as EntryBackend from "@/backend/EntryBackend";
import * as Setting from "@/lib/setting";

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
    {type: "text", name: "name", labelKey: "general:Name"},
    {type: "text", name: "displayName", labelKey: "general:Display name"},
    {type: "text", name: "provider", labelKey: "general:Provider", disabled: () => true},
    {type: "text", name: "type", labelKey: "general:Type", disabled: () => true},
    {type: "text", name: "clientIp", labelKey: "general:Client IP", disabled: () => true},
    {type: "text", name: "userAgent", labelKey: "general:User agent", disabled: () => true},
    {type: "textarea", name: "message", labelKey: "payment:Message", rows: 8, disabled: () => true},
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
    />
  );
}
