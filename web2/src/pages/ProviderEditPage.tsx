import i18next from "i18next";
import {useParams} from "react-router-dom";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useOrganizationOptions} from "@/hooks/use-options";
import * as ProviderBackend from "@/backend/ProviderBackend";
import * as Setting from "@/lib/setting";

const CATEGORIES = [
  "Captcha",
  "Email",
  "Face ID",
  "ID Verification",
  "Log",
  "MFA",
  "Notification",
  "OAuth",
  "Payment",
  "SAML",
  "SMS",
  "Scan",
  "Storage",
  "Web3",
].sort((a, b) => a.localeCompare(b));

/** Categories whose credentials are the generic clientId/clientSecret pair. */
const CLIENT_CREDENTIAL_CATEGORIES = ["OAuth", "Payment", "SMS", "Storage", "Captcha", "MFA", "Notification", "ID Verification", "Scan", "Face ID", "Web3"];

export default function ProviderEditPage() {
  const {organizationName = "", providerName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();

  const isCategory = (...categories: string[]) => (ctx: {record: any}) => categories.includes(ctx.record.category);

  const fields: EditField[] = [
    {
      type: "select",
      name: "owner",
      labelKey: "general:Organization",
      options: () => organizations,
      disabled: () => !Setting.isAdminUser(account),
    },
    {type: "text", name: "name", labelKey: "general:Name"},
    {type: "text", name: "displayName", labelKey: "general:Display name"},
    {
      type: "select",
      name: "category",
      labelKey: "general:Category",
      options: () => CATEGORIES.map((item) => ({value: item, label: item})),
    },
    {
      type: "select",
      name: "type",
      labelKey: "general:Type",
      options: (ctx) =>
        (Setting.getProviderTypeOptions(ctx.record.category) as any[]).map((item) => ({
          value: item.id,
          label: item.name,
        })),
    },
    {
      type: "select",
      name: "subType",
      labelKey: "provider:Sub type",
      when: (ctx) => ctx.record.category === "SMS" && ctx.record.type === "Volc Engine SMS",
      options: () => [{value: "Volc Engine SMS", label: "Volc Engine SMS"}],
    },
    {
      type: "select",
      name: "method",
      labelKey: "provider:Method",
      when: isCategory("Face ID", "Captcha"),
      options: () => ["Normal", "Silent"].map((item) => ({value: item, label: item})),
    },
    {
      type: "text",
      name: "clientId",
      labelKey: "provider:Client ID",
      when: (ctx) => CLIENT_CREDENTIAL_CATEGORIES.includes(ctx.record.category),
    },
    {
      type: "password",
      name: "clientSecret",
      labelKey: "provider:Client secret",
      when: (ctx) => CLIENT_CREDENTIAL_CATEGORIES.includes(ctx.record.category),
    },
    {
      type: "text",
      name: "clientId2",
      labelKey: "provider:Client ID 2",
      when: isCategory("OAuth", "SMS", "Payment"),
    },
    {
      type: "password",
      name: "clientSecret2",
      labelKey: "provider:Client secret 2",
      when: isCategory("OAuth", "SMS", "Payment"),
    },
    {type: "text", name: "appId", labelKey: "provider:App ID", when: isCategory("OAuth", "SMS", "Payment", "Notification")},
    {type: "text", name: "scopes", labelKey: "provider:Scope", when: isCategory("OAuth")},
    {type: "text", name: "domain", labelKey: "provider:Domain", when: isCategory("OAuth", "Storage", "SAML")},
    {type: "text", name: "customAuthUrl", labelKey: "provider:Custom auth URL", when: (ctx) => String(ctx.record.type ?? "").startsWith("Custom")},
    {type: "text", name: "customTokenUrl", labelKey: "provider:Custom token URL", when: (ctx) => String(ctx.record.type ?? "").startsWith("Custom")},
    {type: "text", name: "customUserInfoUrl", labelKey: "provider:Custom userinfo URL", when: (ctx) => String(ctx.record.type ?? "").startsWith("Custom")},

    {type: "text", name: "host", labelKey: "general:Host", when: isCategory("Email", "Storage", "Log")},
    {type: "number", name: "port", labelKey: "general:Port", when: isCategory("Email", "Storage", "Log")},
    {type: "switch", name: "disableSsl", labelKey: "provider:Disable SSL", when: isCategory("Email", "Storage", "OAuth")},
    {type: "text", name: "title", labelKey: "provider:Email title", when: isCategory("Email")},
    {type: "textarea", name: "content", labelKey: "provider:Email content", rows: 10, when: isCategory("Email", "SMS", "Notification")},
    {type: "text", name: "receiver", labelKey: "provider:Test email/phone", when: isCategory("Email", "SMS")},

    {type: "text", name: "endpoint", labelKey: "provider:Endpoint", when: isCategory("Storage", "Log", "SAML", "Web3")},
    {type: "text", name: "providerUrl", labelKey: "provider:Provider URL", when: isCategory("OAuth", "SAML", "Payment", "Storage")},
    {type: "text", name: "idP", labelKey: "provider:IdP", when: isCategory("SAML")},
    {type: "text", name: "issuerUrl", labelKey: "provider:Issuer URL", when: isCategory("SAML")},
    {type: "textarea", name: "metadata", labelKey: "provider:Metadata", rows: 8, when: isCategory("SAML")},
    {
      type: "select",
      name: "sslMode",
      labelKey: "syncer:SSL mode",
      when: isCategory("Log"),
      options: () => ["Default", "SSL", "TLS"].map((item) => ({value: item, label: item})),
    },
  ];

  return (
    <SimpleEditPage
      titleKey="provider:Edit Provider"
      backTo="/providers"
      deps={[organizationName, providerName]}
      fields={fields}
      fetch={() => ProviderBackend.getProvider(organizationName, providerName)}
      add={(record) => ProviderBackend.addProvider(record)}
      update={(record) => ProviderBackend.updateProvider(organizationName, providerName, record)}
      editUrl={(record) => `/providers/${record.owner}/${record.name}`}
    >
      {(ctx) =>
        ctx.record.category === "OAuth" ? (
          <p className="pt-2 text-xs text-muted-foreground">
            {i18next.t("provider:Redirect URL")}: {window.location.origin}/callback
          </p>
        ) : null
      }
    </SimpleEditPage>
  );
}
