import * as React from "react";
import i18next from "i18next";
import {Link, useLocation, useNavigate, useParams} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {Checkbox} from "@/components/ui/checkbox";
import {Input} from "@/components/ui/input";
import {Label} from "@/components/ui/label";
import {Avatar, AvatarFallback, AvatarImage} from "@/components/ui/avatar";
import {Alert, AlertDescription} from "@/components/ui/alert";
import {Loading} from "@/components/common/Loading";
import {AuthLayout} from "@/components/auth/AuthLayout";
import {SigninMethodTabs} from "@/components/auth/SigninMethodTabs";
import {MfaVerify, NextMfa, RequiredMfa} from "@/components/auth/MfaVerify";
import {AgreementCheckbox, getAgreementDefaultValue, isAgreementRequired} from "@/components/auth/AgreementModal";
import {DeviceLoginPanel} from "@/components/auth/DeviceLoginPanel";
import {GoogleOneTap} from "@/components/auth/GoogleOneTap";
import {FaceRecognitionCommonModal} from "@/components/common/FaceRecognitionCommonModal";
import {FaceRecognitionModal} from "@/components/common/FaceRecognitionModal";
import {ProviderButtons} from "@/components/auth/ProviderButtons";
import {WeChatLoginPanel} from "@/components/auth/WeChatLoginPanel";
import {OrganizationSelect} from "@/components/common/OrganizationSelect";
import {RedirectForm} from "@/components/auth/RedirectForm";
import {SendCodeInput, type CaptchaValues} from "@/components/auth/SendCodeInput";
import {CaptchaModal, type CaptchaHandle} from "@/components/common/CaptchaModal";
import {getCaptchaProvider} from "@/lib/captcha";
import {useAccount} from "@/hooks/use-account";
import {authConfig} from "@/auth/Auth";
import * as Util from "@/auth/Util";
import {signInWithWebAuthn} from "@/auth/webauthn";
import * as Obfuscator from "@/auth/Obfuscator";
import * as ApplicationBackend from "@/backend/ApplicationBackend";
import * as AuthBackend from "@/backend/AuthBackend";
import * as OrganizationBackend from "@/backend/OrganizationBackend";
import * as Setting from "@/lib/setting";

type LoginType = "login" | "code" | "cas" | "saml" | "device";
type LoginMethod = "password" | "verificationCode" | "verificationCodeEmail" | "verificationCodePhone" | "ldap" | "webAuthn" | "wechat" | "faceId";

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
  case "WebAuthn":
    return "webAuthn";
  case "Face ID":
    return "faceId";
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
  if (method === "webAuthn") {
    return "WebAuthn";
  }
  if (method === "faceId") {
    return "Face ID";
  }
  return "Password";
}

/**
 * Port of the antd `renderOrganizationChoiceBox`: before the sign-in form an
 * application may ask which organization the visitor belongs to, either by
 * picking one or by typing its name.
 */
function OrganizationChoiceBox({mode}: {mode: string}) {
  const [name, setName] = React.useState("");
  const go = (organization: string) => {
    if (!organization) {
      Setting.showMessage("error", i18next.t("login:Please input your organization name!"));
      return;
    }
    Setting.goToLink(`/login/${organization}?orgChoiceMode=None`);
  };

  if (mode === "Select") {
    return (
      <div className="space-y-3">
        <p className="text-base">{i18next.t("login:Please select an organization to sign in")}</p>
        <OrganizationSelect value="" onChange={go} />
      </div>
    );
  }

  return (
    <form
      className="space-y-3"
      onSubmit={(e) => {
        e.preventDefault();
        go(name.trim());
      }}
    >
      <p className="text-base">{i18next.t("login:Please type an organization to sign in")}</p>
      <Input autoFocus value={name} onChange={(e) => setName(e.target.value)} />
      <Button type="submit" className="w-full">
        {i18next.t("general:Confirm")}
      </Button>
    </form>
  );
}

/**
 * "Forgot password?" link. The application may point it at a page of its own via
 * `forgetUrl`, and the sign-in URL is remembered so the OAuth flow can resume
 * once the password has been reset.
 */
