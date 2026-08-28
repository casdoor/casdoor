import * as React from "react";
import i18next from "i18next";
import {useNavigate} from "react-router-dom";
import {Badge} from "@/components/ui/badge";
import {Button} from "@/components/ui/button";
import {Card, CardContent, CardHeader, CardTitle} from "@/components/ui/card";
import {Input} from "@/components/ui/input";
import {Loading} from "@/components/common/Loading";
import {MultiSelect} from "@/components/common/MultiSelect";
import {PageHeader} from "@/components/crud/PageHeader";
import {useAccount} from "@/hooks/use-account";
import * as ServerBackend from "@/backend/ServerBackend";
import * as Setting from "@/lib/setting";

interface OnlineServer {
  id: string;
  name: string;
  nameText: string;
  categoriesRaw: string[];
  categoriesLower: string[];
  endpoint: string;
  description: string;
  website?: string;
}

function getOnlineServersFromResponse(data: any): any[] {
  if (Array.isArray(data?.servers)) {
    return data.servers;
  }
  if (Array.isArray(data)) {
    return data;
  }
  if (Array.isArray(data?.data)) {
    return data.data;
  }
  return [];
}

function normalizeOnlineServers(onlineServers: any[]): OnlineServer[] {
  return onlineServers
    .map((server, index) => {
      const categoriesRaw = [server?.category].filter(
        (category) => typeof category === "string" && category.trim() !== "",
      ) as string[];

      return {
        id: server.id ?? `${server.name ?? "server"}-${index}`,
        name: server.name ?? "",
        nameText: (server.name ?? "").toLowerCase(),
        categoriesRaw,
        categoriesLower: categoriesRaw.map((category) => category.toLowerCase()),
        endpoint: server.endpoints?.production ?? server.endpoint ?? "",
        description: server.description ?? "",
        website: server?.maintainer?.website ?? server?.website,
      };
    })
    .filter((server) => server.endpoint.startsWith("http"));
}

function getWebsiteUrl(website: string) {
  if (!website) {
    return "";
  }
  return /^https?:\/\//i.test(website) ? website : `https://${website}`;
}

function getOnlineServerName(onlineServer: OnlineServer) {
  const source = onlineServer.id || onlineServer.name || `server_${Setting.getRandomName()}`;
  const normalized = String(source).toLowerCase().replace(/[^a-z0-9_-]/g, "_").replace(/_+/g, "_").replace(/^_+|_+$/g, "");
  return normalized || `server_${Setting.getRandomName()}`;
}

/** The public MCP server directory: pick one and it is handed to the server edit page as a new server. */
export default function ServerStorePage() {
  const {account} = useAccount();
  const navigate = useNavigate();

  const [loading, setLoading] = React.useState(false);
  const [servers, setServers] = React.useState<OnlineServer[]>([]);
  const [nameFilter, setNameFilter] = React.useState("");
  const [categoryFilter, setCategoryFilter] = React.useState<string[]>([]);

  const fetchOnlineServers = React.useCallback(() => {
    setLoading(true);
    setNameFilter("");
    setCategoryFilter([]);

    ServerBackend.getOnlineServers()
      .then((res: any) => {
        setLoading(false);
        if (res.status === "ok") {
          setServers(normalizeOnlineServers(getOnlineServersFromResponse(res.data)));
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to get")}: ${res.msg}`);
        }
      })
      .catch((error: any) => {
        setLoading(false);
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }, []);

  React.useEffect(() => {
    fetchOnlineServers();
  }, [fetchOnlineServers]);

  const categoryOptions = React.useMemo(() => {
    const categories = servers.flatMap((server) => server.categoriesRaw);
    return [...new Set(categories)]
      .sort((a, b) => a.localeCompare(b))
      .map((category) => ({label: category, value: category.toLowerCase()}));
  }, [servers]);

  const filteredServers = React.useMemo(() => {
    const filter = nameFilter.trim().toLowerCase();
    return servers.filter((server) => {
      const nameMatched = !filter || server.nameText.includes(filter);
      const categoryMatched =
        categoryFilter.length === 0 || categoryFilter.some((category) => server.categoriesLower.includes(category));
      return nameMatched && categoryMatched;
    });
  }, [servers, nameFilter, categoryFilter]);

  const createServerFromOnline = (onlineServer: OnlineServer) => {
    if (!account) {
      return;
    }
    if (!onlineServer.endpoint) {
      Setting.showMessage("error", i18next.t("server:Production endpoint is empty"));
      return;
    }

    const owner = Setting.getRequestOrganization(account);
    const newServer = {
      owner,
      name: getOnlineServerName(onlineServer) + Setting.getRandomName(),
      createdTime: new Date().toISOString(),
      displayName: onlineServer.name || getOnlineServerName(onlineServer),
      url: onlineServer.endpoint,
      application: "",
    };

    navigate(`/servers/${newServer.owner}/${newServer.name}`, {state: {mode: "add", record: newServer}});
  };

  return (
    <div className="space-y-4">
      <PageHeader
        title={i18next.t("general:MCP Store")}
        actions={
          <>
            <Button variant="outline" onClick={() => {
              setNameFilter("");
              setCategoryFilter([]);
            }}>
              {i18next.t("general:Clear")}
            </Button>
            <Button variant="outline" onClick={fetchOnlineServers} disabled={loading}>
              {i18next.t("general:Refresh")}
            </Button>
          </>
        }
      />
      <div className="flex flex-col gap-2 sm:flex-row">
        <Input
          className="sm:max-w-xs"
          placeholder={i18next.t("general:Name")}
          value={nameFilter}
          onChange={(e) => setNameFilter(e.target.value)}
        />
        <div className="sm:w-72">
          <MultiSelect
            value={categoryFilter}
            onChange={setCategoryFilter}
            options={categoryOptions.map((option) => ({value: option.value, label: option.label}))}
            placeholder={i18next.t("general:Category")}
          />
        </div>
      </div>

      {loading ? (
        <Loading />
      ) : filteredServers.length === 0 ? (
        <p className="py-16 text-center text-muted-foreground">{i18next.t("general:No data")}</p>
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {filteredServers.map((server) => (
            <Card key={server.id} className="flex h-full flex-col">
              <CardHeader className="flex-row items-start justify-between gap-2 space-y-0">
                <CardTitle className="text-base">{server.name || "-"}</CardTitle>
                <Button size="sm" onClick={() => createServerFromOnline(server)}>
                  {i18next.t("general:Add")}
                </Button>
              </CardHeader>
              <CardContent className="flex-1 space-y-2 text-sm">
                <p className="min-h-[3rem] text-muted-foreground">{server.description || "-"}</p>
                <p className="break-all">
                  <span className="font-medium">{i18next.t("general:Url")}: </span>
                  {server.website ? (
                    <a
                      target="_blank"
                      rel="noreferrer"
                      href={getWebsiteUrl(server.endpoint)}
                      className="underline-offset-4 hover:underline"
                    >
                      {server.endpoint}
                    </a>
                  ) : "-"}
                </p>
                <p className="break-all">
                  <span className="font-medium">{i18next.t("general:Website")}: </span>
                  {server.website ? (
                    <a
                      target="_blank"
                      rel="noreferrer"
                      href={getWebsiteUrl(server.website)}
                      className="underline-offset-4 hover:underline"
                    >
                      {server.website}
                    </a>
                  ) : "-"}
                </p>
                <div className="flex flex-wrap gap-1">
                  {server.categoriesRaw.map((category) => (
                    <Badge key={`${server.id}-${category}`} variant="secondary">{category}</Badge>
                  ))}
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
