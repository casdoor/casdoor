import {useParams} from "react-router-dom";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useApplicationOptions, useOrganizationOptions, usePlanOptions} from "@/hooks/use-options";
import * as PricingBackend from "@/backend/PricingBackend";
import * as Setting from "@/lib/setting";

export default function PricingEditPage() {
  const {organizationName = "", pricingName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();
  const applications = useApplicationOptions(organizationName);
  const plans = usePlanOptions(organizationName);

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
    {type: "select", name: "application", labelKey: "general:Application", options: () => applications},
    {type: "multiselect", name: "plans", labelKey: "general:Plans", options: () => plans},
    {type: "number", name: "trialDuration", labelKey: "pricing:Trial duration"},
    {type: "switch", name: "isEnabled", labelKey: "general:Is enabled"},
  ];

  return (
    <SimpleEditPage
      titleKey="pricing:Edit Pricing"
      backTo="/pricings"
      deps={[organizationName, pricingName]}
      fields={fields}
      fetch={() => PricingBackend.getPricing(organizationName, pricingName)}
      add={(record) => PricingBackend.addPricing(record)}
      update={(record) => PricingBackend.updatePricing(organizationName, pricingName, record)}
      editUrl={(record) => `/pricings/${record.owner}/${record.name}`}
    />
  );
}
