import * as React from "react";
import i18next from "i18next";

/** Auto-submitting form used by the SAML POST binding. */
export function RedirectForm({
  samlResponse,
  redirectUrl,
  relayState,
}: {
  samlResponse: string;
  redirectUrl: string;
  relayState: string;
}) {
  React.useEffect(() => {
    (document.getElementById("saml") as HTMLFormElement | null)?.submit();
  }, []);

  return (
    <div className="flex min-h-screen items-center justify-center">
      <p className="text-muted-foreground">{i18next.t("login:Redirecting, please wait.")}</p>
      <form id="saml" method="post" action={redirectUrl}>
        <input type="hidden" name="SAMLResponse" id="samlResponse" value={samlResponse} readOnly />
        <input type="hidden" name="RelayState" id="relayState" value={relayState} readOnly />
      </form>
    </div>
  );
}

export default RedirectForm;
