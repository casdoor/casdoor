import * as React from "react";
import i18next from "i18next";
import {ArrowLeft, ExternalLink} from "lucide-react";
import {Link, useNavigate, useParams} from "react-router-dom";
import {Badge} from "@/components/ui/badge";
import {Button} from "@/components/ui/button";
import {Card, CardContent, CardDescription, CardHeader, CardTitle} from "@/components/ui/card";
import {ConfirmButton} from "@/components/common/ConfirmButton";
import {DescriptionList, type DescriptionItem} from "@/components/common/DescriptionList";
import {Loading} from "@/components/common/Loading";
import {useAccount} from "@/hooks/use-account";
import * as ApplicationBackend from "@/backend/ApplicationBackend";
import * as IntegrationBackend from "@/backend/IntegrationBackend";
import {
  IntegrationLogo,
  MarketplaceSiteUrl,
  fillVariables,
  getIntegrationTypeLabel,
  getLocalized,
  parseBundle,
  renderInlineText,
} from "@/lib/integration";
import * as Setting from "@/lib/setting";

/** An installed integration: what it created, and the rest of the setup with this Casdoor's own values filled in. */
export default function IntegrationEditPage() {
  const navigate = useNavigate();
  const {organizationName = "", integrationName = ""} = useParams();
  const {account} = useAccount();

  const [integration, setIntegration] = React.useState<any>(undefined);
  const [application, setApplication] = React.useState<any>(null);
  const [latest, setLatest] = React.useState("");

  React.useEffect(() => {
    IntegrationBackend.getIntegration(organizationName, integrationName).then((res: any) => {
      if (res.status !== "ok") {
        Setting.showMessage("error", `${i18next.t("general:Failed to get")}: ${res.msg}`);
        setIntegration(null);
        return;
      }
      setIntegration(res.data);
      if (res.data?.type === "app" && res.data.application) {
        ApplicationBackend.getApplication("admin", res.data.application).then((appRes: any) => {
          if (appRes.status === "ok") {
            setApplication(appRes.data);
          }
        });
      }
      if (res.data) {
        IntegrationBackend.getMarketplaceIndex(organizationName).then((indexRes: any) => {
          const item = indexRes.data?.integrations?.find((i: any) => i.id === res.data.integrationId);
          setLatest(item?.latest ?? "");
        });
      }
    });
  }, [organizationName, integrationName]);

  if (integration === undefined) {
    return <Loading />;
  }
  if (integration === null) {
    return <p className="py-16 text-center text-muted-foreground">{i18next.t("general:No data")}</p>;
  }

  const bundle = parseBundle(integration.bundle);
  const files = bundle?.files ?? {};
  const guide = files["guide.json"];
  const values: Record<string, string> = {
    ...(integration.variables ?? {}),
    clientId: application?.clientId ?? "",
    clientSecret: application?.clientSecret ?? "",
  };
  const fill = (text: string) => fillVariables(text, values);

  const uninstall = async() => {
    const res: any = await IntegrationBackend.deleteIntegration(integration);
    if (res.status === "ok") {
      Setting.showMessage("success", i18next.t("integration:Uninstalled"));
      navigate("/integrations");
    } else {
      Setting.showMessage("error", `${i18next.t("integration:Failed to uninstall")}: ${res.msg}`);
    }
  };

  const items: DescriptionItem[] = [
    {label: i18next.t("general:Organization"), children: <Link className="underline-offset-4 hover:underline" to={`/organizations/${integration.owner}`}>{integration.owner}</Link>},
    {label: i18next.t("general:Name"), children: integration.name},
    {
      label: i18next.t("integration:Integration"),
      children: (
        <a className="inline-flex items-center gap-1 underline-offset-4 hover:underline" href={`${MarketplaceSiteUrl}/integrations/${integration.integrationId}/`} target="_blank" rel="noreferrer">
          {integration.integrationId}
          <ExternalLink className="size-3.5" />
        </a>
      ),
    },
    {
      label: i18next.t("system:Version"),
      children: (
        <span className="flex items-center gap-2">
          {integration.version}
          {latest && latest !== integration.version ? <Badge variant="warning">{i18next.t("integration:Update available")}: {latest}</Badge> : null}
        </span>
      ),
    },
    {
      label: i18next.t("general:Application"),
      hidden: !integration.application,
      children: <Link className="underline-offset-4 hover:underline" to={`/applications/${integration.owner}/${integration.application}`}>{integration.application}</Link>,
    },
    {
      label: i18next.t("general:Provider"),
      hidden: !integration.provider,
      children: <Link className="underline-offset-4 hover:underline" to={`/providers/${integration.owner}/${integration.provider}`}>{integration.provider}</Link>,
    },
    {label: i18next.t("provider:Client ID"), hidden: !application, children: <code className="font-mono text-xs">{application?.clientId}</code>},
    {
      label: i18next.t("application:Redirect URLs"),
      hidden: !application,
      children: (
        <div className="space-y-0.5">
          {(application?.redirectUris ?? []).map((uri: string) => <div key={uri}><code className="font-mono text-xs">{uri}</code></div>)}
        </div>
      ),
    },
    ...Object.entries(integration.variables ?? {})
      .filter(([name]) => name !== "casdoorUrl")
      .map(([name, value]) => {
        const variable = bundle?.manifest?.variables?.find((v: any) => v.name === name);
        return {key: name, label: getLocalized(variable?.label) || name, children: String(value) || "-"};
      }),
    {label: i18next.t("general:Created time"), children: Setting.getFormattedDate(integration.createdTime)},
  ];

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex min-w-0 items-center gap-3">
          <Button variant="ghost" size="icon" onClick={() => navigate("/integrations")} aria-label={i18next.t("general:Back")}>
            <ArrowLeft className="size-4" />
          </Button>
          <IntegrationLogo logo={integration.logo} name={integration.displayName} />
          <div className="min-w-0">
            <h1 className="truncate text-2xl font-semibold tracking-tight">{integration.displayName}</h1>
            <div className="mt-0.5 flex gap-1">
              <Badge variant="secondary">{getIntegrationTypeLabel(integration.type)}</Badge>
              <Badge variant="outline">{integration.version}</Badge>
            </div>
          </div>
        </div>
        {Setting.isLocalAdminUser(account) ? (
          <ConfirmButton
            variant="outline"
            title={i18next.t("integration:Uninstall")}
            description={getUninstallDescription(integration)}
            confirmText={i18next.t("integration:Uninstall")}
            onConfirm={uninstall}
          >
            {i18next.t("integration:Uninstall")}
          </ConfirmButton>
        ) : null}
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">{i18next.t("integration:Installed in Casdoor")}</CardTitle>
          <CardDescription>{getInstalledDescription(integration)}</CardDescription>
        </CardHeader>
        <CardContent>
          <DescriptionList items={items} />
        </CardContent>
      </Card>

      {guide ? (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">{fill(guide.headline)}</CardTitle>
            <CardDescription>{renderInlineText(fill(guide.summary))}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            {integration.type === "provider" ? (
              <p className="text-sm text-muted-foreground">{i18next.t("integration:Provider guide note")}</p>
            ) : null}
            <ol className="space-y-6">
              {guide.steps.map((step: any, index: number) => (
                <li key={index} className="flex gap-3">
                  <span className="flex size-6 shrink-0 items-center justify-center rounded-full bg-primary text-xs font-semibold text-primary-foreground">{index + 1}</span>
                  <div className="min-w-0 flex-1 space-y-2">
                    <h3 className="font-medium">{fill(step.title)}</h3>
                    <p className="text-sm leading-relaxed">{renderInlineText(fill(step.text))}</p>
                    {step.fields ? (
                      <div className="overflow-x-auto rounded-md border">
                        <table className="w-full text-sm">
                          <tbody>
                            {step.fields.map((field: any) => (
                              <tr key={field.label} className="border-b last:border-0">
                                <td className="whitespace-nowrap bg-muted/50 px-3 py-2 font-medium">{fill(field.label)}</td>
                                <td className="px-3 py-2 font-mono text-xs break-all">{fill(field.value)}</td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    ) : null}
                    {step.code ? (
                      <CodeBlock content={fill(step.code.content)} />
                    ) : null}
                  </div>
                </li>
              ))}
            </ol>

            {guide.tips?.length ? (
              <div className="space-y-2">
                <h3 className="font-medium">{i18next.t("integration:Good to know")}</h3>
                <ul className="list-disc space-y-1 pl-5 text-sm text-muted-foreground">
                  {guide.tips.map((tip: string, index: number) => <li key={index}>{renderInlineText(fill(tip))}</li>)}
                </ul>
              </div>
            ) : null}

            {guide.sources?.length ? (
              <p className="text-xs text-muted-foreground">
                {i18next.t("integration:Sources")}:{" "}
                {guide.sources.map((source: any, index: number) => (
                  <React.Fragment key={source.href}>
                    {index > 0 ? ", " : null}
                    <a className="underline-offset-4 hover:underline" href={source.href} target="_blank" rel="noreferrer">{source.label}</a>
                  </React.Fragment>
                ))}
              </p>
            ) : null}
          </CardContent>
        </Card>
      ) : null}
    </div>
  );
}

function CodeBlock({content}: {content: string}) {
  const copy = () => {
    navigator.clipboard?.writeText(content).then(() => Setting.showMessage("success", i18next.t("general:Copied to clipboard successfully")));
  };
  return (
    <div className="relative">
      <pre className="overflow-x-auto rounded-md bg-muted p-3 pr-16 font-mono text-xs leading-relaxed">{content}</pre>
      <Button variant="outline" size="sm" className="absolute right-2 top-2 h-7 px-2 text-xs" onClick={copy}>
        {i18next.t("general:Copy")}
      </Button>
    </div>
  );
}

function getInstalledDescription(integration: any): string {
  switch (integration.type) {
  case "app":
    return i18next.t("integration:App installed description");
  case "provider":
    return integration.application ? i18next.t("integration:Provider installed in application description") : i18next.t("integration:Provider installed description");
  case "theme":
    return i18next.t("integration:Theme installed description");
  default:
    return "";
  }
}

function getUninstallDescription(integration: any): string {
  switch (integration.type) {
  case "app":
    return i18next.t("integration:Uninstall app description").replace("{name}", integration.application);
  case "provider":
    return i18next.t("integration:Uninstall provider description").replace("{name}", integration.provider);
  case "theme":
    return i18next.t("integration:Uninstall theme description").replace("{name}", integration.application);
  default:
    return "";
  }
}
