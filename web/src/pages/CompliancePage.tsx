import * as React from "react";
import i18next from "i18next";
import {Printer, ShieldCheck} from "lucide-react";
import {Tabs, TabsContent, TabsList, TabsTrigger} from "@/components/ui/tabs";
import {Card, CardContent, CardHeader, CardTitle} from "@/components/ui/card";
import {Badge} from "@/components/ui/badge";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Textarea} from "@/components/ui/textarea";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table";
import {PageHeader} from "@/components/crud/PageHeader";
import {Loading} from "@/components/common/Loading";
import {SearchableSelect} from "@/components/common/SearchableSelect";
import {EditionNotice} from "@/components/common/EditionNotice";
import {useAccount} from "@/hooks/use-account";
import {useOrganizationOptions} from "@/hooks/use-options";
import * as EnterpriseBackend from "@/backend/EnterpriseBackend";
import * as UserBackend from "@/backend/UserBackend";
import * as Setting from "@/lib/setting";

/** Loads one response's data, shows its error as a message, reloads when the deps change. */
function useData<T>(load: () => Promise<any> | null, deps: React.DependencyList): [T | null, boolean, () => void] {
  const [data, setData] = React.useState<T | null>(null);
  const [loading, setLoading] = React.useState(false);
  const [version, setVersion] = React.useState(0);
  React.useEffect(() => {
    const promise = load();
    if (!promise) {
      return;
    }
    let stopped = false;
    setLoading(true);
    promise.then((res: any) => {
      if (stopped) {
        return;
      }
      setLoading(false);
      if (res.status === "ok") {
        setData(res.data);
      } else {
        setData(null);
        Setting.showMessage("error", res.msg);
      }
    });
    return () => {
      stopped = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [...deps, version]);
  return [data, loading, () => setVersion((v) => v + 1)];
}

const statusVariants: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  pass: "default",
  warn: "secondary",
  fail: "destructive",
};

function getStatusText(status: string) {
  if (status === "pass") {
    return i18next.t("enterprise:Pass");
  }
  if (status === "warn") {
    return i18next.t("enterprise:Warning");
  }
  return i18next.t("enterprise:Fail");
}

function escapeHtml(text: any) {
  return String(text ?? "").replace(/[&<>"']/g, (c) => ({"&": "&amp;", "<": "&lt;", ">": "&gt;", "\"": "&quot;", "'": "&#39;"}[c] as string));
}

/** Opens a printable page, which the browser can save as PDF. */
function printHtml(title: string, body: string) {
  const win = window.open("", "_blank");
  if (!win) {
    return;
  }
  win.document.write(`<!doctype html><html><head><meta charset="utf-8"><title>${escapeHtml(title)}</title><style>
body{font-family:-apple-system,"PingFang SC","Microsoft YaHei",sans-serif;margin:32px;color:#111;font-size:13px}
h1{font-size:20px}h2{font-size:15px;margin-top:24px}table{border-collapse:collapse;width:100%;margin-top:8px}
th,td{border:1px solid #ccc;padding:6px 8px;text-align:left;vertical-align:top}th{background:#f3f4f6}
.pass{color:#15803d}.warn{color:#b45309}.fail{color:#b91c1c}.muted{color:#666}
</style></head><body>${body}</body></html>`);
  win.document.close();
  win.focus();
  win.print();
}

