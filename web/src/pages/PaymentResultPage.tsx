import * as React from "react";
import i18next from "i18next";
import {AlertTriangle, CheckCircle2, Info, XCircle} from "lucide-react";
import {useNavigate, useParams, useSearchParams} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {Loading} from "@/components/common/Loading";
import {useAccount} from "@/hooks/use-account";
import * as PaymentBackend from "@/backend/PaymentBackend";
import * as PricingBackend from "@/backend/PricingBackend";
import * as SubscriptionBackend from "@/backend/SubscriptionBackend";
import * as UserBackend from "@/backend/UserBackend";
import * as Setting from "@/lib/setting";

/** Payment types the frontend nudges with /api/notify-payment while they are pending. */
const NOTIFIABLE = ["PayPal", "Stripe", "AirWallex", "Alipay", "WeChat Pay", "Balance", "Dummy"];

type Status = "success" | "info" | "warning" | "error";

function Result({status, title, subTitle, extra}: {status: Status; title: React.ReactNode; subTitle?: React.ReactNode; extra?: React.ReactNode}) {
  const Icon = {success: CheckCircle2, info: Info, warning: AlertTriangle, error: XCircle}[status];
  const color = {
    success: "text-emerald-500",
    info: "text-sky-500",
    warning: "text-amber-500",
    error: "text-destructive",
  }[status];

  return (
    <div className="mx-auto flex max-w-2xl flex-col items-center gap-4 py-16 text-center">
      <Icon className={`h-16 w-16 ${color}`} />
      <h1 className="text-xl font-semibold">{title}</h1>
      {subTitle ? <p className="text-muted-foreground">{subTitle}</p> : null}
      {extra ? <div className="flex flex-wrap justify-center gap-2 pt-2">{extra}</div> : null}
    </div>
  );
}

/**
 * The page a payment provider returns to. Ported from web/src/PaymentResultPage.js:
 * while the payment is still "Created" it polls once a second, calling
 * /api/notify-payment first for the providers that need the nudge.
 */
export default function PaymentResultPage() {
  const params = useParams();
  const [search] = useSearchParams();
  const navigate = useNavigate();
  const {account} = useAccount();

  const owner = params.organizationName ?? params.owner ?? "";
  const pricingName = params.pricingName ?? "";
  const subscriptionName = search.get("subscription");

  const [paymentName, setPaymentName] = React.useState(params.paymentName ?? "");
  const [payment, setPayment] = React.useState<any>(null);
  const [user, setUser] = React.useState<any>(null);
  const [nonce, setNonce] = React.useState(0);

  // resolve the payment name from the subscription when arriving from /buy-plan
  React.useEffect(() => {
    if (!owner || !pricingName || !subscriptionName) {
      return;
    }
    let cancelled = false;
    (async() => {
      try {
        const pricingRes = await PricingBackend.getPricing(owner, pricingName);
        if (pricingRes.status !== "ok") {
          throw new Error(pricingRes.msg);
        }
        const subscriptionRes = await SubscriptionBackend.getSubscription(owner, subscriptionName);
        if (subscriptionRes.status !== "ok") {
          throw new Error(subscriptionRes.msg);
        }
        if (!cancelled) {
          setPaymentName(subscriptionRes.data.payment);
        }
      } catch (err: any) {
        Setting.showMessage("error", err.message);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [owner, pricingName, subscriptionName]);

  React.useEffect(() => {
    if (!owner || !paymentName) {
      return;
    }
    let cancelled = false;
    let timer: number | undefined;

    PaymentBackend.getPayment(owner, paymentName)
      .then((res: any) => {
        if (cancelled) {
          return;
        }
        if (res.status !== "ok") {
          Setting.showMessage("error", res.msg);
          return;
        }
        const next = res.data;
        setPayment(next);

        if (next.state === "Created") {
          timer = window.setTimeout(async() => {
            if (NOTIFIABLE.includes(next.type)) {
              await PaymentBackend.notifyPayment(owner, paymentName);
            }
            setNonce((n) => n + 1);
          }, 1000);
        } else if (next.state === "Paid" && account) {
          UserBackend.getUser(account.owner, account.name).then((userRes: any) => {
            if (!cancelled && userRes.status === "ok") {
              setUser(userRes.data);
            }
          });
        }
      });

    return () => {
      cancelled = true;
      if (timer !== undefined) {
        window.clearTimeout(timer);
      }
    };
  }, [owner, paymentName, nonce, account]);

  if (payment === null) {
    return <Loading />;
  }

  const goToViewOrder = () => {
    if (payment.order) {
      navigate(`/orders/${payment.owner}/${payment.order}/pay`);
    } else {
      Setting.showMessage("error", i18next.t("order:Order not found"));
    }
  };

  const actions = (
    <>
      <Button onClick={goToViewOrder}>{i18next.t("order:View Order")}</Button>
      <Button variant="outline" onClick={() => navigate("/orders")}>
        {i18next.t("order:Return to Order List")}
      </Button>
    </>
  );
  const subTitle = i18next.t("payment:You can view your order details or return to the order list");
  const currency = Setting.getCurrencyText(payment.currency);

  if (payment.state === "Paid") {
    if (payment.isRecharge) {
      return (
        <Result
          status="success"
          title={i18next.t("payment:Recharged successfully")}
          subTitle={`${i18next.t("payment:You have successfully recharged")} ${payment.price} ${currency}, ${i18next.t("payment:Your current balance is")} ${user?.balance} ${currency}`}
          extra={actions}
        />
      );
    }
    return (
      <Result
        status="success"
        title={`${i18next.t("payment:You have successfully completed the payment")}: ${payment.productsDisplayName}`}
        subTitle={subTitle}
        extra={actions}
      />
    );
  }

  if (payment.state === "Created") {
    return (
      <Result
        status="info"
        title={`${i18next.t("payment:The payment is still under processing")}: ${payment.productsDisplayName}, ${i18next.t("payment:the current state is")}: ${payment.state}, ${i18next.t("payment:please wait for a few seconds...")}`}
        subTitle={subTitle}
        extra={<Loading />}
      />
    );
  }

  if (payment.state === "Canceled" || payment.state === "Timeout") {
    const title = payment.state === "Canceled"
      ? i18next.t("payment:The payment has been canceled")
      : i18next.t("payment:The payment has timed out");
    return (
      <Result
        status="warning"
        title={`${title}: ${payment.productsDisplayName}, ${i18next.t("payment:the current state is")}: ${payment.state}`}
        subTitle={subTitle}
        extra={actions}
      />
    );
  }

  return (
    <Result
      status="error"
      title={`${i18next.t("payment:The payment has failed")}: ${payment.productsDisplayName}, ${i18next.t("payment:the current state is")}: ${payment.state}`}
      subTitle={`${i18next.t("payment:Failed reason")}: ${payment.message}`}
      extra={actions}
    />
  );
}
