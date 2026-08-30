import * as React from "react";
import i18next from "i18next";
import {useNavigate, useParams} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {Loading} from "@/components/common/Loading";
import {FormRow} from "@/components/crud/FormRow";
import {getPaymentEnv} from "@/pages/ProductBuyPage";
import * as OrderBackend from "@/backend/OrderBackend";
import * as ProductBackend from "@/backend/ProductBackend";
import * as Setting from "@/lib/setting";

const PROVIDER_TEXT: Record<string, string> = {
  "Dummy": "product:Dummy",
  "Balance": "user:Balance",
  "Alipay": "product:Alipay",
  "WeChat Pay": "product:WeChat Pay",
  "PayPal": "product:PayPal",
  "Stripe": "product:Stripe",
  "AirWallex": "product:AirWallex",
};

/**
 * Pays an order: shows what was ordered and one button per payment provider the
 * product supports. Ported from web/src/OrderPayPage.js, including the WeChat
 * JSAPI path inside the WeChat browser and the QR-code redirect outside it.
 */
export default function OrderPayPage() {
  const params = useParams();
  const navigate = useNavigate();
  const owner = params.organizationName ?? params.owner ?? "";
  const orderName = params.orderName ?? "";

  const [order, setOrder] = React.useState<any>(null);
  const [firstProduct, setFirstProduct] = React.useState<any>(null);
  const [processing, setProcessing] = React.useState(false);
  const paymentEnv = React.useMemo(() => getPaymentEnv(), []);

  React.useEffect(() => {
    if (!owner || !orderName) {
      return;
    }
    OrderBackend.getOrder(owner, orderName).then((res: any) => {
      if (res.status !== "ok") {
        Setting.showMessage("error", res.msg);
        return;
      }
      setOrder(res.data);

      const firstProductName = res.data?.products?.[0] ?? res.data?.productInfos?.[0]?.name;
      if (!firstProductName) {
        return;
      }
      ProductBackend.getProduct(res.data.owner, firstProductName).then((productRes: any) => {
        if (productRes.status === "ok") {
          setFirstProduct(productRes.data);
        } else {
          Setting.showMessage("error", productRes.msg);
        }
      });
    });
  }, [owner, orderName]);

  // WeChat Pay inside the WeChat browser goes through the JSAPI bridge
  const callWechatPay = (attachInfo: any) => {
    const invoke = () => {
      setProcessing(false);
      (window as any).WeixinJSBridge.invoke(
        "getBrandWCPayRequest",
        {
          appId: attachInfo.appId,
          timeStamp: attachInfo.timeStamp,
          nonceStr: attachInfo.nonceStr,
          package: attachInfo.package,
          signType: attachInfo.signType,
          paySign: attachInfo.paySign,
        },
        (res: any) => {
          if (res.err_msg === "get_brand_wcpay_request:ok") {
            Setting.goToLink(attachInfo.payment.successUrl);
          } else if (res.err_msg === "get_brand_wcpay_request:cancel") {
            Setting.showMessage("error", i18next.t("product:Payment cancelled"));
          } else {
            Setting.showMessage("error", i18next.t("product:Payment failed"));
          }
        },
      );
    };

    if (typeof (window as any).WeixinJSBridge === "undefined") {
      document.addEventListener("WeixinJSBridgeReady", invoke, false);
    } else {
      invoke();
    }
  };

  const payOrder = (provider: any) => {
    if (!firstProduct || !order) {
      return;
    }
    setProcessing(true);

    OrderBackend.payOrder(order.owner, order.name, provider.name, paymentEnv)
      .then((res: any) => {
        if (res.status !== "ok") {
          Setting.showMessage("error", `${i18next.t("product:Payment failed")}: ${res.msg}`);
          setProcessing(false);
          return;
        }

        const payment = res.data;
        const attachInfo = res.data2;
        let payUrl = payment.payUrl;

        if (provider.type === "WeChat Pay") {
          if (paymentEnv === "WechatBrowser") {
            attachInfo.payment = payment;
            callWechatPay(attachInfo);
            return;
          }
          payUrl = `/qrcode/${payment.owner}/${payment.name}?providerName=${provider.name}&payUrl=${encodeURIComponent(payment.payUrl)}&successUrl=${encodeURIComponent(payment.successUrl)}`;
        }
        Setting.goToLink(payUrl);
      })
      .catch((error: any) => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
        setProcessing(false);
      });
  };

  if (order === null || !order.productInfos) {
    return <Loading />;
  }

  const isViewMode = order.state !== "Created";
  const updateTimeLabels: Record<string, string> = {
    Paid: "order:Payment time",
    Canceled: "order:Cancel time",
    Failed: "order:Payment failed time",
    Timeout: "order:Timeout time",
  };
  const updateTimeLabel = i18next.t(updateTimeLabels[order.state] ?? "general:Updated time");
  const showUpdateTime = order.state !== "Created" && (order.updateTime ?? "") !== "";

  const productPrice = (product: any) => {
    const price = product.price * (product.quantity ?? 1);
    return `${Setting.getCurrencySymbol(order.currency)}${price.toFixed(2)} (${Setting.getCurrencyText(order.currency)})`;
  };

  return (
    <div className="mx-auto max-w-4xl space-y-6 py-4">
      <div className="space-y-2">
        <h1 className="text-2xl font-semibold">{i18next.t("application:Order")}</h1>
        <div className="divide-y divide-border/60 rounded-lg border px-4">
          <FormRow labelKey="general:ID"><span>{order.name}</span></FormRow>
          <FormRow label={i18next.t("general:Status")}><span>{order.state}</span></FormRow>
          <FormRow labelKey="general:Created time">
            <span>{Setting.getFormattedDate(order.createdTime)}</span>
          </FormRow>
          {showUpdateTime ? (
            <FormRow label={updateTimeLabel}>
              <span>{Setting.getFormattedDate(order.updateTime)}</span>
            </FormRow>
          ) : null}
          <FormRow labelKey="general:User"><span>{order.user}</span></FormRow>
        </div>
      </div>

      <div className="space-y-2">
        <h2 className="text-xl font-semibold">{i18next.t("product:Information")}</h2>
        {order.productInfos.map((product: any) => (
          <div key={product.name} className="divide-y divide-border/60 rounded-lg border px-4">
            <FormRow labelKey="general:Name">
              <span className="text-lg font-medium">{Setting.getLanguageText(product.displayName)}</span>
            </FormRow>
            <FormRow label={i18next.t("product:Image")}>
              {product.image ? (
                <img src={product.image} alt={product.name} className="h-[90px] object-contain" />
              ) : null}
            </FormRow>
            <FormRow label={i18next.t("order:Price")}>
              <span className="font-semibold">{productPrice(product)}</span>
            </FormRow>
            <FormRow label={i18next.t("product:Quantity")}>
              <span>{product.quantity ?? 1}</span>
            </FormRow>
            {product.detail ? (
              <FormRow labelKey="general:Detail">
                <span>{Setting.getLanguageText(product.detail)}</span>
              </FormRow>
            ) : null}
            {product.pricingName && product.planName ? (
              <>
                <FormRow label={i18next.t("subscription:Subscription plan")}>
                  <span>{Setting.getLanguageText(product.planName)}</span>
                </FormRow>
                <FormRow label={i18next.t("subscription:Subscription pricing")}>
                  <span>{Setting.getLanguageText(product.pricingName)}</span>
                </FormRow>
              </>
            ) : null}
          </div>
        ))}
      </div>

      <div className="space-y-2">
        <h2 className="text-xl font-semibold">{i18next.t("general:Payment")}</h2>
        <div className="divide-y divide-border/60 rounded-lg border px-4">
          <FormRow label={i18next.t("order:Price")}>
            <span className="text-2xl font-bold text-destructive">
              {`${Setting.getCurrencySymbol(order.currency)}${order.price} (${Setting.getCurrencyText(order.currency)})`}
            </span>
          </FormRow>
          {!isViewMode ? (
            <FormRow label={i18next.t("order:Pay")} block>
              {!firstProduct?.providerObjs?.length ? (
                <p className="text-muted-foreground">{i18next.t("product:There is no payment channel for this product.")}</p>
              ) : (
                <div className="flex flex-wrap gap-3">
                  {firstProduct.providerObjs.map((provider: any) => (
                    <Button
                      key={provider.name}
                      variant="outline"
                      size="lg"
                      className="h-[50px] rounded-full"
                      loading={processing}
                      onClick={() => payOrder(provider)}
                    >
                      <img
                        src={Setting.getProviderLogoURL(provider)}
                        alt={provider.displayName}
                        className="!h-8 !w-8 object-contain"
                      />
                      {i18next.t(PROVIDER_TEXT[provider.type] ?? provider.type)}
                    </Button>
                  ))}
                </div>
              )}
            </FormRow>
          ) : null}
        </div>
      </div>

      <div>
        <Button variant="outline" onClick={() => navigate("/orders")}>
          {i18next.t("order:Return to Order List")}
        </Button>
      </div>
    </div>
  );
}
