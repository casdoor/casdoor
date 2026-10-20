import * as React from "react";
import i18next from "i18next";
import {useParams, useSearchParams} from "react-router-dom";
import {Alert, AlertDescription} from "@/components/ui/alert";
import {Badge} from "@/components/ui/badge";
import {Button} from "@/components/ui/button";
import {Checkbox} from "@/components/ui/checkbox";
import {Loading} from "@/components/common/Loading";
import {AuthLayout} from "@/components/auth/AuthLayout";
import * as Util from "@/auth/Util";
import {useAccount} from "@/hooks/use-account";
import * as ApplicationBackend from "@/backend/ApplicationBackend";
import * as ConsentBackend from "@/backend/ConsentBackend";
import * as Setting from "@/lib/setting";

interface ScopeDescription {
  scope: string;
  displayName: string;
  description?: string;
  /** openid and scopes the application doesn't define are always granted */
  required: boolean;
  /** not asked before, while the user has already answered for other scopes of this application */
  isNew: boolean;
  defaultChecked: boolean;
}

/** OAuth consent screen: lets the user pick which of the requested scopes to grant, or deny the code. */
export default function ConsentPage() {
  const params = useParams();
  const [searchParams] = useSearchParams();
  const applicationName = params.applicationName ?? searchParams.get("application") ?? "";

  const [application, setApplication] = React.useState<any>(undefined);
  const [granting, setGranting] = React.useState(false);
  const [checkedScopes, setCheckedScopes] = React.useState<Record<string, boolean>>({});
  const oAuthParams = React.useMemo(() => Util.getOAuthGetParameters(), []);
  const {account} = useAccount();

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

  const scopeDescriptions: ScopeDescription[] = React.useMemo(() => {
    if (!application) {
      return [];
    }

    // Scopes the client asked for, described by the application's customScopes.
    const customScopesMap: Record<string, any> = {};
    (application.customScopes ?? []).forEach((item: any) => {
      if (item?.scope) {
        customScopesMap[item.scope] = item;
      }
    });

    // What the user answered before for this application: granted scopes start checked, denied ones unchecked.
    const record = (account?.applicationScopes ?? []).find((item: any) => item?.application === `${application.owner}/${application.name}`);
    const granted: string[] = record?.grantedScopes ?? [];
    const denied: string[] = record?.deniedScopes ?? [];

    return Array.from(new Set<string>((oAuthParams?.scope ?? "").split(" ").map((s: string) => s.trim()).filter(Boolean)))
      .map((scope: string) => {
        const item = customScopesMap[scope];
        const required = scope === "openid" || !item;
        const base = item
          ? {...item, displayName: item.displayName || item.scope}
          : {
            scope,
            displayName: scope,
            description: i18next.t("consent:This scope is not defined in the application"),
          };
        return {
          ...base,
          scope,
          required,
          isNew: !!record && !required && !granted.includes(scope) && !denied.includes(scope),
          defaultChecked: required || !denied.includes(scope),
        };
      });
  }, [application, account, oAuthParams]);

  React.useEffect(() => {
    const checked: Record<string, boolean> = {};
    scopeDescriptions.forEach((item) => {
      checked[item.scope] = item.defaultChecked;
    });
    setCheckedScopes(checked);
  }, [scopeDescriptions]);

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

  const grantedScopes = scopeDescriptions.filter((item) => item.required || checkedScopes[item.scope]).map((item) => item.scope);

  const grant = () => {
    setGranting(true);
    ConsentBackend.grantConsent(
      {
        owner: application.owner,
        application: `${application.owner}/${application.name}`,
        grantedScopes,
      },
      oAuthParams,
    )
      .then((res: any) => {
        if (res.status === "ok") {
          // a hybrid flow has the access token in data2 and the ID token in data3
          const responseTypes = (oAuthParams.responseType || "code").split(" ");
          const payload: Record<string, string> = {code: res.data};
          if (responseTypes.includes("token")) {
            payload.access_token = res.data2;
            payload.token_type = "Bearer";
          }
          if (responseTypes.includes("id_token")) {
            payload.id_token = res.data3;
          }
          Util.sendOAuthResponse(oAuthParams, payload);
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
    Util.sendOAuthResponse(oAuthParams, {error: "access_denied", error_description: "User denied consent"});
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
              <li key={item.scope}>
                <label
                  htmlFor={`consent-scope-${item.scope}`}
                  className={`flex items-start gap-3 p-3 ${item.required ? "" : "cursor-pointer"}`}
                >
                  <Checkbox
                    id={`consent-scope-${item.scope}`}
                    className="mt-0.5"
                    checked={item.required || !!checkedScopes[item.scope]}
                    disabled={item.required || granting}
                    onCheckedChange={(v) => setCheckedScopes((prev) => ({...prev, [item.scope]: v === true}))}
                  />
                  <div className="min-w-0">
                    <div className="flex flex-wrap items-center gap-2 text-sm font-medium">
                      {item.displayName}
                      {item.isNew ? <Badge variant="info">{i18next.t("consent:New")}</Badge> : null}
                      {item.required ? (
                        <span className="text-xs font-normal text-muted-foreground">{i18next.t("organization:Required")}</span>
                      ) : null}
                    </div>
                    {item.description ? (
                      <p className="text-xs text-muted-foreground">{item.description}</p>
                    ) : null}
                  </div>
                </label>
              </li>
            ))}
          </ul>
        )}

        <div className="flex gap-2">
          <Button variant="outline" className="flex-1" onClick={deny} disabled={granting}>
            {i18next.t("permission:Deny")}
          </Button>
          <Button className="flex-1" loading={granting} onClick={grant} disabled={granting || grantedScopes.length === 0}>
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
