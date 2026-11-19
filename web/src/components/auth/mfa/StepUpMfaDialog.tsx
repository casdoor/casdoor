import * as React from "react";
import i18next from "i18next";
import {ShieldCheck} from "lucide-react";
import {Button} from "@/components/ui/button";
import {Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle} from "@/components/ui/dialog";
import {Input} from "@/components/ui/input";
import {CodeInput} from "@/components/auth/CodeInput";
import {EmailMfaType, RadiusMfaType, RecoveryMfaType, SmsMfaType, TotpMfaType} from "@/components/auth/mfa/constants";
import * as MfaBackend from "@/backend/MfaBackend";
import {cn} from "@/lib/utils";

export interface StepUpMethod {
  mfaType: string;
  dest?: string;
  countryCode?: string;
  isPreferred?: boolean;
}

export interface StepUpResult {
  token: string;
  tokenType: string;
  expiresIn: number;
  mfaType: string;
}

const methodLabelKeys: Record<string, string> = {
  [TotpMfaType]: "mfa:Use Authenticator App",
  [EmailMfaType]: "mfa:Use Email",
  [SmsMfaType]: "mfa:Use SMS",
  [RadiusMfaType]: "mfa:Use Radius",
  [RecoveryMfaType]: "mfa:Use a recovery code",
};

export function getStepUpMethodLabel(mfaType: string) {
  return i18next.t(methodLabelKeys[mfaType] ?? mfaType);
}

/** a short name of the method for tables and summaries */
export function getStepUpMethodName(mfaType: string) {
  switch (mfaType) {
  case TotpMfaType:
    return "TOTP";
  case EmailMfaType:
    return i18next.t("general:Email");
  case SmsMfaType:
    return i18next.t("general:Phone");
  case RadiusMfaType:
    return "RADIUS";
  case RecoveryMfaType:
    return i18next.t("mfa:Recovery code");
  default:
    return mfaType;
  }
}

interface StepUpMfaDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  /** the client ID of the application the proof token is issued for */
  clientId: string;
  /** names the action the check approves, copied into the proof token */
  action?: string;
  title?: string;
  description?: string;
  onVerified: (result: StepUpResult) => void;
}

/**
 * Asks the signed-in user for a fresh code of an MFA factor they have already enrolled, e.g. before
 * a payment or an approval, and hands back a short-lived single-use proof token for the application.
 */
