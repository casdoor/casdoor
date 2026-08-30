import * as React from "react";
import i18next from "i18next";
import {Button} from "@/components/ui/button";
import {Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle} from "@/components/ui/dialog";
import {Input} from "@/components/ui/input";
import {Label} from "@/components/ui/label";
import {SendCodeInput} from "@/components/auth/SendCodeInput";
import * as UserBackend from "@/backend/UserBackend";
import * as Setting from "@/lib/setting";

interface ResetModalProps {
  application: any;
  destType: "email" | "phone";
  buttonText: string;
  countryCode?: string;
  disabled?: boolean;
}

/**
 * "Reset Email..." / "Reset Phone..." of the user page. Ported from
 * web/src/common/modal/ResetModal.js: the code is sent with method "reset" and
 * the new destination is confirmed through /api/reset-email-or-phone.
 */
export function ResetModal({application, destType, buttonText, countryCode = "", disabled}: ResetModalProps) {
  const [open, setOpen] = React.useState(false);
  const [submitting, setSubmitting] = React.useState(false);
  const [dest, setDest] = React.useState("");
  const [code, setCode] = React.useState("");

  const close = () => {
    setOpen(false);
    setDest("");
    setCode("");
  };

  const handleOk = () => {
    if (dest === "") {
      Setting.showMessage("error", destType === "phone"
        ? i18next.t("user:Phone cannot be empty")
        : i18next.t("user:Email cannot be empty"));
      return;
    }
    if (code === "") {
      Setting.showMessage("error", i18next.t("code:Empty code"));
      return;
    }

    setSubmitting(true);
    UserBackend.resetEmailOrPhone(dest, destType, code).then((res: any) => {
      if (res.status === "ok") {
        Setting.showMessage("success", i18next.t("user:Email/phone reset successfully"));
        window.location.reload();
      } else {
        Setting.showMessage("error", i18next.t("user:" + res.msg));
        setSubmitting(false);
      }
    });
  };

  return (
    <>
      <Button variant="outline" disabled={disabled} onClick={() => setOpen(true)}>
        {buttonText}
      </Button>
      <Dialog open={open} onOpenChange={(next) => (next ? setOpen(true) : close())}>
        <DialogContent className="sm:max-w-[600px]">
          <DialogHeader>
            <DialogTitle>{buttonText}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>{destType === "email" ? i18next.t("user:New Email") : i18next.t("user:New phone")}</Label>
              <div className="flex items-center gap-2">
                {destType === "phone" && countryCode !== "" ? (
                  <span className="shrink-0 text-sm text-muted-foreground">+{Setting.getCountryCode(countryCode)}</span>
                ) : null}
                <Input
                  value={dest}
                  placeholder={destType === "email"
                    ? i18next.t("user:Input your email")
                    : i18next.t("user:Input your phone number")}
                  onChange={(e) => setDest(e.target.value)}
                />
              </div>
            </div>
            <div className="space-y-2">
              <Label>{i18next.t("code:Code you received")}</Label>
              <SendCodeInput
                value={code}
                onChange={setCode}
                method="reset"
                dest={dest}
                destType={destType}
                countryCode={countryCode}
                applicationId={Setting.getApplicationName(application)}
                application={application}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={close}>{i18next.t("general:Cancel")}</Button>
            <Button loading={submitting} onClick={handleOk}>{buttonText}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
