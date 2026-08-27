import * as React from "react";
import i18next from "i18next";
import {Button} from "@/components/ui/button";
import {Checkbox} from "@/components/ui/checkbox";
import {Input} from "@/components/ui/input";
import {Label} from "@/components/ui/label";
import {SendCodeInput} from "@/components/auth/SendCodeInput";
import * as AuthBackend from "@/backend/AuthBackend";
import * as Setting from "@/lib/setting";

export const NextMfa = "NextMfa";
export const RequiredMfa = "RequiredMfa";
export const SmsMfaType = "sms";
export const EmailMfaType = "email";
export const TotpMfaType = "app";
export const RecoveryMfaType = "recovery";

interface MfaVerifyProps {
  /** the login values that produced the "NextMfa" answer */
  formValues: Record<string, any>;
  authParams: any;
  /** res.data2 of the login call */
  mfaProps: any;
  application: any;
  onSuccess: (res: any) => void;
}

/**
 * Second factor step of the sign-in flow. It re-posts the original login payload
 * with `passcode`/`recoveryCode` added, exactly like the antd MfaAuthVerifyForm.
 */
export function MfaVerify({formValues, authParams, mfaProps, application, onSuccess}: MfaVerifyProps) {
  const [loading, setLoading] = React.useState(false);
  const [passcode, setPasscode] = React.useState("");
  const [recoveryCode, setRecoveryCode] = React.useState("");
  const [useRecovery, setUseRecovery] = React.useState(false);
  const [remember, setRemember] = React.useState(false);

  const post = (extra: Record<string, any>) => {
    setLoading(true);
    const values = {...formValues, password: "", username: "", mfaType: mfaProps?.mfaType, ...extra};
    const loginFunction = formValues.type === "cas" ? AuthBackend.loginCas : AuthBackend.login;
    loginFunction(values, authParams)
      .then((res: any) => {
        if (res.status === "ok") {
          onSuccess(res);
        } else {
          Setting.showMessage("error", res.msg);
        }
      })
      .catch((err: any) => Setting.showMessage("error", err.message))
      .finally(() => setLoading(false));
  };

  if (useRecovery) {
    return (
      <div className="space-y-4">
        <h2 className="text-center text-lg font-semibold">{i18next.t("mfa:Multi-factor authentication")}</h2>
        <div className="space-y-2">
          <Label htmlFor="recoveryCode">{i18next.t("mfa:Recovery code")}</Label>
          <Input id="recoveryCode" value={recoveryCode} onChange={(e) => setRecoveryCode(e.target.value)} />
        </div>
        <Button className="w-full" loading={loading} onClick={() => post({recoveryCode})}>
          {i18next.t("mfa:Verify Code")}
        </Button>
        <Button variant="link" className="w-full" onClick={() => setUseRecovery(false)}>
          {i18next.t("general:Back")}
        </Button>
      </div>
    );
  }

  const isCodeType = mfaProps?.mfaType === SmsMfaType || mfaProps?.mfaType === EmailMfaType;

  return (
    <div className="space-y-4">
      <h2 className="text-center text-lg font-semibold">{i18next.t("mfa:Multi-factor authentication")}</h2>
      {isCodeType ? (
        <>
          <p className="text-sm text-muted-foreground">
            {i18next.t("mfa:You have enabled Multi-Factor Authentication, Please click 'Get Code' to continue")}
          </p>
          <SendCodeInput
            value={passcode}
            onChange={setPasscode}
            method={mfaProps.mfaType === EmailMfaType ? "email" : "phone"}
            dest={mfaProps?.secret ?? ""}
            countryCode={mfaProps?.countryCode ?? ""}
            type="mfaAuth"
            applicationId={`${application?.owner}/${application?.name}`}
          />
        </>
      ) : (
        <>
          <p className="text-sm text-muted-foreground">
            {i18next.t("mfa:You have enabled Multi-Factor Authentication, please enter the TOTP code")}
          </p>
          <Input
            autoFocus
            inputMode="numeric"
            value={passcode}
            placeholder={i18next.t("code:Enter your code")}
            onChange={(e) => setPasscode(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") {
                post({passcode, enableMfaRemember: remember});
              }
            }}
          />
        </>
      )}

      <div className="flex items-center gap-2">
        <Checkbox id="mfa-remember" checked={remember} onCheckedChange={(v) => setRemember(v === true)} />
        <Label htmlFor="mfa-remember" className="text-sm font-normal">
          {i18next.t("mfa:Remember this account for")}
        </Label>
      </div>

      <Button className="w-full" loading={loading} onClick={() => post({passcode, enableMfaRemember: remember})}>
        {i18next.t("mfa:Verify Code")}
      </Button>
      <Button variant="link" className="w-full" onClick={() => setUseRecovery(true)}>
        {i18next.t("mfa:Have problems?")}
      </Button>
    </div>
  );
}
