import * as React from "react";
import * as Provider from "@/auth/Provider";
import * as Setting from "@/lib/setting";

const GSI_SRC = "https://accounts.google.com/gsi/client";

function loadGsiScript(): Promise<void> {
  return new Promise((resolve, reject) => {
    if ((window as any).google?.accounts?.id) {
      resolve();
      return;
    }
    const existing = document.querySelector(`script[src="${GSI_SRC}"]`) as HTMLScriptElement | null;
    if (existing) {
      existing.addEventListener("load", () => resolve());
      existing.addEventListener("error", () => reject(new Error("failed to load Google Identity Services")));
      return;
    }
    const script = document.createElement("script");
    script.src = GSI_SRC;
    script.async = true;
    script.onload = () => resolve();
    script.onerror = () => reject(new Error("failed to load Google Identity Services"));
    document.body.appendChild(script);
  });
}

/**
 * Google One Tap, shown when a Google provider has rule "OneTap". The antd
 * frontend wraps `react-google-one-tap-login`; here the Google Identity Services
 * script is used directly, and the credential is handed to /callback in the same
 * "GoogleIdToken-<json>" shape the backend already parses.
 */
export function GoogleOneTap({application}: {application: any}) {
  const providerItem = (application?.providers ?? []).find(
    (item: any) => item.provider?.type === "Google" && item.rule === "OneTap",
  );
  const clientId = providerItem?.provider?.clientId;

  React.useEffect(() => {
    if (!clientId || Setting.inIframe()) {
      return;
    }
    let cancelled = false;

    loadGsiScript()
      .then(() => {
        if (cancelled) {
          return;
        }
        const google = (window as any).google;
        google.accounts.id.initialize({
          client_id: clientId,
          callback: (response: any) => {
            const code = "GoogleIdToken-" + JSON.stringify(response);
            const authUrlParams = new URLSearchParams(Provider.getAuthUrl(application, providerItem.provider, "signup"));
            const state = authUrlParams.get("state");
            const redirectUri = authUrlParams.get("redirect_uri");
            Setting.goToLink(`${redirectUri}?state=${state}&code=${encodeURIComponent(code)}`);
          },
        });
        google.accounts.id.prompt();
      })
      .catch((error: Error) => Setting.showMessage("error", error.message));

    return () => {
      cancelled = true;
      (window as any).google?.accounts?.id?.cancel?.();
    };
  }, [clientId, application, providerItem]);

  return null;
}
