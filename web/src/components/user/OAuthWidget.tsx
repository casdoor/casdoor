import * as React from "react";
import i18next from "i18next";
import {Button} from "@/components/ui/button";
import {WeChatQrDialog, needsWeChatQrDialog} from "@/components/auth/WeChatQrDialog";
import * as AuthBackend from "@/backend/AuthBackend";
import * as Provider from "@/auth/Provider";
import {authViaMetaMask} from "@/auth/Web3Auth";
import * as Setting from "@/lib/setting";

const BLANK_AVATAR =
  "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAB4AAAAeCAQAAACROWYpAAAAHElEQVR42mNkoAAwjmoe1TyqeVTzqOZRzcNZMwB18wAfEFQkPQAAAABJRU5ErkJggg==";

function getUserProperty(user: any, providerType: string, propertyName: string) {
  if (user.properties === null || user.properties === undefined) {
    return "";
  }
  return user.properties[`oauth_${providerType}_${propertyName}`];
}

function getProviderLink(user: any, provider: any) {
  if (provider.type === "GitHub") {
    return `https://github.com/${getUserProperty(user, provider.type, "username")}`;
  } else if (provider.type === "Google") {
    return "https://mail.google.com";
  }
  return "";
}

interface WidgetProps {
  user: any;
  application: any;
  providerItem: any;
  account: any;
  onUnlinked: () => void;
}

function Row({provider, children}: {provider: any; children: React.ReactNode}) {
  return (
    <div className="flex flex-wrap items-center gap-3 py-2">
      <div className="flex w-40 shrink-0 items-center gap-2">
        {Setting.getProviderLogo(provider)}
        <span className="text-sm">{`${provider.type}:`}</span>
      </div>
      {children}
    </div>
  );
}

/** A SAML provider only shows which user it maps to (web/src/common/SamlWidget.js). */
function SamlWidget({user, providerItem}: {user: any; providerItem: any}) {
  return (
    <Row provider={providerItem.provider}>
      <span className="text-sm">{user.name}</span>
    </Row>
  );
}

/**
 * One third-party account row of the user page: who it is linked to, plus the
 * Link/Unlink buttons. Ported from web/src/common/OAuthWidget.js — the unlink
 * payload and the authorize URL are unchanged.
 */
function OAuthRow({user, application, providerItem, account, onUnlinked}: WidgetProps) {
  const provider = providerItem.provider;
  const [wechatOpen, setWechatOpen] = React.useState(false);

  let linkedValue: string;
  if (provider.type === "Custom Flexible") {
    const link = (user.thirdPartyLinks || []).find((item: any) => item.providerName === provider.name);
    linkedValue = link ? link.providerId : "";
  } else {
    linkedValue = user[provider.type.toLowerCase()];
  }

  const profileUrl = getProviderLink(user, provider);
  const id = getUserProperty(user, provider.type, "id");
  const username = getUserProperty(user, provider.type, "username");
  const displayName = getUserProperty(user, provider.type, "displayName");
  const email = getUserProperty(user, provider.type, "email");
  const avatarUrl = getUserProperty(user, provider.type, "avatarUrl") || BLANK_AVATAR;

  let name = username === undefined ? displayName : `${displayName} (${username})`;
  if (name === undefined) {
    name = id !== undefined ? id : (email !== undefined ? email : linkedValue);
  }

  const isSelf = user.id === account?.id;

  const link = () => {
    if (provider.category === "Web3") {
      if (provider.type === "MetaMask") {
        authViaMetaMask(application, provider, "link");
      }
      return;
    }
    if (needsWeChatQrDialog(provider)) {
      setWechatOpen(true);
      return;
    }
    Setting.goToLink(Provider.getAuthUrl(application, provider, "link"));
  };

  const unlink = () => {
    if (provider.type === "MetaMask" || provider.type === "Web3Onboard") {
      // the signed token is kept per address in localStorage, see web/src/auth/Web3Auth.js
      localStorage.removeItem(`Web3AuthToken_${linkedValue}`);
    }
    AuthBackend.unlink({providerType: provider.type, providerName: provider.name || "", user}).then((res: any) => {
      if (res.status === "ok") {
        Setting.showMessage("success", "Unlinked successfully");
        onUnlinked();
      } else {
        Setting.showMessage("error", `${i18next.t("general:Failed to unlink")}: ${res.msg}`);
      }
    });
  };

  return (
    <Row provider={provider}>
      <img
        src={avatarUrl}
        alt={name}
        referrerPolicy="no-referrer"
        className="h-[30px] w-[30px] rounded object-contain"
      />
      <span className="min-w-0 flex-1 truncate text-sm" title={name}>
        {linkedValue === "" || linkedValue === undefined ? (
          `(${i18next.t("general:empty")})`
        ) : profileUrl === "" ? name : (
          <a target="_blank" rel="noreferrer" href={profileUrl} className="underline-offset-4 hover:underline">
            {name}
          </a>
        )}
      </span>
      {linkedValue === "" || linkedValue === undefined ? (
        <>
          <Button
            size="sm"
            // Web3Onboard needs the @web3-onboard wallet modules, which are not ported yet
            disabled={!isSelf || provider.type === "Web3Onboard"}
            onClick={link}
          >
            {i18next.t("user:Link")}
          </Button>
          {wechatOpen ? (
            <WeChatQrDialog
              application={application}
              provider={provider}
              method="link"
              open={wechatOpen}
              onClose={() => setWechatOpen(false)}
            />
          ) : null}
        </>
      ) : (
        <Button
          variant="outline"
          size="sm"
          disabled={!providerItem.canUnlink && !Setting.isAdminUser(account)}
          onClick={unlink}
        >
          {i18next.t("user:Unlink")}
        </Button>
      )}
    </Row>
  );
}

/** The whole "3rd-party logins" block of the user page. */
export function ThirdPartyLogins({user, application, account, onUnlinked, filter}: Omit<WidgetProps, "providerItem"> & {
  /** the prompt page only lists the providers the application asks to be bound */
  filter?: (providerItem: any) => boolean;
}) {
  const items = (application?.providers ?? []).filter(filter ?? ((item: any) => Setting.isProviderVisible(item)));

  if (items.length === 0) {
    return null;
  }

  return (
    <div className="divide-y divide-border/60">
      {items.map((providerItem: any) =>
        providerItem.provider.category === "OAuth" || providerItem.provider.category === "Web3" ? (
          <OAuthRow
            key={providerItem.name}
            user={user}
            application={application}
            providerItem={providerItem}
            account={account}
            onUnlinked={onUnlinked}
          />
        ) : (
          <SamlWidget key={providerItem.name} user={user} providerItem={providerItem} />
        ),
      )}
    </div>
  );
}
