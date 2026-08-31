// Default objects for newly created records, ported from the `newXxx()` methods
// of the antd list pages so that a record created from the new UI is identical
// to one created from the old one.

import dayjs from "dayjs";
import * as Conf from "@/Conf";
import * as Setting from "@/lib/setting";
import {SignupTableDefaultCssMap} from "@/lib/signup-css";
import type {Account} from "@/hooks/use-account";

const now = () => dayjs().format();

export const rbacModel = `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act`;

export function newAdapter(account: Account) {
  const randomName = Setting.getRandomName();
  const owner = Setting.getRequestOrganization(account);
  return {
    owner,
    name: `adapter_${randomName}`,
    createdTime: now(),
    table: "table_name",
    useSameDb: true,
  };
}

export function newAgent(account: Account) {
  const randomName = Setting.getRandomName();
  const owner = Setting.getRequestOrganization(account);
  return {
    owner,
    name: `agent_${randomName}`,
    createdTime: now(),
    displayName: `New Agent - ${randomName}`,
    url: "",
    token: "",
    application: "",
  };
}

export function newApplication(account: Account) {
  const randomName = Setting.getRandomName();
  const organizationName = Setting.getRequestOrganization(account);
  return {
    owner: "admin",
    name: `application_${randomName}`,
    organization: organizationName,
    createdTime: now(),
    displayName: `New Application - ${randomName}`,
    category: "Default",
    type: "All",
    scopes: [],
    logo: `${Setting.StaticBaseUrl}/img/casdoor-logo_1185x256.png`,
    enablePassword: true,
    enableSignUp: true,
    disableSignin: false,
    enableSigninSession: false,
    enableCodeSignin: false,
    enableSamlCompress: false,
    disableSamlAttributes: false,
    providers: [
      {name: "provider_captcha_default", canSignUp: false, canSignIn: false, canUnlink: false, prompted: false, signupGroup: "", rule: ""},
    ],
    signinMethods: [
      {name: "Password", displayName: "Password", rule: "All"},
      {name: "Verification code", displayName: "Verification code", rule: "All"},
      {name: "WebAuthn", displayName: "WebAuthn", rule: "None"},
      {name: "Face ID", displayName: "Face ID", rule: "None"},
    ],
    signupItems: [
      {name: "ID", visible: false, required: true, rule: "Random"},
      {name: "Username", visible: true, required: true, rule: "None"},
      {name: "Display name", visible: true, required: true, rule: "None"},
      {name: "Password", visible: true, required: true, rule: "None"},
      {name: "Confirm password", visible: true, required: true, rule: "None"},
      {name: "Email", visible: true, required: true, rule: "Normal"},
      {name: "Phone", visible: true, required: true, rule: "None"},
      {name: "Agreement", visible: true, required: true, rule: "None"},
      {name: "Signup button", visible: true, required: true, rule: "None"},
      {name: "Providers", visible: true, required: true, rule: "None", customCss: SignupTableDefaultCssMap["Providers"]},
    ],
    grantTypes: ["authorization_code", "password", "client_credentials", "token", "id_token", "refresh_token"],
    cert: "cert-built-in",
    redirectUris: ["http://localhost:9000/callback"],
    // the empty tokenFormat will be filled with the organization's "defaultTokenFormat" by the backend
    tokenFormat: "",
    tokenFields: [],
    expireInHours: 24 * 7,
    refreshExpireInHours: 24 * 7,
    cookieExpireInHours: 24 * 30,
    formOffset: 2,
  };
}

export function newCert(account: Account, ownerOverride?: string) {
  const randomName = Setting.getRandomName();
  const owner = Setting.isDefaultOrganizationSelected(account) ? ownerOverride ?? "admin" : Setting.getRequestOrganization(account);
  return {
    owner,
    name: `cert_${randomName}`,
    createdTime: now(),
    displayName: `New Cert - ${randomName}`,
    scope: "JWT",
    type: "x509",
    cryptoAlgorithm: "RS256",
    bitSize: 4096,
    expireInYears: 20,
    certificate: "",
    privateKey: "",
  };
}

export function newCoupon(account: Account) {
  const randomName = Setting.getRandomName();
  const owner = Setting.getRequestOrganization(account);
  return {
    owner,
    name: `coupon_${randomName}`,
    createdTime: now(),
    displayName: `New Coupon - ${randomName}`,
    description: "",
    code: `CODE_${randomName}`.toUpperCase(),
    discountType: "percentage",
    discount: 10,
    maxDiscount: 0,
    scope: "universal",
    products: [],
    users: [],
    quantity: 100,
    usedCount: 0,
    maxUsagePerUser: 1,
    startTime: now(),
    expireTime: dayjs().add(30, "days").format(),
    minOrderAmount: 0,
    currency: "USD",
    state: "Active",
  };
}

