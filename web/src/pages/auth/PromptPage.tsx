import * as React from "react";
import i18next from "i18next";
import {useNavigate, useParams, useSearchParams} from "react-router-dom";
import {Alert, AlertDescription} from "@/components/ui/alert";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Label} from "@/components/ui/label";
import {CountryCodeSelect} from "@/components/common/CountryCodeSelect";
import {Loading} from "@/components/common/Loading";
import {RegionSelect} from "@/components/common/RegionSelect";
import {AuthLayout} from "@/components/auth/AuthLayout";
import {SendCodeInput} from "@/components/auth/SendCodeInput";
import {AffiliationAddressSelect, AffiliationField, useAffiliation} from "@/components/user/AffiliationSelect";
import {ThirdPartyLogins} from "@/components/user/OAuthWidget";
import {useAccount} from "@/hooks/use-account";
import {authConfig} from "@/auth/Auth";
import * as ApplicationBackend from "@/backend/ApplicationBackend";
import * as UserBackend from "@/backend/UserBackend";
import * as Setting from "@/lib/setting";

/**
 * Collects what an application marks as "prompted" after sign-in — the third-party
 * accounts it wants bound plus the profile fields it still needs — then continues
 * the OAuth redirect it was interrupted from.
 */
export default function PromptPage({application: applicationProp}: {application?: any} = {}) {
  const params = useParams();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const {account, reload} = useAccount();

  const applicationName = params.applicationName ?? authConfig.appName;
  const [application, setApplication] = React.useState<any>(undefined);
  const [user, setUser] = React.useState<any>(null);
  const [savedUser, setSavedUser] = React.useState<any>(null);
  const [emailCode, setEmailCode] = React.useState("");
  const [phoneCode, setPhoneCode] = React.useState("");
  const [saving, setSaving] = React.useState(false);

  React.useEffect(() => {
    // the application editor's preview hands its own object over
    if (applicationProp) {
      setApplication(applicationProp);
      return;
    }
    ApplicationBackend.getApplication("admin", applicationName)
      .then((res: any) => setApplication(res.status === "ok" ? res.data : null))
      .catch(() => setApplication(null));
  }, [applicationName, applicationProp]);

  const loadUser = React.useCallback(() => {
    if (!account) {
      return;
    }
    UserBackend.getUser(account.owner, account.name).then((res: any) => {
      if (res.status === "ok") {
        setUser(res.data);
        setSavedUser(res.data);
      }
    });
  }, [account]);

  React.useEffect(loadUser, [loadUser]);

  const affiliation = useAffiliation(application, user);

  const getRedirectUrl = () => {
    const redirectUri = searchParams.get("redirectUri");
    const code = searchParams.get("code");
    const state = searchParams.get("state");
    const oauth = searchParams.get("oauth");
    if (redirectUri === null || code === null || state === null) {
      return oauth === "true" ? sessionStorage.getItem("signinUrl") ?? "" : "";
    }
    return `${redirectUri}?code=${encodeURIComponent(code)}&state=${encodeURIComponent(state)}${Setting.getOAuthIssuerParam()}`;
  };

  const finishAndJump = () => {
    const redirectUrl = getRedirectUrl();
    if (redirectUrl) {
      Setting.goToLink(redirectUrl);
      return;
    }
    const loginLink = Setting.getLoginLink(application);
    if (loginLink.startsWith("http://") || loginLink.startsWith("https://")) {
      Setting.goToLink(loginLink);
    } else {
      navigate(loginLink);
    }
  };

  if (application === undefined || (account && user === null)) {
    return <Loading className="min-h-screen" />;
  }

  // Nothing is prompted, so the visitor should never have landed here.
  if (application && !Setting.hasPromptPage(application)) {
    return (
      <AuthLayout preview={!!applicationProp} application={application}>
        <div className="space-y-4">
          <Alert variant="warning">
            <AlertDescription>{i18next.t("application:You are unexpected to see this prompt page")}</AlertDescription>
          </Alert>
          <Button className="w-full" onClick={finishAndJump}>
            {i18next.t("login:Sign In")}
          </Button>
        </div>
      </AuthLayout>
    );
  }

  const update = (fields: Record<string, any>) => setUser((prev: any) => ({...prev, ...fields}));

  // update-user drops email and phone for a non-admin, so a new one is confirmed with a code
  // through /api/reset-email-or-phone, the same way the account page changes them
  const emailChanged = (user?.email ?? "") !== (savedUser?.email ?? "");
  const phoneChanged = (user?.phone ?? "") !== (savedUser?.phone ?? "");

  const submit = async() => {
    setSaving(true);
    try {
      const res: any = await UserBackend.updateUser(user.owner, user.name, {
        ...Setting.deepCopy(user),
        email: savedUser.email,
        phone: savedUser.phone,
        countryCode: savedUser.countryCode,
      });
      if (res.status !== "ok") {
        Setting.showMessage("error", res.msg);
        return;
      }
      for (const [changed, dest, destType, code] of [
        [emailChanged, user.email, "email", emailCode],
        [phoneChanged, user.phone, "phone", phoneCode],
      ]) {
        if (!changed) {
          continue;
        }
        const resetRes: any = await UserBackend.resetEmailOrPhone(dest, destType, code, phoneCountryCode);
        if (resetRes.status !== "ok") {
          Setting.showMessage("error", resetRes.msg);
          return;
        }
      }
      Setting.showMessage("success", i18next.t("general:Successfully saved"));
      reload();
      finishAndJump();
    } catch (error) {
      Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
    } finally {
      setSaving(false);
    }
  };

  const promptedItems = (application?.signupItems ?? []).filter((item: any) => Setting.isSignupItemPrompted(item));
  const defaultCountryCode = application?.organizationObj?.countryCodes?.[0] ?? "";
  const phoneCountryCode = user?.countryCode || defaultCountryCode;
  const hasPromptedProviders = (application?.providers ?? []).some((item: any) => Setting.isProviderPrompted(item));

  return (
    <AuthLayout preview={!!applicationProp} application={application}>
      <div className="space-y-5">
        <h1 className="text-center text-lg font-semibold">{i18next.t("application:Binding providers")}</h1>

        {hasPromptedProviders && user ? (
          <ThirdPartyLogins
            user={user}
            application={application}
            account={account}
            filter={(providerItem: any) => Setting.isProviderPrompted(providerItem)}
            onUnlinked={loadUser}
          />
        ) : null}

        {Setting.isAffiliationPrompted(application) ? (
          <div className="space-y-3">
            {affiliation.enabled ? (
              <div className="space-y-2">
                <Label>{i18next.t("user:Address")}</Label>
                <AffiliationAddressSelect
                  value={user?.address}
                  options={affiliation.addressOptions}
                  onChange={(value) => {
                    // a new address invalidates the affiliation picked under the old one
                    update({address: value, affiliation: "", score: 0});
                    affiliation.loadAffiliationOptions(value);
                  }}
                />
              </div>
            ) : null}
            <div className="space-y-2">
              <Label>{i18next.t("user:Affiliation")}</Label>
              <AffiliationField
                enabled={affiliation.enabled}
                value={user?.affiliation}
                options={affiliation.affiliationOptions}
                onChange={(name, score) =>
                  score === undefined ? update({affiliation: name}) : update({affiliation: name, score})
                }
              />
            </div>
          </div>
        ) : null}

        {promptedItems.map((item: any) => {
          if (item.name === "Country/Region") {
            return (
              <div key={item.name} className="space-y-2">
                <Label>{i18next.t("user:Country/Region")}</Label>
                <RegionSelect value={user?.region ?? ""} onChange={(value) => update({region: value})} />
              </div>
            );
          }
          if (item.name === "Display name") {
            return (
              <div key={item.name} className="space-y-2">
                <Label htmlFor="displayName">{item.label || i18next.t("general:Display name")}</Label>
                <Input
                  id="displayName"
                  placeholder={item.placeholder}
                  value={user?.displayName ?? ""}
                  onChange={(e) => update({displayName: e.target.value})}
                />
              </div>
            );
          }
          if (item.name === "Email") {
            return (
              <div key={item.name} className="space-y-2">
                <Label htmlFor="email">{item.label || i18next.t("general:Email")}</Label>
                <Input
                  id="email"
                  type="email"
                  autoComplete="email"
                  placeholder={item.placeholder}
                  value={user?.email ?? ""}
                  onChange={(e) => update({email: e.target.value})}
                />
                {emailChanged ? (
                  <SendCodeInput
                    value={emailCode}
                    onChange={setEmailCode}
                    method="reset"
                    destType="email"
                    dest={user?.email ?? ""}
                    disabled={!Setting.isValidEmail(user?.email ?? "")}
                    application={application}
                    applicationId={Setting.getApplicationName(application)}
                  />
                ) : null}
              </div>
            );
          }
          if (item.name === "Phone") {
            return (
              <div key={item.name} className="space-y-2">
                <Label htmlFor="phone">{item.label || i18next.t("general:Phone")}</Label>
                <div className="flex gap-2">
                  <div className="w-28 max-w-[50%] shrink-0">
                    <CountryCodeSelect
                      className="px-2"
                      value={phoneCountryCode}
                      onChange={(value) => update({countryCode: value})}
                      countryCodes={application?.organizationObj?.countryCodes}
                    />
                  </div>
                  <Input
                    id="phone"
                    className="min-w-0 flex-1"
                    autoComplete="tel"
                    placeholder={item.placeholder}
                    value={user?.phone ?? ""}
                    // an account a provider created has no calling code yet
                    onChange={(e) => update({phone: e.target.value, countryCode: phoneCountryCode})}
                  />
                </div>
                {phoneChanged ? (
                  <SendCodeInput
                    value={phoneCode}
                    onChange={setPhoneCode}
                    method="reset"
                    destType="phone"
                    dest={user?.phone ?? ""}
                    countryCode={phoneCountryCode}
                    disabled={!Setting.isValidPhone(user?.phone ?? "", phoneCountryCode)}
                    application={application}
                    applicationId={Setting.getApplicationName(application)}
                  />
                ) : null}
              </div>
            );
          }
          return null;
        })}

        <Button
          className="w-full"
          loading={saving}
          disabled={saving || !Setting.isPromptAnswered(user, application) || (emailChanged && emailCode === "") || (phoneChanged && phoneCode === "")}
          onClick={submit}
        >
          {i18next.t("code:Submit and complete")}
        </Button>
      </div>
    </AuthLayout>
  );
}
