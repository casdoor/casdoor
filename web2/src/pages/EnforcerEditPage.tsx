import * as React from "react";
import i18next from "i18next";
import {useNavigate, useParams} from "react-router-dom";
import {UnauthorizedPage} from "@/components/common/UnauthorizedPage";
import {Input} from "@/components/ui/input";
import {Loading} from "@/components/common/Loading";
import {SearchableSelect} from "@/components/common/SearchableSelect";
import {PolicyTable} from "@/components/casbin/PolicyTable";
import {EditPageShell} from "@/components/crud/EditPageShell";
import {FormRow} from "@/components/crud/FormRow";
import {useAccount} from "@/hooks/use-account";
import {useEditRecord} from "@/hooks/use-edit-record";
import {useOrganizationOptions} from "@/hooks/use-options";
import * as AdapterBackend from "@/backend/AdapterBackend";
import * as EnforcerBackend from "@/backend/EnforcerBackend";
import * as ModelBackend from "@/backend/ModelBackend";
import {submitEdit} from "@/lib/crud";
import * as Setting from "@/lib/setting";

const PAGE_SIZE = 1000;

export default function EnforcerEditPage() {
  const {organizationName = "", enforcerName = ""} = useParams();
  const {account} = useAccount();
  const navigate = useNavigate();
  const organizations = useOrganizationOptions();

  const [owner, setOwner] = React.useState(organizationName);
  const [name, setName] = React.useState(enforcerName);
  const [models, setModels] = React.useState<any[]>([]);
  const [adapters, setAdapters] = React.useState<any[]>([]);
  const [saving, setSaving] = React.useState(false);

  // loadModelCfg=true so the policy table knows the model's p/g sections
  const {record: enforcer, updateField, loading, denied, mode, setMode, reload} = useEditRecord<any>({
    fetch: () => EnforcerBackend.getEnforcer(organizationName, enforcerName, true),
    deps: [organizationName, enforcerName],
  });

  const loadOptions = React.useCallback((forOwner: string) => {
    if (!forOwner) {
      return;
    }
    ModelBackend.getModels(forOwner, 1, PAGE_SIZE).then((res: any) => {
      if (res.status === "ok") {
        setModels(res.data ?? []);
      }
    });
    AdapterBackend.getAdapters(forOwner, 1, PAGE_SIZE).then((res: any) => {
      if (res.status === "ok") {
        setAdapters(res.data ?? []);
      }
    });
  }, []);

  React.useEffect(() => {
    loadOptions(enforcer?.owner ?? owner);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [enforcer?.owner]);

  const save = async(exitAfterSave: boolean) => {
    if (enforcer === null) {
      return;
    }
    const isAdd = mode === "add";
    setSaving(true);
    await submitEdit({
      mode,
      record: Setting.deepCopy(enforcer),
      add: (record) => EnforcerBackend.addEnforcer(record),
      update: (record) => EnforcerBackend.updateEnforcer(owner, name, record),
      onSaved: () => {
        setOwner(enforcer.owner);
        setName(enforcer.name);
        setMode("edit");
        if (exitAfterSave) {
          navigate("/enforcers");
          return;
        }
        const next = `/enforcers/${enforcer.owner}/${enforcer.name}`;
        if (next !== window.location.pathname) {
          navigate(next, {replace: true});
        } else if (isAdd) {
          // the enforcer exists now, reload it so that its modelCfg and policies are available
          reload();
        }
      },
      onFailed: () => {
        if (!isAdd) {
          updateField("name", name);
        }
      },
    });
    setSaving(false);
  };

  if (denied) {
    return <UnauthorizedPage />;
  }

  if (loading || enforcer === null) {
    return <Loading />;
  }

  const isBuiltIn = Setting.builtInObject(enforcer);
  // the backend stores "owner/name" for both, see web/src/EnforcerEditPage.js
  const toIdOptions = (items: any[]) =>
    items.map((item) => ({value: `${item.owner}/${item.name}`, label: `${item.owner}/${item.name}`}));

  return (
    <EditPageShell
      title={mode === "add" ? i18next.t("enforcer:New Enforcer") : i18next.t("enforcer:Edit Enforcer")}
      mode={mode}
      backTo="/enforcers"
      saving={saving}
      onSave={save}
    >
      <FormRow labelKey="general:Organization">
        <SearchableSelect
          disabled={!Setting.isAdminUser(account) || isBuiltIn}
          value={enforcer.owner ?? ""}
          onChange={(value) => updateField("owner", value)}
          options={organizations}
        />
      </FormRow>
      <FormRow labelKey="general:Name">
        <Input disabled={isBuiltIn} value={enforcer.name ?? ""} onChange={(e) => updateField("name", e.target.value)} />
      </FormRow>
      <FormRow labelKey="general:Display name">
        <Input value={enforcer.displayName ?? ""} onChange={(e) => updateField("displayName", e.target.value)} />
      </FormRow>
      <FormRow labelKey="general:Description">
        <Input value={enforcer.description ?? ""} onChange={(e) => updateField("description", e.target.value)} />
      </FormRow>
      <FormRow labelKey="general:Model">
        <SearchableSelect
          disabled={isBuiltIn}
          value={enforcer.model ?? ""}
          onChange={(value) => updateField("model", value)}
          options={toIdOptions(models)}
        />
      </FormRow>
      <FormRow labelKey="general:Adapter">
        <SearchableSelect
          disabled={isBuiltIn}
          value={enforcer.adapter ?? ""}
          onChange={(value) => updateField("adapter", value)}
          options={toIdOptions(adapters)}
        />
      </FormRow>
      {/* policies live in the adapter and are looked up through the enforcer, so they need a saved enforcer */}
      {mode === "add" ? null : (
        <FormRow labelKey="adapter:Policies" block>
          <PolicyTable enforcer={enforcer} modelCfg={enforcer.modelCfg} mode={mode} />
        </FormRow>
      )}
    </EditPageShell>
  );
}
