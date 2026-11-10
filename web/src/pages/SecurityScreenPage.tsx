import * as React from "react";
import i18next from "i18next";
import {useSearchParams} from "react-router-dom";
import {AlertTriangle, Ban, Fingerprint, Laptop, MapPin, Maximize, ShieldAlert, ShieldX, UserCog, XCircle} from "lucide-react";
import {Bar, BarChart, CartesianGrid, Legend, ResponsiveContainer, Tooltip, XAxis, YAxis} from "recharts";
import {getPostureScore, getSecurityChecks} from "@/components/dashboard/EditionDashboard";
import {useAccount} from "@/hooks/use-account";
import * as EnterpriseBackend from "@/backend/EnterpriseBackend";

const STAGE_WIDTH = 1920;
const STAGE_HEIGHT = 1080;
const REFRESH_SECONDS = 30;

// validated against the panel surface for color vision deficiencies
const SUCCESS = "#0284c7";
const FAILURE = "#f43f5e";

const panelClass = "rounded-lg border border-[#1d3557] bg-[#0b1626]/90 shadow-[inset_0_0_24px_rgba(14,165,233,0.06)]";

function Panel({title, extra, className, children}: {title: string; extra?: React.ReactNode; className?: string; children: React.ReactNode}) {
  return (
    <section className={`${panelClass} flex min-h-0 flex-col p-4 ${className ?? ""}`}>
      <div className="mb-3 flex items-center gap-2">
        <span className="h-4 w-1 rounded bg-sky-500" />
        <h2 className="text-[17px] font-semibold tracking-wide text-slate-100">{title}</h2>
        {extra ? <span className="ml-auto text-xs text-slate-400">{extra}</span> : null}
      </div>
      <div className="min-h-0 flex-1 overflow-hidden">{children}</div>
    </section>
  );
}

function Kpi({label, value, sub, tone}: {label: string; value: React.ReactNode; sub?: React.ReactNode; tone?: "bad" | "warn"}) {
  const color = tone === "bad" ? "text-rose-400" : tone === "warn" ? "text-amber-300" : "text-slate-50";
  return (
    <div className={`${panelClass} flex flex-col justify-center px-5 py-3`}>
      <div className="truncate text-[14px] text-slate-400" title={label}>{label}</div>
      <div className={`mt-1 text-[40px] font-bold leading-none tabular-nums ${color}`}>{value}</div>
      <div className="mt-1.5 h-4 text-xs text-slate-500">{sub}</div>
    </div>
  );
}

function RankedBars({items, empty, nameClassName}: {items: {name: string; count: number; failure?: number}[]; empty?: string; nameClassName?: string}) {
  if (items.length === 0) {
    return <div className="flex h-full items-center justify-center text-sm text-slate-500">{empty ?? i18next.t("general:No data")}</div>;
  }
  const max = Math.max(1, ...items.map((item) => item.count));
  return (
    <ol className="space-y-2">
      {items.map((item, index) => (
        <li key={item.name} className="grid grid-cols-[22px_1fr_auto] items-center gap-2 text-[13px]" title={`${item.name}: ${item.count}`}>
          <span className={`text-center text-xs font-semibold tabular-nums ${index < 3 ? "text-sky-300" : "text-slate-500"}`}>{index + 1}</span>
          <div className="min-w-0">
            <div className={`truncate text-slate-200 ${nameClassName ?? ""}`}>{item.name}</div>
            <div className="mt-1 flex h-1.5 overflow-hidden rounded-full bg-slate-800">
              <div className="h-full rounded-l-full" style={{width: `${((item.count - (item.failure ?? 0)) / max) * 100}%`, background: SUCCESS}} />
              {item.failure ? <div className="ml-[2px] h-full rounded-r-full" style={{width: `${(item.failure / max) * 100}%`, background: FAILURE}} /> : null}
            </div>
          </div>
          <span className="tabular-nums text-slate-300">{item.count.toLocaleString()}</span>
        </li>
      ))}
    </ol>
  );
}

