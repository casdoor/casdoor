import * as React from "react";
import i18next from "i18next";
import {useNavigate} from "react-router-dom";
import {Input} from "@/components/ui/input";
import {Switch} from "@/components/ui/switch";
import {Textarea} from "@/components/ui/textarea";
import {Loading} from "@/components/common/Loading";
import {UnauthorizedPage} from "@/components/common/UnauthorizedPage";
import {MultiSelect, type MultiSelectOption} from "@/components/common/MultiSelect";
import {SearchableSelect, type SearchableOption} from "@/components/common/SearchableSelect";
import {TagsInput} from "@/components/common/TagsInput";
import {CodeEditor} from "@/components/common/CodeEditor";
import {EditPageShell} from "@/components/crud/EditPageShell";
import {FormGrid, FormRow} from "@/components/crud/FormRow";
import {useEditRecord} from "@/hooks/use-edit-record";
import {getModeTitleKey, submitEdit, type CasdoorResponse, type EditMode} from "@/lib/crud";
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
  /** the backend rejects an empty value, so say so before the round trip */
  required?: boolean;
  /**
   * Returns a message when the value is not acceptable, or undefined when it is.
   * Only for what the frontend can decide on its own — the authority is still the
   * Go backend, and its rejection still arrives as a message on save.
   */
  validate?: (value: any, ctx: Ctx) => string | undefined;
  /**
   * Runs instead of the plain `updateField` when the control changes, for the
   * fields that have to reset or derive their neighbours (the syncer type
   * rewriting the table columns, the cert type clearing the SSL credentials...).
   */
  onChange?: (value: any, ctx: Ctx, updateFields: UpdateFields) => void;
}

