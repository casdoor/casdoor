import * as React from "react";
import i18next from "i18next";
import {Lock} from "lucide-react";
import {useParams, useSearchParams} from "react-router-dom";
import {Alert, AlertDescription} from "@/components/ui/alert";
import {Button} from "@/components/ui/button";
import {Loading} from "@/components/common/Loading";
import {AuthLayout} from "@/components/auth/AuthLayout";
import * as Util from "@/auth/Util";
import * as ApplicationBackend from "@/backend/ApplicationBackend";
import * as ConsentBackend from "@/backend/ConsentBackend";
import * as Setting from "@/lib/setting";

interface ScopeDescription {
  scope: string;
  displayName: string;
  description?: string;
}

/** OAuth consent screen: shows the requested scopes and grants or denies the code. */
export default function ConsentPage() {
  const params = useParams();
  const [searchParams] = useSearchParams();
  const applicationName = params.applicationName ?? searchParams.get("application") ?? "";

  const [application, setApplication] = React.useState<any>(undefined);
  const [granting, setGranting] = React.useState(false);
  const oAuthParams = React.useMemo(() => Util.getOAuthGetParameters(), []);

  React.useEffect(() => {
    if (!applicationName) {
      setApplication(null);
      return;
    }
    ApplicationBackend.getApplication("admin", applicationName)
      .then((res: any) => {
        if (res.status === "ok") {
          setApplication(res.data);
        } else {
          setApplication(null);
          Setting.showMessage("error", res.msg);
        }
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
          <AlertDescription>{i18next.t("general:Invalid application")}</AlertDescription>
        </Alert>
      </AuthLayout>
    );
  }

  // Scopes the client asked for, described by the application's customScopes.
  const customScopesMap: Record<string, any> = {};
  (application.customScopes ?? []).forEach((item: any) => {
    if (item?.scope) {
      customScopesMap[item.scope] = item;
    }
  });
  const scopeDescriptions: ScopeDescription[] = (oAuthParams?.scope ?? "")
    .split(" ")
    .map((s: string) => s.trim())
    .filter(Boolean)
    .map((scope: string) => {
      const item = customScopesMap[scope];
      return item
        ? {...item, displayName: item.displayName || item.scope}
        : {
          scope,
          displayName: scope,
          description: i18next.t("consent:This scope is not defined in the application"),
        };
    });

  const grant = () => {
    setGranting(true);
    ConsentBackend.grantConsent(
      {
        owner: application.owner,
        application: `${application.owner}/${application.name}`,
        grantedScopes: scopeDescriptions.map((s) => s.scope),
      },
      oAuthParams,
    )
      .then((res: any) => {
        if (res.status === "ok") {
          const concatChar = oAuthParams?.redirectUri?.includes("?") ? "&" : "?";
          Setting.goToLink(`${oAuthParams.redirectUri}${concatChar}code=${res.data}&state=${oAuthParams.state}`);
        } else {
          Setting.showMessage("error", res.msg);
          setGranting(false);
        }
      })
      .catch((error) => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
        setGranting(false);
      });
  };

  const deny = () => {
    const concatChar = oAuthParams?.redirectUri?.includes("?") ? "&" : "?";
    Setting.goToLink(
      `${oAuthParams.redirectUri}${concatChar}error=access_denied&error_description=User denied consent&state=${oAuthParams.state}`,
    );
  };

  return (
    <AuthLayout application={application}>
      <div className="space-y-5">
        <div className="space-y-1 text-center">
          <h1 className="text-lg font-semibold">{i18next.t("consent:Authorization Request")}</h1>
          <p className="text-sm font-medium">{application.displayName || application.name}</p>
          <p className="text-sm text-muted-foreground">{i18next.t("consent:wants to access your account")}</p>
          {application.homepageUrl ? (
            <a
              href={application.homepageUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="text-xs text-primary underline-offset-4 hover:underline"
            >
              {application.homepageUrl}
            </a>
          ) : null}
        </div>

        {scopeDescriptions.length === 0 ? (
          <Alert variant="info">
            <AlertDescription>{i18next.t("consent:This application is requesting")}</AlertDescription>
          </Alert>
        ) : (
          <ul className="divide-y rounded-lg border">
            {scopeDescriptions.map((item) => (
              <li key={item.scope} className="flex items-start gap-3 p-3">
                <Lock className="mt-0.5 h-4 w-4 shrink-0 text-muted-foreground" />
                <div className="min-w-0">
                  <div className="text-sm font-medium">{item.displayName}</div>
                  {item.description ? (
                    <p className="text-xs text-muted-foreground">{item.description}</p>
                  ) : null}
                </div>
              </li>
            ))}
          </ul>
        )}

        <div className="flex gap-2">
          <Button variant="outline" className="flex-1" onClick={deny} disabled={granting}>
            {i18next.t("permission:Deny")}
          </Button>
          <Button className="flex-1" loading={granting} onClick={grant} disabled={granting || scopeDescriptions.length === 0}>
            {i18next.t("permission:Allow")}
          </Button>
        </div>

        <p className="rounded-lg border bg-muted/40 px-4 py-3 text-center text-xs text-muted-foreground">
          {i18next.t("consent:By clicking Allow, you allow this app to use your information")}
        </p>
      </div>
    </AuthLayout>
  );
}
