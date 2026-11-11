import * as React from "react";
import i18next from "i18next";
import {Badge} from "@/components/ui/badge";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Switch} from "@/components/ui/switch";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table";
import {EditionNotice} from "@/components/common/EditionNotice";
import {MultiSelect} from "@/components/common/MultiSelect";
import {SelectField} from "@/components/common/SelectField";
import {FormRow} from "@/components/crud/FormRow";
import {useGroupOptions} from "@/hooks/use-options";
import * as ApplicationBackend from "@/backend/ApplicationBackend";
import * as SyncerBackend from "@/backend/SyncerBackend";
import * as Setting from "@/lib/setting";

const stateVariant = (state: string) => (state === "Active" ? "success" : state === "Disabled" ? "warning" : "secondary");

const stateText = (state: string) => {
  switch (state) {
  case "Active":
    return i18next.t("general:Enabled");
  case "Disabled":
    return i18next.t("general:Disabled");
  case "Deleted":
    return i18next.t("application:Deleted");
  default:
    return state;
  }
};

/** The accounts pushed to the system behind the application, and the queue that keeps them up to date. */
function ProvisioningStatus({application}: {application: any}) {
  const id = `${application.owner}/${application.name}`;
  const [accounts, setAccounts] = React.useState<any[]>([]);
  const [tasks, setTasks] = React.useState<any[]>([]);
  const [summary, setSummary] = React.useState<any>(null);
  const [state, setState] = React.useState("");

  const load = React.useCallback(() => {
    ApplicationBackend.getProvisionedAccounts(id, state).then((res: any) => {
      if (res.status === "ok") {
        setAccounts(res.data ?? []);
      }
    });
    ApplicationBackend.getProvisioningTasks(id).then((res: any) => {
      if (res.status === "ok") {
        setTasks(res.data ?? []);
        setSummary(res.data2);
      }
    });
  }, [id, state]);

  React.useEffect(() => {
    load();
  }, [load]);

  const run = (action: () => Promise<any>, successKey: string) => {
    action().then((res: any) => {
      if (res.status === "ok") {
        Setting.showMessage("success", i18next.t(successKey).replace("%s", String(res.data ?? "")));
        setTimeout(load, 1500);
      } else {
        Setting.showMessage("error", res.msg);
      }
    });
  };

  return (
    <div className="col-span-full space-y-3">
      <div className="flex flex-wrap items-center gap-2">
        <Button variant="outline" size="sm" onClick={() => run(() => ApplicationBackend.syncProvisioning(id), "application:Queued %s users to sync")}>
          {i18next.t("application:Sync all now")}
        </Button>
        <Button variant="outline" size="sm" onClick={() => run(() => ApplicationBackend.retryProvisioning(id), "application:Queued %s failed tasks again")}>
          {i18next.t("application:Retry failed")}
        </Button>
        <Button variant="outline" size="sm" onClick={load}>{i18next.t("general:Refresh")}</Button>
        {summary ? (
          <span className="text-xs text-muted-foreground">
            {`${i18next.t("general:Enabled")} ${summary.accounts.Active ?? 0} | ${i18next.t("general:Disabled")} ${summary.accounts.Disabled ?? 0} | ${i18next.t("application:Deleted")} ${summary.accounts.Deleted ?? 0} | ${i18next.t("application:With errors")} ${summary.accounts.Error ?? 0} | ${i18next.t("application:Queued")} ${(summary.tasks.Pending ?? 0) + (summary.tasks.Retrying ?? 0)} | ${i18next.t("webhook:Failed")} ${summary.tasks.Failed ?? 0}`}
          </span>
        ) : null}
        <div className="ml-auto w-40">
          <SelectField
            value={state || "All"}
            onChange={(value) => setState(value === "All" ? "" : value)}
            options={["All", "Active", "Disabled", "Deleted"].map((value) => ({id: value, name: value === "All" ? i18next.t("general:All") : stateText(value)}))}
          />
        </div>
      </div>
      <div className="overflow-x-auto rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow className="hover:bg-transparent">
              <TableHead className="w-[200px]">{i18next.t("general:User")}</TableHead>
              <TableHead className="w-[220px]">{i18next.t("application:Account ID in the system")}</TableHead>
              <TableHead className="w-[100px]">{i18next.t("general:State")}</TableHead>
              <TableHead className="w-[170px]">{i18next.t("general:Updated time")}</TableHead>
              <TableHead>{i18next.t("webhook:Last Error")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {accounts.length === 0 ? (
              <TableRow className="hover:bg-transparent">
                <TableCell colSpan={5} className="h-16 text-center text-muted-foreground">{i18next.t("general:No data")}</TableCell>
              </TableRow>
            ) : accounts.map((account) => (
              <TableRow key={account.userId}>
                <TableCell className="break-all">{account.user}</TableCell>
                <TableCell className="break-all font-mono text-xs">{account.remoteId}</TableCell>
                <TableCell><Badge variant={stateVariant(account.state) as any}>{stateText(account.state)}</Badge></TableCell>
                <TableCell className="whitespace-nowrap">{account.updatedTime ? new Date(account.updatedTime).toLocaleString() : ""}</TableCell>
                <TableCell className="max-w-[360px] truncate text-xs text-destructive" title={account.lastError}>{account.lastError}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
      <p className="text-sm font-medium">{i18next.t("application:Recent tasks")}</p>
      <div className="overflow-x-auto rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow className="hover:bg-transparent">
              <TableHead className="w-[170px]">{i18next.t("general:Updated time")}</TableHead>
              <TableHead className="w-[200px]">{i18next.t("general:User")}</TableHead>
              <TableHead className="w-[100px]">{i18next.t("general:State")}</TableHead>
              <TableHead className="w-[100px]">{i18next.t("general:Action")}</TableHead>
              <TableHead className="w-[80px]">{i18next.t("application:Attempts")}</TableHead>
              <TableHead>{i18next.t("webhook:Last Error")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {tasks.length === 0 ? (
              <TableRow className="hover:bg-transparent">
                <TableCell colSpan={6} className="h-16 text-center text-muted-foreground">{i18next.t("general:No data")}</TableCell>
              </TableRow>
            ) : tasks.map((task) => (
              <TableRow key={task.id}>
                <TableCell className="whitespace-nowrap">{new Date(task.updatedTime).toLocaleString()}</TableCell>
                <TableCell className="break-all">{task.user}</TableCell>
                <TableCell>{task.state}</TableCell>
                <TableCell>{task.action}</TableCell>
                <TableCell>{task.attempts}</TableCell>
                <TableCell className="max-w-[360px] truncate text-xs text-destructive" title={task.lastError}>{task.lastError}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}

/** Pushes the accounts of the application's users to the system behind it, by SCIM 2.0 or into a legacy user table. */
export function ApplicationProvisioningTab({application, updateField, mode}: {application: any; updateField: (field: string, value: any) => void; mode: string}) {
  const config = application.provisioning ?? {};
  const set = (key: string, value: any) => updateField("provisioning", {...config, [key]: value});
  const groups = useGroupOptions(application.organization);
  const [syncers, setSyncers] = React.useState<any[]>([]);
  const [testing, setTesting] = React.useState(false);
  const type = config.type || "SCIM";

  React.useEffect(() => {
    if (!application.organization) {
      return;
    }
    SyncerBackend.getSyncers("admin", application.organization, 1, 1000).then((res: any) => {
      if (res.status === "ok") {
        setSyncers((res.data ?? []).filter((syncer: any) => syncer.type === "Database"));
      }
    });
  }, [application.organization]);

  const test = () => {
    setTesting(true);
    ApplicationBackend.testProvisioning(`${application.owner}/${application.name}`, {...config, type}).then((res: any) => {
      setTesting(false);
      if (res.status === "ok") {
        Setting.showMessage("success", i18next.t("application:Connected to the system"));
      } else {
        Setting.showMessage("error", res.msg);
      }
    });
  };

  return (
    <>
      <div className="col-span-full">
        <EditionNotice edition={type === "Database" ? "Enterprise" : "Commercial"} />
        <p className="text-xs text-muted-foreground">
          {i18next.t("application:Casdoor creates the accounts of the users in the system behind the application, updates them when the users change, and disables or deletes them when the users leave, are forbidden or are no longer in the chosen groups. Failed pushes are retried, and a full sync reconciles every account")}
        </p>
      </div>
      <FormRow labelKey="application:Enable provisioning">
        <Switch checked={!!config.isEnabled} onCheckedChange={(checked) => updateField("provisioning", {deprovision: "Disable", userNameField: "Name", ...config, type, isEnabled: checked})} />
      </FormRow>
      <FormRow labelKey="general:Type">
        <SelectField
          value={type}
          onChange={(value) => set("type", value)}
          options={[
            {id: "SCIM", name: "SCIM 2.0"},
            {id: "Database", name: i18next.t("application:Database of a legacy system")},
          ]}
        />
      </FormRow>
      {type === "SCIM" ? (
        <>
          <FormRow labelKey="application:SCIM URL">
            <Input placeholder="https://app.example.com/scim/v2" value={config.url ?? ""} onChange={(e) => set("url", e.target.value)} />
          </FormRow>
          <FormRow labelKey="token:Access token">
            <Input type="password" value={config.token ?? ""} onChange={(e) => set("token", e.target.value)} />
          </FormRow>
          <FormRow labelKey="application:User name from">
            <SelectField
              value={config.userNameField || "Name"}
              onChange={(value) => set("userNameField", value)}
              options={[
                {id: "Name", name: i18next.t("general:Name")},
                {id: "Email", name: i18next.t("general:Email")},
                {id: "Phone", name: i18next.t("general:Phone")},
                {id: "Id", name: "ID"},
              ]}
            />
          </FormRow>
        </>
      ) : (
        <FormRow labelKey="application:Database syncer">
          <SelectField
            value={config.syncer ?? ""}
            onChange={(value) => set("syncer", value)}
            options={syncers.map((syncer) => ({id: `${syncer.owner}/${syncer.name}`, name: `${syncer.name} (${syncer.databaseType} ${syncer.table})`}))}
          />
          <p className="mt-1.5 text-xs text-muted-foreground">
            {i18next.t("application:The accounts are written into the table of the database syncer with its column mapping. The syncer may stay disabled. To disable accounts, map a column to IsForbidden")}
          </p>
        </FormRow>
      )}
      <FormRow labelKey="enterprise:Only these groups">
        <MultiSelect value={config.groups ?? []} onChange={(value) => set("groups", value)} options={groups} placeholder={i18next.t("application:Every user of the organization")} />
      </FormRow>
      <FormRow labelKey="application:When a user leaves">
        <SelectField
          value={config.deprovision || "Disable"}
          onChange={(value) => set("deprovision", value)}
          options={[
            {id: "Disable", name: i18next.t("application:Disable the account")},
            {id: "Delete", name: i18next.t("application:Delete the account")},
          ]}
        />
      </FormRow>
      <FormRow labelKey="application:Full sync every (minutes)">
        <Input type="number" min={0} value={config.syncIntervalMinutes ?? 0} onChange={(e) => set("syncIntervalMinutes", Number(e.target.value))} />
      </FormRow>
      <FormRow labelKey="application:Last full sync">
        <Input readOnly value={config.lastSyncTime ? new Date(config.lastSyncTime).toLocaleString() : ""} />
      </FormRow>
      {mode !== "add" ? (
        <>
          <div className="col-span-full">
            <Button variant="outline" size="sm" loading={testing} onClick={test}>{i18next.t("syncer:Test Connection")}</Button>
          </div>
          {config.isEnabled ? <ProvisioningStatus application={application} /> : null}
        </>
      ) : null}
    </>
  );
}
