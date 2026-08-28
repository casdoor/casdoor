import i18next from "i18next";
import {useParams} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useOrganizationOptions, useProviderOptions, useUserNameOptions} from "@/hooks/use-options";
import * as PaymentBackend from "@/backend/PaymentBackend";
import * as Setting from "@/lib/setting";

const STATES = ["Paid", "Created", "Canceled", "Timeout", "Error"];

export default function PaymentEditPage() {
  const {organizationName = "", paymentName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();
  const users = useUserNameOptions(organizationName);
  const providers = useProviderOptions(organizationName, "Payment");

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
    {type: "select", name: "provider", labelKey: "general:Provider", options: () => providers},
    {type: "text", name: "type", labelKey: "general:Type"},
    {type: "select", name: "user", labelKey: "general:User", options: () => users},
    {type: "text", name: "productsDisplayName", labelKey: "payment:Products display name"},
    {type: "text", name: "detail", labelKey: "payment:Detail"},
    {type: "text", name: "tag", labelKey: "general:Tag"},
    {
      type: "select",
      name: "currency",
      labelKey: "payment:Currency",
      options: () => (Setting.CurrencyOptions as any[]).map((item) => ({value: item.id, label: item.name})),
    },
    {type: "number", name: "price", labelKey: "order:Price", step: "0.01"},
    {type: "text", name: "payUrl", labelKey: "payment:Pay url"},
    {type: "text", name: "invoiceUrl", labelKey: "payment:Invoice URL", disabled: () => true},
    {
      type: "select",
      name: "state",
      labelKey: "general:State",
      options: () => STATES.map((item) => ({value: item, label: i18next.t(`payment:${item}`)})),
    },
    {type: "text", name: "message", labelKey: "payment:Message"},
    {
      type: "select",
      name: "invoiceType",
      labelKey: "payment:Invoice type",
      options: () => [
        {value: "Individual", label: i18next.t("payment:Individual")},
        {value: "Organization", label: i18next.t("general:Organization")},
      ],
    },
    {type: "text", name: "invoiceTitle", labelKey: "payment:Invoice title"},
    {type: "text", name: "invoiceTaxId", labelKey: "payment:Invoice tax ID"},
    {type: "text", name: "invoiceRemark", labelKey: "payment:Invoice remark"},
    {type: "text", name: "personName", labelKey: "payment:Person name"},
    {type: "text", name: "personIdCard", labelKey: "payment:Person ID card"},
    {type: "email", name: "personEmail", labelKey: "payment:Person Email"},
    {type: "text", name: "personPhone", labelKey: "payment:Person phone"},
  ];

  return (
    <SimpleEditPage
      titleKey="payment:Edit Payment"
      backTo="/payments"
      deps={[organizationName, paymentName]}
      fields={fields}
      fetch={() => PaymentBackend.getPayment(organizationName, paymentName)}
      add={(record) => PaymentBackend.addPayment(record)}
      update={(record) => PaymentBackend.updatePayment(organizationName, paymentName, record)}
      editUrl={(record) => `/payments/${record.owner}/${record.name}`}
      extraActions={(ctx) => (
        <Button
          variant="outline"
          onClick={() => {
            PaymentBackend.invoicePayment(ctx.record.owner, ctx.record.name).then((res: any) => {
              if (res.status === "ok") {
                Setting.goToLink(res.data);
              } else {
                Setting.showMessage("error", res.msg);
              }
            });
          }}
        >
          {i18next.t("payment:Download Invoice")}
        </Button>
      )}
    />
  );
}
