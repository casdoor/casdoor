import * as React from "react";
import i18next from "i18next";
import {useNavigate, useParams} from "react-router-dom";
import {UnauthorizedPage} from "@/components/common/UnauthorizedPage";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Switch} from "@/components/ui/switch";
import {Textarea} from "@/components/ui/textarea";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table";
import {Loading} from "@/components/common/Loading";
import {CodeEditor} from "@/components/common/CodeEditor";
import {MultiSelect} from "@/components/common/MultiSelect";
import {SearchableSelect} from "@/components/common/SearchableSelect";
import {SelectField} from "@/components/common/SelectField";
import {EditableTable} from "@/components/crud/EditableTable";
import {EditPageShell} from "@/components/crud/EditPageShell";
import {FormRow} from "@/components/crud/FormRow";
import {CaptchaPreview} from "@/components/provider/CaptchaPreview";
import {useAccount} from "@/hooks/use-account";
import {useEditRecord} from "@/hooks/use-edit-record";
import * as CertBackend from "@/backend/CertBackend";
import * as OrganizationBackend from "@/backend/OrganizationBackend";
import * as ProviderBackend from "@/backend/ProviderBackend";
import * as ServerBackend from "@/backend/ServerBackend";
import {mapToRows, rowsToMap, submitEdit} from "@/lib/crud";
import * as ProviderTest from "@/lib/provider-test";
import * as Setting from "@/lib/setting";
import {authConfig} from "@/auth/Auth";
import copy from "copy-to-clipboard";

const defaultUserMapping: Record<string, string> = {
  id: "id",
  username: "username",
  displayName: "displayName",
  email: "email",
  avatarUrl: "avatarUrl",
  phone: "phone",
  countryCode: "country_code",
  firstName: "given_name",
  lastName: "family_name",
  region: "region",
  location: "location",
  affiliation: "affiliation",
  title: "title",
};

const defaultEmailMapping: Record<string, string> = {
  fromName: "fromName",
  fromAddress: "fromAddress",
  toAddress: "toAddress",
  subject: "subject",
  content: "content",
};

const defaultSmsMapping: Record<string, string> = {
  phoneNumber: "phoneNumber",
  content: "content",
};

const CATEGORIES = [
  "Captcha", "Email", "Face ID", "ID Verification", "Log", "MFA", "Notification",
  "OAuth", "Payment", "SAML", "Scan", "SMS", "Storage", "Web3",
].sort((a, b) => a.localeCompare(b));

const HTTP_METHODS = ["GET", "POST", "PUT", "DELETE"];
const CONTENT_TYPES = ["application/json", "application/x-www-form-urlencoded"];

const SMS_PROVIDERS_WITHOUT_SIGN_NAME = ["Custom HTTP SMS", "Twilio SMS", "Amazon SNS", "Msg91 SMS", "Infobip SMS"];
const SMS_PROVIDERS_WITHOUT_TEMPLATE_CODE = ["Infobip SMS"];

const SCAN_HOST_OPTIONS = ["127.0.0.1/32", "10.0.0.0/24", "172.16.0.0/24", "192.168.1.0/24"];
const SCAN_PORT_OPTIONS = ["80", "3000", "8080"];
const SCAN_PATH_OPTIONS = ["/", "/mcp", "/sse", "/mcp/sse"];

/** the keys of web3Wallets in web/src/auth/Web3Auth.js — the values the backend stores in `metadata` */
const WEB3_ONBOARD_WALLETS = [
  {value: "injected", label: "Injected"},
  {value: "phantom", label: "Phantom"},
  {value: "coinbase", label: "Coinbase"},
  {value: "trust", label: "Trust"},
  {value: "gnosis", label: "Gnosis"},
  {value: "sequence", label: "Sequence"},
  {value: "taho", label: "Taho"},
  {value: "frontier", label: "Frontier"},
  {value: "infinityWallet", label: "Infinity Wallet"},
];

function isDefaultProviderName(name: string) {
  return /^provider_[a-z0-9]+$/.test(name ?? "");
}

function isDefaultProviderDisplayName(displayName: string) {
  return /^New Provider - [a-z0-9]+$/.test(displayName ?? "");
}

function slug(value: string) {
  return value.toLowerCase().replace(/[\s-]+/g, "_").replace(/[^a-z0-9_]/g, "");
}

function getAutoProviderName(category: string, type: string, subType: string) {
  if (subType) {
    return `provider_${slug(category)}_${slug(type)}_${slug(subType)}`;
  }
  return `provider_${slug(category)}_${slug(type)}`;
}

function getAutoProviderDisplayName(category: string, type: string, subType: string) {
  return subType ? `${category} ${type} ${subType}` : `${category} ${type}`;
}

/** the (label, tooltip) pair a row shows, the shadcn counterpart of Setting.getLabel() */
type Label = {label: string; tooltip?: string};

function label(text: string, tooltip?: string): Label {
  return {label: text, tooltip};
}

function toList(rawValue: any) {
  return `${rawValue || ""}`.split(",").map(item => item.trim()).filter(item => item !== "");
}

function normalizeAndJoin(values: string[]) {
  return (values || []).map(item => `${item}`.trim()).filter(item => item !== "").join(",");
}

function getProviderSubTypeOptions(type: string) {
  if (type === "Agent") {
    return [{id: "OpenClaw", name: "OpenClaw"}];
  } else if (type === "Security Scan") {
    return [{id: "Site", name: "Site"}, {id: "Url", name: "Url"}];
  } else if (type === "MCP Scan") {
    return [{id: "Intranet Scan", name: "Intranet Scan"}];
  } else if (type === "WeCom" || type === "Infoflow") {
    return [
      {id: "Internal", name: i18next.t("provider:Internal")},
      {id: "Third-party", name: i18next.t("provider:Third-party")},
    ];
  } else if (type === "WeChat") {
    return [
      {id: "Web", name: i18next.t("provider:Web")},
      {id: "Mobile", name: i18next.t("provider:Mobile")},
    ];
  }
  return [];
}

function getClientIdLabel(provider: any): Label {
  if (provider.category === "OAuth") {
    return provider.type === "Apple"
      ? label(i18next.t("provider:Service ID identifier"), i18next.t("provider:Service ID identifier - Tooltip"))
      : label(i18next.t("provider:Client ID"), i18next.t("provider:Client ID - Tooltip"));
  } else if (provider.category === "Email") {
    return label(i18next.t("signup:Username"), i18next.t("signup:Username - Tooltip"));
  } else if (provider.category === "SMS") {
    if (["Volc Engine SMS", "Amazon SNS", "Baidu Cloud SMS"].includes(provider.type)) {
      return label(i18next.t("general:Access key"), i18next.t("general:Access key - Tooltip"));
    } else if (provider.type === "Huawei Cloud SMS") {
      return label(i18next.t("provider:App key"), i18next.t("provider:App key - Tooltip"));
    } else if (provider.type === "UCloud SMS") {
      return label(i18next.t("provider:Public key"), i18next.t("provider:Public key - Tooltip"));
    } else if (["Msg91 SMS", "Infobip SMS", "OSON SMS"].includes(provider.type)) {
      return label(i18next.t("provider:Sender Id"), i18next.t("provider:Sender Id - Tooltip"));
    }
    return label(i18next.t("provider:Client ID"), i18next.t("provider:Client ID - Tooltip"));
  } else if (provider.category === "Captcha") {
    return provider.type === "Aliyun Captcha"
      ? label(i18next.t("general:Access key"), i18next.t("general:Access key - Tooltip"))
      : label(i18next.t("provider:Site key"), i18next.t("provider:Site key - Tooltip"));
  } else if (provider.category === "Notification") {
    return provider.type === "DingTalk"
      ? label(i18next.t("general:Access key"), i18next.t("general:Access key - Tooltip"))
      : label(i18next.t("provider:Client ID"), i18next.t("provider:Client ID - Tooltip"));
  } else if (provider.category === "ID Verification") {
    return provider.type === "Alibaba Cloud"
      ? label(i18next.t("general:Access key"), i18next.t("general:Access key - Tooltip"))
      : label(i18next.t("provider:Client ID"), i18next.t("provider:Client ID - Tooltip"));
  }
  return label(i18next.t("provider:Client ID"), i18next.t("provider:Client ID - Tooltip"));
}

