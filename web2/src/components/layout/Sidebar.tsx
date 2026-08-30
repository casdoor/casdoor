import * as React from "react";
import {Link, useLocation} from "react-router-dom";
import {ChevronDown, PanelLeftClose, PanelLeftOpen} from "lucide-react";
import {Button} from "@/components/ui/button";
import {ScrollArea} from "@/components/ui/scroll-area";
import {Tooltip, TooltipContent, TooltipTrigger} from "@/components/ui/tooltip";
import {getNavGroups, shouldFlattenNav} from "@/lib/nav";
import * as Setting from "@/lib/setting";
import {useAccount} from "@/hooks/use-account";
import {cn} from "@/lib/utils";

const OPEN_GROUPS_KEY = "web2.openNavGroups";

function readOpenGroups(): string[] | null {
  try {
    const raw = localStorage.getItem(OPEN_GROUPS_KEY);
    return raw ? JSON.parse(raw) : null;
  } catch {
    return null;
  }
}

interface SidebarProps {
  collapsed: boolean;
  onCollapsedChange: (collapsed: boolean) => void;
  onNavigate?: () => void;
  className?: string;
}

export function Sidebar({collapsed, onCollapsedChange, onNavigate, className}: SidebarProps) {
  const {account} = useAccount();
  const location = useLocation();
  // recomputed every render on purpose: the labels come from i18next, which
  // changes language without re-rendering this component's inputs
  const groups = getNavGroups(account);
  // an organization trimmed down to a handful of entries gets a flat list
  const flatten = shouldFlattenNav(groups, account);

  const activeGroupKey = React.useMemo(() => {
    const match = groups.find((group) =>
      group.items.some((item) => item.key !== "/" && location.pathname.startsWith(item.key)),
    );
    if (match) {
      return match.key;
    }
    return location.pathname === "/" ? "/home" : groups[0]?.key;
  }, [groups, location.pathname]);

  const [openGroups, setOpenGroups] = React.useState<string[]>(() => readOpenGroups() ?? []);

  React.useEffect(() => {
    if (activeGroupKey && !openGroups.includes(activeGroupKey)) {
      setOpenGroups((prev) => {
        const next = [...prev, activeGroupKey];
        try {
          localStorage.setItem(OPEN_GROUPS_KEY, JSON.stringify(next));
        } catch {
          // ignore
        }
        return next;
      });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [activeGroupKey]);

  const toggleGroup = (key: string) => {
    setOpenGroups((prev) => {
      const next = prev.includes(key) ? prev.filter((k) => k !== key) : [...prev, key];
      try {
        localStorage.setItem(OPEN_GROUPS_KEY, JSON.stringify(next));
      } catch {
        // ignore
      }
      return next;
    });
  };

  const isItemActive = (key: string) =>
    key === "/" ? location.pathname === "/" : location.pathname === key || location.pathname.startsWith(key + "/");

  return (
    <aside
      className={cn(
        "flex h-full flex-col border-r bg-sidebar text-sidebar-foreground transition-[width] duration-200",
        collapsed ? "w-[68px]" : "w-64",
        className,
      )}
    >
      <div className={cn("flex h-14 shrink-0 items-center gap-2 border-b px-3", collapsed && "justify-center px-0")}>
        <Link to="/" className="flex items-center gap-2 overflow-hidden" onClick={onNavigate}>
          <img src={`${Setting.StaticBaseUrl}/img/casdoor.png`} alt="Casdoor" className="h-7 w-7 rounded" />
          {!collapsed && <span className="truncate text-base font-semibold text-foreground">Casdoor</span>}
        </Link>
        {!collapsed && (
          <Button
            variant="ghost"
            size="iconSm"
            className="ml-auto"
            onClick={() => onCollapsedChange(true)}
            aria-label="Collapse sidebar"
          >
            <PanelLeftClose />
          </Button>
        )}
      </div>

      <ScrollArea className="flex-1">
        <nav className="space-y-1 p-2">
          {flatten && !collapsed
            ? groups.flatMap((group) =>
              group.items.map((item) =>
                item.href ? (
                  <a
                    key={item.key}
                    href={item.href}
                    target="_blank"
                    rel="noreferrer"
                    className="block truncate rounded-md px-2 py-1.5 text-sm hover:bg-sidebar-accent"
                  >
                    {item.label}
                  </a>
                ) : (
                  <Link
                    key={item.key}
                    to={item.key}
                    onClick={onNavigate}
                    className={cn(
                      "block truncate rounded-md px-2 py-1.5 text-sm hover:bg-sidebar-accent",
                      isItemActive(item.key) && "bg-sidebar-accent font-medium text-sidebar-accent-foreground",
                    )}
                  >
                    {item.label}
                  </Link>
                ),
              ),
            )
            : collapsed
              ? groups.map((group) => {
                const Icon = group.icon;
                return (
                  <Tooltip key={group.key}>
                    <TooltipTrigger asChild>
                      <Link
                        to={group.to}
                        onClick={onNavigate}
                        className={cn(
                          "flex h-9 items-center justify-center rounded-md hover:bg-sidebar-accent",
                          activeGroupKey === group.key && "bg-sidebar-accent text-sidebar-accent-foreground",
                        )}
                      >
                        <Icon className="h-4 w-4" />
                      </Link>
                    </TooltipTrigger>
                    <TooltipContent side="right">{group.label}</TooltipContent>
                  </Tooltip>
                );
              })
              : groups.map((group) => {
                const Icon = group.icon;
                const open = openGroups.includes(group.key);
                return (
                  <div key={group.key}>
                    <button
                      type="button"
                      onClick={() => toggleGroup(group.key)}
                      className={cn(
                        "flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm font-medium hover:bg-sidebar-accent",
                        activeGroupKey === group.key && "text-foreground",
                      )}
                    >
                      <Icon className="h-4 w-4 shrink-0" />
                      <span className="flex-1 truncate text-left">{group.label}</span>
                      <ChevronDown className={cn("h-3.5 w-3.5 transition-transform", open && "rotate-180")} />
                    </button>
                    {open && (
                      <ul className="ml-[15px] mt-0.5 space-y-0.5 border-l pl-3">
                        {group.items.map((item) =>
                          item.href ? (
                            <li key={item.key}>
                              <a
                                href={item.href}
                                target="_blank"
                                rel="noreferrer"
                                className="block truncate rounded-md px-2 py-1.5 text-sm hover:bg-sidebar-accent"
                              >
                                {item.label}
                              </a>
                            </li>
                          ) : (
                            <li key={item.key}>
                              <Link
                                to={item.key}
                                onClick={onNavigate}
                                className={cn(
                                  "block truncate rounded-md px-2 py-1.5 text-sm hover:bg-sidebar-accent",
                                  isItemActive(item.key) &&
                                    "bg-sidebar-accent font-medium text-sidebar-accent-foreground",
                                )}
                              >
                                {item.label}
                              </Link>
                            </li>
                          ),
                        )}
                      </ul>
                    )}
                  </div>
                );
              })}
        </nav>
      </ScrollArea>

      {collapsed && (
        <div className="border-t p-2">
          <Button
            variant="ghost"
            size="iconSm"
            className="mx-auto flex"
            onClick={() => onCollapsedChange(false)}
            aria-label="Expand sidebar"
          >
            <PanelLeftOpen />
          </Button>
        </div>
      )}
    </aside>
  );
}
