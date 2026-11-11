import * as ApplicationBackend from "@/backend/ApplicationBackend";
import * as CertBackend from "@/backend/CertBackend";
import * as ModelBackend from "@/backend/ModelBackend";
import * as PermissionBackend from "@/backend/PermissionBackend";
import * as ServerBackend from "@/backend/ServerBackend";
import * as Setting from "@/lib/setting";

export type PaidEdition = "Commercial" | "Enterprise";

/** The kinds of object whose edit page a stop opens: the first one of the current organization, or its list page when there is none. */
export type StopObject = "organization" | "application" | "cert" | "permission" | "model" | "server";

/** Where a feature is in the console, for the feature tour to go to and highlight. */
export interface FeatureStop {
  object?: StopObject;
  /** a fixed page instead of an object, may carry a query */
  path?: string;
  /** the value of the tab to open on the page, and the i18n key of its label */
  tab?: string;
  tabLabel?: string;
  /** the i18n label key of the form row to highlight, see FormRow's data-label-key */
  label?: string;
  /** or a CSS selector of the element to highlight */
  target?: string;
}

export interface EditionFeature {
  id: string;
  edition: PaidEdition;
  /** i18n key of the category */
  category: string;
  /** i18n keys of the name and description */
  name: string;
  description: string;
  stop?: FeatureStop;
  /** set for a feature that is turned on in conf/app.conf instead of the console */
  config?: string;
}

const org = (tab: string, tabLabel: string, label?: string, target?: string): FeatureStop => ({object: "organization", tab, tabLabel, label, target});
const app = (tab: string, tabLabel: string, label: string): FeatureStop => ({object: "application", tab, tabLabel, label});
const compliance = (tab: string, tabLabel: string): FeatureStop => ({path: "/compliance", tab, tabLabel});
const marketplace = (query: string): FeatureStop => ({path: `/marketplace?${query}`, target: "[data-tour='marketplace-items']"});

