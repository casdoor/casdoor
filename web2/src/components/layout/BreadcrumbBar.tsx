import i18next from "i18next";
import {ChevronRight} from "lucide-react";
import {Link, useLocation} from "react-router-dom";

/**
 * Console breadcrumbs, ported from `web/src/common/BreadcrumbBar.js`. Only the
 * routes listed here get a trail; anything else renders nothing, exactly as the
 * antd version does.
 */
const RESOURCE_LABELS: Record<string, string> = {
  "apps": "general:Apps",
  "shortcuts": "general:Shortcuts",
  "account": "account:My Account",
  "organizations": "general:Organizations",
  "users": "general:Users",
  "groups": "general:Groups",
  "trees": "general:Groups",
  "invitations": "general:Invitations",
  "applications": "general:Applications",
  "providers": "application:Providers",
  "resources": "general:Resources",
  "certs": "general:Certs",
  "keys": "general:Keys",
  "agents": "general:Agents",
  "servers": "general:MCP Servers",
  "server-store": "general:MCP Store",
  "entries": "general:Entries",
  "sites": "general:Sites",
  "rules": "general:Rules",
  "roles": "general:Roles",
  "permissions": "general:Permissions",
  "models": "general:Models",
  "adapters": "general:Adapters",
  "enforcers": "general:Enforcers",
  "sessions": "general:Sessions",
  "records": "general:Records",
  "tokens": "general:Tokens",
  "verifications": "general:Verifications",
  "product-store": "general:Product Store",
  "products": "general:Products",
  "cart": "general:Cart",
  "orders": "general:Orders",
  "payments": "general:Payments",
  "plans": "general:Plans",
  "pricings": "general:Pricings",
  "subscriptions": "general:Subscriptions",
  "transactions": "general:Transactions",
  "sysinfo": "general:System Info",
  "forms": "general:Forms",
  "syncers": "general:Syncers",
  "webhooks": "general:Webhooks",
  "webhook-events": "general:Webhook Events",
  "tickets": "general:Tickets",
  "ldap": "general:LDAP",
  "mfa": "general:MFA",
};

interface Crumb {
  label: string;
  to?: string;
}

export function buildBreadcrumbItems(pathname: string): Crumb[] | null {
  const pathSegments = (pathname || "").split("/").filter(Boolean);
  if (pathSegments.length === 0) {
    return null;
  }

  const rootSegment = pathSegments[0];
  const listLabelKey = RESOURCE_LABELS[rootSegment];
  if (!listLabelKey) {
    return null;
  }

  const home: Crumb = {label: i18next.t("general:Home"), to: "/"};
  if (pathSegments.length === 1) {
    return [home, {label: i18next.t(listLabelKey)}];
  }

  const lastSegment = pathSegments[pathSegments.length - 1];
  const lastLabelKey = RESOURCE_LABELS[lastSegment];
  return [
    home,
    {label: i18next.t(listLabelKey), to: `/${rootSegment}`},
    {label: lastLabelKey ? i18next.t(lastLabelKey) : lastSegment},
  ];
}

export function BreadcrumbBar() {
  const location = useLocation();
  const items = buildBreadcrumbItems(location.pathname);
  if (!items) {
    return null;
  }

  return (
    <nav aria-label="Breadcrumb" className="min-w-0">
      <ol className="flex min-w-0 items-center gap-1 text-sm text-muted-foreground">
        {items.map((item, index) => (
          <li key={`${item.label}-${index}`} className="flex min-w-0 items-center gap-1">
            {index > 0 ? <ChevronRight className="h-3.5 w-3.5 shrink-0 opacity-60" aria-hidden /> : null}
            {item.to ? (
              <Link to={item.to} className="truncate underline-offset-4 hover:text-foreground hover:underline">
                {item.label}
              </Link>
            ) : (
              <span className="truncate font-medium text-foreground" aria-current="page">
                {item.label}
              </span>
            )}
          </li>
        ))}
      </ol>
    </nav>
  );
}

export default BreadcrumbBar;
