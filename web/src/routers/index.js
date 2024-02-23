import Dashboard from "../basic/Dashboard";
import AppListPage from "../basic/AppListPage";
import ShortcutsPage from "../basic/ShortcutsPage";
import AccountPage from "../account/AccountPage";
import OrganizationListPage from "../OrganizationListPage";
import OrganizationEditPage from "../OrganizationEditPage";
import UserListPage from "../UserListPage";
import GroupTreePage from "../GroupTreePage";
import GroupList from "../GroupList";
import GroupEdit from "../GroupEdit";
import UserEditPage from "../UserEditPage";
import InvitationListPage from "../InvitationListPage";
import InvitationEditPage from "../InvitationEditPage";
import ApplicationListPage from "../ApplicationListPage";
import ApplicationEditPage from "../ApplicationEditPage";
import ProviderListPage from "../ProviderListPage";
import ProviderEditPage from "../ProviderEditPage";
import ResourceListPage from "../ResourceListPage";
import CertListPage from "../CertListPage";
import CertEditPage from "../CertEditPage";
import RoleListPage from "../RoleListPage";
import RoleEditPage from "../RoleEditPage";
import PermissionListPage from "../PermissionListPage";
import PermissionEditPage from "../PermissionEditPage";
import ModelListPage from "../ModelListPage";
import ModelEditPage from "../ModelEditPage";
import AdapterListPage from "../AdapterListPage";
import AdapterEditPage from "../AdapterEditPage";
import EnforcerListPage from "../EnforcerListPage";
import EnforcerEditPage from "../EnforcerEditPage";
import SessionListPage from "../SessionListPage";
import TokenListPage from "../TokenListPage";
import TokenEditPage from "../TokenEditPage";
import ProductListPage from "../ProductListPage";
import ProductEditPage from "../ProductEditPage";
import ProductBuyPage from "../ProductBuyPage";
import RecordListPage from "../RecordListPage";
import PaymentListPage from "../PaymentListPage";
import PaymentEditPage from "../PaymentEditPage";
import PaymentResultPage from "../PaymentResultPage";
import PlanListPage from "../PlanListPage";
import PlanEditPage from "../PlanEditPage";
import PricingListPage from "../PricingListPage";
import PricingEditPage from "../PricingEditPage";
import SubscriptionListPage from "../SubscriptionListPage";
import SubscriptionEditPage from "../SubscriptionEditPage";
import SystemInfo from "../SystemInfo";
import SyncerListPage from "../SyncerListPage";
import SyncerEditPage from "../SyncerEditPage";
import WebhookListPage from "../WebhookListPage";
import WebhookEditPage from "../WebhookEditPage";
import LdapEditPage from "../LdapEditPage";
import LdapSyncPage from "../LdapSyncPage";
import MfaSetupPage from "../auth/MfaSetupPage";
import OidcDiscoveryPage from "../auth/OidcDiscoveryPage";

