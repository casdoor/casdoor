// Copyright 2021 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import * as React from "react";
import {Navigate, Route, Routes, useLocation, useNavigate} from "react-router-dom";
import {Loading} from "@/components/common/Loading";
import {AppLayout} from "@/components/layout/AppLayout";
import {useAccount} from "@/hooks/use-account";
import * as Auth from "@/auth/Auth";
import * as Conf from "@/Conf";
import * as Setting from "@/lib/setting";

// ---- console pages -----------------------------------------------------------
const Dashboard = React.lazy(() => import("@/pages/Dashboard"));
const NotFoundPage = React.lazy(() => import("@/pages/NotFoundPage"));
const AppListPage = React.lazy(() => import("@/pages/AppListPage"));
const ShortcutsPage = React.lazy(() => import("@/pages/ShortcutsPage"));
const AccountPage = React.lazy(() => import("@/pages/AccountPage"));
const MfaSetupPage = React.lazy(() => import("@/pages/auth/MfaSetupPage"));
const SystemInfoPage = React.lazy(() => import("@/pages/SystemInfoPage"));

const OrganizationListPage = React.lazy(() => import("@/pages/OrganizationListPage"));
const OrganizationEditPage = React.lazy(() => import("@/pages/OrganizationEditPage"));
const UserListPage = React.lazy(() => import("@/pages/UserListPage"));
const UserEditPage = React.lazy(() => import("@/pages/UserEditPage"));
const GroupListPage = React.lazy(() => import("@/pages/GroupListPage"));
const GroupEditPage = React.lazy(() => import("@/pages/GroupEditPage"));
const GroupTreePage = React.lazy(() => import("@/pages/GroupTreePage"));
const InvitationListPage = React.lazy(() => import("@/pages/InvitationListPage"));
const InvitationEditPage = React.lazy(() => import("@/pages/InvitationEditPage"));

const ApplicationListPage = React.lazy(() => import("@/pages/ApplicationListPage"));
const ApplicationEditPage = React.lazy(() => import("@/pages/ApplicationEditPage"));
const ProviderListPage = React.lazy(() => import("@/pages/ProviderListPage"));
const ProviderEditPage = React.lazy(() => import("@/pages/ProviderEditPage"));
const ResourceListPage = React.lazy(() => import("@/pages/ResourceListPage"));
const CertListPage = React.lazy(() => import("@/pages/CertListPage"));
const CertEditPage = React.lazy(() => import("@/pages/CertEditPage"));
const KeyListPage = React.lazy(() => import("@/pages/KeyListPage"));
const KeyEditPage = React.lazy(() => import("@/pages/KeyEditPage"));

const RoleListPage = React.lazy(() => import("@/pages/RoleListPage"));
const RoleEditPage = React.lazy(() => import("@/pages/RoleEditPage"));
const PermissionListPage = React.lazy(() => import("@/pages/PermissionListPage"));
const PermissionEditPage = React.lazy(() => import("@/pages/PermissionEditPage"));
const ModelListPage = React.lazy(() => import("@/pages/ModelListPage"));
const ModelEditPage = React.lazy(() => import("@/pages/ModelEditPage"));
const AdapterListPage = React.lazy(() => import("@/pages/AdapterListPage"));
const AdapterEditPage = React.lazy(() => import("@/pages/AdapterEditPage"));
const EnforcerListPage = React.lazy(() => import("@/pages/EnforcerListPage"));
const EnforcerEditPage = React.lazy(() => import("@/pages/EnforcerEditPage"));

const AgentListPage = React.lazy(() => import("@/pages/AgentListPage"));
const AgentEditPage = React.lazy(() => import("@/pages/AgentEditPage"));
const ServerListPage = React.lazy(() => import("@/pages/ServerListPage"));
const ServerEditPage = React.lazy(() => import("@/pages/ServerEditPage"));
const ServerStorePage = React.lazy(() => import("@/pages/ServerStorePage"));
const EntryListPage = React.lazy(() => import("@/pages/EntryListPage"));
const EntryEditPage = React.lazy(() => import("@/pages/EntryEditPage"));
const SiteListPage = React.lazy(() => import("@/pages/SiteListPage"));
const SiteEditPage = React.lazy(() => import("@/pages/SiteEditPage"));
const RuleListPage = React.lazy(() => import("@/pages/RuleListPage"));
const RuleEditPage = React.lazy(() => import("@/pages/RuleEditPage"));

const SessionListPage = React.lazy(() => import("@/pages/SessionListPage"));
const RecordListPage = React.lazy(() => import("@/pages/RecordListPage"));
const TokenListPage = React.lazy(() => import("@/pages/TokenListPage"));
const TokenEditPage = React.lazy(() => import("@/pages/TokenEditPage"));
const VerificationListPage = React.lazy(() => import("@/pages/VerificationListPage"));

