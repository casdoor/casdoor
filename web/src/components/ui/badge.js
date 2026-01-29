import * as React from "react";
import {cva} from "class-variance-authority";
import {cn} from "../../lib/utils";

const badgeVariants = cva(
  "inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-slate-950 focus:ring-offset-2",
  {
    variants: {
      variant: {
        default: "border-transparent bg-slate-900 text-slate-50 hover:bg-slate-900/80",
        secondary: "border-transparent bg-slate-100 text-slate-900 hover:bg-slate-100/80",
        destructive: "border-transparent bg-red-500 text-slate-50 hover:bg-red-500/80",
        success: "border-transparent bg-green-500 text-slate-50 hover:bg-green-500/80",
        warning: "border-transparent bg-yellow-500 text-slate-50 hover:bg-yellow-500/80",
        outline: "text-slate-950",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
);

const Badge = React.forwardRef(({className, variant, ...props}, ref) => (
  <div ref={ref} className={cn(badgeVariants({variant}), className)} {...props} />
));
Badge.displayName = "Badge";

export {Badge, badgeVariants};
