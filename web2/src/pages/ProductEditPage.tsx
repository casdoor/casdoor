import i18next from "i18next";
import {useParams} from "react-router-dom";
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
    {type: "text", name: "detail", labelKey: "product:Detail"},
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
