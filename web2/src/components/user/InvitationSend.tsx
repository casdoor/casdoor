import * as React from "react";
import i18next from "i18next";
import copy from "copy-to-clipboard";
import {Copy} from "lucide-react";
import {Button} from "@/components/ui/button";
import {Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle} from "@/components/ui/dialog";
import {Textarea} from "@/components/ui/textarea";
import * as InvitationBackend from "@/backend/InvitationBackend";
import * as Conf from "@/Conf";
import * as Setting from "@/lib/setting";

/**
 * "Copy signup page URL" + the invitation email sender of the invitation edit page,
 * ported from `web/src/InvitationEditPage.js`. One address per line; only the ones
 * that parse as an email are sent, and the confirmation dialog lists them first.
 */
export function InvitationSend({
  invitation,
  organizations,
  isAdd,
}: {
  invitation: any;
  organizations: any[];
  isAdd: boolean;
}) {
  const [emailsText, setEmailsText] = React.useState("");
  const [open, setOpen] = React.useState(false);
  const [sending, setSending] = React.useState(false);

  const emails = emailsText
    .split("\n")
    .map((line) => line.trim())
    .filter((line) => Setting.isValidEmail(line));

  const copySignupLink = () => {
    let defaultApplication: string | undefined;
    if (invitation.owner === "built-in") {
      defaultApplication = Conf.DefaultApplication;
    } else {
      const selected = (organizations ?? []).find((item: any) => item.name === invitation.owner);
      defaultApplication = selected?.defaultApplication;
      if (!defaultApplication) {
        Setting.showMessage(
          "error",
          i18next.t("invitation:You need to first specify a default application for organization: ") + invitation.owner,
        );
        return;
      }
    }
    copy(`${window.location.origin}/signup/${defaultApplication}?invitationCode=${invitation?.defaultCode}`);
    Setting.showMessage("success", i18next.t("general:Copied to clipboard successfully"));
  };

  const send = () => {
    setSending(true);
    InvitationBackend.sendInvitation(invitation, emails)
      .then((res: any) => {
        if (res.status === "error") {
          Setting.showMessage("error", res.msg);
          return;
        }
        Setting.showMessage("success", i18next.t("general:Successfully sent"));
        setOpen(false);
      })
      .catch((err) => Setting.showMessage("error", err.message))
      .finally(() => setSending(false));
  };

  return (
    <div className="space-y-2">
      <Button variant="outline" size="sm" onClick={copySignupLink}>
        <Copy />
        {i18next.t("application:Copy signup page URL")}
      </Button>
      <Textarea
        rows={3}
        value={emailsText}
        placeholder={i18next.t("general:Email")}
        onChange={(e) => setEmailsText(e.target.value)}
      />
      {/* the emails go out through the saved invitation, so it has to exist first */}
      <Button disabled={isAdd || emails.length === 0} onClick={() => setOpen(true)}>
        {i18next.t("general:Send")}
      </Button>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{i18next.t("general:Send")}</DialogTitle>
          </DialogHeader>
          <div className="space-y-2 text-sm">
            <p>You will send invitation email to:</p>
            <ul className="max-h-64 space-y-1 overflow-auto rounded-lg border p-3 font-mono text-xs">
              {emails.map((email) => (
                <li key={email}>{email}</li>
              ))}
            </ul>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setOpen(false)}>{i18next.t("general:Cancel")}</Button>
            <Button loading={sending} onClick={send}>{i18next.t("general:Send")}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
