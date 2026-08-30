import * as React from "react";
import i18next from "i18next";
import {useSearchParams} from "react-router-dom";
import {AuthLayout} from "@/components/auth/AuthLayout";
import * as ProviderBackend from "@/backend/ProviderBackend";
import * as Util from "@/auth/Util";
import * as Setting from "@/lib/setting";

/**
 * Hosts Telegram's own login widget. Ported from web/src/auth/TelegramLogin.js:
 * the bot username comes from the provider's clientId and the widget posts back
 * to /callback with the same state.
 */
export default function TelegramLogin() {
  const [search] = useSearchParams();
  const containerRef = React.useRef<HTMLDivElement>(null);

  React.useEffect(() => {
    const state = search.get("state") ?? "";
    const innerParams = new URLSearchParams(Util.getQueryParamsFromState(state));
    const providerName = innerParams.get("provider") ?? "";

    ProviderBackend.getProvider("admin", providerName).then((res: any) => {
      if (res.status !== "ok") {
        Setting.showMessage("error", `Failed to get provider: ${res.msg}`);
        return;
      }

      const botUsername = res.data.clientId;
      const authUrl = `${window.location.origin}/callback?state=${state}`;
      if (!botUsername || !containerRef.current) {
        return;
      }

      // loaded from the official Telegram domain; the script has no integrity hash to pin
      const script = document.createElement("script");
      script.src = "https://telegram.org/js/telegram-widget.js";
      script.setAttribute("data-telegram-login", botUsername);
      script.setAttribute("data-size", "large");
      script.setAttribute("data-auth-url", authUrl);
      script.setAttribute("data-request-access", "write");
      script.async = true;

      containerRef.current.innerHTML = "";
      containerRef.current.appendChild(script);
    });
  }, [search]);

  return (
    <AuthLayout>
      <div className="space-y-4 text-center">
        <h1 className="flex items-center justify-center gap-2 text-lg font-semibold">
          <img
            src={Setting.getProviderLogoURL({type: "Telegram", category: "OAuth"})}
            alt="Telegram"
            className="h-8 w-8 object-contain"
          />
          {i18next.t("login:Sign in with Telegram")}
        </h1>
        <p className="text-sm text-muted-foreground">
          {i18next.t("login:Click the button below to sign in with Telegram")}
        </p>
        <div ref={containerRef} className="flex justify-center pt-2" />
      </div>
    </AuthLayout>
  );
}
