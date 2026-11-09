import * as React from "react";
import i18next from "i18next";
import {Download, Eye, Trash2} from "lucide-react";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Label} from "@/components/ui/label";
import {Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle} from "@/components/ui/dialog";
import * as UserBackend from "@/backend/UserBackend";
import * as Setting from "@/lib/setting";

/** Shows the stored value of a field the console masks; the backend records who asked. */
export function RevealButton({userId, field, value, onReveal}: {
  userId: string;
  field: "phone" | "email" | "idCard";
  value?: string;
  onReveal: (value: string) => void;
}) {
  const [loading, setLoading] = React.useState(false);
  if (!value || !value.includes("*")) {
    return null;
  }
  return (
    <Button
      type="button"
      variant="ghost"
      size="icon"
      className="shrink-0"
      title={i18next.t("user:Show the full value")}
      disabled={loading}
      onClick={() => {
        setLoading(true);
        UserBackend.getUserPersonalInfo(userId, field)
          .then((res: any) => {
            if (res.status === "ok") {
              onReveal(res.data);
            } else {
              Setting.showMessage("error", res.msg);
            }
          })
          .finally(() => setLoading(false));
      }}
    >
      <Eye className="h-4 w-4" />
    </Button>
  );
}

/** "Export my data" and "Delete my account" on the user's own account page, when the organization allows them. */
export function PersonalDataActions({user, organization}: {user: any; organization: any}) {
  const policy = organization?.securityPolicy;
  const [open, setOpen] = React.useState(false);
  const [password, setPassword] = React.useState("");
  const [deleting, setDeleting] = React.useState(false);
  const [exporting, setExporting] = React.useState(false);

  if (!policy?.allowDataExport && !policy?.allowAccountDeletion) {
    return null;
  }

  const exportData = () => {
    setExporting(true);
    UserBackend.exportOwnData()
      .then((res: any) => {
        if (res.status !== "ok") {
          Setting.showMessage("error", res.msg);
          return;
        }
        const blob = new Blob([JSON.stringify(res.data, null, 2)], {type: "application/json"});
        const link = document.createElement("a");
        link.href = URL.createObjectURL(blob);
        link.download = `${user.owner}-${user.name}.json`;
        link.click();
        URL.revokeObjectURL(link.href);
      })
      .finally(() => setExporting(false));
  };

  const deleteAccount = () => {
    setDeleting(true);
    UserBackend.deleteOwnAccount(password)
      .then((res: any) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("user:Your account has been deleted"));
          window.location.href = "/";
        } else {
          Setting.showMessage("error", res.msg);
        }
      })
      .finally(() => setDeleting(false));
  };

  return (
    <div className="col-span-full flex flex-wrap gap-2 border-t pt-4">
      {policy.allowDataExport ? (
        <Button type="button" variant="outline" loading={exporting} onClick={exportData}>
          <Download className="h-4 w-4" />
          {i18next.t("user:Export my data")}
        </Button>
      ) : null}
      {policy.allowAccountDeletion && !user.isAdmin ? (
        <>
          <Button type="button" variant="outline" className="text-destructive" onClick={() => setOpen(true)}>
            <Trash2 className="h-4 w-4" />
            {i18next.t("user:Delete my account")}
          </Button>
          <Dialog open={open} onOpenChange={setOpen}>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>{i18next.t("user:Delete my account")}</DialogTitle>
              </DialogHeader>
              <p className="text-sm text-muted-foreground">
                {i18next.t("user:Your account and its sign-in sessions will be deleted, and you can't sign in with it any more")}
              </p>
              <div className="space-y-2">
                <Label htmlFor="delete-password">{i18next.t("general:Password")}</Label>
                <Input id="delete-password" type="password" value={password} onChange={(e) => setPassword(e.target.value)} />
              </div>
              <DialogFooter>
                <Button variant="outline" onClick={() => setOpen(false)}>{i18next.t("general:Cancel")}</Button>
                <Button variant="destructive" loading={deleting} onClick={deleteAccount}>
                  {i18next.t("general:Delete")}
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </>
      ) : null}
    </div>
  );
}
