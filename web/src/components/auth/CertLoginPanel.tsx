import * as React from "react";
import i18next from "i18next";
import {KeyRound} from "lucide-react";
import {Button} from "@/components/ui/button";
import * as Setting from "@/lib/setting";

/** The params that the certificate sign-in port adds to the sign-in page it goes back to. */
export const CertLoginParams = ["certLoginToken", "certLoginError"];

/**
 * The address of the certificate sign-in: the port that asks the browser for a client certificate (the
 * organization's "Certificate login URL"), or this site's own /api/login/cert behind a reverse proxy that
 * passes the certificate on. It comes back to this page with a one-time ticket.
 */
export function getCertLoginUrl(application: any): string {
  const configured = application?.organizationObj?.enterprisePolicy?.certLogin?.loginUrl;
  const target = new URL(configured || `${Setting.ServerUrl}/api/login/cert`, window.location.href);

  const back = new URL(window.location.href);
  CertLoginParams.forEach((param) => back.searchParams.delete(param));

  target.searchParams.set("application", application?.name ?? "");
  if (application?.isShared && application?.organization) {
    target.searchParams.set("organization", application.organization);
  }
  target.searchParams.set("return", back.toString());
  return target.toString();
}

/** The "UKey or certificate" sign-in method: the browser shows the certificates of the inserted UKey. */
export function CertLoginPanel({application}: {application: any}) {
  const [leaving, setLeaving] = React.useState(false);
  return (
    <div className="cert-login-panel space-y-3 text-center">
      <KeyRound className="mx-auto size-10 text-muted-foreground" />
      <p className="text-sm text-muted-foreground">{i18next.t("login:Insert your UKey, then choose its certificate when the browser asks for one")}</p>
      <Button
        type="button"
        className="cert-login-button w-full"
        loading={leaving}
        onClick={() => {
          setLeaving(true);
          window.location.href = getCertLoginUrl(application);
        }}
      >
        {i18next.t("login:Sign in with UKey or certificate")}
      </Button>
    </div>
  );
}
