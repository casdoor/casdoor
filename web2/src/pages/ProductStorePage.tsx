import * as React from "react";
import i18next from "i18next";
import {useNavigate} from "react-router-dom";
import {Badge} from "@/components/ui/badge";
import {Button} from "@/components/ui/button";
import {Card, CardContent} from "@/components/ui/card";
import {Loading} from "@/components/common/Loading";
import {FloatingCartButton, QuantityStepper, useCartItemCount} from "@/components/product/CartControls";
import {PageHeader} from "@/components/crud/PageHeader";
import {useAccount} from "@/hooks/use-account";
import * as ProductBackend from "@/backend/ProductBackend";
import * as UserBackend from "@/backend/UserBackend";
import * as Setting from "@/lib/setting";

/** Max products shown in the store, as in web/src/ProductStorePage.js. */
const PAGE_SIZE = 100;

export default function ProductStorePage() {
  const {account} = useAccount();
  const navigate = useNavigate();

  const [products, setProducts] = React.useState<any[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [addingNames, setAddingNames] = React.useState<string[]>([]);
  const [quantities, setQuantities] = React.useState<Record<string, number>>({});
  const [cartItemCount, setCartItemCount] = useCartItemCount(account);

  React.useEffect(() => {
    if (!account) {
      return;
    }
    const owner = Setting.isDefaultOrganizationSelected(account) ? "" : Setting.getRequestOrganization(account);
    setLoading(true);
    ProductBackend.getProducts(owner, 1, PAGE_SIZE, "state", "Published", "", "")
      .then((res: any) => {
        setLoading(false);
        if (res.status === "ok") {
          setProducts(res.data ?? []);
        } else {
          Setting.showMessage("error", res.msg);
        }
      })
      .catch((error: any) => {
        setLoading(false);
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }, [account]);

  const quantityOf = (product: any) => quantities[product.name] || 1;

  const addToCart = (product: any) => {
    if (!account || addingNames.includes(product.name)) {
      return;
    }
    setAddingNames((prev) => [...prev, product.name]);
    const done = () => setAddingNames((prev) => prev.filter((name) => name !== product.name));

    UserBackend.getUser(account.owner, account.name)
      .then((res: any) => {
        if (res.status !== "ok") {
          Setting.showMessage("error", res.msg);
          done();
          return;
        }

        const user = res.data;
        const cart = user.cart || [];

        if (cart.length > 0) {
          const firstItem = cart[0];
          if (firstItem.currency && product.currency && firstItem.currency !== product.currency) {
            Setting.showMessage("error", i18next.t("product:The currency of the product you are adding is different from the currency of the items in the cart"));
            done();
            return;
          }
        }
        if (product.isRecharge) {
          Setting.showMessage("error", i18next.t("product:Recharge products need to go to the product detail page to set custom amount"));
          done();
          return;
        }

        const quantityToAdd = quantityOf(product);
        const existingIndex = cart.findIndex((item: any) => item.name === product.name);
        if (existingIndex !== -1) {
          cart[existingIndex].quantity = (cart[existingIndex].quantity ?? 1) + quantityToAdd;
        } else {
          cart.push({
            name: product.name,
            createdTime: new Date().toISOString(),
            currency: product.currency,
            pricingName: "",
            planName: "",
            quantity: quantityToAdd,
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

  const buy = (product: any) => {
    navigate(`/products/${product.owner}/${product.name}/buy?quantity=${quantityOf(product)}`);
  };

  return (
    <div className="space-y-4">
      <PageHeader title={i18next.t("general:Product Store")} />
      <FloatingCartButton itemCount={cartItemCount} />

      {loading ? (
        <Loading />
      ) : products.length === 0 ? (
        <p className="py-16 text-center text-muted-foreground">{i18next.t("general:No products available")}</p>
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {products.map((product) => {
            const isAdding = addingNames.includes(product.name);
            const quantity = quantityOf(product);

            return (
              <Card key={`${product.owner}/${product.name}`} className="flex h-full flex-col overflow-hidden">
                <button
                  type="button"
                  onClick={() => buy(product)}
                  className="flex h-48 items-center justify-center bg-muted/50 p-2"
                >
                  {product.image ? (
                    <img src={product.image} alt={product.displayName} className="h-full w-full object-contain" />
                  ) : null}
                </button>
                <CardContent className="flex flex-1 flex-col gap-2 p-4">
                  <h3 className="line-clamp-2 min-h-[2.75rem] font-semibold">
                    {Setting.getLanguageText(product.displayName)}
                  </h3>
                  {product.detail ? (
                    <p className="line-clamp-2 text-sm text-muted-foreground">
                      {Setting.getLanguageText(product.detail)}
                    </p>
                  ) : null}
                  {product.tag ? <div><Badge variant="secondary">{product.tag}</Badge></div> : null}

                  <div className="mt-auto space-y-2 pt-2">
                    {product.isRecharge ? (
                      <div className="space-y-1">
                        {product.rechargeOptions?.length > 0 ? (
                          <>
                            <p className="text-xs text-muted-foreground">{i18next.t("product:Recharge options")}:</p>
                            <div className="flex flex-wrap items-center gap-1">
                              {product.rechargeOptions.map((amount: number) => (
                                <Badge key={amount} variant="secondary">
                                  {Setting.getCurrencySymbol(product.currency)}{amount}
                                </Badge>
                              ))}
                              <span className="text-xs text-muted-foreground">
                                {Setting.getCurrencyWithFlag(product.currency)}
                              </span>
                            </div>
                          </>
                        ) : null}
                        {product.disableCustomRecharge !== true ? (
                          <p className="text-sm font-medium text-primary">
                            {i18next.t("product:Custom amount available")}
                            {!product.rechargeOptions?.length ? (
                              <span className="ml-2 text-xs font-normal text-muted-foreground">
                                {Setting.getCurrencyWithFlag(product.currency)}
                              </span>
                            ) : null}
                          </p>
                        ) : null}
                        {!product.rechargeOptions?.length && product.disableCustomRecharge === true ? (
                          <p className="text-xs text-muted-foreground">
                            {i18next.t("product:No recharge options available")} · {Setting.getCurrencyWithFlag(product.currency)}
                          </p>
                        ) : null}
                      </div>
                    ) : (
                      <div>
                        <span className="text-2xl font-semibold text-destructive">
                          {Setting.getCurrencySymbol(product.currency)}{product.price}
                        </span>
                        <span className="ml-2 text-xs text-muted-foreground">
                          {Setting.getCurrencyWithFlag(product.currency)}
                        </span>
                        <p className="text-xs text-muted-foreground">
                          {i18next.t("product:Sold")}: {product.sold}
                        </p>
                      </div>
                    )}

                    <div className="flex flex-wrap items-center gap-2">
                      {!product.isRecharge ? (
                        <>
                          <QuantityStepper
                            value={quantity}
                            disabled={isAdding}
                            onIncrease={() => setQuantities((prev) => ({...prev, [product.name]: quantity + 1}))}
                            onDecrease={() => setQuantities((prev) => ({...prev, [product.name]: Math.max(1, quantity - 1)}))}
                            onChange={(value) => setQuantities((prev) => ({...prev, [product.name]: value || 1}))}
                          />
                          <Button variant="outline" loading={isAdding} onClick={() => addToCart(product)}>
                            {i18next.t("product:Add to cart")}
                          </Button>
                        </>
                      ) : null}
                      <Button onClick={() => buy(product)}>{i18next.t("product:Buy")}</Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}
    </div>
  );
}
