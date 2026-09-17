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
import {CustomHtml, CustomStyle} from "@/components/common/CustomHtml";
import {AuthLayout} from "@/components/auth/AuthLayout";
import {PasswordRequirements, checkPasswordComplexity} from "@/lib/password-checker";
import {authConfig} from "@/auth/Auth";
import * as Obfuscator from "@/auth/Obfuscator";
import * as ApplicationBackend from "@/backend/ApplicationBackend";
import * as UserBackend from "@/backend/UserBackend";
import * as Setting from "@/lib/setting";
import {useAccount} from "@/hooks/use-account";

// the order the page is built in when the application carries no "Update password items"
const DEFAULT_ITEMS = ["Logo", "Languages", "Title", "Old password", "New password", "Confirm password", "Submit button"]
  .map((name) => ({name, visible: true}));

/**
 * The page a sign-in lands on when the account is flagged "need update password":
 * the same form as the account page's password modal, rendered inside the
 * application's auth layout and built from its "Update password items", then
 * back to the sign-in that was interrupted.
 */
export default function UpdatePasswordPage() {
  const params = useParams();
  const applicationName = params.applicationName ?? authConfig.appName;
  const {account, setAccount} = useAccount();

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

  const signinUrl = Setting.getStoredSigninUrl();

  React.useEffect(() => {
    if (account === null) {
      Setting.goToLink(signinUrl || "/login");
    }
  }, [account, signinUrl]);

  if (application === undefined || !account) {
    return <Loading className="min-h-screen" />;
  }

  const organization = application?.organizationObj;
  const items: any[] = application?.updatePasswordItems?.length ? application.updatePasswordItems : DEFAULT_ITEMS;
  const isVisible = (name: string) => items.find((item) => item.name === name)?.visible !== false;

  const submit = () => {
    if (password === "" || confirm === "") {
      Setting.showMessage("error", i18next.t("user:Empty input!"));
      return;
    }
    if (password !== confirm) {
      Setting.showMessage("error", i18next.t("user:Two passwords you typed do not match."));
      return;
    }
    const complexityError = checkPasswordComplexity(password, organization?.passwordOptions);
    if (complexityError !== "") {
      Setting.showMessage("error", complexityError);
      return;
    }

    let encryptedOldPassword = oldPassword;
    let encryptedNewPassword = password;
    if (organization?.passwordObfuscatorType && organization.passwordObfuscatorType !== "Plain") {
      for (const [plain, assign] of [
        [oldPassword, (v: string) => (encryptedOldPassword = v)],
        [password, (v: string) => (encryptedNewPassword = v)],
      ] as const) {
        const [cipher, error] = Obfuscator.encryptByPasswordObfuscator(
          organization.passwordObfuscatorType,
          organization.passwordObfuscatorKey,
          plain,
        );
        if (error) {
          Setting.showMessage("error", error);
          return;
        }
        assign(cipher);
      }
    }

    setLoading(true);
    UserBackend.setPassword(account.owner, account.name, encryptedOldPassword, encryptedNewPassword)
      .then((res: any) => {
        if (res.status !== "ok") {
          Setting.showMessage("error", i18next.t(`user:${res.msg}`));
          return;
        }
        Setting.showMessage("success", i18next.t("user:Password set successfully"));
        setAccount({...account, needUpdatePassword: false});
        sessionStorage.removeItem("signinUrl");
        Setting.goToLink(signinUrl ? `${signinUrl}${signinUrl.includes("?") ? "&" : "?"}silentSignin=1` : "/");
      })
      .finally(() => setLoading(false));
  };

  const passwordField = (
    item: any,
    id: string,
    className: string,
    labelKey: string,
    value: string,
    onChange: (v: string) => void,
    extra?: React.ReactNode,
  ) => (
    <div key={item.name} className={`${className} space-y-2`}>
      <Label htmlFor={id}>{item.label || i18next.t(labelKey)}</Label>
      <PasswordInput
        id={id}
        className={`${className}-input`}
        autoFocus={id === "oldPassword"}
        autoComplete={id === "oldPassword" ? "current-password" : "new-password"}
        placeholder={item.placeholder || undefined}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onFocus={id === "newPassword" ? () => setPasswordFocused(true) : undefined}
      />
      {extra}
    </div>
  );

  const renderItem = (item: any) => {
    if (item.visible === false) {
      return null;
    }
    if (Setting.isCustomFormItem(item)) {
      return <CustomHtml key={item.name} html={item.customCss} />;
    }
    switch (item.name) {
    case "Title":
      return (
        <h1 key={item.name} className="update-password-title text-xl font-semibold tracking-tight">
          {item.label || i18next.t("user:You need to update your password")}
        </h1>
      );
    case "Old password":
      return passwordField(item, "oldPassword", "update-password-old", "user:Old Password", oldPassword, setOldPassword);
    case "New password":
      return passwordField(
        item, "newPassword", "update-password-new", "user:New Password", password, setPassword,
        passwordFocused ? <PasswordRequirements options={organization?.passwordOptions} password={password} /> : null,
      );
    case "Confirm password":
      return passwordField(item, "confirmPassword", "update-password-confirm", "user:Re-enter New", confirm, setConfirm);
    case "Submit button":
      return (
        <div key={item.name} className="update-password-button-box">
          <Button type="submit" className="update-password-button w-full" loading={loading}>
            {item.label || i18next.t("user:Set Password")}
          </Button>
        </div>
      );
    default:
      // "Logo" and "Languages" are the layout's, see hideLogo / hideLanguages below
      return null;
    }
  };

  return (
    <AuthLayout application={application} hideLogo={!isVisible("Logo")} hideLanguages={!isVisible("Languages")}>
      {items.map((item) => (Setting.isCustomFormItem(item) ? null : <CustomStyle key={`css-${item.name}`} css={item.customCss} />))}
      <form
        className="space-y-4"
        onSubmit={(e) => {
          e.preventDefault();
          submit();
        }}
      >
        {items.map(renderItem)}
      </form>
    </AuthLayout>
  );
}
