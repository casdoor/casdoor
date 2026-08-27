import * as React from "react";
import i18next from "i18next";
import {Link, useLocation, useNavigate, useParams} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {Checkbox} from "@/components/ui/checkbox";
import {Input} from "@/components/ui/input";
import {Label} from "@/components/ui/label";
import {Tabs, TabsList, TabsTrigger} from "@/components/ui/tabs";
import {Alert, AlertDescription} from "@/components/ui/alert";
import {Loading} from "@/components/common/Loading";
import {AuthLayout} from "@/components/auth/AuthLayout";
import {MfaVerify, NextMfa, RequiredMfa} from "@/components/auth/MfaVerify";
import {ProviderButtons} from "@/components/auth/ProviderButtons";
import {RedirectForm} from "@/components/auth/RedirectForm";
import {SendCodeInput} from "@/components/auth/SendCodeInput";
import {useAccount} from "@/hooks/use-account";
import {authConfig} from "@/auth/Auth";
import * as Util from "@/auth/Util";
import * as Obfuscator from "@/auth/Obfuscator";
import * as ApplicationBackend from "@/backend/ApplicationBackend";
import * as AuthBackend from "@/backend/AuthBackend";
import * as OrganizationBackend from "@/backend/OrganizationBackend";
import * as Setting from "@/lib/setting";

type LoginType = "login" | "code" | "cas" | "saml" | "device";
type LoginMethod = "password" | "verificationCode" | "verificationCodeEmail" | "verificationCodePhone" | "ldap";

function getDefaultLoginMethod(application: any): LoginMethod {
  const first = application?.signinMethods?.[0];
  switch (first?.name) {
  case "Password":
    return "password";
  case "Verification code":
    switch (first?.rule) {
    case "All":
      return "verificationCode";
    case "Email only":
      return "verificationCodeEmail";
    case "Phone only":
      return "verificationCodePhone";
    }
    break;
  case "LDAP":
    return "ldap";
  }
  return "password";
}

function getSigninMethodName(method: LoginMethod) {
  if (method === "password") {
    return "Password";
  }
  if (method.startsWith("verificationCode")) {
    return "Verification code";
  }
  if (method === "ldap") {
    return "LDAP";
  }
  return "Password";
}