export function newEnforcer(account: Account) {
  const randomName = Setting.getRandomName();
  const owner = Setting.getRequestOrganization(account);
  return {
    owner,
    name: `enforcer_${randomName}`,
    createdTime: now(),
    displayName: `New Enforcer - ${randomName}`,
  };
}

export function newEntry(account: Account) {
  const randomHex = Math.random().toString(16).slice(2, 18);
  const owner = Setting.getRequestOrganization(account);
  return {
    owner,
    name: randomHex,
    createdTime: now(),
    displayName: randomHex,
    provider: "",
    application: "",
    type: "",
    clientIp: "",
    userAgent: "",
    message: "",
  };
}

export function newForm(account: Account) {
  const randomName = Setting.getRandomName();
  return {
    owner: account.owner,
    name: `form_${randomName}`,
    createdTime: now(),
    displayName: `New Form - ${randomName}`,
    formItems: [],
  };
}

export function newGroup(account: Account) {
  const randomName = Setting.getRandomName();
  const owner = Setting.getRequestOrganization(account);
  return {
    owner,
    name: `group_${randomName}`,
    createdTime: now(),
    updatedTime: now(),
    displayName: `New Group - ${randomName}`,
    type: "Virtual",
    parentId: account.owner,
    isTopGroup: true,
    isEnabled: true,
  };
}

export function newInvitation(account: Account) {
  const randomName = Setting.getRandomName();
  const owner = Setting.getRequestOrganization(account);
  const code = Math.random().toString(36).slice(-10);
  return {
    owner,
    name: `invitation_${randomName}`,
    createdTime: now(),
    updatedTime: now(),
    displayName: `New Invitation - ${randomName}`,
    code,
    defaultCode: code,
    quota: 1,
    usedCount: 0,
    application: "All",
    username: "",
    email: "",
    phone: "",
    signupGroup: "",
    state: "Active",
  };
}

export function newKey(account: Account) {
  const randomName = Setting.getRandomName();
  const owner = Setting.getRequestOrganization(account);
  return {
    owner,
    name: `key_${randomName}`,
    createdTime: now(),
    updatedTime: now(),
    displayName: `New Key - ${randomName}`,
    type: "Organization",
    organization: owner,
    application: "",
    user: "",
    accessKey: "",
    accessSecret: "",
    expireTime: "",
    state: "Active",
  };
}

export function newModel(account: Account) {
  const randomName = Setting.getRandomName();
  const owner = Setting.getRequestOrganization(account);
  return {
    owner,
    name: `model_${randomName}`,
    createdTime: now(),
    displayName: `New Model - ${randomName}`,
    modelText: rbacModel,
  };
}

export function newOrder(account: Account) {
  const randomName = Setting.getRandomName();
  const owner = Setting.getRequestOrganization(account);
  return {
    owner,
    name: `order_${randomName}`,
    createdTime: now(),
    displayName: `New Order - ${randomName}`,
    products: [],
    user: "",
    payment: "",
    state: "Created",
    message: "",
  };
}

export function newPayment(account: Account) {
  const randomName = Setting.getRandomName();
  const organizationName = Setting.getRequestOrganization(account);
  return {
    owner: organizationName,
    name: `payment_${randomName}`,
    createdTime: now(),
    displayName: `New Payment - ${randomName}`,
    provider: "provider_pay_paypal",
    type: "PayPal",
    user: "admin",
    products: [],
    productsDisplayName: "",
    detail: "This is a payment",
    tag: "Promotion-1",
    currency: "USD",
    price: 300.0,
    payUrl: "https://pay.com/pay.php",
    state: "Paid",
    message: "",
  };
}

export function newPermission(account: Account) {
  const randomName = Setting.getRandomName();
  const owner = Setting.getRequestOrganization(account);
  return {
    owner,
    name: `permission_${randomName}`,
    createdTime: now(),
    displayName: `New Permission - ${randomName}`,
    users: [`${account.owner}/${account.name}`],
    groups: [],
    roles: [],
    domains: [],
    resourceType: "Application",
    resources: [Conf.DefaultApplication],
    actions: ["Read"],
    effect: "Allow",
    isEnabled: true,
    submitter: account.name,
    approver: "",
    approveTime: "",
    state: Setting.isLocalAdminUser(account) ? "Approved" : "Pending",
  };
}

