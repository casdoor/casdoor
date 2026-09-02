
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

/** The "Security" tab: the token certificate, the password policy, and the terms of use. */
export function ApplicationSecurityTab({application, updateField, mode, account, certs}: ApplicationSecurityTabProps) {
  const [uploadingTerms, setUploadingTerms] = React.useState(false);
  const termsFileRef = React.useRef<HTMLInputElement>(null);

  const uploadTermsOfUse = (file: File | undefined) => {
    if (termsFileRef.current) {
      termsFileRef.current.value = "";
    }
    if (!file) {
      return;
    }
    if (file.type !== "text/html") {
      Setting.showMessage("error", i18next.t("application:Please select a HTML file"));
      return;
    }
    setUploadingTerms(true);
    const fullFilePath = `termsOfUse/${application.owner}/${application.name}.html`;
    ResourceBackend.uploadResource(account?.owner, account?.name, "termsOfUse", "ApplicationEditPage", fullFilePath, file)
      .then((res: any) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("application:File uploaded successfully"));
          updateField("termsOfUse", res.data);
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
        }
      })
      .finally(() => setUploadingTerms(false));
  };
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
      <FormRow labelKey="signup:Terms of Use">
        <div className="flex flex-wrap items-center gap-2">
          <Input
            className="min-w-[16rem] flex-1"
            value={application.termsOfUse ?? ""}
            onChange={(e) => updateField("termsOfUse", e.target.value)}
          />
          {/* the upload writes the stored URL back onto the application, so it
              can only be done once the application exists */}
          {mode === "add" ? null : (
            <>
              <input
                ref={termsFileRef}
                type="file"
                accept=".html"
                className="hidden"
                onChange={(e) => uploadTermsOfUse(e.target.files?.[0])}
              />
              <Button
                variant="outline"
                loading={uploadingTerms}
                onClick={() => termsFileRef.current?.click()}
              >
                <Upload />
                {i18next.t("general:Click to Upload")}
              </Button>
            </>
          )}
        </div>
      </FormRow>
    </>
  );
}
