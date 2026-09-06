import * as React from "react";
import i18next from "i18next";
import {Link, useNavigate, useParams, useSearchParams} from "react-router-dom";
import {Alert, AlertDescription} from "@/components/ui/alert";
import {Button} from "@/components/ui/button";
import {Checkbox} from "@/components/ui/checkbox";
import {Input} from "@/components/ui/input";
import {Label} from "@/components/ui/label";
import {CountryCodeSelect} from "@/components/common/CountryCodeSelect";
import {Loading} from "@/components/common/Loading";
import {MultiSelect} from "@/components/common/MultiSelect";
import {PasswordInput} from "@/components/common/PasswordInput";
import {SearchableSelect} from "@/components/common/SearchableSelect";
import {RegionSelect} from "@/components/common/RegionSelect";
import {CustomHtml, CustomStyle} from "@/components/common/CustomHtml";
import {AuthDivider, AuthLayout} from "@/components/auth/AuthLayout";
import {AgreementCheckbox, getAgreementDefaultValue} from "@/components/auth/AgreementModal";
import {ProviderButtons, getVisibleProviders} from "@/components/auth/ProviderButtons";
import {SendCodeInput} from "@/components/auth/SendCodeInput";
import {CaptchaModal} from "@/components/common/CaptchaModal";
import {getCaptchaProvider} from "@/lib/captcha";
import {PasswordRequirements} from "@/lib/password-checker";
import {useAccount} from "@/hooks/use-account";
import {authConfig} from "@/auth/Auth";
import * as Util from "@/auth/Util";
import * as ApplicationBackend from "@/backend/ApplicationBackend";
import * as InvitationBackend from "@/backend/InvitationBackend";
import * as AuthBackend from "@/backend/AuthBackend";
import {getSignupItemField, validateSignupItems} from "@/lib/signup-validation";
import * as Setting from "@/lib/setting";
import {cn} from "@/lib/utils";

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

/** Where the antd class name differs from the field, so `customCss` keeps matching. */
const FIELD_CLASS_NAMES: Record<string, string> = {
  firstName: "first-name",
  lastName: "last-name",
  idCard: "idcard",
  invitationCode: "invitation-code",
};

function fieldClass(field: string): string {
  return `signup-${FIELD_CLASS_NAMES[field] ?? field}`;
}

