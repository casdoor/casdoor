import * as React from "react";
import i18next from "i18next";
import {useNavigate, useParams} from "react-router-dom";
import {Alert, AlertDescription} from "@/components/ui/alert";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Label} from "@/components/ui/label";
import {Loading} from "@/components/common/Loading";
import {AuthLayout} from "@/components/auth/AuthLayout";
import {SendCodeInput} from "@/components/auth/SendCodeInput";
import {authConfig} from "@/auth/Auth";
import * as ApplicationBackend from "@/backend/ApplicationBackend";
import * as AuthBackend from "@/backend/AuthBackend";
import * as UserBackend from "@/backend/UserBackend";
import * as Setting from "@/lib/setting";

type Step = "account" | "verify" | "password";

/** Password recovery: find the account, verify a code, then set a new password. */
export default function ForgetPage() {
  const params = useParams();
  const navigate = useNavigate();
  const applicationName = params.applicationName ?? authConfig.appName;

  const [application, setApplication] = React.useState<any>(undefined);
  const [step, setStep] = React.useState<Step>("account");
  const [username, setUsername] = React.useState("");
  const [dest, setDest] = React.useState({email: "", phone: "", countryCode: ""});
  const [method, setMethod] = React.useState<"email" | "phone">("email");
  const [code, setCode] = React.useState("");
  const [password, setPassword] = React.useState("");
  const [confirm, setConfirm] = React.useState("");
  const [loading, setLoading] = React.useState(false);

  React.useEffect(() => {
    ApplicationBackend.getApplication("admin", applicationName)
      .then((res: any) => {
        setApplication(res.status === "ok" ? res.data : null);
      })
      .catch(() => setApplication(null));
  }, [applicationName]);

  if (application === undefined) {
    return <Loading className="min-h-screen" />;
  }

  if (application === null) {
    return (
      <AuthLayout>
        <Alert variant="destructive">
          <AlertDescription>{i18next.t("forget:Unknown forget type")}</AlertDescription>
        </Alert>
      </AuthLayout>
    );
  }

  const findAccount = () => {
    setLoading(true);
    AuthBackend.getEmailAndPhone(application.organization, username)
      .then((res: any) => {
        if (res.status === "ok") {
          setDest({
            email: res.data?.email ?? "",
            phone: res.data?.phone ?? "",
            countryCode: res.data?.countryCode ?? "",
          });
          setMethod(res.data?.email ? "email" : "phone");
          setStep("verify");
        } else {
          Setting.showMessage("error", res.msg);
        }
      })
      .finally(() => setLoading(false));
  };

  const verifyCode = () => {
    setLoading(true);
    UserBackend.verifyCode({
      application: application.name,
      organization: application.organization,
      username,
      name: username,
      code,
      type: "login",
      dest: method === "email" ? dest.email : dest.phone,
      countryCode: dest.countryCode,
    })
      .then((res: any) => {
        if (res.status === "ok") {
          setStep("password");
        } else {
          Setting.showMessage("error", res.msg);
        }
      })
      .finally(() => setLoading(false));
  };

  const resetPassword = () => {
    if (password !== confirm) {
      Setting.showMessage("error", i18next.t("signup:Your confirmed password is inconsistent with the password!"));
      return;
    }
    setLoading(true);
    UserBackend.setPassword(application.organization, username, "", password, code)
      .then((res: any) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("user:Password set successfully"));
          navigate(`/login/${application.organization}`);
        } else {
          Setting.showMessage("error", res.msg);
        }
      })
      .finally(() => setLoading(false));
  };

  return (
    <AuthLayout application={application}>
      <div className="space-y-5">
        <h1 className="text-center text-xl font-semibold">{i18next.t("forget:Reset password")}</h1>

        {step === "account" ? (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="account">{i18next.t("cert:Account")}</Label>
              <Input
                id="account"
                autoFocus
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && findAccount()}
              />
            </div>
            <Button className="w-full" loading={loading} onClick={findAccount}>
              {i18next.t("forget:Next Step")}
            </Button>
          </div>
        ) : step === "verify" ? (
          <div className="space-y-4">
            <div className="flex gap-2">
              {dest.email ? (
                <Button
                  variant={method === "email" ? "default" : "outline"}
                  className="flex-1"
                  onClick={() => setMethod("email")}
                >
                  {i18next.t("general:Email")}
                </Button>
              ) : null}
              {dest.phone ? (
                <Button
                  variant={method === "phone" ? "default" : "outline"}
                  className="flex-1"
                  onClick={() => setMethod("phone")}
                >
                  {i18next.t("general:Phone")}
                </Button>
              ) : null}
            </div>
            <p className="text-sm text-muted-foreground">{method === "email" ? dest.email : dest.phone}</p>
            <SendCodeInput
              value={code}
              onChange={setCode}
              method="forget"
              destType={method}
              dest={method === "email" ? dest.email : dest.phone}
              countryCode={dest.countryCode}
              application={application}
              applicationId={Setting.getApplicationName(application)}
              checkUser={username}
            />
            <Button className="w-full" loading={loading} onClick={verifyCode}>
              {i18next.t("forget:Next Step")}
            </Button>
          </div>
        ) : (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="newPassword">{i18next.t("user:New Password")}</Label>
              <Input
                id="newPassword"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="confirmPassword">{i18next.t("user:Re-enter New")}</Label>
              <Input
                id="confirmPassword"
                type="password"
                value={confirm}
                onChange={(e) => setConfirm(e.target.value)}
              />
            </div>
            <Button className="w-full" loading={loading} onClick={resetPassword}>
              {i18next.t("forget:Change Password")}
            </Button>
          </div>
        )}
      </div>
    </AuthLayout>
  );
}
