import * as React from "react";
import * as Cookie from "cookie";
import {Outlet} from "react-router-dom";
import {SidebarInset, SidebarProvider} from "@/components/ui/sidebar";
import {EnableMfaNotification} from "@/components/auth/EnableMfaNotification";
import {CommandPalette, useCommandPalette} from "@/components/common/CommandPalette";
import {ConsoleTour} from "@/components/common/ConsoleTour";
import {GithubCorner} from "@/components/common/GithubCorner";
import {ThemeShowcaseBar} from "@/components/common/ThemeShowcaseBar";
import {Loading} from "@/components/common/Loading";
import {Header} from "@/components/layout/Header";
import {PoweredBy} from "@/components/layout/PoweredBy";
import {AppSidebar} from "@/components/layout/Sidebar";
import {useAccount} from "@/hooks/use-account";
import {useThemeData} from "@/hooks/use-application-chrome";
import * as OrganizationBackend from "@/backend/OrganizationBackend";
import {getSafeHtml} from "@/lib/custom-html";
import * as Setting from "@/lib/setting";
import {getConsoleCss, useShowcaseApplication} from "@/lib/theme-showcase";

/**
 * `SidebarProvider` persists the rail through its own `sidebar_state` cookie, so
 * the first render has to start from that cookie or the rail flashes open before
 * settling shut.
 */
function readSidebarCookie(): boolean {
  try {
    const value = Cookie.parse(document.cookie)["sidebar_state"];
    return value === undefined ? true : value === "true";
  } catch {
    return true;
  }
}

/**
 * The footer HTML of the default application of the signed-in user's organization,
 * the same application whose sign-in page the organization's users see.
 */
function useFooterHtml(organizationName: string | undefined): string {
  const [footerHtml, setFooterHtml] = React.useState("");

  React.useEffect(() => {
    if (!organizationName) {
      setFooterHtml("");
      return;
    }

    let cancelled = false;
    OrganizationBackend.getDefaultApplication("admin", organizationName)
      .then((res: any) => {
        if (!cancelled) {
          setFooterHtml(res.status === "ok" ? res.data?.footerHtml ?? "" : "");
        }
      })
      .catch(() => undefined);
    return () => {
      cancelled = true;
    };
  }, [organizationName]);

  return footerHtml;
}

export function AppLayout() {
  const {account} = useAccount();
  const footerHtml = useFooterHtml(account?.organization?.name);
  const palette = useCommandPalette();
  // the console follows the signed-in user's organization theme, title and favicon, or on
  // a demo site the Marketplace theme on show
  const showcase = useShowcaseApplication(null);
  useThemeData(showcase?.themeData ?? Setting.getThemeData(account?.organization, null));
  const showcaseCss = React.useMemo(() => getConsoleCss(showcase?.formCss), [showcase]);

  return (
    <div className="flex h-screen flex-col overflow-hidden bg-background">
      <GithubCorner />
      {showcaseCss ? <style>{showcaseCss}</style> : null}
      {/* the console's own navigation is on the left */}
      <ThemeShowcaseBar side="right" />
      <ConsoleTour />
      <EnableMfaNotification />
      <CommandPalette open={palette.open} onOpenChange={palette.setOpen} />

      {/* the provider owns the rail's open state, its ⌘B shortcut and the mobile
          sheet; it is sized to the viewport here rather than min-h-svh so the
          footer below keeps its place */}
      <SidebarProvider defaultOpen={readSidebarCookie()} className="min-h-0 flex-1">
        <AppSidebar />
        <SidebarInset className="min-h-0 overflow-hidden">
          <Header onOpenPalette={() => palette.setOpen(true)} />
          <main className="min-h-0 flex-1 overflow-y-auto scrollbar-thin">
            <div className="mx-auto w-full max-w-[1600px] p-4 md:p-6">
              {/* keeps a lazy page's suspension from tearing down the console chrome */}
              <React.Suspense fallback={<Loading />}>
                <Outlet />
              </React.Suspense>
            </div>
          </main>
        </SidebarInset>
      </SidebarProvider>

      {/* outside the scroll area and the full width of the window, so it stays at
          the bottom of the viewport the way antd's Layout.Footer does */}
      {footerHtml ? (
        // never trusted here: the console is where admins work, so no script runs in it
        <footer
          id="footer"
          className="shrink-0 border-t bg-background py-3 text-center text-xs text-muted-foreground"
          dangerouslySetInnerHTML={{__html: getSafeHtml(footerHtml, false)}}
        />
      ) : (
        <footer
          id="footer"
          className="shrink-0 border-t bg-background py-3 text-center text-xs text-muted-foreground"
        >
          <PoweredBy />
        </footer>
      )}
    </div>
  );
}
