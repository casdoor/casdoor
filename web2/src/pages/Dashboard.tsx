import * as React from "react";
import i18next from "i18next";
import {Link} from "react-router-dom";
import {
  CartesianGrid,
  Legend,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip as RechartsTooltip,
  XAxis,
  YAxis,
} from "recharts";
import {Card, CardContent, CardHeader, CardTitle} from "@/components/ui/card";
import {Loading} from "@/components/common/Loading";
import {PageHeader} from "@/components/crud/PageHeader";
import {useAccount} from "@/hooks/use-account";
import * as DashboardBackend from "@/backend/DashboardBackend";
import * as Setting from "@/lib/setting";

// Same multi-hue palette the antd dashboard used.
const CHART_COLORS = [
  "#1677ff", "#0ea5e9", "#06b6d4", "#14b8a6", "#6366f1", "#8b5cf6", "#0958d9",
  "#0284c7", "#0891b2", "#0f766e", "#5734d3", "#7c3aed", "#38bdf8", "#5eead4",
];

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

export default function Dashboard() {
  const {account} = useAccount();
  const [data, setData] = React.useState<any>(null);
  const [loading, setLoading] = React.useState(true);

  React.useEffect(() => {
    if (!account) {
      return;
    }
    setLoading(true);
    DashboardBackend.getDashboard(Setting.getRequestOrganization(account))
      .then((res: any) => {
        if (res.status === "ok") {
          setData(res.data);
        } else {
          Setting.showMessage("error", res.msg);
        }
      })
      .finally(() => setLoading(false));
  }, [account]);

  if (loading) {
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
    {key: "subscriptionCounts", name: i18next.t("general:Subscriptions"), data: data.subscriptionCounts, to: "/subscriptions"},
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

  const latest = (s: Series) => (s.data.length > 0 ? s.data[s.data.length - 1] : 0);
  const previous = (s: Series) => (s.data.length > 1 ? s.data[s.data.length - 2] : 0);

  return (
    <div className="space-y-5">
      <PageHeader title={i18next.t("general:Dashboard")} />

      <div className="grid grid-cols-2 gap-3 md:grid-cols-3 xl:grid-cols-5">
        {series
          .filter((s) => !s.hidden)
          .map((s, index) => {
            const delta = latest(s) - previous(s);
            return (
              <Link key={s.key} to={s.to}>
                <Card className="transition-colors hover:border-foreground/20">
                  <CardHeader className="pb-2">
                    <CardTitle className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                      {s.name}
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="flex items-baseline gap-2">
                      <span
                        className="text-2xl font-semibold tabular-nums"
                        style={{color: CHART_COLORS[index % CHART_COLORS.length]}}
                      >
                        {latest(s)}
                      </span>
                      {delta > 0 ? <span className="text-xs text-success">+{delta}</span> : null}
                    </div>
                  </CardContent>
                </Card>
              </Link>
            );
          })}
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">{i18next.t("general:Dashboard")}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-[420px] w-full">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={chartData} margin={{top: 8, right: 16, bottom: 8, left: 0}}>
                <CartesianGrid strokeDasharray="3 3" className="stroke-border" />
                <XAxis dataKey="date" tick={{fontSize: 11}} className="fill-muted-foreground" />
                <YAxis tick={{fontSize: 11}} className="fill-muted-foreground" width={44} />
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
    </div>
  );
}
