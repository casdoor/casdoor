import i18next from "i18next";
import {Link, useParams} from "react-router-dom";
import {Input} from "@/components/ui/input";
import {EditableTable} from "@/components/crud/EditableTable";
import {Badge} from "@/components/ui/badge";
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
    {type: "text", name: "name", labelKey: "general:Name", required: true},
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
    {
      // read-only: membership is edited from the user page, as in the antd frontend
      type: "custom",
      name: "users",
      labelKey: "general:Users",
      render: (ctx) => {
        const users: string[] = ctx.record.users ?? [];
        if (users.length === 0) {
          return <span className="text-sm text-muted-foreground">-</span>;
        }
        return (
          <div className="flex flex-wrap gap-1">
            {users.map((user) => (
              <Link key={user} to={`/users/${user}`}>
                <Badge variant="secondary" className="font-normal hover:bg-secondary/70">{user}</Badge>
              </Link>
            ))}
          </div>
        );
      },
    },
    {type: "text", name: "contactEmail", labelKey: "general:Email"},
    {type: "number", name: "gidNumber", labelKey: "general:GID number"},
    {
      type: "custom",
      name: "properties",
      labelKey: "user:Properties",
      block: true,
      render: (ctx, update) => (
        <EditableTable
          rows={Object.entries(ctx.record.properties ?? {}).map(([key, value]) => ({key, value}))}
          onChange={(rows) =>
            update(
              "properties",
              Object.fromEntries(rows.filter((row: any) => row.key).map((row: any) => [row.key, row.value])),
            )
          }
          newRow={() => ({key: "", value: ""})}
          reorderable={false}
          columns={[
            {
              key: "key",
              title: i18next.t("general:Name"),
              width: 240,
              render: (row: any, _i, patch) => (
                <Input value={row.key ?? ""} onChange={(e) => patch({key: e.target.value})} />
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