export function StepUpMfaDialog({open, onOpenChange, clientId, action, title, description, onVerified}: StepUpMfaDialogProps) {
  const [methods, setMethods] = React.useState<StepUpMethod[] | null>(null);
  const [loadError, setLoadError] = React.useState("");
  const [mfaType, setMfaType] = React.useState("");
  const [passcode, setPasscode] = React.useState("");
  const [error, setError] = React.useState("");
  const [verifying, setVerifying] = React.useState(false);
  const [sending, setSending] = React.useState(false);
  const [sentTo, setSentTo] = React.useState("");
  const [countdown, setCountdown] = React.useState(0);

  React.useEffect(() => {
    if (!open) {
      return;
    }
    setMethods(null);
    setLoadError("");
    setPasscode("");
    setError("");
    setSentTo("");
    setCountdown(0);
    MfaBackend.getStepUpMethods(clientId).then((res: any) => {
      if (res.status !== "ok") {
        setLoadError(res.msg);
        return;
      }
      const list: StepUpMethod[] = res.data ?? [];
      setMethods(list);
      setMfaType(list[0]?.mfaType ?? "");
    }).catch((e) => setLoadError(String(e)));
  }, [open, clientId]);

  React.useEffect(() => {
    if (countdown <= 0) {
      return;
    }
    const timer = setTimeout(() => setCountdown(countdown - 1), 1000);
    return () => clearTimeout(timer);
  }, [countdown]);

  const method = methods?.find((m) => m.mfaType === mfaType);
  const needsSend = mfaType === EmailMfaType || mfaType === SmsMfaType;
  const usesBoxes = mfaType === TotpMfaType || needsSend;

  const selectMethod = (type: string) => {
    setMfaType(type);
    setPasscode("");
    setError("");
    setSentTo("");
    setCountdown(0);
  };

  const sendCode = () => {
    setSending(true);
    setError("");
    MfaBackend.sendStepUpCode(clientId, mfaType).then((res: any) => {
      if (res.status === "ok") {
        setSentTo(res.data ?? method?.dest ?? "");
        setCountdown(60);
      } else {
        setError(res.msg);
      }
    }).finally(() => setSending(false));
  };

  const verify = (code = passcode) => {
    if (!code || verifying) {
      return;
    }
    setVerifying(true);
    setError("");
    MfaBackend.verifyStepUp({clientId, mfaType, passcode: code, action}).then((res: any) => {
      if (res.status === "ok") {
        onVerified(res.data);
        onOpenChange(false);
      } else {
        setError(res.msg);
        setPasscode("");
      }
    }).finally(() => setVerifying(false));
  };

  const hint = () => {
    switch (mfaType) {
    case TotpMfaType:
      return i18next.t("application:Enter the 6-digit code from your authenticator app");
    case RadiusMfaType:
      return i18next.t("mfa:You have enabled Multi-Factor Authentication, please enter the RADIUS password");
    case RecoveryMfaType:
      return i18next.t("mfa:Multi-factor recover description");
    default:
      return sentTo ? `${i18next.t("application:A code has been sent to")} ${sentTo}` : `${i18next.t("application:We will send a code to")} ${method?.dest ?? ""}`;
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <ShieldCheck className="size-5 text-primary" />
            {title ?? i18next.t("application:Confirm it is you")}
          </DialogTitle>
          <DialogDescription>
            {description ?? i18next.t("application:This action needs a fresh code from your multi-factor authentication")}
          </DialogDescription>
        </DialogHeader>

        {loadError ? (
          <p className="text-sm text-destructive">{loadError}</p>
        ) : methods === null ? (
          <p className="text-sm text-muted-foreground">{i18next.t("general:Loading")}...</p>
        ) : methods.length === 0 ? (
          <p className="text-sm text-muted-foreground">{i18next.t("application:You have not enabled any multi-factor authentication this application accepts, please set it up on your account page first")}</p>
        ) : (
          <div className="space-y-4">
            {methods.length > 1 ? (
              <div className="flex flex-wrap gap-2">
                {methods.map((m) => (
                  <button
                    key={m.mfaType}
                    type="button"
                    onClick={() => selectMethod(m.mfaType)}
                    className={cn(
                      "rounded-md border px-3 py-1.5 text-sm transition-colors",
                      m.mfaType === mfaType ? "border-primary bg-primary/10 text-primary" : "hover:bg-accent",
                    )}
                  >
                    {getStepUpMethodLabel(m.mfaType)}
                  </button>
                ))}
              </div>
            ) : null}

            <p className="text-sm text-muted-foreground">{hint()}</p>

            {needsSend ? (
              <Button variant="outline" className="w-full" loading={sending} disabled={countdown > 0} onClick={sendCode}>
                {countdown > 0 ? `${i18next.t("mfa:Send again")} (${countdown}s)` : sentTo ? i18next.t("mfa:Send again") : i18next.t("application:Send code")}
              </Button>
            ) : null}

            {usesBoxes ? (
              <CodeInput
                value={passcode}
                onChange={setPasscode}
                autoFocus={mfaType === TotpMfaType}
                disabled={verifying || (needsSend && !sentTo)}
                onComplete={() => verify()}
              />
            ) : (
              <Input
                autoFocus
                type={mfaType === RadiusMfaType ? "password" : "text"}
                value={passcode}
                placeholder={mfaType === RecoveryMfaType ? i18next.t("mfa:Recovery code") : ""}
                onChange={(e) => setPasscode(e.target.value.trim())}
                onKeyDown={(e) => {
                  if (e.key === "Enter") {
                    verify();
                  }
                }}
              />
            )}

            {error ? <p className="text-sm text-destructive">{error}</p> : null}
          </div>
        )}

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>{i18next.t("general:Cancel")}</Button>
          <Button loading={verifying} disabled={!methods?.length || !passcode} onClick={() => verify()}>
            {i18next.t("mfa:Verify Code")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
