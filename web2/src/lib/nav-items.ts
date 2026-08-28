// The navbar / widget entries an organization can show or hide, mirroring the
// trees in web/src/common/NavItemTree.js and WidgetItemTree.js. "all" means
// "everything", which is also the value Casdoor stores by default.

export const NavItemKeys = [
  "all",
  "/home-top", "/", "/shortcuts", "/apps",
  "/orgs-top", "/organizations", "/groups", "/users", "/invitations",
  "/applications-top", "/applications", "/providers", "/resources", "/certs", "/keys",
  "/sites-top", "/agents", "/servers", "/server-store", "/entries", "/sites", "/rules",
  "/roles-top", "/roles", "/permissions", "/models", "/adapters", "/enforcers",
  "/sessions-top", "/sessions", "/records", "/tokens", "/verifications",
  "/business-top", "/product-store", "/products", "/coupons", "/cart", "/orders",
  "/payments", "/plans", "/pricings", "/subscriptions", "/transactions",
  "/admin-top", "/sysinfo", "/forms", "/syncers", "/webhooks", "/webhook-events",
  "/tickets", "/swagger",
];

export const WidgetItemKeys = ["all", "theme", "language", "ai-assistant", "tour"];
