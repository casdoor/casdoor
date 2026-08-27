import * as React from "react";
import i18next from "i18next";
import {ArrowLeft} from "lucide-react";
import {useNavigate} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {Card, CardContent, CardHeader} from "@/components/ui/card";
import {PageHeader} from "@/components/crud/PageHeader";
import type {EditMode} from "@/lib/crud";
import {cn} from "@/lib/utils";

interface EditPageShellProps {
  title: React.ReactNode;
  description?: React.ReactNode;
  mode: EditMode;
  backTo: string;
  onSave: (exitAfterSave: boolean) => void | Promise<void>;
  /** in "add" mode Cancel simply leaves the page, dropping the unsaved object */
  onCancel?: () => void;
  extraActions?: React.ReactNode;
  children: React.ReactNode;
  saving?: boolean;
  className?: string;
}

export function EditPageShell({
  title,
  description,
  mode,
  backTo,
  onSave,
  onCancel,
  extraActions,
  children,
  saving,
  className,
}: EditPageShellProps) {
  const navigate = useNavigate();

  const footer = (
    <div className="flex flex-wrap items-center gap-2">
      {extraActions}
      <Button variant="outline" loading={saving} onClick={() => onSave(false)}>
        {i18next.t("general:Save")}
      </Button>
      <Button loading={saving} onClick={() => onSave(true)}>
        {i18next.t("general:Save & Exit")}
      </Button>
      {mode === "add" ? (
        <Button variant="ghost" onClick={() => (onCancel ? onCancel() : navigate(backTo))}>
          {i18next.t("general:Cancel")}
        </Button>
      ) : null}
    </div>
  );

  return (
    <div className={cn("space-y-4", className)}>
      <PageHeader
        title={
          <span className="flex items-center gap-2">
            <Button variant="ghost" size="iconSm" onClick={() => navigate(backTo)} aria-label={i18next.t("general:Back")}>
              <ArrowLeft />
            </Button>
            {title}
          </span>
        }
        description={description}
        actions={footer}
      />
      <Card>
        <CardContent className="pt-6">{children}</CardContent>
      </Card>
      <div className="flex justify-end pb-8">{footer}</div>
    </div>
  );
}

export {CardHeader};
