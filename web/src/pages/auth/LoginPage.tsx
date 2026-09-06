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
import {CustomHtml, CustomStyle} from "@/components/common/CustomHtml";
import {AuthDivider, AuthLayout} from "@/components/auth/AuthLayout";
import {SigninMethodTabs} from "@/components/auth/SigninMethodTabs";
import {MfaVerify, NextMfa, RequiredMfa} from "@/components/auth/MfaVerify";
import {AgreementCheckbox, getAgreementDefaultValue, isAgreementRequired} from "@/components/auth/AgreementModal";
import {DeviceLoginPanel} from "@/components/auth/DeviceLoginPanel";
import {GoogleOneTap} from "@/components/auth/GoogleOneTap";
import {FaceRecognitionCommonModal} from "@/components/common/FaceRecognitionCommonModal";
import {FaceRecognitionModal} from "@/components/common/FaceRecognitionModal";
import {ProviderButtons, getVisibleProviders} from "@/components/auth/ProviderButtons";
import {WeChatLoginPanel} from "@/components/auth/WeChatLoginPanel";
import {OrganizationSelect} from "@/components/common/OrganizationSelect";
import {RedirectForm} from "@/components/auth/RedirectForm";
import {CountryCodeSelect} from "@/components/common/CountryCodeSelect";
import {PasswordInput} from "@/components/common/PasswordInput";
import {SendCodeInput, type CaptchaValues} from "@/components/auth/SendCodeInput";
import {CaptchaModal, type CaptchaHandle} from "@/components/common/CaptchaModal";
import {getCaptchaProvider} from "@/lib/captcha";
import {useAccount} from "@/hooks/use-account";
import {authConfig} from "@/auth/Auth";
import * as Util from "@/auth/Util";
import * as Provider from "@/auth/Provider";
import {signInWithWebAuthn} from "@/auth/webauthn";
import * as Obfuscator from "@/auth/Obfuscator";
import * as ApplicationBackend from "@/backend/ApplicationBackend";
import * as AuthBackend from "@/backend/AuthBackend";
import * as OrganizationBackend from "@/backend/OrganizationBackend";
import * as Setting from "@/lib/setting";

type LoginType = "login" | "code" | "cas" | "saml" | "device";
type LoginMethod = "password" | "verificationCode" | "verificationCodeEmail" | "verificationCodePhone" | "ldap" | "webAuthn" | "wechat" | "faceId" | "device";

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
  case "Device login":
    if (first?.rule === "Tab") {
      return "device";
    }
    break;
  }
  return "password";
}

/**
 * The method strip of the antd page: the application lists its sign-in methods in
 * the order they should appear, each with its own rule and display name.
 */
const SIGNIN_METHOD_KEYS = new Map<string, LoginMethod>([
  ["Password-All", "password"],
  ["Password-Non-LDAP", "password"],
  ["Verification code-All", "verificationCode"],
  ["Verification code-Email only", "verificationCodeEmail"],
  ["Verification code-Phone only", "verificationCodePhone"],
  ["WebAuthn-None", "webAuthn"],
  ["LDAP-None", "ldap"],
  ["Face ID-None", "faceId"],
  ["Device login-Tab", "device"],
  ["WeChat-Tab", "wechat"],
  ["WeChat-None", "wechat"],
]);

function getSigninMethods(application: any, type: string) {
  const methods: {value: LoginMethod; label: string}[] = [];
  const signinMethods = (application?.signinMethods ?? []) as any[];

  signinMethods.forEach((signinMethod) => {
    if (Setting.isSigninMethodHidden(signinMethod)) {
      return;
    }
    // on the device-approval page the device tab would loop back to itself
    if (type === "device" && signinMethod.name === "Device login") {
      return;
    }
    const value = SIGNIN_METHOD_KEYS.get(`${signinMethod.name}-${signinMethod.rule}`);
    if (!value) {
      return;
    }
    let label = signinMethod.name === signinMethod.displayName ? getSigninMethodLabel(value) : signinMethod.displayName;
    if (signinMethods.length >= 4 && label === "Verification code") {
      label = "Code";
    }
    methods.push({value, label});
  });

  return methods;
}

