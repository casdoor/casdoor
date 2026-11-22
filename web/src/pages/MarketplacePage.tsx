import * as React from "react";
import i18next from "i18next";
import {ExternalLink, Monitor, Search, ShieldCheck, Smartphone, Upload, X} from "lucide-react";
import {Link, useNavigate, useSearchParams} from "react-router-dom";
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
import {IntegrationLogo, IntegrationTypes, ThemePreviewMessage, applyTheme, getActionKind, getIntegrationTypeLabel, getMarketplaceSiteUrl, getLocalized, getActionText} from "@/lib/integration";
import * as Setting from "@/lib/setting";
import {getEditionText} from "@/components/common/EditionNotice";

/** The Casdoor Marketplace: browse the catalog and install app integrations, provider templates and themes. */
export default function MarketplacePage() {
  const {account} = useAccount();
  const organizationName = useRequestOrganization();

  const [loading, setLoading] = React.useState(false);
  const [index, setIndex] = React.useState<any>(null);
  const [installed, setInstalled] = React.useState<any[]>([]);
  const [type, setType] = React.useState("all");
  const [searchParams, setSearchParams] = useSearchParams();
  const [search, setSearch] = React.useState(() => searchParams.get("q") ?? "");
  // "?edition=Enterprise" lists only the paid entries of that edition, as linked from the editions page
  const editionFilter = searchParams.get("edition") ?? "";
  const [selected, setSelected] = React.useState<any>(null);
  const fileInputRef = React.useRef<HTMLInputElement>(null);

  // an integration that isn't in the Marketplace, from the bundle.json its build wrote and, for an action, the module next to it
  const openFiles = async(fileList: FileList | null) => {
    const files = Array.from(fileList ?? []);
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
    let bundle: any = null;
    let text = "";
    for (const file of files.filter((f) => f.name.toLowerCase().endsWith(".json"))) {
      text = await file.text();
      try {
        bundle = JSON.parse(text);
      } catch {
        bundle = null;
      }
      if (bundle?.manifest) {
        break;
      }
    }
    const manifest = bundle?.manifest;
    if (!manifest?.id || !IntegrationTypes.includes(manifest.type)) {
      Setting.showMessage("error", i18next.t("integration:Not a bundle file"));
      return;
    }

    let module = "";
    if (manifest.type === "action") {
      const moduleName = bundle.files?.["action.json"]?.module ?? "";
      const moduleFile = files.find((file) => file.name === moduleName);
      if (!moduleFile) {
        Setting.showMessage("error", i18next.t("integration:Choose the module too").replace("{name}", moduleName || ".wasm"));
        return;
      }
      module = await readBase64(moduleFile);
    }

    setSelected({
      id: manifest.id,
      type: manifest.type,
      name: manifest.name,
      description: manifest.description,
      latest: manifest.version,
      local: {bundle, text, module},
    });
  };

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
      .filter((item: any) => !editionFilter || (item.edition ?? "").toLowerCase() === editionFilter.toLowerCase())
      .filter((item: any) => {
        if (!filter) {
          return true;
        }
        const text = [item.id, getLocalized(item.name), getLocalized(item.description), ...(item.tags ?? []), ...(item.categories ?? [])].join(" ");
        return text.toLowerCase().includes(filter);
      });
  }, [index, type, search, editionFilter]);

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
            {Setting.isAdminUser(account) ? (
              <>
                <Button variant="outline" className="gap-1.5" title={i18next.t("integration:Install from file help")} onClick={() => fileInputRef.current?.click()}>
                  <Upload className="size-4" />
                  {i18next.t("integration:Install from file")}
                </Button>
                <input ref={fileInputRef} type="file" accept=".json,.wasm" multiple className="hidden" onChange={(e) => openFiles(e.target.files)} />
              </>
            ) : null}
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
        {editionFilter ? (
          <Button
            variant="secondary"
            size="sm"
            className="gap-1 rounded-full sm:ml-auto"
            onClick={() => {
              searchParams.delete("edition");
              setSearchParams(searchParams, {replace: true});
            }}
          >
            {getEditionText(editionFilter.charAt(0).toUpperCase() + editionFilter.slice(1).toLowerCase())}
            <X className="size-3.5" />
          </Button>
        ) : null}
        <div className="relative sm:w-72">
          <Search className="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input className="pl-8" placeholder={i18next.t("general:Search")} value={search} onChange={(e) => setSearch(e.target.value)} />
        </div>
      </div>

      {loading && !index ? (
        <Loading />
      ) : integrations.length === 0 ? (
        <p className="py-16 text-center text-muted-foreground" data-tour="marketplace-items">{i18next.t("general:No data")}</p>
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4" data-tour="marketplace-items">
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
                        {item.edition ? <Badge variant="warning">{getEditionText(item.edition)}</Badge> : null}
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
          organizationName={organizationName}
          installed={installed}
          onClose={() => setSelected(null)}
        />
      ) : null}
    </div>
  );
}

