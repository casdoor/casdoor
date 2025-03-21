import i18next from "i18next";
import {Tree} from "antd";
import React from "react";

export const WidgetItemTree = ({disabled, checkedKeys, defaultExpandedKeys, onCheck}) => {
  const WidgetItemNodes = [
    {
      title: i18next.t("organization:All"),
      key: "all",
      children: [
        {title: i18next.t("general:Tour"), key: "tour"},
        {title: i18next.t("general:AI Assistant"), key: "ai-assistant"},
        {title: i18next.t("user:Language"), key: "language"},
        {title: i18next.t("theme:Theme"), key: "theme"},
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
      treeData={WidgetItemNodes}
    />
  );
};
