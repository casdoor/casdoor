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
import ColorPicker from "./ColorPicker";
import RadiusPicker from "./RadiusPicker";
import * as React from "react";
import {GREEN_COLOR, PINK_COLOR} from "./ColorPicker";
import {Content} from "antd/es/layout/layout";
import i18next from "i18next";
import {useEffect} from "react";
import * as Setting from "../../Setting";

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
  const themeData = props.themeData ?? Setting.ThemeDefault;
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
    const mergedData = Object.assign(Object.assign(Object.assign({}, Setting.ThemeDefault), {themeType}), ThemesInfo[themeType]);
    onThemeChange(null, mergedData);
    form.setFieldsValue(mergedData);
  }, [themeType]);

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
              <Form.Item label={i18next.t("theme:Compact ")} valuePropName="checked" name="isCompact">
                <Switch />
              </Form.Item>
            </Form>
          </Card>
        </Content>
      </Layout>
    </ConfigProvider>
  );
}
