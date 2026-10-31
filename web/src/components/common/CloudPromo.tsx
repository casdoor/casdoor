import i18next from "i18next";
import {ArrowUpRight, Cloud} from "lucide-react";
import {Button} from "@/components/ui/button";
import {DropdownMenuItem} from "@/components/ui/dropdown-menu";
import {SidebarMenu, SidebarMenuButton, SidebarMenuItem} from "@/components/ui/sidebar";
import * as Conf from "@/Conf";
import {useAccount} from "@/hooks/use-account";
import * as Setting from "@/lib/setting";
import {cn} from "@/lib/utils";

/**
 * The entries that point a deployment's admins at the hosted offering, unless a
 * global admin turned them off in the global settings. Each one tags its link
 * with the place it sits in, so www.casdoor.com's analytics can tell them apart.
 */
export function showCloudPromo(account: any): boolean {
  return Conf.ShowCloudPromotion && Setting.isLocalAdminUser(account);
}

/**
 * The sign-in page belongs to the deployment's own users rather than its admins,
 * so only the demo deployment (door and demo.casdoor.com) advertises there.
 */
export function showCloudLoginLink(): boolean {
  return Conf.ShowCloudPromotion && Conf.ShowGithubCorner;
}

function useCloudPromo(): boolean {
  const {account} = useAccount();
  return showCloudPromo(account);
}

export function getCloudUrl(placement: string, path = "/"): string {
  const params = new URLSearchParams({
    utm_source: window.location.hostname,
    utm_medium: "referral",
    utm_campaign: placement,
  });
  return `https://www.casdoor.com${path}?${params}`;
}

function CloudIcon({className}: {className?: string}) {
  return (
    <span
      className={cn(
        "flex shrink-0 items-center justify-center rounded-md bg-gradient-to-br from-sky-500 to-indigo-500 text-white shadow-sm",
        className,
      )}
    >
      <Cloud className="h-[60%] w-[60%]" />
    </span>
  );
}

export function CloudHeaderButton() {
  const visible = useCloudPromo();
  if (!visible) {
    return null;
  }

  return (
    <a
      href={getCloudUrl("header")}
      target="_blank"
      rel="noreferrer"
      className="group mr-1 inline-flex h-8 items-center gap-1.5 rounded-full border border-sky-500/30 bg-gradient-to-r from-sky-500/10 to-indigo-500/10 pl-1 pr-2.5 text-sm font-medium text-foreground transition-colors hover:border-sky-500/60 hover:from-sky-500/20 hover:to-indigo-500/20"
    >
      <CloudIcon className="h-6 w-6 rounded-full" />
      <span className="hidden whitespace-nowrap sm:inline">Casdoor Cloud</span>
      <ArrowUpRight className="h-3.5 w-3.5 text-muted-foreground transition-transform group-hover:-translate-y-0.5 group-hover:translate-x-0.5" />
    </a>
  );
}

export function CloudSidebarCard({collapsed}: {collapsed: boolean}) {
  const visible = useCloudPromo();
  if (!visible) {
    return null;
  }

  if (collapsed) {
    return (
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton asChild tooltip="Casdoor Cloud">
            <a href={getCloudUrl("sidebar")} target="_blank" rel="noreferrer">
              <CloudIcon className="h-4 w-4 rounded" />
              <span>Casdoor Cloud</span>
            </a>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    );
  }

  return (
    <div className="rounded-lg border border-sky-500/20 bg-gradient-to-br from-sky-500/10 via-transparent to-indigo-500/10 p-3">
      <div className="flex items-center gap-2">
        <CloudIcon className="h-6 w-6" />
        <span className="text-sm font-semibold">Casdoor Cloud</span>
      </div>
      <p className="mt-2 text-xs leading-relaxed text-muted-foreground">
        {i18next.t("general:Hosted Casdoor, kept up to date for you")}
      </p>
      <Button asChild size="sm" className="mt-3 w-full">
        <a href={getCloudUrl("sidebar", "/pricing")} target="_blank" rel="noreferrer">
          {i18next.t("general:Start free trial")}
        </a>
      </Button>
    </div>
  );
}

export function CloudDashboardBanner() {
  const visible = useCloudPromo();
  if (!visible) {
    return null;
  }

  return (
    <div className="relative overflow-hidden rounded-xl border border-sky-500/20 bg-gradient-to-r from-sky-500/10 via-indigo-500/5 to-transparent p-4 sm:p-5">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center">
        <CloudIcon className="h-10 w-10 rounded-lg" />
        <div className="min-w-0 flex-1">
          <div className="font-semibold">{i18next.t("general:Get your own Casdoor in minutes")}</div>
          <p className="mt-1 text-sm text-muted-foreground">
            {i18next.t("general:Casdoor Cloud runs a dedicated Casdoor instance for you, with no servers to set up. Upgrades and security patches are handled for you.")}
          </p>
        </div>
        <div className="flex shrink-0 flex-wrap gap-2">
          <Button asChild>
            <a href={getCloudUrl("dashboard", "/pricing")} target="_blank" rel="noreferrer">
              {i18next.t("general:Start free trial")}
            </a>
          </Button>
          <Button asChild variant="outline">
            <a href={getCloudUrl("dashboard")} target="_blank" rel="noreferrer">
              {i18next.t("general:Learn more")}
              <ArrowUpRight />
            </a>
          </Button>
        </div>
      </div>
    </div>
  );
}

export function CloudMenuItem() {
  const visible = useCloudPromo();
  if (!visible) {
    return null;
  }

  return (
    <DropdownMenuItem asChild>
      <a href={getCloudUrl("account-menu")} target="_blank" rel="noreferrer">
        <Cloud />
        Casdoor Cloud
        <ArrowUpRight className="ml-auto opacity-60" />
      </a>
    </DropdownMenuItem>
  );
}

export function CloudLoginLink() {
  if (!showCloudLoginLink()) {
    return null;
  }

  return (
    <a
      href={getCloudUrl("login")}
      target="_blank"
      rel="noreferrer"
      className="group inline-flex items-center gap-2 rounded-full border bg-background/80 py-1 pl-1 pr-3 text-sm text-muted-foreground shadow-sm backdrop-blur transition-colors hover:border-sky-500/50 hover:text-foreground"
    >
      <CloudIcon className="h-6 w-6 rounded-full" />
      {i18next.t("general:Want your own Casdoor?")}
      <span className="font-medium text-foreground">Casdoor Cloud</span>
      <ArrowUpRight className="h-3.5 w-3.5 transition-transform group-hover:-translate-y-0.5 group-hover:translate-x-0.5" />
    </a>
  );
}
