import i18next from "i18next";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import * as SettingBackend from "@/backend/SettingBackend";

export default function SettingEditPage() {
  const fields: EditField[] = [
    {type: "number", name: "dbMaxOpenConns", labelKey: "setting:Max open connections", min: 0},
    {
      type: "number",
      name: "dbMaxIdleConns",
      labelKey: "setting:Max idle connections",
      min: 0,
      validate: (value, ctx) =>
        ctx.record.dbMaxOpenConns > 0 && value > ctx.record.dbMaxOpenConns
          ? i18next.t("setting:Cannot be more than the max open connections")
          : undefined,
    },
    {
      type: "number",
      name: "dbConnMaxLifetime",
      labelKey: "setting:Connection max lifetime",
      min: 0,
      suffix: i18next.t("usage:seconds"),
    },
    {type: "number", name: "recordRetentionDays", labelKey: "setting:Default record retention days", min: 0},
    {type: "switch", name: "hideCloudPromotion", labelKey: "setting:Hide Casdoor Cloud links"},
  ];

  return (
    <SimpleEditPage
      titleKey="general:Global Settings"
      backTo="/sysinfo"
      fields={fields}
      fetch={() => SettingBackend.getSetting()}
      add={(record) => SettingBackend.updateSetting(record)}
      update={(record) => SettingBackend.updateSetting(record)}
    />
  );
}
