
import i18next from "i18next";
import {Input} from "@/components/ui/input";
import {SearchableSelect} from "@/components/common/SearchableSelect";
import {SelectField} from "@/components/common/SelectField";
import {TagsInput} from "@/components/common/TagsInput";
import {FormRow} from "@/components/crud/FormRow";
import type {SearchableOption} from "@/components/common/SearchableSelect";
import type {ApplicationTabProps} from "@/components/application/types";

const SSL_MODES = ["", "HTTP", "HTTPS and HTTP", "HTTPS Only"];

/** The "Reverse Proxy" tab: the domain and certificate Casdoor fronts for this application. */
interface ApplicationReverseProxyTabProps extends ApplicationTabProps {
  certs: SearchableOption[];
}

export function ApplicationReverseProxyTab({application, updateField, certs}: ApplicationReverseProxyTabProps) {
  return (
    <>
      <FormRow labelKey="provider:Domain">
        <Input
          value={application.domain ?? ""}
          placeholder="e.g., blog.example.com"
          onChange={(e) => updateField("domain", e.target.value)}
        />
      </FormRow>
      <FormRow labelKey="application:Other domains">
        <TagsInput value={application.otherDomains ?? []} onChange={(v) => updateField("otherDomains", v)} />
      </FormRow>
      <FormRow labelKey="application:Upstream host">
        <Input
          value={application.upstreamHost ?? ""}
          placeholder="e.g., localhost:8080 or 192.168.1.100:3000"
          onChange={(e) => updateField("upstreamHost", e.target.value)}
        />
      </FormRow>
      <FormRow labelKey="provider:SSL mode">
        <SelectField
          value={application.sslMode ?? ""}
          onChange={(v) => updateField("sslMode", v)}
          options={SSL_MODES.map((item) => ({id: item, name: item === "" ? i18next.t("general:None") : item}))}
        />
      </FormRow>
      <FormRow labelKey="application:SSL cert">
        <SearchableSelect
          value={application.sslCert ?? ""}
          onChange={(v) => updateField("sslCert", v)}
          options={[{value: "", label: i18next.t("general:None")}, ...certs]}
        />
      </FormRow>
    </>
  );
}
