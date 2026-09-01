import * as React from "react";
import i18next from "i18next";
import {useNavigate} from "react-router-dom";
import {Alert, AlertDescription} from "@/components/ui/alert";
import {Button} from "@/components/ui/button";
import {Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle} from "@/components/ui/dialog";
import * as PaymentBackend from "@/backend/PaymentBackend";
import * as Setting from "@/lib/setting";

/**
 * Everything the payment edit page can do with an invoice, ported from
 * `web/src/PaymentEditPage.js`: validate the invoice details, confirm them in a
 * dialog, POST `/api/invoice-payment`, and afterwards offer the issued PDF.
 *
 * `checkError` mirrors the antd version exactly — the backend does not re-check
 * these, and an invoice cannot be withdrawn once issued.
 */
export function checkInvoiceError(payment: any): string {
  if (payment.state !== "Paid") {
    return i18next.t("payment:Please pay the order first!");
  }
  if (!Setting.isValidPersonName(payment.personName)) {
    return i18next.t("signup:Please input your real name!");
  }
  if (!Setting.isValidIdCard(payment.personIdCard)) {
    return i18next.t("signup:Please input the correct ID card number!");
  }
  if (!Setting.isValidEmail(payment.personEmail)) {
    return i18next.t("login:The input is not valid Email!");
  }
  if (!Setting.isValidPhone(payment.personPhone)) {
    return i18next.t("signup:The input is not valid Phone!");
  }

  if (payment.invoiceType === "Individual") {
    if (payment.invoiceTitle !== payment.personName) {
      return i18next.t("signup:The input is not invoice title!");
    }
    if (payment.invoiceTaxId !== "") {
      return i18next.t("signup:The input is not invoice Tax ID!");
    }
  } else {
    if (!Setting.isValidInvoiceTitle(payment.invoiceTitle)) {
      return i18next.t("signup:The input is not invoice title!");
    }
    if (!Setting.isValidTaxId(payment.invoiceTaxId)) {
      return i18next.t("signup:The input is not invoice Tax ID!");
    }
  }

  return "";
}

export function InvoiceActions({payment, isAdd, onIssued}: {payment: any; isAdd: boolean; onIssued: () => void}) {
  const navigate = useNavigate();
  const [open, setOpen] = React.useState(false);
  const [issuing, setIssuing] = React.useState(false);

  const issue = () => {
    setOpen(false);
    setIssuing(true);
    PaymentBackend.invoicePayment(payment.owner, payment.name)
      .then((res: any) => {
        if (res.status === "ok") {
          Setting.showMessage("success", "Successfully invoiced");
          Setting.openLinkSafe(res.data);
          onIssued();
        } else {
          // some providers answer with a Chinese success sentence on a non-ok status
          Setting.showMessage(res.msg?.includes("成功") ? "info" : "error", res.msg);
        }
      })
      .catch((error) => Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`))
      .finally(() => setIssuing(false));
  };

  const goToViewOrder = () => {
    if (payment?.order) {
      navigate(`/orders/${payment.owner}/${payment.order}/pay`);
    } else {
      Setting.showMessage("error", i18next.t("order:Order not found"));
    }
  };

  const rows: [string, React.ReactNode][] = [
    [i18next.t("payment:Person name"), payment?.personName],
    [i18next.t("payment:Person ID card"), payment?.personIdCard],
    [i18next.t("payment:Person Email"), payment?.personEmail],
    [i18next.t("payment:Person phone"), payment?.personPhone],
    [
      i18next.t("payment:Invoice type"),
      payment?.invoiceType === "Individual" ? i18next.t("payment:Individual") : i18next.t("general:Organization"),
    ],
    [i18next.t("payment:Invoice title"), payment?.invoiceTitle],
    [i18next.t("payment:Invoice tax ID"), payment?.invoiceTaxId],
    [i18next.t("payment:Invoice remark"), payment?.invoiceRemark],
  ];

  return (
    <div className="flex flex-wrap gap-2">
      {!payment?.invoiceUrl ? (
        // the invoice is issued for the saved payment, so it needs the payment to exist
        <Button
          disabled={isAdd}
          loading={issuing}
          onClick={() => {
            const errorText = checkInvoiceError(payment);
            if (errorText !== "") {
              Setting.showMessage("error", errorText);
              return;
            }
            setOpen(true);
          }}
        >
          {i18next.t("payment:Issue Invoice")}
        </Button>
      ) : (
        <Button onClick={() => Setting.openLinkSafe(payment.invoiceUrl)}>
          {i18next.t("payment:Download Invoice")}
        </Button>
      )}
      <Button variant="outline" onClick={goToViewOrder}>{i18next.t("order:View Order")}</Button>
      <Button variant="outline" onClick={() => navigate("/orders")}>
        {i18next.t("order:Return to Order List")}
      </Button>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="sm:max-w-[600px]">
          <DialogHeader>
            <DialogTitle>{i18next.t("payment:Confirm your invoice information")}</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            <Alert variant="warning">
              <AlertDescription>
                {i18next.t(
                  "payment:Please carefully check your invoice information. Once the invoice is issued, it cannot be withdrawn or modified.",
                )}
              </AlertDescription>
            </Alert>
            <dl className="grid grid-cols-[minmax(120px,180px)_1fr] gap-x-4 gap-y-2 text-sm">
              {rows.map(([label, value]) => (
                <React.Fragment key={label}>
                  <dt className="text-muted-foreground">{label}</dt>
                  <dd className="break-words">{value || "-"}</dd>
                </React.Fragment>
              ))}
            </dl>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setOpen(false)}>{i18next.t("general:Cancel")}</Button>
            <Button onClick={issue}>{i18next.t("payment:Issue Invoice")}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
