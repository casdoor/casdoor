import i18next from "i18next";
import {useParams} from "react-router-dom";
import copy from "copy-to-clipboard";
import {Copy} from "lucide-react";
import {Button} from "@/components/ui/button";
import PricingPage from "@/pages/PricingPage";
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
    {type: "text", name: "name", labelKey: "general:Name", required: true},
    {type: "text", name: "displayName", labelKey: "general:Display name"},
    {type: "text", name: "description", labelKey: "general:Description"},
    {type: "select", name: "application", labelKey: "general:Application", options: () => applications},
    {type: "multiselect", name: "plans", labelKey: "general:Plans", options: () => plans},
    {type: "number", name: "trialDuration", labelKey: "pricing:Trial duration"},
    {type: "switch", name: "isEnabled", labelKey: "general:Is enabled"},
    {
      // antd renders the real pricing page here, so the plan cards can be checked
      // against the pricing being edited without leaving the form
      type: "custom",
      name: "preview",
      labelKey: "general:Preview",
      block: true,
      when: (ctx) => (ctx.record.plans ?? []).length > 0,
      render: (ctx) => (
        <div className="rounded-lg border p-4">
          <PricingPage pricing={ctx.record} owner={ctx.record.owner} embedded />
        </div>
      ),
    },
  ];

  return (
    <SimpleEditPage
      extraActions={(ctx) => (
        <Button
          variant="outline"
          onClick={() => {
            copy(`${window.location.origin}/select-plan/${ctx.record.owner}/${ctx.record.name}`);
            Setting.showMessage("success", i18next.t("general:Copied to clipboard successfully"));
          }}
        >
          <Copy />
          {i18next.t("pricing:Copy pricing page URL")}
        </Button>
      )}
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