export function newPlan(account: Account) {
  const randomName = Setting.getRandomName();
  const owner = Setting.getRequestOrganization(account);
  return {
    owner,
    name: `plan_${randomName}`,
    createdTime: now(),
    displayName: `New Plan - ${randomName}`,
    description: "",
    price: 10,
    currency: "USD",
    period: "Monthly",
    isEnabled: true,
    paymentProviders: [],
    role: "",
    options: [],
  };
}

export function newPricing(account: Account) {
  const randomName = Setting.getRandomName();
  const owner = Setting.getRequestOrganization(account);
  return {
    owner,
    name: `pricing_${randomName}`,
    createdTime: now(),
    plans: [],
    displayName: `New Pricing - ${randomName}`,
    isEnabled: true,
    trialDuration: 7,
  };
}

export function newProduct(account: Account) {
  const randomName = Setting.getRandomName();
  const owner = Setting.getRequestOrganization(account);
  return {
    owner,
    name: `product_${randomName}`,
    createdTime: now(),
    displayName: `New Product - ${randomName}`,
    image: `${Setting.StaticBaseUrl}/img/casdoor-logo_1185x256.png`,
    tag: "Casdoor Summit 2022",
    currency: "USD",
    price: 300,
    quantity: 99,
    sold: 10,
    isRecharge: false,
    providers: [],
    state: "Published",
    properties: {},
  };
}

export function newProvider(account: Account, ownerOverride?: string) {
  const randomName = Setting.getRandomName();
  const owner = Setting.isDefaultOrganizationSelected(account) ? ownerOverride ?? "admin" : Setting.getRequestOrganization(account);
  return {
    owner,
    name: `provider_${randomName}`,
    createdTime: now(),
    displayName: `New Provider - ${randomName}`,
    category: "OAuth",
    type: "GitHub",
    method: "Normal",
    clientId: "",
    clientSecret: "",
    enableSignUp: true,
    host: "",
    port: 0,
    providerUrl: "",
  };
}

export function newRole(account: Account) {
  const randomName = Setting.getRandomName();
  const owner = Setting.getRequestOrganization(account);
  return {
    owner,
    name: `role_${randomName}`,
    createdTime: now(),
    displayName: `New Role - ${randomName}`,
    users: [],
    groups: [],
    roles: [],
    domains: [],
    isEnabled: true,
  };
}

export function newRule(account: Account) {
  const randomName = Setting.getRandomName();
  const owner = Setting.getRequestOrganization(account);
  return {
    owner,
    name: `rule_${randomName}`,
    createdTime: now(),
    type: "User-Agent",
    expressions: [],
    action: "Block",
    reason: "Your request is blocked.",
  };
}

export function newServer(account: Account) {
  const randomName = Setting.getRandomName();
  const owner = Setting.getRequestOrganization(account);
  return {
    owner,
    name: `server_${randomName}`,
    createdTime: now(),
    displayName: `New Server - ${randomName}`,
    url: "",
    application: "",
  };
}

export function newSite(account: Account) {
  const randomName = Setting.getRandomName();
  const owner = Setting.getRequestOrganization(account);
  return {
    owner,
    name: `site_${randomName}`,
    createdTime: now(),
    displayName: `New Site - ${randomName}`,
    domain: "door.casdoor.com",
    otherDomains: [],
    needRedirect: false,
    disableVerbose: false,
    rules: [],
    enableAlert: false,
    alertInterval: 60,
    alertTryTimes: 3,
    alertProviders: [],
    challenges: [],
    host: "",
    port: 8000,
    hosts: [],
    sslMode: "HTTPS Only",
    sslCert: "",
    publicIp: "8.131.81.162",
    node: "",
    isSelf: false,
    nodes: [],
    casdoorApplication: "",
    organizations: [],
  };
}

export function newSubscription(account: Account) {
  const randomName = Setting.getRandomName();
  const owner = Setting.getRequestOrganization(account);
  return {
    owner,
    name: `sub_${randomName}`,
    createdTime: now(),
    displayName: `New Subscription - ${randomName}`,
    startTime: now(),
    endTime: dayjs().add(30, "d").format(),
    period: "Monthly",
    description: "",
    user: "",
    plan: "",
    state: "Active",
  };
}

