import * as React from "react";
import i18next from "i18next";
import {ArrowRight, ArrowUpRight, Check, Gem, Lock, ShieldAlert, ShieldCheck, Sparkles} from "lucide-react";
import {Link} from "react-router-dom";
import {Bar, BarChart, CartesianGrid, Legend, ResponsiveContainer, Tooltip as RechartsTooltip, XAxis, YAxis} from "recharts";
import * as Conf from "@/Conf";
import {Badge} from "@/components/ui/badge";
import {Button} from "@/components/ui/button";
import {Card, CardContent, CardHeader, CardTitle} from "@/components/ui/card";
import {getEditionText, PricingUrl} from "@/components/common/EditionNotice";
import * as Setting from "@/lib/setting";
import {cn} from "@/lib/utils";

const ranks: Record<string, number> = {Community: 0, Commercial: 1, Enterprise: 2};

export function hasEdition(current: string, required: string) {
  return (ranks[current] ?? 0) >= (ranks[required] ?? 0);
}

/** what GET /api/get-dashboard-security returns */
export interface DashboardSecurity {
  users: number;
  inactiveUsers: number;
  inactiveDays: number;
  admins: number;
  adminsWithoutMfa: number;
  organizations: number;
  policies: Record<string, number>;
  applications: number;
  applicationsWithSessionControl: number;
  certs: number;
  certsExpiring: number;
  certsRotated: number;
  certAlertDays: number;
  nearestCertExpiry: number;
  signinDates: string[];
  signins: number[];
  failedSignins: number[];
  recentDays: number;
  newLocationSignins: number;
  newDeviceSignins: number;
  mcpCalls: number;
  mcpBlockedCalls: number;
}

export interface SecurityCheck {
  key: string;
  edition: string;
  ok: boolean;
  title: string;
  detail: string;
  fix: string;
  to: string;
}

