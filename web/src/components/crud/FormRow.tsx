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
   * a transfer list. Everything else is capped at a readable measure.
   */
  block?: boolean;
  htmlFor?: string;
}

/**
 * The label/control row used by every Casdoor edit page.
 *
 * The antd version was a `<Row><Col span={2}>label</Col><Col span={22}>control
 * </Col></Row>`, and the first port kept that shape: a right-aligned label in a
 * fixed 190px column and a control filling the rest. In a 1600px page that
 * stretched every text field to well over a thousand pixels, which is both ugly
 * and unreadable — the eye has to travel the whole width to get from the label to
 * the end of the value.
 *
 * So the label now sits above its control and the control is capped at a
 * comfortable measure. `block` rows (tables, code editors, transfer lists) opt out
 * of the cap because they genuinely want the width.
 */
export function FormRow({label, labelKey, tooltip, children, className, block, htmlFor}: FormRowProps) {
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
    <div className={cn("space-y-1.5 py-2.5", className)}>
      <label htmlFor={htmlFor} className="flex items-center gap-1.5 text-sm font-medium leading-none">
        <span>{resolvedLabel}</span>
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
      <div className={cn("min-w-0", !block && "max-w-xl")}>{children}</div>
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
    <section className={cn("space-y-2", className)}>
      {title ? (
        <header className="space-y-0.5">
          <h3 className="text-sm font-semibold tracking-tight">{title}</h3>
          {description ? <p className="text-sm text-muted-foreground">{description}</p> : null}
        </header>
      ) : null}
      {/* the rows carry their own rhythm now; a rule between each one turned the
          form into a spreadsheet */}
      <div>{children}</div>
    </section>
  );
}
