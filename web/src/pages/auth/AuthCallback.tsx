import * as React from "react";
import i18next from "i18next";
import {useLocation, useNavigate} from "react-router-dom";
import {Alert, AlertDescription} from "@/components/ui/alert";
import {Loading} from "@/components/common/Loading";
import {AuthLayout} from "@/components/auth/AuthLayout";
import {MfaVerify, NextMfa} from "@/components/auth/MfaVerify";
import {RedirectForm} from "@/components/auth/RedirectForm";
import {authConfig} from "@/auth/Auth";
import * as Provider from "@/auth/Provider";
import * as Util from "@/auth/Util";
import * as ApplicationBackend from "@/backend/ApplicationBackend";
import * as AuthBackend from "@/backend/AuthBackend";
import * as Setting from "@/lib/setting";

/**
 * Landing page of the OAuth/OIDC/SAML/CAS round trip. It re-posts the provider's
 * code to /api/login and then performs whatever redirect the original request asked
 * for — the same contract the antd AuthCallback implemented.
 */
/**
 * `routers/lightweight_auth_filter.go` answers `/callback?state=...` with a tiny
 * static page that posts the provider's code itself, so the browser does not have
 * to download the whole bundle first. When that page meets something it cannot
 * finish — an MFA challenge, a plan to pick — it stores what it already got and
 * sends the browser here with `__casdoor_callback_react=1`.
 *
 * The authorization code is single-use, so this page must continue from that
 * stored response instead of POSTing the code a second time.
 */
const REACT_FALLBACK_KEY = "__casdoor_callback_react";
const REACT_FALLBACK_PAYLOAD_KEY = "casdoor_callback_react_fallback";

interface ReactFallbackPayload {
  search: string;
  res: any;
  body?: Record<string, any>;
  flow?: "cas" | "oauth";
  casService?: string;
  responseType?: string;
  /** the decoded `state` query string, so the payload survives a lost sessionStorage entry */
  innerParams?: string;
  queryString?: string;
}

/** The callback URL without the marker the static page adds, for comparing the two. */
function normalizedSearch(search: string): string {
  const url = new URL(`${window.location.origin}/callback${search || ""}`);
  url.searchParams.delete(REACT_FALLBACK_KEY);
  return url.search;
}

/**
 * Reads and clears the handover, but only when it belongs to *this* callback —
 * a stale payload from an earlier sign-in must not be replayed onto a new code.
 */
function consumeReactFallbackPayload(currentSearch: string): ReactFallbackPayload | null {
  const raw = sessionStorage.getItem(REACT_FALLBACK_PAYLOAD_KEY);
  if (!raw) {
    return null;
  }

  try {
    const payload = JSON.parse(raw) as ReactFallbackPayload;
    if (normalizedSearch(payload.search) !== normalizedSearch(currentSearch)) {
      return null;
    }
    sessionStorage.removeItem(REACT_FALLBACK_PAYLOAD_KEY);
    return payload;
  } catch {
    sessionStorage.removeItem(REACT_FALLBACK_PAYLOAD_KEY);
    return null;
  }
}

/** A login the backend answered with "NextMfa", waiting on the second factor. */
interface PendingMfa {
  props: any;
  values: Record<string, any>;
  authParams: any;
  onSuccess: (res: any) => void;
}

