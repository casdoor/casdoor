// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

import * as React from "react";
import i18next from "i18next";
import {Badge} from "@/components/ui/badge";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table";
import {Tabs, TabsContent, TabsList, TabsTrigger} from "@/components/ui/tabs";
import {Loading} from "@/components/common/Loading";
import {ConfirmButton} from "@/components/common/ConfirmButton";
import {SearchableSelect} from "@/components/common/SearchableSelect";
import {EditionNotice, useEdition} from "@/components/common/EditionNotice";
import {PageHeader} from "@/components/crud/PageHeader";
import {McpRecordTable} from "@/ee/components/server/McpRecordTable";
import {useAccount} from "@/hooks/use-account";
import {useOrganizationOptions} from "@/hooks/use-options";
import * as McpChangeBackend from "@/ee/backend/McpChangeBackend";
import * as Setting from "@/lib/setting";

// the records and approvals of Casdoor's own MCP server are kept under this server name
const BuiltinServer = "casdoor";

const actionText = (action: string) => {
  switch (action) {
  case "Add":
    return i18next.t("general:Add");
  case "Update":
    return i18next.t("general:Edit");
  case "Delete":
    return i18next.t("general:Delete");
  default:
    return action;
  }
};

const actionVariant = (action: string) => (action === "Delete" ? "destructive" : action === "Add" ? "success" : "secondary");

