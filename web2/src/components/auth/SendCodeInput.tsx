import * as React from "react";
import i18next from "i18next";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import * as UserBackend from "@/backend/UserBackend";
import * as Setting from "@/lib/setting";

interface SendCodeInputProps {
  value: string;
  onChange: (value: string) => void;
  /** "email" | "phone" */
  method: string;
  /** email address or phone number the code is sent to */
  dest: string;
  countryCode?: string;
  /** "login" | "signup" | "reset" | "forget" ... */
  type: string;
  applicationId: string;
  checkUser?: string;
  disabled?: boolean;
}

const COUNTDOWN_SECONDS = 60;

/** Verification-code field with the "Send code" button, replacing common/SendCodeInput.js. */
export function SendCodeInput({
  value,
  onChange,
  method,
  dest,
  countryCode = "",
  type,
  applicationId,
  checkUser,
  disabled,
}: SendCodeInputProps) {
  const [seconds, setSeconds] = React.useState(0);
  const [sending, setSending] = React.useState(false);

  React.useEffect(() => {
    if (seconds <= 0) {
      return;
    }
    const timer = window.setTimeout(() => setSeconds((s) => s - 1), 1000);
    return () => window.clearTimeout(timer);
  }, [seconds]);

  const send = () => {
    if (!dest) {
      Setting.showMessage("error", i18next.t("login:Please input your Email or Phone!"));
      return;
    }
    setSending(true);
    UserBackend.sendCode("none", "", "", method, countryCode, dest, type, applicationId, checkUser ?? "")
      .then((res: any) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("code:Code Sent"));
          setSeconds(COUNTDOWN_SECONDS);
        } else {
          Setting.showMessage("error", res.msg);
        }
      })
      .finally(() => setSending(false));
  };

  return (
    <div className="flex gap-2">
      <Input
        value={value}
        disabled={disabled}
        placeholder={i18next.t("code:Enter your code")}
        onChange={(e) => onChange(e.target.value)}
      />
      <Button
        type="button"
        variant="outline"
        className="shrink-0"
        loading={sending}
        disabled={disabled || seconds > 0}
        onClick={send}
      >
        {seconds > 0 ? `${seconds}s` : i18next.t("code:Get Code")}
      </Button>
    </div>
  );
}
