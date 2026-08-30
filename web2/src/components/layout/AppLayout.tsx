import * as React from "react";
import {Outlet} from "react-router-dom";
import {Sheet, SheetContent} from "@/components/ui/sheet";
import {ConsoleTour} from "@/components/common/ConsoleTour";
import {GithubCorner} from "@/components/common/GithubCorner";
import {Header} from "@/components/layout/Header";
import {Sidebar} from "@/components/layout/Sidebar";
import * as Conf from "@/Conf";

export function AppLayout() {
  const [collapsed, setCollapsed] = React.useState(() => localStorage.getItem("siderCollapsed") === "true");
  const [mobileOpen, setMobileOpen] = React.useState(false);

  const handleCollapsed = (next: boolean) => {
    setCollapsed(next);
    localStorage.setItem("siderCollapsed", String(next));
  };

  return (
    <div className="flex h-screen overflow-hidden bg-background">
      <GithubCorner />
      <ConsoleTour />
      <Sidebar className="hidden lg:flex" collapsed={collapsed} onCollapsedChange={handleCollapsed} />

      <Sheet open={mobileOpen} onOpenChange={setMobileOpen}>
        <SheetContent side="left" className="w-64 p-0">
          <Sidebar collapsed={false} onCollapsedChange={() => undefined} onNavigate={() => setMobileOpen(false)} />
        </SheetContent>
      </Sheet>

      <div className="flex min-w-0 flex-1 flex-col">
        <Header onOpenMobileNav={() => setMobileOpen(true)} />
        <main className="flex-1 overflow-y-auto scrollbar-thin">
          <div className="mx-auto w-full max-w-[1600px] p-4 md:p-6">
            <Outlet />
          </div>
          <footer className="pb-6 text-center text-xs text-muted-foreground">
            {Conf.CustomFooter ?? (
              <span>
                Powered by{" "}
                <a
                  href="https://casdoor.org"
                  target="_blank"
                  rel="noreferrer"
                  className="underline-offset-4 hover:underline"
                >
                  Casdoor
                </a>
              </span>
            )}
          </footer>
        </main>
      </div>
    </div>
  );
}
