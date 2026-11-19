import * as React from "react";
import i18next from "i18next";
import {Copy, RefreshCw, ShieldCheck} from "lucide-react";
import {Alert, AlertDescription, AlertTitle} from "@/components/ui/alert";
import {Badge} from "@/components/ui/badge";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Switch} from "@/components/ui/switch";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table";
import {MultiSelect} from "@/components/common/MultiSelect";
import {FormRow} from "@/components/crud/FormRow";
import {EmailMfaType, RadiusMfaType, RecoveryMfaType, SmsMfaType, TotpMfaType} from "@/components/auth/mfa/constants";
import {getStepUpMethodName, StepUpMfaDialog, type StepUpResult} from "@/components/auth/mfa/StepUpMfaDialog";
import type {ApplicationTabProps} from "@/components/application/types";
import * as ApplicationBackend from "@/backend/ApplicationBackend";
import * as Setting from "@/lib/setting";
import {cn} from "@/lib/utils";

const stepUpMfaTypes = [TotpMfaType, EmailMfaType, SmsMfaType, RadiusMfaType, RecoveryMfaType];

function decodeJwtPart(part: string) {
  try {
    const base64 = part.replace(/-/g, "+").replace(/_/g, "/");
    const json = decodeURIComponent(atob(base64).split("").map((c) => `%${`00${c.charCodeAt(0).toString(16)}`.slice(-2)}`).join(""));
    return JSON.parse(json);
  } catch {
    return null;
  }
}

function CodeBlock({text, wrap}: {text: string; wrap?: boolean}) {
  return (
    <div className="relative">
      <pre className={cn("overflow-x-auto rounded-md border bg-muted/50 p-3 pr-10 font-mono text-xs leading-relaxed", wrap && "whitespace-pre-wrap break-all")}>{text}</pre>
      <Button variant="ghost" size="icon" className="absolute right-1 top-1 size-7" onClick={() => Setting.copyToClipboard(text)}>
        <Copy className="size-3.5" />
      </Button>
    </div>
  );
}

function TryStepUp({application, saved}: {application: any; saved: boolean}) {
  const [open, setOpen] = React.useState(false);
  const [action, setAction] = React.useState("approve-payment-1024");
  const [result, setResult] = React.useState<StepUpResult | null>(null);
  const claims = result ? decodeJwtPart(result.token.split(".")[1] ?? "") : null;

  return (
    <div className="space-y-3">
      <div className="flex flex-wrap items-end gap-2">
        <div className="min-w-[14rem] flex-1 space-y-1">
          <div className="text-xs text-muted-foreground">{i18next.t("application:Action to approve (optional)")}</div>
          <Input value={action} onChange={(e) => setAction(e.target.value)} />
        </div>
        <Button disabled={!saved} onClick={() => setOpen(true)}>
          <ShieldCheck />
          {i18next.t("application:Try step-up verification")}
        </Button>
      </div>
      <p className="text-xs text-muted-foreground">{i18next.t("application:Save the application first, the check uses the saved settings and your own account")}</p>
      {result ? (
        <div className="space-y-2">
          <div className="flex flex-wrap items-center gap-2 text-sm">
            <Badge variant="success">{i18next.t("user:Verified")}</Badge>
            <span className="text-muted-foreground">
              {getStepUpMethodName(result.mfaType)}, {i18next.t("application:the proof token expires in {seconds} seconds and can be redeemed once").replace("{seconds}", String(result.expiresIn))}
            </span>
          </div>
          <CodeBlock text={result.token} wrap />
          {claims ? <CodeBlock text={JSON.stringify(claims, null, 2)} /> : null}
        </div>
      ) : null}
      <StepUpMfaDialog
        open={open}
        onOpenChange={setOpen}
        clientId={application.clientId}
        action={action}
        onVerified={(r) => {
          setResult(r);
          Setting.showMessage("success", i18next.t("user:Verified"));
        }}
      />
    </div>
  );
}

