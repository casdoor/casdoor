// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

/** @jsxImportSource @emotion/react */

import {css} from "@emotion/react";
import {Space, theme} from "antd";
import * as React from "react";
import i18next from "i18next";
import * as Setting from "../../Setting";

const {useToken} = theme;

export const THEMES = {
  default: `${Setting.StaticBaseUrl}/img/theme_default.svg`,
  dark: `${Setting.StaticBaseUrl}/img/theme_dark.svg`,
  lark: `${Setting.StaticBaseUrl}/img/theme_lark.svg`,
  comic: `${Setting.StaticBaseUrl}/img/theme_comic.svg`,
};

const themeTypes = {
  default: "Default", // i18next.t("theme:Default")
  dark: "Dark",       // i18next.t("theme:Dark")
  lark: "Document",   // i18next.t("theme:Document")
  comic: "Blossom",   // i18next.t("theme:Blossom")
};

const useStyle = () => {
  const {token} = useToken();
  return {
    themeCard: css `
      border-radius: ${token.borderRadius}px;
      cursor: pointer;
      transition: all ${token.motionDurationSlow};
      overflow: hidden;
      display: inline-block;

      & > input[type="radio"] {
        width: 0;
        height: 0;
        opacity: 0;
      }

      img {
        vertical-align: top;
        box-shadow: 0 3px 6px -4px rgba(0, 0, 0, 0.12), 0 6px 16px 0 rgba(0, 0, 0, 0.08),
          0 9px 28px 8px rgba(0, 0, 0, 0.05);
      }

      &:focus-within,
      &:hover {
        transform: scale(1.04);
      }
    `,
    themeCardActive: css `
      box-shadow: 0 0 0 1px ${token.colorBgContainer},
        0 0 0 ${token.controlOutlineWidth * 2 + 1}px ${token.colorPrimary};
        
      &:hover:not(:focus-within) {
        transform: scale(1);
      }
    `,
  };
};

export default function ThemePicker({value, onChange}) {
  const {token} = useToken();
  const style = useStyle();

  return (
    <Space size={token.paddingLG}>
      {Object.keys(THEMES).map((theme) => {
        const url = THEMES[theme];
        return (
          <Space key={theme} direction="vertical" align="center">
            <label
              css={[style.themeCard, value === theme && style.themeCardActive]}
              onClick={() => {
                onChange?.(theme);
              }}
            >
              <input type="radio" name="theme" />
              <img src={url} alt={theme} />
            </label>
            <span>{i18next.t(`theme:${themeTypes[theme]}`)}</span>
          </Space>
        );
      })}
    </Space>
  );
}