const indexRouters = [
  {
    path: "/",
    exact: true,
    component: Dashboard,
    auth: true,
  },
  {
    path: "/apps",
    exact: true,
    component: AppListPage,
    auth: true,

  },
  {
    path: "/shortcuts",
    exact: true,
    component: ShortcutsPage,
    auth: true,

  },
  {
    path: "/account",
    exact: true,
    component: AccountPage,
    auth: true,

  },
  {
    path: "/organizations",
    exact: true,
    component: OrganizationListPage,
    auth: true,

  },
  {
    path: "/organizations/:organizationName",
    exact: true,
    component: OrganizationEditPage,
    auth: true,

  },
  {
    path: "/organizations/:organizationName/users",
    exact: true,
    component: UserListPage,
    auth: true,
  },
  {
    path: "/trees/:organizationName",
    exact: true,
    component: GroupTreePage,
    auth: true,
  },
  {
    path: "/trees/:organizationName/:groupName",
    exact: true,
    component: GroupTreePage,
    auth: true,
  },
  {
    path: "/groups",
    exact: true,
    component: GroupList,
    auth: true,
  },
  {
    path: "/groups/:organizationName/:groupName",
    exact: true,
    component: GroupEdit,
    auth: true,
  },
  {
    path: "/users",
    exact: true,
    component: UserListPage,
    auth: true,
  },
  {
    path: "/users/:organizationName/:userName",
    component: UserEditPage,
    auth: true,
  },
  {
    path: "/invitations",
    exact: true,
    component: InvitationListPage,
    auth: true,
  },
  {
    path: "/invitations/:organizationName/:invitationName",
    component: InvitationEditPage,
    auth: true,
  },
  {
    path: "/applications",
    exact: true,
    component: ApplicationListPage,
    auth: true,
  },
  {
    path: "/applications/:organizationName/:applicationName",
    component: ApplicationEditPage,
    auth: true,
  },
  {
    path: "/providers",
    exact: true,
    component: ProviderListPage,
    auth: true,
  },
  {
    path: "/providers/:organizationName/:providerName",
    component: ProviderEditPage,
    auth: true,
  },
  {
    path: "/resources",
    exact: true,
    component: ResourceListPage,
    auth: true,
  },
  {
    path: "/certs",
    exact: true,
    component: CertListPage,
    auth: true,
  },
  {
    path: "/certs/:organizationName/:certName",
    component: CertEditPage,
    auth: true,
  },
  {
    path: "/roles",
    exact: true,
    component: RoleListPage,
    auth: true,
  },
  {
    path: "/roles/:organizationName/:roleName",
    component: RoleEditPage,
    auth: true,
  },
  {
    path: "/permissions",
    exact: true,
    component: PermissionListPage,
    auth: true,
  },
  {
    path: "/permissions/:organizationName/:permissionName",
    component: PermissionEditPage,
    auth: true,
  },
  {
    path: "/models",
    exact: true,
    component: ModelListPage,
    auth: true,
  },
  {
    path: "/models/:organizationName/:modelName",
    component: ModelEditPage,
    auth: true,
  },
  {
    path: "/adapters",
    exact: true,
    component: AdapterListPage,
    auth: true,
  },
  {
    path: "/adapters/:organizationName/:adapterName",
    component: AdapterEditPage,
    auth: true,
  },
  {
    path: "/enforcers",
    exact: true,
    component: EnforcerListPage,
    auth: true,
  },
  {
    path: "/enforcers/:organizationName/:enforcerName",
    component: EnforcerEditPage,
    auth: true,
  },
  {
    path: "/sessions",
    exact: true,
    component: SessionListPage,
    auth: true,
  },
  {
    path: "/tokens",
    exact: true,
    component: TokenListPage,
    auth: true,
  },
  {
    path: "/tokens/:tokenName",
    component: TokenEditPage,
    auth: true,
  },
  {
    path: "/products",
    exact: true,
    component: ProductListPage,
    auth: true,
  },
  {
    path: "/products/:organizationName/:productName",
    component: ProductEditPage,
    auth: true,
  },
  {
    path: "/products/:organizationName/:productName/buy",
    component: ProductBuyPage,
    auth: true,
  },
  {
    path: "/records",
    component: RecordListPage,
    auth: true,
  },
  {
    path: "/payments",
    exact: true,
    component: PaymentListPage,
    auth: true,
  },
  {
    path: "/payments/:organizationName/:paymentName",
    component: PaymentEditPage,
    auth: true,
  },
  {
    path: "/payments/:organizationName/:paymentName/result",
    component: PaymentResultPage,
    auth: true,
  },
  {
    path: "/plans",
    exact: true,
    component: PlanListPage,
    auth: true,
  },
  {
    path: "/plans/:organizationName/:planName",
    component: PlanEditPage,
    auth: true,
  },
  {
    path: "/pricings",
    exact: true,
    component: PricingListPage,
    auth: true,
  },
  {
    path: "/pricings/:organizationName/:pricingName",
    component: PricingEditPage,
    auth: true,
  },
  {
    path: "/subscriptions",
    exact: true,
    component: SubscriptionListPage,
    auth: true,
  },
  {
    path: "/subscriptions/:organizationName/:subscriptionName",
    component: SubscriptionEditPage,
    auth: true,
  },
  {
    path: "/sysinfo",
    exact: true,
    component: SystemInfo,
    auth: true,
  },
  {
    path: "/syncers",
    exact: true,
    component: SyncerListPage,
    auth: true,
  },
  {
    path: "/syncers/:syncerName",
    component: SyncerEditPage,
    auth: true,
  },
  {
    path: "/webhooks",
    exact: true,
    component: WebhookListPage,
    auth: true,
  },
  {
    path: "/webhooks/:webhookName",
    component: WebhookEditPage,
    auth: true,
  },
  {
    path: "/ldap/:organizationName/:ldapId",
    component: LdapEditPage,
    auth: true,
  },
  {
    path: "/ldap/sync/:organizationName/:ldapId",
    component: LdapSyncPage,
    auth: true,
  },
  {
    path: "/mfa/setup",
    exact: true,
    component: MfaSetupPage,
    auth: true,
  },
  {
    path: "/.well-known/openid-configuration",
    exact: true,
    component: OidcDiscoveryPage,
  },
];

export default indexRouters;