/** The checks of the dashboard, run on the organization's own data, so each one says what it found there. */
export function getSecurityChecks(data: DashboardSecurity, organization: string): SecurityCheck[] {
  const orgs = data.organizations;
  const policy = (key: string) => data.policies?.[key] ?? 0;
  const orgLink = organization ? `/organizations/${organization}` : "/organizations";
  const missing = (key: string) => Math.max(0, orgs - policy(key));

  return [
    {
      key: "adminMfa",
      edition: "Community",
      ok: data.adminsWithoutMfa === 0,
      title: i18next.t("home:MFA for admins"),
      detail: data.adminsWithoutMfa === 0
        ? i18next.t("home:Every admin signs in with MFA")
        : i18next.t("home:Admins without MFA (%s)").replace("%s", String(data.adminsWithoutMfa)),
      fix: i18next.t("home:Turn on MFA for them"),
      to: "/users",
    },
    {
      key: "inactive",
      edition: "Commercial",
      ok: data.inactiveUsers === 0 || (orgs > 0 && missing("disableInactive") === 0),
      title: i18next.t("home:Inactive accounts"),
      detail: data.inactiveUsers === 0
        ? i18next.t("home:No account has gone unused for long")
        : i18next.t("home:Accounts unused for %d days (%s)").replace("%d", String(data.inactiveDays)).replace("%s", String(data.inactiveUsers)),
      fix: i18next.t("home:Disable unused accounts automatically"),
      to: orgLink,
    },
    {
      key: "session",
      edition: "Commercial",
      ok: data.applications === 0 || data.applicationsWithSessionControl === data.applications || (orgs > 0 && missing("inactiveTimeout") === 0),
      title: i18next.t("home:Session control"),
      detail: i18next.t("home:Applications with session control (%s / %d)")
        .replace("%s", String(data.applicationsWithSessionControl)).replace("%d", String(data.applications)),
      fix: i18next.t("home:Set an idle timeout and exclusive sign-in"),
      to: "/applications",
    },
    {
      key: "passwordCheck",
      edition: "Commercial",
      ok: orgs > 0 && missing("passwordCheck") === 0,
      title: i18next.t("home:Weak and leaked passwords"),
      detail: missing("passwordCheck") === 0
        ? i18next.t("home:New passwords are checked against weak and leaked ones")
        : i18next.t("home:Organizations without password checks (%s)").replace("%s", String(missing("passwordCheck"))),
      fix: i18next.t("home:Check passwords when they are set"),
      to: orgLink,
    },
    {
      key: "stuffing",
      edition: "Commercial",
      ok: orgs > 0 && missing("stuffing") === 0,
      title: i18next.t("home:Credential stuffing"),
      detail: missing("stuffing") === 0
        ? i18next.t("home:IPs that try many accounts are blocked")
        : i18next.t("home:Organizations without credential stuffing defense (%s)").replace("%s", String(missing("stuffing"))),
      fix: i18next.t("home:Block credential stuffing"),
      to: orgLink,
    },
    {
      key: "signinAlert",
      edition: "Commercial",
      ok: orgs > 0 && missing("signinAlert") === 0,
      title: i18next.t("organization:Sign-in alerts"),
      detail: missing("signinAlert") === 0
        ? i18next.t("home:Users are told of sign-ins from a new place or device")
        : i18next.t("home:Organizations without sign-in alerts (%s)").replace("%s", String(missing("signinAlert"))),
      fix: i18next.t("home:Alert on new places and devices"),
      to: orgLink,
    },
    {
      key: "certs",
      edition: "Commercial",
      ok: data.certsExpiring === 0,
      title: i18next.t("home:Signing keys and certificates"),
      detail: data.certsExpiring > 0
        ? i18next.t("home:Certificates expiring within %d days (%s)").replace("%d", String(data.certAlertDays)).replace("%s", String(data.certsExpiring))
        : data.nearestCertExpiry >= 0
          ? i18next.t("home:The next certificate expires in %s days").replace("%s", String(data.nearestCertExpiry))
          : i18next.t("general:None"),
      fix: i18next.t("organization:Rotate signing keys automatically"),
      to: "/certs",
    },
    {
      key: "masking",
      edition: "Commercial",
      ok: orgs > 0 && missing("masking") === 0,
      title: i18next.t("home:Personal data masking"),
      detail: missing("masking") === 0
        ? i18next.t("home:Phone numbers, emails and ID numbers are masked in the console")
        : i18next.t("home:Organizations without data masking (%s)").replace("%s", String(missing("masking"))),
      fix: i18next.t("home:Mask personal data"),
      to: orgLink,
    },
  ];
}

/** 0 to 100, how far the organization is from a clean bill: the share of each check met, weighted. */
export function getPostureScore(data: DashboardSecurity, checks: SecurityCheck[]): number {
  const weights: Record<string, number> = {adminMfa: 20, inactive: 15, session: 10, passwordCheck: 12, stuffing: 12, signinAlert: 10, certs: 11, masking: 10};
  const share = (key: string, ok: boolean) => {
    if (ok) {
      return 1;
    }
    if (key === "adminMfa" && data.admins > 0) {
      return 1 - data.adminsWithoutMfa / data.admins;
    }
    if (key === "inactive" && data.users > 0) {
      return Math.max(0, 1 - data.inactiveUsers / data.users);
    }
    if (key === "session" && data.applications > 0) {
      return data.applicationsWithSessionControl / data.applications;
    }
    const policyKey: Record<string, string> = {passwordCheck: "passwordCheck", stuffing: "stuffing", signinAlert: "signinAlert", masking: "masking"};
    if (policyKey[key] && data.organizations > 0) {
      return (data.policies?.[policyKey[key]] ?? 0) / data.organizations;
    }
    return 0;
  };
  const total = checks.reduce((sum, check) => sum + (weights[check.key] ?? 0) * share(check.key, check.ok), 0);
  return Math.round(total);
}

