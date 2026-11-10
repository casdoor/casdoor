import * as React from "react";
import i18next from "i18next";
import {useParams} from "react-router-dom";
import {Badge} from "@/components/ui/badge";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table";
import {Loading} from "@/components/common/Loading";
import {EditionNotice} from "@/components/common/EditionNotice";
import {SearchableSelect} from "@/components/common/SearchableSelect";
import {PageHeader} from "@/components/crud/PageHeader";
import {useOrganizationFilter} from "@/hooks/use-organization";
import * as ServerBackend from "@/backend/ServerBackend";
import * as Setting from "@/lib/setting";

const stateVariant = (state: string) =>
  state === "Pending" ? "warning" : state === "Approved" || state === "Used" ? "success" : state === "Rejected" ? "destructive" : "secondary";

const stateText = (state: string) => {
  switch (state) {
  case "Pending":
    return i18next.t("webhook:Pending");
  case "Approved":
    return i18next.t("permission:Approved");
  case "Rejected":
    return i18next.t("pam:Rejected");
  case "Used":
    return i18next.t("server:Used");
  case "Expired":
    return i18next.t("subscription:Expired");
  default:
    return state;
  }
};

/** The tool calls of agents that wait for a person's approval: approving lets the same call through once within an hour. */
export default function McpApprovalListPage() {
  const params = useParams();
  const requestOrganization = useOrganizationFilter();
  const organizationName = params.organizationName ?? requestOrganization;
  const [approvals, setApprovals] = React.useState<any[] | null>(null);
  const [state, setState] = React.useState(params.approvalName ? "" : "Pending");
  const [comments, setComments] = React.useState<Record<string, string>>({});
  const [version, setVersion] = React.useState(0);

  React.useEffect(() => {
    if (params.approvalName) {
      ServerBackend.getMcpApproval(organizationName, params.approvalName).then((res: any) => {
        if (res.status === "ok") {
          setApprovals(res.data ? [res.data] : []);
        } else {
          setApprovals([]);
          Setting.showMessage("error", res.msg);
        }
      });
      return;
    }
    ServerBackend.getMcpApprovals(organizationName, state).then((res: any) => {
      if (res.status === "ok") {
        setApprovals(res.data ?? []);
      } else {
        setApprovals([]);
        Setting.showMessage("error", res.msg);
      }
    });
  }, [organizationName, params.approvalName, state, version]);

  const decide = (approval: any, approved: boolean) => {
    ServerBackend.decideMcpApproval(approval.owner, approval.name, approved, comments[approval.name] ?? "").then((res: any) => {
      if (res.status === "ok") {
        Setting.showMessage("success", approved ? i18next.t("server:Approved, the agent can call the tool again") : i18next.t("pam:Rejected"));
        setVersion((v) => v + 1);
      } else {
        Setting.showMessage("error", res.msg);
      }
    });
  };

  const stateOptions = ["", "Pending", "Approved", "Used", "Rejected"].map((value) => ({
    value,
    label: value === "" ? i18next.t("general:All") : stateText(value),
  }));

  return (
    <div className="space-y-4">
      <PageHeader
        title={i18next.t("server:Tool call approvals")}
        description={i18next.t("server:Tool call approvals - Tooltip")}
        actions={params.approvalName ? null : (
          <div className="w-[180px]">
            <SearchableSelect value={state} onChange={setState} options={stateOptions} />
          </div>
        )}
      />
      <EditionNotice edition="Enterprise" />
      {approvals === null ? (
        <Loading />
      ) : (
        <div className="overflow-x-auto rounded-lg border">
          <Table>
            <TableHeader>
              <TableRow className="hover:bg-transparent">
                <TableHead className="w-[160px]">{i18next.t("general:Created time")}</TableHead>
                <TableHead className="w-[150px]">{i18next.t("general:User")}</TableHead>
                <TableHead className="w-[150px]">{i18next.t("server:Agent")}</TableHead>
                <TableHead className="w-[180px]">{i18next.t("general:Tool")}</TableHead>
                <TableHead>{i18next.t("server:Arguments")}</TableHead>
                <TableHead className="w-[100px]">{i18next.t("general:State")}</TableHead>
                <TableHead className="w-[320px]">{i18next.t("general:Action")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {approvals.length === 0 ? (
                <TableRow className="hover:bg-transparent">
                  <TableCell colSpan={7} className="h-20 text-center text-muted-foreground">{i18next.t("general:No data")}</TableCell>
                </TableRow>
              ) : approvals.map((approval) => (
                <TableRow key={approval.name}>
                  <TableCell className="whitespace-nowrap">{new Date(approval.createdTime).toLocaleString()}</TableCell>
                  <TableCell className="break-all">{approval.user}</TableCell>
                  <TableCell className="break-all">{approval.agent || approval.application}</TableCell>
                  <TableCell className="break-all">{`${approval.server} / ${approval.tool}`}</TableCell>
                  <TableCell>
                    <pre className="max-h-40 max-w-[420px] overflow-auto whitespace-pre-wrap break-all font-mono text-xs">{formatArguments(approval.arguments)}</pre>
                    <div className="mt-1 font-mono text-xs text-muted-foreground">{`${i18next.t("entry:Trace ID")}: ${approval.traceId}`}</div>
                  </TableCell>
                  <TableCell>
                    <Badge variant={stateVariant(approval.state) as any} title={approval.approver ? `${approval.approver} ${approval.comment ?? ""}` : undefined}>
                      {stateText(approval.state)}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    {approval.state === "Pending" ? (
                      <div className="flex flex-wrap items-center gap-2">
                        <Input
                          className="w-[150px]"
                          placeholder={i18next.t("enterprise:Comment")}
                          value={comments[approval.name] ?? ""}
                          onChange={(e) => setComments({...comments, [approval.name]: e.target.value})}
                        />
                        <Button size="sm" onClick={() => decide(approval, true)}>{i18next.t("mfa:Approve")}</Button>
                        <Button size="sm" variant="outline" onClick={() => decide(approval, false)}>{i18next.t("server:Reject")}</Button>
                      </div>
                    ) : (
                      <span className="text-xs text-muted-foreground">
                        {approval.approver ? `${approval.approver} ${approval.approveTime ? new Date(approval.approveTime).toLocaleString() : ""} ${approval.comment ?? ""}` : ""}
                      </span>
                    )}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}
    </div>
  );
}

function formatArguments(text: string) {
  try {
    return JSON.stringify(JSON.parse(text), null, 2);
  } catch {
    return text;
  }
}
