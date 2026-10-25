
import * as React from "react";
import i18next from "i18next";
import {Upload} from "lucide-react";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {SearchableSelect, type SearchableOption} from "@/components/common/SearchableSelect";
import {FormRow} from "@/components/crud/FormRow";
import type {ApplicationTabProps} from "@/components/application/types";
import * as ResourceBackend from "@/backend/ResourceBackend";
import * as Setting from "@/lib/setting";

interface ApplicationSecurityTabProps extends ApplicationTabProps {
  mode: string;
  account: any;
  certs: SearchableOption[];
}

/** An HTML document the signin and signup pages ask users to accept, typed in as a URL or uploaded. */
function AgreementDocRow({application, updateField, mode, account, field, labelKey}: ApplicationTabProps & {
  mode: string;
  account: any;
  field: "termsOfUse" | "privacyPolicy";
  labelKey: string;
}) {
  const [uploading, setUploading] = React.useState(false);
  const fileRef = React.useRef<HTMLInputElement>(null);

  const upload = (file: File | undefined) => {
    if (fileRef.current) {
      fileRef.current.value = "";
    }
    if (!file) {
      return;
    }
    if (file.type !== "text/html") {
      Setting.showMessage("error", i18next.t("application:Please select a HTML file"));
      return;
    }
    setUploading(true);
    const fullFilePath = `${field}/${application.owner}/${application.name}.html`;
    ResourceBackend.uploadResource(account?.owner, account?.name, field, "ApplicationEditPage", fullFilePath, file)
      .then((res: any) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("application:File uploaded successfully"));
          updateField(field, res.data);
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
        }
      })
      .finally(() => setUploading(false));
  };

  return (
    <FormRow labelKey={labelKey}>
      <div className="flex flex-wrap items-center gap-2">
        <Input
          className="min-w-[16rem] flex-1"
          value={application[field] ?? ""}
          onChange={(e) => updateField(field, e.target.value)}
        />
        {/* the upload writes the stored URL back onto the application, so it
            can only be done once the application exists */}
        {mode === "add" ? null : (
          <>
            <input
              ref={fileRef}
              type="file"
              accept=".html"
              className="hidden"
              onChange={(e) => upload(e.target.files?.[0])}
            />
            <Button
              variant="outline"
              loading={uploading}
              onClick={() => fileRef.current?.click()}
            >
              <Upload />
              {i18next.t("general:Click to Upload")}
            </Button>
          </>
        )}
      </div>
    </FormRow>
  );
}

/** The "Security" tab: the token certificate, the password policy, the terms of use and the privacy policy. */
export function ApplicationSecurityTab({application, updateField, mode, account, certs}: ApplicationSecurityTabProps) {
  return (
    <>
      <FormRow labelKey="application:Token cert">
        <SearchableSelect
          value={application.cert ?? ""}
          onChange={(v) => updateField("cert", v)}
          options={certs}
        />
      </FormRow>
      <FormRow labelKey="application:Client cert">
        <SearchableSelect
          value={application.clientCert ?? ""}
          onChange={(v) => updateField("clientCert", v)}
          options={certs}
        />
      </FormRow>
      <FormRow labelKey="application:Failed signin limit">
        <Input
          type="number"
          value={application.failedSigninLimit ?? 5}
          onChange={(e) => updateField("failedSigninLimit", Setting.myParseInt(e.target.value))}
        />
      </FormRow>
      <FormRow labelKey="application:Failed signin frozen time">
        <Input
          type="number"
          value={application.failedSigninFrozenTime ?? 15}
          onChange={(e) => updateField("failedSigninFrozenTime", Setting.myParseInt(e.target.value))}
        />
      </FormRow>
      <FormRow labelKey="application:Code resend timeout">
        <Input
          type="number"
          value={application.codeResendTimeout ?? 60}
          onChange={(e) => updateField("codeResendTimeout", Setting.myParseInt(e.target.value))}
        />
      </FormRow>
      <FormRow labelKey="general:IP whitelist">
        <Input
          placeholder={application.organizationObj?.ipWhitelist}
          value={application.ipWhitelist ?? ""}
          onChange={(e) => updateField("ipWhitelist", e.target.value)}
        />
      </FormRow>
      <AgreementDocRow application={application} updateField={updateField} mode={mode} account={account} field="termsOfUse" labelKey="signup:Terms of Use" />
      <AgreementDocRow application={application} updateField={updateField} mode={mode} account={account} field="privacyPolicy" labelKey="signup:Privacy Policy" />
    </>
  );
}
