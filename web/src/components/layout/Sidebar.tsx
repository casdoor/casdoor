import * as React from "react";
import {Link, useLocation} from "react-router-dom";
import {ChevronDown, PanelLeftClose, PanelLeftOpen} from "lucide-react";
import {Button} from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {ScrollArea} from "@/components/ui/scroll-area";
import {Tooltip, TooltipContent, TooltipTrigger} from "@/components/ui/tooltip";
import {getNavGroups, shouldFlattenNav} from "@/lib/nav";
import * as Setting from "@/lib/setting";
import {useAccount} from "@/hooks/use-account";
import {useIsDark} from "@/hooks/use-theme";
import {cn} from "@/lib/utils";

const OPEN_GROUPS_KEY = "casdoor.openNavGroups";

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

/**
 * One navigation row. `nested` ones hang off a group's rule and mark themselves
 * with a bar on that rule, so the active page is visible without having to read
 * the labels.
 */
function NavItem({
  item,
  active,
  onNavigate,
  nested,
}: {
  item: {key: string; label: React.ReactNode; href?: string};
  active: boolean;
  onNavigate?: () => void;
  nested?: boolean;
}) {
  const className = cn(
    "relative flex h-8 items-center truncate rounded-md px-2 text-sm text-sidebar-foreground/80 transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
    active && "bg-sidebar-accent font-medium text-sidebar-accent-foreground",
    active && nested && "before:absolute before:-left-2 before:top-1 before:h-6 before:w-0.5 before:rounded-full before:bg-sidebar-primary",
  );

  if (item.href) {
    return (
      <a href={item.href} target="_blank" rel="noreferrer" className={className}>
        {item.label}
      </a>
    );
  }
  return (
    <Link to={item.key} onClick={onNavigate} className={className}>
      {item.label}
    </Link>
  );
}

export function Sidebar({collapsed, onCollapsedChange, onNavigate, className}: SidebarProps) {
  const {account} = useAccount();
  const location = useLocation();
  const isDark = useIsDark();
  const organization = account?.organization;
  // the organization brands its own console, and may brand light and dark apart
  const siderLogo = Setting.getThemedLogo(organization?.logo, organization?.logoDark, [isDark ? "dark" : "light"]);
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
          {/* collapsed the organization's favicon fits where its wordmark does not */}
          <img
            src={collapsed ? organization?.favicon || siderLogo : siderLogo}
            alt={organization?.displayName || "Casdoor"}
            className={cn("object-contain", collapsed ? "h-7 w-7 rounded" : "h-9 max-w-[160px]")}
          />
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
        <nav className="space-y-0.5 p-2">
          {flatten && !collapsed
            ? groups.flatMap((group) => group.items.map((item) => <NavItem key={item.key} item={item} active={isItemActive(item.key)} onNavigate={onNavigate} />))
            : collapsed
              ? groups.map((group) => {
                const Icon = group.icon;
                return (
                  // collapsing used to drop the second level entirely and leave the
                  // group icon pointing at one arbitrary page; the menu keeps every
                  // destination reachable at 68px wide
                  <DropdownMenu key={group.key}>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <DropdownMenuTrigger asChild>
                          <button
                            type="button"
                            aria-label={group.label}
                            className={cn(
                              "flex h-9 w-full items-center justify-center rounded-md text-sidebar-foreground/80 transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
                              activeGroupKey === group.key && "bg-sidebar-accent text-sidebar-accent-foreground",
                            )}
                          >
                            <Icon className="h-[18px] w-[18px]" />
                          </button>
                        </DropdownMenuTrigger>
                      </TooltipTrigger>
                      <TooltipContent side="right">{group.label}</TooltipContent>
                    </Tooltip>
                    <DropdownMenuContent side="right" align="start" className="w-52">
                      <DropdownMenuLabel>{group.label}</DropdownMenuLabel>
                      <DropdownMenuSeparator />
                      {group.items.map((item) =>
                        item.href ? (
                          <DropdownMenuItem key={item.key} asChild>
                            <a href={item.href} target="_blank" rel="noreferrer">{item.label}</a>
                          </DropdownMenuItem>
                        ) : (
                          <DropdownMenuItem key={item.key} asChild>
                            <Link
                              to={item.key}
                              onClick={onNavigate}
                              className={cn(isItemActive(item.key) && "font-medium")}
                            >
                              {item.label}
                            </Link>
                          </DropdownMenuItem>
                        ),
                      )}
                    </DropdownMenuContent>
                  </DropdownMenu>
                );
              })
              : groups.map((group) => {
                const Icon = group.icon;
                const open = openGroups.includes(group.key);
                // a collapsed group still has to show that something inside it is
                // the page you are on
                const hasActiveChild = !open && group.items.some((item) => isItemActive(item.key));
                return (
                  <div key={group.key}>
                    <button
                      type="button"
                      onClick={() => toggleGroup(group.key)}
                      className={cn(
                        "flex h-8 w-full items-center gap-2 rounded-md px-2 text-sm text-sidebar-foreground/80 transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
                        (activeGroupKey === group.key || hasActiveChild) && "font-medium text-sidebar-accent-foreground",
                        hasActiveChild && "bg-sidebar-accent",
                      )}
                    >
                      <Icon className="h-[18px] w-[18px] shrink-0" />
                      <span className="flex-1 truncate text-left">{group.label}</span>
                      <ChevronDown
                        className={cn("h-3.5 w-3.5 shrink-0 opacity-50 transition-transform", open && "rotate-180")}
                      />
                    </button>
                    {open && (
                      // the rule lines up under the centre of the group icon above it
                      <ul className="ml-[17px] mt-0.5 space-y-0.5 border-l border-sidebar-border pl-2">
                        {group.items.map((item) => (
                          <li key={item.key}>
                            <NavItem item={item} active={isItemActive(item.key)} onNavigate={onNavigate} nested />
                          </li>
                        ))}
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
