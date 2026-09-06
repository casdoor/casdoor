import * as React from "react";
import i18next from "i18next";
import {Button} from "@/components/ui/button";
import {Checkbox} from "@/components/ui/checkbox";
import {Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle} from "@/components/ui/dialog";
import * as Setting from "@/lib/setting";

/** The application requires the terms of use to be accepted before signing up. */
export function isAgreementRequired(application: any) {
  const item = application?.signupItems?.find((signupItem: any) => signupItem.name === "Agreement");
  if (!item || !item.rule || item.rule === "None") {
    return false;
  }
  return !!item.required;
}

/** "Signin (Default True)" pre-checks the box. */
export function getAgreementDefaultValue(application: any) {
  const item = application?.signupItems?.find((signupItem: any) => signupItem.name === "Agreement");
  return isAgreementRequired(application) && item?.rule === "Signin (Default True)";
}

function fetchTermsOfUse(url: string) {
  return fetch(url, {method: "GET"})
    .then((res) => res.text())
    .catch((error) => {
      Setting.showMessage("error", `${i18next.t("general:Failed to get")}: ${url}, ${error}`);
      return "";
    });
}

/**
 * The terms-of-use dialog, ported from web/src/common/modal/AgreementModal.js —
 * the document is loaded from `application.termsOfUse` and shown in an iframe.
 */
export function AgreementModal({
  application,
  open,
  onOk,
  onCancel,
}: {
  application: any;
  open: boolean;
  onOk: () => void;
  onCancel: () => void;
}) {
  const [doc, setDoc] = React.useState("");

  React.useEffect(() => {
    if (!open || !application?.termsOfUse) {
      return;
    }
    let cancelled = false;
    fetchTermsOfUse(application.termsOfUse).then((text) => {
      if (!cancelled) {
        setDoc(text);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [open, application?.termsOfUse]);

  return (
    <Dialog open={open} onOpenChange={(next) => (next ? undefined : onCancel())}>
      {/* 55vw is a readable column on a desktop but a squeezed one on a phone,
          where the dialog has to take the width it can get */}
      <DialogContent className="max-w-[calc(100vw-2rem)] rounded-lg p-4 sm:max-w-[55vw] sm:p-6">
        <DialogHeader>
          <DialogTitle>{i18next.t("signup:Terms of Use")}</DialogTitle>
        </DialogHeader>
        <iframe title="terms" srcDoc={doc} className="h-[55vh] w-full border-0 sm:h-[60vh]" />
        <DialogFooter>
          <Button variant="outline" className="max-sm:h-11" onClick={onCancel}>{i18next.t("signup:Decline")}</Button>
          <Button className="max-sm:h-11" onClick={onOk}>{i18next.t("signup:Accept")}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

/** The "Accept Terms of Use" checkbox plus its dialog. */
export function AgreementCheckbox({
  application,
  checked,
  onChange,
}: {
  application: any;
  checked: boolean;
  onChange: (checked: boolean) => void;
}) {
  const [open, setOpen] = React.useState(false);

  return (
    <div className="flex items-center gap-2">
      <Checkbox id="agreement" checked={checked} onCheckedChange={(v) => onChange(v === true)} />
      <label htmlFor="agreement" className="text-sm">
        {i18next.t("signup:Accept")}{" "}
        <button
          type="button"
          className="underline underline-offset-4"
          onClick={() => setOpen(true)}
        >
          {i18next.t("signup:Terms of Use")}
        </button>
      </label>
      <AgreementModal
        application={application}
        open={open}
        onOk={() => {
          onChange(true);
          setOpen(false);
        }}
        onCancel={() => {
          onChange(false);
          setOpen(false);
        }}
      />
    </div>
  );
}
