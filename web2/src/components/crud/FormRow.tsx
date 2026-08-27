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
  /** stack the control under the label (for tables and editors) */
  block?: boolean;
  htmlFor?: string;
}

/**
 * The label/control row used by every Casdoor edit page. The antd version was a
 * `<Row><Col span={2}>label</Col><Col span={22}>control</Col></Row>`; here it is a
 * responsive grid that collapses to a single column on small screens.
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
    <div
      className={cn(
        "grid items-start gap-x-4 gap-y-1.5 py-3",
        block ? "grid-cols-1" : "grid-cols-1 md:grid-cols-[minmax(150px,190px)_1fr]",
        className,
      )}
    >
      <label
        htmlFor={htmlFor}
        className="flex items-center gap-1 pt-2 text-sm font-medium text-muted-foreground md:justify-end md:text-right"
      >
        <span>{resolvedLabel}</span>
        {resolvedTooltip ? (
          <Tooltip>
            <TooltipTrigger asChild>
              <span className="cursor-help text-muted-foreground/70">
                <HelpCircle className="h-3.5 w-3.5" />
              </span>
            </TooltipTrigger>
            <TooltipContent>{resolvedTooltip}</TooltipContent>
          </Tooltip>
        ) : null}
      </label>
      <div className="min-w-0">{children}</div>
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
    <section className={cn("space-y-1", className)}>
      {title ? (
        <header className="border-b pb-2">
          <h3 className="text-sm font-semibold">{title}</h3>
          {description ? <p className="text-xs text-muted-foreground">{description}</p> : null}
        </header>
      ) : null}
      <div className="divide-y divide-border/60">{children}</div>
    </section>
  );
}
