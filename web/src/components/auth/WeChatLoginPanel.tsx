import * as React from "react";
import i18next from "i18next";
import {QRCodeSVG} from "qrcode.react";
import * as AuthBackend from "@/backend/AuthBackend";
import * as Util from "@/auth/Util";

/**
 * The WeChat QR panel of the sign-in page, ported from
 * web/src/auth/WeChatLoginPanel.js: it fetches a QR code for the application's
 * WeChat provider and polls the scan event once a second.
 */
export function WeChatLoginPanel({application}: {application: any}) {
  const [qrCode, setQrCode] = React.useState<string | null>(null);
  const [expired, setExpired] = React.useState(false);
  const [nonce, setNonce] = React.useState(0);

  const providerItem = (application?.providers ?? []).find((item: any) => item.provider?.type === "WeChat");

  React.useEffect(() => {
    if (!providerItem) {
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
  }, [providerItem, application, nonce]);

  if (!providerItem) {
    return null;
  }

  return (
    <div className="mx-auto flex flex-col items-center gap-3 pt-4">
      <div className={expired ? "opacity-30" : undefined}>
        <QRCodeSVG value={qrCode ?? " "} size={230} />
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
