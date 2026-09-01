import * as React from "react";
import i18next from "i18next";
import {useLocation, useNavigate} from "react-router-dom";
import {Alert, AlertDescription} from "@/components/ui/alert";
import {Loading} from "@/components/common/Loading";
import {AuthLayout} from "@/components/auth/AuthLayout";
import {MfaVerify, NextMfa} from "@/components/auth/MfaVerify";
import {authConfig} from "@/auth/Auth";
import * as ApplicationBackend from "@/backend/ApplicationBackend";
import * as AuthBackend from "@/backend/AuthBackend";
import * as Conf from "@/Conf";
import * as Setting from "@/lib/setting";

/**
 * Landing page of a sign-in through an external SAML identity provider.
 *
 * Unlike `/callback`, the IdP does not hand back an authorization code: it posts a
 * `samlResponse` plus the base64 `relayState` that Casdoor put into the AuthnRequest.
 * The relay state carries "clientId&state&provider&redirectUri", which is what
 * `/api/login` needs to turn the assertion into a session (or an OAuth code).
 * Ported from `web/src/auth/SamlCallback.js`.
 */
/** A login the backend answered with "NextMfa", waiting on the second factor. */
interface PendingMfa {
  props: any;
  values: Record<string, any>;
  authParams: any;
  onSuccess: (res: any) => void;
}

export default function SamlCallback() {
  const location = useLocation();
  const navigate = useNavigate();
  const [msg, setMsg] = React.useState<string | null>(null);
  // The assertion is single-use, so the second factor is collected here rather
  // than by sending the user back to /login with the pending login dropped.
  const [mfa, setMfa] = React.useState<PendingMfa | null>(null);
  const [application, setApplication] = React.useState<any>(null);
  const startedRef = React.useRef(false);

  React.useEffect(() => {
    if (startedRef.current) {
      return;
    }
    startedRef.current = true;

    const params = new URLSearchParams(location.search);
    const relayState = params.get("relayState") ?? "";
    const samlResponse = params.get("samlResponse") ?? "";

    let messages: string[];
    try {
      messages = atob(relayState).split("&");
    } catch (error) {
      setMsg(i18next.t("general:Error"));
      return;
    }

    const clientId = messages[0] === "" ? "" : messages[0];
    const applicationName = messages[0] === "" ? Conf.DefaultApplication : "";
    const state = messages[1];
    const providerName = messages[2];
    const redirectUri = messages[3];

    // Casdoor's own login page does not need an OAuth code, it just needs the session
    const getResponseType = (): string => {
      if (redirectUri === "null") {
        return "login";
      }
      try {
        if (authConfig.serverUrl === new URL(redirectUri).origin) {
          return "login";
        }
      } catch (error) {
        return "login";
      }
      return "code";
    };

    const responseType = getResponseType();

    const body: Record<string, any> = {
      type: responseType,
      clientId: clientId,
      provider: providerName,
      state: state,
      application: applicationName,
      redirectUri: `${window.location.origin}/callback`,
      method: "signup",
      relayState: relayState,
      samlResponse: encodeURIComponent(samlResponse),
    };

    const param =
      clientId === ""
        ? ""
        : `?clientId=${clientId}&responseType=${responseType}&redirectUri=${redirectUri}&scope=read&state=${state}`;

    const handleLogin = (res: any) => {
      if (responseType === "login") {
        Setting.showMessage("success", i18next.t("application:Logged in successfully"));
        navigate(Setting.getFromLink());
      } else if (responseType === "code") {
        Setting.goToLink(
          `${redirectUri}?code=${encodeURIComponent(res.data)}&state=${encodeURIComponent(state)}`,
        );
      }
    };

    // Same second-factor / password-update handling the OAuth callback does.
    const checkMfa = (res: any, onDone: (res: any) => void) => {
      if (res.data === Setting.RequiredUpdatePassword) {
        Setting.goToUpdatePassword();
      } else if (res.data === "RequiredMfa") {
        localStorage.setItem("mfaRedirectUrl", window.location.origin);
        Setting.goToLink(window.location.origin);
      } else if (res.data === NextMfa) {
        // the panel needs the application for its branding and the captcha rule
        // behind "Get Code"; with only a clientId there is no name to fetch by
        if (applicationName) {
          ApplicationBackend.getApplication("admin", applicationName).then((appRes: any) => {
            setApplication(appRes.status === "ok" ? appRes.data : null);
          });
        }
        setMfa({
          props: res.data2,
          values: {...body, providerBack: body.provider, provider: ""},
          // the same params antd hands MfaAuthVerifyForm, so the re-post keeps the OAuth context
          authParams: {clientId, responseType, redirectUri, state},
          onSuccess: onDone,
        });
      } else if (res.data === "SelectPlan") {
        const pricing = res.data2;
        Setting.goToLink(`/select-plan/${pricing.owner}/${pricing.name}`);
      } else if (res.data === "BuyPlanResult") {
        const sub = res.data2;
        Setting.goToLink(`/buy-plan/${sub.owner}/${sub.pricing}/result?subscription=${sub.name}`);
      } else {
        onDone(res);
      }
    };

    AuthBackend.loginWithSaml(body, param)
      .then((res: any) => {
        if (res.status === "ok") {
          checkMfa(res, handleLogin);
        } else {
          setMsg(res.msg);
        }
      })
      .catch((error) => setMsg(`${i18next.t("general:Failed to connect to server")}: ${error}`));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

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