const ProductListPage = React.lazy(() => import("@/pages/ProductListPage"));
const ProductStorePage = React.lazy(() => import("@/pages/ProductStorePage"));
const ProductBuyPage = React.lazy(() => import("@/pages/ProductBuyPage"));
const CartListPage = React.lazy(() => import("@/pages/CartListPage"));
const OrderPayPage = React.lazy(() => import("@/pages/OrderPayPage"));
const PaymentResultPage = React.lazy(() => import("@/pages/PaymentResultPage"));
const ProductEditPage = React.lazy(() => import("@/pages/ProductEditPage"));
const CouponListPage = React.lazy(() => import("@/pages/CouponListPage"));
const CouponEditPage = React.lazy(() => import("@/pages/CouponEditPage"));
const OrderListPage = React.lazy(() => import("@/pages/OrderListPage"));
const OrderEditPage = React.lazy(() => import("@/pages/OrderEditPage"));
const PaymentListPage = React.lazy(() => import("@/pages/PaymentListPage"));
const PaymentEditPage = React.lazy(() => import("@/pages/PaymentEditPage"));
const PlanListPage = React.lazy(() => import("@/pages/PlanListPage"));
const PlanEditPage = React.lazy(() => import("@/pages/PlanEditPage"));
const PricingListPage = React.lazy(() => import("@/pages/PricingListPage"));
const PricingEditPage = React.lazy(() => import("@/pages/PricingEditPage"));
const SubscriptionListPage = React.lazy(() => import("@/pages/SubscriptionListPage"));
const SubscriptionEditPage = React.lazy(() => import("@/pages/SubscriptionEditPage"));
const TransactionListPage = React.lazy(() => import("@/pages/TransactionListPage"));
const TransactionEditPage = React.lazy(() => import("@/pages/TransactionEditPage"));

const FormListPage = React.lazy(() => import("@/pages/FormListPage"));
const FormEditPage = React.lazy(() => import("@/pages/FormEditPage"));
const SyncerListPage = React.lazy(() => import("@/pages/SyncerListPage"));
const SyncerEditPage = React.lazy(() => import("@/pages/SyncerEditPage"));
const WebhookListPage = React.lazy(() => import("@/pages/WebhookListPage"));
const WebhookEditPage = React.lazy(() => import("@/pages/WebhookEditPage"));
const WebhookEventListPage = React.lazy(() => import("@/pages/WebhookEventListPage"));
const TicketListPage = React.lazy(() => import("@/pages/TicketListPage"));
const TicketEditPage = React.lazy(() => import("@/pages/TicketEditPage"));
const LdapEditPage = React.lazy(() => import("@/pages/LdapEditPage"));
const LdapSyncPage = React.lazy(() => import("@/pages/LdapSyncPage"));

// ---- public pages ------------------------------------------------------------
const LoginPage = React.lazy(() => import("@/pages/auth/LoginPage"));
const SignupPage = React.lazy(() => import("@/pages/auth/SignupPage"));
const ForgetPage = React.lazy(() => import("@/pages/auth/ForgetPage"));
const AuthCallback = React.lazy(() => import("@/pages/auth/AuthCallback"));
const ResultPage = React.lazy(() => import("@/pages/auth/ResultPage"));
const ConsentPage = React.lazy(() => import("@/pages/auth/ConsentPage"));
const PricingPage = React.lazy(() => import("@/pages/PricingPage"));
const QrCodePage = React.lazy(() => import("@/pages/QrCodePage"));
const PromptPage = React.lazy(() => import("@/pages/auth/PromptPage"));
const TelegramLogin = React.lazy(() => import("@/pages/auth/TelegramLogin"));
const CasLogout = React.lazy(() => import("@/pages/auth/CasLogout"));
const CaptchaPage = React.lazy(() => import("@/pages/auth/CaptchaPage"));
const OidcDiscoveryPage = React.lazy(() => import("@/pages/auth/OidcDiscoveryPage"));

Setting.initServerUrl();
Setting.initWebConfig();
Auth.initAuthWithConfig({
  serverUrl: Setting.ServerUrl,
  appName: Conf.DefaultApplication,
});

/**
 * When the organization marks an MFA item as "Required" and the user has not set
 * it up yet, Casdoor parks them on the setup wizard right after sign-in.
 */
function useRequiredMfaRedirect() {
  const {account} = useAccount();
  const navigate = useNavigate();
  const location = useLocation();

  React.useEffect(() => {
    if (!account || location.pathname.startsWith("/mfa/setup")) {
      return;
    }
    if (!Setting.isRequiredEnableMfa(account, account.organization)) {
      return;
    }
    const mfaType = Setting.getMfaItemsByRules(account, account.organization, [Setting.MfaRuleRequired]).find(
      (item: any) => item.rule === Setting.MfaRuleRequired,
    )?.name;
    if (mfaType !== undefined) {
      navigate(`/mfa/setup?mfaType=${mfaType}`, {state: {from: "/login"}});
    }
  }, [account, navigate, location.pathname]);
}

