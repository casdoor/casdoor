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
import {useNavigate} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {Label} from "@/components/ui/label";
import {PasswordInput} from "@/components/common/PasswordInput";
import {Loading} from "@/components/common/Loading";
import {AuthLayout} from "@/components/auth/AuthLayout";
import {PasswordRequirements, checkPasswordComplexity} from "@/lib/password-checker";
import {authConfig} from "@/auth/Auth";
import * as Obfuscator from "@/auth/Obfuscator";
import * as ApplicationBackend from "@/backend/ApplicationBackend";
import * as AuthBackend from "@/backend/AuthBackend";
import * as Setting from "@/lib/setting";

/**
 * The welcome page of a new installation started without initAdminPassword: built-in/admin
 * has no password yet and gets its first one here, then signs in on the login page.
 */
export default function InitAdminPage() {
  const navigate = useNavigate();

  const [application, setApplication] = React.useState<any>(undefined);
  const [isPending, setIsPending] = React.useState<boolean | undefined>(undefined);
  const [password, setPassword] = React.useState("");
  const [confirm, setConfirm] = React.useState("");
  const [passwordFocused, setPasswordFocused] = React.useState(false);
  const [loading, setLoading] = React.useState(false);

  React.useEffect(() => {
    AuthBackend.getInitAdminStatus()
      .then((res: any) => setIsPending(res.status === "ok" && res.data === true))
      .catch(() => setIsPending(false));
    ApplicationBackend.getApplication("admin", authConfig.appName)
      .then((res: any) => setApplication(res.status === "ok" ? res.data : null))
      .catch(() => setApplication(null));
  }, []);

  React.useEffect(() => {
    if (isPending === false) {
      navigate("/login", {replace: true});
    }
  }, [isPending, navigate]);

  if (application === undefined || !isPending) {
    return <Loading className="min-h-screen" />;
  }

  const organization = application?.organizationObj;
  const passwordOptions: string[] = organization?.passwordOptions ?? [];

  const submit = () => {
    if (password === "" || confirm === "") {
      Setting.showMessage("error", i18next.t("user:Empty input!"));
      return;
    }
    if (password !== confirm) {
      Setting.showMessage("error", i18next.t("user:Two passwords you typed do not match."));
      return;
    }
    const complexityError = checkPasswordComplexity(password, passwordOptions);
    if (complexityError !== "") {
      Setting.showMessage("error", complexityError);
      return;
    }

    let encryptedPassword = password;
    if (organization?.passwordObfuscatorType && organization.passwordObfuscatorType !== "Plain") {
      const [cipher, error] = Obfuscator.encryptByPasswordObfuscator(organization.passwordObfuscatorType, organization.passwordObfuscatorKey, password);
      if (error) {
        Setting.showMessage("error", error);
        return;
      }
      encryptedPassword = cipher;
    }

    setLoading(true);
    AuthBackend.initAdminPassword(encryptedPassword)
      .then((res: any) => {
        if (res.status !== "ok") {
          Setting.showMessage("error", res.msg);
          return;
        }
        Setting.showMessage("success", i18next.t("user:Password set successfully"));
        navigate("/login", {replace: true});
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
        <div className="space-y-1">
          <h1 className="text-xl font-semibold tracking-tight">{i18next.t("login:Welcome to Casdoor")}</h1>
          <p className="text-sm text-muted-foreground">
            {i18next.t("login:Set the password of the administrator admin to finish the setup")}
          </p>
        </div>
        <div className="space-y-2">
          <Label>{i18next.t("signup:Username")}</Label>
          <div className="text-sm font-medium">admin</div>
        </div>
        <div className="space-y-2">
          <Label htmlFor="newPassword">{i18next.t("user:New Password")}</Label>
          <PasswordInput
            id="newPassword"
            autoFocus
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