function ComplianceReportTab({owner}: {owner: string}) {
  const [report, loading] = useData<any>(() => (owner ? EnterpriseBackend.getComplianceReport(owner) : null), [owner]);
  if (loading || !report) {
    return loading ? <Loading /> : null;
  }

  const print = () => {
    const rows = report.items.map((item: any) => `<tr><td>${escapeHtml(item.category)}</td><td>${escapeHtml(item.title)}</td><td>${escapeHtml(item.requirement)}</td><td class="${item.status}">${escapeHtml(getStatusText(item.status))}</td><td>${escapeHtml(item.detail)}</td></tr>`).join("");
    printHtml(i18next.t("enterprise:MLPS checklist"), `<h1>${escapeHtml(i18next.t("enterprise:MLPS checklist"))}: ${escapeHtml(owner)}</h1>
<p class="muted">${escapeHtml(Setting.getFormattedDate(report.createdTime))}, ${escapeHtml(getStatusText("pass"))} ${report.pass}, ${escapeHtml(getStatusText("warn"))} ${report.warn}, ${escapeHtml(getStatusText("fail"))} ${report.fail}</p>
<table><tr><th>${escapeHtml(i18next.t("general:Category"))}</th><th>${escapeHtml(i18next.t("enterprise:Item"))}</th><th>${escapeHtml(i18next.t("enterprise:Requirement"))}</th><th>${escapeHtml(i18next.t("general:Status"))}</th><th>${escapeHtml(i18next.t("general:Detail"))}</th></tr>${rows}</table>`);
  };

  return (
    <div className="space-y-3">
      <div className="flex flex-wrap items-center gap-2">
        <Badge>{getStatusText("pass")} {report.pass}</Badge>
        <Badge variant="secondary">{getStatusText("warn")} {report.warn}</Badge>
        <Badge variant="destructive">{getStatusText("fail")} {report.fail}</Badge>
        <span className="text-xs text-muted-foreground">{i18next.t("enterprise:MLPS checklist - Tooltip")}</span>
        <Button variant="outline" size="sm" className="ml-auto" onClick={print}>
          <Printer className="size-4" />
          {i18next.t("enterprise:Print")}
        </Button>
      </div>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-32">{i18next.t("general:Category")}</TableHead>
            <TableHead className="w-40">{i18next.t("enterprise:Item")}</TableHead>
            <TableHead>{i18next.t("enterprise:Requirement")}</TableHead>
            <TableHead className="w-20">{i18next.t("general:Status")}</TableHead>
            <TableHead>{i18next.t("general:Detail")}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {report.items.map((item: any, index: number) => (
            <TableRow key={index}>
              <TableCell className="text-muted-foreground">{item.category}</TableCell>
              <TableCell className="font-medium">{item.title}</TableCell>
              <TableCell className="text-xs">{item.requirement}</TableCell>
              <TableCell><Badge variant={statusVariants[item.status]}>{getStatusText(item.status)}</Badge></TableCell>
              <TableCell className="text-xs">{item.detail}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

function Stat({label, value}: {label: string; value: React.ReactNode}) {
  return (
    <Card>
      <CardContent className="p-4">
        <div className="text-xs text-muted-foreground">{label}</div>
        <div className="mt-1 text-2xl font-semibold tabular-nums">{value}</div>
      </CardContent>
    </Card>
  );
}

function CountList({title, items}: {title: string; items: any[]}) {
  const max = Math.max(1, ...items.map((item) => item.count));
  return (
    <Card>
      <CardHeader className="pb-2"><CardTitle className="text-sm">{title}</CardTitle></CardHeader>
      <CardContent className="space-y-1.5">
        {items.length === 0 ? <p className="text-xs text-muted-foreground">{i18next.t("general:None")}</p> : null}
        {items.map((item) => (
          <div key={item.name} className="text-xs">
            <div className="flex justify-between gap-2"><span className="truncate">{item.name}</span><span className="tabular-nums">{item.count}</span></div>
            <div className="mt-0.5 h-1.5 rounded bg-secondary"><div className="h-full rounded bg-primary" style={{width: `${(item.count / max) * 100}%`}} /></div>
          </div>
        ))}
      </CardContent>
    </Card>
  );
}

function getOverviewStats(overview: any): [string, React.ReactNode][] {
  return [
    [i18next.t("enterprise:Sign-ins"), overview.signinSuccess],
    [i18next.t("home:Failed sign-ins"), overview.signinFailure],
    [i18next.t("enterprise:Failure rate"), `${(overview.failureRate * 100).toFixed(1)}%`],
    [i18next.t("enterprise:Active users"), overview.activeUsers],
    [i18next.t("enterprise:Signups"), overview.signups],
    [i18next.t("enterprise:Admin operations"), overview.adminOperations],
    [i18next.t("enterprise:Denied requests"), overview.denied],
    [i18next.t("enterprise:New locations and devices"), overview.newLocations + overview.newDevices],
  ];
}

function SecurityOverviewTab({owner}: {owner: string}) {
  const [days, setDays] = React.useState(7);
  const [month, setMonth] = React.useState("");
  const [overview, loading] = useData<any>(() => (owner ? EnterpriseBackend.getSecurityOverview(owner, days, month) : null), [owner, days, month]);

  const print = () => {
    const stats = getOverviewStats(overview).map(([label, value]) => `<tr><th>${escapeHtml(label)}</th><td>${escapeHtml(value)}</td></tr>`).join("");
    const list = (title: string, items: any[]) => `<h2>${escapeHtml(title)}</h2><table>${items.map((item) => `<tr><td>${escapeHtml(item.name)}</td><td>${item.count}</td></tr>`).join("") || `<tr><td class="muted">${escapeHtml(i18next.t("general:None"))}</td></tr>`}</table>`;
    const dayRows = overview.days.map((day: any) => `<tr><td>${day.date}</td><td>${day.success}</td><td>${day.failure}</td><td>${day.signups}</td></tr>`).join("");
    printHtml(i18next.t("enterprise:Monthly report"), `<h1>${escapeHtml(i18next.t("enterprise:Security report"))}: ${escapeHtml(owner)}</h1>
<p class="muted">${escapeHtml(Setting.getFormattedDate(overview.from))} ~ ${escapeHtml(Setting.getFormattedDate(overview.to))}</p>
<table>${stats}</table>
<h2>${escapeHtml(i18next.t("enterprise:By day"))}</h2><table><tr><th>${escapeHtml(i18next.t("enterprise:Date"))}</th><th>${escapeHtml(i18next.t("enterprise:Sign-ins"))}</th><th>${escapeHtml(i18next.t("home:Failed sign-ins"))}</th><th>${escapeHtml(i18next.t("enterprise:Signups"))}</th></tr>${dayRows}</table>
${list(i18next.t("enterprise:Top failed IPs"), overview.topFailedIps)}${list(i18next.t("enterprise:Top failed users"), overview.topFailedUsers)}
${list(i18next.t("enterprise:Top admin operations"), overview.topActions)}${list(i18next.t("enterprise:Top admins"), overview.topAdmins)}`);
  };

  const maxDay = overview ? Math.max(1, ...overview.days.map((day: any) => day.success + day.failure)) : 1;
  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center gap-2">
        {[7, 30, 90].map((value) => (
          <Button key={value} size="sm" variant={!month && days === value ? "default" : "outline"} onClick={() => {
            setMonth("");
            setDays(value);
          }}>
            {i18next.t("enterprise:Last %s days").replace("%s", String(value))}
          </Button>
        ))}
        <Input type="month" className="w-44" value={month} onChange={(e) => setMonth(e.target.value)} />
        <Button variant="outline" size="sm" className="ml-auto" disabled={!overview} onClick={print}>
          <Printer className="size-4" />
          {i18next.t(month ? "enterprise:Print monthly report" : "enterprise:Print")}
        </Button>
      </div>
      {loading || !overview ? (loading ? <Loading /> : null) : (
        <>
          {overview.truncated ? <p className="text-xs text-muted-foreground">{i18next.t("enterprise:Only the first records were counted")}</p> : null}
          <div className="grid grid-cols-2 gap-3 md:grid-cols-4">
            {getOverviewStats(overview).map(([label, value]) => <Stat key={label} label={label} value={value} />)}
          </div>
          <Card>
            <CardHeader className="pb-2"><CardTitle className="text-sm">{i18next.t("enterprise:By day")}</CardTitle></CardHeader>
            <CardContent>
              <div className="flex h-40 items-end gap-1">
                {overview.days.map((day: any) => (
                  <div key={day.date} className="flex flex-1 flex-col justify-end" title={`${day.date}: ${day.success} / ${day.failure}`}>
                    <div className="bg-destructive/70" style={{height: `${(day.failure / maxDay) * 140}px`}} />
                    <div className="bg-primary" style={{height: `${(day.success / maxDay) * 140}px`}} />
                  </div>
                ))}
              </div>
              <div className="mt-1 flex justify-between text-[10px] text-muted-foreground">
                <span>{overview.days[0]?.date}</span>
                <span>{overview.days[overview.days.length - 1]?.date}</span>
              </div>
            </CardContent>
          </Card>
          <div className="grid gap-3 md:grid-cols-2 lg:grid-cols-4">
            <CountList title={i18next.t("enterprise:Top failed IPs")} items={overview.topFailedIps} />
            <CountList title={i18next.t("enterprise:Top failed users")} items={overview.topFailedUsers} />
            <CountList title={i18next.t("enterprise:Top admin operations")} items={overview.topActions} />
            <CountList title={i18next.t("enterprise:Top admins")} items={overview.topAdmins} />
          </div>
          <Card>
            <CardHeader className="pb-2"><CardTitle className="text-sm">{i18next.t("enterprise:Recent new places and devices")}</CardTitle></CardHeader>
            <CardContent>
              <Table>
                <TableBody>
                  {overview.recentSigninRisks.map((trace: any, index: number) => (
                    <TableRow key={index}>
                      <TableCell>{Setting.getFormattedDate(trace.createdTime)}</TableCell>
                      <TableCell>{trace.user}</TableCell>
                      <TableCell>{trace.kind === "location" ? i18next.t("user:Location") : i18next.t("entry:Device")}</TableCell>
                      <TableCell className="max-w-xs truncate text-xs">{trace.kind === "location" ? trace.traceKey : trace.userAgent}</TableCell>
                      <TableCell className="text-xs">{trace.ip}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
              {overview.recentSigninRisks.length === 0 ? <p className="text-xs text-muted-foreground">{i18next.t("general:None")}</p> : null}
            </CardContent>
          </Card>
        </>
      )}
    </div>
  );
}

const reasonKeys: Record<string, string> = {
  "never-signed-in": "enterprise:Never signed in",
  "inactive": "key:Inactive",
  "disabled-with-roles": "enterprise:Disabled but holds roles",
};

function GovernanceTab({owner}: {owner: string}) {
  const [days, setDays] = React.useState(90);
  const [accounts, loading, reload] = useData<any[]>(() => (owner ? EnterpriseBackend.getOrphanAccounts(owner, days) : null), [owner, days]);
  const [conflicts] = useData<any[]>(() => (owner ? EnterpriseBackend.getSodConflicts(owner) : null), [owner]);

  const disable = async(userId: string) => {
    const [userOwner, userName] = userId.split("/");
    const res = await UserBackend.getUser(userOwner, userName);
    if (res.status !== "ok" || !res.data) {
      Setting.showMessage("error", res.msg);
      return;
    }
    const result = await UserBackend.updateUser(userOwner, userName, {...res.data, isForbidden: true});
    if (result.status === "ok") {
      Setting.showMessage("success", i18next.t("general:Successfully saved"));
      reload();
    } else {
      Setting.showMessage("error", result.msg);
    }
  };

  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <div className="flex flex-wrap items-center gap-2">
          <h3 className="text-sm font-semibold">{i18next.t("enterprise:Orphan accounts")}</h3>
          <span className="text-xs text-muted-foreground">{i18next.t("enterprise:Days without signing in")}</span>
          <Input type="number" className="w-24" min={1} value={days} onChange={(e) => setDays(Setting.myParseInt(e.target.value) || 90)} />
        </div>
        {loading ? <Loading /> : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{i18next.t("general:User")}</TableHead>
                <TableHead>{i18next.t("enterprise:Reasons")}</TableHead>
                <TableHead>{i18next.t("general:Created time")}</TableHead>
                <TableHead>{i18next.t("enterprise:Last sign-in")}</TableHead>
                <TableHead>{i18next.t("general:Roles")}</TableHead>
                <TableHead />
              </TableRow>
            </TableHeader>
            <TableBody>
              {(accounts ?? []).map((account: any) => (
                <TableRow key={account.user}>
                  <TableCell>{account.displayName ? `${account.displayName} (${account.user})` : account.user}{account.isAdmin ? <Badge variant="outline" className="ml-1">{i18next.t("general:Admin")}</Badge> : null}</TableCell>
                  <TableCell>{account.reasons.map((reason: string) => <Badge key={reason} variant="secondary" className="mr-1">{i18next.t(reasonKeys[reason] ?? reason)}</Badge>)}</TableCell>
                  <TableCell className="text-xs">{Setting.getFormattedDate(account.createdTime)}</TableCell>
                  <TableCell className="text-xs">{account.lastSigninTime ? Setting.getFormattedDate(account.lastSigninTime) : "-"}</TableCell>
                  <TableCell className="text-xs">{(account.roles ?? []).join(", ")}</TableCell>
                  <TableCell>{account.isForbidden ? null : <Button size="sm" variant="outline" onClick={() => disable(account.user)}>{i18next.t("general:Disable")}</Button>}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
        {accounts && accounts.length === 0 ? <p className="text-xs text-muted-foreground">{i18next.t("general:None")}</p> : null}
      </div>

      <div className="space-y-2">
        <h3 className="text-sm font-semibold">{i18next.t("enterprise:Separation-of-duty conflicts")}</h3>
        <p className="text-xs text-muted-foreground">{i18next.t("enterprise:Separation-of-duty conflicts - Tooltip")}</p>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{i18next.t("general:User")}</TableHead>
              <TableHead>{i18next.t("application:Rule")}</TableHead>
              <TableHead>{i18next.t("general:Roles")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {(conflicts ?? []).map((conflict: any, index: number) => (
              <TableRow key={index}>
                <TableCell>{conflict.displayName ? `${conflict.displayName} (${conflict.user})` : conflict.user}</TableCell>
                <TableCell>{conflict.rule}</TableCell>
                <TableCell className="text-xs">{conflict.roles.join(", ")}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
        {conflicts && conflicts.length === 0 ? <p className="text-xs text-muted-foreground">{i18next.t("general:None")}</p> : null}
      </div>
    </div>
  );
}

function ConsentLogTab({owner}: {owner: string}) {
  const [page, setPage] = React.useState(1);
  const [logs, loading] = useData<any[]>(() => (owner ? EnterpriseBackend.getConsentLogs(owner, page, 20) : null), [owner, page]);
  return (
    <div className="space-y-2">
      <p className="text-xs text-muted-foreground">{i18next.t("enterprise:Consent log - Tooltip")}</p>
      {loading ? <Loading /> : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{i18next.t("general:Created time")}</TableHead>
              <TableHead>{i18next.t("general:User")}</TableHead>
              <TableHead>{i18next.t("general:Application")}</TableHead>
              <TableHead>{i18next.t("general:Action")}</TableHead>
              <TableHead>{i18next.t("general:Scopes")}</TableHead>
              <TableHead>{i18next.t("general:Client IP")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {(logs ?? []).map((log: any) => (
              <TableRow key={log.name}>
                <TableCell className="text-xs">{Setting.getFormattedDate(log.createdTime)}</TableCell>
                <TableCell>{log.user}</TableCell>
                <TableCell>{log.application}</TableCell>
                <TableCell><Badge variant={log.action === "grant" ? "default" : "secondary"}>{i18next.t(log.action === "grant" ? "enterprise:Grant" : "enterprise:Revoke")}</Badge></TableCell>
                <TableCell className="text-xs">{(log.grantedScopes ?? []).join(" ")}{(log.deniedScopes ?? []).length > 0 ? ` (-${log.deniedScopes.join(" ")})` : ""}</TableCell>
                <TableCell className="text-xs">{log.clientIp}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
      <div className="flex items-center gap-2">
        <Button size="sm" variant="outline" disabled={page <= 1} onClick={() => setPage(page - 1)}>{i18next.t("enterprise:Previous page")}</Button>
        <span className="text-xs tabular-nums">{page}</span>
        <Button size="sm" variant="outline" disabled={!logs || logs.length < 20} onClick={() => setPage(page + 1)}>{i18next.t("enterprise:Next page")}</Button>
      </div>
    </div>
  );
}

function RecordIntegrityTab({owner}: {owner: string}) {
  const [result, setResult] = React.useState<any>(null);
  const [loading, setLoading] = React.useState(false);
  React.useEffect(() => setResult(null), [owner]);

  const verify = async() => {
    setLoading(true);
    const res = await EnterpriseBackend.verifyRecords(owner);
    setLoading(false);
    if (res.status === "ok") {
      setResult(res.data);
    } else {
      Setting.showMessage("error", res.msg);
    }
  };

  return (
    <div className="space-y-3">
      <p className="text-xs text-muted-foreground">{i18next.t("enterprise:Record integrity - Tooltip")}</p>
      <Button size="sm" loading={loading} onClick={verify}>
        <ShieldCheck className="size-4" />
        {i18next.t("forget:Verify")}
      </Button>
      {result ? (
        <div className="space-y-2 text-sm">
          <p>
            {i18next.t("enterprise:Checked records")}: {result.checked}
            {result.checked > 0 ? ` (${Setting.getFormattedDate(result.firstTime)} ~ ${Setting.getFormattedDate(result.lastTime)})` : ""}
          </p>
          {result.checked === 0 ? <p className="text-muted-foreground">{i18next.t("enterprise:No chained records yet")}</p> : null}
          {result.checked > 0 && result.breaks.length === 0 ? <Badge>{i18next.t("enterprise:The chain is intact")}</Badge> : null}
          {result.breaks.length > 0 ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>ID</TableHead>
                  <TableHead>{i18next.t("general:Created time")}</TableHead>
                  <TableHead>{i18next.t("enterprise:Problem")}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {result.breaks.map((item: any, index: number) => (
                  <TableRow key={index}>
                    <TableCell>{item.id}</TableCell>
                    <TableCell className="text-xs">{Setting.getFormattedDate(item.createdTime)}</TableCell>
                    <TableCell>
                      <Badge variant="destructive">
                        {i18next.t(item.reason === "modified" ? "enterprise:The record was changed" : "enterprise:A record before it was deleted, inserted or reordered")}
                      </Badge>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : null}
        </div>
      ) : null}
    </div>
  );
}

function OtpDeviceTab({owner}: {owner: string}) {
  const [tokens, setTokens] = React.useState("");
  const [assignments, setAssignments] = React.useState("");
  const [result, setResult] = React.useState<any>(null);
  const [loading, setLoading] = React.useState(false);

  const readFile = (file: File | undefined) => {
    if (file) {
      file.text().then(setTokens);
    }
  };
  const submit = async() => {
    setLoading(true);
    const res = await EnterpriseBackend.importOtpDevices(owner, tokens, assignments);
    setLoading(false);
    if (res.status === "ok") {
      setResult(res.data);
      Setting.showMessage("success", i18next.t("enterprise:Imported %s tokens").replace("%s", String(res.data.imported)));
    } else {
      Setting.showMessage("error", res.msg);
    }
  };

  return (
    <div className="max-w-3xl space-y-3">
      <p className="text-xs text-muted-foreground">{i18next.t("enterprise:Hardware tokens - Tooltip")}</p>
      <div className="space-y-1">
        <div className="text-sm font-medium">{i18next.t("enterprise:Token seeds")}</div>
        <Input type="file" accept=".xml,.pskc,.csv,.txt" onChange={(e) => readFile(e.target.files?.[0])} />
        <Textarea rows={6} className="font-mono text-xs" placeholder={"serial,secret,period,digits,user\nFT0001,JBSWY3DPEHPK3PXP,60,6,zhangsan"} value={tokens} onChange={(e) => setTokens(e.target.value)} />
      </div>
      <div className="space-y-1">
        <div className="text-sm font-medium">{i18next.t("enterprise:Token assignments")}</div>
        <Textarea rows={4} className="font-mono text-xs" placeholder={"zhangsan,FT0001\nlisi,FT0002"} value={assignments} onChange={(e) => setAssignments(e.target.value)} />
      </div>
      <Button loading={loading} disabled={!tokens.trim()} onClick={submit}>{i18next.t("enterprise:Import")}</Button>
      {result && result.skipped.length > 0 ? (
        <div className="text-xs">
          <div className="font-medium">{i18next.t("enterprise:Skipped")}</div>
          <ul className="list-disc pl-5">{result.skipped.map((item: string) => <li key={item}>{item}</li>)}</ul>
        </div>
      ) : null}
    </div>
  );
}

/** The enterprise edition's compliance and governance tools for one organization. */
export default function CompliancePage() {
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
        title={i18next.t("general:Compliance")}
        description={i18next.t("enterprise:Compliance - Tooltip")}
        actions={isGlobalAdmin ? <SearchableSelect className="w-60" value={owner} onChange={setOwner} options={organizations} /> : null}
      />
      <EditionNotice edition="Enterprise" />
      {owner ? (
        <Tabs defaultValue="report">
          <TabsList className="mb-2 flex-wrap">
            <TabsTrigger value="report">{i18next.t("enterprise:MLPS checklist")}</TabsTrigger>
            <TabsTrigger value="security">{i18next.t("enterprise:Security overview")}</TabsTrigger>
            <TabsTrigger value="governance">{i18next.t("enterprise:Account governance")}</TabsTrigger>
            <TabsTrigger value="consents">{i18next.t("enterprise:Consent log")}</TabsTrigger>
            <TabsTrigger value="records">{i18next.t("enterprise:Record integrity")}</TabsTrigger>
            <TabsTrigger value="otp">{i18next.t("enterprise:Hardware tokens")}</TabsTrigger>
          </TabsList>
          <TabsContent value="report"><ComplianceReportTab owner={owner} /></TabsContent>
          <TabsContent value="security"><SecurityOverviewTab owner={owner} /></TabsContent>
          <TabsContent value="governance"><GovernanceTab owner={owner} /></TabsContent>
          <TabsContent value="consents"><ConsentLogTab owner={owner} /></TabsContent>
          <TabsContent value="records"><RecordIntegrityTab owner={owner} /></TabsContent>
          <TabsContent value="otp"><OtpDeviceTab owner={owner} /></TabsContent>
        </Tabs>
      ) : null}
    </div>
  );
}
