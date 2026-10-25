import * as React from "react";
import i18next from "i18next";
import {Button} from "@/components/ui/button";
import {Checkbox} from "@/components/ui/checkbox";
import {Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle} from "@/components/ui/dialog";
import * as Setting from "@/lib/setting";

/** The documents an application can ask a user to accept, in the order the checkbox lists them. */
type AgreementDoc = "termsOfUse" | "privacyPolicy";
const agreementDocs: {field: AgreementDoc; titleKey: string}[] = [
  {field: "termsOfUse", titleKey: "signup:Terms of Use"},
  {field: "privacyPolicy", titleKey: "signup:Privacy Policy"},
];

export function hasAgreementDoc(application: any) {
  return agreementDocs.some(({field}) => !!application?.[field]);
}

export function isSigninAgreementShown(application: any) {
  const item = application?.signinItems?.find((signinItem: any) => signinItem.name === "Agreement");
  return hasAgreementDoc(application) && !!item?.visible;
}

function fetchAgreementDoc(url: string) {
  return fetch(url, {method: "GET"})
    .then((res) => res.text())
    .catch((error) => {
      Setting.showMessage("error", `${i18next.t("general:Failed to get")}: ${url}, ${error}`);
      return "";
    });
}

/**
 * The dialog of the terms of use or the privacy policy, ported from
 * web/src/common/modal/AgreementModal.js — the document is loaded from the
 * application's `termsOfUse` or `privacyPolicy` and shown in an iframe.
 */
export function AgreementModal({
  application,
  doc = "termsOfUse",
  open,
  onOk,
  onCancel,
}: {
  application: any;
  doc?: AgreementDoc;
  open: boolean;
  onOk: () => void;
  onCancel: () => void;
}) {
  const [text, setText] = React.useState("");
  const url = application?.[doc] as string | undefined;
  const titleKey = agreementDocs.find(({field}) => field === doc)?.titleKey ?? "signup:Terms of Use";

  React.useEffect(() => {
    if (!open || !url) {
      return;
    }
    let cancelled = false;
    setText("");
    fetchAgreementDoc(url).then((fetched) => {
      if (!cancelled) {
        setText(fetched);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [open, url]);

  return (
    <Dialog open={open} onOpenChange={(next) => (next ? undefined : onCancel())}>
      {/* 55vw is a readable column on a desktop but a squeezed one on a phone,
          where the dialog has to take the width it can get */}
      <DialogContent className="max-w-[calc(100vw-2rem)] rounded-lg p-4 sm:max-w-[55vw] sm:p-6">
        <DialogHeader>
          <DialogTitle>{i18next.t(titleKey)}</DialogTitle>
        </DialogHeader>
        <iframe title={doc} srcDoc={text} sandbox="allow-popups allow-popups-to-escape-sandbox" className="h-[55vh] w-full border-0 sm:h-[60vh]" />
        <DialogFooter>
          <Button variant="outline" className="max-sm:h-11" onClick={onCancel}>{i18next.t("signup:Decline")}</Button>
          <Button className="max-sm:h-11" onClick={onOk}>{i18next.t("signup:Accept")}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

/** The "Accept Terms of Use and Privacy Policy" checkbox, each document opening in its own dialog. */
export function AgreementCheckbox({
  application,
  checked,
  onChange,
}: {
  application: any;
  checked: boolean;
  onChange: (checked: boolean) => void;
}) {
  const [openDoc, setOpenDoc] = React.useState<AgreementDoc | null>(null);
  const docs = agreementDocs.filter(({field}) => !!application?.[field]);

  return (
    <div className="flex items-center gap-2">
      <Checkbox id="agreement" checked={checked} onCheckedChange={(v) => onChange(v === true)} />
      <label htmlFor="agreement" className="text-sm">
        {i18next.t("signup:Accept")}{" "}
        {docs.map(({field, titleKey}, i) => (
          <React.Fragment key={field}>
            {i > 0 ? i18next.t("signup:Agreement separator") : null}
            <button
              type="button"
              className="underline underline-offset-4"
              onClick={() => setOpenDoc(field)}
            >
              {i18next.t(titleKey)}
            </button>
          </React.Fragment>
        ))}
      </label>
      <AgreementModal
        application={application}
        doc={openDoc ?? "termsOfUse"}
        open={openDoc !== null}
        onOk={() => {
          onChange(true);
          setOpenDoc(null);
        }}
        onCancel={() => {
          onChange(false);
          setOpenDoc(null);
        }}
      />
    </div>
  );
}
