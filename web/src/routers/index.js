import {lazy} from "react";

const indexRouters = [
  {
    path: "/",
    exact: true,
    component: lazy(() => import("../basic/Dashboard")),
    auth: true,

  },
  {
    path: "/apps",
    exact: true,
    component: lazy(() => import("../basic/AppListPage")),
    auth: true,

  },
  {
    path: "/shortcuts",
    exact: true,
    component: lazy(() => import("../basic/ShortcutsPage")),
    auth: true,

  },
  {
    path: "/account",
    exact: true,
    component: lazy(() => import("../account/AccountPage")),
    auth: true,

  },
  {
    path: "/organizations",
    exact: true,
    component: lazy(() => import("../OrganizationListPage")),
    auth: true,

  },
  {
    path: "/organizations/:organizationName",
    exact: true,
    component: lazy(() => import("../OrganizationEditPage")),
    auth: true,

  },
  {
    path: "/organizations/:organizationName/users",
    exact: true,
    component: lazy(() => import("../UserListPage")),
    auth: true,
  },
  {
    path: "/trees/:organizationName",
    exact: true,
    component: lazy(() => import("../GroupTreePage")),
    auth: true,
  },
  {
    path: "/trees/:organizationName/:groupName",
    exact: true,
    component: lazy(() => import("../GroupTreePage")),
    auth: true,
  },
  {
    path: "/groups",
    exact: true,
    component: lazy(() => import("../GroupList")),
    auth: true,
  },
  {
    path: "/groups/:organizationName/:groupName",
    exact: true,
    component: lazy(() => import("../GroupEdit")),
    auth: true,
  },
  {
    path: "/users",
    exact: true,
    component: lazy(() => import("../UserListPage")),
    auth: true,
  },
  {
    path: "/users/:organizationName/:userName",
    component: lazy(() => import("../UserEditPage")),
    auth: true,
  },
  {
    path: "/invitations",
    exact: true,
    component: lazy(() => import("../InvitationListPage")),
    auth: true,
  },
  {
    path: "/invitations/:organizationName/:invitationName",
    component: lazy(() => import("../InvitationEditPage")),
    auth: true,
  },
  {
    path: "/applications",
    exact: true,
    component: lazy(() => import("../ApplicationListPage")),
    auth: true,
  },
  {
    path: "/applications/:organizationName/:applicationName",
    component: lazy(() => import("../ApplicationEditPage")),
    auth: true,
  },
  {
    path: "/providers",
    exact: true,
    component: lazy(() => import("../ProviderListPage")),
    auth: true,
  },
  {
    path: "/providers/:organizationName/:providerName",
    component: lazy(() => import("../ProviderEditPage")),
    auth: true,
  },
  {
    path: "/resources",
    exact: true,
    component: lazy(() => import("../ResourceListPage")),
    auth: true,
  },
  {
    path: "/certs",
    exact: true,
    component: lazy(() => import("../CertListPage")),
    auth: true,
  },
  {
    path: "/certs/:organizationName/:certName",
    component: lazy(() => import("../CertEditPage")),
    auth: true,
  },
  {
    path: "/roles",
    exact: true,
    component: lazy(() => import("../RoleListPage")),
    auth: true,
  },
  {
    path: "/roles/:organizationName/:roleName",
    component: lazy(() => import("../RoleEditPage")),
    auth: true,
  },
  {
    path: "/permissions",
    exact: true,
    component: lazy(() => import("../PermissionListPage")),
    auth: true,
  },
  {
    path: "/permissions/:organizationName/:permissionName",
    component: lazy(() => import("../PermissionEditPage")),
    auth: true,
  },
  {
    path: "/models",
    exact: true,
    component: lazy(() => import("../ModelListPage")),
    auth: true,
  },
  {
    path: "/models/:organizationName/:modelName",
    component: lazy(() => import("../ModelEditPage")),
    auth: true,
  },
  {
    path: "/adapters",
    exact: true,
    component: lazy(() => import("../AdapterListPage")),
    auth: true,
  },
  {
    path: "/adapters/:organizationName/:adapterName",
    component: lazy(() => import("../AdapterEditPage")),
    auth: true,
  },
  {
    path: "/enforcers",
    exact: true,
    component: lazy(() => import("../EnforcerListPage")),
    auth: true,
  },
  {
    path: "/enforcers/:organizationName/:enforcerName",
    component: lazy(() => import("../EnforcerEditPage")),
    auth: true,
  },
  {
    path: "/sessions",
    exact: true,
    component: lazy(() => import("../SessionListPage")),
    auth: true,
  },
  {
    path: "/tokens",
    exact: true,
    component: lazy(() => import("../TokenListPage")),
    auth: true,
  },
  {
    path: "/tokens/:tokenName",
    component: lazy(() => import("../TokenEditPage")),
    auth: true,
  },
  {
    path: "/products",
    exact: true,
    component: lazy(() => import("../ProductListPage")),
    auth: true,
  },
  {
    path: "/products/:organizationName/:productName",
    component: lazy(() => import("../ProductEditPage")),
    auth: true,
  },
  {
    path: "/products/:organizationName/:productName/buy",
    component: lazy(() => import("../ProductBuyPage")),
    auth: true,
  },
  {
    path: "/records",
    component: lazy(() => import("../RecordListPage")),
    auth: true,
  },
  {
    path: "/payments",
    exact: true,
    component: lazy(() => import("../PaymentListPage")),
    auth: true,
  },
  {
    path: "/payments/:organizationName/:paymentName",
    component: lazy(() => import("../PaymentEditPage")),
    auth: true,
  },
  {
    path: "/payments/:organizationName/:paymentName/result",
    component: lazy(() => import("../PaymentResultPage")),
    auth: true,
  },
  {
    path: "/plans",
    exact: true,
    component: lazy(() => import("../PlanListPage")),
    auth: true,
  },
  {
    path: "/plans/:organizationName/:planName",
    component: lazy(() => import("../PlanEditPage")),
    auth: true,
  },
  {
    path: "/pricings",
    exact: true,
    component: lazy(() => import("../PricingListPage")),
    auth: true,
  },
  {
    path: "/pricings/:organizationName/:pricingName",
    component: lazy(() => import("../PricingEditPage")),
    auth: true,
  },
  {
    path: "/subscriptions",
    exact: true,
    component: lazy(() => import("../SubscriptionListPage")),
    auth: true,
  },
  {
    path: "/subscriptions/:organizationName/:subscriptionName",
    component: lazy(() => import("../SubscriptionEditPage")),
    auth: true,
  },
  {
    path: "/sysinfo",
    exact: true,
    component: lazy(() => import("../SystemInfo")),
    auth: true,
  },
  {
    path: "/syncers",
    exact: true,
    component: lazy(() => import("../SyncerListPage")),
    auth: true,
  },
  {
    path: "/syncers/:syncerName",
    component: lazy(() => import("../SyncerEditPage")),
    auth: true,
  },
  {
    path: "/webhooks",
    exact: true,
    component: lazy(() => import("../WebhookListPage")),
    auth: true,
  },
  {
    path: "/webhooks/:webhookName",
    component: lazy(() => import("../WebhookEditPage")),
    auth: true,
  },
  {
    path: "/ldap/:organizationName/:ldapId",
    component: lazy(() => import("../LdapEditPage")),
    auth: true,
  },
  {
    path: "/ldap/sync/:organizationName/:ldapId",
    component: lazy(() => import("../LdapSyncPage")),
    auth: true,
  },
  {
    path: "/mfa/setup",
    exact: true,
    component: lazy(() => import("../auth/MfaSetupPage")),
    auth: true,
  },
  {
    path: "/.well-known/openid-configuration",
    exact: true,
    component: lazy(() => import("../auth/OidcDiscoveryPage")),
  },
];

export default indexRouters;
