import * as React from "react";
import i18next from "i18next";
import {useNavigate, useParams} from "react-router-dom";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Switch} from "@/components/ui/switch";
import {Loading} from "@/components/common/Loading";
import {SelectField} from "@/components/common/SelectField";
import {TagsInput} from "@/components/common/TagsInput";
import {EditPageShell} from "@/components/crud/EditPageShell";
import {EditableTable} from "@/components/crud/EditableTable";
import {FormRow} from "@/components/crud/FormRow";
import * as LdapBackend from "@/backend/LdapBackend";
import * as Setting from "@/lib/setting";

const PASSWORD_TYPES = ["Plain", "SSHA", "MD5"];

/** Create or edit one LDAP server of an organization. */
export default function LdapEditPage() {
  const {organizationName = "", ldapId = ""} = useParams();
  const navigate = useNavigate();
  const isNew = ldapId === "new";

  const [ldap, setLdap] = React.useState<any>(
    isNew
      ? {
        owner: organizationName,
        id: "",
        serverName: "",
        host: "",
        port: 389,
        enableSsl: false,
        allowSelfSignedCert: false,
        username: "",
        password: "",
        baseDn: "",
        filter: "",
        filterFields: ["uid", "mail", "mobile"],
        autoSync: 0,
        defaultGroups: [],
        enableGroups: false,
        passwordType: "Plain",
      }
      : null,
  );
  const [loading, setLoading] = React.useState(!isNew);
  const [saving, setSaving] = React.useState(false);

  React.useEffect(() => {
    if (isNew) {
      return;
    }
    LdapBackend.getLdap(organizationName, ldapId)
      .then((res: any) => {
        if (res.status === "ok") {
          setLdap(res.data);
        } else {
          Setting.showMessage("error", res.msg);
        }
      })
      .finally(() => setLoading(false));
  }, [organizationName, ldapId, isNew]);

  if (loading || ldap === null) {
    return <Loading />;
  }

  const update = (field: string, value: any) => setLdap((prev: any) => ({...prev, [field]: value}));

  const save = async(exitAfterSave: boolean) => {
    setSaving(true);
    const call = isNew ? LdapBackend.addLdap(ldap) : LdapBackend.updateLdap(ldap);
    try {
      const res: any = await call;
      if (res.status === "ok") {
        Setting.showMessage("success", i18next.t("general:Successfully saved"));
        if (exitAfterSave) {
          navigate(`/organizations/${organizationName}`);
        } else if (isNew && res.data?.id) {
          navigate(`/ldap/${organizationName}/${res.data.id}`, {replace: true});
        }
      } else {
        Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
      }
    } catch (error: any) {
      Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
    } finally {
      setSaving(false);
    }
  };

  return (
    <EditPageShell
      title={`${i18next.t("ldap:Edit LDAP")} - ${ldap.serverName || organizationName}`}
      mode={isNew ? "add" : "edit"}
      backTo={`/organizations/${organizationName}`}
      onSave={save}
      saving={saving}
      extraActions={
        !isNew ? (
          <Button variant="outline" onClick={() => navigate(`/ldap/sync/${organizationName}/${ldap.id}`)}>
            {i18next.t("general:Sync")}
          </Button>
        ) : null
      }
    >
      <FormRow labelKey="general:Organization">
        <Input value={ldap.owner ?? organizationName} disabled />
      </FormRow>
      <FormRow labelKey="ldap:Server name">
        <Input value={ldap.serverName ?? ""} onChange={(e) => update("serverName", e.target.value)} />
      </FormRow>
      <FormRow labelKey="ldap:Server Host">
        <Input value={ldap.host ?? ""} onChange={(e) => update("host", e.target.value)} />
      </FormRow>
      <FormRow labelKey="ldap:Server Port">
        <Input
          type="number"
          value={ldap.port ?? 389}
          onChange={(e) => update("port", Setting.myParseInt(e.target.value))}
        />
      </FormRow>
      <FormRow labelKey="ldap:Enable SSL">
        <Switch checked={!!ldap.enableSsl} onCheckedChange={(v) => update("enableSsl", v)} />
      </FormRow>
      {ldap.enableSsl ? (
        <FormRow labelKey="ldap:Allow self-signed certificate">
          <Switch
            checked={!!ldap.allowSelfSignedCert}
            onCheckedChange={(v) => update("allowSelfSignedCert", v)}
          />
        </FormRow>
      ) : null}
      <FormRow labelKey="ldap:Base DN">
        <Input value={ldap.baseDn ?? ""} onChange={(e) => update("baseDn", e.target.value)} />
      </FormRow>
      <FormRow labelKey="ldap:Search Filter">
        <Input value={ldap.filter ?? ""} onChange={(e) => update("filter", e.target.value)} />
      </FormRow>
      <FormRow labelKey="ldap:Filter fields">
        <TagsInput value={ldap.filterFields ?? []} onChange={(v) => update("filterFields", v)} />
      </FormRow>
      <FormRow labelKey="ldap:Admin">
        <Input value={ldap.username ?? ""} onChange={(e) => update("username", e.target.value)} />
      </FormRow>
      <FormRow labelKey="ldap:Admin Password">
        <Input type="password" value={ldap.password ?? ""} onChange={(e) => update("password", e.target.value)} />
      </FormRow>
      <FormRow labelKey="ldap:Password type">
        <SelectField
          value={ldap.passwordType ?? "Plain"}
          onChange={(v) => update("passwordType", v)}
          options={PASSWORD_TYPES.map((item) => ({id: item, name: item}))}
        />
      </FormRow>
      <FormRow labelKey="ldap:Auto Sync">
        <Input
          type="number"
          value={ldap.autoSync ?? 0}
          onChange={(e) => update("autoSync", Setting.myParseInt(e.target.value))}
        />
      </FormRow>
      <FormRow labelKey="ldap:Enable groups">
        <Switch checked={!!ldap.enableGroups} onCheckedChange={(v) => update("enableGroups", v)} />
      </FormRow>
      <FormRow labelKey="ldap:Custom attributes" block>
        <EditableTable
          rows={ldap.customAttributes ?? []}
          onChange={(rows) => update("customAttributes", rows)}
          newRow={() => ({attributeName: "", userPropertyName: ""})}
          reorderable={false}
          columns={[
            {
              key: "attributeName",
              title: i18next.t("ldap:LDAP attribute name"),
              width: 260,
              render: (row: any, _i, patch) => (
                <Input value={row.attributeName ?? ""} onChange={(e) => patch({attributeName: e.target.value})} />
              ),
            },
            {
              key: "userPropertyName",
              title: i18next.t("ldap:User property name"),
              render: (row: any, _i, patch) => (
                <Input
                  value={row.userPropertyName ?? ""}
                  onChange={(e) => patch({userPropertyName: e.target.value})}
                />
              ),
            },
          ]}
        />
      </FormRow>
      <FormRow labelKey="ldap:Default groups">
        <TagsInput value={ldap.defaultGroups ?? []} onChange={(v) => update("defaultGroups", v)} />
      </FormRow>
    </EditPageShell>
  );
}
