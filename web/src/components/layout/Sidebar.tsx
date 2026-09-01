import * as React from "react";
import {Link, useLocation} from "react-router-dom";
import {ChevronDown} from "lucide-react";
import {Collapsible, CollapsibleContent, CollapsibleTrigger} from "@/components/ui/collapsible";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
  SidebarRail,
  useSidebar,
} from "@/components/ui/sidebar";
import {getNavGroups, shouldFlattenNav, type NavGroup, type NavItem} from "@/lib/nav";
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

function writeOpenGroups(keys: string[]) {
  try {
    localStorage.setItem(OPEN_GROUPS_KEY, JSON.stringify(keys));
  } catch {
    // a blocked store just means the groups reopen on the active one next visit
  }
}

/**
 * An external nav entry opens in a new tab; everything else is a router link.
 *
 * It forwards its ref and spreads what it is given, because every caller renders
 * it under an `asChild` slot — Radix merges the menu item's own class names and
 * handlers onto this element, and a plain function component would swallow them.
 */
const NavLink = React.forwardRef<
  HTMLAnchorElement,
  {item: NavItem; onNavigate?: () => void} & React.ComponentPropsWithoutRef<"a">
    >(({item, onNavigate, onClick, ...props}, ref) => {
      const handleClick = (event: React.MouseEvent<HTMLAnchorElement>) => {
        onClick?.(event);
        onNavigate?.();
      };

      if (item.href) {
        return (
          <a ref={ref} href={item.href} target="_blank" rel="noreferrer" onClick={handleClick} {...props}>
            {item.label}
          </a>
        );
      }
      return (
        <Link ref={ref} to={item.key} onClick={handleClick} {...props}>
          {item.label}
        </Link>
      );
    });
NavLink.displayName = "NavLink";

/**
 * A group rendered as an icon while the rail is collapsed. Collapsing used to
 * drop the second level entirely and leave the group icon pointing at one
 * arbitrary page; the menu keeps every destination reachable at 3rem wide.
 */
function CollapsedGroup({
  group,
  active,
  isItemActive,
  onNavigate,
}: {
  group: NavGroup;
  active: boolean;
  isItemActive: (key: string) => boolean;
  onNavigate?: () => void;
}) {
  const Icon = group.icon;
  return (
    <SidebarMenuItem>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <SidebarMenuButton tooltip={group.label} isActive={active}>
            <Icon />
            <span>{group.label}</span>
          </SidebarMenuButton>
        </DropdownMenuTrigger>
        <DropdownMenuContent side="right" align="start" className="w-52">
          <DropdownMenuLabel>{group.label}</DropdownMenuLabel>
          <DropdownMenuSeparator />
          {group.items.map((item) => (
            <DropdownMenuItem key={item.key} asChild className={cn(isItemActive(item.key) && "font-medium")}>
              <NavLink item={item} onNavigate={onNavigate} />
            </DropdownMenuItem>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>
    </SidebarMenuItem>
  );
}

function ExpandedGroup({
  group,
  open,
  onOpenChange,
  active,
  isItemActive,
  onNavigate,
}: {
  group: NavGroup;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  active: boolean;
  isItemActive: (key: string) => boolean;
  onNavigate?: () => void;
}) {
  const Icon = group.icon;
  // a shut group still has to show that something inside it is the page you are on
  const hasActiveChild = !open && group.items.some((item) => isItemActive(item.key));

  return (
    <Collapsible asChild open={open} onOpenChange={onOpenChange} className="group/collapsible">
      <SidebarMenuItem>
        <CollapsibleTrigger asChild>
          <SidebarMenuButton tooltip={group.label} isActive={hasActiveChild} className={cn(active && "font-medium")}>
            <Icon />
            <span>{group.label}</span>
            <ChevronDown className="ml-auto opacity-50 transition-transform group-data-[state=open]/collapsible:rotate-180" />
          </SidebarMenuButton>
        </CollapsibleTrigger>
        <CollapsibleContent>
          <SidebarMenuSub>
            {group.items.map((item) => (
              <SidebarMenuSubItem key={item.key}>
                <SidebarMenuSubButton asChild isActive={isItemActive(item.key)}>
                  <NavLink item={item} onNavigate={onNavigate} />
                </SidebarMenuSubButton>
              </SidebarMenuSubItem>
            ))}
          </SidebarMenuSub>
        </CollapsibleContent>
      </SidebarMenuItem>
    </Collapsible>
  );
}

export function AppSidebar({onNavigate}: {onNavigate?: () => void}) {
  const {account} = useAccount();
  const location = useLocation();
  const isDark = useIsDark();
  const {state, isMobile} = useSidebar();
  // the rail only shrinks to icons on a desktop; the mobile sheet is always full width
  const collapsed = state === "collapsed" && !isMobile;

  const organization = account?.organization;
  // the organization brands its own console, and may brand light and dark apart
  const siderLogo = Setting.getThemedLogo(organization?.logo, organization?.logoDark, [isDark ? "dark" : "light"]);
  // recomputed every render on purpose: the labels come from i18next, which
  // changes language without re-rendering this component's inputs
  const groups = getNavGroups(account);
  // an organization trimmed down to a handful of entries gets a flat list
  const flatten = shouldFlattenNav(groups, account);

  const isItemActive = (key: string) =>
    key === "/" ? location.pathname === "/" : location.pathname === key || location.pathname.startsWith(key + "/");

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
      setOpenGroups((previous) => {
        const next = [...previous, activeGroupKey];
        writeOpenGroups(next);
        return next;
      });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [activeGroupKey]);

  const setGroupOpen = (key: string, open: boolean) => {
    setOpenGroups((previous) => {
      const next = open ? [...new Set([...previous, key])] : previous.filter((k) => k !== key);
      writeOpenGroups(next);
      return next;
    });
  };

  return (
    <Sidebar collapsible="icon">
      <SidebarHeader className="h-14 justify-center border-b">
        <Link to="/" onClick={onNavigate} className="flex items-center overflow-hidden px-1">
          {/* collapsed the organization's favicon fits where its wordmark does not */}
          <img
            src={collapsed ? organization?.favicon || siderLogo : siderLogo}
            alt={organization?.displayName || "Casdoor"}
            className={cn("object-contain", collapsed ? "h-6 w-6 rounded" : "h-8 max-w-[160px]")}
          />
        </Link>
      </SidebarHeader>

      <SidebarContent className="scrollbar-thin">
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              {flatten && !collapsed
                ? groups.flatMap((group) =>
                  group.items.map((item) => (
                    <SidebarMenuItem key={item.key}>
                      <SidebarMenuButton asChild isActive={isItemActive(item.key)}>
                        <NavLink item={item} onNavigate={onNavigate} />
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                  )),
                )
                : groups.map((group) =>
                  collapsed ? (
                    <CollapsedGroup
                      key={group.key}
                      group={group}
                      active={activeGroupKey === group.key}
                      isItemActive={isItemActive}
                      onNavigate={onNavigate}
                    />
                  ) : (
                    <ExpandedGroup
                      key={group.key}
                      group={group}
                      open={openGroups.includes(group.key)}
                      onOpenChange={(open) => setGroupOpen(group.key, open)}
                      active={activeGroupKey === group.key}
                      isItemActive={isItemActive}
                      onNavigate={onNavigate}
                    />
                  ),
                )}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      {/* the thin strip along the border: click or drag it to collapse the rail */}
      <SidebarRail />
    </Sidebar>
  );
}
