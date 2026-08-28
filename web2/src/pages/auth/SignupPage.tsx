import * as React from "react";
import i18next from "i18next";
import {Link, useNavigate, useParams, useSearchParams} from "react-router-dom";
import {Alert, AlertDescription} from "@/components/ui/alert";
import {Button} from "@/components/ui/button";
import {Checkbox} from "@/components/ui/checkbox";
import {Input} from "@/components/ui/input";
import {Label} from "@/components/ui/label";
import {Loading} from "@/components/common/Loading";
import {SearchableSelect} from "@/components/common/SearchableSelect";
import {AuthLayout} from "@/components/auth/AuthLayout";
import {ProviderButtons} from "@/components/auth/ProviderButtons";
import {SendCodeInput} from "@/components/auth/SendCodeInput";
import {CaptchaModal} from "@/components/common/CaptchaModal";
import {getCaptchaProvider} from "@/lib/captcha";
import {authConfig} from "@/auth/Auth";
import * as Util from "@/auth/Util";
import * as ApplicationBackend from "@/backend/ApplicationBackend";
import * as AuthBackend from "@/backend/AuthBackend";
import * as Setting from "@/lib/setting";

/** The signup items Casdoor can render; anything else falls back to a text input. */
const SIMPLE_TEXT_ITEMS: Record<string, string> = {
  "Affiliation": "user:Affiliation",
  "ID card": "user:ID card",
  "Real name": "application:Real name",
  "Bio": "user:Bio",
  "Tag": "user:Tag",
  "Education": "user:Education",
  "Gender": "user:Gender",
  "First name": "general:First name",
  "Last name": "general:Last name",
  "Invitation code": "application:Invitation code",
};

