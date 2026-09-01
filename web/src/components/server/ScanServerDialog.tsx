import * as React from "react";
import i18next from "i18next";
import {Loader2} from "lucide-react";
import {Button} from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {SearchableSelect} from "@/components/common/SearchableSelect";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table";
import * as ProviderBackend from "@/backend/ProviderBackend";
import * as ServerBackend from "@/backend/ServerBackend";
import * as Setting from "@/lib/setting";

interface ScannedServer {
  host: string;
  port: number;
  path: string;
  url: string;
}

interface ScanServerDialogProps {
  organizationName: string;
  /** re-fetch the server list after a scanned server was added */
  onAdded: () => void;
}

/**
 * "Scan server" on the MCP server list: picks an Intranet Scan provider, asks the
 * backend to sweep the intranet, and adds any MCP server it found. Port of
 * web/src/common/modal/ScanServerModal.js plus the scan state of ServerListPage.
 */
export function ScanServerDialog({organizationName, onAdded}: ScanServerDialogProps) {
  const [open, setOpen] = React.useState(false);
  const [loading, setLoading] = React.useState(false);
  const [providers, setProviders] = React.useState<any[]>([]);
  const [selectedProvider, setSelectedProvider] = React.useState("");
  const [result, setResult] = React.useState<any>(null);
  const [servers, setServers] = React.useState<ScannedServer[]>([]);

  const openDialog = () => {
    setResult(null);
    setServers([]);
    ProviderBackend.getProviders(organizationName, 1, 200)
      .then((res: any) => {
        if (res.status !== "ok") {
          Setting.showMessage("error", `${i18next.t("general:Failed to get")}: ${res.msg}`);
          return;
        }
        const scanProviders = (res.data ?? []).filter(
          (provider: any) =>
            provider.category === "Scan" && provider.type === "MCP Scan" && provider.subType === "Intranet Scan",
        );
        setProviders(scanProviders);
        setSelectedProvider(scanProviders.length > 0 ? `${scanProviders[0].owner}/${scanProviders[0].name}` : "");
        setOpen(true);
      })
      .catch((error: any) => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  };

  const scan = () => {
    if (!selectedProvider) {
      Setting.showMessage("error", i18next.t("server:Please select a provider"));
      return;
    }
    const [providerOwner, providerName] = selectedProvider.split("/");
    setLoading(true);
    ServerBackend.syncIntranetServers(providerOwner, providerName)
      .then((res: any) => {
        if (res.status === "ok") {
          const scanResult = res.data ?? {};
          const scanServers = scanResult.servers ?? [];
          setResult(scanResult);
          setServers(scanServers);
          Setting.showMessage("success", `${i18next.t("general:Successfully got")}: ${scanServers.length} server(s)`);
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to get")}: ${res.msg}`);
        }
      })
      .catch((error: any) => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      })
      .finally(() => setLoading(false));
  };

  const addScannedServer = (scanServer: ScannedServer) => {
    ServerBackend.addServer({
      owner: organizationName,
      name: `server_${Setting.getRandomName()}`,
      createdTime: new Date().toISOString(),
      displayName: `Scanned MCP ${scanServer.host}:${scanServer.port}`,
      url: scanServer.url,
      application: "",
    })
      .then((res: any) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully added"));
          onAdded();
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to add")}: ${res.msg}`);
        }
      })
      .catch((error: any) => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  };

  return (
    <>
      <Button variant="outline" onClick={openDialog}>
        {i18next.t("server:Scan server")}
      </Button>
      <Dialog open={open} onOpenChange={(next) => (loading ? undefined : setOpen(next))}>
        <DialogContent className="max-w-4xl">
          <DialogHeader>
            <DialogTitle>{i18next.t("server:Scan server")}</DialogTitle>
            {result ? (
              <DialogDescription>
                {`${i18next.t("server:Scanned hosts")}:${result.scannedHosts ?? 0}, ` +
                  `${i18next.t("server:Online hosts")}:${result.onlineHosts?.length ?? 0}, ` +
                  `${i18next.t("server:Found servers")}:${servers.length}`}
              </DialogDescription>
            ) : null}
          </DialogHeader>

          <SearchableSelect
            value={selectedProvider}
            onChange={setSelectedProvider}
            options={providers.map((provider: any) => ({
              value: `${provider.owner}/${provider.name}`,
              label: provider.displayName || provider.name,
            }))}
          />

          {result === null ? null : (
            <div className="max-h-80 overflow-auto rounded-lg border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-36">{i18next.t("general:Host")}</TableHead>
                    <TableHead className="w-24">{i18next.t("general:Port")}</TableHead>
                    <TableHead className="w-32">{i18next.t("general:Path")}</TableHead>
                    <TableHead>{i18next.t("general:URL")}</TableHead>
                    <TableHead className="w-28">{i18next.t("general:Action")}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {servers.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={5} className="text-center text-sm text-muted-foreground">
                        {i18next.t("general:No data")}
                      </TableCell>
                    </TableRow>
                  ) : (
                    servers.map((server, index) => (
                      <TableRow key={`${server.url}-${index}`}>
                        <TableCell>{server.host}</TableCell>
                        <TableCell>{server.port}</TableCell>
                        <TableCell>{server.path}</TableCell>
                        <TableCell>
                          {server.url ? (
                            <a
                              className="text-primary underline-offset-4 hover:underline"
                              target="_blank"
                              rel="noreferrer"
                              href={server.url}
                            >
                              {Setting.getShortText(server.url, 60)}
                            </a>
                          ) : null}
                        </TableCell>
                        <TableCell>
                          <Button size="sm" onClick={() => addScannedServer(server)}>
                            {i18next.t("general:Add")}
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </div>
          )}

          <DialogFooter>
            <Button variant="outline" onClick={() => setOpen(false)} disabled={loading}>
              {i18next.t("general:Cancel")}
            </Button>
            <Button onClick={scan} disabled={loading}>
              {loading ? <Loader2 className="animate-spin" /> : null}
              {i18next.t("general:Sync")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
