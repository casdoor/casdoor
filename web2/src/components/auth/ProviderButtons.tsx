import * as React from "react";
import i18next from "i18next";
import {Button} from "@/components/ui/button";
import {Tooltip, TooltipContent, TooltipTrigger} from "@/components/ui/tooltip";
import * as Provider from "@/auth/Provider";
import * as Setting from "@/lib/setting";

interface ProviderButtonsProps {
  application: any;
  /** "signup" | "signin" | "link" */
  method: "signup" | "signin" | "link";
}

/**
 * The third-party login buttons of an application. The visibility rules and the
 * authorize URL are the ones the antd frontend used, so the round-trip through
 * /callback keeps working unchanged.
 */
export function ProviderButtons({application, method}: ProviderButtonsProps) {
  const items = (application?.providers ?? []).filter((item: any) =>
    method === "signup" ? Setting.isProviderVisibleForSignUp(item) : Setting.isProviderVisibleForSignIn(item),
  );

  if (items.length === 0) {
    return null;
  }

  const goTo = (providerItem: any) => {
    const url = Provider.getAuthUrl(application, providerItem.provider, method);
    if (url) {
      Setting.setSigninLanguage(Setting.getLanguage());
      Setting.goToLink(url);
    }
  };

  // A short list gets full-width buttons, a long one gets a grid of logos.
  if (items.length <= 3) {
    return (
      <div className="space-y-2">
        {items.map((item: any) => (
          <Button
            key={item.name}
            type="button"
            variant="outline"
            className="w-full justify-center gap-2"
            onClick={() => goTo(item)}
          >
            <img
              src={Setting.getProviderLogoURL(item.provider)}
              alt={item.provider.displayName}
              className="h-5 w-5 object-contain"
            />
            {i18next.t("login:Sign in with {type}").replace("{type}", item.provider.displayName || item.provider.type)}
          </Button>
        ))}
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
              onClick={() => goTo(item)}
              className="flex h-10 w-10 items-center justify-center rounded-md border transition-colors hover:bg-accent"
            >
              <img
                src={Setting.getProviderLogoURL(item.provider)}
                alt={item.provider.displayName}
                className="h-6 w-6 object-contain"
              />
            </button>
          </TooltipTrigger>
          <TooltipContent>{item.provider.displayName || item.provider.type}</TooltipContent>
        </Tooltip>
      ))}
    </div>
  );
}
