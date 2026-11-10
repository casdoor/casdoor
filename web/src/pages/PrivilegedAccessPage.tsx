import * as React from "react";
import i18next from "i18next";
import {Copy, Eye, KeyRound} from "lucide-react";
import {Tabs, TabsContent, TabsList, TabsTrigger} from "@/components/ui/tabs";
import {Badge} from "@/components/ui/badge";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Textarea} from "@/components/ui/textarea";
import {Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle} from "@/components/ui/dialog";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table";
import {PageHeader} from "@/components/crud/PageHeader";
import {Loading} from "@/components/common/Loading";
import {EditionNotice} from "@/components/common/EditionNotice";
import {useAccount} from "@/hooks/use-account";
import * as PrivilegedAccountBackend from "@/backend/PrivilegedAccountBackend";
import {EnumBadge, enumLabel} from "@/lib/enum-labels";
import {PAM_ACCOUNT_TYPES, PAM_REQUEST_STATES} from "@/lib/pam";
import * as Setting from "@/lib/setting";

/** the revealed password is hidden again after this many seconds */
const revealSeconds = 60;

function useNow() {
  const [now, setNow] = React.useState(Date.now());
  React.useEffect(() => {
    const timer = window.setInterval(() => setNow(Date.now()), 1000);
    return () => window.clearInterval(timer);
  }, []);
  return now;
}

function formatLeft(expireTime: string, now: number) {
  const left = Math.max(0, Math.floor((new Date(expireTime).getTime() - now) / 1000));
  const hours = Math.floor(left / 3600);
  const minutes = Math.floor((left % 3600) / 60);
  const seconds = left % 60;
  const pad = (n: number) => String(n).padStart(2, "0");
  return hours > 0 ? `${hours}:${pad(minutes)}:${pad(seconds)}` : `${pad(minutes)}:${pad(seconds)}`;
}

function getTarget(request: any) {
  return request.accountDisplayName || request.account;
}

function TimeLeft({request, now}: {request: any; now: number}) {
  if (request.state === "Active") {
    return <span className="font-mono tabular-nums">{formatLeft(request.expireTime, now)}</span>;
  }
  if (request.state === "Approved") {
    return (
      <span className="text-xs text-muted-foreground">
        {i18next.t("pam:Starts when you reveal the password, check out within")} <span className="font-mono tabular-nums">{formatLeft(request.expireTime, now)}</span>
      </span>
    );
  }
  return <span className="text-muted-foreground">-</span>;
}

