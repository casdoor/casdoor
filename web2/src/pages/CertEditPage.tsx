import i18next from "i18next";
import copy from "copy-to-clipboard";
import FileSaver from "file-saver";
import {useParams} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {CodeEditor} from "@/components/common/CodeEditor";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {useAccount} from "@/hooks/use-account";
import {useOrganizationOptions} from "@/hooks/use-options";
import * as CertBackend from "@/backend/CertBackend";
import * as Setting from "@/lib/setting";

const TYPES = ["SSL", "x509", "Payment"];

/** SSL certs are issued by an ACME provider, so they only pick the key family */
const SSL_ALGORITHMS = [
  {value: "RSA", label: "RSA"},
  {value: "ECC", label: "ECC"},
];

const JWT_ALGORITHMS = [
  {value: "RS256", label: "RS256 (RSA + SHA256)"},
  {value: "RS384", label: "RS384 (RSA + SHA384)"},
  {value: "RS512", label: "RS512 (RSA + SHA512)"},
  {value: "ES256", label: "ES256 (ECDSA using P-256 + SHA256)"},
  {value: "ES384", label: "ES384 (ECDSA using P-384 + SHA384)"},
  {value: "ES512", label: "ES512 (ECDSA using P-521 + SHA512)"},
  {value: "PS256", label: "PS256 (RSASSA-PSS using SHA256 and MGF1 with SHA256)"},
  {value: "PS384", label: "PS384 (RSASSA-PSS using SHA384 and MGF1 with SHA384)"},
  {value: "PS512", label: "PS512 (RSASSA-PSS using SHA512 and MGF1 with SHA512)"},
];

/** the DNS providers Casdoor can solve the ACME challenge with */
const SSL_PROVIDERS = ["GoDaddy", "Aliyun"];

