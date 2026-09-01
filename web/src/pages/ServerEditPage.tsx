import * as React from "react";
import i18next from "i18next";
import {useNavigate, useParams} from "react-router-dom";
import {UnauthorizedPage} from "@/components/common/UnauthorizedPage";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Switch} from "@/components/ui/switch";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table";
import {Loading} from "@/components/common/Loading";
import {SearchableSelect} from "@/components/common/SearchableSelect";
import {EditPageShell} from "@/components/crud/EditPageShell";
import {FormRow} from "@/components/crud/FormRow";
import {useAccount} from "@/hooks/use-account";
import {useApplicationOptions, useOrganizationOptions} from "@/hooks/use-options";
import {useEditRecord} from "@/hooks/use-edit-record";
import * as ServerBackend from "@/backend/ServerBackend";
import {submitEdit} from "@/lib/crud";
import * as Setting from "@/lib/setting";

/** The tools the MCP server exposes; only "isAllowed" is editable, see web/src/ToolTable.js. */
function ToolTable({tools, onChange}: {tools: any[]; onChange: (tools: any[]) => void}) {
  const rows = tools ?? [];

  return (
    <div className="overflow-x-auto rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow className="hover:bg-transparent">
            <TableHead className="w-[260px]">{i18next.t("general:Name")}</TableHead>
            <TableHead>{i18next.t("general:Description")}</TableHead>
            <TableHead className="w-[120px]">{i18next.t("general:Is allowed")}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.length === 0 ? (
            <TableRow className="hover:bg-transparent">
              <TableCell colSpan={3} className="h-20 text-center text-muted-foreground">
                {i18next.t("general:No data")}
              </TableCell>
            </TableRow>
          ) : (
            rows.map((tool, index) => (
              <TableRow key={tool.name || `tool-${index}`}>
                <TableCell>{tool.name}</TableCell>
                <TableCell className="text-muted-foreground">{tool.description}</TableCell>
                <TableCell>
                  <Switch
                    checked={!!tool.isAllowed}
                    onCheckedChange={(checked) => {
                      const next = [...rows];
                      next[index] = {...next[index], isAllowed: checked};
                      onChange(next);
                    }}
                  />
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
}

export default function ServerEditPage() {
  const {organizationName = "", serverName = ""} = useParams();
  const {account} = useAccount();
  const navigate = useNavigate();
  const organizations = useOrganizationOptions();

  const [owner, setOwner] = React.useState(organizationName);
  const [name, setName] = React.useState(serverName);
  const [saving, setSaving] = React.useState(false);

  const {record: server, updateField, setRecord, loading, denied, mode, setMode, reload} = useEditRecord<any>({
    fetch: () => ServerBackend.getServer(organizationName, serverName),
    deps: [organizationName, serverName],
  });

  const applications = useApplicationOptions(server?.owner ?? owner);

  const updateServerField = (key: string, value: any) => {
    if (key === "owner") {
      // the applications belong to the organization, so the old choice no longer applies
      setRecord((prev: any) => (prev === null ? prev : {...prev, owner: value, application: ""}));
      return;
    }
    updateField(key, value);
  };

  const save = async(exitAfterSave: boolean) => {
    if (server === null) {
      return;
    }
    setSaving(true);
    await submitEdit({
      mode,
      record: Setting.deepCopy(server),
      add: (record) => ServerBackend.addServer(record),
      update: (record) => ServerBackend.updateServer(owner, name, record),
      onSaved: () => {
        setOwner(server.owner);
        setName(server.name);
        setMode("edit");
        if (exitAfterSave) {
          navigate("/servers");
          return;
        }
        const next = `/servers/${server.owner}/${server.name}`;
        if (next !== window.location.pathname) {
          navigate(next, {replace: true});
        } else {
          reload();
        }
      },
    });
    setSaving(false);
  };

  const syncMcpTool = (isCleared: boolean) => {
    if (server === null) {
      return;
    }
    ServerBackend.syncMcpTool(owner, name, Setting.deepCopy(server), isCleared)
      .then((res: any) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully modified"));
          reload();
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to update")}: ${res.msg}`);
        }
      })
      .catch((error: any) => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  };

  const getMcpAccessToken = () => {
    if (server === null) {
      return;
    }
    ServerBackend.getMcpAccessToken(server.owner, server.application).then((res: any) => {
      if (res.status === "ok") {
        updateField("token", res.data);
        Setting.showMessage("success", i18next.t("general:Successfully got"));
      } else {
        Setting.showMessage("error", `${i18next.t("general:Failed to get")}: ${res.msg}`);
      }
    });
  };

  if (denied) {
    return <UnauthorizedPage />;
  }

  if (loading || server === null) {
    return <Loading />;
  }

  return (
    <EditPageShell
      title={mode === "add" ? i18next.t("server:New MCP Server") : i18next.t("server:Edit MCP Server")}
      mode={mode}
      backTo="/servers"
      saving={saving}
      onSave={save}
    >
      <FormRow labelKey="general:Organization">
        <SearchableSelect
          disabled={!Setting.isAdminUser(account)}
          value={server.owner ?? ""}
          onChange={(value) => updateServerField("owner", value)}
          options={organizations}
        />
      </FormRow>
      <FormRow labelKey="general:Name">
        <Input value={server.name ?? ""} onChange={(e) => updateServerField("name", e.target.value)} />
      </FormRow>
      <FormRow labelKey="general:Display name">
        <Input value={server.displayName ?? ""} onChange={(e) => updateServerField("displayName", e.target.value)} />
      </FormRow>
      <FormRow labelKey="general:Description">
        <Input value={server.description ?? ""} onChange={(e) => updateServerField("description", e.target.value)} />
      </FormRow>
      <FormRow labelKey="general:URL">
        <Input value={server.url ?? ""} onChange={(e) => updateServerField("url", e.target.value)} />
      </FormRow>
      <FormRow labelKey="token:Access token">
        <div className="flex items-center gap-2">
          <Input
            type="password"
            placeholder="***"
            value={server.token ?? ""}
            onChange={(e) => updateServerField("token", e.target.value)}
          />
          <Button variant="outline" disabled={!server.application} onClick={getMcpAccessToken}>
            {i18next.t("token:Get access token")}
          </Button>
        </div>
      </FormRow>
      <FormRow block labelKey="general:Application">
        <SearchableSelect
          value={server.application ?? ""}
          onChange={(value) => updateServerField("application", value)}
          options={applications}
        />
      </FormRow>
      <FormRow labelKey="general:Tool" block>
        <div className="space-y-2">
          {mode !== "add" ? (
            <div className="flex flex-wrap gap-2">
              <Button onClick={() => syncMcpTool(false)}>{i18next.t("general:Sync")}</Button>
              <Button variant="outline" onClick={() => syncMcpTool(true)}>{i18next.t("general:Clear")}</Button>
            </div>
          ) : null}
          <ToolTable tools={server.tools ?? []} onChange={(tools) => updateServerField("tools", tools)} />
        </div>
      </FormRow>
      <FormRow labelKey="provider:Base URL">
        <Input readOnly value={`${window.location.origin}/api/server/${server.owner}/${server.name}`} />
      </FormRow>
    </EditPageShell>
  );
}
