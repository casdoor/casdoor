import * as React from "react";
import {useLocation, useParams} from "react-router-dom";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useApplicationOptions, useOrganizationOptions, useUserNameOptions} from "@/hooks/use-options";
import * as TransactionBackend from "@/backend/TransactionBackend";
import * as Setting from "@/lib/setting";

const STATES = ["Paid", "Created", "Canceled", "Error"];

export default function TransactionEditPage() {
  const {organizationName = "", transactionName = ""} = useParams();
  const location = useLocation();
  const {account} = useAccount();
  // "Recharge" on the list page opens the row it just added under its own title
  const isRecharge = Boolean((location.state as any)?.recharge);
  // the application and user pickers follow the organization the form currently holds
  const [owner, setOwner] = React.useState(organizationName);
  const organizations = useOrganizationOptions();
  const applications = useApplicationOptions(owner);
  const users = useUserNameOptions(owner);

  const fields: EditField[] = [
    {
      type: "select",
      name: "owner",
      labelKey: "general:Organization",
      options: () => organizations,
      disabled: () => !Setting.isAdminUser(account),
      onChange: (value, _ctx, updateFields) => {
        setOwner(value);
        // the old application belongs to the old organization
        updateFields({owner: value, application: ""});
      },
    },
    {type: "text", name: "name", labelKey: "general:Name"},
    {type: "text", name: "displayName", labelKey: "general:Display name"},
    {type: "select", name: "application", labelKey: "general:Application", options: () => applications},
    {type: "text", name: "domain", labelKey: "provider:Domain"},
    {type: "text", name: "category", labelKey: "general:Category"},
    {type: "text", name: "type", labelKey: "general:Type"},
    {type: "text", name: "subtype", labelKey: "provider:Subtype"},
    {type: "text", name: "provider", labelKey: "general:Provider"},
    {type: "select", name: "user", labelKey: "general:User", options: () => users},
    {type: "text", name: "tag", labelKey: "general:Tag"},
    {type: "number", name: "amount", labelKey: "product:Amount", step: "0.01"},
    {
      type: "select",
      name: "currency",
      labelKey: "payment:Currency",
      options: () => (Setting.CurrencyOptions as any[]).map((item) => ({value: item.id, label: item.name})),
    },
    {type: "text", name: "payment", labelKey: "general:Payment"},
    {
      type: "select",
      name: "state",
      labelKey: "general:State",
      options: () => STATES.map((item) => ({value: item, label: item})),
    },
  ];

  return (
    <SimpleEditPage
      titleKey={isRecharge ? "transaction:Recharge" : "transaction:Edit Transaction"}
      backTo="/transactions"
      deps={[organizationName, transactionName]}
      fields={fields}
      fetch={() => TransactionBackend.getTransaction(organizationName, transactionName)}
      add={(record) => TransactionBackend.addTransaction(record)}
      update={(record) => TransactionBackend.updateTransaction(organizationName, transactionName, record)}
      editUrl={(record) => `/transactions/${record.owner}/${record.name}`}
    />
  );
}
