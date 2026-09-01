import * as React from "react";
import i18next from "i18next";
import {Button} from "@/components/ui/button";
import {Checkbox} from "@/components/ui/checkbox";
import {Input} from "@/components/ui/input";
import {Label} from "@/components/ui/label";
import {SendCodeInput} from "@/components/auth/SendCodeInput";
import * as AuthBackend from "@/backend/AuthBackend";
import * as Setting from "@/lib/setting";
import {
  EmailMfaType,
  PushMfaType,
  RadiusMfaType,
  SmsMfaType,
  TotpMfaType,
} from "@/components/auth/mfa/constants";

export {
  NextMfa,
  RequiredMfa,
  SmsMfaType,
  EmailMfaType,
  TotpMfaType,
  RadiusMfaType,
  PushMfaType,
  RecoveryMfaType,
} from "@/components/auth/mfa/constants";

/** The factor the user marked preferred, falling back to the first one. */
export function getPreferredMfaProp(mfaProps: any[]) {
  return mfaProps.find((item) => item?.isPreferred) ?? mfaProps[0];
}

/** Label of the "use another factor" link, as antd's renderMfaAuthVerifyForm has it. */
function getSwitchLabel(mfaType: string) {
  switch (mfaType) {
  case SmsMfaType:
    return i18next.t("mfa:Use SMS");
  case TotpMfaType:
    return i18next.t("mfa:Use Authenticator App");
  case EmailMfaType:
    return i18next.t("mfa:Use Email");
  default:
    return "";
  }
}

interface MfaVerifyProps {
  /** the login values that produced the "NextMfa" answer */
  formValues: Record<string, any>;
  authParams: any;
  /** res.data2 of the login call: every factor the user has enabled */
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

  // The backend answers with every enabled factor; the user starts on their
  // preferred one and can switch to any of the others.
  const factors: any[] = React.useMemo(
    () => (Array.isArray(mfaProps) ? mfaProps : [mfaProps]).filter(Boolean),
    [mfaProps],
  );
  const [selectedType, setSelectedType] = React.useState<string | undefined>(
    () => getPreferredMfaProp(factors)?.mfaType,
  );
  const selected = factors.find((item) => item.mfaType === selectedType) ?? factors[0];

  const post = (extra: Record<string, any>) => {
    setLoading(true);
    const values = {...formValues, password: "", username: "", mfaType: selected?.mfaType, ...extra};
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
        <h2 className="text-center text-lg font-semibold">{i18next.t("mfa:Multi-factor recover")}</h2>
        <p className="text-sm text-muted-foreground">{i18next.t("mfa:Multi-factor recover description")}</p>
        <div className="space-y-2">
          <Label htmlFor="recoveryCode">{i18next.t("mfa:Recovery code")}</Label>
          <Input
            id="recoveryCode"
            placeholder={i18next.t("mfa:Recovery code")}
            value={recoveryCode}
            onChange={(e) => setRecoveryCode(e.target.value)}
          />
        </div>
        <Button className="w-full" loading={loading} onClick={() => post({recoveryCode})}>
          {i18next.t("forget:Verify")}
        </Button>
        <Button variant="link" className="w-full" onClick={() => setUseRecovery(false)}>
          {i18next.t("mfa:Have problems?")} {i18next.t("mfa:Use SMS verification code")}
        </Button>
      </div>
    );
  }

  const isCodeType = selected?.mfaType === SmsMfaType || selected?.mfaType === EmailMfaType;

  // The antd MfaAuthVerifyForm branches on the factor: SMS/email get a "Get Code"
  // field, TOTP an authenticator code, push the code from the notification, and
  // RADIUS the account's RADIUS password.
  const {prompt, placeholder} = (() => {
    if (selected?.mfaType === PushMfaType) {
      return {
        prompt: i18next.t("mfa:You have enabled Multi-Factor Authentication, please enter the verification code from push notification"),
        placeholder: i18next.t("code:Enter your code"),
      };
    }
    if (selected?.mfaType === RadiusMfaType) {
      return {
        prompt: i18next.t("mfa:You have enabled Multi-Factor Authentication, please enter the RADIUS password"),
        placeholder: i18next.t("general:Password"),
      };
    }
    return {
      prompt: i18next.t("mfa:You have enabled Multi-Factor Authentication, please enter the TOTP code"),
      placeholder: i18next.t("code:Enter your code"),
    };
  })();

  return (
    <div className="space-y-4">
      <h2 className="text-center text-lg font-semibold">{i18next.t("mfa:Multi-factor authentication")}</h2>
      {isCodeType ? (
        <>
          <p className="text-sm text-muted-foreground">
            {i18next.t("mfa:You have enabled Multi-Factor Authentication, Please click 'Get Code' to continue")}
          </p>
          <SendCodeInput
            // remount on a factor switch so the countdown and code do not carry over
            key={selected?.mfaType}
            value={passcode}
            onChange={setPasscode}
            method="mfaAuth"
            destType={selected.mfaType === EmailMfaType ? "email" : "phone"}
            dest={selected?.secret ?? ""}
            countryCode={selected?.countryCode ?? ""}
            application={application}
            applicationId={Setting.getApplicationName(application)}
          />
        </>
      ) : (
        <>
          <p className="text-sm text-muted-foreground">{prompt}</p>
          <Input
            autoFocus
            // RADIUS takes a password, not digits, so only the code factors get a numeric keypad
            inputMode={selected?.mfaType === RadiusMfaType ? "text" : "numeric"}
            value={passcode}
            placeholder={placeholder}
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
          {i18next.t("mfa:Remember this account for {hour} hours").replace("{hour}", selected?.mfaRememberInHours)}
        </Label>
      </div>

      <Button className="w-full" loading={loading} onClick={() => post({passcode, enableMfaRemember: remember})}>
        {i18next.t("mfa:Verify Code")}
      </Button>
      {factors
        .filter((item) => item.mfaType !== selected?.mfaType && getSwitchLabel(item.mfaType))
        .map((item) => (
          <Button
            key={item.mfaType}
            variant="link"
            className="w-full"
            onClick={() => {
              setSelectedType(item.mfaType);
              setPasscode("");
            }}
          >
            {getSwitchLabel(item.mfaType)}
          </Button>
        ))}

      <Button variant="link" className="w-full" onClick={() => setUseRecovery(true)}>
        {i18next.t("mfa:Have problems?")} {i18next.t("mfa:Use a recovery code")}
      </Button>
    </div>
  );
}
