import i18next from "i18next";
import {useParams} from "react-router-dom";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useOrganizationOptions, useProviderOptions, useRoleNameOptions} from "@/hooks/use-options";
import * as PlanBackend from "@/backend/PlanBackend";
import * as Setting from "@/lib/setting";

const PERIODS = ["Monthly", "Yearly"];

export default function PlanEditPage() {
  const {organizationName = "", planName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();
  const roles = useRoleNameOptions(organizationName);
  const paymentProviders = useProviderOptions(organizationName, "Payment");

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
    {type: "number", name: "price", labelKey: "order:Price", step: "0.01"},
    {
      type: "select",
      name: "currency",
      labelKey: "payment:Currency",
      options: () => (Setting.CurrencyOptions as any[]).map((item) => ({value: item.id, label: item.name})),
    },
    {
      type: "select",
      name: "period",
      labelKey: "plan:Period",
      options: () => PERIODS.map((item) => ({value: item, label: i18next.t(`plan:${item}`)})),
    },
    {type: "multiselect", name: "paymentProviders", labelKey: "product:Payment providers", options: () => paymentProviders},
    {type: "select", name: "role", labelKey: "general:Role", options: () => roles},
    {type: "tags", name: "options", labelKey: "signup:Options"},
    {type: "switch", name: "isExclusive", labelKey: "plan:Is exclusive"},
    {type: "switch", name: "isEnabled", labelKey: "general:Is enabled"},
  ];

  return (
    <SimpleEditPage
      titleKey="plan:Edit Plan"
      backTo="/plans"
      deps={[organizationName, planName]}
      fields={fields}
      fetch={() => PlanBackend.getPlan(organizationName, planName)}
      add={(record) => PlanBackend.addPlan(record)}
      update={(record) => PlanBackend.updatePlan(organizationName, planName, record)}
      editUrl={(record) => `/plans/${record.owner}/${record.name}`}
    />
  );
}
