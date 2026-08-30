import i18next from "i18next";
import {useParams} from "react-router-dom";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {InvoiceActions, checkInvoiceError} from "@/components/product/InvoiceActions";
import {useAccount} from "@/hooks/use-account";
import {useOrganizationOptions, useProviderOptions, useUserNameOptions} from "@/hooks/use-options";
import * as PaymentBackend from "@/backend/PaymentBackend";
import * as Setting from "@/lib/setting";

const STATES = ["Paid", "Created", "Canceled", "Timeout", "Error"];

/** every invoice field freezes once the invoice has been issued */
const issued = (ctx: {record: any}) => !!ctx.record.invoiceUrl;

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
    {type: "text", name: "productsDisplayName", labelKey: "general:Products"},
    {type: "text", name: "detail", labelKey: "general:Detail"},
    {type: "text", name: "tag", labelKey: "general:Tag"},
    {
      type: "select",
      name: "currency",
      labelKey: "payment:Currency",
      options: () => (Setting.CurrencyOptions as any[]).map((item) => ({value: item.id, label: item.name})),
    },
    {type: "number", name: "price", labelKey: "order:Price", step: "0.01"},
    {type: "text", name: "payUrl", labelKey: "general:URL"},
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
      disabled: issued,
      // an individual invoice is always made out to the payer, with no tax ID
      onChange: (value, ctx, updateFields) =>
        updateFields(
          value === "Individual"
            ? {invoiceType: value, invoiceTitle: ctx.record.personName, invoiceTaxId: ""}
            : {invoiceType: value},
        ),
    },
    {
      type: "text",
      name: "invoiceTitle",
      labelKey: "payment:Invoice title",
      disabled: (ctx) => issued(ctx) || ctx.record.invoiceType === "Individual",
    },
    {
      type: "text",
      name: "invoiceTaxId",
      labelKey: "payment:Invoice tax ID",
      disabled: (ctx) => issued(ctx) || ctx.record.invoiceType === "Individual",
    },
    {type: "text", name: "invoiceRemark", labelKey: "payment:Invoice remark", disabled: issued},
    {
      type: "text",
      name: "personName",
      labelKey: "payment:Person name",
      disabled: issued,
      onChange: (value, ctx, updateFields) =>
        updateFields(
          ctx.record.invoiceType === "Individual"
            ? {personName: value, invoiceTitle: value, invoiceTaxId: ""}
            : {personName: value},
        ),
    },
    {type: "text", name: "personIdCard", labelKey: "payment:Person ID card", disabled: issued},
    {type: "email", name: "personEmail", labelKey: "payment:Person Email", disabled: issued},
    {type: "text", name: "personPhone", labelKey: "payment:Person phone", disabled: issued},
    {
      type: "custom",
      name: "invoiceActions",
      labelKey: "payment:Invoice actions",
      block: true,
      render: (ctx) => (
        <InvoiceActions payment={ctx.record} isAdd={ctx.mode === "add"} onIssued={ctx.reload} />
      ),
    },
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
      beforeSave={(record) => {
        // antd refuses to save a payment whose invoice details do not validate
        const errorText = checkInvoiceError(record);
        if (errorText !== "") {
          Setting.showMessage("error", errorText);
          return null;
        }
        return record;
      }}
      editUrl={(record) => `/payments/${record.owner}/${record.name}`}
    />
  );
}
