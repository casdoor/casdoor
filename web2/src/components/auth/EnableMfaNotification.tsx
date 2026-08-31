import * as React from "react";
import i18next from "i18next";
import {useLocation, useNavigate} from "react-router-dom";
import {toast} from "sonner";
import {Badge} from "@/components/ui/badge";
import {Button} from "@/components/ui/button";
import {useAccount} from "@/hooks/use-account";
import * as Setting from "@/lib/setting";

/**
 * antd's `notification.open` auto-closed after 4.5s. These toasts carry buttons
 * the user is meant to press, so they get longer on screen than a plain message.
 */
const NOTIFICATION_DURATION = 12000;

/** Fixed ids so a re-run (StrictMode, a second account fetch) replaces the toast rather than stacking one. */
const REQUIRED_TOAST_ID = "enable-mfa-required";
const PROMPTED_TOAST_ID = "enable-mfa-prompted";

function MfaTypeTags({mfaTypes}: {mfaTypes: string[]}) {
  return (
    <div className="mt-2 flex flex-wrap gap-1">
      {mfaTypes.map((item) => (
        <Badge key={item} variant="warning">{item}</Badge>
      ))}
    </div>
  );
}

/**
 * Port of `web/src/common/notifaction/EnableMfaNotification.js`.
 *
 * Right after sign-in, an organization that marks an MFA item as "Prompted" gets
 * a toast recommending the user turn it on; one marked "Required" gets a
 * read-only warning (the redirect to `/mfa/setup` itself lives in `App.tsx`).
 */
export function EnableMfaNotification() {
  const {account} = useAccount();
  const navigate = useNavigate();
  const location = useLocation();
  const from = (location.state as {from?: string} | null)?.from;

  React.useEffect(() => {
    if (!account || from !== "/login") {
      return;
    }

    const mfaItems = Setting.getMfaItemsByRules(account, account.organization, [
      Setting.MfaRuleRequired,
      Setting.MfaRulePrompted,
    ]);
    if (mfaItems.length === 0) {
      return;
    }

    const requiredItem = mfaItems.find((item: any) => item.rule === Setting.MfaRuleRequired);
    if (requiredItem) {
      toast.warning(i18next.t("mfa:Enable multi-factor authentication"), {
        id: REQUIRED_TOAST_ID,
        duration: NOTIFICATION_DURATION,
        description: (
          <>
            {i18next.t("mfa:To ensure the security of your account, it is required to enable multi-factor authentication")}
            <MfaTypeTags mfaTypes={[requiredItem.name]} />
          </>
        ),
        action: (
          <Button size="sm" onClick={() => toast.dismiss(REQUIRED_TOAST_ID)}>
            {i18next.t("general:Confirm")}
          </Button>
        ),
      });
      return;
    }

    const promptedTypes = mfaItems
      .filter((item: any) => item.rule === Setting.MfaRulePrompted)
      .map((item: any) => item.name);
    if (promptedTypes.length === 0) {
      return;
    }

    toast.info(i18next.t("mfa:Enable multi-factor authentication"), {
      id: PROMPTED_TOAST_ID,
      duration: NOTIFICATION_DURATION,
      description: (
        <>
          {i18next.t("mfa:To ensure the security of your account, it is recommended that you enable multi-factor authentication")}
          <MfaTypeTags mfaTypes={promptedTypes} />
        </>
      ),
      action: (
        <div className="flex items-center gap-1">
          <Button variant="ghost" size="sm" onClick={() => toast.dismiss(PROMPTED_TOAST_ID)}>
            {i18next.t("general:Later")}
          </Button>
          <Button
            size="sm"
            onClick={() => {
              toast.dismiss(PROMPTED_TOAST_ID);
              navigate(`/mfa/setup?mfaType=${promptedTypes[0]}`, {state: {from: "/"}});
            }}
          >
            {i18next.t("general:Go to enable")}
          </Button>
        </div>
      ),
    });
  }, [account, from, navigate]);

  return null;
}

export default EnableMfaNotification;