function ForgetLink({application}: {application: any}) {
  const url = Setting.getForgetLink(application);
  const className = "text-xs text-muted-foreground underline-offset-4 hover:underline";
  const text = i18next.t("login:Forgot password?");

  if (url?.startsWith("/")) {
    return <Link to={url} onClick={Setting.storeSigninUrl} className={className}>{text}</Link>;
  }
  if (url?.startsWith("http")) {
    return <a href={url} onClick={Setting.storeSigninUrl} className={className}>{text}</a>;
  }
  return null;
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
  const [captchaVisible, setCaptchaVisible] = React.useState(false);
  const [pendingValues, setPendingValues] = React.useState<any>(null);
  const [agreed, setAgreed] = React.useState(false);
  const [faceValues, setFaceValues] = React.useState<any>(null);
  const [captchaValues, setCaptchaValues] = React.useState<CaptchaValues | undefined>(undefined);
  const captchaRef = React.useRef<CaptchaHandle | null>(null);
  const [saml, setSaml] = React.useState<{response: string; redirectUrl: string; relayState: string} | null>(null);

  const owner = params.owner;
  const applicationName = params.applicationName ?? authConfig.appName;

  // remember where the sign-in started, for the flows that have to come back to it
  React.useEffect(() => {
    localStorage.setItem("signinUrl", location.pathname + location.search);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  React.useEffect(() => {
    let cancelled = false;
    const onLoaded = (app: any) => {
      if (cancelled) {
        return;
      }
      setApplication(app);
      setLoginMethod(getDefaultLoginMethod(app));
      setAgreed(getAgreementDefaultValue(app));
      setAutoSignin(Setting.getAutoSigninDefaultValue(app));
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
      navigate("/", {replace: true, state: {from: "/login"}});
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
      reload().then(() => navigate(Setting.getFromLink(), {state: {from: "/login"}}));
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
      localStorage.setItem("mfaRedirectUrl", window.location.href);
      reload().then(() => navigate("/mfa/setup", {state: {from: "/login"}}));
    } else if (res.data === NextMfa) {
      // hand over every enabled factor: MfaVerify starts on the preferred one and lets the user switch
      setMfa({props: res.data2, values: {...values, providerBack: values.provider, provider: ""}, authParams});
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

  /** Stamps the payload with what the current request is: CAS, SAML, OAuth or a plain login. */
  const applyRequestType = (values: Record<string, any>) => {
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

  /**
   * "Continue with <current account>": the session cookie already identifies the
   * user, so the payload carries no credentials — the antd page posts the same.
   */
  const loginAsCurrentAccount = () => {
    doLogin(applyRequestType({
      application: application.name,
      organization: application.organization,
      language: Setting.getLanguage(),
    }));
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
    } else if (loginMethod !== "webAuthn" && loginMethod !== "faceId") {
      values.code = code;
      values.username = username;
      values.password = "";
    }

    return applyRequestType(values);
  };

  const refreshInlineCaptcha = () => captchaRef.current?.loadCaptcha();

  const doLogin = (values: any) => {
    setLoading(true);
    Setting.setSigninLanguage(Setting.getLanguage());

    // a widget captcha is single-use, so a rejected sign-in needs a fresh one
    const usedInlineCaptcha = Setting.isInlineCaptchaEnabled(application) && captchaValues !== undefined;
    const onRejected = () => {
      if (usedInlineCaptcha && !String(values.signinMethod).includes("Verification code")) {
        refreshInlineCaptcha();
      }
    };

    if (type === "cas") {
      const casParams = Util.getCasParameters();
      AuthBackend.loginCas(values, casParams)
        .then((res: any) => {
          if (res.status === "ok") {
            checkMfa(res, values, casParams, (ok) => handleCasLoginResult(ok, casParams));
          } else {
            Setting.showMessage("error", `${i18next.t("application:Failed to sign in")}: ${res.msg}`);
            onRejected();
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
          onRejected();
        }
      })
      .finally(() => {
        localStorage.setItem("lastLoginOrg", values?.organization || "");
        setLoading(false);
      });
  };

  /** the user rejected this device-code sign-in from the approval page */
  const [deviceCanceled, setDeviceCanceled] = React.useState(false);

  const cancelDeviceLogin = () => {
    const cancelToken = new URLSearchParams(location.search).get("cancelToken") || "";
    AuthBackend.cancelDeviceLogin(params.userCode, cancelToken)
      .then((res: any) => {
        if (res.status === "ok") {
          setDeviceCanceled(true);
        } else {
          Setting.showMessage("error", res.msg || i18next.t("general:Failed to cancel"));
        }
      })
      .catch(() => Setting.showMessage("error", i18next.t("general:Failed to cancel")));
  };

  /** another signed-in device approved the device code, finish the OAuth flow here */
  const completeDeviceLogin = (deviceCode: string) => {
    const oAuthParams = Util.getOAuthGetParameters();
    AuthBackend.completeDeviceLogin(deviceCode, oAuthParams)
      .then((res: any) => {
        if (res.status === "ok") {
          handleLoginResult(res, {type: oAuthParams?.responseType ?? "login"}, oAuthParams);
        } else {
          Setting.showMessage("error", `${i18next.t("application:Failed to sign in")}: ${res.msg}`);
        }
      })
      .catch((err: any) => Setting.showMessage("error", err.message));
  };

  const doWebAuthnLogin = (values: any) => {
    const oAuthParams = Util.getOAuthGetParameters();
    setLoading(true);
    signInWithWebAuthn(application, username, values, oAuthParams)
      .then((res: any) => {
        if (res?.status === "ok") {
          handleLoginResult(res, values, oAuthParams);
        } else {
          Setting.showMessage("error", res?.msg);
        }
      })
      .catch((error: any) => {
        Setting.showMessage("error", error.message);
      })
      .finally(() => setLoading(false));
  };

  const submit = (e?: React.FormEvent) => {
    e?.preventDefault();
    if (isAgreementRequired(application) && !agreed) {
      Setting.showMessage("error", i18next.t("signup:Please accept the agreement!"));
      return;
    }
    const values = buildValues();
    if (values === null) {
      return;
    }

    if (loginMethod === "webAuthn") {
      doWebAuthnLogin(values);
      return;
    }

    if (loginMethod === "faceId") {
      // the backend checks the user exists and picks the provider before the camera opens
      setLoading(true);
      fetch(`${Setting.ServerUrl}/api/faceid-signin-begin?owner=${application.organization}&name=${encodeURIComponent(username)}`, {
        method: "GET",
        credentials: "include",
        headers: {"Accept-Language": Setting.getAcceptLanguage()},
      })
        .then((res) => res.json())
        .then((res: any) => {
          if (res.status === "error") {
            setLoading(false);
            Setting.showMessage("error", res.msg);
            return;
          }
          setFaceValues(values);
        });
      return;
    }

    // Password/LDAP sign-in may need a captcha first; the code flow is already
    // rate-limited by the code itself.
    if (loginMethod === "password" || loginMethod === "ldap") {
      const captchaRule = Setting.getCaptchaRule(application);
      if (Setting.isInlineCaptchaEnabled(application)) {
        if (captchaRule === Setting.CaptchaRule.Always && !captchaValues?.captchaToken) {
          Setting.showMessage("error", i18next.t("general:Please complete the captcha correctly"));
          return;
        }
        values.captchaType = captchaValues?.captchaType;
        values.captchaToken = captchaValues?.captchaToken;
        values.clientSecret = captchaValues?.clientSecret;
      } else if (captchaRule === Setting.CaptchaRule.Always) {
        setPendingValues(values);
        setCaptchaVisible(true);
        return;
      } else if (
        captchaRule === Setting.CaptchaRule.Dynamic ||
        captchaRule === Setting.CaptchaRule.InternetOnly
      ) {
        AuthBackend.getCaptchaStatus(values)
          .then((res: any) => {
            if (res.status === "ok" && res.data) {
              setPendingValues(values);
              setCaptchaVisible(true);
            } else {
              doLogin(values);
            }
          })
          .catch(() => doLogin(values));
        return;
      }
    }

    doLogin(values);
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

  if (deviceCanceled) {
    return (
      <AuthLayout application={application}>
        <Alert variant="warning">
          <AlertDescription>{i18next.t("login:Device login was canceled")}</AlertDescription>
        </Alert>
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

  const captchaProvider = getCaptchaProvider(application);
  const passwordEnabled = Setting.isPasswordEnabled(application);
  const codeEnabled = Setting.isCodeSigninEnabled(application);
  const ldapEnabled = Setting.isLdapEnabled(application);
  const webAuthnEnabled = Setting.isWebAuthnEnabled(application);
  const faceIdEnabled = Setting.isFaceIdEnabled(application);
  // an application with a Face ID provider lets the backend do the recognition
  const hasFaceIdProvider = (application.providers ?? []).some((item: any) => item.provider?.category === "Face ID");
  const wechatEnabled = (application.providers ?? []).some(
    (item: any) => item.provider?.type === "WeChat" && Setting.isProviderVisibleForSignIn(item),
  );
  const isCodeMethod = (loginMethod ?? "").startsWith("verificationCode");
  // "Email only" and "Phone only" share the "Verification code" tab
  const activeTab = isCodeMethod ? "verificationCode" : loginMethod;
  const methods = [
    passwordEnabled ? {value: "password", label: i18next.t("general:Password")} : null,
    codeEnabled ? {value: "verificationCode", label: i18next.t("login:Verification code")} : null,
    ldapEnabled ? {value: "ldap", label: i18next.t("login:LDAP")} : null,
    webAuthnEnabled ? {value: "webAuthn", label: i18next.t("login:WebAuthn")} : null,
    faceIdEnabled ? {value: "faceId", label: i18next.t("login:Face ID")} : null,
    wechatEnabled ? {value: "wechat", label: i18next.t("login:WeChat")} : null,
  ].filter(Boolean) as {value: LoginMethod; label: string}[];
  const showTabs = methods.length > 1;
  // each block of the form can be hidden from "Signin items" in the application
  const isVisible = (name: string) => Setting.isSigninItemVisible(application, name);
  // the QR panels replace the credential form entirely
  const isPanelMethod = loginMethod === "wechat";
  // device login is offered next to the form when the application asks for it
  const deviceLoginOnLoginPage = type !== "device" && (application.signinMethods ?? []).some(
    (item: any) => item.name === "Device login" && item.rule === "Login page",
  );

  // An application can ask the visitor which organization they belong to before
  // showing the form. Choosing one reloads /login/<org> with the box turned off,
  // exactly as the antd page does.
  const orgChoiceMode = new URLSearchParams(location.search).get("orgChoiceMode") === "None"
    ? "None"
    : application.orgChoiceMode;
  if (type === "login" && (orgChoiceMode === "Select" || orgChoiceMode === "Input")) {
    return (
      <AuthLayout application={application}>
        <OrganizationChoiceBox mode={orgChoiceMode} />
      </AuthLayout>
    );
  }

  // Already signed in to this organization: offer the one-click path before the
  // form, which is how an OAuth or device request gets approved.
  const showSignedInBox =
    !!account && account.owner === application.organization && !(type === "device" && deviceCanceled);

  return (
    <AuthLayout application={application}>
      <div className="space-y-5">
        <h1 className="text-center text-xl font-semibold">
          {application.displayName || i18next.t("login:Sign In")}
        </h1>

        {type === "device" && params.userCode ? (
          <div className="space-y-1 text-center">
            <h2 className="text-base font-semibold">{i18next.t("login:Approve sign-in on this device")}</h2>
            <div className="font-medium">{application.displayName}</div>
            <div className="text-sm text-muted-foreground">
              {i18next.t("login:Confirmation code")}: {params.userCode}
            </div>
          </div>
        ) : null}

        {showSignedInBox ? (
          <div className="space-y-3">
            <div className="text-sm">
              {type === "device"
                ? i18next.t("login:Continue with your current account to approve this sign-in")
                : i18next.t("login:Continue with")}
              &nbsp;:
            </div>
            <Button
              type="button"
              variant="outline"
              className="h-12 w-full justify-center gap-3"
              onClick={loginAsCurrentAccount}
            >
              <Avatar className="h-8 w-8">
                {Setting.getEffectiveAvatarUrl(account) ? (
                  <AvatarImage src={Setting.getEffectiveAvatarUrl(account)} alt={account.name} />
                ) : null}
                <AvatarFallback style={{backgroundColor: Setting.getAvatarColor(account.name), color: "#fff"}}>
                  {(account.name || "?").charAt(0).toUpperCase()}
                </AvatarFallback>
              </Avatar>
              {account.displayName ? `${account.name} (${account.displayName})` : account.name}
            </Button>
            <div className="text-sm">{i18next.t("login:Or sign in with another account")}&nbsp;:</div>
          </div>
        ) : null}

        {showTabs && isVisible("Signin methods") ? (
          <SigninMethodTabs
            methods={methods}
            value={activeTab}
            onChange={(v) => setLoginMethod(v as LoginMethod)}
          />
        ) : null}

        {isPanelMethod ? <WeChatLoginPanel application={application} /> : (
          <form className="space-y-4" onSubmit={submit}>
            {isVisible("Username") ? (
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
            ) : null}

            {isCodeMethod && isVisible("Verification code") ? (
              <div className="space-y-2">
                <Label>{i18next.t("login:Verification code")}</Label>
                <SendCodeInput
                  value={code}
                  onChange={setCode}
                  method="login"
                  destType={username.includes("@") ? "email" : "phone"}
                  dest={username}
                  application={application}
                  applicationId={Setting.getApplicationName(application)}
                  useInlineCaptcha={Setting.isInlineCaptchaEnabled(application)}
                  captchaValue={captchaValues}
                  refreshCaptcha={refreshInlineCaptcha}
                />
              </div>
            ) : null}

            {!isCodeMethod && loginMethod !== "webAuthn" && loginMethod !== "faceId" && isVisible("Password") ? (
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <Label htmlFor="password">{i18next.t("general:Password")}</Label>
                  {isVisible("Forgot password?") ? <ForgetLink application={application} /> : null}
                </div>
                <Input
                  id="password"
                  type="password"
                  autoComplete="current-password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                />
              </div>
            ) : null}

            {captchaProvider && Setting.isInlineCaptchaEnabled(application) ? (
              <CaptchaModal
                noModal
                owner={captchaProvider.owner}
                name={captchaProvider.name}
                isCurrentProvider
                innerRef={captchaRef}
                onUpdateToken={(captchaType, captchaToken, clientSecret) =>
                  setCaptchaValues({captchaType, captchaToken, clientSecret})
                }
              />
            ) : null}

            {isVisible("Auto sign in") ? (
              <div className="flex items-center gap-2">
                <Checkbox id="autoSignin" checked={autoSignin} onCheckedChange={(v) => setAutoSignin(v === true)} />
                <Label htmlFor="autoSignin" className="text-sm font-normal">
                  {i18next.t("login:Auto sign in")}
                </Label>
              </div>
            ) : null}

            {application.termsOfUse && isVisible("Agreement") ? (
              <AgreementCheckbox application={application} checked={agreed} onChange={setAgreed} />
            ) : null}

            {isVisible("Login button") ? (
              <Button type="submit" className="w-full" loading={loading}>
                {loginMethod === "webAuthn"
                  ? i18next.t("login:Sign in with WebAuthn")
                  : loginMethod === "faceId"
                    ? i18next.t("login:Sign in with Face ID")
                    : type === "device"
                      ? i18next.t("login:Approve and sign in")
                      : i18next.t("login:Sign In")}
              </Button>
            ) : null}

            {type === "device" ? (
              <Button type="button" variant="outline" className="w-full" onClick={cancelDeviceLogin}>
                {i18next.t("general:Cancel")}
              </Button>
            ) : null}
          </form>
        )}

        {deviceLoginOnLoginPage ? (
          <div className="border-t pt-4">
            <DeviceLoginPanel application={application} onSuccess={completeDeviceLogin} />
          </div>
        ) : null}

        {application.enableSignUp && isVisible("Signup link") ? (
          <p className="text-center text-sm text-muted-foreground">
            {i18next.t("login:No account?")}{" "}
            <Link to={`/signup/${application.name}`} className="text-foreground underline-offset-4 hover:underline">
              {i18next.t("login:sign up now")}
            </Link>
          </p>
        ) : null}

        {isVisible("Providers") ? <ProviderButtons application={application} method="signin" /> : null}

        <GoogleOneTap application={application} />

        {faceValues !== null ? (
          hasFaceIdProvider ? (
            <FaceRecognitionCommonModal
              visible={true}
              onOk={(faceIdImage) => {
                doLogin({...faceValues, faceIdImage});
                setFaceValues(null);
              }}
              onCancel={() => {
                setFaceValues(null);
                setLoading(false);
              }}
            />
          ) : (
            <FaceRecognitionModal
              visible={true}
              onOk={(faceId) => {
                doLogin({...faceValues, faceId});
                setFaceValues(null);
              }}
              onCancel={() => {
                setFaceValues(null);
                setLoading(false);
              }}
            />
          )
        ) : null}

        {captchaProvider && !Setting.isInlineCaptchaEnabled(application) ? (
          <CaptchaModal
            owner={captchaProvider.owner}
            name={captchaProvider.name}
            visible={captchaVisible}
            isCurrentProvider
            innerRef={captchaRef}
            onOk={(captchaType, captchaToken, clientSecret) => {
              setCaptchaVisible(false);
              doLogin({...pendingValues, captchaType, captchaToken, clientSecret});
            }}
            onCancel={() => {
              setCaptchaVisible(false);
              setLoading(false);
            }}
          />
        ) : null}
      </div>
    </AuthLayout>
  );
}
