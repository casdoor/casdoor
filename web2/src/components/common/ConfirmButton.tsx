import * as React from "react";
import i18next from "i18next";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import {Button, type ButtonProps} from "@/components/ui/button";
import {cn} from "@/lib/utils";

interface ConfirmButtonProps extends Omit<ButtonProps, "onClick" | "title"> {
  title?: React.ReactNode;
  description?: React.ReactNode;
  confirmText?: React.ReactNode;
  cancelText?: React.ReactNode;
  onConfirm: () => void | Promise<void>;
  destructive?: boolean;
}

/** A button that asks for confirmation before running its action. */
export function ConfirmButton({
  title,
  description,
  confirmText,
  cancelText,
  onConfirm,
  destructive = true,
  children,
  ...buttonProps
}: ConfirmButtonProps) {
  const [open, setOpen] = React.useState(false);
  const [busy, setBusy] = React.useState(false);

  return (
    <AlertDialog open={open} onOpenChange={setOpen}>
      <AlertDialogTrigger asChild>
        <Button {...buttonProps}>{children}</Button>
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{title ?? i18next.t("general:Sure to delete") + "?"}</AlertDialogTitle>
          {description ? <AlertDialogDescription>{description}</AlertDialogDescription> : null}
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={busy}>{cancelText ?? i18next.t("general:Cancel")}</AlertDialogCancel>
          <AlertDialogAction
            className={cn(destructive && "bg-destructive text-destructive-foreground hover:bg-destructive/90")}
            disabled={busy}
            onClick={async(e) => {
              e.preventDefault();
              setBusy(true);
              try {
                await onConfirm();
                setOpen(false);
              } finally {
                setBusy(false);
              }
            }}
          >
            {confirmText ?? i18next.t("general:OK")}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