export const EditionFeatures: EditionFeature[] = [
  // the commercial edition
  {id: "exclusive-signin", edition: "Commercial", category: "edition:Sign-in and sessions", name: "edition:Exclusive sign-in", description: "edition:Exclusive sign-in - Description",
    stop: app("authentication", "application:Authentication", "application:Enable exclusive signin")},
  {id: "inactive-timeout", edition: "Commercial", category: "edition:Sign-in and sessions", name: "edition:Idle timeout", description: "edition:Idle timeout - Description",
    stop: org("security", "application:Security", "organization:Inactive timeout minutes")},
  {id: "disable-inactive", edition: "Commercial", category: "edition:Sign-in and sessions", name: "edition:Disable inactive accounts", description: "edition:Disable inactive accounts - Description",
    stop: org("security", "application:Security", "organization:Disable inactive users after days")},
  {id: "mobile-one-click", edition: "Commercial", category: "edition:Sign-in and sessions", name: "edition:Mobile one-click login", description: "edition:Mobile one-click login - Description",
    stop: app("oidc-oauth", "OIDC/OAuth", "application:Grant types")},
  {id: "push-mfa", edition: "Commercial", category: "edition:Sign-in and sessions", name: "edition:Push MFA", description: "edition:Push MFA - Description",
    stop: org("password", "general:Password type", "general:MFA items")},

  {id: "password-check", edition: "Commercial", category: "edition:Passwords and attacks", name: "edition:Weak and leaked password checks", description: "edition:Weak and leaked password checks - Description",
    stop: org("security", "application:Security", "organization:Password check plugins")},
  {id: "credential-stuffing", edition: "Commercial", category: "edition:Passwords and attacks", name: "edition:Credential stuffing defense", description: "edition:Credential stuffing defense - Description",
    stop: org("security", "application:Security", "organization:Credential stuffing threshold")},
  {id: "signin-alerts", edition: "Commercial", category: "edition:Passwords and attacks", name: "edition:New location and device alerts", description: "edition:New location and device alerts - Description",
    stop: org("security", "application:Security", "organization:Alert on new location")},
  {id: "password-migration", edition: "Commercial", category: "edition:Passwords and attacks", name: "edition:Password migration", description: "edition:Password migration - Description",
    stop: org("security", "application:Security", "organization:Password migration plugin")},

  {id: "masking", edition: "Commercial", category: "edition:Privacy and messaging", name: "edition:Masking in the console", description: "edition:Masking in the console - Description",
    stop: org("security", "application:Security", "organization:Mask personal information in the console")},
  {id: "self-service", edition: "Commercial", category: "edition:Privacy and messaging", name: "edition:Data export and account deletion", description: "edition:Data export and account deletion - Description",
    stop: org("security", "application:Security", "organization:Allow users to export their data")},
  {id: "unionid", edition: "Commercial", category: "edition:Privacy and messaging", name: "edition:WeChat member merging", description: "edition:WeChat member merging - Description",
    stop: org("security", "application:Security", "organization:Merge members by WeChat unionid")},
  {id: "sms-limits", edition: "Commercial", category: "edition:Privacy and messaging", name: "edition:SMS limits", description: "edition:SMS limits - Description",
    stop: org("security", "application:Security", "organization:SMS per number per hour")},

  {id: "key-rotation", edition: "Commercial", category: "edition:Keys and access control", name: "edition:Signing key rotation", description: "edition:Signing key rotation - Description",
    stop: {object: "cert", target: "[data-tour='rotate-key']"}},
  {id: "auto-rotation", edition: "Commercial", category: "edition:Keys and access control", name: "edition:Automatic rotation and expiry alerts", description: "edition:Automatic rotation and expiry alerts - Description",
    stop: org("security", "application:Security", "organization:Rotate signing keys automatically")},
  {id: "rbac", edition: "Commercial", category: "edition:Keys and access control", name: "edition:Role-based access control", description: "edition:Role-based access control - Description",
    stop: {object: "permission", label: "role:Sub roles"}},
  {id: "abac", edition: "Commercial", category: "edition:Keys and access control", name: "edition:Attribute-based access control", description: "edition:Attribute-based access control - Description",
    stop: {object: "model", label: "model:Model text"}},
  {id: "syncers", edition: "Commercial", category: "edition:Keys and access control", name: "edition:More syncers", description: "edition:More syncers - Description",
    stop: {path: "/syncers"}},

  {id: "mcp-gateway", edition: "Commercial", category: "edition:AI agents", name: "edition:MCP gateway", description: "edition:MCP gateway - Description",
    stop: {object: "server", label: "server:Token header"}},
  {id: "mcp-records", edition: "Commercial", category: "edition:AI agents", name: "edition:MCP tool call records", description: "edition:MCP tool call records - Description",
    stop: org("security", "application:Security", "organization:Record MCP tool calls")},
  {id: "mcp-tool-permissions", edition: "Commercial", category: "edition:AI agents", name: "edition:Tool permissions for agents", description: "edition:Tool permissions for agents - Description",
    stop: {object: "server", label: "server:Tool permissions"}},
  {id: "mcp-identity", edition: "Commercial", category: "edition:AI agents", name: "edition:Identity and trace for MCP servers", description: "edition:Identity and trace for MCP servers - Description",
    stop: {object: "server", label: "server:Identity header"}},
  {id: "scim-provisioning", edition: "Commercial", category: "edition:Keys and access control", name: "edition:SCIM provisioning", description: "edition:SCIM provisioning - Description",
    stop: app("provisioning", "application:Provisioning", "application:Enable provisioning")},
  {id: "id-jag", edition: "Commercial", category: "edition:AI agents", name: "edition:Identity assertion grants", description: "edition:Identity assertion grants - Description",
    stop: org("advanced", "provider:Advanced", "organization:Trusted ID-JAG issuers")},

  {id: "security-checks", edition: "Commercial", category: "edition:Operations", name: "edition:Security checks on the dashboard", description: "edition:Security checks on the dashboard - Description",
    stop: {path: "/", target: "[data-tour='security-checks']"}},
  {id: "paid-integrations", edition: "Commercial", category: "edition:Operations", name: "edition:Paid integrations", description: "edition:Paid integrations - Description",
    stop: marketplace("edition=Commercial")},
  {id: "tracing", edition: "Commercial", category: "edition:Operations", name: "edition:OpenTelemetry tracing", description: "edition:OpenTelemetry tracing - Description",
    config: "otelEndpoint"},

  // the enterprise edition, on top of everything in the commercial edition
  {id: "sm2-cert", edition: "Enterprise", category: "edition:Chinese SM cryptography", name: "edition:SM2 certificates", description: "edition:SM2 certificates - Description",
    stop: {object: "cert", label: "cert:Crypto algorithm"}},
  {id: "sm2-token", edition: "Enterprise", category: "edition:Chinese SM cryptography", name: "edition:SM2 token signing", description: "edition:SM2 token signing - Description",
    stop: app("oidc-oauth", "OIDC/OAuth", "application:Token signing method")},
  {id: "sm3-password", edition: "Enterprise", category: "edition:Chinese SM cryptography", name: "edition:SM3 password hashing", description: "edition:SM3 password hashing - Description",
    stop: org("password", "general:Password type", "general:Password type")},
  {id: "sm4-fields", edition: "Enterprise", category: "edition:Chinese SM cryptography", name: "edition:SM4 field encryption", description: "edition:SM4 field encryption - Description",
    stop: org("enterprise", "general:Compliance", "organization:Encrypted fields")},
  {id: "hsm", edition: "Enterprise", category: "edition:Keys and access control", name: "edition:Hardware security module", description: "edition:Hardware security module - Description",
    stop: {object: "cert", label: "cert:Key storage"}},

  {id: "admin-separation", edition: "Enterprise", category: "edition:Administration and audit", name: "edition:Three-admin separation", description: "edition:Three-admin separation - Description",
    stop: org("enterprise", "general:Compliance", "organization:Separate the admins")},
  {id: "delegated-admin", edition: "Enterprise", category: "edition:Administration and audit", name: "edition:Delegated administration", description: "edition:Delegated administration - Description",
    stop: org("enterprise", "general:Compliance", undefined, "[data-tour='delegated-admins']")},
  {id: "tamper-proof", edition: "Enterprise", category: "edition:Administration and audit", name: "edition:Tamper-proof audit records", description: "edition:Tamper-proof audit records - Description",
    stop: org("enterprise", "general:Compliance", "organization:Tamper-proof records")},
  {id: "record-integrity", edition: "Enterprise", category: "edition:Administration and audit", name: "edition:Record integrity verification", description: "edition:Record integrity verification - Description",
    stop: compliance("records", "enterprise:Record integrity")},
  {id: "consent-log", edition: "Enterprise", category: "edition:Administration and audit", name: "edition:Consent records", description: "edition:Consent records - Description",
    stop: org("enterprise", "general:Compliance", "organization:Record consents")},
  {id: "siem", edition: "Enterprise", category: "edition:Administration and audit", name: "edition:CEF and LEEF log forwarding", description: "edition:CEF and LEEF log forwarding - Description",
    stop: {path: "/providers"}},

  {id: "db-provisioning", edition: "Enterprise", category: "edition:Enterprise connections", name: "edition:Provisioning to legacy databases", description: "edition:Provisioning to legacy databases - Description",
    stop: app("provisioning", "application:Provisioning", "application:Enable provisioning")},
  {id: "mcp-approval", edition: "Enterprise", category: "edition:AI agents", name: "edition:Approval of tool calls", description: "edition:Approval of tool calls - Description",
    stop: {object: "server", label: "server:Needs approval"}},

  {id: "sod", edition: "Enterprise", category: "edition:Governance and compliance", name: "edition:Separation-of-duty rules", description: "edition:Separation-of-duty rules - Description",
    stop: org("enterprise", "general:Compliance", undefined, "[data-tour='sod-rules']")},
  {id: "governance", edition: "Enterprise", category: "edition:Governance and compliance", name: "edition:Orphan account cleanup", description: "edition:Orphan account cleanup - Description",
    stop: compliance("governance", "enterprise:Account governance")},
  {id: "access-review", edition: "Enterprise", category: "edition:Governance and compliance", name: "edition:Access reviews", description: "edition:Access reviews - Description",
    stop: {path: "/access-reviews", target: "#add-button"}},
  {id: "mlps", edition: "Enterprise", category: "edition:Governance and compliance", name: "edition:MLPS self-check", description: "edition:MLPS self-check - Description",
    stop: compliance("report", "enterprise:MLPS checklist")},
  {id: "security-overview", edition: "Enterprise", category: "edition:Governance and compliance", name: "edition:Security overview and monthly report", description: "edition:Security overview and monthly report - Description",
    stop: compliance("security", "enterprise:Security overview")},
  {id: "security-screen", edition: "Enterprise", category: "edition:Governance and compliance", name: "edition:Security situation screen", description: "edition:Security situation screen - Description",
    stop: {path: "/compliance", tab: "security", tabLabel: "enterprise:Security overview", target: "[data-tour='security-screen']"}},
  {id: "security-report-word", edition: "Enterprise", category: "edition:Governance and compliance", name: "edition:Security report in Word", description: "edition:Security report in Word - Description",
    stop: {path: "/compliance", tab: "security", tabLabel: "enterprise:Security overview", target: "[data-tour='security-report-word']"}},
  {id: "posture-score", edition: "Enterprise", category: "edition:Governance and compliance", name: "edition:Security score", description: "edition:Security score - Description",
    stop: {path: "/", target: "[data-tour='edition-hero']"}},

  {id: "hardware-tokens", edition: "Enterprise", category: "edition:Enterprise connections", name: "edition:Hardware OTP tokens", description: "edition:Hardware OTP tokens - Description",
    stop: compliance("otp", "enterprise:Hardware tokens")},
  {id: "hr-sync", edition: "Enterprise", category: "edition:Enterprise connections", name: "edition:HR system sync", description: "edition:HR system sync - Description",
    stop: marketplace("edition=Enterprise&q=syncer")},
  {id: "gov-signin", edition: "Enterprise", category: "edition:Enterprise connections", name: "edition:Government and office sign-in", description: "edition:Government and office sign-in - Description",
    stop: marketplace("edition=Enterprise&q=idp")},
  {id: "id-verification", edition: "Enterprise", category: "edition:Enterprise connections", name: "edition:Real-name verification", description: "edition:Real-name verification - Description",
    stop: marketplace("edition=Enterprise&q=verification")},
  {id: "form-fill", edition: "Enterprise", category: "edition:Enterprise connections", name: "edition:Legacy system form autofill", description: "edition:Legacy system form autofill - Description",
    stop: app("form-fill", "application:Form autofill", "application:Enable form autofill")},
  {id: "pam", edition: "Enterprise", category: "edition:Administration and audit", name: "edition:Privileged account management", description: "edition:Privileged account management - Description",
    stop: {path: "/privileged-accounts"}},
  {id: "cert-login", edition: "Enterprise", category: "edition:Enterprise connections", name: "edition:UKey and certificate login", description: "edition:UKey and certificate login - Description",
    stop: org("enterprise", "general:Compliance", "organization:Enable certificate login")},
];