/** Sends anonymous visitors to the sign-in page of their organization. */
function RequireAuth({children}: {children: React.ReactNode}) {
  const {account, loading} = useAccount();
  const location = useLocation();

  if (loading || account === undefined) {
    return <Loading className="min-h-screen" />;
  }
  if (account === null) {
    const lastOrg = localStorage.getItem("lastLoginOrg");
    const to = lastOrg && lastOrg !== "built-in" ? `/login/${lastOrg}` : "/login";
    return <Navigate to={to} replace state={{from: location.pathname + location.search}} />;
  }
  return <>{children}</>;
}

export default function App() {
  useRequiredMfaRedirect();

  return (
    <React.Suspense fallback={<Loading className="min-h-screen" />}>
      <Routes>
        {/* Authentication */}
        <Route path="/login" element={<LoginPage type="login" />} />
        <Route path="/login/:owner" element={<LoginPage type="login" />} />
        <Route path="/login/oauth/authorize" element={<LoginPage type="code" />} />
        <Route path="/login/oauth/device/:userCode" element={<LoginPage type="device" />} />
        <Route path="/login/saml/authorize/:owner/:applicationName" element={<LoginPage type="saml" />} />
        <Route path="/cas/:owner/:casApplicationName/login" element={<LoginPage type="cas" />} />
        <Route path="/cas/:owner/:casApplicationName/logout" element={<CasLogout />} />
        <Route path="/signup" element={<SignupPage />} />
        <Route path="/signup/:applicationName" element={<SignupPage />} />
        <Route path="/signup/oauth/authorize" element={<SignupPage />} />
        <Route path="/forget" element={<ForgetPage />} />
        <Route path="/forget/:applicationName" element={<ForgetPage />} />
        <Route path="/callback" element={<AuthCallback />} />
        <Route path="/callback/saml" element={<AuthCallback />} />
        <Route path="/telegram-login" element={<TelegramLogin />} />
        <Route path="/captcha" element={<CaptchaPage />} />
        <Route path="/.well-known/openid-configuration" element={<OidcDiscoveryPage />} />
        <Route path="/consent/:applicationName" element={<ConsentPage />} />
        <Route path="/prompt" element={<PromptPage />} />
        <Route path="/prompt/:applicationName" element={<PromptPage />} />
        <Route path="/result" element={<ResultPage />} />
        <Route path="/result/:applicationName" element={<ResultPage />} />

        {/* Pricing / checkout pages reachable without the console chrome */}
        <Route path="/select-plan/:owner/:pricingName" element={<PricingPage />} />
        <Route path="/buy-plan/:owner/:pricingName" element={<ProductBuyPage />} />
        <Route path="/buy-plan/:owner/:pricingName/result" element={<PaymentResultPage />} />
        <Route path="/qrcode/:owner/:paymentName" element={<QrCodePage />} />

        {/* Console */}
        <Route
          element={
            <RequireAuth>
              <AppLayout />
            </RequireAuth>
          }
        >
          <Route path="/" element={<Dashboard />} />
          <Route path="/apps" element={<AppListPage />} />
          <Route path="/shortcuts" element={<ShortcutsPage />} />
          <Route path="/account" element={<AccountPage />} />
          <Route path="/mfa/setup" element={<MfaSetupPage />} />
          <Route path="/sysinfo" element={<SystemInfoPage />} />

          <Route path="/organizations" element={<OrganizationListPage />} />
          <Route path="/organizations/:organizationName" element={<OrganizationEditPage />} />
          <Route path="/organizations/:organizationName/users" element={<UserListPage />} />
          <Route path="/users" element={<UserListPage />} />
          <Route path="/users/:organizationName/:userName" element={<UserEditPage />} />
          <Route path="/groups" element={<GroupListPage />} />
          <Route path="/groups/:organizationName/:groupName" element={<GroupEditPage />} />
          <Route path="/trees/:organizationName" element={<GroupTreePage />} />
          <Route path="/trees/:organizationName/:groupName" element={<GroupTreePage />} />
          <Route path="/invitations" element={<InvitationListPage />} />
          <Route path="/invitations/:organizationName/:invitationName" element={<InvitationEditPage />} />

          <Route path="/applications" element={<ApplicationListPage />} />
          <Route path="/applications/:organizationName/:applicationName" element={<ApplicationEditPage />} />
          <Route path="/providers" element={<ProviderListPage />} />
          <Route path="/providers/:organizationName/:providerName" element={<ProviderEditPage />} />
          <Route path="/resources" element={<ResourceListPage />} />
          <Route path="/certs" element={<CertListPage />} />
          <Route path="/certs/:organizationName/:certName" element={<CertEditPage />} />
          <Route path="/keys" element={<KeyListPage />} />
          <Route path="/keys/:organizationName/:keyName" element={<KeyEditPage />} />

          <Route path="/roles" element={<RoleListPage />} />
          <Route path="/roles/:organizationName/:roleName" element={<RoleEditPage />} />
          <Route path="/permissions" element={<PermissionListPage />} />
          <Route path="/permissions/:organizationName/:permissionName" element={<PermissionEditPage />} />
          <Route path="/models" element={<ModelListPage />} />
          <Route path="/models/:organizationName/:modelName" element={<ModelEditPage />} />
          <Route path="/adapters" element={<AdapterListPage />} />
          <Route path="/adapters/:organizationName/:adapterName" element={<AdapterEditPage />} />
          <Route path="/enforcers" element={<EnforcerListPage />} />
          <Route path="/enforcers/:organizationName/:enforcerName" element={<EnforcerEditPage />} />

          <Route path="/agents" element={<AgentListPage />} />
          <Route path="/agents/:organizationName/:agentName" element={<AgentEditPage />} />
          <Route path="/servers" element={<ServerListPage />} />
          <Route path="/servers/:organizationName/:serverName" element={<ServerEditPage />} />
          <Route path="/server-store" element={<ServerStorePage />} />
          <Route path="/entries" element={<EntryListPage />} />
          <Route path="/entries/:organizationName/:entryName" element={<EntryEditPage />} />
          <Route path="/sites" element={<SiteListPage />} />
          <Route path="/sites/:organizationName/:siteName" element={<SiteEditPage />} />
          <Route path="/rules" element={<RuleListPage />} />
          <Route path="/rules/:organizationName/:ruleName" element={<RuleEditPage />} />

          <Route path="/sessions" element={<SessionListPage />} />
          <Route path="/records" element={<RecordListPage />} />
          <Route path="/tokens" element={<TokenListPage />} />
          <Route path="/tokens/:tokenName" element={<TokenEditPage />} />
          <Route path="/verifications" element={<VerificationListPage />} />

          <Route path="/product-store" element={<ProductStorePage />} />
          <Route path="/products" element={<ProductListPage />} />
          <Route path="/products/:organizationName/:productName" element={<ProductEditPage />} />
          <Route path="/products/:organizationName/:productName/buy" element={<ProductBuyPage />} />
          <Route path="/cart" element={<CartListPage />} />
          <Route path="/coupons" element={<CouponListPage />} />
          <Route path="/coupons/:organizationName/:couponName" element={<CouponEditPage />} />
          <Route path="/orders" element={<OrderListPage />} />
          <Route path="/orders/:organizationName/:orderName" element={<OrderEditPage />} />
          <Route path="/orders/:organizationName/:orderName/pay" element={<OrderPayPage />} />
          <Route path="/payments" element={<PaymentListPage />} />
          <Route path="/payments/:organizationName/:paymentName" element={<PaymentEditPage />} />
          <Route path="/payments/:organizationName/:paymentName/result" element={<PaymentResultPage />} />
          <Route path="/plans" element={<PlanListPage />} />
          <Route path="/plans/:organizationName/:planName" element={<PlanEditPage />} />
          <Route path="/pricings" element={<PricingListPage />} />
          <Route path="/pricings/:organizationName/:pricingName" element={<PricingEditPage />} />
          <Route path="/subscriptions" element={<SubscriptionListPage />} />
          <Route path="/subscriptions/:organizationName/:subscriptionName" element={<SubscriptionEditPage />} />
          <Route path="/transactions" element={<TransactionListPage />} />
          <Route path="/transactions/:organizationName/:transactionName" element={<TransactionEditPage />} />

          <Route path="/forms" element={<FormListPage />} />
          <Route path="/forms/:formName" element={<FormEditPage />} />
          <Route path="/syncers" element={<SyncerListPage />} />
          <Route path="/syncers/:organizationName/:syncerName" element={<SyncerEditPage />} />
          <Route path="/webhooks" element={<WebhookListPage />} />
          <Route path="/webhooks/:webhookName" element={<WebhookEditPage />} />
          <Route path="/webhook-events" element={<WebhookEventListPage />} />
          <Route path="/tickets" element={<TicketListPage />} />
          <Route path="/tickets/:organizationName/:ticketName" element={<TicketEditPage />} />
          <Route path="/ldap/:organizationName/:ldapId" element={<LdapEditPage />} />
          <Route path="/ldap/sync/:organizationName/:ldapId" element={<LdapSyncPage />} />

          <Route path="*" element={<NotFoundPage />} />
        </Route>
      </Routes>
    </React.Suspense>
  );
}
