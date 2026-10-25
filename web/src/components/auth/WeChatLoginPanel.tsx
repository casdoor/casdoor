import * as React from "react";
import i18next from "i18next";
import {QRCodeSVG} from "qrcode.react";
import * as AuthBackend from "@/backend/AuthBackend";
import * as Provider from "@/auth/Provider";
import * as Util from "@/auth/Util";

/**
 * The WeChat QR panel of the sign-in page, ported from
 * web/src/auth/WeChatLoginPanel.js: with an official account (Client ID 2) it fetches
 * that account's QR code and polls the scan event once a second; with only the open
 * platform website app it frames WeChat's own QR code, whose scan comes back to
 * /callback inside the frame and is finished by AuthCallback in the page around it.
 */
export function WeChatLoginPanel({application, showHint = false}: {application: any; showHint?: boolean}) {
  const [qrCode, setQrCode] = React.useState<string | null>(null);
  const [expired, setExpired] = React.useState(false);
  const [nonce, setNonce] = React.useState(0);

  const providerItem = (application?.providers ?? []).find((item: any) => item.provider?.type === "WeChat");
  const officialAccount = !!providerItem?.provider?.clientId2;
  const qrConnectUrl = React.useMemo(
    () => (providerItem && !officialAccount ? Provider.getWechatQrConnectUrl(application, providerItem.provider) : ""),
    [providerItem, officialAccount, application],
  );

  React.useEffect(() => {
    if (!providerItem || !officialAccount) {
      return;
    }
    let cancelled = false;
    let timer: number | undefined;

    setQrCode(null);
    setExpired(false);

    AuthBackend.getWechatQRCode(`${providerItem.provider.owner}/${providerItem.provider.name}`)
      .then((res: any) => {
        if (cancelled) {
          return;
        }
        if (res.status === "ok" && res.data) {
          setQrCode(res.data);
          timer = window.setInterval(() => {
            Util.getEvent(application, providerItem.provider, res.data2, "signup");
          }, 1000);
        } else {
          setExpired(true);
        }
      })
      .catch(() => {
        if (!cancelled) {
          setExpired(true);
        }
      });

    return () => {
      cancelled = true;
      if (timer !== undefined) {
        window.clearInterval(timer);
      }
    };
  }, [providerItem, officialAccount, application, nonce]);

  if (!providerItem) {
    return null;
  }

  return (
    <div className="mx-auto flex flex-col items-center gap-3 pt-4">
      {showHint && officialAccount ? (
        <p className="text-center text-sm text-muted-foreground">
          {i18next.t("provider:Please use WeChat to scan the QR code and follow the official account for sign in")}
        </p>
      ) : null}
      {officialAccount ? (
        <div className={expired ? "opacity-30" : undefined}>
          <QRCodeSVG value={qrCode ?? " "} size={230} />
        </div>
      ) : (
        <iframe key={nonce} src={qrConnectUrl} name={Provider.WECHAT_QR_FRAME_NAME} title={i18next.t("login:Scan to sign in")} className="h-[310px] w-full max-w-[300px] border-0" />
      )}
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
