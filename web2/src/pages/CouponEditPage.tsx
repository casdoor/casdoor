import {useParams} from "react-router-dom";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useOrganizationOptions, useProductOptions, useUserNameOptions} from "@/hooks/use-options";
import * as CouponBackend from "@/backend/CouponBackend";
import * as Setting from "@/lib/setting";

const DISCOUNT_TYPES = ["percentage", "fixed"];
const SCOPES = ["universal", "product", "user"];
const STATES = ["Active", "Suspended"];

export default function CouponEditPage() {
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
    {type: "text", name: "name", labelKey: "general:Name"},
    {type: "text", name: "displayName", labelKey: "general:Display name"},
    {type: "text", name: "description", labelKey: "general:Description"},
    {type: "text", name: "code", labelKey: "invitation:Code"},
    {
      type: "select",
      name: "discountType",
      labelKey: "coupon:Discount type",
      options: () => DISCOUNT_TYPES.map((item) => ({value: item, label: item})),
    },
    {type: "number", name: "discount", labelKey: "coupon:Discount", step: "0.01"},
    {type: "number", name: "maxDiscount", labelKey: "coupon:Max discount", step: "0.01"},
    {
      type: "select",
      name: "scope",
      labelKey: "provider:Scope",
      options: () => SCOPES.map((item) => ({value: item, label: item})),
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
    {type: "number", name: "usedCount", labelKey: "coupon:Used count", disabled: () => true},
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
      options: () => STATES.map((item) => ({value: item, label: item})),
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
    />
  );
}
