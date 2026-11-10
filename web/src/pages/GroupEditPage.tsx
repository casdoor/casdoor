import * as React from "react";
import i18next from "i18next";
import {Link, useLocation, useParams} from "react-router-dom";
import {MapTable} from "@/components/crud/MapTable";
import {Badge} from "@/components/ui/badge";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useGroupOptions, useOrganizationOptions, useUserNameOptions} from "@/hooks/use-options";
import * as GroupBackend from "@/backend/GroupBackend";
import * as Setting from "@/lib/setting";
import {isDelegatedAdmin, isDelegatedRootGroup} from "@/lib/delegated-admin";

export default function GroupEditPage() {
  const {organizationName = "", groupName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();
  const groups = useGroupOptions(organizationName);
  const users = useUserNameOptions(organizationName);
  // a delegated admin keeps their own top groups where they are and puts the others under the groups they administer
  const delegated = isDelegatedAdmin(account);
  const location = useLocation();
  // the group tree hands over where to return to; kept in a ref because "Save"
  // re-navigates to the saved name and drops the router state
  const backTo = React.useRef<string>((location.state as {backTo?: string} | null)?.backTo ?? "/groups");

  const fields: EditField[] = [
    {
      type: "select",
      name: "owner",
      labelKey: "general:Organization",
      options: () => organizations,
      disabled: () => !Setting.isAdminUser(account),
    },
    {type: "text", name: "name", labelKey: "general:Name", required: true, disabled: (ctx) => delegated && ctx.mode !== "add"},
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
      disabled: (ctx) => delegated && isDelegatedRootGroup(account, ctx.record.name),
      options: (ctx) => {
        const parents = groups
          .filter((option) => option.value !== `${ctx.record.owner}/${ctx.record.name}`)
          .map((option) => ({...option, value: option.value.split("/")[1]}));
        if (!delegated) {
          return [{value: ctx.record.owner, label: ctx.record.owner}, ...parents];
        }
        return isDelegatedRootGroup(account, ctx.record.name) ? [{value: ctx.record.parentId, label: ctx.record.parentId}] : parents;
      },
    },
    {
      type: "select",
      name: "manager",
      labelKey: "group:Manager",
      options: () => [{value: "", label: "-"}, ...users],
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
        <MapTable
          value={ctx.record.properties}
          onChange={(value) => update("properties", value)}
          keyTitle={i18next.t("general:Name")}
          valueTitle={i18next.t("webhook:Value")}
          keyWidth={240}
        />
      ),
    },
    {type: "switch", name: "isEnabled", labelKey: "general:Is enabled"},
  ];

  return (
    <SimpleEditPage
      titleKey="group:Edit Group"
      backTo={backTo.current}
      deps={[organizationName, groupName]}
      fields={fields}
      fetch={() => GroupBackend.getGroup(organizationName, groupName)}
      add={(record) => GroupBackend.addGroup(record)}
      update={(record) => GroupBackend.updateGroup(organizationName, groupName, record)}
      editUrl={(record) => `/groups/${record.owner}/${record.name}`}
    />
  );
}