export function getEditionFeatures(edition: PaidEdition): EditionFeature[] {
  return EditionFeatures.filter((feature) => feature.edition === edition);
}

export function getEditionFeature(id: string): EditionFeature | undefined {
  return EditionFeatures.find((feature) => feature.id === id);
}

const ranks: Record<string, number> = {Community: 0, Commercial: 1, Enterprise: 2};

export function isEditionCovered(current: string | null | undefined, required: string): boolean {
  return (ranks[current ?? "Community"] ?? 0) >= (ranks[required] ?? 0);
}

const objectListPaths: Record<StopObject, string> = {
  organization: "/organizations",
  application: "/applications",
  cert: "/certs",
  permission: "/permissions",
  model: "/models",
  server: "/servers",
};

export function getObjectListPath(object: StopObject) {
  return objectListPaths[object];
}

function firstOf(promise: Promise<any>): Promise<any | null> {
  return promise
    .then((res: any) => (res?.status === "ok" && Array.isArray(res.data) && res.data.length > 0 ? res.data[0] : null))
    .catch(() => null);
}

/** The edit page of the first object of a kind in the organization the console is on, or the list page when there is none. */
export async function resolveObjectPath(object: StopObject, account: any): Promise<string> {
  const organization = Setting.getRequestOrganization(account);
  let record: any = null;
  switch (object) {
  case "organization":
    return `/organizations/${organization}`;
  case "application":
    record = await firstOf(ApplicationBackend.getApplicationsByOrganization("admin", organization, 1, 1));
    return record ? `/applications/${record.organization}/${record.name}` : objectListPaths[object];
  case "cert":
    record = await firstOf(CertBackend.getCerts(organization, 1, 1));
    if (!record && Setting.isAdminUser(account)) {
      record = await firstOf(CertBackend.getGlobalCerts(1, 1));
    }
    return record ? `/certs/${record.owner}/${record.name}` : objectListPaths[object];
  case "permission":
    record = await firstOf(PermissionBackend.getPermissions(organization, 1, 1));
    return record ? `/permissions/${record.owner}/${record.name}` : objectListPaths[object];
  case "model":
    record = await firstOf(ModelBackend.getModels(organization, 1, 1));
    return record ? `/models/${record.owner}/${record.name}` : objectListPaths[object];
  case "server":
    record = await firstOf(ServerBackend.getServers(organization, 1, 1));
    return record ? `/servers/${record.owner}/${record.name}` : objectListPaths[object];
  default:
    return "/";
  }
}
