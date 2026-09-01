import i18next from "i18next";
import {Badge} from "@/components/ui/badge";
import {CrudListPage} from "@/components/crud/CrudListPage";
import {dateColumn, linkColumn, organizationColumn, tagsColumn, textColumn} from "@/components/crud/columns";
import type {ColumnDef} from "@/components/crud/types";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as ProductBackend from "@/backend/ProductBackend";
import * as Setting from "@/lib/setting";
import {newProduct} from "@/pages/defaults";

export default function ProductListPage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();
  // the antd list pages let a non-admin look but not touch these
  const readOnly = !Setting.isLocalAdminUser(account);

  const columns: ColumnDef<any>[] = [
    linkColumn({dataIndex: "name", to: (r) => `/products/${r.owner}/${r.name}`}),
    organizationColumn(),
    dateColumn(),
    textColumn({dataIndex: "displayName", title: i18next.t("general:Display name"), searchable: true, width: 170}),
    {
      dataIndex: "image",
      title: i18next.t("product:Image"),
      width: 150,
      render: (value) =>
        value ? (
          <a href={value} target="_blank" rel="noreferrer">
            <img src={value} alt="product" className="h-8 max-w-[130px] object-contain" />
          </a>
        ) : null,
    },
    textColumn({dataIndex: "tag", title: i18next.t("general:Tag"), width: 130, searchable: true}),
    {
      dataIndex: "price",
      searchable: true,
      title: i18next.t("order:Price"),
      width: 120,
      sortable: true,
      render: (value, record) => Setting.getPriceDisplay(value, record.currency),
    },
    textColumn({dataIndex: "quantity", title: i18next.t("product:Quantity"), width: 110, searchable: true}),
    textColumn({dataIndex: "sold", title: i18next.t("product:Sold"), width: 100, searchable: true}),
    {
      dataIndex: "state",
      searchable: true,
      title: i18next.t("general:State"),
      width: 120,
      sortable: true,
      render: (value) => <Badge variant={value === "Published" ? "success" : "secondary"}>{value}</Badge>,
    },
    tagsColumn({dataIndex: "providers", title: i18next.t("product:Payment providers"), width: 200}),
  ];

  return (
    <CrudListPage
      title={i18next.t("general:Products")}
      columns={columns}
      deps={[organizationName]}
      fetch={(q) =>
        ProductBackend.getProducts(organizationName, q.page, q.pageSize, q.searchedColumn, q.searchText, q.sortField, q.sortOrder)
      }
      newRecord={account ? () => newProduct(account) : undefined}
      readOnly={readOnly}
      editUrl={(r) => `/products/${r.owner}/${r.name}`}
      remove={(r) => ProductBackend.deleteProduct(r)}
      rowActions={(record) => [
        {key: "buy", label: i18next.t("product:Buy"), href: `/products/${record.owner}/${record.name}/buy`},
      ]}
      actionColumnWidth={240}
    />
  );
}
