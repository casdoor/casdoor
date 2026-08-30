import * as React from "react";
import i18next from "i18next";
import {useNavigate} from "react-router-dom";
import {Input} from "@/components/ui/input";
import {Switch} from "@/components/ui/switch";
import {Textarea} from "@/components/ui/textarea";
import {Loading} from "@/components/common/Loading";
import {MultiSelect, type MultiSelectOption} from "@/components/common/MultiSelect";
import {SearchableSelect, type SearchableOption} from "@/components/common/SearchableSelect";
import {TagsInput} from "@/components/common/TagsInput";
import {CodeEditor} from "@/components/common/CodeEditor";
import {EditPageShell} from "@/components/crud/EditPageShell";
import {FormRow} from "@/components/crud/FormRow";
import {useEditRecord} from "@/hooks/use-edit-record";
import {submitEdit, type CasdoorResponse, type EditMode} from "@/lib/crud";
import * as Setting from "@/lib/setting";

type Ctx = {record: any; mode: EditMode; reload: () => void};

/** patches several fields at once, for values that have to move together */
type UpdateFields = (patch: Record<string, any>) => void;

interface BaseField {
  name: string;
  /** an i18n key, or a function of the record when the label depends on the type */
  labelKey?: string | ((ctx: Ctx) => string);
  label?: React.ReactNode | ((ctx: Ctx) => React.ReactNode);
  /** hide the row unless this returns true */
  when?: (ctx: Ctx) => boolean;
  disabled?: (ctx: Ctx) => boolean;
  block?: boolean;
  /**
   * Runs instead of the plain `updateField` when the control changes, for the
   * fields that have to reset or derive their neighbours (the syncer type
   * rewriting the table columns, the cert type clearing the SSL credentials...).
   */
  onChange?: (value: any, ctx: Ctx, updateFields: UpdateFields) => void;
}

export type EditField =
  | (BaseField & {type: "text" | "password" | "email" | "url"})
  | (BaseField & {type: "number"; step?: string})
  | (BaseField & {type: "textarea"; rows?: number; placeholder?: string})
  | (BaseField & {type: "switch"})
  | (BaseField & {type: "tags"})
  | (BaseField & {type: "select"; options: (ctx: Ctx) => SearchableOption[]})
  | (BaseField & {type: "multiselect"; options: (ctx: Ctx) => MultiSelectOption[]; creatable?: boolean})
  | (BaseField & {type: "code"; language?: string; height?: number})
  | (BaseField & {type: "custom"; render: (ctx: Ctx, update: (field: string, value: any) => void) => React.ReactNode});

export interface SimpleEditPageProps {
  titleKey: string;
  backTo: string;
  fields: EditField[];
  fetch: () => Promise<CasdoorResponse<any>>;
  add: (record: any) => Promise<CasdoorResponse>;
  update: (record: any) => Promise<CasdoorResponse>;
  /** deps of the fetch, usually the route params */
  deps?: React.DependencyList;
  /** where to go after "Save" (not "Save & Exit") when the name changed */
  editUrl?: (record: any) => string;
  transform?: (record: any) => any;
  /**
   * Last chance to adjust the payload before it is sent. Returning `null` aborts
   * the save, which is how a page rejects a record its own validation refuses.
   */
  beforeSave?: (record: any) => any;
  extraActions?: (ctx: Ctx) => React.ReactNode;
  children?: (ctx: Ctx, update: (field: string, value: any) => void) => React.ReactNode;
}

/**
 * Config-driven edit page. The Casdoor edit pages are all "load object, render a
 * list of labelled fields, POST it back", so they are expressed as data here and
 * only the genuinely custom ones (application, user, provider) are hand-written.
 */
export function SimpleEditPage({
  titleKey,
  backTo,
  fields,
  fetch,
  add,
  update,
  deps = [],
  editUrl,
  transform,
  beforeSave,
  extraActions,
  children,
}: SimpleEditPageProps) {
  const navigate = useNavigate();
  const [saving, setSaving] = React.useState(false);
  const {record, updateField, updateFields, loading, mode, setMode, reload} = useEditRecord<any>({fetch, transform, deps});

  if (loading || record === null) {
    return <Loading />;
  }

  const ctx: Ctx = {record, mode, reload};

  const save = async(exitAfterSave: boolean) => {
    const payload = beforeSave ? beforeSave(Setting.deepCopy(record)) : Setting.deepCopy(record);
    if (payload === null) {
      return;
    }
    setSaving(true);
    await submitEdit({
      mode,
      record: payload,
      add,
      update,
      onSaved: () => {
        setMode("edit");
        if (exitAfterSave) {
          navigate(backTo);
        } else if (editUrl) {
          const next = editUrl(record);
          if (next !== window.location.pathname) {
            navigate(next, {replace: true});
          }
        }
      },
    });
    setSaving(false);
  };

  const renderField = (field: EditField) => {
    if (field.when && !field.when(ctx)) {
      return null;
    }
    const value = record[field.name];
    const disabled = field.disabled ? field.disabled(ctx) : false;
    const set = (next: any) =>
      field.onChange ? field.onChange(next, ctx, updateFields) : updateField(field.name, next);

    let control: React.ReactNode;
    switch (field.type) {
    case "textarea":
      control = (
        <Textarea
          rows={field.rows ?? 4}
          disabled={disabled}
          placeholder={field.placeholder}
          value={value ?? ""}
          onChange={(e) => set(e.target.value)}
        />
      );
      break;
    case "number":
      control = (
        <Input
          type="number"
          step={field.step}
          disabled={disabled}
          value={value ?? 0}
          onChange={(e) => set(field.step ? Number(e.target.value) : Setting.myParseInt(e.target.value))}
        />
      );
      break;
    case "switch":
      control = (
        <Switch disabled={disabled} checked={!!value} onCheckedChange={(v) => set(v)} />
      );
      break;
    case "tags":
      control = (
        <TagsInput disabled={disabled} value={value ?? []} onChange={(v) => set(v)} />
      );
      break;
    case "select":
      control = (
        <SearchableSelect
          disabled={disabled}
          value={value ?? ""}
          onChange={(v) => set(v)}
          options={field.options(ctx)}
        />
      );
      break;
    case "multiselect":
      control = (
        <MultiSelect
          disabled={disabled}
          creatable={field.creatable}
          value={value ?? []}
          onChange={(v) => set(v)}
          options={field.options(ctx)}
        />
      );
      break;
    case "code":
      control = (
        <CodeEditor
          language={field.language}
          height={field.height}
          value={value ?? ""}
          onChange={(v) => set(v)}
        />
      );
      break;
    case "custom":
      control = field.render(ctx, updateField);
      break;
    default:
      control = (
        <Input
          type={field.type === "text" ? "text" : field.type}
          disabled={disabled}
          value={value ?? ""}
          onChange={(e) => set(e.target.value)}
        />
      );
    }

    return (
      <FormRow
        key={field.name}
        labelKey={typeof field.labelKey === "function" ? field.labelKey(ctx) : field.labelKey}
        label={typeof field.label === "function" ? field.label(ctx) : field.label}
        block={field.block || field.type === "code"}
      >
        {control}
      </FormRow>
    );
  };

  return (
    <EditPageShell
      title={`${i18next.t(titleKey)} - ${record.displayName || record.name}`}
      mode={mode}
      backTo={backTo}
      onSave={save}
      saving={saving}
      extraActions={extraActions?.(ctx)}
    >
      {fields.map(renderField)}
      {children?.(ctx, updateField)}
    </EditPageShell>
  );
}