function getClientSecretLabel(provider: any): Label {
  if (provider.category === "OAuth") {
    return provider.type === "Apple"
      ? label(i18next.t("provider:Team ID"), i18next.t("provider:Team ID - Tooltip"))
      : label(i18next.t("provider:Client secret"), i18next.t("provider:Client secret - Tooltip"));
  } else if (provider.category === "Storage") {
    return provider.type === "Google Cloud Storage"
      ? label(i18next.t("provider:Service account JSON"), i18next.t("provider:Service account JSON - Tooltip"))
      : label(i18next.t("provider:Client secret"), i18next.t("provider:Client secret - Tooltip"));
  } else if (provider.category === "Email") {
    return ["Azure ACS", "SendGrid", "Resend"].includes(provider.type)
      ? label(i18next.t("provider:Secret key"), i18next.t("provider:Secret key - Tooltip"))
      : label(i18next.t("general:Password"), i18next.t("general:Password - Tooltip"));
  } else if (provider.category === "SMS") {
    if (["Volc Engine SMS", "Amazon SNS", "Baidu Cloud SMS", "OSON SMS"].includes(provider.type)) {
      return label(i18next.t("provider:Secret access key"), i18next.t("provider:Secret access key - Tooltip"));
    } else if (provider.type === "Huawei Cloud SMS") {
      return label(i18next.t("provider:App secret"), i18next.t("provider:AppSecret - Tooltip"));
    } else if (provider.type === "UCloud SMS") {
      return label(i18next.t("provider:Private Key"), i18next.t("provider:Private Key - Tooltip"));
    } else if (provider.type === "Msg91 SMS") {
      return label(i18next.t("provider:Auth Key"), i18next.t("provider:Auth Key - Tooltip"));
    } else if (provider.type === "Infobip SMS") {
      return label(i18next.t("provider:Api Key"), i18next.t("provider:Api Key - Tooltip"));
    }
    return label(i18next.t("provider:Client secret"), i18next.t("provider:Client secret - Tooltip"));
  } else if (provider.category === "Captcha") {
    return provider.type === "Aliyun Captcha"
      ? label(i18next.t("provider:Secret access key"), i18next.t("provider:Secret access key - Tooltip"))
      : label(i18next.t("provider:Secret key"), i18next.t("provider:Secret key - Tooltip"));
  } else if (provider.category === "Notification") {
    if (["Line", "Telegram", "Bark", "DingTalk", "Discord", "Slack", "Pushover", "Pushbullet"].includes(provider.type)) {
      return label(i18next.t("provider:Secret key"), i18next.t("provider:Secret key - Tooltip"));
    } else if (["Lark", "Microsoft Teams", "WeCom"].includes(provider.type)) {
      return label(i18next.t("provider:Endpoint"), i18next.t("provider:Endpoint - Tooltip"));
    }
    return label(i18next.t("provider:Client secret"), i18next.t("provider:Client secret - Tooltip"));
  } else if (provider.category === "ID Verification") {
    return provider.type === "Alibaba Cloud"
      ? label(i18next.t("provider:Secret access key"), i18next.t("provider:Secret access key - Tooltip"))
      : label(i18next.t("provider:Client secret"), i18next.t("provider:Client secret - Tooltip"));
  }
  return label(i18next.t("provider:Client secret"), i18next.t("provider:Client secret - Tooltip"));
}

function getClientId2Label(provider: any): Label {
  if (provider.category === "OAuth") {
    return provider.type === "Apple"
      ? label(i18next.t("provider:Key ID"), i18next.t("provider:Key ID - Tooltip"))
      : label(i18next.t("provider:Client ID 2"), i18next.t("provider:Client ID 2 - Tooltip"));
  } else if (provider.category === "Email") {
    return label(i18next.t("provider:From address"), i18next.t("provider:From address - Tooltip"));
  } else if (provider.type === "Aliyun Captcha") {
    return label(i18next.t("provider:Scene"), i18next.t("provider:Scene - Tooltip"));
  } else if (provider.type === "WeChat Pay" || provider.type === "CUCloud") {
    return label(i18next.t("provider:App ID"), i18next.t("provider:App ID - Tooltip"));
  }
  return label(i18next.t("provider:Client ID 2"), i18next.t("provider:Client ID 2 - Tooltip"));
}

function getClientSecret2Label(provider: any): Label {
  if (provider.category === "OAuth") {
    return provider.type === "Apple"
      ? label(i18next.t("provider:Key text"), i18next.t("provider:Key text - Tooltip"))
      : label(i18next.t("provider:Client secret 2"), i18next.t("provider:Client secret 2 - Tooltip"));
  } else if (provider.category === "Email") {
    return label(i18next.t("provider:From name"), i18next.t("provider:From name - Tooltip"));
  } else if (provider.type === "Aliyun Captcha") {
    return label(i18next.t("provider:App key"), i18next.t("provider:App key - Tooltip"));
  }
  return label(i18next.t("provider:Client secret 2"), i18next.t("provider:Client secret 2 - Tooltip"));
}

/** the extra "appId" row, whose meaning depends on the provider — null hides it */
function getAppIdLabel(provider: any): Label | null {
  if (provider.category === "OAuth") {
    if ((provider.type === "WeCom" && provider.subType === "Internal") || provider.type === "Infoflow") {
      return label(i18next.t("provider:Agent ID"), i18next.t("provider:Agent ID - Tooltip"));
    } else if (provider.type === "AzureADB2C") {
      return label(i18next.t("provider:User flow"), i18next.t("provider:User flow - Tooltip"));
    }
  } else if (provider.category === "SMS") {
    if (provider.type === "Twilio SMS" || provider.type === "Azure ACS") {
      return label(i18next.t("provider:Sender number"), i18next.t("provider:Sender number - Tooltip"));
    } else if (provider.type === "Tencent Cloud SMS") {
      return label(i18next.t("provider:App ID"), i18next.t("provider:App ID - Tooltip"));
    } else if (provider.type === "Volc Engine SMS") {
      return label(i18next.t("provider:SMS account"), i18next.t("provider:SMS account - Tooltip"));
    } else if (provider.type === "Huawei Cloud SMS") {
      return label(i18next.t("provider:Channel No."), i18next.t("provider:Channel No. - Tooltip"));
    } else if (provider.type === "Amazon SNS") {
      return label(i18next.t("provider:Region"), i18next.t("provider:Region - Tooltip"));
    } else if (provider.type === "Baidu Cloud SMS") {
      return label(i18next.t("provider:Endpoint"), i18next.t("provider:Endpoint - Tooltip"));
    } else if (provider.type === "Infobip SMS") {
      return label(i18next.t("provider:Base URL"), i18next.t("provider:Base URL - Tooltip"));
    } else if (provider.type === "UCloud SMS") {
      return label(i18next.t("provider:Project Id"), i18next.t("provider:Project Id - Tooltip"));
    }
  } else if (provider.category === "Email") {
    if (provider.type === "SUBMAIL") {
      return label(i18next.t("provider:App ID"), i18next.t("provider:App ID - Tooltip"));
    }
  } else if (provider.category === "Notification") {
    if (provider.type === "Viber") {
      return label(i18next.t("provider:Domain"), i18next.t("provider:Domain - Tooltip"));
    } else if (["Line", "Matrix", "Rocket Chat"].includes(provider.type)) {
      return label(i18next.t("provider:App Key"), i18next.t("provider:App Key - Tooltip"));
    } else if (provider.type === "CUCloud") {
      return label("Topic name", "Topic name - Tooltip");
    }
  }
  return null;
}

/** the notification "receiver" row — null when the provider needs no receiver */
function getReceiverLabel(provider: any): Label | null {
  if (["Telegram", "Pushover", "Pushbullet", "Slack", "Discord", "Line", "Twitter", "Reddit", "Rocket Chat", "Viber"].includes(provider.type)) {
    return label(i18next.t("provider:Chat ID"), i18next.t("provider:Chat ID - Tooltip"));
  } else if (["Custom HTTP", "Webpush", "Matrix"].includes(provider.type)) {
    return label(i18next.t("provider:Endpoint"), i18next.t("provider:Endpoint - Tooltip"));
  }
  return null;
}

function hasClientIdRow(provider: any) {
  if ((provider.category === "Storage" && provider.type === "Google Cloud Storage") ||
      (provider.category === "Email" && ["Azure ACS", "SendGrid", "Resend"].includes(provider.type)) ||
      (provider.category === "Face ID" && provider.type === "Local UniFace") ||
      (provider.category === "Notification" && ["Line", "Telegram", "Bark", "Discord", "Slack", "Pushbullet", "Pushover", "Lark", "Microsoft Teams", "WeCom"].includes(provider.type))) {
    return false;
  }
  return true;
}

function hasCredentialRows(provider: any) {
  if ((provider.category === "Captcha" && provider.type === "Default") ||
      provider.category === "Web3" ||
      provider.category === "MFA" ||
      provider.category === "Log" ||
      provider.category === "Scan" ||
      (provider.category === "Storage" && provider.type === "Local File System") ||
      (provider.category === "SMS" && provider.type === "Custom HTTP SMS") ||
      (provider.category === "Email" && provider.type === "Custom HTTP Email") ||
      (provider.category === "Notification" && ["Google Chat", "Custom HTTP", "Balance"].includes(provider.type))) {
    return false;
  }
  return true;
}

function hasCredential2Rows(provider: any) {
  return provider.category === "Email" ||
    ["WeChat", "Apple", "Aliyun Captcha", "WeChat Pay", "Twitter", "Reddit", "CUCloud"].includes(provider.type);
}

