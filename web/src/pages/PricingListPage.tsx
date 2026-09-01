import i18next from "i18next";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {boolColumn, dateColumn, linkColumn, organizationColumn, refsColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import * as Setting from "@/lib/setting";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as PricingBackend from "@/backend/PricingBackend";
import {newPricing} from "@/pages/defaults";

export default function PricingListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();
  // the antd list pages let a non-admin look but not touch these
  const readOnly = !Setting.isLocalAdminUser(account);

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/pricings/${r.owner}/${r.name}`}),
    organizationColumn(),
    dateColumn(),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 170}),
    textColumn({dataIndex: "application", title: i18next.t("general:Application"), width: 160, searchable: true, link: (v, r: any) => `/applications/${r.owner}/${v}`}),
    refsColumn({dataIndex: "plans", title: i18next.t("general:Plans"), urlPrefix: "/plans", width: 220}),
    boolColumn({dataIndex: "isEnabled", title: i18next.t("general:Is enabled")}),
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Pricings")}
      columns={columns}
      deps={[organizationName]}
      fetch={(q) =>
        PricingBackend.getPricings(organizationName, q.page, q.pageSize, q.searchedColumn, q.searchText, q.sortField, q.sortOrder)
      }
      newRecord={account ? () => newPricing(account) : undefined}
      readOnly={readOnly}
      editUrl={(r) => `/pricings/${r.owner}/${r.name}`}
      remove={(r) => PricingBackend.deletePricing(r)}
    />
  );
}
