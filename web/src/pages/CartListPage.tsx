import * as React from "react";
import i18next from "i18next";
import {Link, useNavigate} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table";
import {ConfirmButton} from "@/components/common/ConfirmButton";
import {Loading} from "@/components/common/Loading";
import {QuantityStepper} from "@/components/product/CartControls";
import {PageHeader} from "@/components/crud/PageHeader";
import {useAccount} from "@/hooks/use-account";
import * as OrderBackend from "@/backend/OrderBackend";
import * as ProductBackend from "@/backend/ProductBackend";
import * as UserBackend from "@/backend/UserBackend";
import * as Setting from "@/lib/setting";

interface CartRow {
  name: string;
  displayName?: string;
  image?: string;
  price: number;
  currency?: string;
  quantity: number;
  pricingName?: string;
  planName?: string;
  isRecharge?: boolean;
  isInvalid?: boolean;
  createdTime?: string;
}

/** The cart entry a row belongs to, matched the way web/src/CartListPage.js does. */
function matches(item: any, record: CartRow) {
  return item.name === record.name &&
    (record.isRecharge ? item.price === record.price : true) &&
    (item.pricingName || "") === (record.pricingName || "") &&
    (item.planName || "") === (record.planName || "");
}

export default function CartListPage() {
  const {account} = useAccount();
  const navigate = useNavigate();

  const [rows, setRows] = React.useState<CartRow[]>([]);
  const [user, setUser] = React.useState<any>(null);
  const [loading, setLoading] = React.useState(true);
  const [placingOrder, setPlacingOrder] = React.useState(false);
  const [updatingKeys, setUpdatingKeys] = React.useState<string[]>([]);

  const owner = user?.owner || account?.owner || "";

  const fetch = React.useCallback(() => {
    if (!account) {
      return;
    }
    setLoading(true);
    UserBackend.getUser(account.owner, account.name)
      .then(async(res: any) => {
        if (res.status !== "ok") {
          setLoading(false);
          Setting.showMessage("error", res.msg);
          return;
        }

        const cart = res.data.cart || [];
        const loaded = await Promise.all(cart.map((item: any) =>
          ProductBackend.getProduct(account.owner, item.name)
            .then((productRes: any) => {
              if (productRes.status === "ok" && productRes.data) {
                const currencyChanged = item.currency && productRes.data.currency && item.currency !== productRes.data.currency;
                return {
                  ...productRes.data,
                  createdTime: item.createdTime,
                  pricingName: item.pricingName,
                  planName: item.planName,
                  quantity: item.quantity,
                  price: productRes.data.isRecharge ? item.price : productRes.data.price,
                  isInvalid: currencyChanged,
                };
              }
              return {...item, isInvalid: true};
            })
            .catch(() => ({...item, isInvalid: true}))));

        loaded.sort((a: any, b: any) => (b.createdTime > a.createdTime ? 1 : -1));
        setRows(loaded);
        setUser(res.data);
        setLoading(false);

        loaded.filter((item: any) => item.isInvalid).forEach((item: any) => {
          Setting.showMessage("error", `${i18next.t("product:Product not found or invalid")}: ${item.name}`);
        });
      })
      .catch((error: any) => {
        setLoading(false);
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }, [account]);

  React.useEffect(() => {
    fetch();
  }, [fetch]);

  const saveCart = (nextUser: any, onOk: () => void) => {
    UserBackend.updateUser(nextUser.owner, nextUser.name, nextUser)
      .then((res: any) => {
        if (res.status === "ok") {
          onOk();
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to delete")}: ${res.msg}`);
        }
      })
      .catch((error: any) => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  };

  const clearCart = () => {
    if (!user) {
      Setting.showMessage("error", i18next.t("general:Failed to delete"));
      return;
    }
    const nextUser = Setting.deepCopy(user);
    nextUser.cart = [];
    saveCart(nextUser, () => {
      Setting.showMessage("success", i18next.t("general:Successfully deleted"));
      fetch();
    });
  };

  const deleteCart = (record: CartRow) => {
    if (!user || !Array.isArray(user.cart)) {
      Setting.showMessage("error", i18next.t("general:Failed to delete"));
      return;
    }
    const nextUser = Setting.deepCopy(user);
    const index = nextUser.cart.findIndex((item: any) => matches(item, record));
    if (index === -1) {
      Setting.showMessage("error", i18next.t("general:Failed to delete"));
      return;
    }
    nextUser.cart.splice(index, 1);
    saveCart(nextUser, () => {
      Setting.showMessage("success", i18next.t("general:Successfully deleted"));
      fetch();
    });
  };

  const rowKey = (record: CartRow) =>
    `${record.name}-${record.price ?? "null"}-${record.pricingName || ""}-${record.planName || ""}`;

  const updateQuantity = (record: CartRow, quantity: number) => {
    if (quantity < 1 || !user) {
      return;
    }
    const key = rowKey(record);
    if (updatingKeys.includes(key)) {
      return;
    }

    const nextUser = Setting.deepCopy(user);
    const index = nextUser.cart.findIndex((item: any) => matches(item, record));
    if (index === -1) {
      return;
    }
    nextUser.cart[index].quantity = quantity;

    setUpdatingKeys((prev) => [...prev, key]);
    setRows((prev) => prev.map((item) => (rowKey(item) === key ? {...item, quantity} : item)));

    UserBackend.updateUser(nextUser.owner, nextUser.name, nextUser)
      .then((res: any) => {
        if (res.status === "ok") {
          setUser(nextUser);
        } else {
          Setting.showMessage("error", res.msg);
          fetch();
        }
      })
      .catch((error: any) => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
        fetch();
      })
      .finally(() => setUpdatingKeys((prev) => prev.filter((item) => item !== key)));
  };

  const placeOrder = () => {
    if (placingOrder) {
      return;
    }
    if (rows.some((item) => item.isInvalid)) {
      Setting.showMessage("error", i18next.t("product:Cart contains invalid products, please delete them before placing an order"));
      return;
    }
    if (rows.length === 0) {
      Setting.showMessage("error", i18next.t("product:Product list cannot be empty"));
      return;
    }

    setPlacingOrder(true);
    const productInfos = rows.map((item) => ({
      name: item.name,
      price: item.price,
      quantity: item.quantity,
      pricingName: item.pricingName,
      planName: item.planName,
    }));

    OrderBackend.placeOrder(owner, productInfos, user?.name)
      .then((res: any) => {
        if (res.status === "ok") {
          const nextUser = Setting.deepCopy(user);
          nextUser.cart = [];
          UserBackend.updateUser(nextUser.owner, nextUser.name, nextUser);
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

  const isEmpty = rows.length === 0;
  const hasInvalid = rows.some((item) => item.isInvalid);
  const validRows = rows.filter((item) => !item.isInvalid);
  const total = validRows.reduce((sum, item) => sum + item.price * item.quantity, 0);
  const currency = validRows[0]?.currency ?? rows[0]?.currency ?? "USD";

  return (
    <div className="space-y-4">
      <PageHeader
        title={i18next.t("general:Cart")}
        actions={
          <>
            <Button variant="outline" onClick={() => navigate("/product-store")}>{i18next.t("general:Add")}</Button>
            <ConfirmButton
              variant="outline"
              disabled={isEmpty}
              title={`${i18next.t("general:Sure to delete")}: ${i18next.t("general:Cart")} ?`}
              onConfirm={clearCart}
            >
              {i18next.t("general:Clear")}
            </ConfirmButton>
            <Button loading={placingOrder} disabled={isEmpty || hasInvalid} onClick={placeOrder}>
              {i18next.t("general:Place Order")}
            </Button>
          </>
        }
      />

      {loading ? (
        <Loading />
      ) : (
        <>
          <div className="overflow-x-auto rounded-lg border">
            <Table>
              <TableHeader>
                <TableRow className="hover:bg-transparent">
                  <TableHead className="sticky left-0 z-20 w-[140px] bg-card after:absolute after:inset-y-0 after:-right-px after:w-px after:bg-border">
                    {i18next.t("general:Name")}
                  </TableHead>
                  <TableHead className="w-[170px]">{i18next.t("general:Display name")}</TableHead>
                  <TableHead className="w-[170px]">{i18next.t("product:Image")}</TableHead>
                  <TableHead className="w-[160px]">{i18next.t("order:Price")}</TableHead>
                  <TableHead className="w-[140px]">{i18next.t("pricing:Pricing name")}</TableHead>
                  <TableHead className="w-[140px]">{i18next.t("plan:Plan name")}</TableHead>
                  <TableHead className="w-[130px]">{i18next.t("product:Quantity")}</TableHead>
                  <TableHead className="sticky right-0 z-20 w-[180px] bg-card before:absolute before:inset-y-0 before:-left-px before:w-px before:bg-border">
                    {i18next.t("general:Action")}
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {isEmpty ? (
                  <TableRow className="hover:bg-transparent">
                    <TableCell colSpan={8} className="h-20 text-center text-muted-foreground">
                      {i18next.t("general:No data")}
                    </TableCell>
                  </TableRow>
                ) : (
                  rows.map((record, index) => {
                    const key = rowKey(record);
                    return (
                      <TableRow key={`${key}-${index}`} className="bg-card hover:bg-muted">
                        <TableCell className="sticky left-0 z-20 bg-inherit after:absolute after:inset-y-0 after:-right-px after:w-px after:bg-border">
                          {record.isInvalid ? (
                            <span className="text-destructive">{record.name}</span>
                          ) : (
                            <Link to={`/products/${owner}/${record.name}`} className="underline-offset-4 hover:underline">
                              {record.name}
                            </Link>
                          )}
                        </TableCell>
                        <TableCell>
                          {record.isInvalid ? (
                            <span className="text-destructive">{i18next.t("product:Invalid product")}</span>
                          ) : record.displayName}
                        </TableCell>
                        <TableCell>
                          {record.image ? (
                            <a target="_blank" rel="noreferrer" href={record.image}>
                              <img src={record.image} alt={record.name} className="h-16 object-contain" />
                            </a>
                          ) : null}
                        </TableCell>
                        <TableCell>
                          {Setting.getPriceDisplay((record.price * record.quantity).toFixed(2), record.currency)}
                        </TableCell>
                        <TableCell>
                          {!record.pricingName ? null : record.isInvalid ? (
                            <span className="text-destructive">{record.pricingName}</span>
                          ) : (
                            <Link to={`/pricings/${owner}/${record.pricingName}`} className="underline-offset-4 hover:underline">
                              {record.pricingName}
                            </Link>
                          )}
                        </TableCell>
                        <TableCell>
                          {!record.planName ? null : record.isInvalid ? (
                            <span className="text-destructive">{record.planName}</span>
                          ) : (
                            <Link to={`/plans/${owner}/${record.planName}`} className="underline-offset-4 hover:underline">
                              {record.planName}
                            </Link>
                          )}
                        </TableCell>
                        <TableCell>
                          <QuantityStepper
                            value={record.quantity}
                            disabled={updatingKeys.includes(key) || record.isInvalid}
                            onIncrease={() => updateQuantity(record, record.quantity + 1)}
                            onDecrease={() => updateQuantity(record, record.quantity - 1)}
                          />
                        </TableCell>
                        <TableCell className="sticky right-0 z-20 bg-inherit before:absolute before:inset-y-0 before:-left-px before:w-px before:bg-border">
                          <div className="flex flex-wrap gap-2">
                            <Button
                              size="sm"
                              disabled={record.isInvalid}
                              onClick={() => navigate(`/products/${owner}/${record.name}/buy`)}
                            >
                              {i18next.t("general:Detail")}
                            </Button>
                            <ConfirmButton
                              variant="destructive"
                              size="sm"
                              title={`${i18next.t("general:Sure to delete")}: ${record.name} ?`}
                              onConfirm={() => deleteCart(record)}
                            >
                              {i18next.t("general:Delete")}
                            </ConfirmButton>
                          </div>
                        </TableCell>
                      </TableRow>
                    );
                  })
                )}
              </TableBody>
            </Table>
          </div>

          {!isEmpty ? (
            <div className="flex flex-col items-center gap-4 pt-4">
              <div className="flex items-center text-lg font-semibold">
                {i18next.t("product:Total Price")}:&nbsp;
                <span className="text-2xl text-destructive">
                  {Setting.getCurrencySymbol(currency)}{total.toFixed(2)} ({Setting.getCurrencyText(currency)})
                </span>
              </div>
              <Button size="lg" loading={placingOrder} disabled={hasInvalid} onClick={placeOrder}>
                {i18next.t("general:Place Order")}
              </Button>
            </div>
          ) : null}
        </>
      )}
    </div>
  );
}
