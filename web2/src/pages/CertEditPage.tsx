import {useParams} from "react-router-dom";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useOrganizationOptions} from "@/hooks/use-options";
import * as CertBackend from "@/backend/CertBackend";
import * as Setting from "@/lib/setting";

const SCOPES = ["JWT", "CA"];
const CRYPTO_ALGORITHMS = ["RS256", "RS512", "ES256", "ES384", "ES512"];

export default function CertEditPage() {
  const {organizationName = "", certName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();

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
      name: "scope",
      labelKey: "provider:Scope",
      options: () => SCOPES.map((item) => ({value: item, label: item})),
    },
    {
      type: "select",
      name: "type",
      labelKey: "general:Type",
      options: (ctx) =>
        (ctx.record.scope === "JWT" ? ["x509"] : ["x509", "Let's Encrypt"]).map((item) => ({value: item, label: item})),
    },
    {
      type: "select",
      name: "cryptoAlgorithm",
      labelKey: "cert:Crypto algorithm",
      options: () => CRYPTO_ALGORITHMS.map((item) => ({value: item, label: item})),
    },
    {
      type: "select",
      name: "bitSize",
      labelKey: "cert:Bit size",
      when: (ctx) => !String(ctx.record.cryptoAlgorithm ?? "").startsWith("ES"),
      options: () => [1024, 2048, 4096].map((item) => ({value: String(item), label: String(item)})),
    },
    {type: "number", name: "expireInYears", labelKey: "cert:Expire in years"},
    {type: "code", name: "certificate", labelKey: "cert:Certificate", height: 260},
    {type: "code", name: "privateKey", labelKey: "cert:Private key", height: 260},
  ];

  return (
    <SimpleEditPage
      titleKey="cert:Edit Cert"
      backTo="/certs"
      deps={[organizationName, certName]}
      fields={fields}
      fetch={() => CertBackend.getCert(organizationName, certName)}
      add={(record) => CertBackend.addCert(record)}
      update={(record) => CertBackend.updateCert(organizationName, certName, record)}
      editUrl={(record) => `/certs/${record.owner}/${record.name}`}
      beforeSave={(record) => ({...record, bitSize: Setting.myParseInt(record.bitSize)})}
    />
  );
}
