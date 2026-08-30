import * as React from "react";
import i18next from "i18next";
import {Card, CardContent, CardHeader, CardTitle} from "@/components/ui/card";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table";
import {Loading} from "@/components/common/Loading";
import {PageHeader} from "@/components/crud/PageHeader";
import * as SystemBackend from "@/backend/SystemInfo";
import * as Setting from "@/lib/setting";

const REFRESH_MS = 5000;

function usageColor(percent: number) {
  if (percent >= 90) {
    return "bg-destructive";
  }
  if (percent >= 70) {
    return "bg-warning";
  }
  return "bg-primary";
}

function Usage({label, value, detail}: {label: string; value: number; detail?: string}) {
  const percent = Number.isFinite(value) ? Math.min(100, Math.max(0, value)) : 0;
  return (
    <div className="space-y-1.5">
      <div className="flex items-baseline justify-between text-sm">
        <span className="text-muted-foreground">{label}</span>
        <span className="tabular-nums">{percent.toFixed(1)}%</span>
      </div>
      <div className="relative h-2 w-full overflow-hidden rounded-full bg-secondary">
        <div className={`h-full transition-all ${usageColor(percent)}`} style={{width: `${percent}%`}} />
      </div>
      {detail ? <p className="text-xs text-muted-foreground">{detail}</p> : null}
    </div>
  );
}

/** The per-API rows of `/api/get-prometheus-info`, port of table/PrometheusInfoTable.js. */
function PrometheusTable({
  rows,
  columns,
}: {
  rows: any[] | null | undefined;
  columns: {key: string; title: string}[];
}) {
  if (rows === null || rows === undefined || rows.length === 0) {
    return <Loading />;
  }

  return (
    <div className="max-h-[300px] overflow-auto rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow className="hover:bg-transparent">
            {columns.map((column) => (
              <TableHead key={column.key}>{column.title}</TableHead>
            ))}
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((row, index) => (
            <TableRow key={index}>
              {columns.map((column) => (
                <TableCell key={column.key} className="tabular-nums">
                  {row[column.key]}
                </TableCell>
              ))}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

export default function SystemInfoPage() {
  const [systemInfo, setSystemInfo] = React.useState<any>(null);
  const [versionInfo, setVersionInfo] = React.useState<any>(null);
  const [prometheusInfo, setPrometheusInfo] = React.useState<any>(null);
  const [loading, setLoading] = React.useState(true);

  React.useEffect(() => {
    let stopped = false;

    const load = () => {
      SystemBackend.getSystemInfo()
        .then((res: any) => {
          if (stopped) {
            return;
          }
          if (res.status === "ok") {
            setSystemInfo(res.data);
            SystemBackend.getPrometheusInfo().then((prometheusRes: any) => {
              if (!stopped && prometheusRes.status === "ok") {
                setPrometheusInfo(prometheusRes.data);
              }
            });
          } else {
            Setting.showMessage("error", res.msg);
            stopped = true;
          }
        })
        .finally(() => setLoading(false));
    };

    load();
    SystemBackend.getVersionInfo().then((res: any) => {
      if (res.status === "ok") {
        setVersionInfo(res.data);
      }
    });

    const id = window.setInterval(() => {
      if (!stopped) {
        load();
      }
    }, REFRESH_MS);
    return () => {
      stopped = true;
      window.clearInterval(id);
    };
  }, []);

  if (loading || systemInfo === null) {
    return <Loading />;
  }

  const cpuUsage: number[] = systemInfo.cpuUsage ?? [];
  const memoryPercent = systemInfo.memoryTotal ? (systemInfo.memoryUsed / systemInfo.memoryTotal) * 100 : 0;
  const diskPercent = systemInfo.diskTotal ? (systemInfo.diskUsed / systemInfo.diskTotal) * 100 : 0;

  return (
    <div className="space-y-4">
      <PageHeader
        title={i18next.t("general:System Info")}
        description={
          versionInfo?.version
            ? `${versionInfo.version}${versionInfo.commitOffset > 0 ? ` (ahead+${versionInfo.commitOffset})` : ""}`
            : undefined
        }
      />

      <div className="grid gap-4 md:grid-cols-3">
        <Card id="cpu-card">
          <CardHeader className="pb-3">
            <CardTitle className="text-base">{i18next.t("system:CPU Usage")}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {cpuUsage.length === 0 ? (
              <p className="text-sm text-muted-foreground">{i18next.t("general:Failed to get")}</p>
            ) : (
              cpuUsage.map((usage, index) => <Usage key={index} label={`CPU ${index + 1}`} value={usage} />)
            )}
          </CardContent>
        </Card>

        <Card id="memory-card">
          <CardHeader className="pb-3">
            <CardTitle className="text-base">{i18next.t("system:Memory Usage")}</CardTitle>
          </CardHeader>
          <CardContent>
            <Usage
              label={i18next.t("system:Memory Usage")}
              value={memoryPercent}
              detail={`${Setting.getFriendlyFileSize(systemInfo.memoryUsed)} / ${Setting.getFriendlyFileSize(
                systemInfo.memoryTotal,
              )}`}
            />
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-base">{i18next.t("system:Disk Usage")}</CardTitle>
          </CardHeader>
          <CardContent>
            <Usage
              label={i18next.t("system:Disk Usage")}
              value={diskPercent}
              detail={`${Setting.getFriendlyFileSize(systemInfo.diskUsed)} / ${Setting.getFriendlyFileSize(
                systemInfo.diskTotal,
              )}`}
            />
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <Card id="latency-card">
          <CardHeader className="pb-3">
            <CardTitle className="text-base">{i18next.t("system:API Latency")}</CardTitle>
          </CardHeader>
          <CardContent>
            <PrometheusTable
              rows={prometheusInfo?.apiLatency}
              columns={[
                {key: "name", title: i18next.t("general:Name")},
                {key: "method", title: i18next.t("general:Method")},
                {key: "count", title: i18next.t("system:Count")},
                {key: "latency", title: `${i18next.t("system:Latency")}(ms)`},
              ]}
            />
          </CardContent>
        </Card>

        <Card id="throughput-card">
          <CardHeader className="pb-3">
            <CardTitle className="text-base">{i18next.t("system:API Throughput")}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <div className="text-sm text-muted-foreground">
              {i18next.t("system:Total Throughput")}: {prometheusInfo?.totalThroughput ?? "-"}
            </div>
            <PrometheusTable
              rows={prometheusInfo?.apiThroughput}
              columns={[
                {key: "name", title: i18next.t("general:Name")},
                {key: "method", title: i18next.t("general:Method")},
                {key: "throughput", title: i18next.t("system:Throughput")},
              ]}
            />
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-base">{i18next.t("system:Network Usage")}</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-4 sm:grid-cols-3">
          <div>
            <div className="text-sm text-muted-foreground">{i18next.t("system:Sent")}</div>
            <div className="text-xl font-semibold tabular-nums">
              {Setting.getFriendlyFileSize(systemInfo.networkSent)}
            </div>
          </div>
          <div>
            <div className="text-sm text-muted-foreground">{i18next.t("system:Received")}</div>
            <div className="text-xl font-semibold tabular-nums">
              {Setting.getFriendlyFileSize(systemInfo.networkRecv)}
            </div>
          </div>
          <div>
            <div className="text-sm text-muted-foreground">{i18next.t("system:Total Throughput")}</div>
            <div className="text-xl font-semibold tabular-nums">
              {Setting.getFriendlyFileSize(systemInfo.networkTotal)}
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
