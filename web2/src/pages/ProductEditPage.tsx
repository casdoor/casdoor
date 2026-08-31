import i18next from "i18next";
import {useParams} from "react-router-dom";
import {Input} from "@/components/ui/input";
import {EditableTable} from "@/components/crud/EditableTable";
import {Button} from "@/components/ui/button";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useOrganizationOptions, useProviderOptions} from "@/hooks/use-options";
import * as ProductBackend from "@/backend/ProductBackend";
import * as Setting from "@/lib/setting";

const STATES = ["Published", "Draft"];

export default function ProductEditPage() {
  const {organizationName = "", productName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();
  const paymentProviders = useProviderOptions(organizationName, "Payment");

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
    {
      type: "custom",
      name: "image",
      labelKey: "product:Image",
      render: (ctx, update) => (
        <div className="space-y-2">
          <input
            className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            value={ctx.record.image ?? ""}
            onChange={(e) => update("image", e.target.value)}
          />
          {ctx.record.image ? <img src={ctx.record.image} alt="product" className="h-24 object-contain" /> : null}
        </div>
      ),
    },
    {type: "text", name: "detail", labelKey: "general:Detail"},
    {type: "text", name: "description", labelKey: "general:Description"},
    {type: "text", name: "tag", labelKey: "general:Tag"},
    {
      type: "select",
      name: "currency",
      labelKey: "payment:Currency",
      options: () => (Setting.CurrencyOptions as any[]).map((item) => ({value: item.id, label: item.name})),
    },
    {type: "number", name: "price", labelKey: "order:Price", step: "0.01"},
    {type: "number", name: "quantity", labelKey: "product:Quantity"},
    {type: "number", name: "sold", labelKey: "product:Sold"},
    {type: "switch", name: "isRecharge", labelKey: "product:Is recharge"},
    {
      type: "switch",
      name: "disableCustomRecharge",
      labelKey: "product:Disable custom amount",
      when: (ctx) => !!ctx.record.isRecharge,
    },
    {
      type: "tags",
      name: "rechargeOptions",
      labelKey: "product:Recharge options",
      placeholder: i18next.t("product:Enter preset amounts"),
      when: (ctx) => !!ctx.record.isRecharge,
    },
    {type: "text", name: "successUrl", labelKey: "product:Success URL"},
    {
      type: "custom",
      name: "properties",
      labelKey: "user:Properties",
      block: true,
      render: (ctx, update) => (
        <EditableTable
          rows={Object.entries(ctx.record.properties ?? {}).map(([key, value]) => ({key, value}))}
          onChange={(rows) =>
            update(
              "properties",
              Object.fromEntries(rows.filter((row: any) => row.key).map((row: any) => [row.key, row.value])),
            )
          }
          newRow={() => ({key: "", value: ""})}
          reorderable={false}
          columns={[
            {
              key: "key",
              title: i18next.t("general:Name"),
              width: 240,
              render: (row: any, _i, patch) => (
                <Input value={row.key ?? ""} onChange={(e) => patch({key: e.target.value})} />
              ),
            },
            {
              key: "value",
              title: i18next.t("webhook:Value"),
              render: (row: any, _i, patch) => (
                <Input value={row.value ?? ""} onChange={(e) => patch({value: e.target.value})} />
              ),
            },
          ]}
        />
      ),
    },
    {
      type: "multiselect",
      name: "providers",
      labelKey: "product:Payment providers",
      options: () => paymentProviders,
    },
    {
      type: "select",
      name: "state",
      labelKey: "general:State",
      options: () => STATES.map((item) => ({value: item, label: i18next.t(`product:${item}`)})),
    },
  ];

  return (
    <SimpleEditPage
      extraActions={(ctx) =>
        ctx.mode === "add" ? null : (
          <Button variant="outline" asChild>
            <a href={`/products/${ctx.record.owner}/${ctx.record.name}/buy`} target="_blank" rel="noreferrer">
              {i18next.t("product:Test buy page..")}
            </a>
          </Button>
        )
      }
      beforeSave={(product) => {
        // the same three checks the antd page runs before it POSTs
        if (!product.currency) {
          Setting.showMessage("error", i18next.t("product:Please select a currency"));
          return null;
        }
        if (!product.isCreatedByPlan && (!product.providers || product.providers.length === 0)) {
          Setting.showMessage("error", i18next.t("product:Please select at least one payment provider"));
          return null;
        }
        if (product.isRecharge && product.disableCustomRecharge && (!product.rechargeOptions || product.rechargeOptions.length === 0)) {
          Setting.showMessage(
            "error",
            i18next.t("product:Please add at least one recharge option when custom amount is disabled"),
          );
          return null;
        }
        return product;
      }}
      titleKey="product:Edit Product"
      backTo="/products"
      deps={[organizationName, productName]}
      fields={fields}
      fetch={() => ProductBackend.getProduct(organizationName, productName)}
      add={(record) => ProductBackend.addProduct(record)}
      update={(record) => ProductBackend.updateProduct(organizationName, productName, record)}
      editUrl={(record) => `/products/${record.owner}/${record.name}`}
    />
  );
}
