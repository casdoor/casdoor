import * as React from "react";
import i18next from "i18next";
import {ExternalLink, Search, ShieldCheck} from "lucide-react";
import {Link, useNavigate} from "react-router-dom";
import {Badge} from "@/components/ui/badge";
import {Button} from "@/components/ui/button";
import {Card} from "@/components/ui/card";
import {Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle} from "@/components/ui/dialog";
import {Input} from "@/components/ui/input";
import {Label} from "@/components/ui/label";
import {Tabs, TabsList, TabsTrigger} from "@/components/ui/tabs";
import {Loading} from "@/components/common/Loading";
import {SelectField} from "@/components/common/SelectField";
import {PageHeader} from "@/components/crud/PageHeader";
import {useAccount} from "@/hooks/use-account";
import {useRequestOrganization} from "@/hooks/use-organization";
import * as ApplicationBackend from "@/backend/ApplicationBackend";
import * as IntegrationBackend from "@/backend/IntegrationBackend";
import {IntegrationLogo, IntegrationTypes, getIntegrationTypeLabel, getMarketplaceSiteUrl, getLocalized} from "@/lib/integration";
import * as Setting from "@/lib/setting";

/** The Casdoor Marketplace: browse the catalog and install app integrations, provider templates and themes. */
export default function MarketplacePage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();

  const [loading, setLoading] = React.useState(false);
  const [index, setIndex] = React.useState<any>(null);
  const [installed, setInstalled] = React.useState<any[]>([]);
  const [type, setType] = React.useState("all");
  const [search, setSearch] = React.useState("");
  const [selected, setSelected] = React.useState<any>(null);

  const fetchIndex = React.useCallback((refresh: boolean) => {
    setLoading(true);
    IntegrationBackend.getMarketplaceIndex(organizationName, refresh)
      .then((res: any) => {
        setLoading(false);
        if (res.status === "ok") {
          setIndex(res.data);
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to get")}: ${res.msg}`);
        }
      })
      .catch((error: any) => {
        setLoading(false);
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }, [organizationName]);

  const fetchInstalled = React.useCallback(() => {
    IntegrationBackend.getIntegrations(organizationName).then((res: any) => {
      if (res.status === "ok") {
        setInstalled(res.data ?? []);
      }
    });
  }, [organizationName]);

  React.useEffect(() => {
    fetchIndex(false);
    fetchInstalled();
  }, [fetchIndex, fetchInstalled]);

  const integrations = React.useMemo(() => {
    const filter = search.trim().toLowerCase();
    return (index?.integrations ?? [])
      .filter((item: any) => IntegrationTypes.includes(item.type))
      .filter((item: any) => type === "all" || item.type === type)
      .filter((item: any) => {
        if (!filter) {
          return true;
        }
        const text = [item.id, getLocalized(item.name), getLocalized(item.description), ...(item.tags ?? []), ...(item.categories ?? [])].join(" ");
        return text.toLowerCase().includes(filter);
      });
  }, [index, type, search]);

  const getInstalledCount = (id: string) => installed.filter((integration) => integration.integrationId === id).length;

  if (!account) {
    return null;
  }

  return (
    <div className="space-y-4">
      <PageHeader
        title={i18next.t("general:Marketplace")}
        description={i18next.t("integration:Marketplace description")}
        actions={
          <>
            <Button variant="outline" asChild>
              <Link to="/integrations">{i18next.t("general:Integrations")} ({installed.length})</Link>
            </Button>
            <Button variant="outline" onClick={() => fetchIndex(true)} disabled={loading}>
              {i18next.t("general:Refresh")}
            </Button>
          </>
        }
      />

      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <Tabs value={type} onValueChange={setType}>
          <TabsList>
            <TabsTrigger value="all">{i18next.t("general:All")}</TabsTrigger>
            {IntegrationTypes.map((t) => (
              <TabsTrigger key={t} value={t}>{getIntegrationTypeLabel(t)}</TabsTrigger>
            ))}
          </TabsList>
        </Tabs>
        <div className="relative sm:w-72">
          <Search className="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input className="pl-8" placeholder={i18next.t("general:Search")} value={search} onChange={(e) => setSearch(e.target.value)} />
        </div>
      </div>

      {loading && !index ? (
        <Loading />
      ) : integrations.length === 0 ? (
        <p className="py-16 text-center text-muted-foreground">{i18next.t("general:No data")}</p>
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {integrations.map((item: any) => {
            const name = getLocalized(item.name) || item.id;
            const installedCount = getInstalledCount(item.id);
            return (
              <Card key={item.id} className="flex h-full flex-col overflow-hidden">
                {item.type === "theme" && item.screenshots?.length ? (
                  <img src={item.screenshots[0]} alt="" className="aspect-[16/10] w-full border-b object-cover" />
                ) : null}
                <div className="flex flex-1 flex-col gap-3 p-4">
                  <div className="flex items-start gap-3">
                    {item.type !== "theme" ? <IntegrationLogo logo={item.logo} name={name} /> : null}
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center gap-1.5">
                        <span className="truncate font-semibold">{name}</span>
                        {item.verified ? <ShieldCheck className="size-4 shrink-0 text-primary" aria-label={i18next.t("user:Verified")} /> : null}
                      </div>
                      <div className="mt-0.5 flex flex-wrap gap-1">
                        <Badge variant="secondary">{getIntegrationTypeLabel(item.type)}</Badge>
                        {installedCount > 0 ? <Badge variant="success">{i18next.t("integration:Installed")}{installedCount > 1 ? ` (${installedCount})` : ""}</Badge> : null}
                      </div>
                    </div>
                  </div>
                  <p className="line-clamp-3 flex-1 text-sm text-muted-foreground">{getLocalized(item.description)}</p>
                  {item.tags?.length ? (
                    <div className="flex flex-wrap gap-1">
                      {item.tags.map((tag: string) => <Badge key={tag} variant="outline" className="font-normal">{tag}</Badge>)}
                    </div>
                  ) : null}
                  <div className="flex items-center justify-between gap-2 pt-1">
                    <a
                      href={`${getMarketplaceSiteUrl()}/integrations/${item.id}/`}
                      target="_blank"
                      rel="noreferrer"
                      className="inline-flex items-center gap-1 text-sm text-muted-foreground underline-offset-4 hover:text-foreground hover:underline"
                    >
                      {i18next.t("integration:Guide")}
                      <ExternalLink className="size-3.5" />
                    </a>
                    <Button size="sm" onClick={() => setSelected(item)}>{i18next.t("integration:Install")}</Button>
                  </div>
                </div>
              </Card>
            );
          })}
        </div>
      )}

      {selected ? (
        <InstallDialog
          item={selected}
          account={account}
          organizationName={organizationName}
          installed={installed}
          onClose={() => setSelected(null)}
        />
      ) : null}
    </div>
  );
}

function InstallDialog({item, account, organizationName, installed, onClose}: {
  item: any;
  account: any;
  organizationName: string;
  installed: any[];
  onClose: () => void;
}) {
  const navigate = useNavigate();
  const [bundle, setBundle] = React.useState<any>(null);
  const [error, setError] = React.useState("");
  const [applications, setApplications] = React.useState<any[]>([]);
  const [name, setName] = React.useState(() => {
    const taken = new Set(installed.map((integration) => integration.name));
    let candidate = item.id;
    for (let i = 2; taken.has(candidate); i++) {
      candidate = `${item.id}-${i}`;
    }
    return candidate;
  });
  const [variables, setVariables] = React.useState<Record<string, string>>({});
  const [application, setApplication] = React.useState("");
  const [clientId, setClientId] = React.useState("");
  const [clientSecret, setClientSecret] = React.useState("");
  const [installing, setInstalling] = React.useState(false);

  React.useEffect(() => {
    IntegrationBackend.getMarketplaceBundle(organizationName, item.id).then((res: any) => {
      if (res.status === "ok") {
        setBundle(res.data);
        const defaults: Record<string, string> = {};
        for (const variable of res.data?.manifest?.variables ?? []) {
          defaults[variable.name] = variable.type === "color" ? variable.example ?? "" : "";
        }
        setVariables(defaults);
      } else {
        setError(res.msg);
      }
    });
    if (item.type !== "app") {
      ApplicationBackend.getApplicationsByOrganization("admin", organizationName).then((res: any) => {
        if (res.status === "ok") {
          setApplications(res.data ?? []);
        }
      });
    }
  }, [item, organizationName]);

  const manifestVariables: any[] = bundle?.manifest?.variables ?? [];
  const needsGlobalAdmin = item.requiresGlobalAdmin && !Setting.isAdminUser(account);
  const isMissing = !name
    || manifestVariables.some((variable) => variable.required && !variables[variable.name]?.trim())
    || (item.type === "provider" && (!clientId || !clientSecret))
    || (item.type === "theme" && !application);

  const install = () => {
    setInstalling(true);
    IntegrationBackend.installIntegration({
      owner: organizationName,
      name,
      integrationId: item.id,
      version: bundle?.manifest?.version ?? "",
      variables,
      application,
      clientId,
      clientSecret,
    })
      .then((res: any) => {
        setInstalling(false);
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("integration:Installed"));
          navigate(`/integrations/${res.data.owner}/${res.data.name}`);
        } else {
          Setting.showMessage("error", `${i18next.t("integration:Failed to install")}: ${res.msg}`);
        }
      })
      .catch((err: any) => {
        setInstalling(false);
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${err}`);
      });
  };

  const itemName = getLocalized(item.name) || item.id;
  const applicationOptions = applications.map((app) => ({id: app.name, name: app.displayName ? `${app.displayName} (${app.name})` : app.name}));

  return (
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-lg">
        <DialogHeader>
          <div className="flex items-center gap-3">
            <IntegrationLogo logo={item.logo} name={itemName} />
            <div>
              <DialogTitle>{i18next.t("integration:Install")} {itemName}</DialogTitle>
              <DialogDescription>{getIntegrationTypeLabel(item.type)}, {bundle?.manifest?.version ?? item.latest}</DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <p className="text-sm text-muted-foreground">{getLocalized(item.description)}</p>
        <p className="text-sm">{getInstallSummary(item.type, itemName)}</p>

        {error ? (
          <p className="text-sm text-destructive">{error}</p>
        ) : !bundle ? (
          <Loading />
        ) : (
          <div className="space-y-4">
            {needsGlobalAdmin ? (
              <p className="rounded-md border border-warning/30 bg-warning/10 p-3 text-sm text-warning">
                {i18next.t("integration:Only a global admin can install this theme")}
              </p>
            ) : null}

            <FieldRow label={i18next.t("general:Organization")}>
              <Input value={organizationName} disabled />
            </FieldRow>
            <FieldRow label={i18next.t("general:Name")} required help={i18next.t("integration:Name help")}>
              <Input value={name} onChange={(e) => setName(e.target.value)} />
            </FieldRow>

            {manifestVariables.map((variable) => (
              <FieldRow
                key={variable.name}
                label={getLocalized(variable.label) || variable.name}
                required={variable.required}
                help={getLocalized(variable.description)}
              >
                {variable.type === "color" ? (
                  <div className="flex gap-2">
                    <input
                      type="color"
                      className="h-9 w-12 cursor-pointer rounded-md border bg-transparent p-1"
                      value={/^#[0-9a-fA-F]{6}$/.test(variables[variable.name] ?? "") ? variables[variable.name] : "#000000"}
                      onChange={(e) => setVariables({...variables, [variable.name]: e.target.value})}
                    />
                    <Input
                      value={variables[variable.name] ?? ""}
                      placeholder={variable.example}
                      onChange={(e) => setVariables({...variables, [variable.name]: e.target.value})}
                    />
                  </div>
                ) : (
                  <Input
                    value={variables[variable.name] ?? ""}
                    placeholder={variable.example}
                    onChange={(e) => setVariables({...variables, [variable.name]: e.target.value})}
                  />
                )}
              </FieldRow>
            ))}

            {item.type === "provider" ? (
              <>
                <FieldRow label={i18next.t("provider:Client ID")} required help={i18next.t("integration:Client ID help")}>
                  <Input value={clientId} onChange={(e) => setClientId(e.target.value)} />
                </FieldRow>
                <FieldRow label={i18next.t("provider:Client secret")} required>
                  <Input type="password" autoComplete="new-password" value={clientSecret} onChange={(e) => setClientSecret(e.target.value)} />
                </FieldRow>
                <FieldRow label={i18next.t("general:Application")} help={i18next.t("integration:Add to application help")}>
                  <SelectField
                    value={application || "-"}
                    onChange={(value) => setApplication(value === "-" ? "" : value)}
                    options={[{id: "-", name: i18next.t("general:None")}, ...applicationOptions]}
                  />
                </FieldRow>
              </>
            ) : null}

            {item.type === "theme" ? (
              <FieldRow label={i18next.t("general:Application")} required help={i18next.t("integration:Theme application help")}>
                <SelectField value={application} onChange={setApplication} options={applicationOptions} />
              </FieldRow>
            ) : null}
          </div>
        )}

        <DialogFooter>
          <Button variant="outline" onClick={onClose}>{i18next.t("general:Cancel")}</Button>
          <Button onClick={install} disabled={!bundle || isMissing || needsGlobalAdmin || installing}>
            {installing ? i18next.t("integration:Installing") : i18next.t("integration:Install")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function getInstallSummary(type: string, name: string): string {
  switch (type) {
  case "app":
    return i18next.t("integration:App install summary").replace("{name}", name);
  case "provider":
    return i18next.t("integration:Provider install summary").replace("{name}", name);
  case "theme":
    return i18next.t("integration:Theme install summary");
  default:
    return "";
  }
}

function FieldRow({label, required, help, children}: {label: React.ReactNode; required?: boolean; help?: React.ReactNode; children: React.ReactNode}) {
  return (
    <div className="space-y-1.5">
      <Label>
        {label}
        {required ? <span className="ml-0.5 text-destructive">*</span> : null}
      </Label>
      {children}
      {help ? <p className="text-xs text-muted-foreground">{help}</p> : null}
    </div>
  );
}
