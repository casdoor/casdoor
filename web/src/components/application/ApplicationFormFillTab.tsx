import i18next from "i18next";
import {Input} from "@/components/ui/input";
import {Switch} from "@/components/ui/switch";
import {TagsInput} from "@/components/common/TagsInput";
import {EditionNotice} from "@/components/common/EditionNotice";
import {FormRow} from "@/components/crud/FormRow";
import type {ApplicationTabProps} from "@/components/application/types";

export const FormFillExtensionUrl = "https://github.com/casdoor/casdoor-chrome-extension";

/** The "Form autofill" tab: the browser extension signs users into this legacy system by filling its sign-in form with their managed accounts. */
export function ApplicationFormFillTab({application, updateField}: ApplicationTabProps) {
  const formFill = application.formFill ?? {};
  const set = (key: string, value: any) => updateField("formFill", {...formFill, [key]: value});
  const selector = (key: string, labelKey: string) => (
    <FormRow labelKey={labelKey}>
      <Input
        value={formFill[key] ?? ""}
        placeholder={i18next.t("application:Found automatically")}
        onChange={(e) => set(key, e.target.value)}
      />
    </FormRow>
  );

  return (
    <>
      <div className="col-span-full mb-2 text-sm text-muted-foreground">
        {i18next.t("application:Form autofill - Description")}{" "}
        <a href={FormFillExtensionUrl} target="_blank" rel="noreferrer" className="underline underline-offset-2">
          {i18next.t("application:Install the browser extension")}
        </a>
      </div>
      <FormRow labelKey="application:Enable form autofill">
        <Switch checked={!!formFill.enabled} onCheckedChange={(v) => set("enabled", v)} />
        {formFill.enabled ? <EditionNotice edition="Enterprise" /> : null}
      </FormRow>
      <FormRow labelKey="application:Sign-in page URLs">
        <TagsInput value={formFill.urls ?? []} onChange={(v) => set("urls", v)} />
      </FormRow>
      {selector("usernameSelector", "application:Username field selector")}
      {selector("passwordSelector", "application:Password field selector")}
      {selector("submitSelector", "application:Submit button selector")}
      <FormRow labelKey="application:Submit automatically">
        <Switch checked={!!formFill.autoSubmit} onCheckedChange={(v) => set("autoSubmit", v)} />
      </FormRow>
      <FormRow labelKey="application:Let users save their accounts">
        <Switch checked={!!formFill.allowSave} onCheckedChange={(v) => set("allowSave", v)} />
      </FormRow>
    </>
  );
}