function RecentStepUps({application}: {application: any}) {
  const id = `${application.owner}/${application.name}`;
  const [records, setRecords] = React.useState<any[]>([]);

  const load = React.useCallback(() => {
    ApplicationBackend.getStepUpRecords(id).then((res: any) => {
      if (res.status === "ok") {
        setRecords(res.data ?? []);
      }
    });
  }, [id]);

  React.useEffect(() => {
    load();
  }, [load]);

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-end">
        <Button variant="ghost" size="sm" onClick={load}>
          <RefreshCw />
          {i18next.t("general:Refresh")}
        </Button>
      </div>
      {records.length === 0 ? (
        <p className="text-sm text-muted-foreground">{i18next.t("application:No step-up checks have passed yet, failed attempts are in the records page")}</p>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{i18next.t("general:Created time")}</TableHead>
              <TableHead>{i18next.t("general:User")}</TableHead>
              <TableHead>{i18next.t("general:Method")}</TableHead>
              <TableHead>{i18next.t("general:Action")}</TableHead>
              <TableHead>{i18next.t("application:Token")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {records.map((r) => (
              <TableRow key={`${r.owner}/${r.name}`}>
                <TableCell className="whitespace-nowrap">{Setting.getFormattedDate(r.createdTime)}</TableCell>
                <TableCell>{`${r.owner}/${r.user}`}</TableCell>
                <TableCell>{getStepUpMethodName(r.mfaType)}</TableCell>
                <TableCell className="max-w-[16rem] truncate">{r.action}</TableCell>
                <TableCell>
                  {r.redeemedTime ? (
                    <Badge variant="secondary">{i18next.t("application:Redeemed")}</Badge>
                  ) : r.expireTime * 1000 < Date.now() ? (
                    <Badge variant="outline">{i18next.t("subscription:Expired")}</Badge>
                  ) : (
                    <Badge variant="info">{i18next.t("application:Valid")}</Badge>
                  )}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
    </div>
  );
}

/**
 * The "Step-up MFA" tab: lets the application ask a signed-in user for a fresh code of the MFA factor
 * they already enrolled before a sensitive action, without a new sign-in.
 */
export function ApplicationStepUpMfaTab({application, updateField, mode}: ApplicationTabProps & {mode: string}) {
  const serverUrl = Setting.getFullServerUrl();
  const saved = mode !== "add";
  const typeOptions = stepUpMfaTypes.map((t) => ({value: t, label: getStepUpMethodName(t)}));

  const frontendExample = `# In the browser, with the user's access token (or the Casdoor session cookie and clientId)
curl -X POST ${serverUrl}/api/mfa/step-up/verify \\
  -H "Authorization: Bearer <access_token>" \\
  -F mfaType=app -F passcode=123456 -F action=approve-payment-1024
# => {"status":"ok","data":{"token":"<proof token>","expiresIn":${application.stepUpExpireInSeconds || 300},...}}`;

  const backendExample = `# On your server, with the application's own credentials: check and use up the proof token
curl -X POST ${serverUrl}/api/mfa/step-up/redeem \\
  -u ${application.clientId}:<client_secret> \\
  -F token=<proof token> -F action=approve-payment-1024
# => {"status":"ok","data":{"owner":"...","name":"...","mfa_type":"app","amr":["mfa","otp"],"auth_time":...}}

# Or verify a code a user typed into your own page, for a user of the organization
curl -X POST ${serverUrl}/api/mfa/step-up/verify \\
  -u ${application.clientId}:<client_secret> \\
  -F owner=${application.organization} -F name=<user> -F mfaType=app -F passcode=123456`;

  return (
    <>
      <div className="col-span-full">
        <Alert>
          <ShieldCheck className="size-4" />
          <AlertTitle>{i18next.t("application:Step-up MFA")}</AlertTitle>
          <AlertDescription>
            {i18next.t("application:Step-up MFA - Description")}
          </AlertDescription>
        </Alert>
      </div>
      <FormRow labelKey="application:Enable step-up MFA">
        <Switch checked={!!application.enableStepUpMfa} onCheckedChange={(v) => updateField("enableStepUpMfa", v)} />
      </FormRow>
      {application.enableStepUpMfa ? (
        <>
          <FormRow labelKey="application:Allowed methods">
            <MultiSelect
              value={application.stepUpMfaTypes ?? []}
              onChange={(v) => updateField("stepUpMfaTypes", v)}
              options={typeOptions}
              placeholder={i18next.t("application:All methods the user has enabled")}
            />
          </FormRow>
          <FormRow labelKey="application:Proof token lifetime (seconds)">
            <Input
              type="number"
              min={30}
              max={3600}
              placeholder="300"
              value={application.stepUpExpireInSeconds || ""}
              onChange={(e) => updateField("stepUpExpireInSeconds", Setting.myParseInt(e.target.value))}
            />
          </FormRow>
          <FormRow labelKey="application:Try it" block>
            <TryStepUp application={application} saved={saved} />
          </FormRow>
          <FormRow labelKey="application:How to call it" block>
            <div className="space-y-2">
              <CodeBlock text={frontendExample} />
              <CodeBlock text={backendExample} />
              <p className="text-xs text-muted-foreground">{i18next.t("application:The proof token is a JWT signed with the application's token cert, it can also be checked with the JWKS, then your server must remember its jti so it is used only once")}</p>
            </div>
          </FormRow>
          {mode !== "add" ? (
            <FormRow labelKey="application:Recent step-up checks" block>
              <RecentStepUps application={application} />
            </FormRow>
          ) : null}
        </>
      ) : null}
    </>
  );
}
