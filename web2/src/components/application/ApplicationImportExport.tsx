import * as React from "react";
import i18next from "i18next";
import copy from "copy-to-clipboard";
import {Download, Upload} from "lucide-react";
import {Button} from "@/components/ui/button";
import {Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle} from "@/components/ui/dialog";
import {Textarea} from "@/components/ui/textarea";
import * as Setting from "@/lib/setting";

/** Fields that represent UI/theme customization and are safe to transfer between applications. */
const UI_FIELDS = [
  "logo",
  "favicon",
  "formBackgroundUrl",
  "formBackgroundUrlMobile",
  "formCss",
  "formCssMobile",
  "formOffset",
  "formSideHtml",
  "themeData",
  "headerHtml",
  "pageHtml",
  "footerHtml",
  "signupHtml",
  "signinHtml",
];

export function exportApplicationJson(application: any) {
  const payload: Record<string, any> = {
    name: application.name,
    organization: application.organization,
  };
  for (const key of UI_FIELDS) {
    if (application[key] !== undefined) {
      payload[key] = application[key];
    }
  }
  copy(JSON.stringify(payload, null, 2));
  Setting.showMessage("success", i18next.t("general:Copied to clipboard successfully"));
}

interface ApplicationImportModalProps {
  application: any;
  onImport: (updates: Record<string, any>) => void;
}

/** "Import JSON" of the application edit page, ported from web/src/common/ApplicationImportExport.js. */
export function ApplicationImportModal({application, onImport}: ApplicationImportModalProps) {
  const [open, setOpen] = React.useState(false);
  const [jsonText, setJsonText] = React.useState("");

  const close = () => {
    setOpen(false);
    setJsonText("");
  };

  const handleOk = () => {
    let parsed: any;
    try {
      parsed = JSON.parse(jsonText);
    } catch (e: any) {
      Setting.showMessage("error", e.message);
      return;
    }

    if (parsed.name !== application.name || parsed.organization !== application.organization) {
      Setting.showMessage("error", i18next.t("general:Invalid application"));
      return;
    }

    const updates: Record<string, any> = {};
    for (const key of UI_FIELDS) {
      if (Object.prototype.hasOwnProperty.call(parsed, key)) {
        updates[key] = parsed[key];
      }
    }

    onImport(updates);
    Setting.showMessage("success", i18next.t("general:Successfully modified"));
    close();
  };

  return (
    <>
      <Button variant="outline" onClick={() => setOpen(true)}>
        <Upload />
        {i18next.t("application:Import JSON")}
      </Button>
      <Dialog open={open} onOpenChange={(next) => (next ? setOpen(true) : close())}>
        <DialogContent className="sm:max-w-[600px]">
          <DialogHeader>
            <DialogTitle>{i18next.t("application:Import JSON")}</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">{i18next.t("application:Import JSON - description")}</p>
          <Textarea rows={12} value={jsonText} onChange={(e) => setJsonText(e.target.value)} placeholder="{ ... }" />
          <DialogFooter>
            <Button variant="outline" onClick={close}>{i18next.t("general:Cancel")}</Button>
            <Button onClick={handleOk}>{i18next.t("general:OK")}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}

export function ApplicationExportButton({application}: {application: any}) {
  return (
    <Button variant="outline" onClick={() => exportApplicationJson(application)}>
      <Download />
      {i18next.t("application:Export JSON")}
    </Button>
  );
}
