import * as React from "react";
import {Outlet} from "react-router-dom";
import {Sheet, SheetContent} from "@/components/ui/sheet";
import {EnableMfaNotification} from "@/components/auth/EnableMfaNotification";
import {ConsoleTour} from "@/components/common/ConsoleTour";
import {GithubCorner} from "@/components/common/GithubCorner";
import {Header} from "@/components/layout/Header";
import {PoweredBy} from "@/components/layout/PoweredBy";
import {Sidebar} from "@/components/layout/Sidebar";
import {useAccount} from "@/hooks/use-account";
import {useAccountHelmet, useThemeData} from "@/hooks/use-application-chrome";
import * as Setting from "@/lib/setting";

export function AppLayout() {
  const [collapsed, setCollapsed] = React.useState(() => localStorage.getItem("siderCollapsed") === "true");
  const [mobileOpen, setMobileOpen] = React.useState(false);
  const {account} = useAccount();
  // the console follows the signed-in user's organization theme, title and favicon
  useThemeData(Setting.getThemeData(account?.organization, null));
  useAccountHelmet(account);

  const handleCollapsed = (next: boolean) => {
    setCollapsed(next);
    localStorage.setItem("siderCollapsed", String(next));
  };

  return (
    <div className="flex h-screen flex-col overflow-hidden bg-background">
      <GithubCorner />
      <ConsoleTour />
      <EnableMfaNotification />

      <div className="flex min-h-0 flex-1">
        <Sidebar className="hidden lg:flex" collapsed={collapsed} onCollapsedChange={handleCollapsed} />

        <Sheet open={mobileOpen} onOpenChange={setMobileOpen}>
          <SheetContent side="left" className="w-64 p-0">
            <Sidebar collapsed={false} onCollapsedChange={() => undefined} onNavigate={() => setMobileOpen(false)} />
          </SheetContent>
        </Sheet>

        <div className="flex min-w-0 flex-1 flex-col">
          <Header onOpenMobileNav={() => setMobileOpen(true)} />
          <main className="min-h-0 flex-1 overflow-y-auto scrollbar-thin">
            <div className="mx-auto w-full max-w-[1600px] p-4 md:p-6">
              <Outlet />
            </div>
          </main>
        </div>
      </div>

      {/* outside the scroll area and the full width of the window, so it stays at
          the bottom of the viewport the way antd's Layout.Footer does */}
      <footer
        id="footer"
        className="shrink-0 border-t bg-background py-3 text-center text-xs text-muted-foreground"
      >
        <PoweredBy />
      </footer>
    </div>
  );
}