export default function AuthCallback() {
  const location = useLocation();
  const navigate = useNavigate();
  const [msg, setMsg] = React.useState<string | null>(null);
  const [saml, setSaml] = React.useState<{response: string; redirectUrl: string; relayState: string} | null>(null);
  // The provider's authorization code is single-use, so the second factor has to
  // be collected here; sending the user back to /login would drop the pending
  // login and leave them unable to finish.
  const [mfa, setMfa] = React.useState<PendingMfa | null>(null);
  const [application, setApplication] = React.useState<any>(null);
  const startedRef = React.useRef(false);

  React.useEffect(() => {
    if (startedRef.current) {
      return;
    }
    startedRef.current = true;

    const params = new URLSearchParams(location.search);
    const state = params.get("state") ?? "";
    const innerParams = new URLSearchParams(Util.getQueryParamsFromState(state));

    const getResponseType = (): string => {
      const method = innerParams.get("method");
      if (method === "signup" || method === "signin") {
        const userCode = innerParams.get("userCode");
        if (userCode) {
          return "device";
        }
        const realRedirectUri = innerParams.get("redirect_uri");
        if (realRedirectUri === null) {
          const samlRequest = innerParams.get("SAMLRequest");
          const casService = innerParams.get("service");
          if (samlRequest) {
            return "saml";
          }
          if (casService) {
            return "cas";
          }
          return "login";
        }
        const realRedirectUrl = new URL(realRedirectUri).origin;
        if (authConfig.serverUrl === realRedirectUrl) {
          return "login";
        }
        return innerParams.get("response_type") ?? "code";
      }
      if (method === "link") {
        return "link";
      }
      return "unknown";
    };

    // Providers disagree on the parameter name for the authorization code.
    let code =
      params.get("code") ?? params.get("auth_code") ?? params.get("authCode") ?? null;
    if (code === null) {
      const web3Key = params.get("web3AuthTokenKey");
      if (web3Key) {
        code = localStorage.getItem(web3Key);
      }
    }
    const isSteam = params.get("openid.mode");
    if (isSteam !== null && code === null) {
      code = location.search;
    }

    // Telegram passes the auth data as individual query params.
    const telegramId = params.get("id");
    if (telegramId !== null && !code) {
      const telegramAuthData: Record<string, any> = {id: parseInt(telegramId, 10)};
      ["hash", "auth_date", "first_name", "last_name", "username", "photo_url"].forEach((field) => {
        const value = params.get(field);
        if (value) {
          telegramAuthData[field] = value;
        }
      });
      code = JSON.stringify(telegramAuthData);
    }

    const responseType = getResponseType();
    const applicationName = innerParams.get("application");
    const casService = innerParams.get("service") ?? "";
    const codeVerifier = Provider.getCodeVerifier(state);

    const body: Record<string, any> = {
      type: responseType,
      application: applicationName,
      provider: innerParams.get("provider"),
      code,
      samlRequest: innerParams.get("SAMLRequest"),
      state: applicationName,
      invitationCode: innerParams.get("invitationCode") || "",
      redirectUri: `${window.location.origin}/callback`,
      method: innerParams.get("method"),
      userCode: innerParams.get("userCode") || "",
      codeVerifier,
      language: Setting.getSigninLanguage(),
    };

    if (codeVerifier) {
      Provider.clearCodeVerifier(state);
    }

    // the fallback hands over its own `state`, so these are arguments rather than
    // closures over the ones this page decoded
    const handleOAuth = (res: any, params: URLSearchParams = innerParams, type: string = responseType) => {
      const oAuthParams = Util.getOAuthGetParameters(params);
      const concatChar = oAuthParams?.redirectUri?.includes("?") ? "&" : "?";
      const responseMode = oAuthParams?.responseMode || "query";
      const responseTypes = type.split(" ");

      // The backend answers with `data: {required: true}` instead of an authorization
      // code when the user still has to consent. The consent page issues the real code,
      // and reads the original authorization request from the query string, which the
      // state carries rather than the callback URL.
      if (res.data?.required === true) {
        Setting.goToLink(`/consent/${params.get("application") ?? applicationName}?${params.toString()}`);
        return;
      }

      if (type === "login" || type === "device") {
        Setting.showMessage("success", i18next.t("application:Logged in successfully"));
        navigate(Setting.getFromLink());
      } else if (type === "code") {
        if (responseMode === "form_post") {
          Setting.createFormAndSubmit(oAuthParams?.redirectUri, {code: res.data, state: oAuthParams?.state});
        } else {
          Setting.goToLink(
            `${oAuthParams.redirectUri}${concatChar}code=${encodeURIComponent(res.data)}&state=${encodeURIComponent(
              oAuthParams.state,
            )}`,
          );
        }
      } else if (responseTypes.includes("token") || responseTypes.includes("id_token")) {
        if (responseMode === "form_post") {
          Setting.createFormAndSubmit(oAuthParams?.redirectUri, {
            token: responseTypes.includes("token") ? res.data : null,
            id_token: responseTypes.includes("id_token") ? res.data : null,
            token_type: "bearer",
            state: oAuthParams?.state,
          });
        } else {
          Setting.goToLink(
            `${oAuthParams.redirectUri}${concatChar}${type}=${encodeURIComponent(
              res.data,
            )}&state=${encodeURIComponent(oAuthParams.state)}&token_type=bearer`,
          );
        }
      } else if (type === "link") {
        let from = params.get("from") ?? "/";
        const oauth = params.get("oauth");
        if (oauth) {
          from += `?oauth=${oauth}`;
        }
        navigate(from);
      } else if (type === "saml") {
        if (res.data2?.method === "POST") {
          setSaml({response: res.data, redirectUrl: res.data2.redirectUrl, relayState: oAuthParams.relayState});
        } else {
          const redirectUri = res.data2.redirectUrl;
          Setting.goToLink(
            `${redirectUri}${redirectUri.includes("?") ? "&" : "?"}SAMLResponse=${encodeURIComponent(res.data)}&RelayState=${
              oAuthParams.relayState
            }`,
          );
        }
      }
    };

    const checkMfa = (res: any, authParams: any, onDone: (res: any) => void) => {
      if (res.data === Setting.RequiredUpdatePassword) {
        Setting.goToUpdatePassword();
      } else if (res.data === "RequiredMfa") {
        // the account reload in the console then bounces to /mfa/setup
        localStorage.setItem("mfaRedirectUrl", window.location.origin);
        Setting.goToLink(window.location.origin);
      } else if (res.data === NextMfa) {
        // the panel needs the application for its branding and the captcha rule
        // behind "Get Code"
        if (applicationName) {
          ApplicationBackend.getApplication("admin", applicationName).then((appRes: any) => {
            setApplication(appRes.status === "ok" ? appRes.data : null);
          });
        }
        setMfa({
          props: res.data2,
          values: {...body, providerBack: body.provider, provider: ""},
          authParams,
          onSuccess: onDone,
        });
      } else if (res.data === "SelectPlan") {
        const pricing = res.data2;
        Setting.goToLink(`/select-plan/${pricing.owner}/${pricing.name}?user=${body.username}`);
      } else if (res.data === "BuyPlanResult") {
        const sub = res.data2;
        Setting.goToLink(`/buy-plan/${sub.owner}/${sub.pricing}/result?subscription=${sub.name}`);
      } else {
        onDone(res);
      }
    };

    const handleCas = (ok: any, service: string) => {
      let message = "Logged in successfully.";
      if (service === "") {
        message += " Now you can visit apps protected by Casdoor.";
      }
      Setting.showMessage("success", message);
      if (service !== "") {
        const newUrl = new URL(service);
        newUrl.searchParams.append("ticket", ok.data);
        window.location.href = newUrl.toString();
      }
    };

    // The static callback page already spent the authorization code and handed the
    // answer over, so continue from it instead of signing in again.
    const fallback = consumeReactFallbackPayload(location.search);
    if (fallback !== null) {
      if (fallback.flow === "cas") {
        const service = fallback.casService ?? casService;
        checkMfa(fallback.res, {service}, (ok) => handleCas(ok, service));
      } else {
        const fallbackParams = new URLSearchParams(
          fallback.innerParams || Util.getQueryParamsFromState(state),
        );
        const fallbackType = fallback.responseType || responseType;
        checkMfa(fallback.res, Util.getOAuthGetParameters(fallbackParams), (ok) =>
          handleOAuth(ok, fallbackParams, fallbackType),
        );
      }
      return;
    }

    if (responseType === "cas") {
      AuthBackend.loginCas(body, {service: casService}).then((res: any) => {
        if (res.status === "ok") {
          checkMfa(res, {service: casService}, (ok) => handleCas(ok, casService));
        } else {
          setMsg(res.msg);
        }
      });
      return;
    }

    const oAuthParams = Util.getOAuthGetParameters(innerParams);
    AuthBackend.login(body, oAuthParams)
      .then((res: any) => {
        if (res.status === "ok") {
          checkMfa(res, oAuthParams, handleOAuth);
        } else {
          setMsg(res.msg);
        }
      })
      .catch((error) => setMsg(`${i18next.t("general:Failed to connect to server")}: ${error}`));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  if (saml !== null) {
    return <RedirectForm samlResponse={saml.response} redirectUrl={saml.redirectUrl} relayState={saml.relayState} />;
  }

  if (mfa !== null) {
    return (
      <AuthLayout application={application}>
        <MfaVerify
          formValues={mfa.values}
          authParams={mfa.authParams}
          mfaProps={mfa.props}
          application={application}
          onSuccess={(res) => mfa.onSuccess(res)}
        />
      </AuthLayout>
    );
  }

  if (msg !== null) {
    return (
      <AuthLayout>
        <Alert variant="destructive">
          <AlertDescription>{msg}</AlertDescription>
        </Alert>
      </AuthLayout>
    );
  }

  return <Loading className="min-h-screen" />;
}
