import * as React from "react";
import i18next from "i18next";
import {useLocation} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {Tooltip, TooltipContent, TooltipTrigger} from "@/components/ui/tooltip";
import {WeChatQrDialog, needsWeChatQrDialog} from "@/components/auth/WeChatQrDialog";
import * as AuthBackend from "@/backend/AuthBackend";
import * as Provider from "@/auth/Provider";
import {authViaMetaMask} from "@/auth/Web3Auth";
import * as Setting from "@/lib/setting";

interface ProviderButtonsProps {
  application: any;
  /** "signup" | "signin" | "link" */
  method: "signup" | "signin" | "link";
  /** the item rule: "big" for a labelled button per provider, "small" for a logo grid */
  rule?: string;
  /** return false to swallow the click, e.g. on an unaccepted agreement */
  onBeforeClick?: () => boolean;
}

/** SAML sign-in goes through /api/get-saml-login, which answers with a redirect or a POST form. */
function goToSamlUrl(provider: any, search: string) {
  const params = new URLSearchParams(search);
  const clientId = params.get("client_id") ?? "";
  const state = params.get("state");
  const realRedirectUri = params.get("redirect_uri");
  const redirectUri = `${window.location.origin}/callback/saml`;

  const relayState = `${clientId}&${state}&${provider.name}&${realRedirectUri}&${redirectUri}`;
  AuthBackend.getSamlLogin(`${provider.owner}/${provider.name}`, btoa(relayState)).then((res: any) => {
    if (res.status !== "ok") {
      Setting.showMessage("error", res.msg);
      return;
    }
    if (res.data2 === "POST") {
      document.write(res.data);
    } else {
      window.location.href = res.data;
    }
  });
}

/**
 * The third-party login buttons of an application. The visibility rules and the
 * authorize URL are the ones the antd frontend used, so the round-trip through
 * /callback keeps working unchanged. SAML, Web3 and the WeChat media platform
 * take their own paths, as in web/src/auth/ProviderButton.js.
 */
export function ProviderButtons({application, method, rule, onBeforeClick}: ProviderButtonsProps) {
  const location = useLocation();
  const [wechatItem, setWechatItem] = React.useState<any>(null);

  const items = (application?.providers ?? []).filter((item: any) =>
    method === "signup" ? Setting.isProviderVisibleForSignUp(item) : Setting.isProviderVisibleForSignIn(item),
  );

  const goTo = (providerItem: any) => {
    const provider = providerItem.provider;

    if (provider.category === "SAML") {
      goToSamlUrl(provider, location.search);
      return;
    }
    if (provider.category === "Web3") {
      if (provider.type === "MetaMask") {
        authViaMetaMask(application, provider, method);
      } else {
        // Web3Onboard needs the @web3-onboard wallet modules, which are not ported yet
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${provider.type}`);
      }
      return;
    }
    if (needsWeChatQrDialog(provider)) {
      setWechatItem(providerItem);
      return;
    }

    const url = Provider.getAuthUrl(application, provider, method);
    if (url) {
      Setting.setSigninLanguage(Setting.getLanguage());
      Setting.goToLink(url);
    }
  };

  /**
   * `?provider_hint=<name>` skips the picker and goes straight to that provider.
   * `routers/lightweight_auth_filter.go` serves a static page that makes the same
   * hop without downloading the bundle; this is what happens when that page is
   * unavailable or hands the request back, and it is what the antd login page did
   * inline while rendering the buttons.
   */
  const hint = method === "link" ? null : new URLSearchParams(location.search).get("provider_hint");
  const hintedName = items.find((item: any) => item.provider?.name === hint)?.provider?.name;
  const goToRef = React.useRef(goTo);
  goToRef.current = goTo;

  React.useEffect(() => {
    if (!hintedName) {
      return;
    }
    goToRef.current(items.find((item: any) => item.provider?.name === hintedName));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [hintedName]);

  const onClick = (providerItem: any) => {
    if (onBeforeClick && onBeforeClick() === false) {
      return;
    }
    goTo(providerItem);
  };

  if (items.length === 0) {
    return null;
  }

  const dialog = wechatItem ? (
    <WeChatQrDialog
      application={application}
      provider={wechatItem.provider}
      method={method}
      open={true}
      onClose={() => setWechatItem(null)}
    />
  ) : null;

  // "big" is a labelled button per provider, "small" a grid of logos; without a
  // rule a short list still gets the buttons and a long one the grid.
  const big = rule === "big" || (rule !== "small" && items.length <= 3);
  if (big) {
    return (
      <div className="space-y-2">
        {items.map((item: any) => (
          <Button
            key={item.name}
            type="button"
            variant="outline"
            className="provider-big-img w-full justify-center gap-2"
            onClick={() => onClick(item)}
          >
            <img
              src={Setting.getProviderLogoURL(item.provider)}
              alt={item.provider.displayName}
              className="h-5 w-5 object-contain"
            />
            {i18next.t("login:Sign in with {type}").replace("{type}", item.provider.displayName || item.provider.type)}
          </Button>
        ))}
        {dialog}
      </div>
    );
  }

  return (
    <div className="flex flex-wrap justify-center gap-2">
      {items.map((item: any) => (
        <Tooltip key={item.name}>
          <TooltipTrigger asChild>
            <button
              type="button"
              onClick={() => onClick(item)}
              className="flex h-10 w-10 items-center justify-center rounded-md border transition-colors hover:bg-accent"
            >
              <img
                src={Setting.getProviderLogoURL(item.provider)}
                alt={item.provider.displayName}
                className="provider-img h-6 w-6 object-contain"
              />
            </button>
          </TooltipTrigger>
          <TooltipContent>{item.provider.displayName || item.provider.type}</TooltipContent>
        </Tooltip>
      ))}
      {dialog}
    </div>
  );
}
