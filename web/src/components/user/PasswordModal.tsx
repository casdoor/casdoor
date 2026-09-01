import * as React from "react";
import i18next from "i18next";
import {Button} from "@/components/ui/button";
import {Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle} from "@/components/ui/dialog";
import {Input} from "@/components/ui/input";
import {Label} from "@/components/ui/label";
import {PasswordRequirements, checkPasswordComplexity} from "@/lib/password-checker";
import * as Obfuscator from "@/auth/Obfuscator";
import * as UserBackend from "@/backend/UserBackend";
import * as Setting from "@/lib/setting";

interface PasswordModalProps {
  user: any;
  /** the saved user name: the API works on the persisted user, not the edited one */
  userName: string;
  organization: any;
  account: any;
  disabled?: boolean;
  onPasswordUpdated?: () => void;
}

/**
 * "Modify password..." of the user page, ported from
 * `web/src/common/modal/PasswordModal.js`.
 *
 * Three things the plain inline input could not do and that the backend relies on:
 * the old password (`SetPassword` in controllers/user.go rejects a non-admin who does
 * not supply it), the organization password obfuscator, and the `passwordOptions`
 * complexity rules.
 */
export function PasswordModal({user, userName, organization, account, disabled, onPasswordUpdated}: PasswordModalProps) {
  const [open, setOpen] = React.useState(false);
  const [submitting, setSubmitting] = React.useState(false);
  const [oldPassword, setOldPassword] = React.useState("");
  const [newPassword, setNewPassword] = React.useState("");
  const [rePassword, setRePassword] = React.useState("");

  const passwordOptions: string[] = organization?.passwordOptions ?? [];
  const hasOldPassword = user.password !== "" || user.ldap !== "";
  const buttonText = hasOldPassword ? i18next.t("user:Modify password...") : i18next.t("user:Set password...");
  // an admin resetting someone else's password is not asked for the old one
  const needOldPassword = hasOldPassword && !Setting.isLocalAdminUser(account);

  const newPasswordError = newPassword === "" ? "" : checkPasswordComplexity(newPassword, passwordOptions);
  const rePasswordError =
    rePassword !== "" && rePassword !== newPassword
      ? i18next.t("signup:Your confirmed password is inconsistent with the password!")
      : "";

  const close = () => {
    setOpen(false);
    setOldPassword("");
    setNewPassword("");
    setRePassword("");
  };

  const handleOk = () => {
    if (newPassword === "" || rePassword === "") {
      Setting.showMessage("error", i18next.t("user:Empty input!"));
      return;
    }
    if (newPassword !== rePassword) {
      Setting.showMessage("error", i18next.t("user:Two passwords you typed do not match."));
      return;
    }
    if (organization === null || organization === undefined) {
      Setting.showMessage("error", i18next.t("general:Organization is null"));
      return;
    }

    const errorMsg = checkPasswordComplexity(newPassword, passwordOptions);
    if (errorMsg !== "") {
      Setting.showMessage("error", errorMsg);
      return;
    }

    // Encrypt passwords using the password obfuscator if the organization configures one
    let encryptedOldPassword = oldPassword;
    let encryptedNewPassword = newPassword;

    if (organization.passwordObfuscatorType && organization.passwordObfuscatorType !== "Plain") {
      const [oldPasswordCipher, oldPasswordError] = Obfuscator.encryptByPasswordObfuscator(
        organization.passwordObfuscatorType,
        organization.passwordObfuscatorKey,
        oldPassword,
      );
      if (oldPasswordError) {
        Setting.showMessage("error", oldPasswordError);
        return;
      }
      encryptedOldPassword = oldPasswordCipher;

      const [newPasswordCipher, newPasswordCipherError] = Obfuscator.encryptByPasswordObfuscator(
        organization.passwordObfuscatorType,
        organization.passwordObfuscatorKey,
        newPassword,
      );
      if (newPasswordCipherError) {
        Setting.showMessage("error", newPasswordCipherError);
        return;
      }
      encryptedNewPassword = newPasswordCipher;
    }

    setSubmitting(true);
    UserBackend.setPassword(user.owner, userName, encryptedOldPassword, encryptedNewPassword)
      .then((res: any) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("user:Password set successfully"));
          close();
          if (account?.owner === user.owner && account?.name === userName) {
            const wasForced = account.needUpdatePassword;
            account.needUpdatePassword = false;
            onPasswordUpdated?.();

            // continue the login that was interrupted by "Need update password"
            const signinUrl = sessionStorage.getItem("signinUrl");
            if (wasForced && signinUrl) {
              sessionStorage.removeItem("signinUrl");
              Setting.goToLink(signinUrl);
            }
          }
        } else {
          Setting.showMessage("error", i18next.t(`user:${res.msg}`));
        }
      })
      .finally(() => setSubmitting(false));
  };

  return (
    <>
      <Button disabled={disabled} onClick={() => setOpen(true)}>
        {buttonText}
      </Button>
      <Dialog open={open} onOpenChange={(next) => (next ? setOpen(true) : close())}>
        <DialogContent className="sm:max-w-[600px]">
          <DialogHeader>
            <DialogTitle>{i18next.t("general:Password")}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            {needOldPassword ? (
              <div className="space-y-2">
                <Label htmlFor="old-password">{i18next.t("user:Old Password")}</Label>
                <Input
                  id="old-password"
                  type="password"
                  autoComplete="current-password"
                  value={oldPassword}
                  placeholder={i18next.t("user:input password")}
                  onChange={(e) => setOldPassword(e.target.value)}
                />
              </div>
            ) : null}
            <div className="space-y-2">
              <Label htmlFor="new-password">{i18next.t("user:New Password")}</Label>
              <Input
                id="new-password"
                type="password"
                autoComplete="new-password"
                value={newPassword}
                placeholder={i18next.t("user:input password")}
                onChange={(e) => setNewPassword(e.target.value)}
              />
              <PasswordRequirements options={passwordOptions} password={newPassword} />
              {passwordOptions.length === 0 && newPasswordError ? (
                <p className="text-xs text-destructive">{newPasswordError}</p>
              ) : null}
            </div>
            <div className="space-y-2">
              <Label htmlFor="re-password">{i18next.t("user:Re-enter New")}</Label>
              <Input
                id="re-password"
                type="password"
                autoComplete="new-password"
                value={rePassword}
                placeholder={i18next.t("user:input password")}
                onChange={(e) => setRePassword(e.target.value)}
              />
              {rePasswordError ? <p className="text-xs text-destructive">{rePasswordError}</p> : null}
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={close}>{i18next.t("general:Cancel")}</Button>
            <Button loading={submitting} onClick={handleOk}>{i18next.t("user:Set Password")}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
