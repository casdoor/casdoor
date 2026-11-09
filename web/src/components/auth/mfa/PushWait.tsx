import * as React from "react";
import i18next from "i18next";
import {Smartphone} from "lucide-react";
import {Button} from "@/components/ui/button";
import * as MfaBackend from "@/backend/MfaBackend";
import * as Setting from "@/lib/setting";

interface PushWaitProps {
  /** the provider and receiver when setting the factor up, empty at sign-in where the enabled factor is used */
  provider?: string;
  receiver?: string;
  application?: string;
  /** called with the id of the approved request, which is the passcode of the push factor */
  onApproved: (challengeId: string, receiver: string) => void;
  /** sends the card once mounted */
  autoSend?: boolean;
}

/** Sends a confirmation card to the phone and waits until the user approves it there. */
export function PushWait({provider, receiver, application, onApproved, autoSend = true}: PushWaitProps) {
  const [challengeId, setChallengeId] = React.useState("");
  const [sentTo, setSentTo] = React.useState("");
  const [status, setStatus] = React.useState("");
  const [sending, setSending] = React.useState(false);
  const approvedRef = React.useRef(false);

  const send = React.useCallback(() => {
    setSending(true);
    setStatus("");
    approvedRef.current = false;
    MfaBackend.sendMfaPush({provider, receiver, application})
      .then((res: any) => {
        if (res.status === "ok") {
          setChallengeId(res.data);
          setSentTo(res.data2 ?? "");
          setStatus("Pending");
        } else {
          Setting.showMessage("error", res.msg);
        }
      })
      .catch((err: any) => Setting.showMessage("error", err.message))
      .finally(() => setSending(false));
  }, [provider, receiver, application]);

  React.useEffect(() => {
    if (autoSend) {
      send();
    }
    // sends once on mount
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  React.useEffect(() => {
    if (!challengeId || status !== "Pending") {
      return;
    }
    const timer = window.setInterval(() => {
      MfaBackend.getMfaPushChallenge(challengeId).then((res: any) => {
        const next = res.status === "ok" ? res.data.status : "Expired";
        setStatus(next);
        if (next === "Approved" && !approvedRef.current) {
          approvedRef.current = true;
          onApproved(challengeId, sentTo);
        }
      });
    }, 2000);
    return () => window.clearInterval(timer);
  }, [challengeId, status, sentTo, onApproved]);

  const message = (() => {
    switch (status) {
    case "Pending":
      return i18next.t("mfa:A confirmation card has been sent to your phone, please approve the sign-in there");
    case "Approved":
      return i18next.t("mfa:The sign-in is approved, you can go back to the sign-in page");
    case "Denied":
      return i18next.t("mfa:The sign-in is denied");
    case "Expired":
      return i18next.t("mfa:The sign-in request has expired");
    default:
      return i18next.t("mfa:Send a confirmation card to your phone");
    }
  })();

  return (
    <div className="space-y-3">
      <div className="flex items-start gap-3 rounded-md border p-3 text-sm">
        <Smartphone className="mt-0.5 h-5 w-5 shrink-0 text-muted-foreground" />
        <span>{message}</span>
      </div>
      {status !== "Pending" && status !== "Approved" ? (
        <Button type="button" variant="outline" className="w-full" loading={sending} onClick={send}>
          {status ? i18next.t("mfa:Send again") : i18next.t("general:Send")}
        </Button>
      ) : null}
    </div>
  );
}