export type EditField =
  | (BaseField & {type: "text" | "password" | "email" | "url"})
  | (BaseField & {type: "number"; step?: string; min?: number; max?: number; suffix?: React.ReactNode})
  | (BaseField & {type: "textarea"; rows?: number; placeholder?: string})
  | (BaseField & {type: "switch"})
  | (BaseField & {type: "tags"; placeholder?: string})
  | (BaseField & {type: "select"; options: (ctx: Ctx) => SearchableOption[]})
  | (BaseField & {
    type: "multiselect";
    options: (ctx: Ctx) => MultiSelectOption[];
    /** a predicate when only some records may invent their own values */
    creatable?: boolean | ((ctx: Ctx) => boolean);
  })
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
  /**
   * Fields the server decided while adding the record, as a patch to apply to it.
   * Only for what a reload cannot reach: a record the server renamed is no longer
   * where the page would look for it.
   */
  onAdded?: (record: any, res: CasdoorResponse) => Record<string, any> | undefined;
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
  onAdded,
  transform,
  beforeSave,
  extraActions,
  children,
}: SimpleEditPageProps) {
  const navigate = useNavigate();
  const [saving, setSaving] = React.useState(false);
  const [errors, setErrors] = React.useState<Record<string, string>>({});
  const {record, updateField, updateFields, loading, denied, mode, setMode, reload} = useEditRecord<any>({fetch, transform, deps});
  const savedIdentity = React.useRef<{owner: any; name: any} | null>(null);

  React.useEffect(() => {
    savedIdentity.current = null;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, deps);

  if (denied) {
    return <UnauthorizedPage />;
  }

  if (loading || record === null) {
    return <Loading />;
  }

  const ctx: Ctx = {record, mode, reload};

  // the identity the record was last loaded or saved under: a rejected rename has
  // to be rolled back, or the next save would address an object that never existed
  if (savedIdentity.current === null && mode !== "add") {
    savedIdentity.current = {owner: record.owner, name: record.name};
  }

  const isEmpty = (value: any) =>
    value === undefined || value === null || value === "" || (Array.isArray(value) && value.length === 0);

  /** a hidden or locked row is not the reader's to fix, so it is not checked */
  const applies = (field: EditField) =>
    (!field.when || field.when(ctx)) && mode !== "view" && !(field.disabled ? field.disabled(ctx) : false);

  const checkField = (field: EditField, value: any): string | undefined => {
    if (!applies(field)) {
      return undefined;
    }
    if (field.required && isEmpty(value)) {
      return i18next.t("general:This field is required");
    }
    return field.validate?.(value, ctx);
  };

  const save = async(exitAfterSave: boolean) => {
    const found: Record<string, string> = {};
    fields.forEach((field) => {
      const message = checkField(field, record[field.name]);
      if (message) {
        found[field.name] = message;
      }
    });
    if (Object.keys(found).length > 0) {
      setErrors(found);
      // every offending row is marked; the page scrolls to the first one
      document.querySelector(`[data-field="${Object.keys(found)[0]}"]`)?.scrollIntoView({block: "center"});
      return;
    }
    setErrors({});

    const payload = beforeSave ? beforeSave(Setting.deepCopy(record)) : Setting.deepCopy(record);
    if (payload === null) {
      return;
    }
    const isAdd = mode === "add";
    setSaving(true);
    await submitEdit({
      mode,
      record: payload,
      add,
      update,
      onSaved: (saved, res) => {
        setMode("edit");
        const patch = isAdd ? onAdded?.(saved, res) : undefined;
        if (patch) {
          updateFields(patch);
        }
        if (exitAfterSave) {
          navigate(backTo);
          return;
        }
        const next = editUrl ? editUrl({...record, ...patch}) : null;
        if (next && next !== window.location.pathname) {
          navigate(next, {replace: true});
        } else if (isAdd) {
          // the server fills in what the form could not know: generated ids and
          // secrets, computed prices, defaults taken from the organization
          reload();
        }
        savedIdentity.current = {owner: record.owner, name: record.name};
      },
      onFailed: () => {
        // the antd pages put the name back when the backend rejects the save
        if (!isAdd && savedIdentity.current) {
          updateFields(savedIdentity.current);
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
    // a read-only page locks the whole form, whatever each field asked for
    const disabled = mode === "view" || (field.disabled ? field.disabled(ctx) : false);
    const set = (next: any) => {
      if (errors[field.name]) {
        const message = checkField(field, next);
        setErrors((previous) => {
          const rest = {...previous};
          delete rest[field.name];
          return message ? {...rest, [field.name]: message} : rest;
        });
      }
      if (field.onChange) {
        field.onChange(next, ctx, updateFields);
      } else {
        updateField(field.name, next);
      }
    };

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
        <div className="flex items-center gap-2">
          <Input
            type="number"
            step={field.step}
            min={field.min}
            max={field.max}
            disabled={disabled}
            value={value ?? 0}
            onChange={(e) => set(field.step ? Number(e.target.value) : Setting.myParseInt(e.target.value))}
          />
          {/* antd's `addonAfter`, for the fields whose unit matters */}
          {field.suffix ? <span className="shrink-0 text-sm text-muted-foreground">{field.suffix}</span> : null}
        </div>
      );
      break;
    case "switch":
      control = (
        <Switch disabled={disabled} checked={!!value} onCheckedChange={(v) => set(v)} />
      );
      break;
    case "tags":
      control = (
        <TagsInput
          disabled={disabled}
          value={value ?? []}
          onChange={(v) => set(v)}
          placeholder={field.placeholder}
        />
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
          creatable={typeof field.creatable === "function" ? field.creatable(ctx) : field.creatable}
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
        className="scroll-mt-24"
        labelKey={typeof field.labelKey === "function" ? field.labelKey(ctx) : field.labelKey}
        label={typeof field.label === "function" ? field.label(ctx) : field.label}
        block={field.block || field.type === "code"}
        required={field.required && applies(field)}
        error={errors[field.name]}
      >
        <div data-field={field.name}>{control}</div>
      </FormRow>
    );
  };

  return (
    <EditPageShell
      title={`${i18next.t(getModeTitleKey(titleKey, mode))} - ${record.displayName || record.name}`}
      mode={mode}
      backTo={backTo}
      onSave={save}
      saving={saving}
      extraActions={extraActions?.(ctx)}
    >
      <FormGrid>{fields.map(renderField)}</FormGrid>
      {children?.(ctx, updateField)}
    </EditPageShell>
  );
}
