import i18next from "i18next";
import {useParams} from "react-router-dom";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useOrganizationOptions, usePlanOptions, useUserNameOptions} from "@/hooks/use-options";
import * as SubscriptionBackend from "@/backend/SubscriptionBackend";
import * as Setting from "@/lib/setting";

const PERIODS = ["Monthly", "Yearly"];
const STATES = ["Pending", "Active", "Upcoming", "Expired", "Error", "Suspended"];

export default function SubscriptionEditPage() {
  const {organizationName = "", subscriptionName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();
  const users = useUserNameOptions(organizationName);
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
    {type: "select", name: "user", labelKey: "general:User", options: () => users},
    {type: "select", name: "plan", labelKey: "general:Plan", options: () => plans},
    {
      type: "select",
      name: "period",
      labelKey: "plan:Period",
      options: () => PERIODS.map((item) => ({value: item, label: i18next.t(`plan:${item}`)})),
    },
    {type: "text", name: "startTime", labelKey: "subscription:Start time"},
    {type: "text", name: "endTime", labelKey: "subscription:End time"},
    {
      type: "select",
      name: "state",
      labelKey: "general:State",
      options: () => STATES.map((item) => ({value: item, label: i18next.t(`subscription:${item}`)})),
    },
  ];

  return (
    <SimpleEditPage
      titleKey="subscription:Edit Subscription"
      backTo="/subscriptions"
      deps={[organizationName, subscriptionName]}
      fields={fields}
      fetch={() => SubscriptionBackend.getSubscription(organizationName, subscriptionName)}
      add={(record) => SubscriptionBackend.addSubscription(record)}
      update={(record) => SubscriptionBackend.updateSubscription(organizationName, subscriptionName, record)}
      editUrl={(record) => `/subscriptions/${record.owner}/${record.name}`}
    />
  );
}
