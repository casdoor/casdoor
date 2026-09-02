
import i18next from "i18next";
import {Copy, Link as LinkIcon} from "lucide-react";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Switch} from "@/components/ui/switch";
import {CodeEditor} from "@/components/common/CodeEditor";
import {SelectField} from "@/components/common/SelectField";
import {EditableTable} from "@/components/crud/EditableTable";
import {FormRow} from "@/components/crud/FormRow";
import type {ApplicationTabProps} from "@/components/application/types";
import * as Setting from "@/lib/setting";

const SAML_HASH_ALGORITHMS = ["SHA1", "SHA256", "SHA512"];

interface ApplicationSamlTabProps extends ApplicationTabProps {
  mode: string;
  samlMetadata: string;
  samlMetadataUrl: string;
  idpInitiatedSsoUrl: string;
}

/** The "SAML" tab: the IdP side of a SAML application, and the metadata a service provider needs. */
export function ApplicationSamlTab({
  application,
  updateField,
  mode,
  samlMetadata,
  samlMetadataUrl,
  idpInitiatedSsoUrl,
}: ApplicationSamlTabProps) {
  return (
    <>
      <FormRow labelKey="application:SAML reply URL">
        <Input value={application.samlReplyUrl ?? ""} onChange={(e) => updateField("samlReplyUrl", e.target.value)} />
      </FormRow>
      <FormRow labelKey="application:Enable SAML compression">
        <Switch
          checked={!!application.enableSamlCompress}
          onCheckedChange={(v) => updateField("enableSamlCompress", v)}
        />
      </FormRow>
      <FormRow labelKey="application:Enable SAML C14N10">
        <Switch
          checked={!!application.enableSamlC14n10}
          onCheckedChange={(v) => updateField("enableSamlC14n10", v)}
        />
      </FormRow>
      {application.enableSamlC14n10 ? (
        <FormRow labelKey="application:SAML C14N10 prefix">
          <Input
            value={application.samlC14nPrefix ?? ""}
            placeholder="xs"
            onChange={(e) => updateField("samlC14nPrefix", e.target.value)}
          />
        </FormRow>
      ) : null}
      <FormRow labelKey="application:Use Email as NameID">
        <Switch
          checked={!!application.useEmailAsSamlNameId}
          onCheckedChange={(v) => updateField("useEmailAsSamlNameId", v)}
        />
      </FormRow>
      <FormRow labelKey="application:Enable SAML POST binding">
        <Switch
          checked={!!application.enableSamlPostBinding}
          onCheckedChange={(v) => updateField("enableSamlPostBinding", v)}
        />
      </FormRow>
      <FormRow block labelKey="application:SAML hash algorithm">
        <SelectField
          value={application.samlHashAlgorithm ?? "SHA256"}
          onChange={(v) => updateField("samlHashAlgorithm", v)}
          options={SAML_HASH_ALGORITHMS.map((item) => ({id: item, name: item}))}
        />
      </FormRow>
      <FormRow block labelKey="application:Disable SAML attributes">
        <Switch
          checked={!!application.disableSamlAttributes}
          onCheckedChange={(v) => updateField("disableSamlAttributes", v)}
        />
      </FormRow>
      <FormRow block labelKey="application:Enable SAML assertion signature">
        <Switch
          checked={!!application.enableSamlAssertionSignature}
          onCheckedChange={(v) => updateField("enableSamlAssertionSignature", v)}
        />
      </FormRow>
      <FormRow labelKey="general:SAML attributes" block>
        <EditableTable
          rows={application.samlAttributes ?? []}
          onChange={(rows) => updateField("samlAttributes", rows)}
          newRow={() => ({name: "", nameFormat: "", value: ""})}
          reorderable={false}
          columns={[
            {
              key: "name",
              title: i18next.t("general:Name"),
              width: 220,
              render: (row: any, _i, patch) => (
                <Input value={row.name ?? ""} onChange={(e) => patch({name: e.target.value})} />
              ),
            },
            {
              key: "nameFormat",
              title: i18next.t("general:Name format"),
              width: 220,
              render: (row: any, _i, patch) => (
                <Input value={row.nameFormat ?? ""} onChange={(e) => patch({nameFormat: e.target.value})} />
              ),
            },
            {
              key: "value",
              title: i18next.t("webhook:Value"),
              render: (row: any, _i, patch) => (
                <Input value={row.value ?? ""} onChange={(e) => patch({value: e.target.value})} />
              ),
            },
          ]}
        />
      </FormRow>
      {/* both are generated from the saved application, so only after it exists */}
      {mode === "add" ? null : (
        <>
          <FormRow labelKey="application:SAML metadata" block>
            <div className="space-y-2">
              <CodeEditor language="xml" value={samlMetadata} readOnly onChange={() => undefined} />
              <Button variant="outline" size="sm" onClick={() => Setting.copyToClipboard(samlMetadataUrl)}>
                <Copy />
                {i18next.t("application:Copy SAML metadata URL")}
              </Button>
            </div>
          </FormRow>
          <FormRow labelKey="application:IdP-initiated SSO URL" block>
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <LinkIcon className="h-4 w-4 shrink-0 text-muted-foreground" />
                <Input value={idpInitiatedSsoUrl} readOnly />
              </div>
              <Button variant="outline" size="sm" onClick={() => Setting.copyToClipboard(idpInitiatedSsoUrl)}>
                <Copy />
                {i18next.t("application:Copy IdP-initiated SSO URL")}
              </Button>
            </div>
          </FormRow>
        </>
      )}
    </>
  );
}
