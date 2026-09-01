import * as React from "react";
import i18next from "i18next";
import {HelpCircle} from "lucide-react";
import {Tooltip, TooltipContent, TooltipTrigger} from "@/components/ui/tooltip";
import {cn} from "@/lib/utils";

interface FormRowProps {
  label?: React.ReactNode;
  /**
   * i18n key such as "general:Name". The label is `t(key)` and the tooltip is
   * `t(key + " - Tooltip")` when that key exists, matching the antd pages.
   */
  labelKey?: string;
  tooltip?: React.ReactNode;
  children: React.ReactNode;
  className?: string;
  /**
   * This row's content wants the full width of the card — a table, a code editor,
   * a transfer list. Inside a `FormGrid` it takes both columns instead of one.
   */
  block?: boolean;
  htmlFor?: string;
  /** marks the label, so the asterisk is not the control's problem to draw */
  required?: boolean;
  /** shown under the control; the row is only in error while this is set */
  error?: React.ReactNode;
}

/**
 * Lays a run of `FormRow`s out in two columns once the card is wide enough for
 * each one to keep a readable measure (~500px at the `xl` breakpoint, where the
 * content area clears 1000px). Below that it stays a single column.
 *
 * The Casdoor edit pages carry 20-60 fields each, so a single column turned them
 * into a fifteen-screen scroll while two thirds of the card sat empty.
 */
export const formGridClass = "grid grid-cols-1 gap-x-10 xl:grid-cols-2";

export function FormGrid({children, className}: {children: React.ReactNode; className?: string}) {
  return <div className={cn(formGridClass, className)}>{children}</div>;
}

/**
 * The label/control row used by every Casdoor edit page.
 *
 * The antd version was a `<Row><Col span={2}>label</Col><Col span={22}>control
 * </Col></Row>`. Here the label sits above its control and the row is a cell of
 * the enclosing `FormGrid`, which is what decides how many fit side by side.
 */
export function FormRow({label, labelKey, tooltip, children, className, block, htmlFor, required, error}: FormRowProps) {
  let resolvedLabel = label;
  let resolvedTooltip = tooltip;
  if (labelKey) {
    resolvedLabel = resolvedLabel ?? i18next.t(labelKey);
    if (resolvedTooltip === undefined) {
      const tooltipKey = `${labelKey} - Tooltip`;
      const translated = i18next.t(tooltipKey);
      // i18next returns the key itself when the translation is missing
      resolvedTooltip = translated === tooltipKey || translated === `${labelKey.split(":")[1]} - Tooltip` ? undefined : translated;
    }
  }

  return (
    <div className={cn("min-w-0 space-y-1.5 py-2.5", block && "xl:col-span-2", className)}>
      <label htmlFor={htmlFor} className="flex items-center gap-1.5 text-sm font-medium leading-none">
        <span>
          {resolvedLabel}
          {required ? <span className="ml-0.5 text-destructive">*</span> : null}
        </span>
        {resolvedTooltip ? (
          <Tooltip>
            <TooltipTrigger asChild>
              <button type="button" tabIndex={-1} className="text-muted-foreground/60 hover:text-foreground">
                <HelpCircle className="h-3.5 w-3.5" />
              </button>
            </TooltipTrigger>
            <TooltipContent className="max-w-sm">{resolvedTooltip}</TooltipContent>
          </Tooltip>
        ) : null}
      </label>
      {/* the ring is painted on a wrapper rather than the control, so every kind of
          control a page puts in a row — input, select, switch, code editor — shows
          the same error state without having to thread a prop through */}
      <div className={cn("min-w-0", error && "rounded-md ring-1 ring-destructive ring-offset-2 ring-offset-background")}>
        {children}
      </div>
      {error ? <p className="text-xs text-destructive">{error}</p> : null}
    </div>
  );
}

export function FormSection({
  title,
  description,
  children,
  className,
}: {
  title?: React.ReactNode;
  description?: React.ReactNode;
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <section className={cn("space-y-2 xl:col-span-2", className)}>
      {title ? (
        <header className="space-y-0.5">
          <h3 className="text-sm font-semibold tracking-tight">{title}</h3>
          {description ? <p className="text-sm text-muted-foreground">{description}</p> : null}
        </header>
      ) : null}
      {/* the rows carry their own rhythm now; a rule between each one turned the
          form into a spreadsheet */}
      <FormGrid>{children}</FormGrid>
    </section>
  );
}