function getSigninMethodLabel(method: LoginMethod) {
  switch (method) {
  case "password":
    return i18next.t("general:Password");
  case "ldap":
    return i18next.t("login:LDAP");
  case "webAuthn":
    return i18next.t("login:WebAuthn");
  case "faceId":
    return i18next.t("login:Face ID");
  case "device":
    return i18next.t("login:Device login");
  case "wechat":
    return i18next.t("login:WeChat");
  default:
    return i18next.t("login:Verification code");
  }
}

/** The antd page's getPlaceholder(): what the single credential field accepts. */
function getUsernameRequiredMessage(method: LoginMethod | undefined) {
  switch (method) {
  case "verificationCodeEmail":
    return i18next.t("login:Please input your Email!");
  case "verificationCodePhone":
    return i18next.t("login:Please input your Phone!");
  case "ldap":
    return i18next.t("login:Please input your LDAP username!");
  default:
    return i18next.t("login:Please input your Email or Phone!");
  }
}

function getUsernameLabel(method: LoginMethod | undefined) {
  switch (method) {
  case "verificationCode":
    return i18next.t("login:Email or phone");
  case "verificationCodeEmail":
    return i18next.t("general:Email");
  case "verificationCodePhone":
    return i18next.t("general:Phone");
  case "ldap":
    return i18next.t("login:LDAP username, Email or phone");
  default:
    return i18next.t("login:username, Email or phone");
  }
}

