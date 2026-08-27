import i18next from "i18next";
import {useParams} from "react-router-dom";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useGroupOptions, useOrganizationOptions} from "@/hooks/use-options";
import * as GroupBackend from "@/backend/GroupBackend";
import * as Setting from "@/lib/setting";

export default function GroupEditPage() {
  const {organizationName = "", groupName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();
  const groups = useGroupOptions(organizationName);

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
    {
      type: "select",
      name: "type",
      labelKey: "general:Type",
      options: () => [
        {value: "Virtual", label: i18next.t("group:Virtual")},
        {value: "Physical", label: i18next.t("group:Physical")},
      ],
    },
    {
      type: "select",
      name: "parentId",
      labelKey: "group:Parent group",
      options: (ctx) => [
        {value: ctx.record.owner, label: ctx.record.owner},
        ...groups
          .filter((option) => option.value !== `${ctx.record.owner}/${ctx.record.name}`)
          .map((option) => ({...option, value: option.value.split("/")[1]})),
      ],
    },
    {type: "text", name: "contactEmail", labelKey: "group:Contact email"},
    {type: "switch", name: "isEnabled", labelKey: "general:Is enabled"},
  ];

  return (
    <SimpleEditPage
      titleKey="group:Edit Group"
      backTo="/groups"
      deps={[organizationName, groupName]}
      fields={fields}
      fetch={() => GroupBackend.getGroup(organizationName, groupName)}
      add={(record) => GroupBackend.addGroup(record)}
      update={(record) => GroupBackend.updateGroup(organizationName, groupName, record)}
      editUrl={(record) => `/groups/${record.owner}/${record.name}`}
    />
  );
}
