import i18next from "i18next";
import {CheckboxTree, type TreeNode} from "@/components/common/CheckboxTree";

/**
 * The console navigation grouped the way the sidebar groups it, so an
 * organization can hide a whole section in one click. Ported from
 * web/src/common/NavItemTree.js — the keys are the stored `navItems` values and
 * must not change.
 */
function getNavItemNodes(): TreeNode[] {
  return [
    {
      key: "all",
      title: i18next.t("general:All"),
      children: [
        {
          key: "/home-top",
          title: i18next.t("general:Home"),
          children: [
            {key: "/", title: i18next.t("general:Dashboard")},
            {key: "/shortcuts", title: i18next.t("general:Shortcuts")},
            {key: "/apps", title: i18next.t("general:Apps")},
          ],
        },
        {
          key: "/orgs-top",
          title: i18next.t("general:User Management"),
          children: [
            {key: "/organizations", title: i18next.t("general:Organizations")},
            {key: "/groups", title: i18next.t("general:Groups")},
            {key: "/users", title: i18next.t("general:Users")},
            {key: "/invitations", title: i18next.t("general:Invitations")},
          ],
        },
        {
          key: "/applications-top",
          title: i18next.t("general:Identity"),
          children: [
            {key: "/applications", title: i18next.t("general:Applications")},
            {key: "/providers", title: i18next.t("application:Providers")},
            {key: "/resources", title: i18next.t("general:Resources")},
            {key: "/certs", title: i18next.t("general:Certs")},
            {key: "/keys", title: i18next.t("general:Keys")},
          ],
        },
        {
          key: "/sites-top",
          title: i18next.t("general:LLM AI"),
          children: [
            {key: "/agents", title: i18next.t("general:Agents")},
            {key: "/servers", title: i18next.t("general:MCP Servers")},
            {key: "/server-store", title: i18next.t("general:MCP Store")},
            {key: "/entries", title: i18next.t("general:Entries")},
            {key: "/sites", title: i18next.t("general:Sites")},
            {key: "/rules", title: i18next.t("general:Rules")},
          ],
        },
        {
          key: "/roles-top",
          title: i18next.t("general:Authorization"),
          children: [
            {key: "/roles", title: i18next.t("general:Roles")},
            {key: "/permissions", title: i18next.t("general:Permissions")},
            {key: "/models", title: i18next.t("general:Models")},
            {key: "/adapters", title: i18next.t("general:Adapters")},
            {key: "/enforcers", title: i18next.t("general:Enforcers")},
          ],
        },
        {
          key: "/sessions-top",
          title: i18next.t("general:Auditing"),
          children: [
            {key: "/sessions", title: i18next.t("general:Sessions")},
            {key: "/records", title: i18next.t("general:Records")},
            {key: "/tokens", title: i18next.t("general:Tokens")},
            {key: "/verifications", title: i18next.t("general:Verifications")},
          ],
        },
        {
          key: "/business-top",
          title: i18next.t("general:Business"),
          children: [
            {key: "/product-store", title: i18next.t("general:Product Store")},
            {key: "/products", title: i18next.t("general:Products")},
            {key: "/coupons", title: i18next.t("general:Coupons")},
            {key: "/cart", title: i18next.t("general:Cart")},
            {key: "/orders", title: i18next.t("general:Orders")},
            {key: "/payments", title: i18next.t("general:Payments")},
            {key: "/plans", title: i18next.t("general:Plans")},
            {key: "/pricings", title: i18next.t("general:Pricings")},
            {key: "/subscriptions", title: i18next.t("general:Subscriptions")},
            {key: "/transactions", title: i18next.t("general:Transactions")},
          ],
        },
        {
          key: "/admin-top",
          title: i18next.t("general:Admin"),
          children: [
            {key: "/sysinfo", title: i18next.t("general:System Info")},
            {key: "/forms", title: i18next.t("general:Forms")},
            {key: "/syncers", title: i18next.t("general:Syncers")},
            {key: "/webhooks", title: i18next.t("general:Webhooks")},
            {key: "/webhook-events", title: i18next.t("general:Webhook Events")},
            {key: "/tickets", title: i18next.t("general:Tickets")},
            {key: "/swagger", title: i18next.t("general:Swagger")},
          ],
        },
      ],
    },
  ];
}

/** the header buttons an organization can hide, port of web/src/common/WidgetItemTree.js */
function getWidgetItemNodes(): TreeNode[] {
  return [
    {
      key: "all",
      title: i18next.t("general:All"),
      children: [
        {key: "tour", title: i18next.t("general:Tour")},
        {key: "language", title: i18next.t("user:Language")},
        {key: "theme", title: i18next.t("theme:Theme")},
      ],
    },
  ];
}

interface ItemTreeProps {
  checkedKeys: string[];
  onCheck: (checkedKeys: string[]) => void;
  disabled?: boolean;
}

export function NavItemTree({checkedKeys, onCheck, disabled}: ItemTreeProps) {
  return (
    <CheckboxTree
      nodes={getNavItemNodes()}
      checkedKeys={checkedKeys}
      onCheck={onCheck}
      disabled={disabled}
      defaultExpandedKeys={["all"]}
      className="max-h-96 overflow-auto"
    />
  );
}

export function WidgetItemTree({checkedKeys, onCheck, disabled}: ItemTreeProps) {
  return (
    <CheckboxTree
      nodes={getWidgetItemNodes()}
      checkedKeys={checkedKeys}
      onCheck={onCheck}
      disabled={disabled}
      defaultExpandedKeys={["all"]}
    />
  );
}