/** Which of the three verification-code methods the "Verification code" tab stands for. */
function getCodeLoginMethod(application: any): LoginMethod {
  const rule = (application?.signinMethods ?? []).find(
    (item: any) => item.name === "Verification code" && !Setting.isSigninMethodHidden(item),
  )?.rule;
  if (rule === "Email only") {
    return "verificationCodeEmail";
  }
  if (rule === "Phone only") {
    return "verificationCodePhone";
  }
  return "verificationCode";
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
      <div className="space-y-4">
        <p className="text-sm text-muted-foreground">{i18next.t("login:Please select an organization to sign in")}</p>
        <OrganizationSelect value="" onChange={go} />
      </div>
    );
  }

  return (
    <form
      className="space-y-4"
      onSubmit={(e) => {
        e.preventDefault();
        go(name.trim());
      }}
    >
      <p className="text-sm text-muted-foreground">{i18next.t("login:Please type an organization to sign in")}</p>
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
function ForgetLink({application, label}: {application: any; label?: string}) {
  const url = Setting.getForgetLink(application);
  const className = "text-sm font-medium text-foreground underline-offset-4 hover:underline";
  const text = label || i18next.t("login:Forgot password?");

  if (url?.startsWith("/")) {
    return <Link to={url} onClick={Setting.storeSigninUrl} className={className}>{text}</Link>;
  }
  if (url?.startsWith("http")) {
    return <a href={url} onClick={Setting.storeSigninUrl} className={className}>{text}</a>;
  }
  return null;
}

interface LoginPageProps {
  type?: LoginType;
  /**
   * Renders the page for the application editor's live preview: the application
   * comes from the form being edited instead of the route, and the flows that
   * would leave the page are turned off.
   */
  application?: any;
  preview?: string;
}

export default function LoginPage({type = "login", application: applicationProp, preview}: LoginPageProps) {
  const params = useParams();
  const location = useLocation();
  const navigate = useNavigate();
  const {account, reload} = useAccount();

  const [application, setApplication] = React.useState<any>(undefined);
  const [msg, setMsg] = React.useState<string | null>(null);
  const [loginMethod, setLoginMethod] = React.useState<LoginMethod | undefined>(undefined);
  const [username, setUsername] = React.useState("");
  const [countryCode, setCountryCode] = React.useState("");
  const [password, setPassword] = React.useState("");
  const [code, setCode] = React.useState("");
  const [signinErrors, setSigninErrors] = React.useState<Record<string, string>>({});
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
  // the device-code flow ends on the page itself: "" while it runs, then the outcome
  const [userCodeStatus, setUserCodeStatus] = React.useState<"" | "expired" | "canceled" | "success">("");

  const owner = params.owner;
  const applicationName = params.applicationName ?? authConfig.appName;

  // remember where the sign-in started, for the flows that have to come back to it
  React.useEffect(() => {
    if (preview) {
      return;
    }
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
      setCountryCode(app?.organizationObj?.countryCodes?.[0] ?? "");
      setLoginMethod(getDefaultLoginMethod(app));
      setAgreed(getAgreementDefaultValue(app));
      setAutoSignin(Setting.getAutoSigninDefaultValue(app));
    };

    // the preview is handed the application the editor is holding, live
    if (applicationProp) {
      onLoaded(applicationProp);
      return () => {
        cancelled = true;
      };
    }

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
            if (type === "device") {
              setUserCodeStatus("expired");
            }
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
  }, [type, owner, params.applicationName, params.casApplicationName, params.userCode, applicationProp]);

  // Already signed in on a plain /login: go to the console.
  React.useEffect(() => {
    if (!preview && type === "login" && account && !location.search.includes("silentSignin")) {
      navigate("/", {replace: true, state: {from: "/login"}});
    }
  }, [account, type, navigate, location.search, preview]);

  /** An iframe-embedded sign-in reports its progress to the host page. */
  const sendSilentSigninData = (data: string) => {
    if (Setting.inIframe()) {
      window.parent.postMessage({tag: "Casdoor", type: "SilentSignin", data}, "*");
    }
  };

  /** A sign-in opened as a popup hands the result back instead of redirecting. */
  const sendPopupData = (message: any, redirectUri: string) => {
    const searchParams = new URLSearchParams(location.search);
    if (searchParams.get("popup") !== "1") {
      return;
    }
    if ((searchParams.get("popup_type") || "window") === "iframe") {
      window.parent.postMessage(message, new URL(redirectUri).origin);
    } else {
      window.opener?.postMessage(message, redirectUri);
    }
  };

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
    const searchParams = new URLSearchParams(location.search);
    const isIframePopup = searchParams.get("popup") === "1" && (searchParams.get("popup_type") || "window") === "iframe";
    if (!isIframePopup) {
      Setting.goToLink(redirectUrl);
    }
    sendPopupData({type: "loginSuccess", data: {code: codeValue, state: oAuthParams.state}}, oAuthParams.redirectUri);
  };

  const handleLoginResult = (res: any, values: any, authParams: any) => {
    const responseType = values["type"];
    const responseTypes = String(responseType).split(" ");
    const responseMode = authParams?.responseMode || "query";

    if (responseType === "login") {
      Setting.showMessage("success", i18next.t("application:Logged in successfully"));
      reload().then(() => navigate(Setting.getFromLink(), {state: {from: "/login"}}));
    } else if (responseType === "device") {
      Setting.showMessage("success", i18next.t("application:Logged in successfully"));
      setUserCodeStatus("success");
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
    // A code handed to Casdoor's own /callback would be read there as a social-login
    // round trip and the client's "state" as one of Casdoor's own, so such a request
    // signs the user in instead, the same way AuthCallback treats this redirect URI.
    if (values.type !== "device" && Setting.isSelfRedirectUri(oAuthParams?.redirectUri)) {
      values.type = "login";
    }
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
      if (loginMethod === "verificationCodePhone") {
        values.countryCode = countryCode;
      }
    }

    return applyRequestType(values);
  };

  /** the checks the antd form ran as Form rules; without them an empty form is posted */
  const validateSignin = () => {
    const found: Record<string, string> = {};
    if (isPanelMethod) {
      return found;
    }

    const value = username.trim();
    if (isPhoneCodeMethod) {
      if (countryCode === "" && !application.organizationObj?.countryCodes?.[0]) {
        found.countryCode = i18next.t("signup:Please select your country code!");
      } else if (value === "") {
        found.username = i18next.t("signup:Please input your phone number!");
      } else if (!Setting.isValidPhone(value, countryCode)) {
        found.username = i18next.t("signup:The input is not valid Phone!");
      }
    } else if (value === "") {
      if (loginMethod !== "webAuthn") {
        found.username = getUsernameRequiredMessage(loginMethod);
      }
    } else if (loginMethod === "verificationCode" && !Setting.isValidEmail(value) && !Setting.isValidPhone(value)) {
      found.username = i18next.t("login:The input is not valid Email or phone number!");
    } else if (loginMethod === "verificationCodeEmail" && !Setting.isValidEmail(value)) {
      found.username = i18next.t("login:The input is not valid Email!");
    }

    if ((loginMethod === "password" || loginMethod === "ldap") && password === "") {
      found.password = i18next.t("login:Please input your password!");
    }
    if (isCodeMethod && code === "") {
      found.code = i18next.t("login:Please input your code!");
    }
    return found;
  };

  const fieldError = (field: string) =>
    signinErrors[field] ? <p className="text-xs text-destructive">{signinErrors[field]}</p> : null;

  const clearError = (field: string) =>
    setSigninErrors((prev) => (prev[field] ? {...prev, [field]: ""} : prev));

  /** the organization is encoded in the client id or the path, so keep the query string */
  const switchLoginOrganization = (name: string) => {
    const searchParams = new URLSearchParams(window.location.search);

    const clientId = searchParams.get("client_id");
    if (clientId) {
      searchParams.set("client_id", `${clientId.split("-org-")[0]}-org-${name}`);
      Setting.goToLink(`/login/oauth/authorize?${searchParams.toString()}`);
      return;
    }
    if (window.location.pathname.startsWith("/login/saml/authorize")) {
      Setting.goToLink(`/login/saml/authorize/${name}/${application.name}-org-${name}?${searchParams.toString()}`);
      return;
    }
    if (window.location.pathname.startsWith("/cas")) {
      Setting.goToLink(`/cas/${application.name}-org-${name}/${name}/login?${searchParams.toString()}`);
      return;
    }
    searchParams.set("orgChoiceMode", "None");
    Setting.goToLink(`/login/${name}?${searchParams.toString()}`);
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
  const cancelDeviceLogin = () => {
    const cancelToken = new URLSearchParams(location.search).get("cancelToken") || "";
    AuthBackend.cancelDeviceLogin(params.userCode, cancelToken)
      .then((res: any) => {
        if (res.status === "ok") {
          setUserCodeStatus("canceled");
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
    const found = validateSignin();
    setSigninErrors(found);
    const firstError = Object.values(found)[0];
    if (firstError) {
      Setting.showMessage("error", firstError);
      return;
    }
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

  // A popup host wants to know when the visitor closes the window without signing in.
  React.useEffect(() => {
    const searchParams = new URLSearchParams(location.search);
    if (preview || searchParams.get("popup") !== "1") {
      return;
    }
    const onUnload = () => sendPopupData({type: "windowClosed"}, searchParams.get("redirect_uri") ?? "");
    window.addEventListener("beforeunload", onUnload);
    return () => window.removeEventListener("beforeunload", onUnload);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [location.search]);

  /**
   * The antd page's componentDidUpdate: a visitor already signed in to this
   * organization does not have to touch the form. A silent sign-in reports back to
   * the embedding page, and `enableAutoSignin` submits on its own.
   */
  const autoSignedIn = React.useRef(false);
  React.useEffect(() => {
    if (preview || account === undefined) {
      return;
    }
    if (account === null) {
      sendSilentSigninData("user-not-logged-in");
      return;
    }
    if (!application || account.owner !== application.organization || autoSignedIn.current) {
      return;
    }

    const silentSignin = new URLSearchParams(location.search).get("silentSignin");
    if (silentSignin !== null) {
      autoSignedIn.current = true;
      sendSilentSigninData("signing-in");
      loginAsCurrentAccount();
    } else if (application.enableAutoSignin) {
      autoSignedIn.current = true;
      loginAsCurrentAccount();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [account, application, location.search]);

  if (saml !== null) {
    return <RedirectForm samlResponse={saml.response} redirectUrl={saml.redirectUrl} relayState={saml.relayState} />;
  }

  if (userCodeStatus === "expired") {
    return (
      <AuthLayout preview={!!preview}>
        <Alert variant="destructive">
          <AlertDescription>{`Code ${i18next.t("subscription:Expired")}`}</AlertDescription>
        </Alert>
      </AuthLayout>
    );
  }

  if (application === undefined) {
    return <Loading className="min-h-screen" />;
  }

  if (application === null) {
    return (
      <AuthLayout preview={!!preview}>
        <Alert variant="destructive">
          <AlertDescription>{msg ?? i18next.t("application:Failed to sign in")}</AlertDescription>
        </Alert>
      </AuthLayout>
    );
  }

  if (mfa !== null) {
    return (
      <AuthLayout preview={!!preview} application={application}>
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

  if (userCodeStatus === "canceled") {
    return (
      <AuthLayout preview={!!preview} application={application}>
        <Alert variant="warning">
          <AlertDescription>{i18next.t("login:Device login was canceled")}</AlertDescription>
        </Alert>
      </AuthLayout>
    );
  }

  if (userCodeStatus === "success") {
    return (
      <AuthLayout preview={!!preview} application={application}>
        <Alert>
          <AlertDescription>{i18next.t("application:Logged in successfully")}</AlertDescription>
        </Alert>
      </AuthLayout>
    );
  }

  if (application.disableSignin || application.organizationObj?.disableSignin) {
    return (
      <AuthLayout preview={!!preview} application={application}>
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
  const isCodeMethod = (loginMethod ?? "").startsWith("verificationCode");
  // "Email only" and "Phone only" share the "Verification code" tab, but they are
  // different methods: the tab has to lead back to the one the application configured
  const codeMethod = getCodeLoginMethod(application);
  const isPhoneCodeMethod = loginMethod === "verificationCodePhone";
  const activeTab = isCodeMethod ? codeMethod : loginMethod;
  // the application lists its methods in the order it wants them shown, each with
  // its own rule and display name
  const methods = getSigninMethods(application, type);
  const showTabs = methods.length > 1;
  // each block of the form can be hidden from "Signin items" in the application
  const isVisible = (name: string) => Setting.isSigninItemVisible(application, name);
  // the QR panels replace the credential form entirely
  const isPanelMethod = loginMethod === "wechat" || loginMethod === "device";
  // device login is offered next to the form when the application asks for it
  const deviceLoginOnLoginPage = type !== "device" && (application.signinMethods ?? []).some(
    (item: any) => item.name === "Device login" && item.rule === "Login page",
  );
  // the WeChat QR code can sit next to the form instead of replacing it
  const wechatOnLoginPage = (application.signinMethods ?? []).some(
    (item: any) => item.name === "WeChat" && item.rule === "Login page",
  );
  // without any credential method there is no form, only the providers
  const showForm = passwordEnabled || codeEnabled || webAuthnEnabled || ldapEnabled || faceIdEnabled;

  // With no form and a single third-party provider there is nothing to choose.
  const visibleOAuthProviderItems = (application.providers ?? []).filter(
    (item: any) => Setting.isProviderVisibleForSignIn(item) && item.provider?.category !== "SAML",
  );
  if (preview !== "auto" && !passwordEnabled && !codeEnabled && !webAuthnEnabled && !ldapEnabled && visibleOAuthProviderItems.length === 1) {
    Setting.goToLink(Provider.getAuthUrl(application, visibleOAuthProviderItems[0].provider, "signin"));
    return <Loading className="min-h-screen" />;
  }

  // An application can ask the visitor which organization they belong to before
  // showing the form. Choosing one reloads /login/<org> with the box turned off,
  // exactly as the antd page does.
  const orgChoiceMode = new URLSearchParams(location.search).get("orgChoiceMode") === "None"
    ? "None"
    : application.orgChoiceMode;
  if (type === "login" && (orgChoiceMode === "Select" || orgChoiceMode === "Input")) {
    return (
      <AuthLayout preview={!!preview} application={application}>
        <OrganizationChoiceBox mode={orgChoiceMode} />
      </AuthLayout>
    );
  }

  // Already signed in to this organization: offer the one-click path before the
  // form, which is how an OAuth or device request gets approved.
  const showSignedInBox = !!account && account.owner === application.organization;

  // The whole page can be replaced by the application's own markup.
  if (application.signinHtml) {
    return <CustomHtml html={application.signinHtml} />;
  }

  const signinItems = (application.signinItems ?? []) as any[];
  const hasCodeSigninItem = signinItems.some((item: any) => item.name === "Verification code");

  /**
   * One entry of the application's "Signin items", rendered in the configured
   * order. An item carries its own label, placeholder and custom CSS, and the
   * "Text N" ones carry raw HTML instead of a widget.
   */
  /**
   * The verification-code field. The application may place it as its own signin
   * item; when it does not, the antd page renders it in the password slot, so a
   * code-based sign-in still has somewhere to type the code.
   */
  const renderCodeInput = (item: any) => (
    <div key={item.name} className="verification-code space-y-2">
      <Label>{item.label || i18next.t("login:Verification code")}</Label>
      <SendCodeInput
        className="verification-code-input"
        value={code}
        onChange={(v) => {
          setCode(v);
          clearError("code");
        }}
        method="login"
        destType={loginMethod === "verificationCodeEmail" || (!isPhoneCodeMethod && username.includes("@")) ? "email" : "phone"}
        dest={username}
        // the antd page only enabled the button once the identifier parsed
        disabled={isPhoneCodeMethod
          ? !Setting.isValidPhone(username, countryCode)
          : loginMethod === "verificationCodeEmail"
            ? !Setting.isValidEmail(username)
            : !Setting.isValidEmail(username) && !Setting.isValidPhone(username)}
        countryCode={isPhoneCodeMethod ? countryCode : ""}
        placeholder={item.placeholder}
        application={application}
        applicationId={Setting.getApplicationName(application)}
        useInlineCaptcha={Setting.isInlineCaptchaEnabled(application)}
        captchaValue={captchaValues}
        refreshCaptcha={refreshInlineCaptcha}
      />
      {fieldError("code")}
    </div>
  );

  const renderSigninItem = (item: any) => {
    const key = item.name;
    if (Setting.isCustomFormItem(item)) {
      return item.visible ? <CustomHtml key={key} html={item.customCss} /> : null;
    }
    // the antd page keeps the auto sign-in checkbox even when the link is hidden
    if (!item.visible && item.name !== "Forgot password?") {
      return null;
    }

    switch (item.name) {
    // Logo and Languages are rendered by AuthLayout, which is told about their
    // visibility, and the browser's own back gesture replaces the back button
    case "Logo":
    case "Languages":
    case "Back button":
      return null;
    case "Signin methods":
      return showTabs ? (
        <SigninMethodTabs
          key={key}
          className="signin-methods"
          methods={methods}
          value={activeTab}
          onChange={(v) => setLoginMethod(v as LoginMethod)}
        />
      ) : null;
    case "Select organization":
      return (
        <div key={key} className="login-organization-select">
          <OrganizationSelect value={application.organization} onChange={switchLoginOrganization} />
        </div>
      );
    case "Username":
      if (loginMethod === "wechat") {
        return <WeChatLoginPanel key={key} application={application} />;
      }
      if (loginMethod === "device") {
        return <DeviceLoginPanel key={key} application={application} onSuccess={completeDeviceLogin} />;
      }
      if (isPhoneCodeMethod) {
        return (
          <div key={key} className="signin-phone space-y-2">
            <Label htmlFor="username">{item.label || i18next.t("general:Phone")}</Label>
            <div className="flex gap-2">
              <div className="w-28 max-w-[50%] shrink-0">
                <CountryCodeSelect
                  className="px-2"
                  value={countryCode}
                  onChange={setCountryCode}
                  countryCodes={application.organizationObj?.countryCodes}
                />
              </div>
              <Input
                id="username"
                className="login-username-input min-w-0 flex-1"
                autoFocus
                autoComplete="tel"
                placeholder={item.placeholder || i18next.t("general:Phone")}
                value={username}
                onChange={(e) => {
                  setUsername(e.target.value);
                  clearError("username");
                }}
              />
            </div>
            {fieldError("countryCode")}
            {fieldError("username")}
          </div>
        );
      }
      return (
        <div key={key} className="login-username space-y-2">
          <Label htmlFor="username">
            {item.label || (isCodeMethod ? getUsernameLabel(loginMethod) : i18next.t("signup:Username"))}
          </Label>
          <Input
            id="username"
            className="login-username-input"
            autoFocus
            autoComplete="username"
            placeholder={item.placeholder || getUsernameLabel(loginMethod)}
            value={username}
            onChange={(e) => {
              setUsername(e.target.value);
              clearError("username");
            }}
          />
          {fieldError("username")}
        </div>
      );
    case "Verification code":
      if (isPanelMethod || !isCodeMethod) {
        return null;
      }
      return renderCodeInput(item);
    case "Password":
      if (isPanelMethod || loginMethod === "webAuthn" || loginMethod === "faceId") {
        return null;
      }
      if (isCodeMethod) {
        return hasCodeSigninItem ? null : renderCodeInput(item);
      }
      return (
        <div key={key} className="login-password space-y-2">
          <Label htmlFor="password">{item.label || i18next.t("general:Password")}</Label>
          <PasswordInput
            id="password"
            className="login-password-input"
            autoComplete="current-password"
            placeholder={item.placeholder}
            value={password}
            onChange={(e) => {
              setPassword(e.target.value);
              clearError("password");
            }}
          />
          {fieldError("password")}
        </div>
      );
    case "Forgot password?":
      // the item's default CSS pins this row at 320px, wider than some panels
      return (
        <div key={key} className="login-forget-password flex max-w-full flex-wrap items-center justify-between gap-x-3 gap-y-2">
          <div className="flex items-center gap-2">
            <Checkbox id="autoSignin" checked={autoSignin} onCheckedChange={(v) => setAutoSignin(v === true)} />
            <Label htmlFor="autoSignin" className="login-auto-signin cursor-pointer text-sm font-normal text-muted-foreground">
              {i18next.t("login:Auto sign in")}
            </Label>
          </div>
          {item.visible ? <ForgetLink application={application} label={item.label} /> : null}
        </div>
      );
    case "Agreement":
      return application.termsOfUse ? (
        <AgreementCheckbox key={key} application={application} checked={agreed} onChange={setAgreed} />
      ) : null;
    case "Captcha":
      if (item.rule !== "inline" || !captchaProvider) {
        return null;
      }
      return (
        <CaptchaModal
          key={key}
          noModal
          owner={captchaProvider.owner}
          name={captchaProvider.name}
          isCurrentProvider
          innerRef={captchaRef}
          onUpdateToken={(captchaType, captchaToken, clientSecret) =>
            setCaptchaValues({captchaType, captchaToken, clientSecret})
          }
        />
      );
    case "Login button":
      if (isPanelMethod) {
        return null;
      }
      return (
        <div key={key} className="login-button-box space-y-3">
          <Button type="submit" className="login-button w-full" loading={loading}>
            {loginMethod === "webAuthn"
              ? i18next.t("login:Sign in with WebAuthn")
              : loginMethod === "faceId"
                ? i18next.t("login:Sign in with Face ID")
                : type === "device"
                  ? i18next.t("login:Approve and sign in")
                  : item.label || i18next.t("login:Sign In")}
          </Button>
          {type === "device" ? (
            <Button type="button" variant="outline" className="w-full" onClick={cancelDeviceLogin}>
              {i18next.t("general:Cancel")}
            </Button>
          ) : null}
        </div>
      );
    case "Signup link": {
      if (!application.enableSignUp) {
        return null;
      }
      const signupUrl = Setting.getSignupLink(application) ?? "/signup";
      const signupText = item.label || i18next.t("login:sign up now");
      const signupClass = "font-medium text-foreground underline-offset-4 hover:underline";
      return (
        <p key={key} className="login-signup-link pt-1 text-center text-sm text-muted-foreground">
          {item.label ? null : <span className="mr-1">{i18next.t("login:No account?")}</span>}
          {signupUrl.startsWith("/") ? (
            <Link to={signupUrl} onClick={Setting.storeSigninUrl} className={signupClass}>{signupText}</Link>
          ) : (
            <a href={signupUrl} onClick={Setting.storeSigninUrl} className={signupClass}>{signupText}</a>
          )}
        </p>
      );
    }
    case "Providers":
      if (getVisibleProviders(application, "signin").length === 0) {
        return null;
      }
      return (
        <div key={key} className="space-y-4">
          {showForm ? <AuthDivider /> : null}
          <ProviderButtons
            application={application}
            method="signin"
            rule={Setting.getProvidersRule(application, item)}
          />
        </div>
      );
    default:
      return null;
    }
  };

  return (
    <AuthLayout
      preview={!!preview}
      application={application}
      hideLogo={!isVisible("Logo")}
      hideLanguages={!isVisible("Languages")}
    >
      <div className="space-y-6">
        {/* each item styles itself; a "Text N" item holds HTML, not CSS */}
        {signinItems.map((item: any) =>
          Setting.isCustomFormItem(item) ? null : <CustomStyle key={`css-${item.name}`} css={item.customCss} />,
        )}

        {type === "device" && params.userCode ? (
          <div className="space-y-1 rounded-lg border bg-muted/50 p-4 text-center">
            <div className="text-sm text-muted-foreground">{i18next.t("login:Confirmation code")}</div>
            <div className="font-mono text-2xl font-semibold tracking-[0.2em]">{params.userCode}</div>
          </div>
        ) : null}

        {showSignedInBox ? (
          <div className="space-y-3">
            <p className="text-sm text-muted-foreground">
              {type === "device"
                ? i18next.t("login:Continue with your current account to approve this sign-in")
                : i18next.t("login:Continue with")}
            </p>
            <Button
              type="button"
              variant="outline"
              className="h-14 w-full justify-start gap-3 px-3 text-left"
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
              <span className="min-w-0 flex-1 truncate">
                {account.displayName ? `${account.name} (${account.displayName})` : account.name}
              </span>
            </Button>
            <AuthDivider label={i18next.t("login:Or sign in with another account")} />
          </div>
        ) : null}

        {showForm ? (
          <form className="space-y-4" onSubmit={submit}>
            {signinItems.map(renderSigninItem)}
          </form>
        ) : (
          <div className="space-y-4">
            <p className="text-sm text-muted-foreground">
              {i18next.t("login:To access")}&nbsp;
              <a
                target="_blank"
                rel="noreferrer"
                href={application.homepageUrl}
                className="font-medium text-foreground underline-offset-4 hover:underline"
              >
                {application.displayName}
              </a>
            </p>
            {signinItems
              .filter((item: any) => item.name === "Providers" || item.name === "Signup link")
              .map(renderSigninItem)}
          </div>
        )}

        {wechatOnLoginPage ? (
          <div className="space-y-4">
            <AuthDivider />
            <p className="text-center text-sm text-muted-foreground">
              {i18next.t("provider:Please use WeChat to scan the QR code and follow the official account for sign in")}
            </p>
            <WeChatLoginPanel application={application} />
          </div>
        ) : null}

        {deviceLoginOnLoginPage ? (
          <div className="space-y-4">
            <AuthDivider />
            <DeviceLoginPanel application={application} onSuccess={completeDeviceLogin} />
          </div>
        ) : null}

        {preview === "auto" ? null : <GoogleOneTap application={application} />}

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