function InstallDialog({item, organizationName, installed, onClose}: {
  item: any;
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
  const [previewApplication, setPreviewApplication] = React.useState<any>(null);
  // what already has the name, checked for the name in checkedName; the suggested name moves on while the user hasn't typed one
  const [nameConflict, setNameConflict] = React.useState("");
  const [checkedName, setCheckedName] = React.useState("");
  const [isNameEdited, setIsNameEdited] = React.useState(false);

  React.useEffect(() => {
    if (!name) {
      setNameConflict("");
      setCheckedName("");
      return;
    }
    let cancelled = false;
    const timer = setTimeout(() => {
      IntegrationBackend.getIntegrationNameConflict(organizationName, name, item.type)
        .then((res: any) => {
          if (cancelled) {
            return;
          }
          const conflict = res.status === "ok" ? res.data ?? "" : "";
          if (conflict && conflict !== "invalid" && !isNameEdited) {
            const match = name.match(/-(\d+)$/);
            const next = match ? Number(match[1]) + 1 : 2;
            if (next <= 50) {
              setName(`${item.id}-${next}`);
              return;
            }
          }
          setNameConflict(conflict);
          setCheckedName(name);
        })
        .catch(() => {
          if (!cancelled) {
            setNameConflict("");
            setCheckedName(name);
          }
        });
    }, 300);
    return () => {
      cancelled = true;
      clearTimeout(timer);
    };
  }, [name, item, organizationName, isNameEdited]);

  React.useEffect(() => {
    const applyBundle = (data: any) => {
      setBundle(data);
      const defaults: Record<string, string> = {};
      for (const variable of data?.manifest?.variables ?? []) {
        defaults[variable.name] = variable.type === "color" ? variable.example ?? "" : "";
      }
      setVariables(defaults);
    };
    if (item.local) {
      applyBundle(item.local.bundle);
    } else {
      IntegrationBackend.getMarketplaceBundle(organizationName, item.id).then((res: any) => {
        if (res.status === "ok") {
          applyBundle(res.data);
        } else {
          setError(res.msg);
        }
      });
    }
    if (item.type !== "app") {
      ApplicationBackend.getApplicationsByOrganization("admin", organizationName).then((res: any) => {
        if (res.status === "ok") {
          setApplications(res.data ?? []);
          if (item.type === "theme" && res.data?.length) {
            setApplication((current) => current || res.data[0].name);
          }
        }
      });
    }
  }, [item, organizationName]);

  React.useEffect(() => {
    if (item.type !== "theme" || !application) {
      setPreviewApplication(null);
      return;
    }
    ApplicationBackend.getApplication("admin", application).then((res: any) => {
      if (res.status === "ok") {
        setPreviewApplication(res.data);
      }
    });
  }, [item, application]);

  const manifestVariables: any[] = bundle?.manifest?.variables ?? [];
  const actionKind = getActionKind(bundle);
  const getVariableError = (variable: any) => {
    const value = variables[variable.name]?.trim() ?? "";
    if (!value) {
      return "";
    }
    if (variable.type === "color" && !/^#[0-9a-fA-F]{6}$/.test(value)) {
      return i18next.t("integration:Color format error");
    }
    if (variable.type === "url" && !isHttpUrl(value)) {
      return i18next.t("integration:URL format error");
    }
    return "";
  };
  const nameError = checkedName === name ? getNameConflictText(nameConflict, name) : "";
  const isMissing = !name
    || checkedName !== name
    || nameConflict !== ""
    || manifestVariables.some((variable) => getVariableError(variable) !== "")
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
      bundle: item.local?.text ?? "",
      module: item.local?.module ?? "",
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
  const isTheme = item.type === "theme";
  // empty fields are previewed with their examples, so the preview never shows a blank headline
  const previewValues = Object.fromEntries(manifestVariables.map((variable) => {
    const value = variables[variable.name]?.trim() ?? "";
    const valid = variable.type !== "color" || /^#[0-9a-fA-F]{6}$/.test(value);
    return [variable.name, value && valid ? value : variable.example ?? ""];
  }));
  const theme = bundle?.files?.["theme.json"];
  const replacedTheme = isTheme && application ? installed.find((integration) => integration.type === "theme" && integration.application === application) : null;
  const applicationOptions = applications.map((app) => ({id: app.name, name: app.displayName ? `${app.displayName} (${app.name})` : app.name}));

  return (
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent className={isTheme ? "max-h-[92vh] overflow-y-auto sm:max-w-6xl" : "max-h-[90vh] overflow-y-auto sm:max-w-lg"}>
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
        <p className="text-sm">{getInstallSummary(item.type, itemName, actionKind)}</p>
        {item.local ? <p className="text-sm text-warning">{i18next.t("integration:Install from file warning")}</p> : null}

        <div className={isTheme ? "grid gap-6 lg:grid-cols-[minmax(0,340px)_minmax(0,1fr)]" : undefined}>
          {error ? (
            <p className="text-sm text-destructive">{error}</p>
          ) : !bundle ? (
            <Loading />
          ) : (
            <div className="space-y-4">
              <FieldRow label={i18next.t("general:Organization")}>
                <Input value={organizationName} disabled />
              </FieldRow>
              <FieldRow label={i18next.t("general:Name")} required help={i18next.t("integration:Name help")}>
                <Input
                  value={name}
                  aria-invalid={nameError !== ""}
                  onChange={(e) => {
                    setIsNameEdited(true);
                    setName(e.target.value);
                  }}
                />
                {nameError ? <p className="text-xs text-destructive">{nameError}</p> : null}
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
                        value={/^#[0-9a-fA-F]{6}$/.test(variables[variable.name]?.trim() ?? "") ? variables[variable.name].trim() : "#000000"}
                        onChange={(e) => setVariables({...variables, [variable.name]: e.target.value})}
                      />
                      <Input
                        value={variables[variable.name] ?? ""}
                        placeholder={variable.example}
                        aria-invalid={getVariableError(variable) !== ""}
                        onChange={(e) => setVariables({...variables, [variable.name]: e.target.value})}
                      />
                    </div>
                  ) : (
                    <Input
                      value={variables[variable.name] ?? ""}
                      placeholder={variable.example}
                      aria-invalid={getVariableError(variable) !== ""}
                      onChange={(e) => setVariables({...variables, [variable.name]: e.target.value})}
                    />
                  )}
                  {getVariableError(variable) ? <p className="text-xs text-destructive">{getVariableError(variable)}</p> : null}
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

              {item.type === "action" ? (
                <>
                  {actionKind === "sms" || actionKind === "push" ? (
                    <FieldRow label={i18next.t("provider:Client ID")} help={actionKind === "sms" ? i18next.t("integration:SMS action client ID help") : i18next.t("integration:Push action client ID help")}>
                      <Input value={clientId} onChange={(e) => setClientId(e.target.value)} />
                    </FieldRow>
                  ) : null}
                  {["payment", "sms", "push"].includes(actionKind) ? (
                    <FieldRow label={i18next.t("provider:Client secret")} help={actionKind === "sms" ? i18next.t("integration:SMS action client secret help") : i18next.t("integration:Action client secret help")}>
                      <Input type="password" autoComplete="new-password" value={clientSecret} onChange={(e) => setClientSecret(e.target.value)} />
                    </FieldRow>
                  ) : (
                    <p className="text-sm text-muted-foreground">{i18next.t("integration:Plugin action install help")}</p>
                  )}
                  {bundle?.manifest?.permissions?.hosts?.length ? (
                    <FieldRow label={i18next.t("integration:Allowed hosts")} help={i18next.t("integration:Allowed hosts help")}>
                      <div className="space-y-0.5">
                        {bundle.manifest.permissions.hosts.map((host: string) => <div key={host}><code className="font-mono text-xs">{host}</code></div>)}
                      </div>
                    </FieldRow>
                  ) : null}
                </>
              ) : null}

              {bundle?.manifest?.downloads?.length && !item.local ? (
                <FieldRow label={i18next.t("integration:Downloads")} help={i18next.t("integration:Downloads help")}>
                  <div className="flex flex-wrap gap-2">
                    {bundle.manifest.downloads.map((path: string) => (
                      <Button key={path} variant="outline" size="sm" asChild>
                        <a href={`${Setting.ServerUrl}/api/download-marketplace-file?owner=${encodeURIComponent(organizationName)}&integrationId=${encodeURIComponent(item.id)}&path=${encodeURIComponent(path)}`}>
                          {path.split("/").pop()}
                        </a>
                      </Button>
                    ))}
                  </div>
                </FieldRow>
              ) : null}

              {item.type === "theme" ? (
                <FieldRow label={i18next.t("general:Application")} required help={i18next.t("integration:Theme application help")}>
                  <SelectField value={application} onChange={setApplication} options={applicationOptions} />
                  {replacedTheme ? (
                    <p className="text-xs text-amber-600 dark:text-amber-500">
                      {i18next.t("integration:Theme replaces").replace("{name}", replacedTheme.displayName || replacedTheme.name)}
                    </p>
                  ) : null}
                </FieldRow>
              ) : null}
            </div>
          )}
          {isTheme ? (
            <ThemePreview application={previewApplication && theme ? applyTheme(previewApplication, theme, previewValues) : null} />
          ) : null}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={onClose}>{i18next.t("general:Cancel")}</Button>
          <Button onClick={install} disabled={!bundle || isMissing || installing}>
            {installing ? i18next.t("integration:Installing") : i18next.t("integration:Install")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function isHttpUrl(value: string): boolean {
  if (/[\s"'<>`]/.test(value)) {
    return false;
  }
  try {
    const url = new URL(value);
    return (url.protocol === "http:" || url.protocol === "https:") && url.host !== "";
  } catch {
    return false;
  }
}

function getNameConflictText(conflict: string, name: string): string {
  switch (conflict) {
  case "invalid":
    return i18next.t("integration:Name format error");
  case "integration":
    return i18next.t("integration:Integration name taken").replace("{name}", name);
  case "application":
    return i18next.t("integration:Application name taken").replace("{name}", name);
  case "provider":
    return i18next.t("integration:Provider name taken").replace("{name}", name);
  default:
    return "";
  }
}

function readBase64(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result).replace(/^data:[^,]*,/, ""));
    reader.onerror = () => reject(reader.error);
    reader.readAsDataURL(file);
  });
}

