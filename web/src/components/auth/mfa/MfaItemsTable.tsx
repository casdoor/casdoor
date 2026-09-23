import i18next from "i18next";
import {SelectField} from "@/components/common/SelectField";
import {EditableTable} from "@/components/crud/EditableTable";
import {EmailMfaType, PushMfaType, SmsMfaType, TotpMfaType} from "@/components/auth/mfa/constants";
import * as Setting from "@/lib/setting";

/** The stored name is the backend's MFA type, not the label shown in the dropdown. */
const MFA_ITEMS = [
  {id: SmsMfaType, name: "Phone"},
  {id: EmailMfaType, name: "Email"},
  {id: TotpMfaType, name: "App"},
  {id: PushMfaType, name: "Push"},
];

/** `general:Optional` and friends do not exist; the antd table uses these. */
const MFA_RULES: Record<string, string> = {
  "Optional": "organization:Optional",
  "Prompted": "organization:Prompt",
  "Required": "organization:Required",
};

/** The "MFA items" table shared by the organization and user edit pages, as web-old's MfaTable. */
export function MfaItemsTable({rows, onChange}: {rows: any[] | undefined | null; onChange: (rows: any[]) => void}) {
  const items = rows ?? [];

  return (
    <EditableTable
      rows={items}
      onChange={onChange}
      newRow={items.length < MFA_ITEMS.length ? () => ({
        name: Setting.getNewRowNameForTable(items, "Please select a MFA method"),
        rule: "Optional",
      }) : undefined}
      columns={[
        {
          key: "name",
          title: i18next.t("general:Name"),
          width: 220,
          render: (row: any, index, patch) => (
            <SelectField
              value={row.name}
              onChange={(value) => patch({name: value})}
              options={MFA_ITEMS.filter((item) => !items.some((other: any, i) => i !== index && other.name === item.id))}
            />
          ),
        },
        {
          key: "rule",
          title: i18next.t("application:Rule"),
          width: 220,
          render: (row: any, _index, patch) => (
            <SelectField
              value={row.rule}
              onChange={(value) => {
                const required = items.filter((item: any) => item.rule === "Required").length;
                if (value === "Required" && required >= 1 && row.rule !== "Required") {
                  Setting.showMessage("error", i18next.t("general:Only 1 MFA method can be required"));
                  return;
                }
                patch({rule: value});
              }}
              options={Object.entries(MFA_RULES).map(([id, key]) => ({id, name: i18next.t(key)}))}
            />
          ),
        },
      ]}
    />
  );
}
