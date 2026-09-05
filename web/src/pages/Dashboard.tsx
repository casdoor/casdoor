import * as React from "react";
import i18next from "i18next";
import {AppWindow, Building2, KeyRound, Plug, UserPlus, Users} from "lucide-react";
import {Link, useNavigate} from "react-router-dom";
import {
  Cell,
  CartesianGrid,
  Legend,
  Line,
  LineChart,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip as RechartsTooltip,
  XAxis,
  YAxis,
} from "recharts";
import {Card, CardContent, CardHeader, CardTitle} from "@/components/ui/card";
import {ActivityHeatmap} from "@/components/common/ActivityHeatmap";
import {Loading} from "@/components/common/Loading";
import {PageHeader} from "@/components/crud/PageHeader";
import {useAccount} from "@/hooks/use-account";
import * as DashboardBackend from "@/backend/DashboardBackend";
import * as Setting from "@/lib/setting";

/**
 * The eight `--chart-*` hues of the theme, which are redefined for dark mode, so
 * a series keeps its identity while the palette brightens. More series than hues
 * simply wrap around.
 */
const CHART_COLORS = Array.from({length: 8}, (_, i) => `hsl(var(--chart-${i + 1}))`);

function buildDateArray(): string[] {
  const arr: string[] = [];
  const now = new Date();
  for (let i = 30; i >= 0; i--) {
    const d = new Date(now);
    d.setDate(d.getDate() - i);
    arr.push(`${d.getMonth() + 1}-${d.getDate()}`);
  }
  return arr;
}

const DATE_ARRAY = buildDateArray();

interface Series {
  key: string;
  name: string;
  data: number[];
  to: string;
  /** hidden by default, like the antd legend defaults */
  hidden?: boolean;
}

/**
 * Eight of these sit in a row, so the metric is carried by an icon rather than a
 * colour: eight accent hues across the top of the dashboard read as eight
 * unrelated warnings, and none of them meant anything.
 */
function StatCard({title, value, icon, to}: {title: string; value: number; icon: React.ReactNode; to: string}) {
  return (
    <Link to={to} className="group rounded-xl focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
      <Card className="h-full transition-colors group-hover:border-foreground/25">
        <CardContent className="flex items-center gap-3 p-4">
          <span
            aria-hidden
            className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-muted text-muted-foreground transition-colors group-hover:text-foreground [&_svg]:size-[18px]"
          >
            {icon}
          </span>
          <span className="min-w-0">
            <span className="block truncate text-xs text-muted-foreground">{title}</span>
            <span className="block text-2xl font-semibold leading-tight tabular-nums">{value}</span>
          </span>
        </CardContent>
      </Card>
    </Link>
  );
}

/**
 * The organization the dashboard counts. A global admin looking at "All" asks for
 * every organization (an empty owner), which `getRequestOrganization` cannot
 * express — it falls back to the admin's own organization. Port of
 * `getOrganizationName()` in web/src/basic/Dashboard.js.
 */
function getDashboardOrganization(account: any): string {
  if (!Setting.isAdminUser(account) && Setting.isLocalAdminUser(account)) {
    return account.owner;
  }
  const stored = Setting.getOrganization();
  return stored === "All" ? "" : stored || "";
}

