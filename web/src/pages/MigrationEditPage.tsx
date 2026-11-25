import * as React from "react";
import i18next from "i18next";
import {useParams} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table";
import {ConfirmButton} from "@/components/common/ConfirmButton";
import {ActionSelect} from "@/components/common/ActionSelect";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useOrganizationOptions} from "@/hooks/use-options";
import * as MigrationBackend from "@/backend/MigrationBackend";
import * as Setting from "@/lib/setting";

/** the JSON body of a request is limited to 64 MB, and base64 makes the files a third larger */
const MAX_FILES_SIZE = 45 * 1024 * 1024;

type ImportFile = {name: string; size: number; data: string};

function readFile(file: File): Promise<ImportFile> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => {
      const url = String(reader.result ?? "");
      resolve({name: file.name, size: file.size, data: url.slice(url.indexOf(",") + 1)});
    };
    reader.onerror = () => reject(reader.error);
    reader.readAsDataURL(file);
  });
}

export default function MigrationEditPage() {
  const {organizationName = "", migrationName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();
  const [files, setFiles] = React.useState<ImportFile[]>([]);

  const fields: EditField[] = [
    {
      type: "select",
      name: "organization",
      labelKey: "general:Organization",
      options: () => organizations,
      disabled: () => !Setting.isAdminUser(account),
      // the actions are installed per organization
      onChange: (value, ctx, updateFields) =>
        updateFields(value === ctx.record.organization ? {organization: value} : {organization: value, plugin: ""}),
    },
    {type: "text", name: "name", labelKey: "general:Name", required: true},
    {type: "text", name: "displayName", labelKey: "general:Display name"},
    {
      type: "custom",
      name: "plugin",
      labelKey: "general:Action",
      render: (ctx, update) => (
        <div className="space-y-1">
          <ActionSelect owner={ctx.record.organization} kind="import" value={ctx.record.plugin} onChange={(v) => update("plugin", v)} />
          <p className="text-xs text-muted-foreground">{i18next.t("migration:Import action - Tooltip")}</p>
        </div>
      ),
    },
    {type: "text", name: "host", labelKey: "general:Host"},
    {type: "text", name: "clientId", labelKey: "provider:Client ID"},
    {type: "password", name: "clientSecret", labelKey: "provider:Client secret"},
    {type: "textarea", name: "metadata", labelKey: "provider:Metadata", rows: 4, placeholder: "{}"},
    {
      type: "custom",
      name: "files",
      labelKey: "migration:Files",
      render: () => (
        <div className="space-y-2">
          <input
            type="file"
            multiple
            className="block text-sm"
            onChange={async(e) => {
              const picked = Array.from(e.target.files ?? []);
              const total = picked.reduce((sum, file) => sum + file.size, 0);
              if (total > MAX_FILES_SIZE) {
                Setting.showMessage("error", i18next.t("migration:The files are larger than 45 MB, split them into several migrations"));
                e.target.value = "";
                return;
              }
              setFiles(await Promise.all(picked.map(readFile)));
            }}
          />
          <p className="text-xs text-muted-foreground">{i18next.t("migration:Files - Tooltip")}</p>
          {files.length > 0 ? (
            <ul className="text-sm text-muted-foreground">
              {files.map((file) => <li key={file.name}>{file.name} ({Math.ceil(file.size / 1024)} KB)</li>)}
            </ul>
          ) : null}
        </div>
      ),
    },
    {
      type: "custom",
      name: "report",
      labelKey: "migration:Report",
      block: true,
      render: (ctx) => <MigrationReport migration={ctx.record} />,
    },
  ];

  const run = async(record: any, mode: string, reload: () => void) => {
    if (mode !== "undo" && !record.plugin) {
      Setting.showMessage("error", i18next.t("migration:Choose the action that reads the other system"));
      return;
    }
    // the run reads the saved settings, so the form is saved first
    if (mode !== "undo") {
      const saved: any = await MigrationBackend.updateMigration("admin", migrationName, record);
      if (saved.status !== "ok") {
        Setting.showMessage("error", saved.msg);
        return;
      }
    }
    const res: any = await MigrationBackend.runMigration(
      "admin", record.name, record.organization, mode,
      mode === "undo" ? [] : files.map((file) => ({name: file.name, data: file.data})),
    );
    if (res.status === "ok") {
      Setting.showMessage("success", i18next.t("migration:Started, the report below shows the progress"));
    } else {
      Setting.showMessage("error", res.msg);
    }
    reload();
  };

  return (
    <SimpleEditPage
      titleKey="migration:Edit Migration"
      backTo="/migrations"
      deps={[organizationName, migrationName]}
      fields={fields}
      fetch={() => MigrationBackend.getMigration("admin", migrationName, organizationName)}
      add={(record) => MigrationBackend.addMigration(record)}
      update={(record) => MigrationBackend.updateMigration("admin", migrationName, record)}
      editUrl={(record) => `/migrations/${record.organization}/${record.name}`}
      extraActions={(ctx) => {
        const isRunning = ctx.record.state === "Running";
        const disabled = isRunning || ctx.mode === "add";
        return (
          <>
            <RunningPoller isRunning={isRunning} reload={ctx.reload} />
            <Button variant="outline" disabled={disabled} onClick={() => run(ctx.record, "test", ctx.reload)}>
              {i18next.t("migration:Test connection")}
            </Button>
            <Button variant="outline" disabled={disabled} onClick={() => run(ctx.record, "plan", ctx.reload)}>
              {i18next.t("general:Preview")}
            </Button>
            <ConfirmButton
              variant="outline"
              disabled={disabled}
              destructive={false}
              title={i18next.t("migration:Apply the migration") + "?"}
              description={i18next.t("migration:Apply - Tooltip")}
              onConfirm={() => run(ctx.record, "apply", ctx.reload)}
            >
              {i18next.t("migration:Apply")}
            </ConfirmButton>
            <ConfirmButton
              variant="outline"
              disabled={disabled}
              title={i18next.t("migration:Undo the migration") + "?"}
              description={i18next.t("migration:Undo - Tooltip")}
              onConfirm={() => run(ctx.record, "undo", ctx.reload)}
            >
              {i18next.t("migration:Undo")}
            </ConfirmButton>
          </>
        );
      }}
    />
  );
}

/** reloads the migration while it runs in the background, so the report fills in */
function RunningPoller({isRunning, reload}: {isRunning: boolean; reload: () => void}) {
  React.useEffect(() => {
    if (!isRunning) {
      return;
    }
    const timer = window.setInterval(reload, 2000);
    return () => window.clearInterval(timer);
  }, [isRunning, reload]);
  return null;
}

function MigrationReport({migration}: {migration: any}) {
  const report = migration.report;
  if (!report) {
    return <p className="text-sm text-muted-foreground">{i18next.t("migration:Report - Tooltip")}</p>;
  }

  const modeLabels: Record<string, string> = {
    test: i18next.t("migration:Test connection"),
    plan: i18next.t("general:Preview"),
    apply: i18next.t("migration:Apply"),
    undo: i18next.t("migration:Undo"),
  };
  const isUndo = report.mode === "undo";

  return (
    <div className="space-y-3 text-sm">
      <div className="text-muted-foreground">
        {modeLabels[report.mode] ?? report.mode} | {migration.state} | {report.startedTime}{report.finishedTime ? ` - ${report.finishedTime}` : ""}
      </div>
      {report.error ? <div className="text-destructive">{report.error}</div> : null}
      {(report.kinds ?? []).length > 0 ? (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{i18next.t("migration:Kind")}</TableHead>
              <TableHead>{i18next.t("migration:Total")}</TableHead>
              {isUndo ? null : <TableHead>{i18next.t("consent:New")}</TableHead>}
              {isUndo ? null : <TableHead>{i18next.t("migration:Exists")}</TableHead>}
              {isUndo ? null : <TableHead>{i18next.t("migration:Invalid")}</TableHead>}
              {report.mode === "apply" ? <TableHead>{i18next.t("migration:Added")}</TableHead> : null}
              {isUndo ? <TableHead>{i18next.t("application:Deleted")}</TableHead> : null}
              {report.mode === "apply" || isUndo ? <TableHead>{i18next.t("webhook:Failed")}</TableHead> : null}
            </TableRow>
          </TableHeader>
          <TableBody>
            {report.kinds.map((kind: any) => (
              <TableRow key={kind.kind}>
                <TableCell>{i18next.t(`migration:${kind.kind}`)}</TableCell>
                <TableCell>{kind.total}</TableCell>
                {isUndo ? null : <TableCell>{kind.new}</TableCell>}
                {isUndo ? null : <TableCell>{kind.exists}</TableCell>}
                {isUndo ? null : <TableCell>{kind.invalid}</TableCell>}
                {report.mode === "apply" ? <TableCell>{kind.added}</TableCell> : null}
                {isUndo ? <TableCell>{kind.deleted}</TableCell> : null}
                {report.mode === "apply" || isUndo ? <TableCell>{kind.failed}</TableCell> : null}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      ) : null}
      {(report.warnings ?? []).length > 0 ? (
        <div>
          <div className="font-medium">{i18next.t("migration:Warnings")}</div>
          <ul className="list-disc pl-5 text-muted-foreground">
            {report.warnings.map((warning: string) => <li key={warning}>{warning}</li>)}
          </ul>
        </div>
      ) : null}
      {(report.issues ?? []).length > 0 ? (
        <div>
          <div className="font-medium">{i18next.t("migration:Issues")}</div>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{i18next.t("migration:Kind")}</TableHead>
                <TableHead>{i18next.t("general:Name")}</TableHead>
                <TableHead>{i18next.t("rule:Reason")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {report.issues.map((issue: any, index: number) => (
                <TableRow key={index}>
                  <TableCell>{i18next.t(`migration:${issue.kind}`)}</TableCell>
                  <TableCell>{issue.name}</TableCell>
                  <TableCell>{issue.reason}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      ) : null}
    </div>
  );
}
