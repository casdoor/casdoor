import i18next from "i18next";
import {Tree} from "antd";
import React from "react";

export const NavItemTree = ({disabled, checkedKeys, defaultExpandedKeys, onCheck}) => {
  const NavItemNodes = [
    {
      title: i18next.t("organization:All"),
      key: "all",
      children: [
        {
          title: i18next.t("general:Home"),
          key: "/home-top",
          children: [
            {title: i18next.t("general:Dashboard"), key: "/"},
            {title: i18next.t("general:Shortcuts"), key: "/shortcuts"},
            {title: i18next.t("general:Apps"), key: "/apps"},
          ],
        },
        {
          title: i18next.t("general:User Management"),
          key: "/orgs-top",
          children: [
            {title: i18next.t("general:Organizations"), key: "/organizations"},
            {title: i18next.t("general:Groups"), key: "/groups"},
            {title: i18next.t("general:Users"), key: "/users"},
            {title: i18next.t("general:Invitations"), key: "/invitations"},
          ],
        },
        {
          title: i18next.t("general:Identity"),
          key: "/applications-top",
          children: [
            {title: i18next.t("general:Applications"), key: "/applications"},
            {title: i18next.t("general:Providers"), key: "/providers"},
            {title: i18next.t("general:Resources"), key: "/resources"},
            {title: i18next.t("general:Certs"), key: "/certs"},
          ],
        },
        {
          title: i18next.t("general:Authorization"),
          key: "/roles-top",
          children: [
            {title: i18next.t("general:Applications"), key: "/roles"},
            {title: i18next.t("general:Permissions"), key: "/permissions"},
            {title: i18next.t("general:Models"), key: "/models"},
            {title: i18next.t("general:Adapters"), key: "/adapters"},
            {title: i18next.t("general:Enforcers"), key: "/enforcers"},
          ],
        },
        {
          title: i18next.t("general:Logging & Auditing"),
          key: "/sessions-top",
          children: [
            {title: i18next.t("general:Sessions"), key: "/sessions"},
            {title: i18next.t("general:Records"), key: "/records"},
            {title: i18next.t("general:Tokens"), key: "/tokens"},
            {title: i18next.t("general:Verifications"), key: "/verifications"},
          ],
        },
        {
          title: i18next.t("general:Business & Payments"),
          key: "/business-top",
          children: [
            {title: i18next.t("general:Products"), key: "/products"},
            {title: i18next.t("general:Payments"), key: "/payments"},
            {title: i18next.t("general:Plans"), key: "/plans"},
            {title: i18next.t("general:Pricings"), key: "/pricings"},
            {title: i18next.t("general:Subscriptions"), key: "/subscriptions"},
            {title: i18next.t("general:Transactions"), key: "/transactions"},
          ],
        },
        {
          title: i18next.t("general:Admin"),
          key: "/admin-top",
          children: [
            {title: i18next.t("general:System Info"), key: "/sysinfo"},
            {title: i18next.t("general:Syncers"), key: "/syncers"},
            {title: i18next.t("general:Webhooks"), key: "/webhooks"},
            {title: i18next.t("general:Swagger"), key: "/swagger"},
          ],
        },
      ],
    },
  ];

  return (
    <Tree
      disabled={disabled}
      checkable
      checkedKeys={checkedKeys}
      defaultExpandedKeys={defaultExpandedKeys}
      onCheck={onCheck}
      treeData={NavItemNodes}
    />
  );
};
