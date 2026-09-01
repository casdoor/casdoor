import i18next from "i18next";
import {useParams} from "react-router-dom";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useOrganizationOptions, useProductOptions, useUserNameOptions} from "@/hooks/use-options";
import * as OrderBackend from "@/backend/OrderBackend";
import * as Setting from "@/lib/setting";

const STATES = ["Created", "Paid", "Canceled", "Error"];

export default function OrderEditPage() {
  const {organizationName = "", orderName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();
  const users = useUserNameOptions(organizationName);
  const products = useProductOptions(organizationName);

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
    {type: "multiselect", name: "products", labelKey: "general:Products", options: () => products},
    {type: "select", name: "user", labelKey: "general:User", options: () => users},
    {type: "text", name: "payment", labelKey: "general:Payment"},
    {
      type: "select",
      name: "state",
      labelKey: "general:State",
      options: () => STATES.map((item) => ({value: item, label: i18next.t(`payment:${item}`)})),
    },
    {type: "text", name: "message", labelKey: "payment:Message"},
  ];

  return (
    <SimpleEditPage
      titleKey="order:Edit Order"
      backTo="/orders"
      deps={[organizationName, orderName]}
      fields={fields}
      fetch={() => OrderBackend.getOrder(organizationName, orderName)}
      add={(record) => OrderBackend.addOrder(record)}
      update={(record) => OrderBackend.updateOrder(organizationName, orderName, record)}
      editUrl={(record) => `/orders/${record.owner}/${record.name}`}
    />
  );
}
