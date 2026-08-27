import * as React from "react";
import i18next from "i18next";
import {CheckCircle2} from "lucide-react";
import {useLocation, useParams} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {Loading} from "@/components/common/Loading";
import {AuthLayout} from "@/components/auth/AuthLayout";
import {authConfig} from "@/auth/Auth";
import * as ApplicationBackend from "@/backend/ApplicationBackend";
import * as Setting from "@/lib/setting";

export default function ResultPage() {
  const params = useParams();
  const location = useLocation();
  const [application, setApplication] = React.useState<any>(undefined);
  const username = (location.state as any)?.username ?? "";

  const applicationName = params.applicationName ?? authConfig.appName;

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

  const signinUrl = application ? Setting.getLoginLink(application) : "/login";

  return (
    <AuthLayout application={application}>
      <div className="space-y-4 text-center">
        <CheckCircle2 className="mx-auto h-12 w-12 text-success" />
        <h1 className="text-xl font-semibold">{i18next.t("signup:Your account has been created!")}</h1>
        {username ? <p className="text-sm text-muted-foreground">{username}</p> : null}
        <Button className="w-full" onClick={() => Setting.goToLink(signinUrl)}>
          {i18next.t("login:Sign In")}
        </Button>
      </div>
    </AuthLayout>
  );
}