/** the application editor's live preview hands the form's own object over */
export default function SignupPage({application: applicationProp}: {application?: any} = {}) {
  const params = useParams();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const {reload} = useAccount();

  const [application, setApplication] = React.useState<any>(undefined);
  const [msg, setMsg] = React.useState<string | null>(null);
  const [values, setValues] = React.useState<Record<string, any>>({});
  const [agreed, setAgreed] = React.useState(false);
  const [errors, setErrors] = React.useState<Record<string, string>>({});
  const [loading, setLoading] = React.useState(false);
  const [captchaVisible, setCaptchaVisible] = React.useState(false);
  const [pendingValues, setPendingValues] = React.useState<any>(null);
  const [emailOrPhoneMode, setEmailOrPhoneMode] = React.useState("");
  const [invitation, setInvitation] = React.useState<any>(undefined);
  const [passwordFocused, setPasswordFocused] = React.useState(false);
  const [userLang, setUserLang] = React.useState("");

  const applicationName = params.applicationName ?? authConfig.appName;

  React.useEffect(() => {
    if (applicationProp) {
      setApplication(applicationProp);
      setAgreed(getAgreementDefaultValue(applicationProp));
      setValues((prev) => ({
        ...prev,
        application: applicationProp.name,
        organization: applicationProp.organization,
        countryCode: prev.countryCode ?? applicationProp.organizationObj?.countryCodes?.[0] ?? "",
      }));
      return;
    }

    const oAuthParams = Util.getOAuthGetParameters();
    if (oAuthParams) {
      // the OAuth params live on the signin path, remember it to get back there after the signup
      const signinUrl = window.location.pathname.replace("/signup/oauth/authorize", "/login/oauth/authorize");
      sessionStorage.setItem("signinUrl", signinUrl + window.location.search);
    }
    const load = oAuthParams
      ? AuthBackend.getApplicationLogin(oAuthParams)
      : ApplicationBackend.getApplication("admin", applicationName);
    load
      .then((res: any) => {
        if (res.status === "ok" && res.data) {
          setApplication(res.data);
          setAgreed(getAgreementDefaultValue(res.data));
          const invitationCode = searchParams.get("invitationCode") ?? "";
          setValues((prev) => ({
            ...prev,
            application: res.data.name,
            organization: res.data.organization,
            invitationCode,
            // the antd CountryCodeSelect seeded the form with it, so the phone
            // code request and the signup payload carry it even if untouched
            countryCode: prev.countryCode ?? res.data.organizationObj?.countryCodes?.[0] ?? "",
          }));
          // an invitation can pin the username, email or phone the account must be created with
          if (invitationCode !== "") {
            InvitationBackend.getInvitationCodeInfo(invitationCode, `admin/${res.data.name}`).then((infoRes: any) => {
              if (infoRes.status === "error") {
                Setting.showMessage("error", infoRes.msg);
                return;
              }
              setInvitation(infoRes.data);
              setValues((prev) => ({
                ...prev,
                ...(infoRes.data?.username ? {username: infoRes.data.username} : {}),
                ...(infoRes.data?.email ? {email: infoRes.data.email} : {}),
                ...(infoRes.data?.phone ? {phone: infoRes.data.phone} : {}),
              }));
            });
          }
        } else {
          setApplication(null);
          setMsg(res.msg || i18next.t("general:Unknown application name"));
        }
      })
      .catch((error) => {
        setApplication(null);
        setMsg(`${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [applicationName, applicationProp]);

  const signupItems = (application?.signupItems ?? []) as any[];
  const languagesItem = signupItems.find((item: any) => item.name === "Languages");
  const languages = application?.organizationObj?.languages as string[] | undefined;
  // an organization that offers a single language forces it rather than offering a choice
  const forcedLanguage =
    (!languagesItem || languagesItem.visible) && languages && languages.length <= 1
      ? (languages.length === 1 ? languages[0] : "en")
      : "";

  React.useEffect(() => {
    if (forcedLanguage !== "" && Setting.getLanguage() !== forcedLanguage) {
      Setting.setLanguage(forcedLanguage);
    }
  }, [forcedLanguage]);

  const set = (key: string, value: any) => {
    setValues((prev) => ({...prev, [key]: value}));
    // antd re-validates on change; clearing the message as the user types matches that
    setErrors((prev) => (prev[key] ? {...prev, [key]: ""} : prev));
  };

  /** the inline message under a field, as the antd Form.Item shows it */
  const fieldError = (field: string) =>
    errors[field] ? <p className="text-xs text-destructive">{errors[field]}</p> : null;

  const renderLabel = (text: React.ReactNode, required?: boolean, htmlFor?: string) => (
    <Label htmlFor={htmlFor}>
      {text}
      {required ? <span className="ml-0.5 text-destructive">*</span> : null}
    </Label>
  );

  if (application === undefined) {
    return <Loading className="min-h-screen" />;
  }

  if (application === null) {
    return (
      <AuthLayout preview={!!applicationProp}>
        <Alert variant="destructive">
          <AlertDescription>{msg ?? i18next.t("application:Failed to sign in")}</AlertDescription>
        </Alert>
      </AuthLayout>
    );
  }

  const signinLink = Setting.getStoredSigninUrl() || Setting.getLoginLink(application) || "/login";

  if (!application.enableSignUp) {
    return (
      <AuthLayout preview={!!applicationProp} application={application}>
        <div className="space-y-4">
          <Alert variant="warning">
            <AlertDescription>{i18next.t("application:The application does not allow to sign up new account")}</AlertDescription>
          </Alert>
          <Button className="w-full" onClick={() => Setting.goToLink(signinLink)}>
            {i18next.t("login:Sign In")}
          </Button>
        </div>
      </AuthLayout>
    );
  }

  const captchaProvider = getCaptchaProvider(application);
  const items = signupItems.filter((item: any) => item.visible);
  // "First, last" splits the display name into two fields, which then replace the
  // separate First name / Last name items; Chinese names are not split
  const splitDisplayName =
    signupItems.find((item: any) => item.name === "Display name")?.rule === "First, last" &&
    Setting.getLanguage() !== "zh";

  /** The item an "Email or Phone" / "Phone or Email" choice currently stands for. */
  const emailOrPhoneItem = (item: any) => {
    const mode = emailOrPhoneMode || (item.name === "Email or Phone" ? "Email" : "Phone");
    return {...item, name: mode};
  };

  const agreementItem = signupItems.find((item: any) => item.name === "Agreement" && item.visible);

  /** the agreement gates the provider buttons too */
  const checkAgreement = () => {
    if (agreementItem?.required && !agreed) {
      Setting.showMessage("error", i18next.t("signup:Please accept the agreement!"));
      return false;
    }
    return true;
  };

  const submit = (e: React.FormEvent) => {
    e.preventDefault();

    // the antd form runs these as Form rules; without them an organization's
    // signup-item regex and the phone / email / password checks are ignored
    const validated = items.map((item: any) =>
      item.name === "Email or Phone" || item.name === "Phone or Email" ? emailOrPhoneItem(item) : item,
    );
    const found = validateSignupItems(validated, {...values, agreement: agreed}, application);
    setErrors(found);
    const firstError = Object.values(found)[0];
    if (firstError) {
      Setting.showMessage("error", firstError);
      return;
    }

    const payload: Record<string, any> = {
      ...values,
      application: application.name,
      organization: application.organization,
      plan: searchParams.get("plan"),
      pricing: searchParams.get("pricing"),
      agreement: agreed,
      language: userLang,
    };
    // a "Multiple Choices" item edits an array, the backend field is a string
    Object.keys(payload).forEach((key) => {
      if (Array.isArray(payload[key])) {
        payload[key] = payload[key].join(", ");
      }
    });

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

  const getResultPath = (payload: Record<string, any>, username: string) => {
    if (payload.plan && payload.pricing) {
      // the prompt page needs the user to be signed in, so a paid signup goes to buy-plan
      return `/buy-plan/${application.organization}/${payload.pricing}?user=${username}&plan=${payload.plan}`;
    }
    if (authConfig.appName === application.name) {
      return "/result";
    }
    if (Setting.hasPromptPage(application)) {
      return `/prompt/${application.name}?oauth=${Util.getOAuthGetParameters() !== null}`;
    }
    return `/result/${application.name}`;
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
        // the id of the new user comes back from signup(); a phone-only signup has no username
        let username = payload.username;
        if (typeof res.data === "string") {
          username = res.data.split("/")[1];
        }
        const path = getResultPath(payload, username);
        // the prompt page renders the signed-in account, so it has to be loaded first
        if (Setting.hasPromptPage(application) && (!payload.plan || !payload.pricing)) {
          reload().then(() => navigate(path, {state: {username}}));
          return;
        }
        navigate(path, {state: {username}});
      })
      .finally(() => setLoading(false));
  };

  const renderEmail = (item: any) => (
    <React.Fragment key={`${item.name}-email`}>
      <div className="signup-email space-y-2">
        {renderLabel(item.label || i18next.t("general:Email"), item.required, "email")}
        <Input
          id="email"
          className="signup-email-input"
          // not type="email": the browser's own bubble would pre-empt the
          // translated message and skip the signup item's regex
          autoComplete="email"
          required={item.required}
          disabled={!!invitation?.email}
          placeholder={item.placeholder}
          value={values.email ?? ""}
          onChange={(e) => set("email", e.target.value)}
        />
        {fieldError("email")}
      </div>
      {item.rule !== "No verification" ? (
        <div className="signup-email-code space-y-2">
          {renderLabel(i18next.t("code:Email code"), item.required)}
          <SendCodeInput
            className="signup-email-code-input"
            value={values.emailCode ?? ""}
            onChange={(v) => set("emailCode", v)}
            method="signup"
            destType="email"
            dest={values.email ?? ""}
            disabled={!Setting.isValidEmail(values.email ?? "")}
            application={application}
            applicationId={Setting.getApplicationName(application)}
          />
        </div>
      ) : null}
    </React.Fragment>
  );

  const renderPhone = (item: any) => (
    <React.Fragment key={`${item.name}-phone`}>
      <div className="signup-phone space-y-2">
        {renderLabel(item.label || i18next.t("general:Phone"), item.required, "phone")}
        <div className="flex gap-2">
          <div className="w-28 max-w-[50%] shrink-0">
            <CountryCodeSelect
              className="px-2"
              value={values.countryCode ?? ""}
              onChange={(v) => set("countryCode", v)}
              countryCodes={application.organizationObj?.countryCodes}
            />
          </div>
          <Input
            id="phone"
            className="signup-phone-input min-w-0 flex-1"
            autoComplete="tel"
            required={item.required}
            disabled={!!invitation?.phone}
            placeholder={item.placeholder}
            value={values.phone ?? ""}
            onChange={(e) => set("phone", e.target.value)}
          />
        </div>
        {fieldError("phone")}
      </div>
      {item.rule !== "No verification" ? (
        <div className="phone-code space-y-2">
          {renderLabel(i18next.t("code:Phone code"), item.required)}
          <SendCodeInput
            className="signup-phone-code-input"
            value={values.phoneCode ?? ""}
            onChange={(v) => set("phoneCode", v)}
            method="signup"
            destType="phone"
            dest={values.phone ?? ""}
            countryCode={values.countryCode ?? ""}
            disabled={!Setting.isValidPhone(values.phone ?? "", values.countryCode ?? "")}
            application={application}
            applicationId={Setting.getApplicationName(application)}
          />
        </div>
      ) : null}
    </React.Fragment>
  );

  /** A plain text field, used by the items that only differ in their label. */
  const renderTextItem = (item: any, field: string, defaultLabel: string, extra?: Record<string, any>) => (
    <div key={`${item.name}-${field}`} className={`${fieldClass(field)} space-y-2`}>
      {renderLabel(item.label || defaultLabel, item.required, field)}
      <Input
        id={field}
        className={`${fieldClass(field)}-input`}
        required={item.required}
        placeholder={item.placeholder}
        value={values[field] ?? ""}
        onChange={(e) => set(field, e.target.value)}
        {...extra}
      />
      {fieldError(field)}
    </div>
  );

  const renderPassword = (item: any, field: "password" | "confirm", defaultLabel: string) => (
    <div key={`${item.name}-${field}`} className={`${fieldClass(field)} space-y-2`}>
      {renderLabel(item.label || defaultLabel, item.required, field)}
      <PasswordInput
        id={field}
        className={`${fieldClass(field)}-input`}
        autoComplete="new-password"
        required={item.required}
        placeholder={item.placeholder}
        value={values[field] ?? ""}
        onChange={(e) => set(field, e.target.value)}
        onFocus={field === "password" ? () => setPasswordFocused(true) : undefined}
      />
      {field === "password" && passwordFocused ? (
        <PasswordRequirements
          options={application.organizationObj?.passwordOptions}
          password={values.password ?? ""}
        />
      ) : null}
      {field === "confirm" && values.confirm && values.confirm !== values.password ? (
        <p className="text-xs text-destructive">
          {i18next.t("signup:Your confirmed password is inconsistent with the password!")}
        </p>
      ) : (
        fieldError(field)
      )}
    </div>
  );

  const renderItem = (item: any) => {
    if (Setting.isCustomFormItem(item)) {
      // a "Text N" item is raw HTML, kept in the label by the application editor
      return <CustomHtml key={item.name} html={item.label} />;
    }

    switch (item.name) {
    case "Username":
      return renderTextItem(item, "username", i18next.t("signup:Username"), {
        autoComplete: "username",
        disabled: !!invitation?.username,
      });
    case "Display name":
      if (splitDisplayName) {
        return (
          <React.Fragment key={item.name}>
            {renderTextItem({...item, label: ""}, "firstName", i18next.t("general:First name"))}
            {renderTextItem({...item, label: ""}, "lastName", i18next.t("general:Last name"))}
          </React.Fragment>
        );
      }
      return renderTextItem(
        item,
        "name",
        item.rule === "Real name" || item.rule === "First, last"
          ? i18next.t("application:Real name")
          : i18next.t("general:Display name"),
      );
    case "First name":
      return splitDisplayName ? null : renderTextItem(item, "firstName", i18next.t("general:First name"));
    case "Last name":
      return splitDisplayName ? null : renderTextItem(item, "lastName", i18next.t("general:Last name"));
    case "Password":
      return renderPassword(item, "password", i18next.t("general:Password"));
    case "Confirm password":
      return renderPassword(item, "confirm", i18next.t("general:Confirm"));
    case "Email":
      return renderEmail(item);
    case "Phone":
      return renderPhone(item);
    case "Email or Phone":
    case "Phone or Email": {
      const mode = emailOrPhoneMode || (item.name === "Email or Phone" ? "Email" : "Phone");
      const choices = item.name === "Email or Phone" ? ["Email", "Phone"] : ["Phone", "Email"];
      return (
        <React.Fragment key={item.name}>
          <div className="signup-email-or-phone grid grid-cols-2 gap-1 rounded-lg bg-muted p-1">
            {choices.map((choice) => (
              <button
                key={choice}
                type="button"
                onClick={() => setEmailOrPhoneMode(choice)}
                className={cn(
                  "rounded-md px-3 py-1.5 text-sm font-medium transition-colors",
                  mode === choice
                    ? "bg-background text-foreground shadow-sm"
                    : "text-muted-foreground hover:text-foreground",
                )}
              >
                {i18next.t(choice === "Email" ? "general:Email" : "general:Phone")}
              </button>
            ))}
          </div>
          {mode === "Email" ? renderEmail(item) : renderPhone(item)}
        </React.Fragment>
      );
    }
    case "Country/Region":
      return (
        <div key={item.name} className="signup-country-region space-y-2">
          {renderLabel(item.label || i18next.t("user:Country/Region"), item.required)}
          <RegionSelect
            className="signup-region-select"
            value={values.region ?? ""}
            onChange={(v) => set("region", v)}
          />
          {fieldError("region")}
        </div>
      );
    case "Tag": {
      // the item's own options win over the organization's tag list
      const tags = (item.options?.length > 0 ? item.options : application.tags ?? []) as string[];
      return (
        <div key={item.name} className="signup-tag space-y-2">
          {renderLabel(item.label || i18next.t("general:Tag"), item.required)}
          <SearchableSelect
            className="signup-tag-select"
            value={values.tag ?? ""}
            onChange={(v) => set("tag", v)}
            placeholder={item.placeholder || i18next.t("signup:Please select your tag!")}
            options={tags.map((tag) => ({value: tag, label: tag}))}
          />
          {fieldError("tag")}
        </div>
      );
    }
    case "Invitation code":
      return renderTextItem(item, "invitationCode", i18next.t("application:Invitation code"), {
        disabled: invitation !== undefined,
      });
    case "Agreement":
      // the application's own terms document opens in a dialog, as in the antd page
      if (application.termsOfUse) {
        return (
          <div key={item.name} className="login-agreement">
            <AgreementCheckbox application={application} checked={agreed} onChange={setAgreed} />
          </div>
        );
      }
      return (
        <div key={item.name} className="login-agreement flex items-start gap-2">
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
      if (getVisibleProviders(application, "signup").length === 0) {
        return null;
      }
      return (
        <div key={item.name} className="space-y-4">
          <AuthDivider />
          <ProviderButtons
            application={application}
            method="signup"
            rule={Setting.getProvidersRule(application, item)}
            onBeforeClick={checkAgreement}
          />
        </div>
      );
    case "Signup button":
      return (
        <Button key={item.name} type="submit" className="signup-button w-full" loading={loading}>
          {item.label || i18next.t("account:Sign Up")}
        </Button>
      );
    case "ID":
    case "Languages":
      return null;
    default: {
      const labelKey = SIMPLE_TEXT_ITEMS[item.name];
      const field = getSignupItemField(item.name);
      const label = item.label || (labelKey ? i18next.t(labelKey) : item.name);
      const options = (item.options ?? []).map((option: string) => ({value: option, label: option}));
      if (!item.type || item.type === "Input") {
        return renderTextItem(item, field, label);
      }
      return (
        <div key={item.name} className={`${fieldClass(field)} space-y-2`}>
          {renderLabel(label, item.required, field)}
          {/* an item can be a choice list rather than a text box, as in the antd form */}
          {item.type === "Single Choice" ? (
            <SearchableSelect
              value={values[field] ?? ""}
              onChange={(v) => set(field, v)}
              placeholder={item.placeholder}
              options={options}
            />
          ) : (
            <MultiSelect
              value={values[field] ?? []}
              onChange={(v) => set(field, v)}
              placeholder={item.placeholder}
              options={options}
            />
          )}
          {fieldError(field)}
        </div>
      );
    }
    }
  };

  // The whole page can be replaced by the application's own markup.
  if (application.signupHtml) {
    return <CustomHtml html={application.signupHtml} />;
  }

  // an application that lists no signup button at all still gets the default one
  const hasSignupButtonItem = signupItems.some((item: any) => item.name === "Signup button");

  return (
    <AuthLayout
      preview={!!applicationProp}
      application={application}
      wide
      hideLanguages={(languagesItem && !languagesItem.visible) || forcedLanguage !== ""}
      onLanguageChange={(key) => {
        setUserLang(key);
        Setting.setSigninLanguage(key);
      }}
      footer={
        <>
          {i18next.t("signup:Have account?")}{" "}
          {signinLink.startsWith("/") ? (
            <Link to={signinLink} className="signup-link font-medium text-foreground underline-offset-4 hover:underline">
              {i18next.t("signup:sign in now")}
            </Link>
          ) : (
            <a href={signinLink} className="signup-link font-medium text-foreground underline-offset-4 hover:underline">
              {i18next.t("signup:sign in now")}
            </a>
          )}
        </>
      }
    >
      <form className="space-y-4" onSubmit={submit}>
        {/* each item styles itself; a "Text N" item holds the HTML, not the CSS */}
        {signupItems.map((item: any) => (
          <CustomStyle key={`css-${item.name}`} css={item.customCss} />
        ))}
        {items.map(renderItem)}
        {hasSignupButtonItem ? null : (
          <Button type="submit" className="signup-button w-full" loading={loading}>
            {i18next.t("account:Sign Up")}
          </Button>
        )}
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
      </form>
    </AuthLayout>
  );
}
