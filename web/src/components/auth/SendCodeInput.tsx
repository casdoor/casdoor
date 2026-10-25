import * as React from "react";
import i18next from "i18next";
import {Button} from "@/components/ui/button";
import {CodeInput} from "@/components/auth/CodeInput";
import {CaptchaModal} from "@/components/common/CaptchaModal";
import * as AuthBackend from "@/backend/AuthBackend";
import * as UserBackend from "@/backend/UserBackend";
import * as Setting from "@/lib/setting";
import {cn} from "@/lib/utils";

export interface CaptchaValues {
  captchaType?: string;
  captchaToken?: string;
  clientSecret?: string;
}

interface SendCodeInputProps {
  value: string;
  onChange: (value: string) => void;
  /** why the code is sent: "login" | "signup" | "forget" | "mfaSetup" | "mfaAuth" */
  method: string;
  /** where it is sent: "email" | "phone" */
  destType: "email" | "phone";
  /** the email address or phone number itself */
  dest: string;
  countryCode?: string;
  /** "owner/name" of the application, see Setting.getApplicationName */
  applicationId: string;
  /** username the code must belong to, used by the forget-password flow */
  checkUser?: string;
  disabled?: boolean;
  /** the application object, needed for the captcha rule and the resend timeout */
  application?: any;
  /** set when the sign-in form renders its own inline captcha */
  captchaValue?: CaptchaValues;
  useInlineCaptcha?: boolean;
  refreshCaptcha?: () => void;
  /** the signup item adds its own class */
  className?: string;
  /** called once every digit is typed */
  onComplete?: () => void;
}

interface UseSendCodeOptions {
  method: string;
  destType: "email" | "phone";
  dest: string;
  countryCode?: string;
  applicationId: string;
  checkUser?: string;
  application?: any;
  captchaValue?: CaptchaValues;
  useInlineCaptcha?: boolean;
  refreshCaptcha?: () => void;
  /** called once the backend accepted the code request */
  onSent?: () => void;
}

/**
 * Sending a verification code, without the field: the resend countdown and the
 * captcha rule — Never sends straight away, Always challenges first,
 * Dynamic/Internet-Only asks the backend whether this user needs a challenge.
 * `captchaModal` has to be rendered for the non-inline captcha to show up.
 */
export function useSendCode({
  method,
  destType,
  dest,
  countryCode = "",
  applicationId,
  checkUser,
  application,
  captchaValue,
  useInlineCaptcha,
  refreshCaptcha,
  onSent,
}: UseSendCodeOptions) {
  const [seconds, setSeconds] = React.useState(0);
  const [sending, setSending] = React.useState(false);
  const [captchaVisible, setCaptchaVisible] = React.useState(false);

  const resendTimeout = application?.codeResendTimeout > 0 ? application.codeResendTimeout : 60;

  React.useEffect(() => {
    if (seconds <= 0) {
      return;
    }
    const timer = window.setTimeout(() => setSeconds((s) => s - 1), 1000);
    return () => window.clearTimeout(timer);
  }, [seconds]);

  const send = (captchaType: string, captchaToken: string, clientSecret: string) => {
    setCaptchaVisible(false);
    if (!dest) {
      Setting.showMessage("error", i18next.t("login:Please input your Email or Phone!"));
      return;
    }
    setSending(true);
    UserBackend.sendCode(
      captchaType,
      captchaToken,
      clientSecret,
      method,
      countryCode,
      dest,
      destType,
      applicationId,
      checkUser ?? "",
    )
      .then((sent: boolean) => {
        // UserBackend.sendCode() already showed the success or the error message
        if (sent) {
          setSeconds(resendTimeout);
          onSent?.();
        } else if (useInlineCaptcha) {
          refreshCaptcha?.();
        }
      })
      .catch(() => {
        if (useInlineCaptcha) {
          refreshCaptcha?.();
        }
      })
      .finally(() => setSending(false));
  };

  const request = () => {
    const sendWithoutCaptcha = () => send("none", "", "");

    const sendWithCaptcha = () => {
      if (!useInlineCaptcha) {
        setCaptchaVisible(true);
        return;
      }
      // the client secret is validated server-side
      if (!captchaValue?.captchaType || !captchaValue?.captchaToken) {
        Setting.showMessage("error", i18next.t("general:Please complete the captcha correctly"));
        return;
      }
      send(captchaValue.captchaType, captchaValue.captchaToken, captchaValue.clientSecret ?? "");
    };

    if (!application) {
      sendWithoutCaptcha();
      return;
    }

    const captchaRule = Setting.getCaptchaRule(application);
    if (captchaRule === Setting.CaptchaRule.Never) {
      sendWithoutCaptcha();
      return;
    }
    if (captchaRule === Setting.CaptchaRule.Always) {
      sendWithCaptcha();
      return;
    }
    if (captchaRule === Setting.CaptchaRule.Dynamic || captchaRule === Setting.CaptchaRule.InternetOnly) {
      AuthBackend.getCaptchaStatus({
        organization: application?.organization,
        username: checkUser || dest,
        application: application?.name,
      })
        .then((res: any) => {
          if (res.status === "ok" && res.data) {
            sendWithCaptcha();
          } else {
            sendWithoutCaptcha();
          }
        })
        .catch(sendWithoutCaptcha);
      return;
    }
    sendWithoutCaptcha();
  };

  const captchaModal = !useInlineCaptcha && application ? (
    <CaptchaModal
      owner={application.owner}
      name={application.name}
      visible={captchaVisible}
      onOk={send}
      onCancel={() => setCaptchaVisible(false)}
    />
  ) : null;

  return {seconds, sending, request, captchaModal, resetCountdown: () => setSeconds(0)};
}

/**
 * Verification-code boxes with the "Get Code" link under them, replacing
 * common/SendCodeInput.js.
 */
export function SendCodeInput({
  value,
  onChange,
  method,
  destType,
  dest,
  countryCode = "",
  applicationId,
  checkUser,
  disabled,
  application,
  captchaValue,
  useInlineCaptcha,
  refreshCaptcha,
  className,
  onComplete,
}: SendCodeInputProps) {
  const [sent, setSent] = React.useState(false);
  const {seconds, sending, request, captchaModal} = useSendCode({
    method,
    destType,
    dest,
    countryCode,
    applicationId,
    checkUser,
    application,
    captchaValue,
    useInlineCaptcha,
    refreshCaptcha,
    onSent: () => setSent(true),
  });

  return (
    <>
      <div className="space-y-2">
        <CodeInput className={className} value={value} disabled={disabled} onChange={onChange} onComplete={onComplete} />
        <div className="flex justify-end">
          <Button
            type="button"
            variant="link"
            className={cn("h-auto p-0", seconds > 0 && "text-muted-foreground")}
            loading={sending}
            disabled={disabled || seconds > 0}
            onClick={request}
          >
            {seconds > 0
              ? `${i18next.t("login:Resend code")} (${seconds}s)`
              : sent ? i18next.t("login:Resend code") : i18next.t("code:Get Code")}
          </Button>
        </div>
      </div>
      {captchaModal}
    </>
  );
}
