import * as React from "react";
import i18next from "i18next";
import {RefreshCw} from "lucide-react";
import {useParams} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import {EditionNotice} from "@/components/common/EditionNotice";
import {useAccount} from "@/hooks/use-account";
import {useOrganizationOptions, useRoleNameOptions, useUserNameOptions} from "@/hooks/use-options";
import * as PrivilegedAccountBackend from "@/backend/PrivilegedAccountBackend";
import {EnumBadge, enumOptions} from "@/lib/enum-labels";
import {isPamRotatable, PAM_ACCOUNT_TYPES, PAM_DEFAULT_PORTS, PAM_PASSWORD_CHARSETS, PAM_ROTATION_STATES} from "@/lib/pam";
import * as Setting from "@/lib/setting";

/** A password that comes back as "***" stays as it is until a new one is typed. */
function SecretInput({value, onChange, disabled}: {value: string; onChange: (value: string) => void; disabled?: boolean}) {
  const saved = value === "***";
  return (
    <Input
      type="password"
      autoComplete="new-password"
      disabled={disabled}
      value={saved ? "" : value ?? ""}
      placeholder={saved ? i18next.t("user:Saved, type to change") : ""}
      onChange={(e) => onChange(e.target.value === "" && saved ? "***" : e.target.value)}
    />
  );
}

