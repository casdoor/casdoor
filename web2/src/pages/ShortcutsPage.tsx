import i18next from "i18next";
import {GridCards} from "@/components/common/GridCards";
import {PageHeader} from "@/components/crud/PageHeader";
import * as Setting from "@/lib/setting";

export default function ShortcutsPage() {
  const items = [
    {link: "/organizations", name: i18next.t("general:Organizations"), description: i18next.t("general:User containers")},
    {link: "/users", name: i18next.t("general:Users"), description: i18next.t("general:Users under all organizations")},
    {link: "/providers", name: i18next.t("application:Providers"), description: i18next.t("general:OAuth providers")},
    {
      link: "/applications",
      name: i18next.t("general:Applications"),
      description: i18next.t("general:Applications that require authentication"),
    },
  ].map((item) => ({...item, logo: `${Setting.StaticBaseUrl}/img${item.link}.png`}));

  return (
    <div className="space-y-4">
      <PageHeader title={i18next.t("general:Shortcuts")} />
      <GridCards items={items} />
    </div>
  );
}