/** What agents changed through Casdoor's own MCP server: who, with which tool, the fields before and after, and rolling a change back. */
function ChangeTable({owner}: {owner: string}) {
  const edition = useEdition();
  const canRollback = edition === "Enterprise";
  const [changes, setChanges] = React.useState<any[] | null>(null);
  const [error, setError] = React.useState("");
  const [traceId, setTraceId] = React.useState("");
  const [state, setState] = React.useState("");
  const [version, setVersion] = React.useState(0);

  React.useEffect(() => {
    setChanges(null);
    McpChangeBackend.getMcpChanges(owner, 1, 100, {traceId: traceId.trim(), state}).then((res: any) => {
      if (res.status === "ok") {
        setChanges(res.data ?? []);
        setError("");
      } else {
        setChanges([]);
        setError(res.msg);
      }
    });
    // the trace ID is searched on Enter or with the button
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [owner, state, version]);

  const rollback = async(change: any, wholeCall: boolean) => {
    const res: any = await McpChangeBackend.rollbackMcpChange(change.owner, change.name, wholeCall);
    if (res.status === "ok") {
      Setting.showMessage("success", i18next.t("system:Rolled back"));
    } else {
      Setting.showMessage("error", res.msg);
    }
    setVersion((v) => v + 1);
  };

  const stateOptions = [
    {value: "", label: i18next.t("general:All")},
    {value: "Applied", label: i18next.t("enterprise:Applied")},
    {value: "Rolled back", label: i18next.t("system:Rolled back")},
  ];

  return (
    <div className="space-y-2">
      <div className="flex flex-wrap items-center gap-2">
        <Input
          className="w-[320px]"
          placeholder={i18next.t("entry:Trace ID")}
          value={traceId}
          onChange={(e) => setTraceId(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter") {
              setVersion((v) => v + 1);
            }
          }}
        />
        <Button variant="outline" size="sm" onClick={() => setVersion((v) => v + 1)}>{i18next.t("general:Search")}</Button>
        <div className="w-[160px]">
          <SearchableSelect value={state} onChange={setState} options={stateOptions} />
        </div>
      </div>
      {changes === null ? (
        <Loading />
      ) : (
        <div className="overflow-x-auto rounded-lg border">
          <Table>
            <TableHeader>
              <TableRow className="hover:bg-transparent">
                <TableHead className="w-[160px]">{i18next.t("general:Created time")}</TableHead>
                <TableHead className="w-[120px]">{i18next.t("general:User")}</TableHead>
                <TableHead className="w-[120px]">{i18next.t("server:Agent")}</TableHead>
                <TableHead className="w-[160px]">{i18next.t("general:Tool")}</TableHead>
                <TableHead className="w-[200px]">{i18next.t("server:Changed object")}</TableHead>
                <TableHead className="w-[190px]">{i18next.t("general:Action")}</TableHead>
                <TableHead className="min-w-[380px]">{i18next.t("server:Changed fields")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {changes.length === 0 ? (
                <TableRow className="hover:bg-transparent">
                  <TableCell colSpan={7} className="h-20 text-center text-muted-foreground">{error || i18next.t("general:No data")}</TableCell>
                </TableRow>
              ) : changes.map((change) => {
                const sameCall = changes.filter((item) => item.traceId === change.traceId && item.state === "Applied");
                return (
                  <TableRow key={change.name}>
                    <TableCell className="whitespace-nowrap">{new Date(change.createdTime).toLocaleString()}</TableCell>
                    <TableCell className="break-all">{change.user}</TableCell>
                    <TableCell className="break-all" title={change.authMethod}>{change.agent || change.application}</TableCell>
                    <TableCell className="break-all">
                      <div>{change.tool}</div>
                      <button type="button" className="font-mono text-xs text-muted-foreground underline-offset-2 hover:underline" title={change.traceId} onClick={() => {
                        setTraceId(change.traceId);
                        setVersion((v) => v + 1);
                      }}>{change.traceId.slice(0, 12)}</button>
                    </TableCell>
                    <TableCell className="break-all">
                      <Badge variant={actionVariant(change.action) as any}>{actionText(change.action)}</Badge>{" "}
                      <span className="text-xs">{`${change.kind} ${change.objectId}`}</span>
                    </TableCell>
                    <TableCell>
                      {change.state === "Applied" ? (
                        <div className="flex flex-col items-start gap-1">
                          <ConfirmButton
                            size="sm"
                            variant="outline"
                            disabled={!canRollback}
                            title={i18next.t("server:Roll back this change") + "?"}
                            description={i18next.t("server:Roll back - Tooltip")}
                            confirmText={i18next.t("server:Roll back")}
                            onConfirm={() => rollback(change, false)}
                          >
                            {i18next.t("server:Roll back")}
                          </ConfirmButton>
                          {sameCall.length > 1 ? (
                            <ConfirmButton
                              size="sm"
                              variant="ghost"
                              className="h-auto whitespace-normal px-2 text-left"
                              disabled={!canRollback}
                              title={i18next.t("server:Roll back the whole call") + "?"}
                              description={i18next.t("server:Roll back - Tooltip")}
                              confirmText={i18next.t("server:Roll back")}
                              onConfirm={() => rollback(change, true)}
                            >
                              {`${i18next.t("server:Roll back the whole call")} (${sameCall.length})`}
                            </ConfirmButton>
                          ) : null}
                        </div>
                      ) : (
                        <span className="text-xs text-muted-foreground" title={change.rollbackTime}>
                          {`${i18next.t("system:Rolled back")} ${change.rollbackUser}`}
                        </span>
                      )}
                    </TableCell>
                    <TableCell>
                      <pre className="max-h-40 min-w-[360px] max-w-[640px] overflow-auto whitespace-pre-wrap break-all font-mono text-xs">{change.diff}</pre>
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </div>
      )}
      {canRollback ? null : <EditionNotice edition="Enterprise" message={i18next.t("server:Rolling back needs the enterprise edition")} />}
    </div>
  );
}

/** The changes and tool calls that AI agents made through Casdoor's own MCP server. */
export default function McpChangeListPage() {
  const {account} = useAccount();
  const organizations = useOrganizationOptions();
  const isGlobalAdmin = Setting.isAdminUser(account);
  const [owner, setOwner] = React.useState("");
  React.useEffect(() => {
    if (account && !owner) {
      setOwner(account.owner);
    }
  }, [account, owner]);

  return (
    <div className="space-y-4">
      <PageHeader
        title={i18next.t("server:Agent changes")}
        description={i18next.t("server:Agent changes - Tooltip")}
        actions={isGlobalAdmin ? <SearchableSelect className="w-60" value={owner} onChange={setOwner} options={organizations} /> : null}
      />
      <EditionNotice edition="Commercial" />
      {owner ? (
        <Tabs defaultValue="changes">
          <TabsList className="mb-2 flex-wrap">
            <TabsTrigger value="changes">{i18next.t("server:Changes")}</TabsTrigger>
            <TabsTrigger value="calls">{i18next.t("entry:Tool calls")}</TabsTrigger>
          </TabsList>
          <TabsContent value="changes"><ChangeTable key={owner} owner={owner} /></TabsContent>
          <TabsContent value="calls"><McpRecordTable key={owner} owner={owner} server={BuiltinServer} /></TabsContent>
        </Tabs>
      ) : null}
    </div>
  );
}