function RequestDialog({target, onClose, onDone}: {target: any; onClose: () => void; onDone: () => void}) {
  const [reason, setReason] = React.useState("");
  const [minutes, setMinutes] = React.useState(Math.min(60, target.maxCheckoutMinutes));
  const [busy, setBusy] = React.useState(false);

  const submit = async() => {
    setBusy(true);
    const res = await PrivilegedAccountBackend.requestPrivilegedAccess(target.name, reason, minutes);
    setBusy(false);
    if (res.status === "ok") {
      Setting.showMessage("success", res.data.state === "Active" ? i18next.t("pam:Access granted") : i18next.t("pam:Request sent to the approvers"));
      onDone();
    } else {
      Setting.showMessage("error", res.msg);
    }
  };

  return (
    <Dialog open onOpenChange={(open) => (open ? undefined : onClose())}>
      <DialogContent className="sm:max-w-[520px]">
        <DialogHeader>
          <DialogTitle>{i18next.t("pam:Request access")}: {target.displayName || target.name}</DialogTitle>
          <DialogDescription>
            {target.requireApproval ? i18next.t("pam:An approver has to approve the request") : i18next.t("pam:The access starts right away")}
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-3">
          <div className="space-y-1">
            <div className="text-sm font-medium">{i18next.t("rule:Reason")}</div>
            <Textarea rows={3} value={reason} onChange={(e) => setReason(e.target.value)} />
          </div>
          <div className="space-y-1">
            <div className="text-sm font-medium">{i18next.t("pam:Duration in minutes")}</div>
            <Input type="number" min={1} max={target.maxCheckoutMinutes} value={minutes} onChange={(e) => setMinutes(Number(e.target.value))} />
            <div className="text-xs text-muted-foreground">{i18next.t("pam:At most %s minutes").replace("%s", String(target.maxCheckoutMinutes))}</div>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>{i18next.t("general:Cancel")}</Button>
          <Button loading={busy} disabled={!reason.trim() || minutes <= 0} onClick={submit}>{i18next.t("general:Request")}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function RevealDialog({revealed, onClose}: {revealed: any; onClose: () => void}) {
  const now = useNow();
  const [shownAt] = React.useState(Date.now());
  const left = revealSeconds - Math.floor((now - shownAt) / 1000);
  React.useEffect(() => {
    if (left <= 0) {
      onClose();
    }
  }, [left, onClose]);

  return (
    <Dialog open onOpenChange={(open) => (open ? undefined : onClose())}>
      <DialogContent className="sm:max-w-[520px]">
        <DialogHeader>
          <DialogTitle>{getTarget(revealed.request)}</DialogTitle>
          <DialogDescription>{i18next.t("pam:Every reveal is recorded in the audit records")}</DialogDescription>
        </DialogHeader>
        <dl className="grid grid-cols-[120px_1fr_auto] items-center gap-x-3 gap-y-2 text-sm">
          <dt className="text-muted-foreground">{i18next.t("signup:Username")}</dt>
          <dd className="break-all font-mono">{revealed.username}</dd>
          <dd><Button variant="ghost" size="sm" onClick={() => Setting.copyToClipboard(revealed.username)}><Copy className="size-4" /></Button></dd>
          <dt className="text-muted-foreground">{i18next.t("general:Password")}</dt>
          <dd className="break-all font-mono">{revealed.password}</dd>
          <dd><Button variant="ghost" size="sm" onClick={() => Setting.copyToClipboard(revealed.password)}><Copy className="size-4" /></Button></dd>
        </dl>
        <div className="text-xs text-muted-foreground">{i18next.t("pam:Hidden again in %s seconds").replace("%s", String(Math.max(0, left)))}</div>
      </DialogContent>
    </Dialog>
  );
}

function DecideDialog({request, approve, onClose, onDone}: {request: any; approve: boolean; onClose: () => void; onDone: () => void}) {
  const [comment, setComment] = React.useState("");
  const [busy, setBusy] = React.useState(false);

  const submit = async() => {
    setBusy(true);
    const res = approve
      ? await PrivilegedAccountBackend.approvePrivilegedRequest(request, comment)
      : await PrivilegedAccountBackend.rejectPrivilegedRequest(request, comment);
    setBusy(false);
    if (res.status === "ok") {
      Setting.showMessage("success", i18next.t("general:Successfully saved"));
      onDone();
    } else {
      Setting.showMessage("error", res.msg);
    }
  };

  return (
    <Dialog open onOpenChange={(open) => (open ? undefined : onClose())}>
      <DialogContent className="sm:max-w-[520px]">
        <DialogHeader>
          <DialogTitle>{approve ? i18next.t("mfa:Approve") : i18next.t("pam:Reject request")}: {getTarget(request)}</DialogTitle>
          <DialogDescription>{request.requester}: {request.reason}</DialogDescription>
        </DialogHeader>
        <div className="space-y-1">
          <div className="text-sm font-medium">{i18next.t("pam:Comment to the requester")}</div>
          <Textarea rows={3} value={comment} onChange={(e) => setComment(e.target.value)} />
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>{i18next.t("general:Cancel")}</Button>
          <Button variant={approve ? "default" : "destructive"} loading={busy} onClick={submit}>
            {approve ? i18next.t("mfa:Approve") : i18next.t("pam:Reject request")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function AccountsTab({accounts, onRequest}: {accounts: any[]; onRequest: (account: any) => void}) {
  if (accounts.length === 0) {
    return <p className="text-sm text-muted-foreground">{i18next.t("pam:There is nothing you can request")}</p>;
  }
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>{i18next.t("general:Name")}</TableHead>
          <TableHead>{i18next.t("general:Type")}</TableHead>
          <TableHead>{i18next.t("pam:Target system")}</TableHead>
          <TableHead>{i18next.t("pam:Require approval")}</TableHead>
          <TableHead>{i18next.t("pam:Max checkout minutes")}</TableHead>
          <TableHead className="w-28" />
        </TableRow>
      </TableHeader>
      <TableBody>
        {accounts.map((account) => (
          <TableRow key={account.name}>
            <TableCell>
              <div className="font-medium">{account.displayName || account.name}</div>
              {account.description ? <div className="text-xs text-muted-foreground">{account.description}</div> : null}
            </TableCell>
            <TableCell>{enumLabel(PAM_ACCOUNT_TYPES, account.type)}</TableCell>
            <TableCell className="font-mono text-xs">
              {account.type === "Role" ? account.role : `${account.username}@${account.host}${account.port ? `:${account.port}` : ""}`}
            </TableCell>
            <TableCell>{account.requireApproval ? i18next.t("pam:Approval needed") : i18next.t("pam:No approval needed")}</TableCell>
            <TableCell>{account.maxCheckoutMinutes}</TableCell>
            <TableCell>
              <Button size="sm" onClick={() => onRequest(account)}>
                <KeyRound className="size-4" />
                {i18next.t("general:Request")}
              </Button>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}

function MyRequestsTab({requests, now, reload, onReveal}: {requests: any[]; now: number; reload: () => void; onReveal: (revealed: any) => void}) {
  const [busy, setBusy] = React.useState("");

  const run = async(request: any, action: () => Promise<any>, onOk?: (res: any) => void) => {
    setBusy(request.name);
    const res = await action();
    setBusy("");
    if (res.status === "ok") {
      onOk?.(res);
    } else {
      Setting.showMessage("error", res.msg);
    }
    reload();
  };

  if (requests.length === 0) {
    return <p className="text-sm text-muted-foreground">{i18next.t("pam:You have no requests yet")}</p>;
  }
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>{i18next.t("pam:Target system")}</TableHead>
          <TableHead>{i18next.t("general:State")}</TableHead>
          <TableHead>{i18next.t("rule:Reason")}</TableHead>
          <TableHead>{i18next.t("pam:Duration in minutes")}</TableHead>
          <TableHead>{i18next.t("pam:Time left")}</TableHead>
          <TableHead>{i18next.t("permission:Approver")}</TableHead>
          <TableHead>{i18next.t("general:Created time")}</TableHead>
          <TableHead className="w-56" />
        </TableRow>
      </TableHeader>
      <TableBody>
        {requests.map((request) => {
          const open = request.state === "Active" || request.state === "Approved";
          return (
            <TableRow key={request.name}>
              <TableCell>
                <div className="font-medium">{getTarget(request)}</div>
                <div className="text-xs text-muted-foreground">{enumLabel(PAM_ACCOUNT_TYPES, request.type)}{request.role ? `: ${request.role}` : ""}</div>
              </TableCell>
              <TableCell><EnumBadge map={PAM_REQUEST_STATES} value={request.state} /></TableCell>
              <TableCell className="max-w-56 break-words text-xs">{request.reason}{request.comment ? <div className="text-muted-foreground">{request.comment}</div> : null}</TableCell>
              <TableCell>{request.durationMinutes}</TableCell>
              <TableCell><TimeLeft request={request} now={now} /></TableCell>
              <TableCell>{request.approver || "-"}</TableCell>
              <TableCell className="whitespace-nowrap text-xs">{Setting.getFormattedDate(request.createdTime)}</TableCell>
              <TableCell>
                <div className="flex flex-wrap gap-1">
                  {open && request.type !== "Role" ? (
                    <Button size="sm" loading={busy === request.name} onClick={() => run(request, () => PrivilegedAccountBackend.revealPrivilegedPassword(request), (res) => onReveal(res.data))}>
                      <Eye className="size-4" />
                      {i18next.t("pam:Reveal password")}
                    </Button>
                  ) : null}
                  {open ? (
                    <Button size="sm" variant="outline" disabled={busy === request.name} onClick={() => run(request, () => PrivilegedAccountBackend.returnPrivilegedRequest(request))}>
                      {i18next.t("pam:Return early")}
                    </Button>
                  ) : null}
                  {request.state === "Pending" ? (
                    <Button size="sm" variant="outline" disabled={busy === request.name} onClick={() => run(request, () => PrivilegedAccountBackend.cancelPrivilegedRequest(request))}>
                      {i18next.t("pam:Withdraw request")}
                    </Button>
                  ) : null}
                </div>
              </TableCell>
            </TableRow>
          );
        })}
      </TableBody>
    </Table>
  );
}

function RequestTable({requests, actions}: {requests: any[]; actions?: (request: any) => React.ReactNode}) {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>{i18next.t("pam:Requested by")}</TableHead>
          <TableHead>{i18next.t("pam:Target system")}</TableHead>
          <TableHead>{i18next.t("general:State")}</TableHead>
          <TableHead>{i18next.t("rule:Reason")}</TableHead>
          <TableHead>{i18next.t("pam:Duration in minutes")}</TableHead>
          <TableHead>{i18next.t("permission:Approver")}</TableHead>
          <TableHead>{i18next.t("pam:Password reveals")}</TableHead>
          <TableHead>{i18next.t("general:Created time")}</TableHead>
          {actions ? <TableHead className="w-48" /> : null}
        </TableRow>
      </TableHeader>
      <TableBody>
        {requests.map((request) => (
          <TableRow key={`${request.owner}/${request.name}`}>
            <TableCell>{request.requester}</TableCell>
            <TableCell>
              <div className="font-medium">{getTarget(request)}</div>
              <div className="text-xs text-muted-foreground">{enumLabel(PAM_ACCOUNT_TYPES, request.type)}{request.role ? `: ${request.role}` : ""}</div>
            </TableCell>
            <TableCell><EnumBadge map={PAM_REQUEST_STATES} value={request.state} /></TableCell>
            <TableCell className="max-w-56 break-words text-xs">{request.reason}{request.comment ? <div className="text-muted-foreground">{request.comment}</div> : null}</TableCell>
            <TableCell>{request.durationMinutes}</TableCell>
            <TableCell>{request.approver || "-"}</TableCell>
            <TableCell>{request.revealCount}</TableCell>
            <TableCell className="whitespace-nowrap text-xs">{Setting.getFormattedDate(request.createdTime)}</TableCell>
            {actions ? <TableCell>{actions(request)}</TableCell> : null}
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}

function AllRequestsTab({owner}: {owner: string}) {
  const [requests, setRequests] = React.useState<any[] | null>(null);
  React.useEffect(() => {
    PrivilegedAccountBackend.getPrivilegedRequests(owner, 1, 200).then((res: any) => {
      if (res.status === "ok") {
        setRequests(res.data ?? []);
      } else {
        setRequests([]);
        Setting.showMessage("error", res.msg);
      }
    });
  }, [owner]);
  if (requests === null) {
    return <Loading />;
  }
  return <RequestTable requests={requests} />;
}

/** Privileged access for everyone: request an account or a role, reveal the password while the access lasts, and approve the requests of others. */
export default function PrivilegedAccessPage() {
  const {account} = useAccount();
  const now = useNow();
  const [data, setData] = React.useState<any>(null);
  const [requesting, setRequesting] = React.useState<any>(null);
  const [revealed, setRevealed] = React.useState<any>(null);
  const [deciding, setDeciding] = React.useState<{request: any; approve: boolean} | null>(null);
  const [tab, setTab] = React.useState("requests");

  const reload = React.useCallback(() => {
    PrivilegedAccountBackend.getPrivilegedAccess().then((res: any) => {
      if (res.status === "ok") {
        setData(res.data);
      } else {
        Setting.showMessage("error", res.msg);
      }
    });
  }, []);
  React.useEffect(reload, [reload]);

  // the countdowns reach zero on the page, the server ends the access within a minute
  const hasExpired = data?.requests?.some((request: any) => request.state === "Active" && new Date(request.expireTime).getTime() < now);
  React.useEffect(() => {
    if (hasExpired) {
      const timer = window.setTimeout(reload, 30000);
      return () => window.clearTimeout(timer);
    }
  }, [hasExpired, reload]);

  if (data === null) {
    return <Loading />;
  }

  const pending = data.approvals.filter((request: any) => request.state === "Pending");
  const decided = data.approvals.filter((request: any) => request.state !== "Pending");

  return (
    <div className="space-y-4">
      <PageHeader title={i18next.t("pam:Privileged access")} description={i18next.t("pam:Privileged access - Tooltip")} />
      <EditionNotice edition="Enterprise" message={i18next.t("pam:Privileged account management needs the enterprise edition")} />
      <Tabs value={tab} onValueChange={setTab}>
        <TabsList className="mb-2 flex-wrap">
          <TabsTrigger value="requests">{i18next.t("pam:My requests")}</TabsTrigger>
          <TabsTrigger value="accounts">{i18next.t("pam:Request access")}</TabsTrigger>
          <TabsTrigger value="approvals">
            {i18next.t("pam:Waiting for my approval")}
            {pending.length > 0 ? <Badge variant="destructive" className="ml-1.5">{pending.length}</Badge> : null}
          </TabsTrigger>
          {Setting.isLocalAdminUser(account) ? <TabsTrigger value="all">{i18next.t("pam:All requests")}</TabsTrigger> : null}
        </TabsList>
        <TabsContent value="requests">
          <MyRequestsTab requests={data.requests} now={now} reload={reload} onReveal={setRevealed} />
        </TabsContent>
        <TabsContent value="accounts">
          <AccountsTab accounts={data.accounts} onRequest={setRequesting} />
        </TabsContent>
        <TabsContent value="approvals" className="space-y-4">
          {pending.length === 0 ? <p className="text-sm text-muted-foreground">{i18next.t("pam:No request is waiting for you")}</p> : (
            <RequestTable
              requests={pending}
              actions={(request) => (
                <div className="flex gap-1">
                  <Button size="sm" onClick={() => setDeciding({request, approve: true})}>{i18next.t("mfa:Approve")}</Button>
                  <Button size="sm" variant="outline" onClick={() => setDeciding({request, approve: false})}>{i18next.t("pam:Reject request")}</Button>
                </div>
              )}
            />
          )}
          {decided.length > 0 ? (
            <div className="space-y-2">
              <div className="text-sm font-medium">{i18next.t("pam:Decided by me")}</div>
              <RequestTable requests={decided} />
            </div>
          ) : null}
        </TabsContent>
        {Setting.isLocalAdminUser(account) && tab === "all" ? (
          <TabsContent value="all"><AllRequestsTab owner={account?.owner ?? ""} /></TabsContent>
        ) : null}
      </Tabs>

      {requesting ? (
        <RequestDialog target={requesting} onClose={() => setRequesting(null)} onDone={() => {
          setRequesting(null);
          setTab("requests");
          reload();
        }} />
      ) : null}
      {revealed ? <RevealDialog revealed={revealed} onClose={() => setRevealed(null)} /> : null}
      {deciding ? (
        <DecideDialog request={deciding.request} approve={deciding.approve} onClose={() => setDeciding(null)} onDone={() => {
          setDeciding(null);
          reload();
        }} />
      ) : null}
    </div>
  );
}