function getGrade(score: number) {
  if (score >= 85) {
    return i18next.t("home:Strong");
  }
  if (score >= 65) {
    return i18next.t("home:Fair");
  }
  return i18next.t("home:Needs attention");
}

/** Where "upgrade" goes: on a demo site with a demo site for that edition there, elsewhere the pricing page. */
function UpgradeButton({edition, variant = "default", className}: {edition: string; variant?: "default" | "outline" | "secondary"; className?: string}) {
  const label = edition === "Enterprise" ? i18next.t("home:Try the enterprise edition") : i18next.t("home:Try the commercial edition");
  if (Conf.ThemeShowcase) {
    return (
      <Button type="button" variant={variant} className={className} onClick={() => Conf.setDemoEdition(edition)}>
        {label}
        <ArrowRight />
      </Button>
    );
  }
  return (
    <Button asChild variant={variant} className={className}>
      <a href={PricingUrl} target="_blank" rel="noreferrer">
        {edition === "Enterprise" ? i18next.t("home:Upgrade to the enterprise edition") : i18next.t("home:Upgrade to the commercial edition")}
        <ArrowUpRight />
      </a>
    </Button>
  );
}

function EditionBadge({edition, className}: {edition: string; className?: string}) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px] font-semibold",
        edition === "Enterprise" ? "bg-[#16294f] text-[#e9cf8f]" : "bg-indigo-600/10 text-indigo-700 dark:bg-indigo-400/15 dark:text-indigo-300",
        className,
      )}
    >
      {edition === "Enterprise" ? <Gem className="size-3" /> : <Sparkles className="size-3" />}
      {getEditionText(edition)}
    </span>
  );
}

/** the circular score, drawn as a plain SVG ring */
function ScoreRing({score, className, trackClassName, barClassName}: {score: number; className?: string; trackClassName?: string; barClassName?: string}) {
  const radius = 52;
  const circumference = 2 * Math.PI * radius;
  const offset = circumference * (1 - Math.min(Math.max(score, 0), 100) / 100);
  return (
    <div className={cn("relative size-[124px] shrink-0", className)}>
      <svg viewBox="0 0 124 124" className="size-full -rotate-90">
        <circle cx="62" cy="62" r={radius} className={cn("fill-none", trackClassName)} strokeWidth="9" />
        <circle
          cx="62"
          cy="62"
          r={radius}
          className={cn("fill-none transition-[stroke-dashoffset] duration-700", barClassName)}
          strokeWidth="9"
          strokeLinecap="round"
          strokeDasharray={circumference}
          strokeDashoffset={offset}
        />
      </svg>
      <div className="absolute inset-0 flex flex-col items-center justify-center">
        <span className="text-3xl font-bold tabular-nums">{score}</span>
        <span className="text-[11px] opacity-70">/ 100</span>
      </div>
    </div>
  );
}

function formatExpiry(license: any) {
  if (!license?.expireTime) {
    return i18next.t("home:Never expires");
  }
  return i18next.t("home:Valid until %s").replace("%s", Setting.getFormattedDate(license.expireTime));
}

