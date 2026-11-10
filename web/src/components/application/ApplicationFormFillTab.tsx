import i18next from "i18next";
import {Plus, Trash2} from "lucide-react";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Switch} from "@/components/ui/switch";
import {TagsInput} from "@/components/common/TagsInput";
import {EditionNotice} from "@/components/common/EditionNotice";
import {SelectField} from "@/components/common/SelectField";
import {FormRow} from "@/components/crud/FormRow";
import type {ApplicationTabProps} from "@/components/application/types";

export const FormFillExtensionUrl = "https://github.com/casdoor/casdoor-chrome-extension";
export const FormFillDesktopClientUrl = "https://github.com/casdoor/casdoor-desktop-autofill";

const AutoFillMethod = "Auto";

/** One desktop (C/S) program whose sign-in window the desktop autofill client fills. */
function FormFillProgramCard({program, onChange, onRemove}: {program: any; onChange: (program: any) => void; onRemove: () => void}) {
  const set = (key: string, value: any) => onChange({...program, [key]: value});
  const field = (key: string, labelKey: string, placeholder?: string) => (
    <FormRow labelKey={labelKey}>
      <Input value={program[key] ?? ""} placeholder={placeholder} onChange={(e) => set(key, e.target.value)} />
    </FormRow>
  );
  const auto = i18next.t("application:Found automatically");

  return (
    <div className="col-span-full rounded-lg border p-4" data-tour="form-fill-program">
      <div className="mb-1 flex items-center justify-between gap-2">
        <span className="text-sm font-medium">{program.name || program.path || i18next.t("application:Desktop program")}</span>
        <Button type="button" variant="ghost" size="icon" onClick={onRemove} aria-label={i18next.t("general:Delete")}>
          <Trash2 className="h-4 w-4" />
        </Button>
      </div>
      <div className="grid gap-x-6 xl:grid-cols-2">
        {field("name", "general:Display name")}
        {field("path", "application:Program", "u8client.exe")}
        {field("windowTitle", "application:Window title", auto)}
        {field("windowClass", "application:Window class", auto)}
        {field("publisher", "application:Publisher")}
        {field("fileHash", "application:File SHA-256")}
        {field("usernameControl", "application:Username control", auto)}
        {field("passwordControl", "application:Password control", auto)}
        {field("submitControl", "application:Submit button control", auto)}
        <FormRow labelKey="application:Fill method">
          <SelectField
            value={program.fillMethod || AutoFillMethod}
            onChange={(v) => set("fillMethod", v === AutoFillMethod ? "" : v)}
            options={[
              {id: AutoFillMethod, name: i18next.t("application:Fill the controls")},
              {id: "Keyboard", name: i18next.t("application:Type with the keyboard")},
            ]}
          />
        </FormRow>
        {program.fillMethod === "Keyboard" ? field("keySequence", "application:Key sequence", "{username}{TAB}{password}") : null}
        {field("launchPath", "application:Launch path")}
        {field("launchArgs", "application:Launch arguments")}
      </div>
    </div>
  );
}

/** The "Form autofill" tab: the browser extension signs users into this legacy system by filling its sign-in form with their managed accounts. */
export function ApplicationFormFillTab({application, updateField}: ApplicationTabProps) {
  const formFill = application.formFill ?? {};
  const set = (key: string, value: any) => updateField("formFill", {...formFill, [key]: value});
  const programs: any[] = formFill.programs ?? [];
  const setPrograms = (value: any[]) => set("programs", value);
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
      <div className="col-span-full mt-4 border-t pt-4" data-label-key="application:Desktop programs">
        <div className="text-sm font-medium">{i18next.t("application:Desktop programs")}</div>
        <div className="mt-1 text-sm text-muted-foreground">
          {i18next.t("application:Desktop programs - Description")}{" "}
          <a href={FormFillDesktopClientUrl} target="_blank" rel="noreferrer" className="underline underline-offset-2">
            {i18next.t("application:Install the desktop client")}
          </a>
        </div>
      </div>
      {programs.map((program, index) => (
        <FormFillProgramCard
          key={index}
          program={program}
          onChange={(value) => setPrograms(programs.map((item, i) => (i === index ? value : item)))}
          onRemove={() => setPrograms(programs.filter((_, i) => i !== index))}
        />
      ))}
      <div className="col-span-full">
        <Button type="button" variant="outline" size="sm" onClick={() => setPrograms([...programs, {name: "", path: ""}])}>
          <Plus className="mr-1 h-4 w-4" />
          {i18next.t("application:Add desktop program")}
        </Button>
      </div>
    </>
  );
}