export default function CertEditPage() {
  const {organizationName = "", certName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();

  const isSsl = (ctx: {record: any}) => ctx.record.type === "SSL";

  const fields: EditField[] = [
    {
      type: "select",
      name: "owner",
      labelKey: "general:Organization",
      disabled: () => !Setting.isAdminUser(account),
      options: () =>
        Setting.isAdminUser(account)
          ? [{value: "admin", label: i18next.t("provider:admin (Shared)")}, ...organizations]
          : organizations,
    },
    {type: "text", name: "name", labelKey: "general:Name"},
    {type: "text", name: "displayName", labelKey: "general:Display name"},
    {
      type: "select",
      name: "scope",
      labelKey: "provider:Scope",
      options: () => [{value: "JWT", label: "JWT"}],
    },
    {
      type: "select",
      name: "type",
      labelKey: "general:Type",
      options: () => TYPES.map((item) => ({value: item, label: item})),
      onChange: (value, ctx, updateFields) => {
        if (value === "SSL") {
          // the SSL key pair is issued by the provider, never pasted in
          updateFields({type: value, cryptoAlgorithm: "RSA", certificate: "", privateKey: ""});
        } else if (ctx.record.type === "SSL") {
          // leaving SSL: drop the ACME credentials and the issued material
          updateFields({
            type: value,
            provider: "",
            account: "",
            accessKey: "",
            accessSecret: "",
            certificate: "",
            privateKey: "",
            expireTime: "",
            domainExpireTime: "",
          });
        } else {
          updateFields({type: value});
        }
      },
    },
    {
      type: "select",
      name: "cryptoAlgorithm",
      labelKey: "cert:Crypto algorithm",
      options: (ctx) => (isSsl(ctx) ? SSL_ALGORITHMS : JWT_ALGORITHMS),
      onChange: (value, ctx, updateFields) => {
        const bitSize = String(value).startsWith("ES")
          ? 0
          : ([1024, 2048, 4096].includes(Setting.myParseInt(ctx.record.bitSize)) ? ctx.record.bitSize : 2048);
        // the stored key pair no longer matches the algorithm, so it is regenerated
        updateFields({cryptoAlgorithm: value, bitSize, certificate: "", privateKey: ""});
      },
    },
    {
      type: "select",
      name: "bitSize",
      labelKey: "cert:Bit size",
      when: (ctx) => !isSsl(ctx) && !String(ctx.record.cryptoAlgorithm ?? "").startsWith("ES"),
      options: (ctx) =>
        Setting.getCryptoAlgorithmOptions(ctx.record.cryptoAlgorithm ?? "").map((item: any) => ({
          value: String(item.id),
          label: item.name,
        })),
      onChange: (value, _ctx, updateFields) =>
        updateFields({bitSize: Setting.myParseInt(value), certificate: "", privateKey: ""}),
    },
    {
      type: "number",
      name: "expireInYears",
      labelKey: "cert:Expire in years",
      when: (ctx) => !isSsl(ctx),
    },
    {
      type: "custom",
      name: "expireTime",
      labelKey: "general:Expire time",
      when: isSsl,
      render: (ctx) => <ReadOnlyDate value={ctx.record.expireTime} />,
    },
    {
      type: "custom",
      name: "domainExpireTime",
      labelKey: "cert:Domain expire",
      when: isSsl,
      render: (ctx) => <ReadOnlyDate value={ctx.record.domainExpireTime} />,
    },
    {
      type: "select",
      name: "provider",
      labelKey: "general:Provider",
      when: isSsl,
      options: () => SSL_PROVIDERS.map((item) => ({value: item, label: item})),
    },
    {type: "text", name: "account", labelKey: "cert:Account", when: isSsl},
    {type: "text", name: "accessKey", labelKey: "general:Access key", when: isSsl},
    {type: "password", name: "accessSecret", labelKey: "cert:Access secret", when: isSsl},
    {
      type: "custom",
      name: "certificate",
      labelKey: "cert:Certificate",
      block: true,
      render: (ctx, update) => (
        <PemField
          value={ctx.record.certificate}
          onChange={(value) => update("certificate", value)}
          copyLabel={i18next.t("cert:Copy certificate")}
          downloadLabel={i18next.t("cert:Download certificate")}
          fileName="token_jwt_key.pem"
        />
      ),
    },
    {
      type: "custom",
      name: "privateKey",
      labelKey: "cert:Private key",
      block: true,
      render: (ctx, update) => (
        <PemField
          value={ctx.record.privateKey}
          onChange={(value) => update("privateKey", value)}
          copyLabel={i18next.t("cert:Copy private key")}
          downloadLabel={i18next.t("cert:Download private key")}
          fileName="token_jwt_key.key"
        />
      ),
    },
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

function ReadOnlyDate({value}: {value: string | undefined}) {
  return (
    <div className="rounded-md border bg-muted/40 px-3 py-2 text-sm text-muted-foreground">
      {Setting.getFormattedDate(value) || "-"}
    </div>
  );
}

interface PemFieldProps {
  value: string | undefined;
  onChange: (value: string) => void;
  copyLabel: string;
  downloadLabel: string;
  fileName: string;
}

/** a PEM block with the copy / download buttons the antd cert page has */
function PemField({value, onChange, copyLabel, downloadLabel, fileName}: PemFieldProps) {
  const empty = !value;
  return (
    <div className="space-y-2">
      <div className="flex flex-wrap gap-2">
        <Button
          variant="outline"
          size="sm"
          disabled={empty}
          onClick={() => {
            copy(value ?? "");
            Setting.showMessage("success", i18next.t("general:Copied to clipboard successfully"));
          }}
        >
          {copyLabel}
        </Button>
        <Button
          size="sm"
          disabled={empty}
          onClick={() => FileSaver.saveAs(new Blob([value ?? ""], {type: "text/plain;charset=utf-8"}), fileName)}
        >
          {downloadLabel}
        </Button>
      </div>
      <CodeEditor value={value ?? ""} onChange={onChange} height={320} />
    </div>
  );
}