/** The sign-in page of the chosen application with the theme applied, live, in a frame of its own. */
function ThemePreview({application}: {application: any}) {
  const frameRef = React.useRef<HTMLIFrameElement>(null);
  const boxRef = React.useRef<HTMLDivElement>(null);
  const [ready, setReady] = React.useState(false);
  const [device, setDevice] = React.useState<"desktop" | "mobile">("desktop");
  const [boxWidth, setBoxWidth] = React.useState(0);

  React.useEffect(() => {
    const onMessage = (event: MessageEvent) => {
      if (event.origin === window.location.origin && event.source === frameRef.current?.contentWindow && event.data?.type === `${ThemePreviewMessage}-ready`) {
        setReady(true);
      }
    };
    window.addEventListener("message", onMessage);
    return () => window.removeEventListener("message", onMessage);
  }, []);

  React.useEffect(() => {
    if (ready && application) {
      frameRef.current?.contentWindow?.postMessage({type: ThemePreviewMessage, application}, window.location.origin);
    }
  }, [ready, application]);

  React.useEffect(() => {
    const box = boxRef.current;
    if (!box) {
      return;
    }
    const observer = new ResizeObserver(() => setBoxWidth(box.clientWidth));
    observer.observe(box);
    return () => observer.disconnect();
  }, []);

  const [frameWidth, frameHeight] = device === "desktop" ? [1280, 800] : [390, 780];
  const scale = boxWidth ? Math.min(1, boxWidth / frameWidth, device === "mobile" ? 520 / frameHeight : 1) : 0;

  return (
    <div className="min-w-0 space-y-2">
      <div className="flex items-center justify-between gap-2">
        <Label>{i18next.t("general:Preview")}</Label>
        <Tabs value={device} onValueChange={(value) => {
          setReady(false);
          setDevice(value as "desktop" | "mobile");
        }}>
          <TabsList className="h-8">
            <TabsTrigger value="desktop" className="h-6 gap-1 px-2 text-xs"><Monitor className="size-3.5" />{i18next.t("integration:Desktop")}</TabsTrigger>
            <TabsTrigger value="mobile" className="h-6 gap-1 px-2 text-xs"><Smartphone className="size-3.5" />{i18next.t("provider:Mobile")}</TabsTrigger>
          </TabsList>
        </Tabs>
      </div>
      <div ref={boxRef} className="w-full">
        <div
          className="relative mx-auto overflow-hidden rounded-lg border bg-muted shadow-sm"
          style={{width: frameWidth * scale, height: frameHeight * scale}}
        >
          {scale ? (
            <iframe
              key={device}
              ref={frameRef}
              title={i18next.t("general:Preview")}
              src={Setting.withBasePath("/marketplace/theme-preview")}
              className="origin-top-left border-0"
              style={{width: frameWidth, height: frameHeight, transform: `scale(${scale})`}}
            />
          ) : null}
          {/* the preview is to look at, not to sign in with */}
          <div className="absolute inset-0 cursor-not-allowed" />
          {!application ? <div className="absolute inset-0 flex items-center justify-center"><Loading /></div> : null}
        </div>
      </div>
      <p className="text-xs text-muted-foreground">{i18next.t("integration:Theme preview note")}</p>
    </div>
  );
}

function getInstallSummary(type: string, name: string, actionKind: string): string {
  switch (type) {
  case "app":
    return i18next.t("integration:App install summary").replace("{name}", name);
  case "provider":
    return i18next.t("integration:Provider install summary").replace("{name}", name);
  case "theme":
    return i18next.t("integration:Theme install summary");
  case "action":
    return getActionText(actionKind, "summary").replace("{name}", name);
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
