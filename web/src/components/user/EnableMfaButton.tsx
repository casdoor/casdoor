import * as React from "react";
import i18next from "i18next";
import {Button} from "@/components/ui/button";
import {Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle} from "@/components/ui/dialog";
import {EmailMfaType} from "@/components/auth/mfa/constants";
import * as MfaBackend from "@/backend/MfaBackend";
import * as Setting from "@/lib/setting";

/**
 * Turns an email or SMS factor on for another user, which an admin can do without
 * the user's own verification code. Port of `web/src/common/modal/EnableMfaModal.js`:
 * opening the dialog initiates the setup, confirming it enables the factor.
 */
export function EnableMfaButton({user, mfaType, onSuccess}: {user: any; mfaType: string; onSuccess: () => void}) {
  const [open, setOpen] = React.useState(false);
  const [loading, setLoading] = React.useState(false);

  const isEmail = mfaType === EmailMfaType;
  const destination = isEmail ? user?.email : user?.phone;

  React.useEffect(() => {
    if (!open || !user) {
      return;
    }
    MfaBackend.MfaSetupInitiate({mfaType, ...user}).then((res: any) => {
      if (res.status === "error") {
        Setting.showMessage("error", i18next.t("mfa:Failed to initiate MFA"));
      }
    });
  }, [open, user, mfaType]);

  const enable = () => {
    setLoading(true);
    MfaBackend.MfaSetupEnable({mfaType, ...user})
      .then((res: any) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Enabled successfully"));
          setOpen(false);
          onSuccess();
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to enable")}: ${res.msg}`);
        }
      })
      .finally(() => setLoading(false));
  };

  return (
    <>
      <Button
        size="sm"
        onClick={() => {
          // there is nothing to send the code to yet
          if (!destination) {
            Setting.showMessage(
              "error",
              i18next.t(isEmail ? "login:Please input your Email!" : "signup:Please input your phone number!"),
            );
            return;
          }
          setOpen(true);
        }}
      >
        {i18next.t("general:Enable")}
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="sm:max-w-[460px]">
          <DialogHeader>
            <DialogTitle>{i18next.t("mfa:Enable multi-factor authentication")}</DialogTitle>
          </DialogHeader>
          <p className="space-y-1 text-sm">
            {i18next.t("mfa:Please confirm the information below")}
            <br />
            <b>{i18next.t("general:User")}</b>: {`${user?.owner}/${user?.name}`}
            <br />
            <b>{i18next.t(isEmail ? "general:Email" : "general:Phone")}</b>: {destination}
          </p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setOpen(false)}>{i18next.t("general:Cancel")}</Button>
            <Button loading={loading} onClick={enable}>{i18next.t("general:OK")}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
