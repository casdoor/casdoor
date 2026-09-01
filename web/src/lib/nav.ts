import i18next from "i18next";
import {
  Bot,
  DollarSign,
  Home,
  LayoutGrid,
  Lock,
  Settings,
  ShieldCheck,
  Wallet,
  type LucideIcon,
} from "lucide-react";
import * as Conf from "@/Conf";
import * as Setting from "@/lib/setting";
import type {Account} from "@/hooks/use-account";

export interface NavItem {
  key: string;
  label: string;
  /** rendered as an external link instead of a router link */
  href?: string;
}

export interface NavGroup {
  key: string;
  label: string;
  icon: LucideIcon;
  /** where the group header itself navigates to */
  to: string;
  items: NavItem[];
}

/** The console navigation, mirroring the antd ManagementPage menu. */
export function getNavGroups(account: Account | null | undefined): NavGroup[] {
  if (!account) {
    return [];
  }

  const groups: NavGroup[] = [];

  groups.push({
    key: "/home",
    label: i18next.t("general:Home"),
    icon: Home,
    to: "/",
    items: [
      {key: "/", label: i18next.t("general:Dashboard")},
      {key: "/shortcuts", label: i18next.t("general:Shortcuts")},
      {key: "/apps", label: i18next.t("general:Apps")},
    ],
  });

  groups.push({
    key: "/orgs",
    label: i18next.t("general:User Management"),
    icon: LayoutGrid,
    to: "/organizations",
    items: [
      {key: "/organizations", label: i18next.t("general:Organizations")},
      {key: "/groups", label: i18next.t("general:Groups")},
      {key: "/users", label: i18next.t("general:Users")},
      {key: "/invitations", label: i18next.t("general:Invitations")},
    ],
  });

  groups.push({
    key: "/identity",
    label: i18next.t("general:Identity"),
    icon: Lock,
    to: "/applications",
    items: [
      {key: "/applications", label: i18next.t("general:Applications")},
      {key: "/providers", label: i18next.t("application:Providers")},
      {key: "/resources", label: i18next.t("general:Resources")},
      {key: "/certs", label: i18next.t("general:Certs")},
      {key: "/keys", label: i18next.t("general:Keys")},
    ],
  });

  groups.push({
    key: "/auth",
    label: i18next.t("general:Authorization"),
    icon: ShieldCheck,
    to: "/roles",
    items: [
      {key: "/roles", label: i18next.t("general:Roles")},
      {key: "/permissions", label: i18next.t("general:Permissions")},
      {key: "/models", label: i18next.t("general:Models")},
      {key: "/adapters", label: i18next.t("general:Adapters")},
      {key: "/enforcers", label: i18next.t("general:Enforcers")},
    ].filter(
      (item) =>
        Setting.isLocalAdminUser(account) || !["/models", "/adapters", "/enforcers"].includes(item.key),
    ),
  });

  groups.push({
    key: "/gateway",
    label: i18next.t("general:LLM AI"),
    icon: Bot,
    to: "/sites",
    items: [
      {key: "/agents", label: i18next.t("general:Agents")},
      {key: "/servers", label: i18next.t("general:MCP Servers")},
      {key: "/server-store", label: i18next.t("general:MCP Store")},
      {key: "/entries", label: i18next.t("general:Entries")},
      {key: "/sites", label: i18next.t("general:Sites")},
      {key: "/rules", label: i18next.t("general:Rules")},
    ],
  });

  groups.push({
    key: "/logs",
    label: i18next.t("general:Auditing"),
    icon: Wallet,
    to: "/sessions",
    items: [
      {key: "/sessions", label: i18next.t("general:Sessions")},
      {key: "/records", label: i18next.t("general:Records")},
      {key: "/tokens", label: i18next.t("general:Tokens")},
      {key: "/verifications", label: i18next.t("general:Verifications")},
    ],
  });

  groups.push({
    key: "/business",
    label: i18next.t("general:Business"),
    icon: DollarSign,
    to: "/products",
    items: [
      {key: "/product-store", label: i18next.t("general:Product Store")},
      {key: "/products", label: i18next.t("general:Products")},
      {key: "/coupons", label: i18next.t("general:Coupons")},
      {key: "/cart", label: i18next.t("general:Cart")},
      {key: "/orders", label: i18next.t("general:Orders")},
      {key: "/payments", label: i18next.t("general:Payments")},
      {key: "/plans", label: i18next.t("general:Plans")},
      {key: "/pricings", label: i18next.t("general:Pricings")},
      {key: "/subscriptions", label: i18next.t("general:Subscriptions")},
      {key: "/transactions", label: i18next.t("general:Transactions")},
    ],
  });

  const adminItems: NavItem[] = [];
  if (Setting.isAdminUser(account)) {
    adminItems.push({key: "/sysinfo", label: i18next.t("general:System Info")});
  }
  adminItems.push(
    {key: "/forms", label: i18next.t("general:Forms")},
    {key: "/syncers", label: i18next.t("general:Syncers")},
    {key: "/webhooks", label: i18next.t("general:Webhooks")},
    {key: "/webhook-events", label: i18next.t("general:Webhook Events")},
    {key: "/tickets", label: i18next.t("general:Tickets")},
  );
  if (Setting.isAdminUser(account)) {
    adminItems.push({
      key: "/swagger",
      label: i18next.t("general:Swagger"),
      href: Setting.isLocalhost() ? `${Setting.getFullServerUrl()}/swagger` : "/swagger",
    });
  }

  groups.push({
    key: "/admin",
    label: i18next.t("general:Admin"),
    icon: Settings,
    to: Setting.isAdminUser(account) ? "/sysinfo" : "/syncers",
    items: adminItems,
  });

  return applyNavItems(groups, account);
}

