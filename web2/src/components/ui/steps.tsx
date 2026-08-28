import * as React from "react";
import {Check} from "lucide-react";
import {cn} from "@/lib/utils";

export interface Step {
  title: React.ReactNode;
  icon?: React.ReactNode;
}

/** Minimal horizontal stepper — the shadcn stand-in for antd's `<Steps>`. */
export function Steps({current, items, className}: {current: number; items: Step[]; className?: string}) {
  return (
    <ol className={cn("flex w-full items-center", className)}>
      {items.map((item, index) => {
        const done = index < current;
        const active = index === current;
        return (
          <li key={index} className={cn("flex items-center", index < items.length - 1 && "flex-1")}>
            <div className="flex items-center gap-2">
              <span
                className={cn(
                  "flex h-8 w-8 shrink-0 items-center justify-center rounded-full border text-sm",
                  done && "border-primary bg-primary text-primary-foreground",
                  active && "border-primary text-primary",
                  !done && !active && "border-border text-muted-foreground",
                )}
              >
                {done ? <Check className="h-4 w-4" /> : (item.icon ?? index + 1)}
              </span>
              <span
                className={cn(
                  "whitespace-nowrap text-sm",
                  active ? "font-medium text-foreground" : "text-muted-foreground",
                )}
              >
                {item.title}
              </span>
            </div>
            {index < items.length - 1 ? (
              <span className={cn("mx-3 h-px flex-1 bg-border", done && "bg-primary")} />
            ) : null}
          </li>
        );
      })}
    </ol>
  );
}
