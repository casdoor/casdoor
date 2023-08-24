import i18next from "i18next";
import React from "react";
import * as Setting from "../Setting";
import GridCards from "./GridCards";

const ShortcutsPage = () => {
  const items = [
    {link: "/organizations", name: i18next.t("general:Organizations"), description: i18next.t("general:User containers")},
    {link: "/users", name: i18next.t("general:Users"), description: i18next.t("general:Users under all organizations")},
    {link: "/providers", name: i18next.t("general:Providers"), description: i18next.t("general:OAuth providers")},
    {link: "/applications", name: i18next.t("general:Applications"), description: i18next.t("general:Applications that require authentication")},
  ];

  const getItems = () => {
    return items.map(item => {
      item.logo = `${Setting.StaticBaseUrl}/img${item.link}.png`;
      item.createdTime = "";
      return item;
    });
  };

  return (
    <div style={{display: "flex", justifyContent: "center", flexDirection: "column", alignItems: "center"}}>
      <GridCards items={getItems()} />
    </div>
  );
};

export default ShortcutsPage;
