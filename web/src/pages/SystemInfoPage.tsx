import * as React from "react";
import i18next from "i18next";
import {Card, CardContent, CardHeader, CardTitle} from "@/components/ui/card";
import {Alert, AlertDescription, AlertTitle} from "@/components/ui/alert";
import {Badge} from "@/components/ui/badge";
import {Progress} from "@/components/ui/progress";
import {ConfirmButton} from "@/components/common/ConfirmButton";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table";
import {Loading} from "@/components/common/Loading";
import {PageHeader} from "@/components/crud/PageHeader";
import * as SystemBackend from "@/backend/SystemInfo";
import * as Setting from "@/lib/setting";

const REFRESH_MS = 5000;
const UPGRADE_POLL_MS = 1500;
const UPGRADE_TIMEOUT_MS = 10 * 60 * 1000;

function isUpgrading(state?: string) {
  return state === "downloading" || state === "installing" || state === "restarting";
}

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

function CommandBlock({command}: {command: string}) {
  return <pre className="overflow-x-auto rounded-md bg-muted p-3 font-mono text-xs leading-relaxed">{command}</pre>;
}

/** Current and latest version, and the one-click upgrade of an official release binary. */
function UpgradeCard() {
  const [info, setInfo] = React.useState<any>(null);
  const [target, setTarget] = React.useState("");
  const [reconnecting, setReconnecting] = React.useState(false);
  const [lostMessage, setLostMessage] = React.useState("");

  React.useEffect(() => {
    SystemBackend.getUpgradeInfo()
      .then((res: any) => {
        if (res.status === "ok") {
          setInfo(res.data);
          if (isUpgrading(res.data.status?.state)) {
            setTarget(res.data.status.version);
          }
        }
      })
      .catch(() => {});
  }, []);

  React.useEffect(() => {
    if (target === "") {
      return;
    }

    let stopped = false;
    let lost = false;
    const startedAt = Date.now();
    const id = window.setInterval(() => {
      if (Date.now() - startedAt > UPGRADE_TIMEOUT_MS) {
        setLostMessage(i18next.t("system:Casdoor did not come back with the new version, please check the server logs"));
        setTarget("");
        return;
      }

      // once the server has gone away for the restart, ask the API that every version has
      const request = lost ? SystemBackend.getVersionInfo() : SystemBackend.getUpgradeInfo();
      request
        .then((res: any) => {
          if (stopped || res.status !== "ok") {
            return;
          }
          if (lost) {
            if (res.data.version === target) {
              stopped = true;
              window.location.reload();
            } else {
              setLostMessage(i18next.t("system:Casdoor did not come back with the new version, please check the server logs"));
              setTarget("");
            }
            return;
          }
          setInfo(res.data);
          if (res.data.status?.state === "failed") {
            setTarget("");
          }
        })
        .catch(() => {
          lost = true;
          if (!stopped) {
            setReconnecting(true);
          }
        });
    }, UPGRADE_POLL_MS);
    return () => {
      stopped = true;
      window.clearInterval(id);
    };
  }, [target]);

  if (info === null) {
    return null;
  }

  const status = info.status ?? {};
  const latestVersion: string = info.latestVersion ?? "";
  const inProgress = target !== "" || isUpgrading(status.state);
  const currentText = info.currentVersion
    ? `${info.currentVersion}${info.commitOffset > 0 ? ` (ahead+${info.commitOffset})` : ""}`
    : i18next.t("system:Unknown version");

  const upgrade = async() => {
    setLostMessage("");
    setReconnecting(false);
    const res: any = await SystemBackend.upgradeSystem(latestVersion);
    if (res.status === "ok") {
      setInfo({...info, status: {state: "downloading", version: latestVersion, downloaded: 0, total: 0, error: ""}});
      setTarget(latestVersion);
    } else {
      Setting.showMessage("error", res.msg);
    }
  };

  const renderProgress = () => {
    if (reconnecting || status.state === "restarting") {
      return <p className="text-muted-foreground">{i18next.t("system:Restarting, this page reloads when Casdoor is back")}</p>;
    }
    if (status.state === "installing") {
      return <p className="text-muted-foreground">{i18next.t("integration:Installing")}...</p>;
    }
    const percent = status.total > 0 ? (status.downloaded / status.total) * 100 : 0;
    return (
      <div className="space-y-1.5">
        <div className="flex items-baseline justify-between">
          <span className="text-muted-foreground">{i18next.t("system:Downloading")} {status.version}...</span>
          <span className="tabular-nums">
            {Setting.getFriendlyFileSize(status.downloaded ?? 0)}
            {status.total > 0 ? ` / ${Setting.getFriendlyFileSize(status.total)}` : ""}
          </span>
        </div>
        <Progress value={percent} />
      </div>
    );
  };

  const renderUpgrade = () => {
    if (!info.hasNewVersion) {
      return null;
    }

    if (info.deploymentType === "docker") {
      return (
        <div className="space-y-2">
          <p className="text-muted-foreground">
            {i18next.t("system:Casdoor runs in a container, pull the new image and recreate the container")}:
          </p>
          <CommandBlock command={`docker pull casbin/casdoor:${latestVersion.replace(/^v/, "")}`} />
        </div>
      );
    }
    if (info.deploymentType === "source") {
      return (
        <div className="space-y-2">
          <p className="text-muted-foreground">
            {i18next.t("system:Casdoor runs from source code, pull the new code, rebuild it and restart")}:
          </p>
          <CommandBlock command={"git pull\ncd web && yarn install && yarn build && cd ..\ngo build"} />
        </div>
      );
    }
    if (info.deploymentType !== "release") {
      return (
        <p className="text-muted-foreground">
          {i18next.t("system:This Casdoor was not installed from an official release package, download the new version from GitHub Releases")}
        </p>
      );
    }
    if (info.isUpgradeDisabled) {
      return (
        <p className="text-muted-foreground">
          {i18next.t("system:Upgrading from this page is disabled by disableUpgrade in app.conf")}
        </p>
      );
    }
    if (inProgress) {
      return renderProgress();
    }
    if (!info.canUpgrade) {
      return null;
    }
    return (
      <ConfirmButton
        size="sm"
        destructive={false}
        title={`${i18next.t("system:Upgrade to")} ${latestVersion}?`}
        description={i18next.t("system:Casdoor will download the release from GitHub, verify its SHA-256 checksum, replace its binary and frontend and restart. The current version is kept in the backup folder next to the binary.")}
        onConfirm={upgrade}
      >
        {i18next.t("system:Upgrade to")} {latestVersion}
      </ConfirmButton>
    );
  };

  return (
    <Card id="upgrade-card">
      <CardHeader className="flex-row items-center justify-between gap-3 space-y-0 pb-3">
        <CardTitle className="text-base">{i18next.t("system:Version")}</CardTitle>
        {info.hasNewVersion ? (
          <Badge variant="warning">{i18next.t("system:New version available")}</Badge>
        ) : latestVersion ? (
          <Badge variant="success">{i18next.t("system:Up to date")}</Badge>
        ) : null}
      </CardHeader>
      <CardContent className="space-y-3 text-sm">
        <div className="grid gap-2 sm:grid-cols-2">
          <div>
            <span className="text-muted-foreground">{i18next.t("system:Current version")}: </span>
            <span className="tabular-nums">{currentText}</span>
          </div>
          <div>
            <span className="text-muted-foreground">{i18next.t("system:Latest version")}: </span>
            {latestVersion ? (
              <a href={info.latestUrl} target="_blank" rel="noreferrer" className="tabular-nums underline-offset-4 hover:underline">
                {latestVersion}
              </a>
            ) : (
              "-"
            )}
          </div>
        </div>
        {info.isCheckDisabled ? (
          <p className="text-muted-foreground">
            {i18next.t("system:Checking for new versions is disabled by disableVersionCheck in app.conf")}
          </p>
        ) : null}
        {info.checkError ? (
          <p className="text-muted-foreground">
            {i18next.t("system:Failed to check the latest version")}: {info.checkError}
          </p>
        ) : null}
        {status.state === "failed" && target === "" ? (
          <Alert variant="destructive">
            <AlertTitle>{i18next.t("system:Upgrade failed")}</AlertTitle>
            <AlertDescription>{status.error}</AlertDescription>
          </Alert>
        ) : null}
        {lostMessage ? (
          <Alert variant="destructive">
            <AlertDescription>{lostMessage}</AlertDescription>
          </Alert>
        ) : null}
        {renderUpgrade()}
      </CardContent>
    </Card>
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
  const versionText = versionInfo?.version
    ? `${versionInfo.version}${versionInfo.commitOffset > 0 ? ` (ahead+${versionInfo.commitOffset})` : ""}`
    : i18next.t("system:Unknown version");

  return (
    <div className="space-y-4">
      <PageHeader
        title={i18next.t("general:System Info")}
        description={versionInfo?.version ? versionText : undefined}
      />

      <UpgradeCard />

      {/* Memory and disk are one meter each; the CPU card is as many meters as the
          machine has cores, so it gets its own full-width row rather than being
          squeezed into a third of one and running thirty rows deep. */}
      <div className="grid gap-4 md:grid-cols-2">
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

      <Card id="cpu-card">
        <CardHeader className="flex-row items-baseline justify-between gap-3 space-y-0 pb-3">
          <CardTitle className="text-base">{i18next.t("system:CPU Usage")}</CardTitle>
          {cpuUsage.length > 0 ? (
            <span className="text-sm text-muted-foreground tabular-nums">
              {cpuUsage.length} × {(cpuUsage.reduce((sum, usage) => sum + usage, 0) / cpuUsage.length).toFixed(1)}%
            </span>
          ) : null}
        </CardHeader>
        <CardContent>
          {cpuUsage.length === 0 ? (
            <p className="text-sm text-muted-foreground">{i18next.t("general:Failed to get")}</p>
          ) : (
            <div className="grid gap-x-8 gap-y-3 sm:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-4">
              {cpuUsage.map((usage, index) => (
                <Usage key={index} label={`CPU ${index + 1}`} value={usage} />
              ))}
            </div>
          )}
        </CardContent>
      </Card>

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

      <Card id="about-card">
        <CardHeader className="pb-3">
          <CardTitle className="text-base">{i18next.t("system:About Casdoor")}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-1 text-sm">
          <p className="text-muted-foreground">
            {i18next.t("system:An Identity and Access Management (IAM) / Single-Sign-On (SSO) platform with web UI supporting OAuth 2.0, OIDC, SAML and CAS")}
          </p>
          <div>
            GitHub:{" "}
            <a href="https://github.com/casdoor/casdoor" target="_blank" rel="noreferrer" className="underline-offset-4 hover:underline">
              Casdoor
            </a>
          </div>
          <div>
            {i18next.t("system:Version")}:{" "}
            {versionInfo?.version ? (
              <a
                href={`https://github.com/casdoor/casdoor/releases/tag/${versionInfo.version}`}
                target="_blank"
                rel="noreferrer"
                className="underline-offset-4 hover:underline"
              >
                {versionText}
              </a>
            ) : (
              versionText
            )}
          </div>
          <div>
            {i18next.t("system:Official website")}:{" "}
            <a href="https://casdoor.org" target="_blank" rel="noreferrer" className="underline-offset-4 hover:underline">
              https://casdoor.org
            </a>
          </div>
          <div>
            {i18next.t("system:Community")}:{" "}
            <a
              href="https://casdoor.org/#:~:text=Casdoor%20API-,Community,-GitHub"
              target="_blank"
              rel="noreferrer"
              className="underline-offset-4 hover:underline"
            >
              Get in Touch!
            </a>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
