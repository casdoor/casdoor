import i18next from "i18next";
import {Badge} from "@/components/ui/badge";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {linkColumn, organizationColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as CouponBackend from "@/backend/CouponBackend";
import * as Setting from "@/lib/setting";
import {newCoupon} from "@/pages/defaults";

export default function CouponListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/coupons/${r.owner}/${r.name}`}),
    organizationColumn(),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 170}),
    textColumn({dataIndex: "code", title: i18next.t("invitation:Code"), width: 150, mono: true}),
    textColumn({dataIndex: "discountType", title: i18next.t("coupon:Discount type"), width: 140}),
    {
      dataIndex: "discount",
      title: i18next.t("coupon:Discount"),
      width: 120,
      sortable: true,
      render: (value, record) =>
        record.discountType === "percentage" ? `${value}%` : Setting.getPriceDisplay(value, record.currency),
    },
    textColumn({dataIndex: "scope", title: i18next.t("provider:Scope"), width: 120}),
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
    {
      dataIndex: "state",
      title: i18next.t("general:State"),
      width: 110,
      sortable: true,
      render: (value) => <Badge variant={value === "Active" ? "success" : "secondary"}>{value}</Badge>,
    },
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
      editUrl={(r) => `/coupons/${r.owner}/${r.name}`}
      remove={(r) => CouponBackend.deleteCoupon(r)}
    />
  );
}
