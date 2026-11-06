import i18next from "i18next";
import {SimpleEditPage, type EditField} from "@/components/crud/SimpleEditPage";
import * as SettingBackend from "@/backend/SettingBackend";
import * as Conf from "@/Conf";
import {useAccount} from "@/hooks/use-account";

export default function SettingEditPage() {
  const {account, setAccount} = useAccount();

  const updateSetting = async(record) => {
    const res = await SettingBackend.updateSetting(record);
    if (res.status === "ok") {
      // the console reads this switch only when the page loads, so the Casdoor Cloud links would stay until a reload;
      // a new account object makes the header, sidebar and dashboard render again
      Conf.setConfig({showCloudPromotion: !record.hideCloudPromotion});
      if (account) {
        setAccount({...account});
      }
    }
    return res;
  };

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
      add={updateSetting}
      update={updateSetting}
    />
  );
}
