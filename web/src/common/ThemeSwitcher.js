// Copyright 2025 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import React, {useEffect, useState} from "react";
import {Button, Popover, Space} from "antd";
import {BgColorsOutlined} from "@ant-design/icons";
import * as ApplicationBackend from "../backend/ApplicationBackend";
import i18next from "i18next";
import * as Conf from "../Conf";

function ThemeSwitcher(props) {
  const {application} = props;
  const [builtInThemes, setBuiltInThemes] = useState([]);
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    if (!Conf.IsDemoMode) {
      return;
    }

    // Fetch built-in themes
    ApplicationBackend.getBuiltInThemes().then((res) => {
      if (res.status === "ok") {
        setBuiltInThemes(res.data || []);
      }
    });
  }, []);

  if (!Conf.IsDemoMode || builtInThemes.length === 0) {
    return null;
  }

  const handleThemeSelect = (themeName) => {
    // Store selected theme in localStorage
    localStorage.setItem("casdoor-demo-theme", themeName);
    // Reload page to apply theme
    window.location.reload();
  };

  const themeButtons = (
    <div style={{maxWidth: "300px"}}>
      <p>{i18next.t("theme:Select a theme to preview")}:</p>
      <Space wrap>
        {builtInThemes.map((theme) => (
          <Button
            key={theme.name}
            size="small"
            onClick={() => handleThemeSelect(theme.name)}
          >
            {i18next.t(`theme:${theme.displayName}`)}
          </Button>
        ))}
      </Space>
    </div>
  );

  return (
    <div style={{
      position: "fixed",
      bottom: "20px",
      right: "20px",
      zIndex: 1000,
    }}>
      <Popover
        content={themeButtons}
        title={i18next.t("theme:Theme Previewer")}
        trigger="click"
        open={visible}
        onOpenChange={setVisible}
        placement="topRight"
      >
        <Button
          shape="circle"
          size="large"
          icon={<BgColorsOutlined />}
          style={{
            backgroundColor: "rgba(255, 255, 255, 0.9)",
            boxShadow: "0 2px 8px rgba(0, 0, 0, 0.15)",
          }}
        />
      </Popover>
    </div>
  );
}

export default ThemeSwitcher;