function ScoreRing({score}: {score: number}) {
  const radius = 62;
  const circumference = 2 * Math.PI * radius;
  const color = score >= 85 ? "#22c55e" : score >= 65 ? "#f59e0b" : FAILURE;
  return (
    <svg viewBox="0 0 150 150" className="size-[136px] shrink-0">
      <circle cx="75" cy="75" r={radius} fill="none" stroke="#1e293b" strokeWidth="12" />
      <circle cx="75" cy="75" r={radius} fill="none" stroke={color} strokeWidth="12" strokeLinecap="round"
        strokeDasharray={circumference} strokeDashoffset={circumference * (1 - Math.min(Math.max(score, 0), 100) / 100)} transform="rotate(-90 75 75)" />
      <text x="75" y="84" textAnchor="middle" className="fill-slate-50 text-[38px] font-bold">{score}</text>
    </svg>
  );
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

const eventKinds: Record<string, {label: string; icon: React.ElementType; className: string}> = {
  "failed-signin": {label: "enterprise:Failed sign-in", icon: XCircle, className: "text-amber-300"},
  "blocked-signin": {label: "enterprise:Blocked sign-in", icon: Ban, className: "text-rose-400"},
  "admin-operation": {label: "enterprise:Admin operation", icon: UserCog, className: "text-sky-300"},
  "denied": {label: "enterprise:Denied request", icon: ShieldX, className: "text-orange-300"},
  "new-location": {label: "enterprise:New location", icon: MapPin, className: "text-violet-300"},
  "new-device": {label: "enterprise:New device", icon: Laptop, className: "text-violet-300"},
  "mcp-blocked": {label: "enterprise:MCP call blocked", icon: ShieldAlert, className: "text-rose-400"},
};

const alertKinds: Record<string, string> = {
  "ip-blocked": "enterprise:IPs blocked for credential stuffing (%s)",
  "accounts-frozen": "enterprise:Sign-ins refused to locked accounts today (%s)",
  "high-failure-rate": "enterprise:Unusually many failed sign-ins today (%s)",
  "admins-without-mfa": "home:Admins without MFA (%s)",
  "cert-expiring": "enterprise:Certificates expiring soon (%s)",
  "orphan-accounts": "enterprise:Orphan accounts to clean up (%s)",
  "sod-conflicts": "enterprise:Separation-of-duty conflicts found (%s)",
  "mcp-blocked": "enterprise:MCP calls blocked today (%s)",
};

function placeName(region: any) {
  if (region.kind === "unknown") {
    return i18next.t("webhook:Unknown");
  }
  if (region.kind === "intranet") {
    return i18next.t("enterprise:Intranet");
  }
  return region.name;
}

function formatEventTime(value: string) {
  const time = new Date(value);
  if (isNaN(time.getTime())) {
    return value;
  }
  const pad = (n: number) => String(n).padStart(2, "0");
  const clock = `${pad(time.getHours())}:${pad(time.getMinutes())}:${pad(time.getSeconds())}`;
  return time.toDateString() === new Date().toDateString() ? clock : `${pad(time.getMonth() + 1)}-${pad(time.getDate())} ${clock}`;
}

/** Keeps the 1920x1080 stage as large as the window allows, centered, whatever the screen's own ratio. */
function useStageScale() {
  const [scale, setScale] = React.useState(1);
  React.useEffect(() => {
    const update = () => setScale(Math.min(window.innerWidth / STAGE_WIDTH, window.innerHeight / STAGE_HEIGHT));
    update();
    window.addEventListener("resize", update);
    return () => window.removeEventListener("resize", update);
  }, []);
  return scale;
}

function useClock() {
  const [now, setNow] = React.useState(() => new Date());
  React.useEffect(() => {
    const timer = window.setInterval(() => setNow(new Date()), 1000);
    return () => window.clearInterval(timer);
  }, []);
  return now;
}

function TrendChart({data, xKey, interval}: {data: any[]; xKey: string; interval: number}) {
  return (
    <ResponsiveContainer width="100%" height="100%">
      <BarChart data={data} margin={{top: 4, right: 8, left: -12, bottom: 0}} barCategoryGap="22%">
        <CartesianGrid vertical={false} stroke="#1e293b" />
        <XAxis dataKey={xKey} tick={{fill: "#94a3b8", fontSize: 12}} tickLine={false} axisLine={{stroke: "#334155"}} interval={interval} />
        <YAxis allowDecimals={false} tickCount={4} tick={{fill: "#94a3b8", fontSize: 12}} tickLine={false} axisLine={false} width={48} />
        <Tooltip cursor={{fill: "rgba(148,163,184,0.08)"}} contentStyle={{background: "#0f172a", border: "1px solid #334155", borderRadius: 6, color: "#e2e8f0"}} labelStyle={{color: "#cbd5e1"}} />
        <Legend iconType="circle" iconSize={8} wrapperStyle={{fontSize: 12, color: "#cbd5e1"}} />
        <Bar dataKey="success" name={i18next.t("home:Successful")} stackId="a" fill={SUCCESS} stroke="#0b1626" strokeWidth={1} isAnimationActive={false} />
        <Bar dataKey="failure" name={i18next.t("webhook:Failed")} stackId="a" fill={FAILURE} stroke="#0b1626" strokeWidth={1} radius={[3, 3, 0, 0]} isAnimationActive={false} />
      </BarChart>
    </ResponsiveContainer>
  );
}

function EventTicker({events}: {events: any[]}) {
  if (events.length === 0) {
    return <div className="flex h-full items-center justify-center text-sm text-slate-500">{i18next.t("general:No data")}</div>;
  }
  const rows = (keyPrefix: string) => events.map((event, index) => {
    const kind = eventKinds[event.kind] ?? {label: event.kind, icon: AlertTriangle, className: "text-slate-300"};
    const Icon = kind.icon;
    return (
      <li key={`${keyPrefix}${index}`} className="grid grid-cols-[86px_1fr] gap-2 border-b border-slate-800/80 py-2 text-[13px]">
        <span className="tabular-nums text-slate-500">{formatEventTime(event.time)}</span>
        <div className="min-w-0">
          <div className={`flex items-center gap-1.5 font-medium ${kind.className}`}>
            <Icon className="size-3.5 shrink-0" />
            <span>{i18next.t(kind.label)}</span>
            <span className="truncate text-slate-200">{event.user}</span>
          </div>
          <div className="truncate text-xs text-slate-500">{[event.ip, event.detail].filter(Boolean).join("  ")}</div>
        </div>
      </li>
    );
  });
  // a long list scrolls by itself, twice in a row so that the loop has no seam
  const scrolling = events.length > 9;
  return (
    <div className="h-full overflow-hidden">
      <ul className={scrolling ? "security-ticker" : ""} style={scrolling ? {animationDuration: `${events.length * 3}s`} : undefined}>
        {rows("a")}
        {scrolling ? rows("b") : null}
      </ul>
    </div>
  );
}

/** The security situation screen of an organization for a wall display: dark, 16:9, refreshed every 30 seconds. */
export default function SecurityScreenPage() {
  const {account} = useAccount();
  const [searchParams] = useSearchParams();
  const owner = searchParams.get("owner") || account?.owner || "";
  const [data, setData] = React.useState<any>(null);
  const [error, setError] = React.useState("");
  const [updatedAt, setUpdatedAt] = React.useState<Date | null>(null);
  const scale = useStageScale();
  const now = useClock();

  React.useEffect(() => {
    if (!owner) {
      return;
    }
    let stopped = false;
    const load = () => EnterpriseBackend.getSecurityScreen(owner).then((res: any) => {
      if (stopped) {
        return;
      }
      if (res.status === "ok") {
        setData(res.data);
        setError("");
        setUpdatedAt(new Date());
      } else {
        setError(res.msg);
      }
    }).catch(() => {});
    load();
    const timer = window.setInterval(load, REFRESH_SECONDS * 1000);
    return () => {
      stopped = true;
      window.clearInterval(timer);
    };
  }, [owner]);

  React.useEffect(() => {
    document.title = `${i18next.t("enterprise:Identity security situation")} - ${data?.displayName ?? owner}`;
  }, [data, owner]);

  const toggleFullscreen = () => {
    if (document.fullscreenElement) {
      document.exitFullscreen();
    } else {
      document.documentElement.requestFullscreen?.();
    }
  };

  const pad = (n: number) => String(n).padStart(2, "0");
  const locale = i18next.language === "zh" ? "zh-CN" : i18next.language;
  const dateText = now.toLocaleDateString(locale, {year: "numeric", month: "long", day: "numeric", weekday: "long"});
  const timeText = `${pad(now.getHours())}:${pad(now.getMinutes())}:${pad(now.getSeconds())}`;

  const security = data?.security;
  const checks = security ? getSecurityChecks(security, owner) : [];
  const score = security ? getPostureScore(security, checks) : 0;
  const mfaRate = data && data.users > 0 ? (data.mfaUsers / data.users) * 100 : 0;

  return (
    <div className="fixed inset-0 overflow-hidden bg-[#050b16] text-slate-100">
      <style>{`
        @keyframes security-ticker { from { transform: translateY(0); } to { transform: translateY(-50%); } }
        .security-ticker { animation-name: security-ticker; animation-timing-function: linear; animation-iteration-count: infinite; }
        .security-ticker:hover { animation-play-state: paused; }
        @media (prefers-reduced-motion: reduce) { .security-ticker { animation: none; } }
      `}</style>
      <div
        className="absolute left-1/2 top-1/2 flex flex-col gap-4 p-6"
        style={{
          width: STAGE_WIDTH,
          height: STAGE_HEIGHT,
          transform: `translate(-50%, -50%) scale(${scale})`,
          background: "radial-gradient(ellipse at 50% -10%, rgba(14,165,233,0.16), transparent 55%), linear-gradient(180deg, #06101f, #050b16)",
        }}
      >
        {/* header */}
        <header className="grid h-[64px] shrink-0 grid-cols-[1fr_auto_1fr] items-center">
          <div className="min-w-0">
            <div className="truncate text-[22px] font-semibold text-slate-100">{data?.displayName ?? owner}</div>
            <div className="text-xs text-slate-500">{owner}</div>
          </div>
          <h1 className="bg-gradient-to-b from-white to-sky-300 bg-clip-text px-10 text-[34px] font-bold tracking-[0.12em] text-transparent">
            {i18next.t("enterprise:Identity security situation")}
          </h1>
          <div className="flex items-center justify-end gap-5">
            <div className="text-right">
              <div className="text-[28px] font-semibold leading-none tabular-nums text-sky-200">{timeText}</div>
              <div className="mt-1 text-xs text-slate-400">{dateText}</div>
            </div>
            <button type="button" onClick={toggleFullscreen} className="rounded-md border border-slate-700 p-2 text-slate-400 hover:text-slate-100" title={i18next.t("enterprise:Full screen")}>
              <Maximize className="size-5" />
            </button>
          </div>
        </header>

        {error && !data ? (
          <div className="flex flex-1 flex-col items-center justify-center gap-3 text-center">
            <ShieldAlert className="size-12 text-amber-300" />
            <p className="max-w-3xl text-lg text-slate-200">{error}</p>
          </div>
        ) : !data ? (
          <div className="flex flex-1 items-center justify-center text-slate-500">{i18next.t("general:Loading")}</div>
        ) : (
          <>
            {/* key numbers */}
            <div className="grid h-[118px] shrink-0 grid-cols-8 gap-4">
              <Kpi label={i18next.t("enterprise:Successful sign-ins today")} value={data.signinSuccessToday.toLocaleString()} />
              <Kpi label={i18next.t("enterprise:Failed sign-ins today")} value={data.signinFailureToday.toLocaleString()} tone={data.signinFailureToday > 0 ? "warn" : undefined}
                sub={`${i18next.t("enterprise:Failure rate")} ${(data.failureRateToday * 100).toFixed(1)}%`} />
              <Kpi label={i18next.t("enterprise:Blocked attempts today")} value={data.blockedToday + data.mcpBlockedToday} tone={data.blockedToday + data.mcpBlockedToday > 0 ? "bad" : undefined}
                sub={data.mcpBlockedToday > 0 ? `MCP ${data.mcpBlockedToday}` : undefined} />
              <Kpi label={i18next.t("enterprise:Active users today")} value={data.activeUsersToday.toLocaleString()} />
              <Kpi label={i18next.t("enterprise:Active users in 7 days")} value={data.activeUsers7Days.toLocaleString()} />
              <Kpi label={i18next.t("enterprise:Active sessions")} value={data.activeSessions.toLocaleString()} />
              <Kpi label={i18next.t("enterprise:MFA coverage")} value={`${mfaRate.toFixed(1)}%`} sub={`${data.mfaUsers} / ${data.users}`} />
              <Kpi label={i18next.t("enterprise:Admins without MFA")} value={security?.adminsWithoutMfa ?? 0} tone={security?.adminsWithoutMfa > 0 ? "bad" : undefined}
                sub={`${security?.admins ?? 0} ${i18next.t("enterprise:admins in total")}`} />
            </div>

            <div className="grid min-h-0 flex-1 grid-cols-[440px_1fr_440px] gap-4">
              {/* left */}
              <div className="grid min-h-0 grid-rows-[220px_1fr_226px] gap-4">
                <Panel title={i18next.t("home:Security posture score")}>
                  <div className="flex h-full items-center gap-4">
                    <ScoreRing score={score} />
                    <div className="min-w-0 flex-1 space-y-1.5">
                      <div className="text-lg font-semibold">{getGrade(score)}</div>
                      {checks.filter((check) => !check.ok).slice(0, 4).map((check) => (
                        <div key={check.key} className="flex items-center gap-1.5 text-[13px] text-slate-300">
                          <XCircle className="size-3.5 shrink-0 text-rose-400" />
                          <span className="truncate">{check.title}</span>
                        </div>
                      ))}
                      {checks.every((check) => check.ok) ? <div className="text-[13px] text-slate-400">{i18next.t("enterprise:Every check is met")}</div> : null}
                    </div>
                  </div>
                </Panel>
                <Panel title={i18next.t("organization:Alerts")} extra={data.alerts.length > 0 ? String(data.alerts.length) : undefined}>
                  {data.alerts.length === 0 ? (
                    <div className="flex h-full items-center justify-center text-sm text-emerald-300">{i18next.t("enterprise:Nothing needs attention")}</div>
                  ) : (
                    <ul className="space-y-1.5">
                      {data.alerts.map((alert: any) => (
                        <li key={alert.kind} className={`flex items-center gap-2 rounded-md border-l-[3px] px-3 py-1.5 text-[13px] ${alert.level === "critical" ? "border-rose-500 bg-rose-500/10" : "border-amber-400 bg-amber-400/10"}`}>
                          {alert.level === "critical" ? <ShieldAlert className="size-4 shrink-0 text-rose-400" /> : <AlertTriangle className="size-4 shrink-0 text-amber-300" />}
                          <span className="shrink-0 font-medium">{i18next.t(alertKinds[alert.kind] ?? alert.kind).replace("%s", String(alert.count))}</span>
                          {alert.detail ? <span className="truncate text-xs text-slate-400" title={alert.detail}>{alert.detail}</span> : null}
                        </li>
                      ))}
                    </ul>
                  )}
                </Panel>
                <Panel title={i18next.t("enterprise:Most used applications")} extra={i18next.t("enterprise:Last %s days").replace("%s", "7")}>
                  <RankedBars items={data.topApplications.slice(0, 4)} />
                </Panel>
              </div>

              {/* center */}
              <div className="grid min-h-0 grid-rows-[1fr_1fr_1.15fr] gap-4">
                <Panel title={i18next.t("enterprise:Sign-ins in the last 24 hours")}>
                  <TrendChart data={data.hours} xKey="hour" interval={1} />
                </Panel>
                <Panel title={i18next.t("enterprise:Sign-ins in the last 30 days")}>
                  <TrendChart data={data.days.map((day: any) => ({...day, day: day.date.slice(5)}))} xKey="day" interval={2} />
                </Panel>
                <Panel title={i18next.t("enterprise:Where sign-ins come from")} extra={data.ipDbLoaded ? i18next.t("enterprise:Last %s days").replace("%s", "30") : i18next.t("enterprise:No IP database, unknown places are shown by network")}>
                  <div className="grid h-full grid-cols-2 gap-6">
                    <div className="min-h-0">
                      <div className="mb-2 text-xs text-slate-400">{i18next.t("enterprise:By province or country")}</div>
                      <RankedBars items={data.provinces.slice(0, 5).map((region: any) => ({...region, name: placeName(region)}))} />
                    </div>
                    <div className="min-h-0">
                      <div className="mb-2 text-xs text-slate-400">{i18next.t("enterprise:By city")}</div>
                      <RankedBars items={data.cities.slice(0, 5).map((region: any) => ({...region, name: placeName(region)}))} />
                    </div>
                  </div>
                </Panel>
              </div>

              {/* right */}
              <div className="grid min-h-0 grid-rows-[232px_232px_1fr] gap-4">
                <Panel title={i18next.t("enterprise:Top failed IPs")} extra={i18next.t("enterprise:Last %s days").replace("%s", "7")}>
                  <RankedBars items={data.topFailedIps.slice(0, 4).map((item: any) => ({...item, failure: item.count}))} empty={i18next.t("enterprise:No failed sign-in")} nameClassName="font-mono" />
                </Panel>
                <Panel title={i18next.t("enterprise:Top failed users")} extra={i18next.t("enterprise:Last %s days").replace("%s", "7")}>
                  <RankedBars items={data.topFailedUsers.slice(0, 4).map((item: any) => ({...item, failure: item.count}))} empty={i18next.t("enterprise:No failed sign-in")} />
                </Panel>
                <Panel title={i18next.t("enterprise:Latest security events")}>
                  <EventTicker events={data.events} />
                </Panel>
              </div>
            </div>
          </>
        )}

        <footer className="flex h-4 shrink-0 items-center justify-between text-[11px] text-slate-500">
          <span className="flex items-center gap-1.5"><Fingerprint className="size-3.5" />Casdoor</span>
          <span>
            {error && data ? <span className="mr-3 text-amber-300">{error}</span> : null}
            {i18next.t("enterprise:Refreshes every %s seconds").replace("%s", String(REFRESH_SECONDS))}
            {updatedAt ? `, ${i18next.t("enterprise:updated at %s").replace("%s", `${pad(updatedAt.getHours())}:${pad(updatedAt.getMinutes())}:${pad(updatedAt.getSeconds())}`)}` : ""}
          </span>
        </footer>
      </div>
    </div>
  );
}
