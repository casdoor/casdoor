import * as React from "react";
import {cn} from "@/lib/utils";

export interface DescriptionItem {
  key?: string;
  label: React.ReactNode;
  /** how many of `columns` this item takes up; the label counts as part of it */
  span?: number;
  children: React.ReactNode;
  /** skip the row entirely */
  hidden?: boolean;
}

interface DescriptionListProps {
  items: DescriptionItem[];
  /** number of label/value pairs per row on md and up; always 1 below that */
  columns?: 1 | 2 | 3;
  className?: string;
}

/**
 * The shadcn stand-in for antd's `<Descriptions bordered size="small">`, which
 * the entry viewers lean on heavily. Cell borders come from a `gap-px` grid over
 * a `bg-border` background, so they stay hairline-thin at any zoom. Below `md`
 * it collapses to antd's "vertical" layout: one item per row, label above value.
 */
export function DescriptionList({items, columns = 1, className}: DescriptionListProps) {
  const visible = items.filter((item) => !item.hidden);
  if (visible.length === 0) {
    return null;
  }

  return (
    <div className={cn("overflow-hidden rounded-md border bg-border", className)}>
      <div
        className="grid gap-px md:[grid-template-columns:var(--desc-cols)]"
        style={{"--desc-cols": `repeat(${columns}, minmax(110px, max-content) minmax(0, 1fr))`} as React.CSSProperties}
      >
        {visible.map((item, index) => {
          const span = Math.min(Math.max(item.span ?? 1, 1), columns);
          return (
            <React.Fragment key={item.key ?? index}>
              <div className="bg-muted/60 px-3 py-2 text-sm font-medium text-muted-foreground">
                {item.label}
              </div>
              <div
                className="min-w-0 break-words bg-background px-3 py-2 text-sm md:[grid-column:var(--desc-span)]"
                style={{"--desc-span": `span ${span * 2 - 1}`} as React.CSSProperties}
              >
                {item.children}
              </div>
            </React.Fragment>
          );
        })}
      </div>
    </div>
  );
}

export default DescriptionList;
