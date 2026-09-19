// Copyright 2026 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import * as React from "react";
import i18next from "i18next";
import {useParams} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {Label} from "@/components/ui/label";
import {PasswordInput} from "@/components/common/PasswordInput";
import {Loading} from "@/components/common/Loading";
import {AuthLayout} from "@/components/auth/AuthLayout";
import {PasswordRequirements, checkPasswordComplexity} from "@/lib/password-checker";
import {authConfig} from "@/auth/Auth";
import * as Obfuscator from "@/auth/Obfuscator";
import * as ApplicationBackend from "@/backend/ApplicationBackend";
import * as UserBackend from "@/backend/UserBackend";
import * as Setting from "@/lib/setting";
import {useAccount} from "@/hooks/use-account";

/**
 * The page a sign-in lands on when the account is flagged "need update password":
 * the account page's password modal as a standalone form in the application's
 * auth layout, which then resumes the sign-in it interrupted.
 */
export default function UpdatePasswordPage() {
  const params = useParams();
  const applicationName = params.applicationName ?? authConfig.appName;
  const {account} = useAccount();

  const [application, setApplication] = React.useState<any>(undefined);
  const [oldPassword, setOldPassword] = React.useState("");
  const [password, setPassword] = React.useState("");
  const [confirm, setConfirm] = React.useState("");
  const [passwordFocused, setPasswordFocused] = React.useState(false);
  const [loading, setLoading] = React.useState(false);

  React.useEffect(() => {
    ApplicationBackend.getApplication("admin", applicationName)
      .then((res: any) => setApplication(res.status === "ok" ? res.data : null))
      .catch(() => setApplication(null));
  }, [applicationName]);

  // the sign-in that was interrupted; without one the console is the destination
  const signinUrl = Setting.getStoredSigninUrl();

  React.useEffect(() => {
    if (account === null) {
      Setting.goToLink(signinUrl || "/login");
    }
  }, [account, signinUrl]);

  if (application === undefined || !account) {
    return <Loading className="min-h-screen" />;
  }

  // same rules as PasswordModal: the organization's password options, and no old
  // password from an account that has none or from a local admin
  const organization = account.organization ?? application?.organizationObj;
  const passwordOptions: string[] = organization?.passwordOptions ?? [];
  const needOldPassword = (account.password !== "" || account.ldap !== "") && !Setting.isLocalAdminUser(account);

  const submit = () => {
    if (password === "" || confirm === "") {
      Setting.showMessage("error", i18next.t("user:Empty input!"));
      return;
    }
    if (password !== confirm) {
      Setting.showMessage("error", i18next.t("user:Two passwords you typed do not match."));
      return;
    }
    if (!organization) {
      Setting.showMessage("error", i18next.t("general:Organization is null"));
      return;
    }
    const complexityError = checkPasswordComplexity(password, passwordOptions);
    if (complexityError !== "") {
      Setting.showMessage("error", complexityError);
      return;
    }

    let encryptedOldPassword = oldPassword;
    let encryptedNewPassword = password;
    if (organization.passwordObfuscatorType && organization.passwordObfuscatorType !== "Plain") {
      const encrypt = (plain: string) =>
        Obfuscator.encryptByPasswordObfuscator(organization.passwordObfuscatorType, organization.passwordObfuscatorKey, plain);
      const [oldCipher, oldError] = encrypt(oldPassword);
      if (oldError) {
        Setting.showMessage("error", oldError);
        return;
      }
      const [newCipher, newError] = encrypt(password);
      if (newError) {
        Setting.showMessage("error", newError);
        return;
      }
      encryptedOldPassword = oldCipher;
      encryptedNewPassword = newCipher;
    }

    setLoading(true);
    UserBackend.setPassword(account.owner, account.name, encryptedOldPassword, encryptedNewPassword)
      .then((res: any) => {
        if (res.status !== "ok") {
          Setting.showMessage("error", i18next.t(`user:${res.msg}`));
          return;
        }
        Setting.showMessage("success", i18next.t("user:Password set successfully"));
        sessionStorage.removeItem("signinUrl");
        // the session survived the change, so the sign-in page finishes the
        // authorization with it instead of asking to pick the account again
        Setting.goToLink(signinUrl ? `${signinUrl}${signinUrl.includes("?") ? "&" : "?"}silentSignin=1` : "/");
      })
      .finally(() => setLoading(false));
  };

  return (
    <AuthLayout application={application}>
      <form
        className="space-y-4"
        onSubmit={(e) => {
          e.preventDefault();
          submit();
        }}
      >
        <h1 className="text-xl font-semibold tracking-tight">{i18next.t("user:You need to update your password")}</h1>
        {needOldPassword ? (
          <div className="space-y-2">
            <Label htmlFor="oldPassword">{i18next.t("user:Old Password")}</Label>
            <PasswordInput
              id="oldPassword"
              autoFocus
              autoComplete="current-password"
              value={oldPassword}
              onChange={(e) => setOldPassword(e.target.value)}
            />
          </div>
        ) : null}
        <div className="space-y-2">
          <Label htmlFor="newPassword">{i18next.t("user:New Password")}</Label>
          <PasswordInput
            id="newPassword"
            autoFocus={!needOldPassword}
            autoComplete="new-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            onFocus={() => setPasswordFocused(true)}
          />
          {passwordFocused ? <PasswordRequirements options={passwordOptions} password={password} /> : null}
        </div>
        <div className="space-y-2">
          <Label htmlFor="confirmPassword">{i18next.t("user:Re-enter New")}</Label>
          <PasswordInput
            id="confirmPassword"
            autoComplete="new-password"
            value={confirm}
            onChange={(e) => setConfirm(e.target.value)}
          />
        </div>
        <Button type="submit" className="w-full" loading={loading}>
          {i18next.t("user:Set Password")}
        </Button>
      </form>
    </AuthLayout>
  );
}
