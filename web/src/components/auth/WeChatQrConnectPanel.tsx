import * as React from "react";
import i18next from "i18next";
import {Monitor, QrCode} from "lucide-react";
import * as Provider from "@/auth/Provider";

/**
 * WeChat's own QR code (the open platform website app) in place of the form, for the "Corner"
 * rule of the WeChat sign-in method. The scan comes back to /callback inside the frame, and
 * AuthCallback finishes the sign-in in the page around it.
 */
export function WeChatQrConnectPanel({application, provider}: {application: any; provider: any}) {
  const src = React.useMemo(() => Provider.getWechatQrConnectUrl(application, provider), [application, provider]);

  return (
    <div className="flex flex-col items-center pt-2">
      <h2 className="text-lg font-semibold">{i18next.t("login:Scan to sign in")}</h2>
      <iframe src={src} name={Provider.WECHAT_QR_FRAME_NAME} title={i18next.t("login:Scan to sign in")} className="mt-4 h-[310px] w-full max-w-[300px] border-0" />
    </div>
  );
}

/** The folded top right corner of the panel that switches between the QR code and the form. */
export function QrCornerSwitch({showingQr, onClick}: {showingQr: boolean; onClick: () => void}) {
  const label = showingQr ? i18next.t("login:Sign in with an account") : i18next.t("login:Scan to sign in");

  return (
    <div className="login-qr-corner absolute right-0 top-0 z-10 flex items-start">
      <span className="relative mr-2 mt-3.5 rounded-md border bg-background px-2 py-1 text-xs text-muted-foreground shadow-sm">
        {label}
        <span className="absolute -right-[5px] top-1/2 size-2 -translate-y-1/2 rotate-45 border-r border-t bg-background" />
      </span>
      <button
        type="button"
        onClick={onClick}
        aria-label={label}
        className="group relative size-16 text-foreground/70 transition-colors hover:text-foreground"
        style={{clipPath: "polygon(0 0, 100% 0, 100% 100%)"}}
      >
        <span className="absolute inset-0 bg-foreground/[0.07] transition-colors group-hover:bg-foreground/[0.12]" />
        {showingQr ? <Monitor className="absolute right-2.5 top-2.5 size-7" /> : <QrCode className="absolute right-2.5 top-2.5 size-7" />}
      </button>
    </div>
  );
}
