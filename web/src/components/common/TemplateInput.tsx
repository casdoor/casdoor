import * as React from "react";
import i18next from "i18next";
import {Input} from "@/components/ui/input";
import {Textarea} from "@/components/ui/textarea";
import {cn} from "@/lib/utils";

export interface TemplateVariable {
  /** written as `${name}` in the text */
  name: string;
  description: string;
}

interface TemplateInputProps {
  value: string;
  onChange: (value: string) => void;
  variables: TemplateVariable[];
  multiline?: boolean;
  rows?: number;
  placeholder?: string;
  disabled?: boolean;
  /** the text with the variables replaced by sample values, shown under the field */
  preview?: string;
}

type Field = HTMLInputElement | HTMLTextAreaElement;

/**
 * A text field for message templates: typing `${` lists the variables to pick
 * from, and the variables below the field insert themselves at the cursor.
 */
export function TemplateInput({value, onChange, variables, multiline, rows, placeholder, disabled, preview}: TemplateInputProps) {
  const ref = React.useRef<Field | null>(null);
  // the `${...` being typed: where it starts and what follows the `${`
  const [typing, setTyping] = React.useState<{start: number; query: string} | null>(null);
  const [active, setActive] = React.useState(0);
  const text = value ?? "";

  const matches = typing === null ? [] : variables.filter((v) => v.name.toLowerCase().includes(typing.query.toLowerCase()));

  const detect = (field: Field) => {
    const caret = field.selectionStart ?? field.value.length;
    const m = /\$\{([\w.]*)$/.exec(field.value.slice(0, caret));
    setTyping(m ? {start: caret - m[0].length, query: m[1]} : null);
    setActive(0);
  };

  // where the cursor goes once the inserted variable is rendered, set before the next keystroke can land
  const caret = React.useRef<number | null>(null);
  React.useLayoutEffect(() => {
    const field = ref.current;
    if (caret.current !== null && field) {
      field.focus();
      field.setSelectionRange(caret.current, caret.current);
      caret.current = null;
    }
  }, [text]);

  const replace = (start: number, end: number, insert: string) => {
    caret.current = start + insert.length;
    onChange(text.slice(0, start) + insert + text.slice(end));
    setTyping(null);
  };

  const pick = (variable: TemplateVariable) => {
    if (typing === null) {
      return;
    }
    let end = typing.start + 2 + typing.query.length;
    // "${us}" with the cursor before "}" must not end up as "${user}}"
    if (text[end] === "}") {
      end += 1;
    }
    replace(typing.start, end, `\${${variable.name}}`);
  };

  const insert = (variable: TemplateVariable) => {
    const field = ref.current;
    const start = field?.selectionStart ?? text.length;
    const end = field?.selectionEnd ?? text.length;
    replace(start, end, `\${${variable.name}}`);
  };

  const onKeyDown = (e: React.KeyboardEvent<Field>) => {
    if (matches.length === 0) {
      return;
    }
    if (e.key === "ArrowDown") {
      e.preventDefault();
      setActive((active + 1) % matches.length);
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setActive((active - 1 + matches.length) % matches.length);
    } else if (e.key === "Enter" || e.key === "Tab") {
      e.preventDefault();
      pick(matches[active]);
    } else if (e.key === "Escape") {
      e.preventDefault();
      setTyping(null);
    }
  };

  const fieldProps = {
    disabled,
    placeholder,
    value: text,
    onChange: (e: React.ChangeEvent<Field>) => {
      onChange(e.target.value);
      detect(e.target);
    },
    onKeyDown,
    onClick: (e: React.MouseEvent<Field>) => detect(e.currentTarget),
    onBlur: () => setTyping(null),
  };

  return (
    <div className="space-y-2">
      <div className="relative">
        {multiline ? (
          <Textarea ref={(el) => {ref.current = el;}} rows={rows ?? 4} {...fieldProps} />
        ) : (
          <Input ref={(el) => {ref.current = el;}} {...fieldProps} />
        )}
        {matches.length > 0 ? (
          <div role="listbox" className="absolute left-0 top-full z-50 mt-1 w-80 max-w-full overflow-hidden rounded-md border bg-popover p-1 text-popover-foreground shadow-md">
            {matches.map((v, i) => (
              <div
                key={v.name}
                role="option"
                aria-selected={i === active}
                className={cn("flex cursor-pointer items-baseline justify-between gap-3 rounded-sm px-2 py-1.5 text-sm", i === active && "bg-accent text-accent-foreground")}
                // mousedown, not click: the field's blur would close the list first
                onMouseDown={(e) => {
                  e.preventDefault();
                  pick(v);
                }}
                onMouseEnter={() => setActive(i)}
              >
                <code className="font-mono text-xs">{`\${${v.name}}`}</code>
                <span className="truncate text-xs text-muted-foreground">{v.description}</span>
              </div>
            ))}
          </div>
        ) : null}
      </div>
      <div className="flex flex-wrap items-center gap-1.5 text-xs text-muted-foreground">
        <span>{i18next.t("webhook:Insert variable")}:</span>
        {variables.map((v) => (
          <button
            key={v.name}
            type="button"
            disabled={disabled}
            title={v.description}
            className="rounded border bg-muted/50 px-1.5 py-0.5 font-mono text-foreground hover:bg-accent disabled:opacity-50"
            onMouseDown={(e) => e.preventDefault()}
            onClick={() => insert(v)}
          >
            {`\${${v.name}}`}
          </button>
        ))}
      </div>
      {preview !== undefined && text !== "" ? (
        <div className="whitespace-pre-wrap rounded-md border border-dashed px-3 py-2 text-xs text-muted-foreground">
          <span className="font-medium">{i18next.t("general:Preview")}: </span>
          {preview}
        </div>
      ) : null}
    </div>
  );
}
