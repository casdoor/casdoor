import i18next from "i18next";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {linkColumn, organizationColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {COUPON_DISCOUNT_TYPES, COUPON_SCOPES, COUPON_STATES, enumColumn} from "@/lib/enum-labels";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as CouponBackend from "@/backend/CouponBackend";
import * as Setting from "@/lib/setting";
import {newCoupon} from "@/pages/defaults";

export default function CouponListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();
  // the antd list pages let a non-admin look but not touch these
  const readOnly = !Setting.isLocalAdminUser(account);

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/coupons/${r.owner}/${r.name}`}),
    organizationColumn(),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 170}),
    textColumn({dataIndex: "code", title: i18next.t("invitation:Code"), width: 150, mono: true, searchable: true}),
    enumColumn({dataIndex: "discountType", title: i18next.t("coupon:Discount type"), map: COUPON_DISCOUNT_TYPES, width: 140}),
    {
      dataIndex: "discount",
      title: i18next.t("coupon:Discount"),
      width: 120,
      sortable: true,
      render: (value, record) =>
        record.discountType === "percentage" ? `${value}%` : Setting.getPriceDisplay(value, record.currency),
    },
    enumColumn({dataIndex: "scope", title: i18next.t("provider:Scope"), map: COUPON_SCOPES, width: 120}),
    {
      dataIndex: "usedCount",
      title: i18next.t("coupon:Usage"),
      width: 120,
      render: (value, record) => `${value ?? 0} / ${record.quantity ?? 0}`,
    },
    {
      dataIndex: "expireTime",
      title: i18next.t("general:Expire time"),
      width: 165,
      sortable: true,
      render: (value) => Setting.getFormattedDate(value),
    },
    enumColumn({dataIndex: "state", title: i18next.t("general:State"), map: COUPON_STATES, searchable: true}),
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Coupons")}
      columns={columns}
      deps={[organizationName]}
      fetch={(q) =>
        CouponBackend.getCoupons(organizationName, q.page, q.pageSize, q.searchedColumn, q.searchText, q.sortField, q.sortOrder)
      }
      newRecord={account ? () => newCoupon(account) : undefined}
      readOnly={readOnly}
      editUrl={(r) => `/coupons/${r.owner}/${r.name}`}
      remove={(r) => CouponBackend.deleteCoupon(r)}
    />
  );
}
