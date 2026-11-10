import * as React from "react";
import i18next from "i18next";
import {useParams} from "react-router-dom";
import {EditionNotice, useFeatureMissing} from "@/components/common/EditionNotice";
import {SearchableSelect} from "@/components/common/SearchableSelect";
import {EditableTable} from "@/components/crud/EditableTable";
import {Input} from "@/components/ui/input";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {
  useApplicationOptions,
  useGroupOptions,
  useOrganizationOptions,
  useRoleOptions,
  useUserOptions,
} from "@/hooks/use-options";
import * as AdapterBackend from "@/backend/AdapterBackend";
import * as AgentBackend from "@/backend/AgentBackend";
import * as ModelBackend from "@/backend/ModelBackend";
import * as PermissionBackend from "@/backend/PermissionBackend";
import * as ServerBackend from "@/backend/ServerBackend";
import {
  enumOptions,
  PERMISSION_ACTIONS,
  PERMISSION_API_ACTIONS,
  PERMISSION_CONDITION_OPERATORS,
  PERMISSION_EFFECTS,
  PERMISSION_MCP_ACTIONS,
  PERMISSION_RESOURCE_TYPES,
  PERMISSION_STATES,
} from "@/lib/enum-labels";
import * as Setting from "@/lib/setting";


export default function PermissionEditPage() {
  const {organizationName = "", permissionName = ""} = useParams();
  const {account} = useAccount();
  const rbacMissing = useFeatureMissing("rbac");
  const abacMissing = useFeatureMissing("abac");
  const organizations = useOrganizationOptions();
  const users = useUserOptions(organizationName);
  const groups = useGroupOptions(organizationName);
  const roles = useRoleOptions(organizationName);
  const applications = useApplicationOptions(organizationName);
  // the backend resolves both through GetModel()/GetAdapter(), which take "owner/name"
  const [models, setModels] = React.useState<any[]>([]);
  const [adapters, setAdapters] = React.useState<any[]>([]);
  const [servers, setServers] = React.useState<any[]>([]);
  const [agents, setAgents] = React.useState<any[]>([]);

  React.useEffect(() => {
    if (!organizationName) {
      return;
    }
    ModelBackend.getModels(organizationName, 1, 1000).then((res: any) => {
      if (res.status === "ok") {
        setModels(res.data ?? []);
      }
    });
    AdapterBackend.getAdapters(organizationName, 1, 1000).then((res: any) => {
      if (res.status === "ok") {
        setAdapters(res.data ?? []);
      }
    });
    ServerBackend.getServers(organizationName, 1, 1000).then((res: any) => {
      if (res.status === "ok") {
        setServers(res.data ?? []);
      }
    });
    AgentBackend.getAgents(organizationName, 1, 1000).then((res: any) => {
      if (res.status === "ok") {
        setAgents(res.data ?? []);
      }
    });
  }, [organizationName]);

  /** "<server>/<tool>" of the MCP servers of the organization, with "*" for all tools of a server */
  const mcpToolOptions = () => {
    const options = [{value: "*", label: i18next.t("general:All")}];
    for (const server of servers) {
      options.push({value: `${server.name}/*`, label: `${server.name}/*`});
      for (const tool of server.tools ?? []) {
        options.push({value: `${server.name}/${tool.name}`, label: `${server.name}/${tool.name}`});
      }
    }
    return options;
  };
  const isMcpTool = (ctx: any) => ctx.record.resourceType === "MCP tool";

  const toIdOptions = (items: any[]) =>
    items.map((item) => ({value: `${item.owner}/${item.name}`, label: `${item.owner}/${item.name}`}));

  /** the Casbin model decides whether roles and domains mean anything for this permission */
  const modelTextOf = (modelId: string) =>
    models.find((item) => `${item.owner}/${item.name}` === modelId)?.modelText ?? "";
  const hasRoleDefinition = (modelId: string) => modelTextOf(modelId).includes("role_definition");
  const hasDomainDefinition = (modelId: string) => {
    const match = modelTextOf(modelId).match(/request_definition\s*\]\s*r\s*=\s*([^\r\n]+)/);
    return match ? match[1].split(",").map((token: string) => token.trim()).includes("dom") : false;
  };

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
    {type: "text", name: "description", labelKey: "general:Description"},
    {type: "multiselect", name: "users", labelKey: "role:Sub users", options: () => users},
    {type: "multiselect", name: "groups", labelKey: "role:Sub groups", options: () => groups},
    {
      type: "multiselect",
      name: "roles",
      labelKey: "role:Sub roles",
      options: () => roles,
      disabled: (ctx) => !hasRoleDefinition(ctx.record.model),
    },
    {
      type: "custom",
      name: "edition",
      labelKey: "system:License",
      when: (ctx) => rbacMissing && ((ctx.record.groups ?? []).length > 0 || (ctx.record.roles ?? []).length > 0),
      render: () => <EditionNotice message={i18next.t("permission:Granting by groups and roles needs a commercial or enterprise edition license")} />,
    },
    {
      type: "multiselect",
      name: "agents",
      labelKey: "general:Agents",
      when: isMcpTool,
      options: () => [
        {value: "*", label: i18next.t("general:All")},
        ...agents.map((agent) => ({value: `${agent.owner}/${agent.name}`, label: agent.displayName ? `${agent.displayName} (${agent.owner}/${agent.name})` : `${agent.owner}/${agent.name}`})),
      ],
    },
    {
      type: "tags",
      name: "domains",
      labelKey: "role:Sub domains",
      disabled: (ctx) => !hasDomainDefinition(ctx.record.model),
    },
    {type: "select", name: "model", labelKey: "general:Model", options: () => toIdOptions(models)},
    {type: "select", name: "adapter", labelKey: "general:Adapter", options: () => toIdOptions(adapters)},
    {
      type: "select",
      name: "resourceType",
      labelKey: "permission:Resource type",
      options: () => enumOptions(PERMISSION_RESOURCE_TYPES),
    },
    {
      type: "multiselect",
      name: "resources",
      labelKey: "general:Resources",
      // antd only lets a "Custom" permission invent its own resource names
      creatable: (ctx) => ctx.record.resourceType === "Custom" || isMcpTool(ctx),
      // an API permission is scoped to backend paths, an MCP tool one to the tools of the MCP servers; everything else to applications
      options: (ctx) =>
        ctx.record.resourceType === "API"
          ? Setting.getApiPaths().map((path: string) => ({value: path, label: path}))
          : isMcpTool(ctx) ? mcpToolOptions() : [{value: "*", label: i18next.t("general:All")}, ...applications],
    },
    {
      type: "multiselect",
      name: "actions",
      labelKey: "permission:Actions",
      creatable: (ctx) => ctx.record.resourceType === "Custom",
      options: (ctx) =>
        enumOptions(ctx.record.resourceType === "API" ? PERMISSION_API_ACTIONS : isMcpTool(ctx) ? PERMISSION_MCP_ACTIONS : PERMISSION_ACTIONS),
    },
    {
      type: "custom",
      name: "conditions",
      labelKey: "permission:Argument conditions",
      block: true,
      when: isMcpTool,
      render: (ctx, update) => (
        <div className="space-y-2">
          <EditableTable
            rows={ctx.record.conditions ?? []}
            onChange={(rows) => update("conditions", rows)}
            newRow={() => ({argument: "", operator: "Equals", value: ""})}
            reorderable={false}
            columns={[
              {
                key: "argument",
                title: i18next.t("permission:Argument"),
                width: 220,
                render: (row: any, _index, patch) => <Input placeholder="factory" value={row.argument ?? ""} onChange={(e) => patch({argument: e.target.value})} />,
              },
              {
                key: "operator",
                title: i18next.t("rule:Operator"),
                width: 170,
                render: (row: any, _index, patch) => (
                  <SearchableSelect value={row.operator ?? "Equals"} onChange={(value) => patch({operator: value})} options={enumOptions(PERMISSION_CONDITION_OPERATORS)} />
                ),
              },
              {
                key: "value",
                title: i18next.t("webhook:Value"),
                render: (row: any, _index, patch) => <Input placeholder="${user.groupNames}" value={row.value ?? ""} onChange={(e) => patch({value: e.target.value})} />,
              },
            ]}
          />
          <p className="text-xs text-muted-foreground">
            {i18next.t("permission:The permission applies only to the calls whose arguments pass every condition. The argument is a dotted path in the tool arguments, and the value a comma-separated list that may use ${user.name}, ${user.owner}, ${user.groups}, ${user.groupNames}, ${user.affiliation} or ${user.properties.<key>} of the calling user")}
          </p>
          {abacMissing && (ctx.record.conditions ?? []).length > 0 ? (
            <EditionNotice message={i18next.t("permission:Argument conditions need a commercial or enterprise edition license")} />
          ) : null}
        </div>
      ),
    },
    {
      type: "select",
      name: "effect",
      labelKey: "permission:Effect",
      options: () => enumOptions(PERMISSION_EFFECTS),
    },
    {type: "switch", name: "isEnabled", labelKey: "general:Is enabled"},
    {type: "text", name: "submitter", labelKey: "permission:Submitter", disabled: () => true},
    {type: "text", name: "approver", labelKey: "permission:Approver", disabled: () => true},
    {type: "text", name: "approveTime", labelKey: "permission:Approve time", disabled: () => true},
    {type: "text", name: "expireTime", labelKey: "general:Expire time"},
    {
      type: "select",
      name: "state",
      labelKey: "general:State",
      options: () => enumOptions(PERMISSION_STATES),
      disabled: () => !Setting.isLocalAdminUser(account),
    },
  ];

  return (
    <SimpleEditPage
      beforeSave={(permission) => {
        // the same checks the antd page runs before it POSTs
        if (permission.resourceType === "MCP tool") {
          if (["users", "groups", "roles", "agents"].every((key) => (permission[key]?.length ?? 0) === 0)) {
            Setting.showMessage("error", i18next.t("permission:Choose the users, groups, roles or agents that the permission is for"));
            return null;
          }
        } else if ((permission.users?.length ?? 0) === 0 && (permission.roles?.length ?? 0) === 0) {
          Setting.showMessage("error", i18next.t("general:The users and roles cannot be empty at the same time"));
          return null;
        }
        if ((permission.resources?.length ?? 0) === 0) {
          Setting.showMessage("error", i18next.t("general:The resources cannot be empty"));
          return null;
        }
        if ((permission.actions?.length ?? 0) === 0) {
          Setting.showMessage("error", i18next.t("general:The actions cannot be empty"));
          return null;
        }
        if (!Setting.isLocalAdminUser(account) && permission.submitter !== account?.name) {
          Setting.showMessage("error", i18next.t("general:A normal user can only modify the permission submitted by itself"));
          return null;
        }
        return permission;
      }}
      titleKey="permission:Edit Permission"
      backTo="/permissions"
      deps={[organizationName, permissionName]}
      fields={fields}
      fetch={() => PermissionBackend.getPermission(organizationName, permissionName)}
      add={(record) => PermissionBackend.addPermission(record)}
      update={(record) => PermissionBackend.updatePermission(organizationName, permissionName, record)}
      editUrl={(record) => `/permissions/${record.owner}/${record.name}`}
    />
  );
}
