import i18next from "i18next";
import copy from "copy-to-clipboard";
import {QRCodeSVG} from "qrcode.react";
import {Alert, AlertDescription} from "@/components/ui/alert";
import {Button} from "@/components/ui/button";
import * as Setting from "@/lib/setting";

/**
 * The deep link the Casdoor Authenticator app scans to take over the account's
 * MFA entries. Port of web/src/common/CasdoorAppConnector.js.
 */
export function generateCasdoorAppUrl(accessToken: string | undefined, forQrCode = true) {
  if (!accessToken) {
    return {qrUrl: "", error: i18next.t("general:Access token is empty")};
  }

  const qrUrl = `casdoor-authenticator://login?serverUrl=${window.location.origin}&accessToken=${accessToken}`;
  // a QR code that large stops being scannable, so the URL is offered instead
  if (forQrCode && qrUrl.length >= 2000) {
    return {qrUrl: "", error: i18next.t("general:QR code is too large")};
  }

  return {qrUrl, error: null as string | null};
}

export function CasdoorAppQrCode({accessToken, icon}: {accessToken?: string; icon?: string}) {
  const {qrUrl, error} = generateCasdoorAppUrl(accessToken, true);

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertDescription>{error}</AlertDescription>
      </Alert>
    );
  }

  return (
    <div className="flex justify-center rounded-md bg-white p-3">
      <QRCodeSVG
        value={qrUrl}
        size={230}
        level="M"
        imageSettings={icon ? {src: icon, height: 40, width: 40, excavate: true} : undefined}
      />
    </div>
  );
}

export function CasdoorAppUrl({accessToken}: {accessToken?: string}) {
  const {qrUrl, error} = generateCasdoorAppUrl(accessToken, false);

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertDescription>{error}</AlertDescription>
      </Alert>
    );
  }

  return (
    <div className="space-y-2">
      {/* the clipboard API is only available over HTTPS, so hide the button otherwise */}
      {window.isSecureContext ? (
        <Button
          size="sm"
          onClick={() => {
            copy(qrUrl);
            Setting.showMessage("success", i18next.t("general:Copied to clipboard successfully"));
          }}
        >
          {i18next.t("resource:Copy Link")}
        </Button>
      ) : null}
      <div className="max-h-24 select-all overflow-auto break-all rounded-md bg-muted p-2 font-mono text-xs">
        {qrUrl}
      </div>
    </div>
  );
}
