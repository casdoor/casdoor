import * as React from "react";
import i18next from "i18next";
import {useNavigate, useParams, useSearchParams} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Label} from "@/components/ui/label";
import {Loading} from "@/components/common/Loading";
import {AuthLayout} from "@/components/auth/AuthLayout";
import {useAccount} from "@/hooks/use-account";
import {authConfig} from "@/auth/Auth";
import * as ApplicationBackend from "@/backend/ApplicationBackend";
import * as UserBackend from "@/backend/UserBackend";
import * as Setting from "@/lib/setting";

/**
 * Collects the profile fields an application marks as "prompted" after sign-in,
 * then continues the OAuth redirect it was interrupted from.
 */
export default function PromptPage() {
  const params = useParams();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const {account, reload} = useAccount();

  const applicationName = params.applicationName ?? authConfig.appName;
  const [application, setApplication] = React.useState<any>(undefined);
  const [user, setUser] = React.useState<any>(null);
  const [saving, setSaving] = React.useState(false);

  React.useEffect(() => {
    ApplicationBackend.getApplication("admin", applicationName)
      .then((res: any) => setApplication(res.status === "ok" ? res.data : null))
      .catch(() => setApplication(null));
  }, [applicationName]);

  React.useEffect(() => {
    if (!account) {
      return;
    }
    UserBackend.getUser(account.owner, account.name).then((res: any) => {
      if (res.status === "ok") {
        setUser(res.data);
      }
    });
  }, [account]);

  if (application === undefined || (account && user === null)) {
    return <Loading className="min-h-screen" />;
  }

  const getRedirectUrl = () => {
    const redirectUri = searchParams.get("redirectUri");
    const code = searchParams.get("code");
    const state = searchParams.get("state");
    const oauth = searchParams.get("oauth");
    if (redirectUri === null || code === null || state === null) {
      return oauth === "true" ? sessionStorage.getItem("signinUrl") ?? "" : "";
    }
    return `${redirectUri}?code=${code}&state=${state}`;
  };

  const finishAndJump = () => {
    const redirectUrl = getRedirectUrl();
    if (redirectUrl) {
      Setting.goToLink(redirectUrl);
    } else {
      const loginLink = Setting.getLoginLink(application);
      if (loginLink.startsWith("http://") || loginLink.startsWith("https://")) {
        Setting.goToLink(loginLink);
      } else {
        navigate(loginLink);
      }
    }
  };

  const submit = () => {
    setSaving(true);
    UserBackend.updateUser(user.owner, user.name, Setting.deepCopy(user))
      .then((res: any) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully saved"));
          reload();
          finishAndJump();
        } else {
          Setting.showMessage("error", res.msg);
        }
      })
      .catch((error) => Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`))
      .finally(() => setSaving(false));
  };

  const update = (field: string, value: any) => setUser((prev: any) => ({...prev, [field]: value}));

  const promptedItems = (application?.signupItems ?? []).filter((item: any) =>
    Setting.isSignupItemPrompted(item),
  );

  return (
    <AuthLayout application={application}>
      <div className="space-y-5">
        <h1 className="text-center text-lg font-semibold">{i18next.t("user:User Profile")}</h1>

        {Setting.isAffiliationPrompted(application) ? (
          <div className="space-y-2">
            <Label htmlFor="affiliation">{i18next.t("user:Affiliation")}</Label>
            <Input
              id="affiliation"
              value={user?.affiliation ?? ""}
              onChange={(e) => update("affiliation", e.target.value)}
            />
          </div>
        ) : null}

        {promptedItems.map((item: any) => {
          if (item.name === "Country/Region") {
            return (
              <div key={item.name} className="space-y-2">
                <Label htmlFor="region">{i18next.t("user:Country/Region")}</Label>
                <Input id="region" value={user?.region ?? ""} onChange={(e) => update("region", e.target.value)} />
              </div>
            );
          }
          return null;
        })}

        <div className="flex gap-2">
          <Button variant="outline" className="flex-1" onClick={finishAndJump} disabled={saving}>
            {i18next.t("general:Cancel")}
          </Button>
          <Button className="flex-1" loading={saving} onClick={submit}>
            {i18next.t("general:Save")}
          </Button>
        </div>
      </div>
    </AuthLayout>
  );
}