export default function PrivilegedAccountEditPage() {
  const {organizationName = "", accountName = ""} = useParams();
  const {account} = useAccount();
  const organizations = useOrganizationOptions();
  const users = useUserNameOptions(organizationName);
  const roles = useRoleNameOptions(organizationName);
  const [rotating, setRotating] = React.useState(false);

  const isRole = (ctx: any) => ctx.record.type === "Role";
  const isTarget = (ctx: any) => !isRole(ctx);
  const canRotate = (ctx: any) => isPamRotatable(ctx.record.type);
  const usesTls = (ctx: any) => ["MySQL", "PostgreSQL", "LDAP", "Active Directory"].includes(ctx.record.type);

  const rotate = async(ctx: any) => {
    setRotating(true);
    const res = await PrivilegedAccountBackend.rotatePrivilegedAccount(organizationName, accountName);
    setRotating(false);
    if (res.status === "ok") {
      Setting.showMessage("success", i18next.t("pam:The password has been rotated"));
    } else {
      Setting.showMessage("error", `${i18next.t("pam:Failed to rotate the password")}: ${res.msg}`);
    }
    ctx.reload();
  };

  const fields: EditField[] = [
    {
      type: "select",
      name: "owner",
      labelKey: "general:Organization",
      options: () => organizations,
      disabled: (ctx) => !Setting.isAdminUser(account) || ctx.mode !== "add",
    },
    {type: "text", name: "name", labelKey: "general:Name", required: true},
    {type: "text", name: "displayName", labelKey: "general:Display name"},
    {type: "textarea", name: "description", labelKey: "general:Description", rows: 2},
    {
      type: "select",
      name: "type",
      labelKey: "general:Type",
      options: () => enumOptions(PAM_ACCOUNT_TYPES),
      onChange: (value, _ctx, updateFields) => {
        const patch: Record<string, any> = {type: value, port: PAM_DEFAULT_PORTS[value] ?? 0};
        if (!isPamRotatable(value)) {
          patch.rotateAfterCheckin = false;
          patch.rotateIntervalDays = 0;
        }
        updateFields(patch);
      },
    },
    {type: "select", name: "role", labelKey: "general:Role", options: () => roles, when: isRole, required: true},
    {type: "text", name: "host", labelKey: "general:Host", when: isTarget},
    {type: "number", name: "port", labelKey: "general:Port", when: isTarget, min: 0, max: 65535},
    {type: "text", name: "database", labelKey: "syncer:Database", when: (ctx) => ctx.record.type === "PostgreSQL"},
    {type: "text", name: "username", labelKey: "pam:Account username", when: isTarget, required: true},
    {
      type: "custom",
      name: "secret",
      labelKey: "general:Password",
      when: isTarget,
      render: (ctx, update) => <SecretInput value={ctx.record.secret} onChange={(v) => update("secret", v)} disabled={ctx.mode === "view"} />,
    },
    {type: "text", name: "adminUsername", labelKey: "pam:Admin username for rotation", when: canRotate},
    {
      type: "custom",
      name: "adminSecret",
      labelKey: "pam:Admin password for rotation",
      when: (ctx) => canRotate(ctx) && !!ctx.record.adminUsername,
      render: (ctx, update) => <SecretInput value={ctx.record.adminSecret} onChange={(v) => update("adminSecret", v)} disabled={ctx.mode === "view"} />,
    },
    {type: "text", name: "hostKey", labelKey: "pam:SSH host key fingerprint", when: (ctx) => ctx.record.type === "Linux"},
    {type: "switch", name: "skipTlsVerify", labelKey: "pam:Skip TLS certificate verification", when: usesTls},
    {type: "multiselect", name: "users", labelKey: "pam:Who can request", options: () => users},
    {type: "switch", name: "requireApproval", labelKey: "pam:Require approval"},
    {type: "multiselect", name: "approvers", labelKey: "pam:Request approvers", options: () => users, when: (ctx) => !!ctx.record.requireApproval},
    {type: "number", name: "maxCheckoutMinutes", labelKey: "pam:Max checkout minutes", min: 1},
    {type: "switch", name: "rotateAfterCheckin", labelKey: "pam:Rotate after check-in", when: canRotate},
    {type: "number", name: "rotateIntervalDays", labelKey: "pam:Rotate every N days", when: canRotate, min: 0},
    {type: "number", name: "passwordLength", labelKey: "pam:Generated password length", when: canRotate, min: 12, max: 128},
    {type: "select", name: "passwordCharset", labelKey: "pam:Generated password characters", when: canRotate, options: () => enumOptions(PAM_PASSWORD_CHARSETS)},
    {
      type: "custom",
      name: "rotationState",
      labelKey: "pam:Last rotation",
      when: (ctx) => canRotate(ctx) && ctx.mode !== "add",
      render: (ctx) => (
        <div className="space-y-1 text-sm">
          <div className="flex flex-wrap items-center gap-2">
            {ctx.record.rotationState ? <EnumBadge map={PAM_ROTATION_STATES} value={ctx.record.rotationState} /> : <span className="text-muted-foreground">{i18next.t("pam:Never rotated")}</span>}
            {ctx.record.rotationTime ? <span className="text-muted-foreground">{Setting.getFormattedDate(ctx.record.rotationTime)}</span> : null}
            {ctx.record.needsRotation ? <span className="text-warning">{i18next.t("pam:Rotation is due after check-in")}</span> : null}
          </div>
          {ctx.record.lastRotatedTime ? (
            <div className="text-muted-foreground">{i18next.t("pam:Last rotated time")}: {Setting.getFormattedDate(ctx.record.lastRotatedTime)}</div>
          ) : null}
          {ctx.record.rotationMessage ? <div className="break-all text-destructive">{ctx.record.rotationMessage}</div> : null}
        </div>
      ),
    },
    {type: "switch", name: "isEnabled", labelKey: "general:Is enabled"},
  ];

  return (
    <div className="space-y-2">
      <EditionNotice edition="Enterprise" message={i18next.t("pam:Privileged account management needs the enterprise edition")} />
      <SimpleEditPage
        titleKey="pam:Edit Privileged Account"
        backTo="/privileged-accounts"
        deps={[organizationName, accountName]}
        fields={fields}
        fetch={() => PrivilegedAccountBackend.getPrivilegedAccount(organizationName, accountName)}
        add={(record) => PrivilegedAccountBackend.addPrivilegedAccount(record)}
        update={(record) => PrivilegedAccountBackend.updatePrivilegedAccount(organizationName, accountName, record)}
        editUrl={(record) => `/privileged-accounts/${record.owner}/${record.name}`}
        extraActions={(ctx) =>
          ctx.mode === "edit" && canRotate(ctx) ? (
            <Button variant="outline" loading={rotating} onClick={() => rotate(ctx)} data-tour="rotate-privileged-password">
              <RefreshCw className="size-4" />
              {i18next.t("pam:Rotate now")}
            </Button>
          ) : null
        }
      />
    </div>
  );
}
