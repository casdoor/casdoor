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

import {Input, Popover, Space, theme} from "antd";
import React, {useEffect, useMemo, useState} from "react";
import {css} from "@emotion/react";
import {TinyColor} from "@ctrl/tinycolor";
import ColorPanel from "antd-token-previewer/es/ColorPanel";

export const BLUE_COLOR = "#1677FF";
export const PINK_COLOR = "#ED4192";
export const GREEN_COLOR = "#00B96B";

export const COLORS = [
  {
    color: BLUE_COLOR,
  },
  {
    color: "#5734d3",
  },
  {
    color: "#9E339F",
  },
  {
    color: PINK_COLOR,
  },
  {
    color: "#E0282E",
  },
  {
    color: "#F4801A",
  },
  {
    color: "#F2BD27",
  },
  {
    color: GREEN_COLOR,
  },
];

export const PRESET_COLORS = COLORS.map(({color}) => color);

const {useToken} = theme;

const useStyle = () => {
  const {token} = useToken();
  return {
    color: css `
      width: ${token.controlHeightLG / 2}px;
      height: ${token.controlHeightLG / 2}px;
      border-radius: 100%;
      cursor: pointer;
      transition: all ${token.motionDurationFast};
      display: inline-block;

      & > input[type="radio"] {
        width: 0;
        height: 0;
        opacity: 0;
      }
    `,
    colorActive: css `
      box-shadow: 0 0 0 1px ${token.colorBgContainer},
        0 0 0 ${token.controlOutlineWidth * 2 + 1}px ${token.colorPrimary};
    `,
  };
};

const DebouncedColorPanel = ({color, onChange}) => {
  const [value, setValue] = useState(color);

  useEffect(() => {
    const timeout = setTimeout(() => {
      onChange?.(value);
    }, 200);
    return () => clearTimeout(timeout);
  }, [value]);

  useEffect(() => {
    setValue(color);
  }, [color]);

  return <ColorPanel color={value} onChange={setValue} />;
};

export default function ColorPicker({value, onChange}) {
  const style = useStyle();

  const matchColors = useMemo(() => {
    const valueStr = new TinyColor(value).toRgbString();
    let existActive = false;

    const colors = PRESET_COLORS.map((color) => {
      const colorStr = new TinyColor(color).toRgbString();
      const active = colorStr === valueStr;
      existActive = existActive || active;

      return {
        color,
        active,
        picker: false,
      };
    });

    return [
      ...colors,
      {
        color: "conic-gradient(red, yellow, lime, aqua, blue, magenta, red)",
        picker: true,
        active: !existActive,
      },
    ];
  }, [value]);

  return (
    <Space size="large">
      <Input
        value={value}
        onChange={(event) => {
          onChange?.(event.target.value);
        }}
        style={{width: 120}}
      />

      <Space size="middle">
        {matchColors.map(({color, active, picker}) => {
          let colorNode = (
            <label
              key={color}
              css={[style.color, active && style.colorActive]}
              style={{
                background: color,
              }}
              onClick={() => {
                if (!picker) {
                  onChange?.(color);
                }
              }}
            >
              <input type="radio" name={picker ? "picker" : "color"} tabIndex={picker ? -1 : 0} />
            </label>
          );

          if (picker) {
            colorNode = (
              <Popover
                key={color}
                overlayInnerStyle={{padding: 0}}
                content={
                  <DebouncedColorPanel color={value || ""} onChange={(c) => onChange?.(c)} />
                }
                trigger="click"
                showArrow={false}
              >
                {colorNode}
              </Popover>
            );
          }

          return colorNode;
        })}
      </Space>
    </Space>
  );
}
