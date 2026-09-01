import * as React from "react";
import i18next from "i18next";
import {QRCodeSVG} from "qrcode.react";
import * as AuthBackend from "@/backend/AuthBackend";
import * as Util from "@/auth/Util";

type Phase = "loading" | "pending" | "success" | "denied" | "expired" | "error";

/**
 * The device-login QR panel of the sign-in page, ported from
 * web/src/auth/DeviceLoginPanel.js: it starts a device authorization, polls the
 * token endpoint at the interval the backend asks for, and hands the device code
 * back once another signed-in device has approved it.
 */
export function DeviceLoginPanel({application, onSuccess}: {application: any; onSuccess: (deviceCode: string) => void}) {
  const [session, setSession] = React.useState<any>(null);
  const [phase, setPhase] = React.useState<Phase>("loading");
  const [error, setError] = React.useState("");
  const [nonce, setNonce] = React.useState(0);

  const clientId = application?.clientId;
  const onSuccessRef = React.useRef(onSuccess);
  onSuccessRef.current = onSuccess;

  React.useEffect(() => {
    if (!clientId) {
      return;
    }
    let cancelled = false;
    let poller: number | undefined;
    let restart: number | undefined;

    const stop = () => {
      if (poller !== undefined) {
        window.clearInterval(poller);
        poller = undefined;
      }
    };

    setSession(null);
    setPhase("loading");
    setError("");

    const scope = Util.getOAuthGetParameters()?.scope || "openid profile email";
    AuthBackend.startDeviceLogin(clientId, scope)
      .then((res: any) => {
        if (cancelled) {
          return;
        }
        if (res.error || !res.device_code || !res.verification_uri) {
          setPhase("error");
          setError(res.error_description || res.error || i18next.t("login:Device login is unavailable"));
          return;
        }

        setSession(res);
        setPhase("pending");

        poller = window.setInterval(() => {
          AuthBackend.pollDeviceLoginToken(clientId, res.device_code)
            .then((tokenRes: any) => {
              if (cancelled) {
                return;
              }
              if (tokenRes.access_token) {
                stop();
                setPhase("success");
                onSuccessRef.current(res.device_code);
                return;
              }
              if (tokenRes.error === "authorization_pending") {
                return;
              }

              stop();
              if (tokenRes.error === "access_denied") {
                setPhase("denied");
                setError(tokenRes.error_description || i18next.t("login:Device login was canceled"));
              } else if (tokenRes.error === "expired_token") {
                setPhase("expired");
                setError(tokenRes.error_description || i18next.t("login:Device login expired"));
                restart = window.setTimeout(() => setNonce((n) => n + 1), 800);
              } else {
                setPhase("error");
                setError(tokenRes.error_description || tokenRes.error || i18next.t("login:Device login is unavailable"));
              }
            })
            .catch((err: any) => {
              stop();
              if (!cancelled) {
                setPhase("error");
                setError(err.message);
              }
            });
        }, (res.interval || 5) * 1000);
      })
      .catch((err: any) => {
        if (!cancelled) {
          setPhase("error");
          setError(err.message);
        }
      });

    return () => {
      cancelled = true;
      stop();
      if (restart !== undefined) {
        window.clearTimeout(restart);
      }
    };
  }, [clientId, nonce]);

  const isBroken = phase === "denied" || phase === "expired" || phase === "error";

  return (
    <div className="mx-auto flex max-w-xs flex-col items-center gap-3 text-center">
      <h2 className="text-lg font-semibold">{i18next.t("login:Device login")}</h2>
      {phase === "success" ? (
        <p>{i18next.t("application:Logged in successfully")}</p>
      ) : isBroken ? (
        <p className="text-destructive">{error}</p>
      ) : (
        <>
          <p className="text-sm">{i18next.t("login:Scan this QR code with a signed-in device to continue")}</p>
          {session?.user_code ? (
            <p className="text-sm text-muted-foreground">
              {i18next.t("login:Confirmation code")}: {session.user_code}
            </p>
          ) : null}
        </>
      )}
      <div className={isBroken ? "opacity-30" : undefined}>
        <QRCodeSVG value={session?.verification_uri ?? " "} size={230} />
      </div>
      <button
        type="button"
        className="text-sm text-muted-foreground underline-offset-4 hover:underline"
        onClick={() => setNonce((n) => n + 1)}
      >
        {i18next.t("general:Refresh")}
      </button>
    </div>
  );
}
