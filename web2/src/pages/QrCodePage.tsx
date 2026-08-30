import * as React from "react";
import i18next from "i18next";
import {QRCodeSVG} from "qrcode.react";
import {useParams, useSearchParams} from "react-router-dom";
import * as PaymentBackend from "@/backend/PaymentBackend";
import * as ProviderBackend from "@/backend/ProviderBackend";
import * as Setting from "@/lib/setting";

/**
 * WeChat Pay QR code shown outside the WeChat browser. Ported from
 * web/src/QrCodePage.js: it polls /api/notify-payment every two seconds and
 * leaves for the success URL as soon as the payment is no longer "Created".
 */
export default function QrCodePage() {
  const params = useParams();
  const [search] = useSearchParams();

  const owner = params.owner ?? "";
  const paymentName = params.paymentName ?? "";
  const providerName = search.get("providerName") ?? "";
  const payUrl = search.get("payUrl") ?? "";
  const successUrl = search.get("successUrl") ?? "";

  const [provider, setProvider] = React.useState<any>(null);

  React.useEffect(() => {
    if (!owner || !providerName) {
      return;
    }
    ProviderBackend.getProvider(owner, providerName).then((res: any) => {
      if (res.status === "ok") {
        setProvider(res.data);
      } else {
        Setting.showMessage("error", res.msg);
      }
    });
  }, [owner, providerName]);

  React.useEffect(() => {
    if (!owner || !paymentName) {
      return;
    }
    let cancelled = false;
    let timer: number | undefined;

    const poll = () => {
      timer = window.setTimeout(async() => {
        try {
          const res = await PaymentBackend.notifyPayment(owner, paymentName);
          if (cancelled) {
            return;
          }
          if (res.status !== "ok") {
            throw new Error(res.msg);
          }
          if (res.data.state !== "Created") {
            Setting.goToLink(successUrl);
            return;
          }
        } catch (err: any) {
          Setting.showMessage("error", err.message);
          return;
        }
        poll();
      }, 2000);
    };

    poll();
    return () => {
      cancelled = true;
      if (timer !== undefined) {
        window.clearTimeout(timer);
      }
    };
  }, [owner, paymentName, successUrl]);

  if (!payUrl || !successUrl || !owner || !paymentName) {
    return null;
  }

  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-6">
      {provider ? (
        <div className="flex h-[50px] items-center gap-3 rounded-full border-2 px-6">
          <img
            src={Setting.getProviderLogoURL(provider)}
            alt={provider.displayName}
            className="h-9 w-9 object-contain"
          />
          <span className="text-lg">{i18next.t(`product:${provider.type}`)}</span>
        </div>
      ) : null}
      <QRCodeSVG value={payUrl} size={200} />
    </div>
  );
}
