import * as React from "react";
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
import {ConfirmButton} from "@/components/common/ConfirmButton";
import {EditionNotice} from "@/components/common/EditionNotice";

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
  {value: "SM2", label: "SM2 (SM2 + SM3, GB/T 32918)"},
];

/** the key size is fixed by the algorithm */
const isFixedKeySize = (algorithm: string) => algorithm.startsWith("ES") || algorithm === "SM2";

/** the key storage of a token signing cert: the database, or the hardware security module set in app.conf */
const KEY_STORAGE_HSM = "HSM";

/** the DNS providers Casdoor can solve the ACME challenge with */
const SSL_PROVIDERS = ["GoDaddy", "Aliyun"];

export default function CertEditPage() {
  const {organizationName = "", certName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();

  const isSsl = (ctx: {record: any}) => ctx.record.type === "SSL";
  const isHsm = (ctx: {record: any}) => ctx.record.type === "x509" && ctx.record.keyStorage === KEY_STORAGE_HSM;
  const isGlobalAdmin = Setting.isAdminUser(account);

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
    {type: "text", name: "name", labelKey: "general:Name", required: true},
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
        const bitSize = isFixedKeySize(String(value))
          ? 0
          : ([1024, 2048, 4096].includes(Setting.myParseInt(ctx.record.bitSize)) ? ctx.record.bitSize : 2048);
        // the stored key pair no longer matches the algorithm, so it is regenerated
        updateFields({cryptoAlgorithm: value, bitSize, certificate: "", privateKey: ""});
      },
    },
    {
      type: "custom",
      name: "edition",
      labelKey: "system:License",
      when: (ctx) => ctx.record.cryptoAlgorithm === "SM2" || isHsm(ctx),
      render: () => <EditionNotice edition="Enterprise" />,
    },
    {
      type: "select",
      name: "keyStorage",
      labelKey: "cert:Key storage",
      when: (ctx) => ctx.record.type === "x509",
      disabled: () => !isGlobalAdmin,
      options: () => [
        {value: "", label: i18next.t("syncer:Database")},
        {value: KEY_STORAGE_HSM, label: i18next.t("cert:Hardware security module (HSM)")},
      ],
      onChange: (value, _ctx, updateFields) =>
        // the key pair of the other storage no longer applies
        updateFields({keyStorage: value, certificate: "", privateKey: ""}),
    },
    {
      type: "text",
      name: "keyReference",
      labelKey: "cert:Key reference",
      when: isHsm,
      required: true,
      disabled: () => !isGlobalAdmin,
    },
    {
      type: "custom",
      name: "hsm",
      labelKey: "cert:HSM",
      when: isHsm,
      render: (ctx, update) => <HsmActions cert={ctx.record} canGenerate={isGlobalAdmin} update={update} />,
    },
    {
      type: "select",
      name: "bitSize",
      labelKey: "cert:Bit size",
      when: (ctx) => !isSsl(ctx) && !isHsm(ctx) && !isFixedKeySize(String(ctx.record.cryptoAlgorithm ?? "")),
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
      name: "keyId",
      labelKey: "provider:Key ID",
      when: (ctx) => !isSsl(ctx) && !!ctx.record.certificate,
      render: (ctx) => <KeyRotation cert={ctx.record} canRotate={!isHsm(ctx)} />,
    },
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
      render: (ctx, update) => isHsm(ctx) ? (
        <div className="rounded-md border bg-muted/40 px-3 py-2 text-sm text-muted-foreground">
          {i18next.t("cert:The private key is kept in the HSM and never leaves it")}
        </div>
      ) : (
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

/** The "kid" of the tokens the cert signs, and the rotation that gives the tokens a new key */
function KeyRotation({cert, canRotate}: {cert: any; canRotate: boolean}) {
  const [rotating, setRotating] = React.useState(false);
  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2" data-tour="rotate-key">
        <div className="flex-1 rounded-md border bg-muted/40 px-3 py-2 font-mono text-sm text-muted-foreground">
          {cert.keyId || cert.name}
        </div>
        {canRotate ? <ConfirmButton
          variant="outline"
          destructive={false}
          loading={rotating}
          title={i18next.t("cert:Rotate the key now? The tokens get a new key ID, and the previous public key stays published for the tokens it signed")}
          onConfirm={() => {
            setRotating(true);
            CertBackend.rotateCert(cert.owner, cert.name)
              .then((res: any) => {
                if (res.status === "ok") {
                  Setting.showMessage("success", i18next.t("general:Successfully modified"));
                  window.location.reload();
                } else {
                  Setting.showMessage("error", res.msg);
                }
              })
              .finally(() => setRotating(false));
          }}
        >
          {i18next.t("cert:Rotate key")}
        </ConfirmButton> : null}
      </div>
      {cert.rotatedTime ? (
        <p className="text-xs text-muted-foreground">
          {i18next.t("cert:Rotated time")}: {Setting.getFormattedDate(cert.rotatedTime)}, {i18next.t("cert:Previous key ID")}: {cert.previousKeyId}
        </p>
      ) : null}
      <EditionNotice />
    </div>
  );
}

