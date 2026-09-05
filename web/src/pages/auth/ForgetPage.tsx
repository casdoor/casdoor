import * as React from "react";
import i18next from "i18next";
import {useNavigate, useParams} from "react-router-dom";
import {Alert, AlertDescription} from "@/components/ui/alert";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Label} from "@/components/ui/label";
import {Loading} from "@/components/common/Loading";
import {AuthLayout} from "@/components/auth/AuthLayout";
import {SendCodeInput} from "@/components/auth/SendCodeInput";
import {PasswordRequirements, checkPasswordComplexity} from "@/lib/password-checker";
import {authConfig} from "@/auth/Auth";
import * as Obfuscator from "@/auth/Obfuscator";
import * as ApplicationBackend from "@/backend/ApplicationBackend";
import * as AuthBackend from "@/backend/AuthBackend";
import * as UserBackend from "@/backend/UserBackend";
import * as Setting from "@/lib/setting";

type Step = "account" | "verify" | "password";

/** Password recovery: find the account, verify a code, then set a new password. */
export default function ForgetPage() {
  const params = useParams();
  const navigate = useNavigate();
  const applicationName = params.applicationName ?? authConfig.appName;

  // A reset link mails the code in the query string, which skips straight to the
  // last step; the code is then verified together with the new password.
  const queryParams = React.useMemo(() => new URLSearchParams(window.location.search), []);
  const linkCode = queryParams.get("code") ?? "";

  const [application, setApplication] = React.useState<any>(undefined);
  const [step, setStep] = React.useState<Step>(linkCode ? "password" : "account");
  const [username, setUsername] = React.useState(queryParams.get("username") ?? "");
  const [account, setAccount] = React.useState(queryParams.get("username") ?? "");
  const [dest, setDest] = React.useState({email: "", phone: "", countryCode: ""});
  const [method, setMethod] = React.useState<"email" | "phone">("email");
  const [code, setCode] = React.useState(linkCode);
  const [password, setPassword] = React.useState("");
  const [confirm, setConfirm] = React.useState("");
  const [passwordFocused, setPasswordFocused] = React.useState(false);
  const [loading, setLoading] = React.useState(false);

  React.useEffect(() => {
    ApplicationBackend.getApplication("admin", applicationName)
      .then((res: any) => {
        setApplication(res.status === "ok" ? res.data : null);
      })
      .catch(() => setApplication(null));
  }, [applicationName]);

  if (application === undefined) {
    return <Loading className="min-h-screen" />;
  }

  if (application === null) {
    return (
      <AuthLayout>
        <Alert variant="destructive">
          <AlertDescription>{i18next.t("forget:Unknown forget type")}</AlertDescription>
        </Alert>
      </AuthLayout>
    );
  }

  // same matching rules as GetProviderByCategoryAndRule() in the backend, for the "forget" method
  const hasProviderOfCategory = (category: string) =>
    application?.providers?.some((providerItem: any) => providerItem?.provider?.category === category &&
      ["forget", "", "All", "all", "None"].includes(providerItem.rule)) ?? false;

  // only advertise the identifiers that can be used: a lookup by email or phone is
  // a dead end when the matching provider is not configured for the application
  const accountPlaceholder = () => {
    const hasEmail = hasProviderOfCategory("Email");
    const hasPhone = hasProviderOfCategory("SMS");
    if (hasEmail && !hasPhone) {
      return i18next.t("login:username or Email");
    }
    if (!hasEmail && hasPhone) {
      return i18next.t("login:username or phone");
    }
    return i18next.t("login:username, Email or phone");
  };

  const findAccount = () => {
    setLoading(true);
    AuthBackend.getEmailAndPhone(application.organization, username)
      .then((res: any) => {
        if (res.status !== "ok") {
          Setting.showMessage("error", res.msg);
          return;
        }

        // only offer a method whose provider is configured for the application
        const email = hasProviderOfCategory("Email") ? (res.data?.email ?? "") : "";
        const phone = hasProviderOfCategory("SMS") ? (res.data?.phone ?? "") : "";
        if (!email && !phone) {
          Setting.showMessage("error", i18next.t("general:No verification method"));
          return;
        }

        setAccount(res.data?.name ?? username);
        setDest({email, phone, countryCode: res.data?.countryCode ?? ""});
        setMethod(email ? "email" : "phone");
        setStep("verify");
      })
      .finally(() => setLoading(false));
  };

  const verifyCode = () => {
    setLoading(true);
    UserBackend.verifyCode({
      application: application.name,
      organization: application.organization,
      // the backend derives the code type from "username", so it must be the destination
      username: method === "email" ? dest.email : dest.phone,
      name: account,
      code,
      type: "login",
      countryCode: dest.countryCode,
    })
      .then((res: any) => {
        if (res.status === "ok") {
          setStep("password");
        } else {
          Setting.showMessage("error", res.msg);
        }
      })
      .finally(() => setLoading(false));
  };

  const resetPassword = async() => {
    const complexityError = checkPasswordComplexity(password, application.organizationObj?.passwordOptions);
    if (complexityError !== "") {
      Setting.showMessage("error", complexityError);
      return;
    }
    if (password !== confirm) {
      Setting.showMessage("error", i18next.t("signup:Your confirmed password is inconsistent with the password!"));
      return;
    }

    setLoading(true);
    // a code that arrived through the link has not been checked yet
    if (linkCode) {
      const verifyRes: any = await UserBackend.verifyCode({
        application: application.name,
        organization: application.organization,
        username: queryParams.get("dest") ?? "",
        name: account,
        code: linkCode,
        type: "login",
      });
      if (verifyRes.status !== "ok") {
        Setting.showMessage("error", verifyRes.msg);
        setLoading(false);
        return;
      }
    }

    const organization = application.organizationObj;
    let newPassword = password;
    if (organization?.passwordObfuscatorType && organization.passwordObfuscatorType !== "Plain") {
      const [cipher, errorMessage] = Obfuscator.encryptByPasswordObfuscator(
        organization.passwordObfuscatorType,
        organization.passwordObfuscatorKey,
        password,
      );
      if (errorMessage.length > 0) {
        Setting.showMessage("error", errorMessage);
        setLoading(false);
        return;
      }
      newPassword = cipher;
    }

    UserBackend.setPassword(application.organization, account, "", newPassword, code)
      .then((res: any) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("user:Password set successfully"));
          // back to where the sign-in started, so the OAuth flow can carry on to
          // the application; without a stored URL the application's own sign-in
          // page is the destination, not Casdoor's
          const link = Setting.getStoredSigninUrl() || Setting.getLoginLink(application) || "/login";
          if (link.startsWith("/")) {
            navigate(link);
          } else {
            Setting.goToLink(link);
          }
        } else {
          Setting.showMessage("error", res.msg);
        }
      })
      .finally(() => setLoading(false));
  };

  return (
    <AuthLayout application={application}>
      <div className="space-y-5">
        <h1 className="text-center text-xl font-semibold">{i18next.t("forget:Reset password")}</h1>

        {step === "account" ? (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="account">{i18next.t("cert:Account")}</Label>
              <Input
                id="account"
                autoFocus
                placeholder={accountPlaceholder()}
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && findAccount()}
              />
            </div>
            <Button className="w-full" loading={loading} onClick={findAccount}>
              {i18next.t("forget:Next Step")}
            </Button>
          </div>
        ) : step === "verify" ? (
          <div className="space-y-4">
            <div className="flex gap-2">
              {dest.email ? (
                <Button
                  variant={method === "email" ? "default" : "outline"}
                  className="flex-1"
                  onClick={() => setMethod("email")}
                >
                  {i18next.t("general:Email")}
                </Button>
              ) : null}
              {dest.phone ? (
                <Button
                  variant={method === "phone" ? "default" : "outline"}
                  className="flex-1"
                  onClick={() => setMethod("phone")}
                >
                  {i18next.t("general:Phone")}
                </Button>
              ) : null}
            </div>
            <p className="text-sm text-muted-foreground">{method === "email" ? dest.email : dest.phone}</p>
            <SendCodeInput
              value={code}
              onChange={setCode}
              method="forget"
              destType={method}
              dest={method === "email" ? dest.email : dest.phone}
              countryCode={dest.countryCode}
              application={application}
              applicationId={Setting.getApplicationName(application)}
              checkUser={account}
            />
            <Button className="w-full" loading={loading} onClick={verifyCode}>
              {i18next.t("forget:Next Step")}
            </Button>
          </div>
        ) : (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="newPassword">{i18next.t("user:New Password")}</Label>
              <Input
                id="newPassword"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                onFocus={() => setPasswordFocused(true)}
              />
              {passwordFocused ? (
                <PasswordRequirements
                  options={application.organizationObj?.passwordOptions}
                  password={password}
                />
              ) : null}
            </div>
            <div className="space-y-2">
              <Label htmlFor="confirmPassword">{i18next.t("user:Re-enter New")}</Label>
              <Input
                id="confirmPassword"
                type="password"
                value={confirm}
                onChange={(e) => setConfirm(e.target.value)}
              />
            </div>
            <Button className="w-full" loading={loading} onClick={resetPassword}>
              {i18next.t("forget:Change Password")}
            </Button>
          </div>
        )}
      </div>
    </AuthLayout>
  );
}