export default function SignupPage() {
  const params = useParams();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();

  const [application, setApplication] = React.useState<any>(undefined);
  const [msg, setMsg] = React.useState<string | null>(null);
  const [values, setValues] = React.useState<Record<string, any>>({});
  const [agreed, setAgreed] = React.useState(false);
  const [loading, setLoading] = React.useState(false);
  const [captchaVisible, setCaptchaVisible] = React.useState(false);
  const [pendingValues, setPendingValues] = React.useState<any>(null);

  const applicationName = params.applicationName ?? authConfig.appName;

  React.useEffect(() => {
    const oAuthParams = Util.getOAuthGetParameters();
    const load = oAuthParams
      ? AuthBackend.getApplicationLogin(oAuthParams)
      : ApplicationBackend.getApplication("admin", applicationName);
    load
      .then((res: any) => {
        if (res.status === "ok") {
          setApplication(res.data);
          setValues((prev) => ({
            ...prev,
            application: res.data.name,
            organization: res.data.organization,
            invitationCode: searchParams.get("invitationCode") ?? "",
          }));
        } else {
          setApplication(null);
          setMsg(res.msg);
        }
      })
      .catch((error) => {
        setApplication(null);
        setMsg(`${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [applicationName]);

  const set = (key: string, value: any) => setValues((prev) => ({...prev, [key]: value}));

  if (application === undefined) {
    return <Loading className="min-h-screen" />;
  }

  if (application === null) {
    return (
      <AuthLayout>
        <Alert variant="destructive">
          <AlertDescription>{msg ?? i18next.t("application:Failed to sign in")}</AlertDescription>
        </Alert>
      </AuthLayout>
    );
  }

  if (!application.enableSignUp) {
    return (
      <AuthLayout application={application}>
        <Alert variant="warning">
          <AlertDescription>{i18next.t("application:The application does not allow to sign up new account")}</AlertDescription>
        </Alert>
      </AuthLayout>
    );
  }

  const captchaProvider = getCaptchaProvider(application);
  const items = (application.signupItems ?? []).filter((item: any) => item.visible);
  const agreementItem = items.find((item: any) => item.name === "Agreement");

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    if (agreementItem?.required && !agreed) {
      Setting.showMessage("error", i18next.t("signup:Please accept the agreement!"));
      return;
    }

    const payload: Record<string, any> = {
      ...values,
      application: application.name,
      organization: application.organization,
      plan: searchParams.get("plan"),
      pricing: searchParams.get("pricing"),
      agreement: agreed,
    };

    const captchaRule = Setting.getCaptchaRule(application);
    if (captchaRule === Setting.CaptchaRule.Always) {
      setPendingValues(payload);
      setCaptchaVisible(true);
      return;
    }
    if (captchaRule === Setting.CaptchaRule.Dynamic || captchaRule === Setting.CaptchaRule.InternetOnly) {
      AuthBackend.getCaptchaStatus({
        organization: application.organization,
        username: payload.username,
        application: application.name,
      })
        .then((res: any) => {
          if (res.status === "ok" && res.data) {
            setPendingValues(payload);
            setCaptchaVisible(true);
          } else {
            submitSignup(payload);
          }
        })
        .catch(() => submitSignup(payload));
      return;
    }

    submitSignup(payload);
  };

  const submitSignup = (payload: Record<string, any>) => {
    setLoading(true);
    const oAuthParams = Util.getOAuthGetParameters();
    AuthBackend.signup(payload, oAuthParams)
      .then((res: any) => {
        if (res.status !== "ok") {
          Setting.showMessage("error", res.msg);
          return;
        }
        // OAuth flow: the backend answers with the authorization code.
        if (oAuthParams && typeof res.data === "string" && !res.data.includes("/")) {
          const redirectUrl = `${oAuthParams.redirectUri}${
            oAuthParams.redirectUri.includes("?") ? "&" : "?"
          }code=${res.data}&state=${oAuthParams.state}`;
          Setting.goToLink(redirectUrl);
          return;
        }
        if (oAuthParams && typeof res.data === "object" && res.data?.required === true) {
          Setting.goToLink(`/consent/${application.name}?${window.location.search.substring(1)}`);
          return;
        }
        let username = payload.username;
        if (typeof res.data === "string") {
          username = res.data.split("/")[1];
        }
        navigate(`/result/${application.name}`, {state: {username}});
      })
      .finally(() => setLoading(false));
  };

  const renderItem = (item: any) => {
    const required = item.required;
    switch (item.name) {
    case "Username":
      return (
        <div key={item.name} className="space-y-2">
          <Label htmlFor="username">{i18next.t("signup:Username")}</Label>
          <Input
            id="username"
            required={required}
            value={values.username ?? ""}
            onChange={(e) => set("username", e.target.value)}
          />
        </div>
      );
    case "Display name":
      return (
        <div key={item.name} className="space-y-2">
          <Label htmlFor="name">{i18next.t("general:Display name")}</Label>
          <Input id="name" required={required} value={values.name ?? ""} onChange={(e) => set("name", e.target.value)} />
        </div>
      );
    case "Password":
      return (
        <div key={item.name} className="space-y-2">
          <Label htmlFor="password">{i18next.t("general:Password")}</Label>
          <Input
            id="password"
            type="password"
            required={required}
            value={values.password ?? ""}
            onChange={(e) => set("password", e.target.value)}
          />
        </div>
      );
    case "Confirm password":
      return (
        <div key={item.name} className="space-y-2">
          <Label htmlFor="confirm">{i18next.t("signup:Confirm")}</Label>
          <Input
            id="confirm"
            type="password"
            required={required}
            value={values.confirm ?? ""}
            onChange={(e) => set("confirm", e.target.value)}
          />
          {values.confirm && values.confirm !== values.password ? (
            <p className="text-xs text-destructive">
              {i18next.t("signup:Your confirmed password is inconsistent with the password!")}
            </p>
          ) : null}
        </div>
      );
    case "Email":
    case "Email or Phone":
    case "Phone or Email":
      return (
        <React.Fragment key={item.name}>
          <div className="space-y-2">
            <Label htmlFor="email">{i18next.t("general:Email")}</Label>
            <Input
              id="email"
              type="email"
              required={required}
              value={values.email ?? ""}
              onChange={(e) => set("email", e.target.value)}
            />
          </div>
          {item.rule !== "No verification" ? (
            <div className="space-y-2">
              <Label>{i18next.t("code:Email code")}</Label>
              <SendCodeInput
                value={values.emailCode ?? ""}
                onChange={(v) => set("emailCode", v)}
                method="signup"
                destType="email"
                dest={values.email ?? ""}
                application={application}
                applicationId={Setting.getApplicationName(application)}
              />
            </div>
          ) : null}
        </React.Fragment>
      );
    case "Phone":
      return (
        <React.Fragment key={item.name}>
          <div className="space-y-2">
            <Label htmlFor="phone">{i18next.t("general:Phone")}</Label>
            <div className="flex gap-2">
              <div className="w-32 shrink-0">
                <SearchableSelect
                  value={values.countryCode ?? application.organizationObj?.countryCodes?.[0] ?? ""}
                  onChange={(v) => set("countryCode", v)}
                  options={Setting.getCountryCodeData(application.organizationObj?.countryCodes).map((country: any) => ({
                    value: country.code,
                    label: `+${country.phone}`,
                    keywords: `${country.name} ${country.code} ${country.phone}`,
                  }))}
                />
              </div>
              <Input
                id="phone"
                required={required}
                value={values.phone ?? ""}
                onChange={(e) => set("phone", e.target.value)}
              />
            </div>
          </div>
          {item.rule !== "No verification" ? (
            <div className="space-y-2">
              <Label>{i18next.t("code:Phone code")}</Label>
              <SendCodeInput
                value={values.phoneCode ?? ""}
                onChange={(v) => set("phoneCode", v)}
                method="signup"
                destType="phone"
                dest={values.phone ?? ""}
                countryCode={values.countryCode ?? ""}
                application={application}
                applicationId={Setting.getApplicationName(application)}
              />
            </div>
          ) : null}
        </React.Fragment>
      );
    case "Country/Region":
      return (
        <div key={item.name} className="space-y-2">
          <Label>{i18next.t("user:Country/Region")}</Label>
          <Input value={values.region ?? ""} onChange={(e) => set("region", e.target.value)} />
        </div>
      );
    case "Agreement":
      return (
        <div key={item.name} className="flex items-start gap-2">
          <Checkbox id="agreement" checked={agreed} onCheckedChange={(v) => setAgreed(v === true)} />
          <Label htmlFor="agreement" className="text-sm font-normal leading-5">
            {i18next.t("signup:Accept")}{" "}
            {item.placeholder ? (
              <a href={item.placeholder} target="_blank" rel="noreferrer" className="underline">
                {i18next.t("signup:Terms of Use")}
              </a>
            ) : (
              i18next.t("signup:Terms of Use")
            )}
          </Label>
        </div>
      );
    case "Providers":
      return <ProviderButtons key={item.name} application={application} method="signup" />;
    case "Signup button":
      return (
        <Button key={item.name} type="submit" className="w-full" loading={loading}>
          {i18next.t("account:Sign Up")}
        </Button>
      );
    case "ID":
    case "Languages":
      return null;
    default: {
      const labelKey = SIMPLE_TEXT_ITEMS[item.name];
      const field = item.name.replace(/\s+/g, "").replace(/^./, (c: string) => c.toLowerCase());
      return (
        <div key={item.name} className="space-y-2">
          <Label htmlFor={field}>{labelKey ? i18next.t(labelKey) : item.label || item.name}</Label>
          <Input
            id={field}
            required={required}
            value={values[field] ?? ""}
            onChange={(e) => set(field, e.target.value)}
          />
        </div>
      );
    }
    }
  };

  return (
    <AuthLayout application={application} wide>
      <form className="space-y-4" onSubmit={submit}>
        <h1 className="text-center text-xl font-semibold">
          {i18next.t("account:Sign Up")} {application.displayName ? `- ${application.displayName}` : ""}
        </h1>
        {items.map(renderItem)}
        {!items.some((item: any) => item.name === "Signup button") ? (
          <Button type="submit" className="w-full" loading={loading}>
            {i18next.t("account:Sign Up")}
          </Button>
        ) : null}
        {captchaProvider ? (
          <CaptchaModal
            owner={captchaProvider.owner}
            name={captchaProvider.name}
            visible={captchaVisible}
            isCurrentProvider
            onOk={(captchaType, captchaToken, clientSecret) => {
              setCaptchaVisible(false);
              submitSignup({...pendingValues, captchaType, captchaToken, clientSecret});
            }}
            onCancel={() => setCaptchaVisible(false)}
          />
        ) : null}
        <p className="text-center text-sm text-muted-foreground">
          {i18next.t("signup:Have account?")}{" "}
          <Link to={`/login/${application.organization}`} className="text-foreground underline-offset-4 hover:underline">
            {i18next.t("signup:sign in now")}
          </Link>
        </p>
      </form>
    </AuthLayout>
  );
}
