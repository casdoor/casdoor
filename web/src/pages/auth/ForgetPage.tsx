import * as React from "react";
import i18next from "i18next";
import {ArrowLeft} from "lucide-react";
import {Link, useNavigate, useParams} from "react-router-dom";
import {Alert, AlertDescription} from "@/components/ui/alert";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {PasswordInput} from "@/components/common/PasswordInput";
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
import {cn} from "@/lib/utils";

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

  const steps: {key: Step; label: string}[] = [
    {key: "account", label: i18next.t("cert:Account")},
    {key: "verify", label: i18next.t("forget:Verify")},
    {key: "password", label: i18next.t("forget:Reset")},
  ];
  const stepIndex = steps.findIndex((item) => item.key === step);

  // a code that arrived through a reset link lands on the last step with nothing
  // behind it, so there is nowhere to go back to
  const back = linkCode
    ? undefined
    : step === "verify"
      ? () => setStep("account")
      : step === "password"
        ? () => setStep("verify")
        : undefined;

  const signinLink = Setting.getLoginLink(application) || "/login";
  const backToSigninClass = "inline-flex items-center gap-1.5 font-medium text-foreground underline-offset-4 hover:underline";

  return (
    <AuthLayout
      application={application}
      onBack={back}
      footer={
        signinLink.startsWith("/") ? (
          <Link to={signinLink} className={backToSigninClass}>
            <ArrowLeft className="h-4 w-4" />
            {i18next.t("login:Sign In")}
          </Link>
        ) : (
          <a href={signinLink} className={backToSigninClass}>
            <ArrowLeft className="h-4 w-4" />
            {i18next.t("login:Sign In")}
          </a>
        )
      }
    >
      <div className="space-y-6">
        {linkCode ? null : (
          <ol className="flex items-center gap-2">
            {steps.map((item, index) => (
              <li key={item.key} className="flex min-w-0 flex-1 flex-col gap-1.5">
                <span
                  className={cn(
                    "h-1 rounded-full transition-colors",
                    index <= stepIndex ? "bg-primary" : "bg-muted",
                  )}
                />
                <span
                  className={cn(
                    "truncate text-xs",
                    index === stepIndex ? "font-medium text-foreground" : "text-muted-foreground",
                  )}
                >
                  {item.label}
                </span>
              </li>
            ))}
          </ol>
        )}

        {step === "account" ? (
          <form
            className="space-y-4"
            onSubmit={(e) => {
              e.preventDefault();
              findAccount();
            }}
          >
            <div className="space-y-2">
              <Label htmlFor="account">{i18next.t("cert:Account")}</Label>
              <Input
                id="account"
                autoFocus
                autoComplete="username"
                placeholder={accountPlaceholder()}
                value={username}
                onChange={(e) => setUsername(e.target.value)}
              />
            </div>
            <Button type="submit" className="w-full" loading={loading}>
              {i18next.t("forget:Next Step")}
            </Button>
          </form>
        ) : step === "verify" ? (
          <form
            className="space-y-4"
            onSubmit={(e) => {
              e.preventDefault();
              verifyCode();
            }}
          >
            {dest.email && dest.phone ? (
              // both are configured, so which one gets the code is a real choice
              <div className="grid grid-cols-2 gap-1 rounded-lg bg-muted p-1">
                {(["email", "phone"] as const).map((option) => (
                  <button
                    key={option}
                    type="button"
                    onClick={() => setMethod(option)}
                    className={cn(
                      "rounded-md px-3 py-1.5 text-sm font-medium transition-colors",
                      method === option
                        ? "bg-background text-foreground shadow-sm"
                        : "text-muted-foreground hover:text-foreground",
                    )}
                  >
                    {i18next.t(option === "email" ? "general:Email" : "general:Phone")}
                  </button>
                ))}
              </div>
            ) : null}

            <div className="rounded-lg border bg-muted/50 px-3 py-2.5 text-sm">
              <span className="text-muted-foreground">
                {i18next.t(method === "email" ? "general:Email" : "general:Phone")}
              </span>
              <span className="ml-2 break-all font-medium">{method === "email" ? dest.email : dest.phone}</span>
            </div>

            <div className="space-y-2">
              <Label>{i18next.t("login:Verification code")}</Label>
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
            </div>

            <Button type="submit" className="w-full" loading={loading}>
              {i18next.t("forget:Next Step")}
            </Button>
          </form>
        ) : (
          <form
            className="space-y-4"
            onSubmit={(e) => {
              e.preventDefault();
              resetPassword();
            }}
          >
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
              {passwordFocused ? (
                <PasswordRequirements
                  options={application.organizationObj?.passwordOptions}
                  password={password}
                />
              ) : null}
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
              {i18next.t("forget:Change Password")}
            </Button>
          </form>
        )}
      </div>
    </AuthLayout>
  );
}
