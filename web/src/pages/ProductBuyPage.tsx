import * as React from "react";
import i18next from "i18next";
import {useNavigate, useParams, useSearchParams} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Loading} from "@/components/common/Loading";
import {FloatingCartButton, QuantityStepper, useCartItemCount} from "@/components/product/CartControls";
import {FormRow} from "@/components/crud/FormRow";
import {useAccount} from "@/hooks/use-account";
import * as OrderBackend from "@/backend/OrderBackend";
import * as PlanBackend from "@/backend/PlanBackend";
import * as PricingBackend from "@/backend/PricingBackend";
import * as ProductBackend from "@/backend/ProductBackend";
import * as UserBackend from "@/backend/UserBackend";
import * as Setting from "@/lib/setting";

/** WeChat Pay only works inside the WeChat mobile browser. */
export function getPaymentEnv() {
  const ua = navigator.userAgent.toLowerCase();
  return ua.indexOf("micromessenger") !== -1 && ua.indexOf("mobile") !== -1 ? "WechatBrowser" : "";
}

/**
 * "Buy Product": reached either directly (/products/:owner/:name/buy) or through a
 * pricing plan (/buy-plan/:owner/:pricingName?plan=&user=). Ported from
 * web/src/ProductBuyPage.js — the place-order payload is unchanged.
 */