function parseFindings(provider: any, scanResult: any) {
  if (Array.isArray(scanResult)) {
    return scanResult;
  }
  if (!provider?.metadata) {
    return [];
  }
  try {
    const parsed = JSON.parse(provider.metadata);
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
}

function getCveLink(cve: any) {
  const references = Array.isArray(cve?.references) ? cve.references : [];
  return references.find((reference: any) => {
    if (typeof reference !== "string") {
      return false;
    }
    const value = reference.trim();
    return value.startsWith("http://") || value.startsWith("https://");
  }) || "";
}

function getEntryPath(subType: string, owner: string, name: string) {
  if (!owner || !name) {
    return "";
  }
  if (subType === "Site") {
    return `/sites/${owner}/${name}`;
  }
  if (subType === "Agent") {
    return `/agents/${owner}/${name}`;
  }
  return "";
}

export default function ProviderEditPage() {
  const {organizationName = "", providerName = ""} = useParams();
  const {account} = useAccount();
  const navigate = useNavigate();

  const [owner, setOwner] = React.useState(organizationName);
  const [name, setName] = React.useState(providerName);
  const [organizations, setOrganizations] = React.useState<any[]>([]);
  const [providers, setProviders] = React.useState<any[]>([]);
  const [certs, setCerts] = React.useState<any[]>([]);
  const [saving, setSaving] = React.useState(false);

  const [nameNotUserEdited, setNameNotUserEdited] = React.useState(false);
  const [displayNameNotUserEdited, setDisplayNameNotUserEdited] = React.useState(false);

  const [requestUrl, setRequestUrl] = React.useState("");
  const [metadataLoading, setMetadataLoading] = React.useState(false);
  const [discoveryLoading, setDiscoveryLoading] = React.useState(false);

  const [scanLoading, setScanLoading] = React.useState(false);
  const [scanResult, setScanResult] = React.useState<any>(null);
  const [scanServers, setScanServers] = React.useState<any[]>([]);

  const transform = React.useCallback((provider: any) => {
    if (provider.type === "Custom HTTP Email") {
      if (!provider.userMapping || !provider.userMapping.fromName) {
        provider.userMapping = {...defaultEmailMapping};
      }
    } else if (provider.type === "Custom HTTP SMS") {
      if (!provider.userMapping || !provider.userMapping.phoneNumber) {
        provider.userMapping = {...defaultSmsMapping};
      }
    } else {
      provider.userMapping = provider.userMapping || {...defaultUserMapping};
    }
    return provider;
  }, []);

  const {record: provider, setRecord, loading, denied, mode, setMode} = useEditRecord<any>({
    fetch: () => ProviderBackend.getProvider(organizationName, providerName),
    transform,
    deps: [organizationName, providerName],
  });

  const getProviders = React.useCallback((forOwner: string) => {
    ProviderBackend.getProviders(forOwner).then((res: any) => {
      if (res.status === "ok") {
        setProviders(res.data ?? []);
      }
    });
  }, []);

  const getCerts = React.useCallback((forOwner: string) => {
    CertBackend.getCerts(forOwner).then((res: any) => {
      if (res.status === "ok") {
        setCerts(res.data ?? []);
      }
    });
  }, []);

  React.useEffect(() => {
    if (Setting.isAdminUser(account)) {
      OrganizationBackend.getOrganizations("admin").then((res: any) => {
        setOrganizations(res.data ?? []);
      });
    }
  }, [account]);

  React.useEffect(() => {
    getProviders(owner);
    getCerts(owner);
  }, [owner, getProviders, getCerts]);

  // the name/display name keep following category+type until the user types their own
  const initializedRef = React.useRef(false);
  React.useEffect(() => {
    if (provider !== null && !initializedRef.current) {
      initializedRef.current = true;
      setNameNotUserEdited(isDefaultProviderName(provider.name));
      setDisplayNameNotUserEdited(isDefaultProviderDisplayName(provider.displayName));
    }
  }, [provider]);

  /** WeChat stores the login mode across three coupled fields, see the antd page. */
  const applyProviderRules = (next: any) => {
    if (next.type === "WeChat") {
      if (!next.clientId) {
        next.signName = "media";
        next.disableSsl = true;
      }
      if (!next.clientId2) {
        next.signName = "open";
        next.disableSsl = false;
      }
      if (!next.disableSsl) {
        next.signName = "open";
      }
    }
    return next;
  };

  const patchProvider = React.useCallback((patch: Record<string, any>) => {
    setRecord((prev: any) => {
      if (prev === null) {
        return prev;
      }
      const next = {...prev, ...patch};
      if (Object.prototype.hasOwnProperty.call(patch, "port")) {
        next.port = Setting.myParseInt(patch.port);
      }
      return applyProviderRules(next);
    });
  }, [setRecord]);

  const updateProviderField = React.useCallback((key: string, value: any) => {
    if (key === "owner" && provider !== null && provider.owner !== value) {
      // the provider changed owner, the cert (and the OpenClaw storage provider) no longer applies
      const patch: Record<string, any> = {owner: value, cert: ""};
      if (provider.category === "Log" && provider.type === "Agent" && provider.subType === "OpenClaw") {
        patch.providerUrl = "";
      }
      patchProvider(patch);
      getProviders(value);
      getCerts(value);
      return;
    }
    patchProvider({[key]: value});
  }, [provider, patchProvider, getProviders, getCerts]);

  const updateUserMappingField = (key: string, value: string) => {
    if (provider === null) {
      return;
    }
    const requiredKeys = ["id", "username", "displayName"];
    if (provider.type === "Custom HTTP Email") {
      if (value === "") {
        Setting.showMessage("error", i18next.t("general:This field is required"));
        return;
      }
    } else if (value === "" && requiredKeys.includes(key)) {
      Setting.showMessage("error", i18next.t("general:This field is required"));
      return;
    }

    const userMapping = {...(provider.userMapping ?? {})};
    if (value === "") {
      delete userMapping[key];
    } else {
      userMapping[key] = value;
    }
    patchProvider({userMapping});
  };

  const save = async(exitAfterSave: boolean) => {
    if (provider === null) {
      return;
    }
    const payload = Setting.deepCopy(provider);
    setSaving(true);
    await submitEdit({
      mode,
      record: payload,
      add: (record) => ProviderBackend.addProvider(record),
      update: (record) => ProviderBackend.updateProvider(owner, name, record),
      onSaved: () => {
        setOwner(provider.owner);
        setName(provider.name);
        setMode("edit");
        if (exitAfterSave) {
          navigate("/providers");
        } else {
          const next = `/providers/${provider.owner}/${provider.name}`;
          if (next !== window.location.pathname) {
            navigate(next, {replace: true});
          }
        }
      },
      onFailed: () => {
        if (mode !== "add") {
          patchProvider({name});
        }
      },
    });
    setSaving(false);
  };

  const loadSamlConfiguration = (metadata: string) => {
    const parser = new DOMParser();
    const xmlDoc = parser.parseFromString(metadata.replace("\n", ""), "text/xml");
    const cert = xmlDoc.querySelector("X509Certificate")!.childNodes[0].nodeValue!.replace(" ", "");
    const endpoint = xmlDoc.querySelector("SingleSignOnService")!.getAttribute("Location");
    const issuerUrl = xmlDoc.querySelector("EntityDescriptor")!.getAttribute("entityID");
    patchProvider({idP: cert, endpoint, issuerUrl});
  };

  const parseSamlMetadata = (metadata: string) => {
    try {
      loadSamlConfiguration(metadata);
      Setting.showMessage("success", i18next.t("provider:Parse metadata successfully"));
    } catch {
      Setting.showMessage("error", i18next.t("provider:Can not parse metadata"));
    }
  };

  const fetchSamlMetadata = () => {
    setMetadataLoading(true);
    fetch(requestUrl, {method: "GET"})
      .then(res => {
        if (!res.ok) {
          return Promise.reject(new Error("error"));
        }
        return res.text();
      })
      .then(text => {
        patchProvider({metadata: text});
        parseSamlMetadata(text);
        Setting.showMessage("success", i18next.t("general:Successfully added"));
      })
      .catch(err => {
        Setting.showMessage("error", err.message);
      })
      .finally(() => {
        setMetadataLoading(false);
      });
  };

  const fetchOidcDiscovery = () => {
    setDiscoveryLoading(true);
    ProviderBackend.getIdpDiscovery(provider.domain ?? "")
      .then((res: any) => {
        if (res.status !== "ok") {
          Setting.showMessage("error", res.msg);
          return;
        }

        const discovery = res.data;
        patchProvider({
          domain: discovery.issuer || provider.domain,
          customAuthUrl: discovery.authorization_endpoint ?? "",
          customTokenUrl: discovery.token_endpoint ?? "",
          customUserInfoUrl: discovery.userinfo_endpoint ?? "",
          customLogoutUrl: discovery.end_session_endpoint ?? "",
        });
        Setting.showMessage("success", i18next.t("general:Successfully added"));
      })
      .catch((err: any) => {
        Setting.showMessage("error", err.message);
      })
      .finally(() => {
        setDiscoveryLoading(false);
      });
  };

  const submitProviderScan = (target = "") => {
    if (provider === null) {
      return;
    }
    if (!provider.owner || !provider.name) {
      Setting.showMessage("error", i18next.t("provider:Provider owner and name are required"));
      return;
    }

    const isSecurityUrlScan = provider.type === "Security Scan" && provider.subType === "Url";
    const rawTarget = isSecurityUrlScan ? (target || provider.content || "") : target;

    setScanLoading(true);
    const scanApi = provider.type === "Security Scan"
      ? ServerBackend.scanProvider(provider.owner, provider.name, rawTarget)
      : ServerBackend.syncIntranetServers(provider.owner, provider.name);

    scanApi
      .then((res: any) => {
        setScanLoading(false);
        if (res.status === "ok") {
          const result = res.data ?? null;
          const servers = result?.servers ?? [];
          patchProvider({metadata: result === null ? "" : JSON.stringify(result)});
          setScanResult(result);
          setScanServers(servers);

          if (Array.isArray(result)) {
            Setting.showMessage("success", `${i18next.t("general:Successfully got")}: ${result.length} finding(s)`);
          } else if (Array.isArray(servers)) {
            Setting.showMessage("success", `${i18next.t("general:Successfully got")}: ${servers.length} server(s)`);
          } else {
            Setting.showMessage("success", i18next.t("general:Successfully saved"));
          }
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to get")}: ${res.msg}`);
        }
      })
      .catch((error: any) => {
        setScanLoading(false);
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  };

  if (denied) {
    return <UnauthorizedPage />;
  }

  if (loading || provider === null) {
    return <Loading />;
  }

  const certOptions = certs.map((cert: any) => ({value: cert.name, label: cert.name}));
  const subTypeOptions = getProviderSubTypeOptions(provider.type);

  const onChangeCategory = (value: string) => {
    const patch: Record<string, any> = {category: value};
    let defaultType = "";
    if (value === "OAuth") {
      defaultType = "Google";
    } else if (value === "Email") {
      defaultType = "Default";
      patch.host = "smtp.example.com";
      patch.port = 465;
      patch.sslMode = "Auto";
      patch.title = "Casdoor Verification Code";
      patch.content = Setting.getDefaultHtmlEmailContent();
      patch.metadata = Setting.getDefaultInvitationHtmlEmailContent();
      patch.receiver = account?.email ?? "";
    } else if (value === "SMS") {
      defaultType = "Twilio SMS";
    } else if (value === "Storage") {
      defaultType = "AWS S3";
    } else if (value === "SAML") {
      defaultType = "Keycloak";
    } else if (value === "Payment") {
      defaultType = "PayPal";
    } else if (value === "Captcha") {
      defaultType = "Default";
    } else if (value === "Web3") {
      defaultType = "MetaMask";
    } else if (value === "Notification") {
      defaultType = "Telegram";
    } else if (value === "Face ID") {
      defaultType = "Alibaba Cloud Facebody";
    } else if (value === "MFA") {
      defaultType = "RADIUS";
      patch.host = "";
      patch.port = 1812;
    } else if (value === "ID Verification") {
      defaultType = "Jumio";
      patch.endpoint = "";
    } else if (value === "Log") {
      defaultType = "Casdoor Permission Log";
      patch.host = "";
      patch.port = 0;
      patch.title = "";
      patch.state = "Enabled";
    } else if (value === "Scan") {
      defaultType = "MCP Scan";
      patch.subType = "Intranet Scan";
      patch.scopes = "127.0.0.1/32";
      patch.content = "3000,8080,80";
      patch.endpoint = "/,/mcp,/sse,/mcp/sse";
    }

    if (defaultType) {
      patch.type = defaultType;
      if (nameNotUserEdited) {
        patch.name = getAutoProviderName(value, defaultType, "");
      }
      if (displayNameNotUserEdited) {
        patch.displayName = getAutoProviderDisplayName(value, defaultType, "");
      }
    }
    patchProvider(patch);
  };

  const onChangeType = (value: string) => {
    const patch: Record<string, any> = {type: value};
    if (value === "Local File System") {
      patch.domain = Setting.getFullServerUrl();
    } else if (value === "OIDC") {
      patch.scopes = "openid profile email";
    } else if (value.startsWith("Custom") && provider.category === "OAuth") {
      patch.customAuthUrl = "https://door.casdoor.com/login/oauth/authorize";
      patch.scopes = "openid profile email";
      patch.customTokenUrl = "https://door.casdoor.com/api/login/oauth/access_token";
      patch.customUserInfoUrl = "https://door.casdoor.com/api/userinfo";
    } else if (value === "Custom HTTP SMS") {
      patch.endpoint = "https://example.com/send-custom-http-sms";
      patch.method = "GET";
      patch.title = "code";
    } else if (value === "Custom HTTP Email") {
      patch.endpoint = "https://example.com/send-custom-http-email";
      patch.method = "POST";
    } else if (value === "Custom HTTP") {
      patch.method = "GET";
      patch.title = "";
    } else if (value === "MCP Scan") {
      patch.subType = "Intranet Scan";
      if (!provider.scopes) {
        patch.scopes = "127.0.0.1/32";
      }
      if (!provider.content) {
        patch.content = "3000,8080,80";
      }
      if (!provider.endpoint) {
        patch.endpoint = "/,/mcp,/sse,/mcp/sse";
      }
    } else if (value === "Security Scan") {
      patch.subType = "Site";
    }
    if (nameNotUserEdited) {
      patch.name = getAutoProviderName(provider.category, value, "");
    }
    if (displayNameNotUserEdited) {
      patch.displayName = getAutoProviderDisplayName(provider.category, value, "");
    }
    patchProvider(patch);
  };

  const renderMappingRows = (rows: {key: string; label: Label}[]) => (
    <div className="space-y-2">
      {rows.map((row) => (
        <div key={row.key} className="grid grid-cols-1 gap-1 sm:grid-cols-[minmax(120px,160px)_1fr] sm:items-center">
          <span className="text-sm text-muted-foreground">{row.label.label}</span>
          <Input
            value={provider.userMapping?.[row.key] ?? ""}
            onChange={(e) => updateUserMappingField(row.key, e.target.value)}
          />
        </div>
      ))}
    </div>
  );

  const renderUserMappingInput = () => renderMappingRows([
    {key: "id", label: label(i18next.t("general:ID"))},
    {key: "username", label: label(i18next.t("signup:Username"))},
    {key: "displayName", label: label(i18next.t("general:Display name"))},
    {key: "email", label: label(i18next.t("general:Email"))},
    {key: "avatarUrl", label: label(i18next.t("general:Avatar"))},
    {key: "phone", label: label(i18next.t("general:Phone"))},
    {key: "countryCode", label: label(i18next.t("user:Country code"))},
    {key: "firstName", label: label(i18next.t("general:First name"))},
    {key: "lastName", label: label(i18next.t("general:Last name"))},
    {key: "region", label: label(i18next.t("provider:Region"))},
    {key: "location", label: label(i18next.t("user:Location"))},
    {key: "affiliation", label: label(i18next.t("user:Affiliation"))},
    {key: "title", label: label(i18next.t("general:Title"))},
  ]);

  const renderEmailMappingInput = () => renderMappingRows([
    {key: "fromName", label: label(i18next.t("provider:From name"))},
    {key: "fromAddress", label: label(i18next.t("provider:From address"))},
    {key: "toAddress", label: label(i18next.t("provider:To address"))},
    {key: "subject", label: label(i18next.t("provider:Subject"))},
    {key: "content", label: label(i18next.t("provider:Email content"))},
  ]);

  const renderSmsMappingInput = () => renderMappingRows([
    {key: "phoneNumber", label: label(i18next.t("general:Phone"))},
    {key: "content", label: label(i18next.t("provider:Content"))},
  ]);

  const renderHttpHeaderTable = () => (
    <EditableTable
      rows={mapToRows(provider.httpHeaders, "name", "value")}
      onChange={(rows) => updateProviderField("httpHeaders", rowsToMap(rows, "name", "value"))}
      newRow={() => ({name: "", value: ""})}
      reorderable={false}
      columns={[
        {
          key: "name",
          title: i18next.t("general:Keys"),
          render: (row: any, _index, update) => (
            <Input value={row.name ?? ""} onChange={(e) => update({name: e.target.value})} />
          ),
        },
        {
          key: "value",
          title: i18next.t("user:Values"),
          render: (row: any, _index, update) => (
            <Input value={row.value ?? ""} onChange={(e) => update({value: e.target.value})} />
          ),
        },
      ]}
    />
  );

  const renderHttpRows = (methods: string[]) => (
    <React.Fragment>
      <FormRow label={i18next.t("general:Method")} tooltip={i18next.t("provider:Method - Tooltip")}>
        <SelectField
          value={provider.method}
          onChange={(v) => updateProviderField("method", v)}
          options={methods.map((m) => ({id: m, name: m}))}
        />
      </FormRow>
      {provider.method !== "GET" ? (
        <FormRow labelKey="webhook:Content type">
          <SelectField
            value={provider.issuerUrl === "" ? "application/x-www-form-urlencoded" : provider.issuerUrl}
            onChange={(v) => updateProviderField("issuerUrl", v)}
            options={CONTENT_TYPES.map((t) => ({id: t, name: t}))}
          />
        </FormRow>
      ) : null}
      <FormRow labelKey="provider:HTTP header" block>
        {renderHttpHeaderTable()}
      </FormRow>
    </React.Fragment>
  );

  const renderOAuthFields = () => (
    <React.Fragment>
      <FormRow labelKey="provider:Email regex">
        <Textarea rows={4} value={provider.emailRegex ?? ""} onChange={(e) => updateProviderField("emailRegex", e.target.value)} />
      </FormRow>
      <FormRow labelKey="provider:Enable proxy">
        <Switch checked={!!provider.enableProxy} onCheckedChange={(v) => updateProviderField("enableProxy", v)} />
      </FormRow>
      {provider.type === "WeChat" ? (
        <React.Fragment>
          <FormRow labelKey="provider:Use WeChat Media Platform in PC">
            <Switch
              disabled={!provider.clientId}
              checked={!!provider.disableSsl}
              onCheckedChange={(v) => updateProviderField("disableSsl", v)}
            />
          </FormRow>
          <FormRow labelKey="token:Access token">
            <Input
              value={provider.content ?? ""}
              disabled={!provider.disableSsl || !provider.clientId2}
              onChange={(e) => updateProviderField("content", e.target.value)}
            />
          </FormRow>
          <FormRow labelKey="provider:Follow-up action">
            <SelectField
              disabled={!provider.disableSsl || !provider.clientId || !provider.clientId2}
              value={provider.signName}
              onChange={(v) => updateProviderField("signName", v)}
              options={[
                {id: "open", name: i18next.t("provider:Use WeChat Open Platform to login")},
                {id: "media", name: i18next.t("provider:Use WeChat Media Platform to login")},
              ]}
            />
          </FormRow>
        </React.Fragment>
      ) : null}
      {["ADFS", "AzureAD", "AzureADB2C", "Okta", "Nextcloud"].includes(provider.type) || (provider.type === "Casdoor" || provider.category === "Storage") ? (
        <FormRow
          label={["AzureAD", "AzureADB2C"].includes(provider.type) ? i18next.t("provider:Tenant ID") : i18next.t("provider:Domain")}
          tooltip={["AzureAD", "AzureADB2C"].includes(provider.type) ? i18next.t("provider:Tenant ID - Tooltip") : i18next.t("provider:Domain - Tooltip")}
        >
          <Input value={provider.domain ?? ""} onChange={(e) => updateProviderField("domain", e.target.value)} />
        </FormRow>
      ) : null}
      {provider.type === "Google" || provider.type === "Lark" ? (
        <FormRow
          label={provider.type === "Google" ? i18next.t("provider:Get phone number") : i18next.t("provider:Use global endpoint")}
          tooltip={provider.type === "Google" ? i18next.t("provider:Get phone number - Tooltip") : i18next.t("provider:Use global endpoint - Tooltip")}
        >
          <Switch disabled={!provider.clientId} checked={!!provider.disableSsl} onCheckedChange={(v) => updateProviderField("disableSsl", v)} />
        </FormRow>
      ) : null}
      {provider.type === "Alipay" ? (
        <React.Fragment>
          <FormRow labelKey="general:Cert">
            <SearchableSelect value={provider.cert ?? ""} onChange={(v) => updateProviderField("cert", v)} options={certOptions} />
          </FormRow>
          <FormRow labelKey="general:Root cert">
            <SearchableSelect value={provider.metadata ?? ""} onChange={(v) => updateProviderField("metadata", v)} options={certOptions} />
          </FormRow>
        </React.Fragment>
      ) : null}
      {provider.type === "OIDC" ? (
        <FormRow labelKey="provider:Issuer URL">
          <div className="flex flex-wrap items-center gap-2">
            <Input className="w-96 max-w-full" value={provider.domain ?? ""} onChange={(e) => updateProviderField("domain", e.target.value)} />
            <Button loading={discoveryLoading} disabled={!provider.domain} onClick={fetchOidcDiscovery}>
              {i18next.t("general:Request")}
            </Button>
          </div>
        </FormRow>
      ) : null}
      {Setting.isCustomOAuthType(provider.type) ? (
        <React.Fragment>
          <FormRow labelKey="provider:Auth URL">
            <Input value={provider.customAuthUrl ?? ""} onChange={(e) => updateProviderField("customAuthUrl", e.target.value)} />
          </FormRow>
          <FormRow labelKey="provider:Token URL">
            <Input value={provider.customTokenUrl ?? ""} onChange={(e) => updateProviderField("customTokenUrl", e.target.value)} />
          </FormRow>
          <FormRow labelKey="provider:Scope">
            <Input value={provider.scopes ?? ""} onChange={(e) => updateProviderField("scopes", e.target.value)} />
          </FormRow>
          <FormRow labelKey="provider:UserInfo URL">
            <Input value={provider.customUserInfoUrl ?? ""} onChange={(e) => updateProviderField("customUserInfoUrl", e.target.value)} />
          </FormRow>
          <FormRow labelKey="provider:Logout URL">
            <Input value={provider.customLogoutUrl ?? ""} onChange={(e) => updateProviderField("customLogoutUrl", e.target.value)} />
          </FormRow>
          <FormRow labelKey="provider:Enable PKCE">
            <Switch checked={!!provider.enablePkce} onCheckedChange={(v) => updateProviderField("enablePkce", v)} />
          </FormRow>
          <FormRow labelKey="provider:User mapping" block>
            {renderUserMappingInput()}
          </FormRow>
          <FormRow labelKey="general:Favicon" block>
            <div className="space-y-2">
              <Input value={provider.customLogo ?? ""} onChange={(e) => updateProviderField("customLogo", e.target.value)} />
              {provider.customLogo ? (
                <a target="_blank" rel="noreferrer" href={provider.customLogo}>
                  <img src={provider.customLogo} alt={provider.customLogo} height={90} className="h-[90px]" />
                </a>
              ) : null}
            </div>
          </FormRow>
        </React.Fragment>
      ) : null}
    </React.Fragment>
  );

  const renderEmailFields = () => (
    <React.Fragment>
      {["Custom HTTP Email", "SendGrid"].includes(provider.type) ? (
        <FormRow label={i18next.t("provider:Endpoint")} tooltip={i18next.t("provider:Region endpoint for Internet")}>
          <Input value={provider.endpoint ?? ""} onChange={(e) => updateProviderField("endpoint", e.target.value)} />
        </FormRow>
      ) : null}
      {provider.type !== "Resend" ? (
        <FormRow label={i18next.t("general:Host")} tooltip={i18next.t("provider:Host - Tooltip")}>
          <Input value={provider.host ?? ""} onChange={(e) => updateProviderField("host", e.target.value)} />
        </FormRow>
      ) : null}
      {!["Azure ACS", "SendGrid", "Resend"].includes(provider.type) ? (
        <React.Fragment>
          <FormRow label={i18next.t("general:Port")} tooltip={i18next.t("provider:Port - Tooltip")}>
            <Input type="number" value={provider.port ?? 0} onChange={(e) => updateProviderField("port", e.target.value)} />
          </FormRow>
          <FormRow labelKey="provider:SSL mode">
            <SelectField
              value={provider.sslMode || "Auto"}
              onChange={(v) => updateProviderField("sslMode", v)}
              options={[
                {id: "Auto", name: i18next.t("general:Auto")},
                {id: "Enable", name: i18next.t("general:Enable")},
                {id: "Disable", name: i18next.t("general:Disable")},
              ]}
            />
          </FormRow>
        </React.Fragment>
      ) : null}
      <FormRow labelKey="provider:Enable proxy">
        <Switch checked={!!provider.enableProxy} onCheckedChange={(v) => updateProviderField("enableProxy", v)} />
      </FormRow>
      {provider.type === "Custom HTTP Email" ? (
        <React.Fragment>
          {renderHttpRows(HTTP_METHODS)}
          {provider.method !== "GET" ? (
            <FormRow labelKey="provider:HTTP body mapping" block>
              {renderEmailMappingInput()}
            </FormRow>
          ) : null}
        </React.Fragment>
      ) : null}
      <FormRow block labelKey="provider:Email title">
        <Input value={provider.title ?? ""} onChange={(e) => updateProviderField("title", e.target.value)} />
      </FormRow>
      <FormRow labelKey="provider:Email content" block>
        <div className="space-y-2">
          <div className="flex flex-wrap gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => updateProviderField("content", "You have requested a verification code at Casdoor. Here is your code: %s, please enter in 5 minutes. <reset-link>Or click %link to reset</reset-link>")}
            >
              {i18next.t("general:Reset to Default")} (Text)
            </Button>
            <Button size="sm" onClick={() => updateProviderField("content", Setting.getDefaultHtmlEmailContent())}>
              {i18next.t("general:Reset to Default")} (HTML)
            </Button>
          </div>
          <div className="grid gap-4 lg:grid-cols-2">
            <CodeEditor
              language="html"
              height={300}
              value={provider.content ?? ""}
              onChange={(v) => updateProviderField("content", v)}
            />
            <div
              className="overflow-auto rounded-md border bg-background p-3"
              dangerouslySetInnerHTML={{
                __html: String(provider.content ?? "").replace("%s", "123456").replace("%{user.friendlyName}", account ? Setting.getFriendlyUserName(account) : ""),
              }}
            />
          </div>
        </div>
      </FormRow>
      <FormRow
        label={`${i18next.t("provider:Email content")}-${i18next.t("general:Invitations")}`}
        tooltip={i18next.t("provider:Email content - Tooltip")}
        block
      >
        <div className="space-y-2">
          <div className="flex flex-wrap gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => updateProviderField("metadata", "You have invited to join Casdoor. Here is your invitation code: %s, please enter in 5 minutes. Or click %link to signup")}
            >
              {i18next.t("general:Reset to Default")} (Text)
            </Button>
            <Button size="sm" onClick={() => updateProviderField("metadata", Setting.getDefaultInvitationHtmlEmailContent())}>
              {i18next.t("general:Reset to Default")} (HTML)
            </Button>
          </div>
          <div className="grid gap-4 lg:grid-cols-2">
            <CodeEditor
              language="html"
              height={300}
              value={provider.metadata ?? ""}
              onChange={(v) => updateProviderField("metadata", v)}
            />
            <div
              className="overflow-auto rounded-md border bg-background p-3"
              dangerouslySetInnerHTML={{__html: String(provider.metadata ?? "").replace("%code", "123456").replace("%s", "123456")}}
            />
          </div>
        </div>
      </FormRow>
      <FormRow labelKey="provider:Test Email">
        <div className="flex flex-wrap items-center gap-2">
          <Input
            className="w-64"
            value={provider.receiver ?? ""}
            placeholder={i18next.t("user:Input your email")}
            onChange={(e) => updateProviderField("receiver", e.target.value)}
          />
          {!["Azure ACS", "SendGrid", "Resend"].includes(provider.type) ? (
            <Button variant="outline" onClick={() => ProviderTest.connectSmtpServer(provider)}>
              {i18next.t("provider:Test SMTP Connection")}
            </Button>
          ) : null}
          <Button
            disabled={!Setting.isValidEmail(provider.receiver ?? "")}
            onClick={() => ProviderTest.sendTestEmail(provider, provider.receiver)}
          >
            {i18next.t("provider:Send Testing Email")}
          </Button>
        </div>
      </FormRow>
    </React.Fragment>
  );

  const renderSmsFields = () => (
    <React.Fragment>
      {!SMS_PROVIDERS_WITHOUT_SIGN_NAME.includes(provider.type) ? (
        <FormRow labelKey="provider:Sign Name">
          <Input value={provider.signName ?? ""} onChange={(e) => updateProviderField("signName", e.target.value)} />
        </FormRow>
      ) : null}
      {!SMS_PROVIDERS_WITHOUT_TEMPLATE_CODE.includes(provider.type) ? (
        <FormRow labelKey="provider:Template code">
          <Input value={provider.templateCode ?? ""} onChange={(e) => updateProviderField("templateCode", e.target.value)} />
        </FormRow>
      ) : null}
      {provider.type === "Custom HTTP SMS" ? (
        <React.Fragment>
          <FormRow label={i18next.t("provider:Endpoint")} tooltip={i18next.t("provider:Region endpoint for Internet")}>
            <Input value={provider.endpoint ?? ""} onChange={(e) => updateProviderField("endpoint", e.target.value)} />
          </FormRow>
          {renderHttpRows(HTTP_METHODS)}
          {provider.method !== "GET" ? (
            <FormRow labelKey="provider:HTTP body mapping" block>
              {renderSmsMappingInput()}
            </FormRow>
          ) : null}
          <FormRow labelKey="provider:Parameter">
            <Input value={provider.title ?? ""} onChange={(e) => updateProviderField("title", e.target.value)} />
          </FormRow>
        </React.Fragment>
      ) : null}
      <FormRow labelKey="provider:Enable proxy">
        <Switch checked={!!provider.enableProxy} onCheckedChange={(v) => updateProviderField("enableProxy", v)} />
      </FormRow>
      <FormRow labelKey="provider:SMS Test">
        <div className="flex flex-wrap items-center gap-2">
          <div className="w-32 shrink-0">
            <SearchableSelect
              value={provider.content ?? ""}
              onChange={(v) => updateProviderField("content", v)}
              options={Setting.getCountryCodeData(account?.organization?.countryCodes).map((country: any) => ({
                value: country.code,
                label: `+${country.phone}`,
                keywords: `${country.name} ${country.code} ${country.phone}`,
              }))}
            />
          </div>
          <Input
            className="w-48"
            value={provider.receiver ?? ""}
            placeholder={i18next.t("user:Input your phone number")}
            onChange={(e) => updateProviderField("receiver", e.target.value)}
          />
          <Button
            disabled={!Setting.isValidPhone(provider.receiver ?? "") || (provider.type === "Custom HTTP SMS" && provider.endpoint === "")}
            onClick={() => ProviderTest.sendTestSms(provider, "+" + Setting.getCountryCode(provider.content) + provider.receiver)}
          >
            {i18next.t("provider:Send Testing SMS")}
          </Button>
        </div>
      </FormRow>
    </React.Fragment>
  );

  const renderNotificationFields = () => {
    const receiverLabel = getReceiverLabel(provider);
    return (
      <React.Fragment>
        {provider.type === "CUCloud" ? (
          <FormRow labelKey="provider:Region ID">
            <Input value={provider.regionId ?? ""} onChange={(e) => updateProviderField("regionId", e.target.value)} />
          </FormRow>
        ) : null}
        {provider.type === "Custom HTTP" ? (
          <FormRow label={i18next.t("general:Method")} tooltip={i18next.t("provider:Method - Tooltip")}>
            <SelectField
              value={provider.method}
              onChange={(v) => updateProviderField("method", v)}
              options={[{id: "GET", name: "GET"}, {id: "POST", name: "POST"}]}
            />
          </FormRow>
        ) : null}
        {["Custom HTTP", "CUCloud"].includes(provider.type) ? (
          <FormRow labelKey="provider:Parameter">
            <Input value={provider.title ?? ""} onChange={(e) => updateProviderField("title", e.target.value)} />
          </FormRow>
        ) : null}
        {["Google Chat", "CUCloud"].includes(provider.type) ? (
          <FormRow labelKey="provider:Metadata">
            <Textarea rows={4} value={provider.metadata ?? ""} onChange={(e) => updateProviderField("metadata", e.target.value)} />
          </FormRow>
        ) : null}
        <FormRow labelKey="provider:Content">
          <Textarea rows={3} value={provider.content ?? ""} onChange={(e) => updateProviderField("content", e.target.value)} />
        </FormRow>
        <FormRow label={receiverLabel ? receiverLabel.label : "Test Notification"} tooltip={receiverLabel?.tooltip}>
          <div className="flex flex-wrap items-center gap-2">
            {receiverLabel ? (
              <Input
                className="w-72"
                value={provider.receiver ?? ""}
                onChange={(e) => updateProviderField("receiver", e.target.value)}
              />
            ) : null}
            <Button onClick={() => ProviderTest.sendTestNotification(provider)}>
              {i18next.t("provider:Send Testing Notification")}
            </Button>
          </div>
        </FormRow>
      </React.Fragment>
    );
  };

  const renderMfaFields = () => (
    <React.Fragment>
      <FormRow label={i18next.t("general:Host")} tooltip={i18next.t("provider:Host - Tooltip")}>
        <Input value={provider.host ?? ""} placeholder="10.10.10.10" onChange={(e) => updateProviderField("host", e.target.value)} />
      </FormRow>
      <FormRow label={i18next.t("general:Port")} tooltip={i18next.t("provider:Port - Tooltip")}>
        <Input type="number" value={provider.port ?? 0} onChange={(e) => updateProviderField("port", e.target.value)} />
      </FormRow>
      <FormRow label={i18next.t("provider:Client secret")} tooltip={i18next.t("provider:RADIUS Shared Secret - Tooltip")}>
        <Input value={provider.clientSecret ?? ""} placeholder="Shared secret" onChange={(e) => updateProviderField("clientSecret", e.target.value)} />
      </FormRow>
    </React.Fragment>
  );

  const renderLogFields = () => {
    const storageProviders = providers.filter((item: any) =>
      item.category === "Storage" &&
      (!provider.owner || item.owner === provider.owner) &&
      (typeof item.state !== "string" || item.state.toLowerCase() !== "disabled"),
    );

    return (
      <React.Fragment>
        {provider.type === "Agent" && provider.subType === "OpenClaw" ? (
          <React.Fragment>
            <FormRow label={i18next.t("general:Host")} tooltip={i18next.t("provider:Host - Tooltip")}>
              <Input value={provider.host ?? ""} onChange={(e) => updateProviderField("host", e.target.value)} />
            </FormRow>
            <FormRow labelKey="provider:Agent ID">
              <Input value={provider.title ?? ""} onChange={(e) => updateProviderField("title", e.target.value)} />
            </FormRow>
            <FormRow label={i18next.t("general:Path")}>
              <Input value={provider.endpoint ?? ""} onChange={(e) => updateProviderField("endpoint", e.target.value)} />
            </FormRow>
            <FormRow label={i18next.t("provider:Storage provider")}>
              <SearchableSelect
                value={provider.providerUrl ?? ""}
                onChange={(v) => updateProviderField("providerUrl", v)}
                options={storageProviders.map((item: any) => ({
                  value: item.name,
                  label: item.displayName || item.name,
                }))}
              />
            </FormRow>
          </React.Fragment>
        ) : null}
        <FormRow labelKey="general:State">
          <SelectField
            value={provider.state || "Enabled"}
            onChange={(v) => updateProviderField("state", v)}
            options={[
              {id: "Enabled", name: i18next.t("general:Enabled")},
              {id: "Disabled", name: i18next.t("general:Disabled")},
            ]}
          />
        </FormRow>
      </React.Fragment>
    );
  };

  const renderSamlFields = () => (
    <React.Fragment>
      <FormRow labelKey="provider:Sign request">
        <Switch checked={!!provider.enableSignAuthnRequest} onCheckedChange={(v) => updateProviderField("enableSignAuthnRequest", v)} />
      </FormRow>
      <FormRow labelKey="provider:Metadata url">
        <div className="flex flex-wrap items-center gap-2">
          <Input className="w-96 max-w-full" value={requestUrl} onChange={(e) => setRequestUrl(e.target.value)} />
          <Button loading={metadataLoading} onClick={fetchSamlMetadata}>
            {i18next.t("general:Request")}
          </Button>
        </div>
      </FormRow>
      <FormRow labelKey="provider:Metadata">
        <div className="space-y-2">
          <Textarea rows={4} value={provider.metadata ?? ""} onChange={(e) => updateProviderField("metadata", e.target.value)} />
          <Button onClick={() => parseSamlMetadata(provider.metadata ?? "")}>
            {i18next.t("provider:Parse")}
          </Button>
        </div>
      </FormRow>
      <FormRow label={i18next.t("provider:Endpoint")} tooltip={i18next.t("provider:SAML 2.0 Endpoint (HTTP)")}>
        <Input value={provider.endpoint ?? ""} onChange={(e) => updateProviderField("endpoint", e.target.value)} />
      </FormRow>
      <FormRow label={i18next.t("provider:IdP")} tooltip={i18next.t("provider:IdP certificate")}>
        <Input value={provider.idP ?? ""} onChange={(e) => updateProviderField("idP", e.target.value)} />
      </FormRow>
      <FormRow labelKey="provider:Issuer URL">
        <Input value={provider.issuerUrl ?? ""} onChange={(e) => updateProviderField("issuerUrl", e.target.value)} />
      </FormRow>
      {["provider:SP ACS URL", "provider:SP Entity ID"].map((key) => (
        <FormRow key={key} labelKey={key}>
          <div className="flex items-center gap-2">
            <Input readOnly value={`${authConfig.serverUrl}/api/acs`} />
            <Button
              onClick={() => {
                copy(`${authConfig.serverUrl}/api/acs`);
                Setting.showMessage("success", i18next.t("general:Copied to clipboard successfully"));
              }}
            >
              {i18next.t("general:Copy")}
            </Button>
          </div>
        </FormRow>
      ))}
    </React.Fragment>
  );

  const renderScanFields = () => {
    const canScan = mode !== "add";
    if (provider.type === "MCP Scan" && provider.subType === "Intranet Scan") {
      return (
        <React.Fragment>
          <FormRow label={i18next.t("general:Host")}>
            <MultiSelect
              creatable
              value={toList(provider.scopes)}
              onChange={(v) => updateProviderField("scopes", normalizeAndJoin(v))}
              options={SCAN_HOST_OPTIONS.map((item) => ({value: item, label: item}))}
            />
          </FormRow>
          <FormRow label={i18next.t("general:Port")}>
            <MultiSelect
              creatable
              value={toList(provider.content)}
              onChange={(v) => updateProviderField("content", normalizeAndJoin(v))}
              options={SCAN_PORT_OPTIONS.map((item) => ({value: item, label: item}))}
            />
          </FormRow>
          <FormRow block label={i18next.t("general:Path")}>
            <MultiSelect
              creatable
              value={toList(provider.endpoint)}
              onChange={(v) => updateProviderField("endpoint", normalizeAndJoin(v))}
              options={SCAN_PATH_OPTIONS.map((item) => ({value: item, label: item}))}
            />
          </FormRow>
          <FormRow label="" block>
            <Button loading={scanLoading} disabled={!canScan} onClick={() => submitProviderScan()}>
              {i18next.t("server:Scan server")}
            </Button>
          </FormRow>
          {scanResult !== null ? (
            <FormRow
              block
              label={`${i18next.t("server:Scanned hosts")}:${scanResult?.scannedHosts ?? 0}, ${i18next.t("server:Online hosts")}:${scanResult?.onlineHosts?.length ?? 0}, ${i18next.t("server:Found servers")}:${scanServers.length}`}
            >
              <div className="rounded-lg border">
                <Table>
                  <TableHeader>
                    <TableRow className="hover:bg-transparent">
                      <TableHead className="w-[140px]">{i18next.t("general:Host")}</TableHead>
                      <TableHead className="w-[90px]">{i18next.t("general:Port")}</TableHead>
                      <TableHead className="w-[120px]">{i18next.t("general:Path")}</TableHead>
                      <TableHead>{i18next.t("general:URL")}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {scanServers.map((server: any, index: number) => (
                      <TableRow key={`${server.url}-${index}`}>
                        <TableCell>{server.host}</TableCell>
                        <TableCell>{server.port}</TableCell>
                        <TableCell>{server.path}</TableCell>
                        <TableCell>
                          {server.url ? (
                            <a target="_blank" rel="noreferrer" href={server.url} className="underline-offset-4 hover:underline">
                              {Setting.getShortText(server.url, 60)}
                            </a>
                          ) : null}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </FormRow>
          ) : null}
        </React.Fragment>
      );
    }

    if (provider.type === "Security Scan") {
      const findings = parseFindings(provider, scanResult);
      return (
        <React.Fragment>
          <FormRow block label={i18next.t("provider:Online list")}>
            <Input value={provider.endpoint ?? ""} onChange={(e) => updateProviderField("endpoint", e.target.value)} />
          </FormRow>
          {provider.subType === "Url" ? (
            <FormRow block label={i18next.t("general:URL")}>
              <Textarea
                rows={3}
                value={provider.content ?? ""}
                placeholder={"https://example.com\nhttps://another.example.com"}
                onChange={(e) => updateProviderField("content", e.target.value)}
              />
            </FormRow>
          ) : null}
          <FormRow label="" block>
            <Button
              loading={scanLoading}
              disabled={!canScan}
              onClick={() => submitProviderScan(provider.subType === "Url" ? provider.content : "")}
            >
              {i18next.t("general:Scan")}
            </Button>
          </FormRow>
          {findings.length > 0 ? (
            <FormRow block label={`${i18next.t("general:Scan")}: ${findings.length}`}>
              <div className="overflow-x-auto rounded-lg border">
                <Table>
                  <TableHeader>
                    <TableRow className="hover:bg-transparent">
                      <TableHead className="w-[160px]">{i18next.t("general:Name")}</TableHead>
                      <TableHead className="w-[160px]">{i18next.t("general:Product")}</TableHead>
                      <TableHead className="w-[160px]">{i18next.t("general:Vendor")}</TableHead>
                      <TableHead className="w-[140px]">{i18next.t("system:Version")}</TableHead>
                      <TableHead className="w-[120px]">{i18next.t("general:Severity")}</TableHead>
                      <TableHead>CVEs</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {findings.map((finding: any, index: number) => {
                      const entryPath = getEntryPath(provider.subType, provider.owner, finding?.name);
                      const cves = Array.isArray(finding?.cves) ? finding.cves : [];
                      return (
                        <TableRow key={`${finding?.targetUrl}-${finding?.name}-${index}`}>
                          <TableCell>
                            {entryPath ? (
                              <a href={entryPath} className="underline-offset-4 hover:underline">{finding.name}</a>
                            ) : finding?.name}
                          </TableCell>
                          <TableCell>{finding?.product}</TableCell>
                          <TableCell>{finding?.vendor}</TableCell>
                          <TableCell>{finding?.version}</TableCell>
                          <TableCell>{finding?.severity}</TableCell>
                          <TableCell>
                            {cves.length === 0 ? "0" : cves.map((cve: any, cveIndex: number) => {
                              const cveLabel = cve?.code || cve?.name || "-";
                              const link = getCveLink(cve);
                              const content = (
                                <React.Fragment>
                                  <div>{cveLabel}{cve?.severity ? ` (${cve.severity})` : ""}</div>
                                  {cve?.summary ? <div className="text-muted-foreground">{cve.summary}</div> : null}
                                </React.Fragment>
                              );
                              return (
                                <div key={`${cveLabel}-${cveIndex}`} className="mb-2 last:mb-0">
                                  {link ? <a target="_blank" rel="noreferrer" href={link} className="block">{content}</a> : content}
                                </div>
                              );
                            })}
                          </TableCell>
                        </TableRow>
                      );
                    })}
                  </TableBody>
                </Table>
              </div>
            </FormRow>
          ) : null}
        </React.Fragment>
      );
    }

    return null;
  };

  const renderPaymentFields = () => (
    <React.Fragment>
      {["Alipay", "WeChat Pay", "Casdoor"].includes(provider.type) ? (
        <FormRow labelKey="general:Cert">
          <SearchableSelect value={provider.cert ?? ""} onChange={(v) => updateProviderField("cert", v)} options={certOptions} />
        </FormRow>
      ) : null}
      {provider.type === "Alipay" ? (
        <FormRow labelKey="general:Root cert">
          <SearchableSelect value={provider.metadata ?? ""} onChange={(v) => updateProviderField("metadata", v)} options={certOptions} />
        </FormRow>
      ) : null}
      {["GC", "FastSpring"].includes(provider.type) ? (
        <FormRow label={i18next.t("general:Host")} tooltip={i18next.t("provider:Host - Tooltip")}>
          <Input value={provider.host ?? ""} onChange={(e) => updateProviderField("host", e.target.value)} />
        </FormRow>
      ) : null}
    </React.Fragment>
  );

  const getWalletValue = () => {
    try {
      const parsed = JSON.parse(provider.metadata);
      return Array.isArray(parsed) ? parsed : ["injected"];
    } catch {
      return ["injected"];
    }
  };

  const renderWeb3Fields = () => (
    <React.Fragment>
      <FormRow labelKey="provider:Enable proxy">
        <Switch checked={!!provider.enableProxy} onCheckedChange={(v) => updateProviderField("enableProxy", v)} />
      </FormRow>
      {provider.type === "Web3Onboard" ? (
        <FormRow labelKey="provider:Wallets">
          <MultiSelect
            value={getWalletValue()}
            onChange={(options) => updateProviderField("metadata", JSON.stringify(options))}
            options={WEB3_ONBOARD_WALLETS}
          />
        </FormRow>
      ) : null}
    </React.Fragment>
  );

  const renderStorageFields = () => (
    <React.Fragment>
      {!["Local File System", "MinIO", "Tencent Cloud COS", "Google Cloud Storage", "Qiniu Cloud Kodo", "Synology", "Casdoor"].includes(provider.type) ? (
        <FormRow label={i18next.t("provider:Endpoint (Intranet)")} tooltip={i18next.t("provider:Region endpoint for Intranet")}>
          <Input value={provider.intranetEndpoint ?? ""} onChange={(e) => updateProviderField("intranetEndpoint", e.target.value)} />
        </FormRow>
      ) : null}
      {provider.type !== "Local File System" ? (
        <React.Fragment>
          <FormRow label={i18next.t("provider:Endpoint")} tooltip={i18next.t("provider:Region endpoint for Internet")}>
            <Input value={provider.endpoint ?? ""} onChange={(e) => updateProviderField("endpoint", e.target.value)} />
          </FormRow>
          <FormRow
            label={provider.type === "Casdoor" ? i18next.t("general:Provider") : i18next.t("provider:Bucket")}
            tooltip={provider.type === "Casdoor" ? i18next.t("general:Provider - Tooltip") : i18next.t("provider:Bucket - Tooltip")}
          >
            <Input value={provider.bucket ?? ""} onChange={(e) => updateProviderField("bucket", e.target.value)} />
          </FormRow>
        </React.Fragment>
      ) : null}
      <FormRow labelKey="provider:Path prefix">
        <Input value={provider.pathPrefix ?? ""} onChange={(e) => updateProviderField("pathPrefix", e.target.value)} />
      </FormRow>
      {!["Synology", "Casdoor"].includes(provider.type) ? (
        <FormRow labelKey="provider:Domain">
          <Input
            value={provider.domain ?? ""}
            disabled={provider.type === "Local File System"}
            onChange={(e) => updateProviderField("domain", e.target.value)}
          />
        </FormRow>
      ) : null}
      {provider.type === "Casdoor" ? (
        <FormRow labelKey="general:Organization">
          <Input value={provider.content ?? ""} onChange={(e) => updateProviderField("content", e.target.value)} />
        </FormRow>
      ) : null}
      {["AWS S3", "Tencent Cloud COS", "Qiniu Cloud Kodo", "Casdoor", "CUCloud OSS", "MinIO"].includes(provider.type) ? (
        <FormRow
          label={provider.type === "Casdoor" ? i18next.t("general:Application") : i18next.t("provider:Region ID")}
          tooltip={provider.type === "Casdoor" ? i18next.t("general:Application - Tooltip") : i18next.t("provider:Region ID - Tooltip")}
        >
          <Input value={provider.regionId ?? ""} onChange={(e) => updateProviderField("regionId", e.target.value)} />
        </FormRow>
      ) : null}
    </React.Fragment>
  );

  const renderEndpointOnlyField = () => (
    <FormRow label={i18next.t("provider:Endpoint")} tooltip={i18next.t("provider:Region endpoint for Internet")}>
      <Input value={provider.endpoint ?? ""} onChange={(e) => updateProviderField("endpoint", e.target.value)} />
    </FormRow>
  );

  const appIdLabel = getAppIdLabel(provider);

  return (
    <EditPageShell
      grid
      title={mode === "add" ? i18next.t("provider:New Provider") : i18next.t("provider:Edit Provider")}
      mode={mode}
      backTo="/providers"
      saving={saving}
      onSave={save}
    >
      <FormRow labelKey="general:Name">
        <Input
          value={provider.name ?? ""}
          onChange={(e) => {
            setNameNotUserEdited(false);
            updateProviderField("name", e.target.value);
          }}
        />
      </FormRow>
      <FormRow labelKey="general:Display name">
        <Input
          value={provider.displayName ?? ""}
          onChange={(e) => {
            setDisplayNameNotUserEdited(false);
            updateProviderField("displayName", e.target.value);
          }}
        />
      </FormRow>
      <FormRow labelKey="general:Organization">
        <SearchableSelect
          disabled={!Setting.isAdminUser(account)}
          value={provider.owner ?? ""}
          onChange={(v) => updateProviderField("owner", v)}
          options={[
            ...(Setting.isAdminUser(account) ? [{value: "admin", label: i18next.t("provider:admin (Shared)")}] : []),
            ...organizations.map((organization: any) => ({value: organization.name, label: organization.name})),
          ]}
        />
      </FormRow>
      <FormRow labelKey="general:Category">
        <SearchableSelect
          value={provider.category ?? ""}
          onChange={onChangeCategory}
          options={CATEGORIES.map((item) => ({value: item, label: item}))}
        />
      </FormRow>
      <FormRow labelKey="general:Type">
        <SearchableSelect
          value={provider.type ?? ""}
          onChange={onChangeType}
          options={(Setting.getProviderTypeOptions(provider.category) as any[])
            .sort((a, b) => a.name.localeCompare(b.name))
            .map((providerType: any) => ({
              value: providerType.id,
              label: (
                <span className="flex items-center gap-2">
                  <img
                    width={20}
                    height={20}
                    src={Setting.getProviderLogoURL({category: provider.category, type: providerType.id})}
                    alt={providerType.id}
                  />
                  {providerType.name}
                </span>
              ),
              keywords: providerType.name,
            }))}
        />
      </FormRow>
      {subTypeOptions.length > 0 ? (
        <React.Fragment>
          <FormRow labelKey="provider:Sub type">
            <SelectField
              value={provider.subType}
              onChange={(v) => {
                const patch: Record<string, any> = {subType: v};
                if (nameNotUserEdited) {
                  patch.name = getAutoProviderName(provider.category, provider.type, v);
                }
                if (displayNameNotUserEdited) {
                  patch.displayName = getAutoProviderDisplayName(provider.category, provider.type, v);
                }
                patchProvider(patch);
              }}
              options={subTypeOptions}
            />
          </FormRow>
          {provider.type === "WeCom" ? (
            <React.Fragment>
              <FormRow label={i18next.t("general:Method")} tooltip={i18next.t("provider:Method - Tooltip")}>
                <SelectField
                  value={provider.method}
                  onChange={(v) => updateProviderField("method", v)}
                  options={[
                    {id: "Normal", name: i18next.t("application:Normal")},
                    {id: "Silent", name: i18next.t("provider:Silent")},
                  ]}
                />
              </FormRow>
              <FormRow labelKey="provider:Scope">
                <SelectField
                  value={provider.scopes}
                  onChange={(v) => updateProviderField("scopes", v)}
                  options={[
                    {id: "snsapi_userinfo", name: "snsapi_userinfo"},
                    {id: "snsapi_privateinfo", name: "snsapi_privateinfo"},
                  ]}
                />
              </FormRow>
              <FormRow labelKey="provider:Use id as name">
                <Switch checked={!!provider.disableSsl} onCheckedChange={(v) => updateProviderField("disableSsl", v)} />
              </FormRow>
            </React.Fragment>
          ) : null}
        </React.Fragment>
      ) : null}

      {provider.category === "OAuth" ? renderOAuthFields() : null}

      {hasCredentialRows(provider) ? (
        <React.Fragment>
          {hasClientIdRow(provider) ? (
            <FormRow label={getClientIdLabel(provider).label} tooltip={getClientIdLabel(provider).tooltip}>
              <Input value={provider.clientId ?? ""} onChange={(e) => updateProviderField("clientId", e.target.value)} />
            </FormRow>
          ) : null}
          <FormRow label={getClientSecretLabel(provider).label} tooltip={getClientSecretLabel(provider).tooltip}>
            <Input value={provider.clientSecret ?? ""} onChange={(e) => updateProviderField("clientSecret", e.target.value)} />
          </FormRow>
        </React.Fragment>
      ) : null}

      {hasCredential2Rows(provider) ? (
        <React.Fragment>
          <FormRow label={getClientId2Label(provider).label} tooltip={getClientId2Label(provider).tooltip}>
            <Input value={provider.clientId2 ?? ""} onChange={(e) => updateProviderField("clientId2", e.target.value)} />
          </FormRow>
          {["WeChat Pay", "CUCloud"].includes(provider.type) || (provider.category === "Email" && provider.type === "Azure ACS") ? null : (
            <FormRow label={getClientSecret2Label(provider).label} tooltip={getClientSecret2Label(provider).tooltip}>
              {provider.category === "OAuth" && provider.type === "Apple" ? (
                <Textarea rows={4} value={provider.clientSecret2 ?? ""} onChange={(e) => updateProviderField("clientSecret2", e.target.value)} />
              ) : (
                <Input value={provider.clientSecret2 ?? ""} onChange={(e) => updateProviderField("clientSecret2", e.target.value)} />
              )}
            </FormRow>
          )}
        </React.Fragment>
      ) : null}

      {appIdLabel ? (
        <FormRow label={appIdLabel.label} tooltip={appIdLabel.tooltip}>
          <Input value={provider.appId ?? ""} onChange={(e) => updateProviderField("appId", e.target.value)} />
        </FormRow>
      ) : null}

      {provider.category === "Notification" ? renderNotificationFields() : null}
      {provider.category === "Email" ? renderEmailFields() : null}
      {provider.category === "SMS" ? renderSmsFields() : null}
      {provider.category === "MFA" ? renderMfaFields() : null}
      {provider.category === "Log" ? renderLogFields() : null}
      {provider.category === "Scan" ? renderScanFields() : null}
      {provider.category === "SAML" ? renderSamlFields() : null}
      {provider.category === "Payment" ? renderPaymentFields() : null}
      {provider.category === "Web3" ? renderWeb3Fields() : null}
      {provider.category === "Storage" ? renderStorageFields() : null}
      {provider.category === "Face ID" ? renderEndpointOnlyField() : null}
      {provider.category === "ID Verification" ? renderEndpointOnlyField() : null}

      {provider.category !== "Log" ? (
        <FormRow block labelKey="provider:Provider URL">
          <Input value={provider.providerUrl ?? ""} onChange={(e) => updateProviderField("providerUrl", e.target.value)} />
        </FormRow>
      ) : null}

      {provider.category === "Captcha" ? (
        <FormRow block labelKey="general:Preview">
          <CaptchaPreview
            owner={provider.owner}
            name={name}
            provider={provider}
            captchaType={provider.type}
            subType={provider.subType}
            clientId={provider.clientId}
            clientSecret={provider.clientSecret}
            clientId2={provider.clientId2}
            clientSecret2={provider.clientSecret2}
            providerUrl={provider.providerUrl}
          />
        </FormRow>
      ) : null}
    </EditPageShell>
  );
}
