import * as React from "react";
import i18next from "i18next";
import {useParams} from "react-router-dom";
import {CheckCircle2, XCircle} from "lucide-react";
import {Alert, AlertDescription} from "@/components/ui/alert";
import {Button} from "@/components/ui/button";
import {AuthLayout} from "@/components/auth/AuthLayout";
import * as MfaBackend from "@/backend/MfaBackend";

/**
 * Opened from the card sent to WeCom, DingTalk, Lark or a notification app: the user approves or denies the
 * sign-in that waits for the push MFA. The answer is a button press, so the link previews of chat apps that
 * fetch the URL don't approve anything.
 */
export default function PushConfirmPage() {
  const {id = ""} = useParams();
  const [challenge, setChallenge] = React.useState<any>(undefined);
  const [error, setError] = React.useState("");
  const [answering, setAnswering] = React.useState(false);

  React.useEffect(() => {
    MfaBackend.getMfaPushChallenge(id)
      .then((res: any) => {
        if (res.status === "ok") {
          setChallenge(res.data);
        } else {
          setError(res.msg);
          setChallenge(null);
        }
      })
      .catch((err: any) => {
        setError(err.message);
        setChallenge(null);
      });
  }, [id]);

  const answer = (approve: boolean) => {
    setAnswering(true);
    MfaBackend.answerMfaPushChallenge(id, approve)
      .then((res: any) => {
        if (res.status === "ok") {
          setChallenge({...challenge, status: res.data});
        } else {
          setError(res.msg);
        }
      })
      .catch((err: any) => setError(err.message))
      .finally(() => setAnswering(false));
  };

  const renderBody = () => {
    if (challenge === undefined) {
      return <p className="text-center text-sm text-muted-foreground">{i18next.t("general:Loading")}</p>;
    }
    if (challenge === null) {
      return (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      );
    }
    if (challenge.status === "Approved" || challenge.status === "Denied") {
      const approved = challenge.status === "Approved";
      return (
        <div className="flex flex-col items-center gap-3 py-4 text-center">
          {approved ? <CheckCircle2 className="h-10 w-10 text-success" /> : <XCircle className="h-10 w-10 text-destructive" />}
          <p className="text-sm">
            {approved ? i18next.t("mfa:The sign-in is approved, you can go back to the sign-in page") : i18next.t("mfa:The sign-in is denied")}
          </p>
        </div>
      );
    }

    const rows: [string, string][] = [
      [i18next.t("general:User"), challenge.userId],
      [i18next.t("general:Application"), challenge.application],
      [i18next.t("general:IP"), `${challenge.ip} ${challenge.location ?? ""}`],
      [i18next.t("general:Created time"), new Date(challenge.createdTime).toLocaleString()],
    ];
    return (
      <div className="space-y-4">
        <dl className="space-y-2 rounded-md border p-3 text-sm">
          {rows.filter(([, value]) => value && value.trim()).map(([label, value]) => (
            <div key={label} className="flex gap-3">
              <dt className="w-24 shrink-0 text-muted-foreground">{label}</dt>
              <dd className="break-all">{value}</dd>
            </div>
          ))}
        </dl>
        {challenge.status === "Expired" ? (
          <Alert variant="destructive">
            <AlertDescription>{i18next.t("mfa:The sign-in request has expired")}</AlertDescription>
          </Alert>
        ) : (
          <div className="grid grid-cols-2 gap-3">
            <Button variant="outline" disabled={answering} onClick={() => answer(false)}>
              {i18next.t("permission:Deny")}
            </Button>
            <Button loading={answering} onClick={() => answer(true)}>
              {i18next.t("mfa:Approve")}
            </Button>
          </div>
        )}
        {error ? (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        ) : null}
      </div>
    );
  };

  return (
    <AuthLayout>
      <div className="space-y-4">
        <h2 className="text-center text-xl font-semibold tracking-tight">{i18next.t("mfa:Confirm your sign-in")}</h2>
        <p className="text-center text-sm text-muted-foreground">{i18next.t("mfa:Approve it only if you are signing in right now")}</p>
        {renderBody()}
      </div>
    </AuthLayout>
  );
}
