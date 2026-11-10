import * as React from "react";
import i18next from "i18next";
import {Badge} from "@/components/ui/badge";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table";
import * as ServerBackend from "@/backend/ServerBackend";

/** The recent tool calls through the MCP gateway, recorded when the organization's security policy asks so. */
export function McpRecordTable({owner, server}: {owner: string; server: string}) {
  const [records, setRecords] = React.useState<any[] | null>(null);
  const [error, setError] = React.useState("");
  const [traceId, setTraceId] = React.useState("");

  const load = React.useCallback((trace = "") => {
    ServerBackend.getMcpRecords(owner, server, 1, 50, {traceId: trace}).then((res: any) => {
      if (res.status === "ok") {
        setRecords(res.data ?? []);
        setError("");
      } else {
        setRecords([]);
        setError(res.msg);
      }
    });
  }, [owner, server]);

  React.useEffect(() => {
    load();
  }, [load]);

  const variant = (result: string) =>
    result === "Success" ? "success" : result === "Blocked" || result === "Pending approval" ? "warning" : "destructive";

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
              load(traceId.trim());
            }
          }}
        />
        <Button variant="outline" size="sm" onClick={() => load(traceId.trim())}>{i18next.t("general:Search")}</Button>
        <Button variant="outline" size="sm" onClick={() => {
          setTraceId("");
          load();
        }}>{i18next.t("general:Refresh")}</Button>
      </div>
      <div className="overflow-x-auto rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow className="hover:bg-transparent">
              <TableHead className="w-[170px]">{i18next.t("general:Created time")}</TableHead>
              <TableHead className="w-[150px]">{i18next.t("general:User")}</TableHead>
              <TableHead className="w-[150px]">{i18next.t("server:Agent")}</TableHead>
              <TableHead className="w-[150px]">{i18next.t("general:Tool")}</TableHead>
              <TableHead>{i18next.t("server:Arguments")}</TableHead>
              <TableHead className="w-[110px]">{i18next.t("payment:Result")}</TableHead>
              <TableHead className="w-[80px]">{i18next.t("entry:Duration")}</TableHead>
              <TableHead className="w-[150px]">{i18next.t("entry:Trace ID")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {!records || records.length === 0 ? (
              <TableRow className="hover:bg-transparent">
                <TableCell colSpan={8} className="h-20 text-center text-muted-foreground">
                  {error || i18next.t("general:No data")}
                </TableCell>
              </TableRow>
            ) : (
              records.map((record) => (
                <TableRow key={record.id}>
                  <TableCell className="whitespace-nowrap">{new Date(record.createdTime).toLocaleString()}</TableCell>
                  <TableCell className="break-all">{record.user}</TableCell>
                  <TableCell className="break-all" title={record.actor}>{record.agent || record.application}</TableCell>
                  <TableCell className="break-all">{record.tool}</TableCell>
                  <TableCell className="max-w-[320px] truncate font-mono text-xs" title={record.arguments}>{record.arguments}</TableCell>
                  <TableCell>
                    <Badge variant={variant(record.result) as any} title={record.response}>{record.result}</Badge>
                  </TableCell>
                  <TableCell>{record.durationMs ? `${record.durationMs} ms` : ""}</TableCell>
                  <TableCell className="max-w-[150px] truncate font-mono text-xs">
                    {record.traceId ? (
                      <button type="button" className="underline-offset-2 hover:underline" title={record.traceId} onClick={() => {
                        setTraceId(record.traceId);
                        load(record.traceId);
                      }}>{record.traceId}</button>
                    ) : null}
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