/**
 * The organization can hide navbar entries: `navItems` for admins,
 * `userNavItems` for everyone else. Either being absent, not an array, or
 * containing "all" means "show everything" — the value Casdoor stores by default.
 */
function getNavItemFilter(account: Account): string[] | null {
  const organization = account.organization;
  const navItems = Setting.isLocalAdminUser(account) ? organization?.navItems : organization?.userNavItems ?? [];
  if (!Array.isArray(navItems) || navItems.includes("all")) {
    return null;
  }
  return navItems;
}

function applyNavItems(groups: NavGroup[], account: Account): NavGroup[] {
  const navItems = getNavItemFilter(account);
  if (navItems === null) {
    return groups;
  }

  return groups
    .map((group) => ({...group, items: group.items.filter((item) => navItems.includes(item.key))}))
    .filter((group) => group.items.length > 0)
    .map((group) => {
      // the group header links somewhere that may have just been filtered out;
      // re-point it at a remaining entry, skipping the external-link ones
      if (group.items.some((item) => item.key === group.to && !item.href)) {
        return group;
      }
      const target = group.items.find((item) => !item.href);
      return target ? {...group, to: target.key} : group;
    });
}

/**
 * With only a handful of entries left after filtering, the antd frontend drops
 * the group headers and shows one flat list. Mirrored here so a trimmed-down
 * organization gets the same shape of menu.
 */
export function shouldFlattenNav(groups: NavGroup[], account: Account | null | undefined): boolean {
  if (!account || getNavItemFilter(account) === null) {
    return false;
  }
  return groups.reduce((count, group) => count + group.items.length, 0) <= Conf.MaxItemsForFlatMenu;
}

/**
 * The header widgets an organization can hide through `widgetItems`; same
 * "absent or contains all" rule as the navbar entries.
 */
export function isWidgetVisible(account: Account | null | undefined, key: string): boolean {
  const widgetItems = account?.organization?.widgetItems;
  if (!Array.isArray(widgetItems) || widgetItems.includes("all")) {
    return true;
  }
  return widgetItems.includes(key);
}