export function newSyncer(account: Account) {
  const randomName = Setting.getRandomName();
  const organizationName = Setting.getRequestOrganization(account);
  return {
    owner: "admin",
    name: `syncer_${randomName}`,
    createdTime: now(),
    organization: organizationName,
    type: "Database",
    host: "localhost",
    port: 3306,
    user: "root",
    password: "123456",
    databaseType: "mysql",
    database: "dbName",
    table: "table_name",
    tableColumns: [],
    affiliationTable: "",
    avatarBaseUrl: "",
    syncInterval: 10,
    isReadOnly: false,
    isEnabled: false,
  };
}

export function newTicket(account: Account) {
  const randomName = Setting.getRandomName();
  const owner = Setting.getRequestOrganization(account);
  return {
    owner,
    name: `ticket_${randomName}`,
    createdTime: now(),
    updatedTime: now(),
    displayName: `New Ticket - ${randomName}`,
    user: account.name,
    title: "",
    content: "",
    state: "Open",
    messages: [],
  };
}

export function newToken(account: Account) {
  const randomName = Setting.getRandomName();
  const organizationName = Setting.getRequestOrganization(account);
  return {
    owner: "admin",
    name: `token_${randomName}`,
    createdTime: now(),
    application: Conf.DefaultApplication,
    organization: organizationName,
    user: "admin",
    accessToken: "",
    expiresIn: 7200,
    scope: "read",
    tokenType: "Bearer",
  };
}

export function newTransaction(account: Account) {
  const organizationName = Setting.getRequestOrganization(account);
  const randomName = Setting.getRandomName();
  return {
    owner: organizationName,
    // the real name is generated by the server when the transaction is added
    name: `transaction_${randomName}`,
    createdTime: now(),
    application: Conf.DefaultApplication,
    domain: "https://ai-admin.casibase.com",
    category: "",
    type: "chat_id",
    subtype: "message_id",
    provider: "provider_chatgpt",
    user: "admin",
    tag: "AI message",
    amount: 0.1,
    currency: "USD",
    payment: "payment_paypal_001",
    state: "Paid",
  };
}

/**
 * The object the transaction list's "Recharge" button POSTs straight away: a
 * paid top-up for the signed-in user, which the backend names and hands back so
 * the edit page can open it in recharge mode.
 */
export function newRechargeTransaction(account: Account) {
  return {
    owner: Setting.getRequestOrganization(account),
    createdTime: now(),
    application: account.signupApplication || "",
    domain: "",
    category: "Recharge",
    type: "",
    subtype: "",
    provider: "",
    user: account.name || "",
    tag: "User",
    amount: 100,
    currency: "USD",
    payment: "",
    state: "Paid",
  };
}

export function newUser(account: Account, organization: Record<string, any>, organizationName: string, groupName?: string) {
  const randomName = Setting.getRandomName();
  const owner =
    Setting.isDefaultOrganizationSelected(account) || groupName ? organizationName : Setting.getRequestOrganization(account);
  return {
    owner,
    name: `user_${randomName}`,
    createdTime: now(),
    type: "normal-user",
    password: "123",
    passwordSalt: "",
    displayName: `New User - ${randomName}`,
    avatar: organization?.defaultAvatar ?? `${Setting.StaticBaseUrl}/img/casbin.svg`,
    email: `${randomName}@example.com`,
    phone: Setting.getRandomNumber(),
    countryCode: organization?.countryCodes?.length > 0 ? organization.countryCodes[0] : "",
    address: [],
    addresses: [],
    groups: groupName ? [`${owner}/${groupName}`] : [],
    roles: [],
    permissions: [],
    managedAccounts: [],
    mfaAccounts: [],
    affiliation: "Example Inc.",
    tag: "staff",
    region: "",
    realName: "",
    isVerified: false,
    isAdmin: owner === "built-in",
    isForbidden: false,
    score: organization?.initScore,
    isDeleted: false,
    properties: {},
    signupApplication: organization?.defaultApplication,
    registerType: "Add User",
    registerSource: `${account.owner}/${account.name}`,
    balanceCurrency: organization?.balanceCurrency || "USD",
  };
}

export function newWebhook(account: Account) {
  const randomName = Setting.getRandomName();
  const organizationName = Setting.getRequestOrganization(account);
  return {
    owner: "admin",
    name: `webhook_${randomName}`,
    createdTime: now(),
    organization: organizationName,
    url: "https://example.com/callback",
    method: "POST",
    contentType: "application/json",
    headers: [],
    events: ["signup", "login", "logout", "update-user"],
    isEnabled: true,
  };
}
