import i18next from "i18next";
import {useParams} from "react-router-dom";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {
  useApplicationOptions,
  useCertOptions,
  useOrganizationOptions,
  useProviderOptions,
  useRuleOptions,
} from "@/hooks/use-options";
import * as SiteBackend from "@/backend/SiteBackend";
import * as Setting from "@/lib/setting";

const SSL_MODES = ["HTTP", "HTTPS Only", "HTTP and HTTPS"];

export default function SiteEditPage() {
  const {organizationName = "", siteName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();
  const applications = useApplicationOptions(organizationName);
  const certs = useCertOptions(organizationName);
  const rules = useRuleOptions(organizationName);
  const notificationProviders = useProviderOptions(organizationName, "Notification");

  const fields: EditField[] = [
    {
      type: "select",
      name: "owner",
      labelKey: "general:Organization",
      options: () => organizations,
      disabled: () => !Setting.isAdminUser(account),
    },
    {type: "text", name: "name", labelKey: "general:Name", required: true},
    {type: "text", name: "displayName", labelKey: "general:Display name"},
    {type: "text", name: "tag", labelKey: "general:Tag"},
    {type: "text", name: "domain", labelKey: "provider:Domain"},
    {type: "tags", name: "otherDomains", labelKey: "application:Other domains"},
    {type: "switch", name: "needRedirect", labelKey: "site:Need redirect"},
    {type: "switch", name: "disableVerbose", labelKey: "site:Disable verbose"},
    {type: "text", name: "host", labelKey: "general:Host"},
    {type: "number", name: "port", labelKey: "general:Port"},
    {type: "tags", name: "hosts", labelKey: "site:Hosts"},
    {type: "text", name: "publicIp", labelKey: "site:Public IP"},
    {type: "text", name: "node", labelKey: "site:Node"},
    {type: "switch", name: "isSelf", labelKey: "general:Self"},
    {
      type: "select",
      name: "sslMode",
      labelKey: "site:Mode",
      options: () => SSL_MODES.map((item) => ({value: item, label: item})),
    },
    {type: "select", name: "sslCert", labelKey: "application:SSL cert", options: () => certs},
    {type: "select", name: "casdoorApplication", labelKey: "site:Casdoor app", options: () => applications},
    {
      type: "select",
      name: "status",
      labelKey: "general:Status",
      options: () => [
        {value: "Active", label: i18next.t("subscription:Active")},
        {value: "Inactive", label: i18next.t("key:Inactive")},
      ],
    },
    {type: "tags", name: "challenges", labelKey: "site:Challenges"},
    {type: "multiselect", name: "rules", labelKey: "general:Rules", options: () => rules},
    {type: "switch", name: "enableAlert", labelKey: "site:Enable alert"},
    {
      type: "number",
      name: "alertInterval",
      labelKey: "site:Alert interval",
      suffix: i18next.t("usage:seconds"),
      when: (ctx) => !!ctx.record.enableAlert,
    },
    {type: "number", name: "alertTryTimes", labelKey: "site:Alert try times", when: (ctx) => !!ctx.record.enableAlert},
    {
      type: "multiselect",
      name: "alertProviders",
      labelKey: "site:Alert providers",
      when: (ctx) => !!ctx.record.enableAlert,
      options: () => notificationProviders,
    },
  ];

  return (
    <SimpleEditPage
      titleKey="site:Edit Site"
      backTo="/sites"
      deps={[organizationName, siteName]}
      fields={fields}
      fetch={() => SiteBackend.getSite(organizationName, siteName)}
      add={(record) => SiteBackend.addSite(record)}
      update={(record) => SiteBackend.updateSite(organizationName, siteName, record)}
      editUrl={(record) => `/sites/${record.owner}/${record.name}`}
    />
  );
}