export default function Dashboard() {
  const {account} = useAccount();
  const navigate = useNavigate();
  const [data, setData] = React.useState<any>(null);
  const [providerData, setProviderData] = React.useState<any[] | null>(null);
  const [mfaData, setMfaData] = React.useState<any>(null);
  const [heatmapData, setHeatmapData] = React.useState<any>(null);
  const [loading, setLoading] = React.useState(true);

  // a regular user has no dashboard; antd sends them to the first page their
  // organization leaves them
  const isConsoleAdmin = Setting.isLocalAdminUser(account);

  const load = React.useCallback(() => {
    if (!account || !Setting.isLocalAdminUser(account)) {
      return;
    }
    const org = getDashboardOrganization(account);
    setLoading(true);

    const apply = (settled: PromiseSettledResult<any>, setter: (value: any) => void) => {
      if (settled.status === "rejected") {
        Setting.showMessage("error", settled.reason?.message ?? String(settled.reason));
        return;
      }
      if (settled.value?.status === "ok") {
        setter(settled.value.data);
      } else if (settled.value?.msg) {
        Setting.showMessage("error", settled.value.msg);
      }
    };

    // allSettled so one failing endpoint still leaves the other cards rendered
    Promise.allSettled([
      DashboardBackend.getDashboard(org),
      DashboardBackend.getDashboardProviders(org),
      DashboardBackend.getDashboardMfa(org),
      DashboardBackend.getDashboardHeatmap(org),
    ])
      .then(([dash, providers, mfa, heatmap]) => {
        apply(dash, setData);
        apply(providers, setProviderData);
        apply(mfa, setMfaData);
        apply(heatmap, setHeatmapData);
      })
      .finally(() => setLoading(false));
  }, [account]);

  React.useEffect(() => {
    if (!account || isConsoleAdmin) {
      load();
      return;
    }
    const navItems = account.organization?.userNavItems;
    const isAllEnabled = !Array.isArray(navItems) || navItems.includes("all");
    if (isAllEnabled || navItems.includes("/apps")) {
      navigate("/apps", {replace: true});
    } else if (navItems.includes("/shortcuts")) {
      navigate("/shortcuts", {replace: true});
    } else {
      navigate("/account", {replace: true});
    }
  }, [account, isConsoleAdmin, load, navigate]);

  React.useEffect(() => {
    window.addEventListener("storageOrganizationChanged", load);
    return () => window.removeEventListener("storageOrganizationChanged", load);
  }, [load]);

  if (loading && !data) {
    return <Loading />;
  }

  if (!data) {
    return (
      <div className="py-16 text-center text-muted-foreground">{i18next.t("general:No data")}</div>
    );
  }

  const series: Series[] = [
    {key: "userCounts", name: i18next.t("general:Users"), data: data.userCounts, to: "/users"},
    {key: "applicationCounts", name: i18next.t("general:Applications"), data: data.applicationCounts, to: "/applications"},
    {key: "providerCounts", name: i18next.t("application:Providers"), data: data.providerCounts, to: "/providers"},
    {key: "organizationCounts", name: i18next.t("general:Organizations"), data: data.organizationCounts, to: "/organizations"},
    {key: "roleCounts", name: i18next.t("general:Roles"), data: data.roleCounts, to: "/roles"},
    {key: "permissionCounts", name: i18next.t("general:Permissions"), data: data.permissionCounts, to: "/permissions"},
    {key: "groupCounts", name: i18next.t("general:Groups"), data: data.groupCounts, to: "/groups"},
    {key: "resourceCounts", name: i18next.t("general:Resources"), data: data.resourceCounts, to: "/resources"},
    {key: "certCounts", name: i18next.t("general:Certs"), data: data.certCounts, to: "/certs"},
    {key: "subscriptionCounts", name: i18next.t("general:Subscriptions"), data: data.subscriptionCounts, to: "/subscriptions", hidden: true},
    {key: "modelCounts", name: i18next.t("general:Models"), data: data.modelCounts, to: "/models", hidden: true},
    {key: "transactionCounts", name: i18next.t("general:Transactions"), data: data.transactionCounts, to: "/transactions"},
    {key: "adapterCounts", name: i18next.t("general:Adapters"), data: data.adapterCounts, to: "/adapters", hidden: true},
    {key: "enforcerCounts", name: i18next.t("general:Enforcers"), data: data.enforcerCounts, to: "/enforcers", hidden: true},
  ].filter((s) => Array.isArray(s.data));

  const chartData = DATE_ARRAY.map((date, index) => {
    const row: Record<string, any> = {date};
    series.forEach((s) => {
      row[s.key] = s.data[index];
    });
    return row;
  });

  const at = (key: string, index: number) => (Array.isArray(data[key]) ? data[key][index] ?? 0 : 0);
  const userCounts: number[] = data.userCounts ?? Array(31).fill(0);

  const mfaRate = mfaData && mfaData.total > 0 ? (mfaData.enabled / mfaData.total) * 100 : 0;
  const providerSlices = (providerData ?? []).map((item: any) => ({
    name: item.type || i18next.t("general:None"),
    value: item.count,
  }));

  return (
    <div className="space-y-5">
      <PageHeader title={i18next.t("general:Dashboard")} />

      {/* Key metrics */}
      {/* Eight across only once there is room for the labels; below that they
          truncate to "New users / 30..." and stop meaning anything */}
      <div id="statistic" className="grid grid-cols-2 gap-3 md:grid-cols-4 2xl:grid-cols-8">
        <StatCard title={i18next.t("home:Total users")} value={userCounts[30] ?? 0} icon={<Users />} to="/users" />
        <StatCard
          title={i18next.t("home:New users today")}
          value={Math.max(0, (userCounts[30] ?? 0) - (userCounts[29] ?? 0))}
          icon={<UserPlus />}
          to="/users"
        />
        <StatCard
          title={i18next.t("home:New users / 7 days")}
          value={Math.max(0, (userCounts[30] ?? 0) - (userCounts[23] ?? 0))}
          icon={<UserPlus />}
          to="/users"
        />
        <StatCard
          title={i18next.t("home:New users / 30 days")}
          value={Math.max(0, (userCounts[30] ?? 0) - (userCounts[0] ?? 0))}
          icon={<UserPlus />}
          to="/users"
        />
        <StatCard title={i18next.t("general:Organizations")} value={at("organizationCounts", 30)} icon={<Building2 />} to="/organizations" />
        <StatCard title={i18next.t("general:Tokens")} value={at("tokenCounts", 30)} icon={<KeyRound />} to="/tokens" />
        <StatCard title={i18next.t("general:Applications")} value={at("applicationCounts", 30)} icon={<AppWindow />} to="/applications" />
        <StatCard title={i18next.t("application:Providers")} value={at("providerCounts", 30)} icon={<Plug />} to="/providers" />
      </div>

      {/* 30-day trend + provider distribution */}
      <div className="grid gap-4 xl:grid-cols-[7fr_5fr]">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">{i18next.t("home:Past 30 days")}</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-[320px] w-full">
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={chartData} margin={{top: 8, right: 16, bottom: 8, left: 0}}>
                  <CartesianGrid strokeDasharray="3 3" className="stroke-border" />
                  <XAxis dataKey="date" tick={{fontSize: 11}} className="fill-muted-foreground" />
                  <YAxis tick={{fontSize: 11}} className="fill-muted-foreground" width={44} allowDecimals={false} />
                  <RechartsTooltip
                    contentStyle={{
                      background: "hsl(var(--popover))",
                      border: "1px solid hsl(var(--border))",
                      borderRadius: 8,
                      fontSize: 12,
                    }}
                  />
                  <Legend wrapperStyle={{fontSize: 12}} />
                  {series
                    .filter((s) => !s.hidden)
                    .map((s, index) => (
                      <Line
                        key={s.key}
                        type="monotone"
                        dataKey={s.key}
                        name={s.name}
                        dot={false}
                        strokeWidth={2}
                        stroke={CHART_COLORS[index % CHART_COLORS.length]}
                      />
                    ))}
                </LineChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">{i18next.t("application:Providers")}</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex h-[320px] w-full items-center justify-center">
              {providerSlices.length > 0 ? (
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie
                      data={providerSlices}
                      dataKey="value"
                      nameKey="name"
                      innerRadius="45%"
                      outerRadius="70%"
                      paddingAngle={2}
                    >
                      {providerSlices.map((_, index) => (
                        <Cell key={index} fill={CHART_COLORS[index % CHART_COLORS.length]} />
                      ))}
                    </Pie>
                    <RechartsTooltip
                      contentStyle={{
                        background: "hsl(var(--popover))",
                        border: "1px solid hsl(var(--border))",
                        borderRadius: 8,
                        fontSize: 12,
                      }}
                    />
                    <Legend wrapperStyle={{fontSize: 12}} />
                  </PieChart>
                </ResponsiveContainer>
              ) : (
                <div className="flex flex-col items-center gap-2 text-muted-foreground">
                  <Plug className="h-8 w-8 opacity-40" strokeWidth={1.5} />
                  <span className="text-sm">{i18next.t("general:No data")}</span>
                  <Link to="/providers" className="text-sm underline underline-offset-4 hover:text-foreground">
                    {i18next.t("application:Providers")}
                  </Link>
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* MFA coverage + sign-in heatmap */}
      <div className="grid gap-4 xl:grid-cols-[4fr_8fr]">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">{i18next.t("user:MFA accounts")}</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex h-[200px] items-center justify-center gap-10">
              {mfaData ? (
                <>
                  <MfaRing percent={mfaRate} />
                  <div className="space-y-4 text-sm">
                    <div>
                      <div className="text-2xl font-semibold text-primary">{mfaData.enabled}</div>
                      <div className="text-muted-foreground">{i18next.t("general:Enabled")}</div>
                    </div>
                    <div>
                      <div className="text-2xl font-semibold text-muted-foreground">{mfaData.disabled}</div>
                      <div className="text-muted-foreground">{i18next.t("general:Disable")}</div>
                    </div>
                  </div>
                </>
              ) : (
                <div className="flex flex-col items-center gap-2 text-muted-foreground">
                  <Plug className="h-8 w-8 opacity-40" strokeWidth={1.5} />
                  <span className="text-sm">{i18next.t("general:No data")}</span>
                  <Link to="/providers" className="text-sm underline underline-offset-4 hover:text-foreground">
                    {i18next.t("application:Providers")}
                  </Link>
                </div>
              )}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">{i18next.t("application:Signin")}</CardTitle>
          </CardHeader>
          <CardContent>
            {heatmapData?.data ? (
              <ActivityHeatmap
                data={heatmapData.data}
                maxCount={heatmapData.maxCount}
                dateRange={heatmapData.dateRange}
              />
            ) : (
              <div className="flex h-[180px] items-center justify-center text-muted-foreground">
                {i18next.t("general:None")}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

/** the circular MFA-adoption gauge, drawn as a plain SVG ring */
function MfaRing({percent}: {percent: number}) {
  const radius = 56;
  const circumference = 2 * Math.PI * radius;
  const offset = circumference * (1 - Math.min(Math.max(percent, 0), 100) / 100);

  return (
    <div className="relative h-[130px] w-[130px]">
      <svg viewBox="0 0 130 130" className="h-full w-full -rotate-90">
        <circle cx="65" cy="65" r={radius} className="fill-none stroke-muted" strokeWidth="10" />
        <circle
          cx="65"
          cy="65"
          r={radius}
          className="fill-none stroke-primary transition-[stroke-dashoffset]"
          strokeWidth="10"
          strokeLinecap="round"
          strokeDasharray={circumference}
          strokeDashoffset={offset}
        />
      </svg>
      <div className="absolute inset-0 flex flex-col items-center justify-center">
        <span className="text-xl font-bold text-primary">{percent.toFixed(2)}%</span>
        <span className="text-xs text-muted-foreground">MFA</span>
      </div>
    </div>
  );
}
