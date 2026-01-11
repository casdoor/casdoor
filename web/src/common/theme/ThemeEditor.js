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

import {Card, ConfigProvider, Form, Layout, Switch, theme} from "antd";
import ThemePicker from "./ThemePicker";
import ColorPicker, {GREEN_COLOR, PINK_COLOR} from "./ColorPicker";
import RadiusPicker from "./RadiusPicker";
import * as React from "react";
import {useEffect, useLayoutEffect} from "react";
import {Content} from "antd/es/layout/layout";
import i18next from "i18next";
import * as Conf from "../../Conf";

const ThemesInfo = {
  default: {},
  dark: {
    borderRadius: 2,
  },
  lark: {
    colorPrimary: GREEN_COLOR,
    borderRadius: 4,
  },
  comic: {
    colorPrimary: PINK_COLOR,
    borderRadius: 16,
  },
};

const onChange = () => {};

export default function ThemeEditor(props) {
  const themeData = props.themeData ?? Conf.ThemeDefault;
  const onThemeChange = props.onThemeChange ?? onChange;

  const {isCompact, themeType, ...themeToken} = themeData;
  const isLight = themeType !== "dark";
  const [form] = Form.useForm();

  const algorithmFn = React.useMemo(() => {
    const algorithms = [isLight ? theme.defaultAlgorithm : theme.darkAlgorithm];

    if (isCompact === true) {
      algorithms.push(theme.compactAlgorithm);
    }

    return algorithms;
  }, [isLight, isCompact]);

  useEffect(() => {
    onThemeChange(null, themeData);
    form.setFieldsValue(themeData);
  }, []);

  useEffect(() => {
    form.setFieldsValue(themeData);
  }, [themeData, form]);

  const prevThemeTypeRef = React.useRef(themeType);
  useLayoutEffect(() => {
    if (prevThemeTypeRef.current !== themeType) {
      const themeInfo = ThemesInfo[themeType] || {};
      const prevThemeInfo = ThemesInfo[prevThemeTypeRef.current] || {};

      const mergedData = {...themeData, themeType};
      let hasChanges = false;

      // Check if colorPrimary is default or from the previous theme (not customized by user)
      const isDefaultColor = themeData.colorPrimary === Conf.ThemeDefault.colorPrimary ||
                            themeData.colorPrimary === prevThemeInfo.colorPrimary;
      if (isDefaultColor && themeInfo.colorPrimary) {
        mergedData.colorPrimary = themeInfo.colorPrimary;
        hasChanges = true;
      }

      // Check if borderRadius is default or from the previous theme (not customized by user)
      const isDefaultBorderRadius = themeData.borderRadius === Conf.ThemeDefault.borderRadius ||
                                   themeData.borderRadius === prevThemeInfo.borderRadius;
      if (isDefaultBorderRadius && themeInfo.borderRadius !== undefined) {
        mergedData.borderRadius = themeInfo.borderRadius;
        hasChanges = true;
      }

      if (hasChanges) {
        onThemeChange(null, mergedData);
        form.setFieldsValue(mergedData);
      }
      prevThemeTypeRef.current = themeType;
    }
  }, [themeType, themeData, onThemeChange, form]);

  return (
    <ConfigProvider
      theme={{
        token: {
          ...themeToken,
        },
        hashed: true,
        algorithm: algorithmFn,
      }}
    >
      <Layout style={{width: "800px", backgroundColor: "white"}}>
        <Content >
          <Card
            title={i18next.t("theme:Theme")}
          >
            <Form
              form={form}
              initialValues={themeData}
              onValuesChange={onThemeChange}
              labelCol={{span: 4}}
              wrapperCol={{span: 20}}
              style={{width: "800px"}}
            >
              <Form.Item label={i18next.t("theme:Theme")} name="themeType">
                <ThemePicker />
              </Form.Item>
              <Form.Item label={i18next.t("theme:Primary color")} name="colorPrimary">
                <ColorPicker />
              </Form.Item>
              <Form.Item label={i18next.t("theme:Border radius")} name="borderRadius">
                <RadiusPicker />
              </Form.Item>
              <Form.Item label={i18next.t("theme:Is compact")} valuePropName="checked" name="isCompact">
                <Switch />
              </Form.Item>
            </Form>
          </Card>
        </Content>
      </Layout>
    </ConfigProvider>
  );
}