export default function LoginPage({type = "login"}: {type?: LoginType}) {
  const params = useParams();
  const location = useLocation();
  const navigate = useNavigate();
  const {account, reload} = useAccount();

  const [application, setApplication] = React.useState<any>(undefined);
  const [msg, setMsg] = React.useState<string | null>(null);
  const [loginMethod, setLoginMethod] = React.useState<LoginMethod | undefined>(undefined);
  const [username, setUsername] = React.useState("");
  const [password, setPassword] = React.useState("");
  const [code, setCode] = React.useState("");
  const [autoSignin, setAutoSignin] = React.useState(true);
  const [loading, setLoading] = React.useState(false);
  const [mfa, setMfa] = React.useState<{props: any; values: any; authParams: any} | null>(null);
  const [saml, setSaml] = React.useState<{response: string; redirectUrl: string; relayState: string} | null>(null);

  const owner = params.owner;
  const applicationName = params.applicationName ?? authConfig.appName;

  React.useEffect(() => {
    let cancelled = false;
    const onLoaded = (app: any) => {
      if (cancelled) {
        return;
      }
      setApplication(app);
      setLoginMethod(getDefaultLoginMethod(app));
    };

    if (type === "code" || type === "cas" || type === "device") {
      const loginParams =
        type === "cas"
          ? Util.getCasLoginParameters("admin", params.casApplicationName)
          : type === "device"
            ? {userCode: params.userCode, type}
            : Util.getOAuthGetParameters();
      AuthBackend.getApplicationLogin(loginParams)
        .then((res: any) => {
          if (res.status === "ok") {
            onLoaded(res.data);
          } else {
            setApplication(null);
            setMsg(res.msg);
          }
        })
        .catch((error) => {
          setApplication(null);
          setMsg(`${i18next.t("general:Failed to connect to server")}: ${error}`);
        });
    } else if (owner && type !== "saml") {
      OrganizationBackend.getDefaultApplication("admin", owner)
        .then((res: any) => {
          if (res.status === "ok") {
            onLoaded(res.data);
          } else {
            setApplication(null);
            setMsg(res.msg);
          }
        })
        .catch((error) => {
          setApplication(null);
          setMsg(`${i18next.t("general:Failed to connect to server")}: ${error}`);
        });
    } else {
      ApplicationBackend.getApplication("admin", params.applicationName ?? applicationName)
        .then((res: any) => {
          if (res.status === "ok") {
            onLoaded(res.data);
          } else {
            setApplication(null);
            setMsg(res.msg);
          }
        })
        .catch((error) => {
          setApplication(null);
          setMsg(`${i18next.t("general:Failed to connect to server")}: ${error}`);
        });
    }
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [type, owner, params.applicationName, params.casApplicationName, params.userCode]);

  // Already signed in on a plain /login: go to the console.
  React.useEffect(() => {
    if (type === "login" && account && !location.search.includes("silentSignin")) {
      navigate("/", {replace: true});
    }
  }, [account, type, navigate, location.search]);

  const postCodeLoginAction = (res: any) => {
    const oAuthParams = Util.getOAuthGetParameters();
    const codeValue = res.data;
    const concatChar = oAuthParams?.redirectUri?.includes("?") ? "&" : "?";
    const redirectUrl = `${oAuthParams.redirectUri}${concatChar}code=${codeValue}&state=${oAuthParams.state}`;

    if (res.data === Setting.RequiredUpdatePassword) {
      Setting.goToUpdatePassword();
      return;
    }
    // The backend asks for an explicit consent before handing out the code.
    if (res.data?.required === true) {
      reload().then(() => navigate(`/consent/${application.name}?${window.location.search.substring(1)}`));
      return;
    }
    if (Setting.hasPromptPage(application)) {
      AuthBackend.getAccount().then((accountRes: any) => {
        if (accountRes.status === "ok") {
          const nextAccount = accountRes.data;
          nextAccount.organization = accountRes.data2;
          if (Setting.isPromptAnswered(nextAccount, application)) {
            Setting.goToLink(redirectUrl);
          } else {
            navigate(
              `/prompt/${application.name}?redirectUri=${oAuthParams.redirectUri}&code=${codeValue}&state=${oAuthParams.state}`,
            );
          }
        } else {
          Setting.showMessage("error", `${i18next.t("application:Failed to sign in")}: ${accountRes.msg}`);
        }
      });
      return;
    }
    if (oAuthParams.noRedirect === "true") {
      window.close();
      window.open(redirectUrl);
      return;
    }
    Setting.goToLink(redirectUrl);
  };

  const handleLoginResult = (res: any, values: any, authParams: any) => {
    const responseType = values["type"];
    const responseTypes = String(responseType).split(" ");
    const responseMode = authParams?.responseMode || "query";

    if (responseType === "login") {
      Setting.showMessage("success", i18next.t("application:Logged in successfully"));
      reload().then(() => navigate(Setting.getFromLink()));
    } else if (responseType === "code") {
      postCodeLoginAction(res);
    } else if (responseTypes.includes("token") || responseTypes.includes("id_token")) {
      const amendatoryResponseType = responseType === "token" ? "access_token" : responseType;
      if (responseMode === "form_post") {
        Setting.createFormAndSubmit(authParams?.redirectUri, {
          token: responseTypes.includes("token") ? res.data : null,
          id_token: responseTypes.includes("id_token") ? res.data : null,
          token_type: "bearer",
          state: authParams?.state,
        });
      } else {
        Setting.goToLink(
          `${authParams.redirectUri}#${amendatoryResponseType}=${res.data}&state=${authParams.state}&token_type=bearer`,
        );
      }
    } else if (responseType === "saml") {
      if (res.data2?.method === "POST") {
        setSaml({response: res.data, redirectUrl: res.data2.redirectUrl, relayState: values["relayState"] ?? ""});
      } else {
        const redirectUri = res.data2.redirectUrl;
        Setting.goToLink(
          `${redirectUri}${redirectUri.includes("?") ? "&" : "?"}SAMLResponse=${encodeURIComponent(
            res.data,
          )}&RelayState=${encodeURIComponent(values["relayState"] ?? "")}`,
        );
      }
    }
  };

  const handleCasLoginResult = (res: any, casParams: any) => {
    let message = "Logged in successfully. ";
    if (casParams.service === "") {
      message += "Now you can visit apps protected by Casdoor.";
    }
    Setting.showMessage("success", message);
    if (casParams.service !== "") {
      const newUrl = new URL(casParams.service);
      newUrl.searchParams.append("ticket", res.data);
      window.location.href = newUrl.toString();
    }
  };

  const checkMfa = (res: any, values: any, authParams: any, onDone: (res: any) => void) => {
    if (res.data === Setting.RequiredUpdatePassword) {
      Setting.goToUpdatePassword();
    } else if (res.data === RequiredMfa) {
      reload().then(() => navigate("/mfa/setup"));
    } else if (res.data === NextMfa) {
      setMfa({props: res.data2?.[0] ?? res.data2, values: {...values, providerBack: values.provider, provider: ""}, authParams});
    } else if (res.data === "SelectPlan") {
      const pricing = res.data2;
      Setting.goToLink(`/select-plan/${pricing.owner}/${pricing.name}?user=${values.username}`);
    } else if (res.data === "BuyPlanResult") {
      const sub = res.data2;
      Setting.goToLink(`/buy-plan/${sub.owner}/${sub.pricing}/result?subscription=${sub.name}`);
    } else {
      onDone(res);
    }
  };

  const buildValues = () => {
    const values: Record<string, any> = {
      application: application.name,
      organization: application.organization,
      username,
      autoSignin,
      signinMethod: getSigninMethodName(loginMethod ?? "password"),
      language: Setting.getLanguage(),
    };

    if (loginMethod === "password" || loginMethod === "ldap") {
      const organization = application.organizationObj;
      const [cipher, errorMessage] = Obfuscator.encryptByPasswordObfuscator(
        organization?.passwordObfuscatorType,
        organization?.passwordObfuscatorKey,
        password,
      );
      if (errorMessage.length > 0) {
        Setting.showMessage("error", errorMessage);
        return null;
      }
      values.password = cipher;
      values.loginMethod = loginMethod;
    } else {
      values.code = code;
      values.username = username;
      values.password = "";
    }

    if (type === "cas") {
      values.type = type;
      return values;
    }

    const oAuthParams = Util.getOAuthGetParameters();
    values.type = oAuthParams?.responseType ?? (type === "device" ? "device" : "login");
    if (params.userCode) {
      values.userCode = params.userCode;
    }
    if (oAuthParams?.samlRequest) {
      values.samlRequest = oAuthParams.samlRequest;
      values.type = "saml";
      values.relayState = oAuthParams.relayState;
    } else if (type === "saml") {
      values.samlRequest = "";
      values.type = "saml";
      values.relayState = Util.getRelayState();
    }
    return values;
  };

  const submit = (e?: React.FormEvent) => {
    e?.preventDefault();
    const values = buildValues();
    if (values === null) {
      return;
    }
    setLoading(true);
    Setting.setSigninLanguage(Setting.getLanguage());

    if (type === "cas") {
      const casParams = Util.getCasParameters();
      AuthBackend.loginCas(values, casParams)
        .then((res: any) => {
          if (res.status === "ok") {
            checkMfa(res, values, casParams, (ok) => handleCasLoginResult(ok, casParams));
          } else {
            Setting.showMessage("error", `${i18next.t("application:Failed to sign in")}: ${res.msg}`);
          }
        })
        .finally(() => setLoading(false));
      return;
    }

    const oAuthParams = Util.getOAuthGetParameters();
    AuthBackend.login(values, oAuthParams)
      .then((res: any) => {
        if (res.status === "ok") {
          checkMfa(res, values, oAuthParams, (ok) => handleLoginResult(ok, values, oAuthParams));
        } else {
          Setting.showMessage("error", `${i18next.t("application:Failed to sign in")}: ${res.msg}`);
        }
      })
      .finally(() => {
        localStorage.setItem("lastLoginOrg", values?.organization || "");
        setLoading(false);
      });
  };

  if (saml !== null) {
    return <RedirectForm samlResponse={saml.response} redirectUrl={saml.redirectUrl} relayState={saml.relayState} />;
  }

  if (application === undefined) {
    return <Loading className="min-h-screen" />;
  }

  if (application === null) {
    return (
      <AuthLayout>
        <Alert variant="destructive">
          <AlertDescription>{msg ?? i18next.t("application:Failed to sign in")}</AlertDescription>
        </Alert>
      </AuthLayout>
    );
  }

  if (mfa !== null) {
    return (
      <AuthLayout application={application}>
        <MfaVerify
          formValues={mfa.values}
          authParams={mfa.authParams}
          mfaProps={mfa.props}
          application={application}
          onSuccess={(res) =>
            type === "cas"
              ? handleCasLoginResult(res, mfa.authParams)
              : handleLoginResult(res, mfa.values, mfa.authParams)
          }
        />
      </AuthLayout>
    );
  }

  if (application.disableSignin || application.organizationObj?.disableSignin) {
    return (
      <AuthLayout application={application}>
        <Alert variant="warning">
          <AlertDescription>{i18next.t("application:Disable signin")}</AlertDescription>
        </Alert>
      </AuthLayout>
    );
  }

  const passwordEnabled = Setting.isPasswordEnabled(application);
  const codeEnabled = Setting.isCodeSigninEnabled(application);
  const ldapEnabled = Setting.isLdapEnabled(application);
  const showTabs = [passwordEnabled, codeEnabled, ldapEnabled].filter(Boolean).length > 1;
  const isCodeMethod = (loginMethod ?? "").startsWith("verificationCode");

  return (
    <AuthLayout application={application}>
      <div className="space-y-5">
        <h1 className="text-center text-xl font-semibold">
          {application.displayName || i18next.t("login:Sign In")}
        </h1>

        {showTabs ? (
          <Tabs value={loginMethod} onValueChange={(v) => setLoginMethod(v as LoginMethod)}>
            <TabsList className="grid w-full" style={{gridTemplateColumns: `repeat(${[passwordEnabled, codeEnabled, ldapEnabled].filter(Boolean).length}, minmax(0, 1fr))`}}>
              {passwordEnabled ? <TabsTrigger value="password">{i18next.t("general:Password")}</TabsTrigger> : null}
              {codeEnabled ? (
                <TabsTrigger value="verificationCode">{i18next.t("login:Verification code")}</TabsTrigger>
              ) : null}
              {ldapEnabled ? <TabsTrigger value="ldap">LDAP</TabsTrigger> : null}
            </TabsList>
          </Tabs>
        ) : null}

        <form className="space-y-4" onSubmit={submit}>
          <div className="space-y-2">
            <Label htmlFor="username">
              {isCodeMethod ? i18next.t("login:Email or phone") : i18next.t("signup:Username")}
            </Label>
            <Input
              id="username"
              autoFocus
              autoComplete="username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
            />
          </div>

          {isCodeMethod ? (
            <div className="space-y-2">
              <Label>{i18next.t("login:Verification code")}</Label>
              <SendCodeInput
                value={code}
                onChange={setCode}
                method={username.includes("@") ? "email" : "phone"}
                dest={username}
                type="login"
                applicationId={`${application.owner}/${application.name}`}
              />
            </div>
          ) : (
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label htmlFor="password">{i18next.t("general:Password")}</Label>
                <Link
                  to={`/forget/${application.name}`}
                  className="text-xs text-muted-foreground underline-offset-4 hover:underline"
                >
                  {i18next.t("login:Forgot password?")}
                </Link>
              </div>
              <Input
                id="password"
                type="password"
                autoComplete="current-password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            </div>
          )}

          <div className="flex items-center gap-2">
            <Checkbox id="autoSignin" checked={autoSignin} onCheckedChange={(v) => setAutoSignin(v === true)} />
            <Label htmlFor="autoSignin" className="text-sm font-normal">
              {i18next.t("login:Auto sign in")}
            </Label>
          </div>

          <Button type="submit" className="w-full" loading={loading}>
            {i18next.t("login:Sign In")}
          </Button>
        </form>

        {application.enableSignUp ? (
          <p className="text-center text-sm text-muted-foreground">
            {i18next.t("login:No account?")}{" "}
            <Link to={`/signup/${application.name}`} className="text-foreground underline-offset-4 hover:underline">
              {i18next.t("login:sign up now")}
            </Link>
          </p>
        ) : null}

        <ProviderButtons application={application} method="signin" />
      </div>
    </AuthLayout>
  );
}