interface HsmActionsProps {
  cert: any;
  /** only a global admin may choose a key in the HSM, which holds the keys of every organization */
  canGenerate: boolean;
  update: (field: string, value: any) => void;
}

/** The HSM set in app.conf, the certificate made from the HSM key and a test signature with it */
function HsmActions({cert, canGenerate, update}: HsmActionsProps) {
  const [status, setStatus] = React.useState<any>(null);
  const [generating, setGenerating] = React.useState(false);
  const [testing, setTesting] = React.useState(false);
  const [result, setResult] = React.useState<string>("");

  React.useEffect(() => {
    if (!canGenerate) {
      return;
    }
    CertBackend.getHsmStatus(false).then((res: any) => {
      if (res.status === "ok") {
        setStatus(res.data);
      }
    });
  }, [canGenerate]);

  const generate = () => {
    setGenerating(true);
    CertBackend.generateHsmCert({...cert, bitSize: Setting.myParseInt(cert.bitSize)})
      .then((res: any) => {
        if (res.status === "ok") {
          update("certificate", res.data);
          update("privateKey", "");
          Setting.showMessage("success", i18next.t("cert:The certificate is signed by the HSM key, save the cert to use it"));
        } else {
          Setting.showMessage("error", res.msg);
        }
      })
      .finally(() => setGenerating(false));
  };

  const test = () => {
    setTesting(true);
    setResult("");
    CertBackend.testHsmCert(cert.owner, cert.name)
      .then((res: any) => {
        if (res.status === "ok") {
          const data = res.data;
          setResult(`${data.algorithm}, ${data.keyType}, ${data.elapsedMs} ms`);
          Setting.showMessage("success", i18next.t("cert:The HSM signed a test token and the certificate verified it"));
        } else {
          Setting.showMessage("error", res.msg);
        }
      })
      .finally(() => setTesting(false));
  };

  return (
    <div className="space-y-2">
      {status ? (
        <p className="text-sm text-muted-foreground">
          {status.configured
            ? `PKCS #11: ${status.library}${status.token ? `, ${status.token}` : ""}`
            : i18next.t("cert:No HSM is set, set hsmType and hsmLibrary in conf/app.conf")}
        </p>
      ) : null}
      <div className="flex flex-wrap gap-2">
        {canGenerate ? (
          <Button size="sm" variant="outline" disabled={generating || !cert.keyReference} onClick={generate}>
            {i18next.t("cert:Generate certificate from the HSM key")}
          </Button>
        ) : null}
        <Button size="sm" variant="outline" disabled={testing || !cert.certificate} onClick={test}>
          {i18next.t("cert:Test signature")}
        </Button>
      </div>
      {result ? <p className="text-sm text-muted-foreground">{result}</p> : null}
      <p className="text-xs text-muted-foreground">
        {i18next.t("cert:The key reference is the label of the key in the PKCS #11 HSM. The test signs with the saved cert")}
      </p>
    </div>
  );
}
