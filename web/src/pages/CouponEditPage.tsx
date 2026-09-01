import i18next from "i18next";
import {useNavigate, useParams} from "react-router-dom";
import {ConfirmButton} from "@/components/common/ConfirmButton";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useOrganizationOptions, useProductOptions, useUserNameOptions} from "@/hooks/use-options";
import * as CouponBackend from "@/backend/CouponBackend";
import {COUPON_DISCOUNT_TYPES, COUPON_SCOPES, COUPON_STATES, enumOptions} from "@/lib/enum-labels";
import * as Setting from "@/lib/setting";


export default function CouponEditPage() {
  const navigate = useNavigate();
  const {organizationName = "", couponName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();
  const products = useProductOptions(organizationName);
  const users = useUserNameOptions(organizationName);

  const fields: EditField[] = [
    {
      type: "select",
      name: "owner",
      labelKey: "general:Organization",
      options: () => organizations,
      disabled: () => !Setting.isAdminUser(account),
    },
    {type: "text", name: "name", labelKey: "general:Name", required: true},
    {type: "text", name: "displayName", labelKey: "general:Display name"},
    {type: "text", name: "description", labelKey: "general:Description"},
    {type: "text", name: "code", labelKey: "invitation:Code"},
    {
      type: "select",
      name: "discountType",
      labelKey: "coupon:Discount type",
      options: () => enumOptions(COUPON_DISCOUNT_TYPES),
    },
    {type: "number", name: "discount", labelKey: "coupon:Discount", step: "0.01"},
    {type: "number", name: "maxDiscount", labelKey: "coupon:Max discount", step: "0.01"},
    {
      type: "select",
      name: "scope",
      labelKey: "provider:Scope",
      options: () => enumOptions(COUPON_SCOPES),
    },
    {
      type: "multiselect",
      name: "products",
      labelKey: "general:Products",
      when: (ctx) => ctx.record.scope === "product",
      options: () => products,
    },
    {
      type: "multiselect",
      name: "users",
      labelKey: "general:Users",
      when: (ctx) => ctx.record.scope === "user",
      options: () => users,
    },
    {type: "number", name: "quantity", labelKey: "product:Quantity"},
    {type: "number", name: "usedCount", labelKey: "invitation:Used count", disabled: () => true},
    {type: "number", name: "maxUsagePerUser", labelKey: "coupon:Max usage per user"},
    {type: "number", name: "minOrderAmount", labelKey: "coupon:Min order amount", step: "0.01"},
    {
      type: "select",
      name: "currency",
      labelKey: "payment:Currency",
      options: () => (Setting.CurrencyOptions as any[]).map((item) => ({value: item.id, label: item.name})),
    },
    {type: "text", name: "startTime", labelKey: "subscription:Start time"},
    {type: "text", name: "expireTime", labelKey: "general:Expire time"},
    {
      type: "select",
      name: "state",
      labelKey: "general:State",
      options: () => enumOptions(COUPON_STATES),
    },
  ];

  return (
    <SimpleEditPage
      titleKey="coupon:Edit Coupon"
      backTo="/coupons"
      deps={[organizationName, couponName]}
      fields={fields}
      fetch={() => CouponBackend.getCoupon(organizationName, couponName)}
      add={(record) => CouponBackend.addCoupon(record)}
      update={(record) => CouponBackend.updateCoupon(organizationName, couponName, record)}
      editUrl={(record) => `/coupons/${record.owner}/${record.name}`}
      // the antd page keeps a Delete next to Save, so a coupon can be dropped
      // without going back to the list
      extraActions={(ctx) =>
        ctx.mode === "edit" ? (
          <ConfirmButton
            variant="outline"
            description={ctx.record.name}
            onConfirm={() =>
              CouponBackend.deleteCoupon(Setting.getDeleteObj(ctx.record, organizationName, couponName)).then(
                (res: any) => {
                  if (res.status === "ok") {
                    Setting.showMessage("success", i18next.t("general:Successfully deleted"));
                    navigate("/coupons");
                  } else {
                    Setting.showMessage("error", `${i18next.t("general:Failed to delete")}: ${res.msg}`);
                  }
                },
              )
            }
          >
            {i18next.t("general:Delete")}
          </ConfirmButton>
        ) : null
      }
    />
  );
}
