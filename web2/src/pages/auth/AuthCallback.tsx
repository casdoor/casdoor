import * as React from "react";
import i18next from "i18next";
import {useLocation, useNavigate} from "react-router-dom";
import {Alert, AlertDescription} from "@/components/ui/alert";
import {Loading} from "@/components/common/Loading";
import {AuthLayout} from "@/components/auth/AuthLayout";
import {RedirectForm} from "@/components/auth/RedirectForm";
import {authConfig} from "@/auth/Auth";
import * as Provider from "@/auth/Provider";
import * as Util from "@/auth/Util";
import * as AuthBackend from "@/backend/AuthBackend";
import * as Setting from "@/lib/setting";

/**
 * Landing page of the OAuth/OIDC/SAML/CAS round trip. It re-posts the provider's
 * code to /api/login and then performs whatever redirect the original request asked
 * for — the same contract the antd AuthCallback implemented.
 */
export default function AuthCallback() {
  const location = useLocation();
  const navigate = useNavigate();
  const [msg, setMsg] = React.useState<string | null>(null);
  const [saml, setSaml] = React.useState<{response: string; redirectUrl: string; relayState: string} | null>(null);
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

    const handleOAuth = (res: any) => {
      const oAuthParams = Util.getOAuthGetParameters(innerParams);
      const concatChar = oAuthParams?.redirectUri?.includes("?") ? "&" : "?";
      const responseMode = oAuthParams?.responseMode || "query";
      const responseTypes = responseType.split(" ");

      if (responseType === "login" || responseType === "device") {
        Setting.showMessage("success", i18next.t("application:Logged in successfully"));
        navigate(Setting.getFromLink());
      } else if (responseType === "code") {
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
            `${oAuthParams.redirectUri}${concatChar}${responseType}=${encodeURIComponent(
              res.data,
            )}&state=${encodeURIComponent(oAuthParams.state)}&token_type=bearer`,
          );
        }
      } else if (responseType === "link") {
        let from = innerParams.get("from") ?? "/";
        const oauth = innerParams.get("oauth");
        if (oauth) {
          from += `?oauth=${oauth}`;
        }
        navigate(from);
      } else if (responseType === "saml") {
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

    const checkMfa = (res: any, onDone: (res: any) => void) => {
      if (res.data === Setting.RequiredUpdatePassword) {
        Setting.goToUpdatePassword();
      } else if (res.data === "RequiredMfa") {
        // the account reload in the console then bounces to /mfa/setup
        localStorage.setItem("mfaRedirectUrl", window.location.origin);
        Setting.goToLink(window.location.origin);
      } else if (res.data === "NextMfa") {
        // The second factor is collected on the login page, which owns the form state.
        setMsg(i18next.t("mfa:Multi-factor authentication"));
        navigate("/login");
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

    if (responseType === "cas") {
      AuthBackend.loginCas(body, {service: casService}).then((res: any) => {
        if (res.status === "ok") {
          checkMfa(res, (ok) => {
            let message = "Logged in successfully.";
            if (casService === "") {
              message += " Now you can visit apps protected by Casdoor.";
            }
            Setting.showMessage("success", message);
            if (casService !== "") {
              const newUrl = new URL(casService);
              newUrl.searchParams.append("ticket", ok.data);
              window.location.href = newUrl.toString();
            }
          });
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
          checkMfa(res, handleOAuth);
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