export default function ProductBuyPage() {
  const params = useParams();
  const [search] = useSearchParams();
  const navigate = useNavigate();
  const {account} = useAccount();

  const owner = params.organizationName ?? params.owner ?? "";
  const pricingName = params.pricingName ?? "";
  const planName = search.get("plan");
  const userName = search.get("user");

  const [productName, setProductName] = React.useState(params.productName ?? "");
  const [product, setProduct] = React.useState<any>(null);
  const [customPrice, setCustomPrice] = React.useState(100);
  const [buyQuantity, setBuyQuantity] = React.useState(
    search.get("quantity") ? parseInt(search.get("quantity") as string, 10) : 1,
  );
  const [placingOrder, setPlacingOrder] = React.useState(false);
  const [addingToCart, setAddingToCart] = React.useState(false);
  const [cartItemCount, setCartItemCount] = useCartItemCount(account);

  React.useEffect(() => {
    let cancelled = false;

    const load = async() => {
      if (!owner || (!params.productName && !pricingName)) {
        return;
      }
      try {
        let name = params.productName ?? "";
        if (pricingName) {
          if (!planName || !userName) {
            return;
          }
          const pricingRes = await PricingBackend.getPricing(owner, pricingName);
          if (pricingRes.status !== "ok") {
            throw new Error(pricingRes.msg);
          }
          const planRes = await PlanBackend.getPlan(owner, planName);
          if (planRes.status !== "ok") {
            throw new Error(planRes.msg);
          }
          name = planRes.data.product;
          if (!cancelled) {
            setProductName(name);
          }
        }

        const res = await ProductBackend.getProduct(owner, name);
        if (res.status !== "ok") {
          throw new Error(res.msg);
        }
        if (cancelled) {
          return;
        }
        setProduct(res.data);
        if (res.data.isRecharge) {
          setCustomPrice(res.data.rechargeOptions?.length > 0 ? res.data.rechargeOptions[0] : 100);
        }
      } catch (err: any) {
        Setting.showMessage("error", err.message);
      }
    };

    load();
    return () => {
      cancelled = true;
    };
  }, [owner, params.productName, pricingName, planName, userName]);

  const addToCart = () => {
    if (!account || addingToCart || product === null) {
      return;
    }
    setAddingToCart(true);
    const done = () => setAddingToCart(false);

    UserBackend.getUser(account.owner, account.name)
      .then((res: any) => {
        if (res.status !== "ok") {
          Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${res.msg}`);
          done();
          return;
        }

        const user = res.data;
        const cart = user.cart || [];

        let actualPrice = product.price;
        if (product.isRecharge) {
          actualPrice = customPrice;
          if (actualPrice <= 0) {
            Setting.showMessage("error", i18next.t("product:Custom price should be greater than zero"));
            done();
            return;
          }
        }

        if (cart.length > 0) {
          const firstItem = cart[0];
          if (firstItem.currency && product.currency && firstItem.currency !== product.currency) {
            Setting.showMessage("error", i18next.t("product:The currency of the product you are adding is different from the currency of the items in the cart"));
            done();
            return;
          }
        }

        const existingIndex = cart.findIndex((item: any) =>
          item.name === product.name &&
          (product.isRecharge ? item.price === actualPrice : true) &&
          (item.pricingName || "") === (pricingName || "") &&
          (item.planName || "") === (planName || ""));

        if (existingIndex !== -1) {
          cart[existingIndex].quantity = (cart[existingIndex].quantity ?? 1) + buyQuantity;
        } else {
          cart.push({
            name: product.name,
            createdTime: new Date().toISOString(),
            price: product.isRecharge ? actualPrice : null,
            currency: product.currency,
            pricingName: pricingName || "",
            planName: planName || "",
            quantity: buyQuantity,
          });
        }

        user.cart = cart;
        UserBackend.updateUser(user.owner, user.name, user)
          .then((updateRes: any) => {
            if (updateRes.status === "ok") {
              Setting.showMessage("success", i18next.t("general:Successfully added"));
              setCartItemCount(cart.length);
            } else {
              Setting.showMessage("error", updateRes.msg);
            }
          })
          .catch((error: any) => {
            Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
          })
          .finally(done);
      })
      .catch((error: any) => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
        done();
      });
  };

  const placeOrder = () => {
    if (product === null) {
      return;
    }
    setPlacingOrder(true);

    const productInfos = [{
      name: product.name,
      price: product.isRecharge ? (customPrice || 0) : product.price,
      pricingName: pricingName || "",
      planName: planName || "",
      quantity: buyQuantity,
    }];

    OrderBackend.placeOrder(product.owner, productInfos, userName ?? "")
      .then((res: any) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("product:Order created successfully"));
          navigate(`/orders/${res.data.owner}/${res.data.name}/pay`);
        } else {
          Setting.showMessage("error", `${i18next.t("product:Failed to create order")}: ${res.msg}`);
          setPlacingOrder(false);
        }
      })
      .catch((error: any) => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
        setPlacingOrder(false);
      });
  };

  if (product === null) {
    return <Loading />;
  }

  const hasOptions = product.rechargeOptions && product.rechargeOptions.length > 0;
  const disableCustom = product.disableCustomRecharge;
  const isRechargeUnpurchasable = product.isRecharge && !hasOptions && disableCustom;
  const isAmountZero = product.isRecharge && (customPrice === 0 || customPrice === null);

  const renderRechargeInput = () => {
    if (isRechargeUnpurchasable) {
      return (
        <p className="text-destructive">
          {i18next.t("product:This product is currently not purchasable (No options available)")}
        </p>
      );
    }

    return (
      <div className="space-y-3">
        {hasOptions ? (
          <div className="flex flex-wrap items-center gap-2">
            <span>{i18next.t("product:Select amount")}:</span>
            {product.rechargeOptions.map((amount: number) => (
              <Button
                key={amount}
                variant={customPrice === amount ? "default" : "outline"}
                size="sm"
                onClick={() => setCustomPrice(amount)}
              >
                {Setting.getCurrencySymbol(product.currency)}{amount}
              </Button>
            ))}
          </div>
        ) : null}
        <div className="flex flex-wrap items-center gap-2">
          <span>{i18next.t("product:Amount")}:</span>
          <Input
            type="number"
            min={0}
            className="w-40"
            value={customPrice}
            disabled={disableCustom}
            onChange={(e) => setCustomPrice(Number(e.target.value))}
          />
          <span>{Setting.getCurrencyText(product.currency)}</span>
        </div>
      </div>
    );
  };

  return (
    <div className="mx-auto max-w-4xl space-y-4 py-4">
      <FloatingCartButton itemCount={cartItemCount} />
      <h1 className="text-2xl font-semibold">{i18next.t("product:Buy Product")}</h1>

      <div className="divide-y divide-border/60 rounded-lg border px-4">
        <FormRow labelKey="general:Name">
          <span className="text-lg font-medium">{Setting.getLanguageText(product.displayName)}</span>
        </FormRow>
        <FormRow labelKey="general:Detail">
          <span>{Setting.getLanguageText(product.detail)}</span>
        </FormRow>
        <FormRow labelKey="general:Tag">
          <span>{product.tag}</span>
        </FormRow>
        <FormRow label={i18next.t("product:SKU")}>
          <span>{productName || product.name}</span>
        </FormRow>
        <FormRow label={i18next.t("product:Image")}>
          {product.image ? (
            <img src={product.image} alt={product.name} className="h-[90px] object-contain" />
          ) : null}
        </FormRow>
        {product.isRecharge ? (
          <FormRow label={i18next.t("order:Price")} block>
            {renderRechargeInput()}
          </FormRow>
        ) : (
          <>
            <FormRow label={i18next.t("order:Price")}>
              <span className="text-2xl font-bold text-destructive">
                {`${Setting.getCurrencySymbol(product.currency)}${product.price} (${Setting.getCurrencyText(product.currency)})`}
              </span>
            </FormRow>
            <FormRow label={i18next.t("product:Quantity")}>
              <span>{product.quantity}</span>
            </FormRow>
            <FormRow label={i18next.t("product:Sold")}>
              <span>{product.sold}</span>
            </FormRow>
          </>
        )}
        <FormRow label={i18next.t("general:Place Order")} block>
          {product.state !== "Published" ? (
            <p className="text-muted-foreground">{i18next.t("product:This product is currently not in sale.")}</p>
          ) : (
            <div className="flex flex-wrap items-center justify-center gap-4 py-2">
              <QuantityStepper
                value={buyQuantity}
                disabled={isRechargeUnpurchasable || addingToCart || isAmountZero}
                onIncrease={() => setBuyQuantity((q) => q + 1)}
                onDecrease={() => setBuyQuantity((q) => Math.max(1, q - 1))}
                onChange={(value) => setBuyQuantity(value || 1)}
              />
              <Button
                variant="outline"
                size="lg"
                loading={addingToCart}
                disabled={isRechargeUnpurchasable || isAmountZero}
                onClick={addToCart}
              >
                {i18next.t("product:Add to cart")}
              </Button>
              <Button
                size="lg"
                loading={placingOrder}
                disabled={isRechargeUnpurchasable || isAmountZero}
                onClick={placeOrder}
              >
                {i18next.t("general:Place Order")}
              </Button>
            </div>
          )}
        </FormRow>
      </div>
    </div>
  );
}