/** The band across the top of the dashboard: which edition this is, and what it gives or would give. */
export function EditionHero({edition, license, data, checks, score}: {
  edition: string;
  license: any;
  data: DashboardSecurity | null;
  checks: SecurityCheck[];
  score: number;
}) {
  const findings = checks.filter((check) => !check.ok).length;
  const lockedFindings = checks.filter((check) => !check.ok && !hasEdition(edition, check.edition)).length;

  if (edition === "Enterprise") {
    return (
      <div className="relative overflow-hidden rounded-xl border border-[#c9a45c]/30 bg-[#0e1a33] p-5 text-white shadow-lg sm:p-6">
        <div
          aria-hidden
          className="pointer-events-none absolute inset-0"
          style={{background: "radial-gradient(40rem 18rem at 85% -20%, hsl(40 60% 55% / 0.22), transparent 60%), radial-gradient(30rem 20rem at -10% 120%, hsl(218 70% 40% / 0.45), transparent 60%)"}}
        />
        <div aria-hidden className="pointer-events-none absolute inset-x-0 top-0 h-1 bg-gradient-to-r from-[#c9a45c] via-[#e9cf8f] to-[#c9a45c]" />
        <div className="relative flex flex-col gap-5 lg:flex-row lg:items-center">
          <ScoreRing score={score} trackClassName="stroke-white/10" barClassName="stroke-[#e0bd6d]" />
          <div className="min-w-0 flex-1 space-y-2">
            <div className="flex flex-wrap items-center gap-2">
              <EditionBadge edition="Enterprise" className="bg-[#c9a45c]/15" />
              <span className="text-sm text-white/60">{i18next.t("home:Security posture")}</span>
            </div>
            <div className="text-2xl font-semibold tracking-tight">
              {getGrade(score)}
              <span className="ml-3 text-base font-normal text-white/70">
                {findings === 0 ? i18next.t("home:Every check passes") : i18next.t("home:%s checks need attention").replace("%s", String(findings))}
              </span>
            </div>
            <p className="max-w-2xl text-sm text-white/65">
              {i18next.t("home:The score weighs MFA, inactive accounts, session control, password checks, credential stuffing, sign-in alerts, key expiry and data masking across your organizations.")}
            </p>
          </div>
          <dl className="grid shrink-0 gap-x-6 gap-y-2 text-sm sm:grid-cols-3 lg:grid-cols-1 lg:border-l lg:border-white/10 lg:pl-6">
            {license?.licensee ? (
              <div>
                <dt className="text-xs text-white/50">{i18next.t("system:Licensee")}</dt>
                <dd className="font-medium">{license.licensee}</dd>
              </div>
            ) : null}
            <div>
              <dt className="text-xs text-white/50">{i18next.t("system:License")}</dt>
              <dd className="font-medium">{formatExpiry(license)}</dd>
            </div>
            <div>
              <dt className="text-xs text-white/50">{i18next.t("home:Support")}</dt>
              <dd className="font-medium text-[#e9cf8f]">{i18next.t("home:24/7 dedicated support")}</dd>
            </div>
          </dl>
        </div>
      </div>
    );
  }

  if (edition === "Commercial") {
    const protections = ["inactiveTimeout", "disableInactive", "passwordCheck", "stuffing", "signinAlert", "masking", "smsLimit", "certAlert"];
    const orgs = data?.organizations ?? 0;
    const enabled = orgs > 0 ? protections.filter((key) => (data?.policies?.[key] ?? 0) > 0).length : 0;
    return (
      <div className="relative overflow-hidden rounded-xl border border-indigo-500/20 bg-gradient-to-br from-indigo-600/[0.08] via-violet-500/[0.05] to-transparent p-5 sm:p-6">
        <div className="flex flex-col gap-5 xl:flex-row xl:items-center">
          <div className="min-w-0 flex-1 space-y-2">
            <div className="flex flex-wrap items-center gap-2">
              <EditionBadge edition="Commercial" />
              {license?.licensee ? <span className="text-sm text-muted-foreground">{license.licensee}</span> : null}
              <span className="text-sm text-muted-foreground">{formatExpiry(license)}</span>
            </div>
            <div className="text-xl font-semibold tracking-tight">
              {i18next.t("home:%s of %d protections are on").replace("%s", String(enabled)).replace("%d", String(protections.length))}
            </div>
            <div className="h-2 max-w-md overflow-hidden rounded-full bg-indigo-500/10">
              <div className="h-full rounded-full bg-gradient-to-r from-indigo-600 to-violet-500" style={{width: `${(enabled / protections.length) * 100}%`}} />
            </div>
            <p className="text-sm text-muted-foreground">{i18next.t("home:Business-hours support is included with your license.")}</p>
          </div>
          <div className="flex w-fit shrink-0 items-center gap-4 rounded-lg border border-dashed border-[#c9a45c]/50 bg-background/60 p-4">
            <ScoreRing score={score} className="size-[72px] blur-[6px]" trackClassName="stroke-muted" barClassName="stroke-[#c9a45c]" />
            <div className="max-w-[220px] space-y-1.5">
              <EditionBadge edition="Enterprise" />
              <div className="text-sm font-medium">{i18next.t("home:Security posture score")}</div>
              <UpgradeButton edition="Enterprise" variant="outline" className="h-8 px-3 text-xs" />
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="rounded-xl border bg-card p-5 sm:p-6">
      <div className="flex flex-col gap-5 xl:flex-row xl:items-center">
        <div className="min-w-0 flex-1 space-y-2">
          <div className="flex flex-wrap items-center gap-2">
            <Badge variant="secondary">{getEditionText("Community")}</Badge>
            {lockedFindings > 0 ? (
              <span className="inline-flex items-center gap-1 text-sm text-warning">
                <ShieldAlert className="size-4" />
                {i18next.t("home:%s findings below can be fixed by the commercial edition").replace("%s", String(lockedFindings))}
              </span>
            ) : null}
          </div>
          <div className="text-xl font-semibold tracking-tight">{i18next.t("home:Protect your users with the commercial edition")}</div>
          <p className="max-w-2xl text-sm text-muted-foreground">
            {i18next.t("home:Idle timeouts, exclusive sign-in, leaked password checks, credential stuffing defense, new device alerts, key rotation and business-hours support, on the same Casdoor you run now.")}
          </p>
        </div>
        <div className="flex shrink-0 flex-wrap gap-2">
          <UpgradeButton edition="Commercial" />
          <UpgradeButton edition="Enterprise" variant="outline" />
        </div>
      </div>
    </div>
  );
}

/** The checks with what each found; a check the edition does not cover shows its finding but locks its fix. */
export function SecurityChecks({edition, checks}: {edition: string; checks: SecurityCheck[]}) {
  return (
    <Card>
      <CardHeader className="flex-row items-center justify-between gap-3 space-y-0">
        <CardTitle className="text-base">{i18next.t("home:Security checks")}</CardTitle>
        <span className="text-xs text-muted-foreground">
          {i18next.t("home:%s of %d pass").replace("%s", String(checks.filter((c) => c.ok).length)).replace("%d", String(checks.length))}
        </span>
      </CardHeader>
      <CardContent className="p-0">
        <ul className="divide-y">
          {checks.map((check) => {
            const unlocked = hasEdition(edition, check.edition);
            return (
              <li key={check.key} className="flex flex-wrap items-center gap-x-3 gap-y-1 px-4 py-3 sm:px-6">
                <span
                  aria-hidden
                  className={cn(
                    "flex size-8 shrink-0 items-center justify-center rounded-full",
                    check.ok ? "bg-success/10 text-success" : "bg-warning/10 text-warning",
                  )}
                >
                  {check.ok ? <ShieldCheck className="size-4" /> : <ShieldAlert className="size-4" />}
                </span>
                <div className="min-w-[12rem] flex-1">
                  <div className="flex flex-wrap items-center gap-2 text-sm font-medium">
                    {check.title}
                    {check.edition !== "Community" && !unlocked ? <EditionBadge edition={check.edition} /> : null}
                  </div>
                  <div className="text-xs text-muted-foreground">{check.detail}</div>
                </div>
                {check.ok ? (
                  <Check className="size-4 shrink-0 text-success" />
                ) : unlocked ? (
                  <Button asChild variant="ghost" size="sm" className="ml-auto h-auto min-h-8 whitespace-normal text-right text-xs">
                    <Link to={check.to}>
                      {check.fix}
                      <ArrowRight />
                    </Link>
                  </Button>
                ) : (
                  <span className="ml-auto inline-flex items-center gap-1 text-right text-xs text-muted-foreground" title={check.fix}>
                    <Lock className="size-3.5" />
                    {check.fix}
                  </span>
                )}
              </li>
            );
          })}
        </ul>
      </CardContent>
    </Card>
  );
}

/** Sign-ins and what the commercial features caught lately; the community edition sees it blurred behind its offer. */
export function SigninRisk({edition, data}: {edition: string; data: DashboardSecurity}) {
  const unlocked = hasEdition(edition, "Commercial");
  const chartData = data.signinDates.map((date, index) => ({
    date: date.slice(5),
    ok: data.signins[index] ?? 0,
    failed: data.failedSignins[index] ?? 0,
  }));
  const ok = data.signins.reduce((a, b) => a + b, 0);
  const failed = data.failedSignins.reduce((a, b) => a + b, 0);
  const failedRate = ok + failed > 0 ? Math.round((failed / (ok + failed)) * 1000) / 10 : 0;
  const stats = [
    {label: i18next.t("home:Failed sign-ins"), value: `${failed}`, sub: `${failedRate}%`},
    {label: i18next.t("home:New places"), value: `${data.newLocationSignins}`, sub: i18next.t("home:%s days").replace("%s", String(data.recentDays))},
    {label: i18next.t("home:New devices"), value: `${data.newDeviceSignins}`, sub: i18next.t("home:%s days").replace("%s", String(data.recentDays))},
    {label: i18next.t("home:Blocked agent calls"), value: `${data.mcpBlockedCalls}`, sub: i18next.t("home:of %s").replace("%s", String(data.mcpCalls))},
  ];

  return (
    <Card className="relative flex flex-col overflow-hidden">
      <CardHeader className="flex-row items-center justify-between gap-3 space-y-0">
        <CardTitle className="text-base">{i18next.t("home:Sign-in risk")}</CardTitle>
        {unlocked ? null : <EditionBadge edition="Commercial" />}
      </CardHeader>
      <CardContent className="flex flex-1 flex-col">
        <div className={cn("flex flex-1 flex-col gap-4", !unlocked && "pointer-events-none select-none blur-[5px]")} aria-hidden={!unlocked}>
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
            {stats.map((stat) => (
              <div key={stat.label} className="rounded-lg border p-3">
                <div className="truncate text-xs text-muted-foreground">{stat.label}</div>
                <div className="flex items-baseline gap-1.5">
                  <span className="text-xl font-semibold tabular-nums">{stat.value}</span>
                  <span className="truncate text-xs text-muted-foreground">{stat.sub}</span>
                </div>
              </div>
            ))}
          </div>
          <div className="min-h-[220px] w-full flex-1">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={chartData} margin={{top: 4, right: 8, bottom: 0, left: 0}}>
                <CartesianGrid strokeDasharray="3 3" vertical={false} className="stroke-border" />
                <XAxis dataKey="date" tick={{fontSize: 11}} className="fill-muted-foreground" />
                <YAxis tick={{fontSize: 11}} className="fill-muted-foreground" width={36} allowDecimals={false} />
                <RechartsTooltip
                  cursor={{fill: "hsl(var(--muted))"}}
                  contentStyle={{background: "hsl(var(--popover))", border: "1px solid hsl(var(--border))", borderRadius: 8, fontSize: 12}}
                />
                <Legend wrapperStyle={{fontSize: 12}} />
                <Bar dataKey="ok" name={i18next.t("home:Successful")} stackId="signin" fill="hsl(var(--chart-1))" />
                <Bar dataKey="failed" name={i18next.t("webhook:Failed")} stackId="signin" fill="hsl(var(--destructive))" radius={[3, 3, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>
        {unlocked ? null : (
          <div className="absolute inset-0 top-14 flex flex-col items-center justify-center gap-3 bg-background/40 px-6 text-center">
            <Lock className="size-6 text-muted-foreground" />
            <div className="max-w-sm text-sm">
              {i18next.t("home:See failed sign-ins, sign-ins from new places and devices, and blocked agent calls with the commercial edition.")}
            </div>
            <UpgradeButton edition="Commercial" />
          </div>
        )}
      </CardContent>
    </Card>
  );
}
