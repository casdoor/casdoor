import * as React from "react";
import i18next from "i18next";
import {ExternalLink} from "lucide-react";
import {useNavigate, useParams} from "react-router-dom";
import {UnauthorizedPage} from "@/components/common/UnauthorizedPage";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Switch} from "@/components/ui/switch";
import {Loading} from "@/components/common/Loading";
import {SelectField} from "@/components/common/SelectField";
import {EditableTable} from "@/components/crud/EditableTable";
import {EditPageShell} from "@/components/crud/EditPageShell";
import {FormRow} from "@/components/crud/FormRow";
import {useAccount} from "@/hooks/use-account";
import {useEditRecord} from "@/hooks/use-edit-record";
import * as FormBackend from "@/backend/FormBackend";
import {submitEdit} from "@/lib/crud";
import * as Setting from "@/lib/setting";

interface FormItem {
  name: string;
  label: string;
  visible: boolean;
  width?: string;
}

const PREVIEW_PAGES: Record<string, React.ComponentType<{formItems?: any[]}>> = {
  users: React.lazy(() => import("@/pages/UserListPage")),
  applications: React.lazy(() => import("@/pages/ApplicationListPage")),
  providers: React.lazy(() => import("@/pages/ProviderListPage")),
  organizations: React.lazy(() => import("@/pages/OrganizationListPage")),
};

/** The list page this form customizes, rendered with the items being edited. */
function FormListPreview({type, formItems}: {type: string; formItems: any[]}) {
  const ListPage = PREVIEW_PAGES[type];
  if (!ListPage) {
    return null;
  }

  return (
    <div className="relative h-[600px] overflow-auto rounded-lg border">
      <div className="pointer-events-none p-4">
        <React.Suspense fallback={null}>
          <ListPage formItems={formItems} />
        </React.Suspense>
      </div>
      {/* the preview is to look at, not to use */}
      <div className="absolute inset-0 z-10 cursor-not-allowed bg-foreground/5" />
    </div>
  );
}

export default function FormEditPage() {
  const {formName = ""} = useParams();
  const {account} = useAccount();
  const navigate = useNavigate();
  const [name, setName] = React.useState(formName);
  const [saving, setSaving] = React.useState(false);

  const {record: form, updateField, setRecord, loading, denied, mode, setMode} = useEditRecord<any>({
    fetch: () => FormBackend.getForm(account?.owner ?? "", formName),
    deps: [account?.owner, formName],
  });

  const save = async(exitAfterSave: boolean) => {
    if (form === null) {
      return;
    }
    setSaving(true);
    await submitEdit({
      mode,
      record: Setting.deepCopy(form),
      add: (record) => FormBackend.addForm(record),
      update: (record) => FormBackend.updateForm(form.owner, name, record),
      onSaved: () => {
        setName(form.name);
        setMode("edit");
        if (exitAfterSave) {
          navigate("/forms");
        } else if (`/forms/${form.name}` !== window.location.pathname) {
          navigate(`/forms/${form.name}`, {replace: true});
        }
      },
      onFailed: () => {
        if (mode !== "add") {
          updateField("name", name);
        }
      },
    });
    setSaving(false);
  };

  if (denied) {
    return <UnauthorizedPage />;
  }

  if (loading || form === null) {
    return <Loading />;
  }

  const items: FormItem[] = form.formItems ?? [];
  const defaultItems = (Setting.getFormTypeItems(form.type) ?? []) as FormItem[];

  const patch = (values: Record<string, any>) => {
    setRecord((prev: any) => (prev === null ? prev : {...prev, ...values}));
  };

  const labelOf = (item: FormItem) => {
    const known = defaultItems.find((candidate) => candidate.name === item.name);
    return i18next.t(known?.label ?? item.label ?? item.name);
  };

  return (
    <EditPageShell
      grid
      title={mode === "add" ? i18next.t("form:New Form") : i18next.t("form:Edit Form")}
      mode={mode}
      backTo="/forms"
      saving={saving}
      onSave={save}
    >
      <FormRow labelKey="general:Name">
        <Input value={form.name ?? ""} disabled />
      </FormRow>
      <FormRow labelKey="general:Display name">
        <Input value={form.displayName ?? ""} onChange={(e) => updateField("displayName", e.target.value)} />
      </FormRow>
      <FormRow labelKey="general:Type">
        <SelectField
          value={form.type}
          onChange={(value) => {
            // the type decides the name, the display name and the default items
            patch({
              type: value,
              name: value,
              displayName: value,
              formItems: Setting.getFormTypeItems(value),
            });
          }}
          options={Setting.getFormTypeOptions().map((option: any) => ({id: option.id, name: i18next.t(option.name)}))}
        />
      </FormRow>
      <FormRow block label={i18next.t("general:Tag")} tooltip={i18next.t("product:Tag - Tooltip")}>
        <Input
          value={form.tag ?? ""}
          onChange={(e) => {
            const tag = e.target.value;
            patch({tag, name: tag ? `${form.type}-tag-${tag}` : form.type});
          }}
        />
      </FormRow>
      <FormRow block labelKey="general:Tag">
        <Input
          value={form.tag ?? ""}
          onChange={(e) => {
            const tag = e.target.value;
            // the name is derived from type and tag, so a form stays addressable
            patch({tag, name: tag ? `${form.type}-tag-${tag}` : form.type});
          }}
        />
      </FormRow>
      <FormRow labelKey="form:Form items" block>
        <EditableTable<FormItem>
          title={
            <div className="flex items-center gap-2">
              <span>{i18next.t("form:Form items")}</span>
              <Button variant="outline" size="sm" onClick={() => updateField("formItems", Setting.getFormTypeItems(form.type))}>
                {i18next.t("general:Reset to Default")}
              </Button>
            </div>
          }
          rows={items}
          onChange={(rows) => updateField("formItems", rows)}
          newRow={() => ({name: "", label: "", visible: false})}
          rowKey={(row, index) => `${row.name}-${index}`}
          columns={[
            {
              key: "name",
              title: i18next.t("general:Name"),
              width: 220,
              render: (row, index, update) => (
                <SelectField
                  value={row.name}
                  onChange={(value) => update({name: value})}
                  options={[
                    ...(row.name ? [{id: row.name, name: labelOf(row)}] : []),
                    ...Setting.getDeduplicatedArray(defaultItems, items, "name").map((item: FormItem) => ({
                      id: item.name,
                      name: i18next.t(item.label),
                    })),
                  ]}
                />
              ),
            },
            {
              key: "label",
              title: i18next.t("signup:Label"),
              width: 220,
              render: (row, index, update) => (
                <Input value={row.label ?? ""} onChange={(e) => update({label: e.target.value})} />
              ),
            },
            {
              key: "visible",
              title: i18next.t("organization:Visible"),
              width: 110,
              render: (row, index, update) => (
                <Switch checked={!!row.visible} onCheckedChange={(v) => update({visible: v})} />
              ),
            },
            {
              key: "width",
              title: i18next.t("form:Width"),
              width: 140,
              render: (row, index, update) => (
                <Input value={row.width ?? ""} onChange={(e) => update({width: e.target.value})} />
              ),
            },
          ]}
        />
      </FormRow>
      <FormRow labelKey="general:Preview" block>
        <div className="space-y-2">
          <FormListPreview type={form.type} formItems={items} />
          {form.type ? (
            <Button variant="outline" size="sm" onClick={() => Setting.openLink(`/${form.type}`)}>
              <ExternalLink />
              {i18next.t("general:Preview")}
            </Button>
          ) : null}
        </div>
      </FormRow>
    </EditPageShell>
  );
}
