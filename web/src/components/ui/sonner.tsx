import * as React from "react";
import i18next from "i18next";
import {CircleCheck, Info, Loader2, OctagonX, TriangleAlert} from "lucide-react";
import {Toaster as SonnerToaster} from "sonner";
import {useTheme} from "@/hooks/use-theme";

export function Toaster() {
  const {resolvedTheme} = useTheme();
  return (
    <SonnerToaster
      theme={resolvedTheme as "light" | "dark"}
      position="top-center"
      containerAriaLabel={i18next.t("general:Notifications")}
      className="toaster group"
      icons={{
        success: <CircleCheck className="size-4 text-success" />,
        info: <Info className="size-4 text-sky-600 dark:text-sky-400" />,
        warning: <TriangleAlert className="size-4 text-warning" />,
        error: <OctagonX className="size-4 text-destructive" />,
        loading: <Loader2 className="size-4 animate-spin" />,
      }}
      style={
        {
          "--normal-bg": "hsl(var(--popover))",
          "--normal-text": "hsl(var(--popover-foreground))",
          "--normal-border": "hsl(var(--border))",
          "--border-radius": "var(--radius)",
        } as React.CSSProperties
      }
    />
  );
}
