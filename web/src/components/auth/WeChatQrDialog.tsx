import * as React from "react";
import i18next from "i18next";
import {QRCodeSVG} from "qrcode.react";
import {Dialog, DialogContent, DialogHeader, DialogTitle} from "@/components/ui/dialog";
import * as AuthBackend from "@/backend/AuthBackend";
import * as Util from "@/auth/Util";
import * as Setting from "@/lib/setting";

interface WeChatQrDialogProps {
  application: any;
  provider: any;
  /** "signup" | "signin" | "link" */
  method: string;
  open: boolean;
  onClose: () => void;
}

/**
 * The WeChat official-account QR code, shown when the provider is configured for
 * the media platform on desktop. Ported from `WechatOfficialAccountModal` in
 * web/src/auth/Util.js: it polls the scan event once a second and leaves for the
 * OAuth redirect as soon as the user has scanned.
 */
export function WeChatQrDialog({application, provider, method, open, onClose}: WeChatQrDialogProps) {
  const [qrCode, setQrCode] = React.useState("");

  React.useEffect(() => {
    if (!open) {
      setQrCode("");
      return;
    }

    let timer: number | undefined;
    AuthBackend.getWechatQRCode(`${provider.owner}/${provider.name}`).then((res: any) => {
      if (res.status !== "ok") {
        Setting.showMessage("error", res.msg);
        onClose();
        return;
      }
      setQrCode(res.data);
      timer = window.setInterval(() => {
        Util.getEvent(application, provider, res.data2, method);
      }, 1000);
    });

    return () => {
      if (timer !== undefined) {
        window.clearInterval(timer);
      }
    };
  }, [open, application, provider, method, onClose]);

  return (
    <Dialog open={open} onOpenChange={(next) => (next ? undefined : onClose())}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {i18next.t("provider:Please use WeChat to scan the QR code and follow the official account for sign in")}
          </DialogTitle>
        </DialogHeader>
        <div className="flex justify-center p-5">
          {qrCode ? <QRCodeSVG value={qrCode} size={230} /> : null}
        </div>
      </DialogContent>
    </Dialog>
  );
}

/** The provider needs the media-platform QR only outside the WeChat browser. */
export function needsWeChatQrDialog(provider: any) {
  return provider.type === "WeChat" &&
    provider.clientId2 !== "" &&
    provider.clientSecret2 !== "" &&
    provider.disableSsl === true &&
    !navigator.userAgent.includes("MicroMessenger");
}
