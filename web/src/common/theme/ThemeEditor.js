import {Card, ConfigProvider, Form, Layout, Radio, theme} from "antd";
import ThemePicker from "./ThemePicker";
import ColorPicker from "./ColorPicker";
import RadiusPicker from "./RadiusPicker";
import * as React from "react";
import {GREEN_COLOR, PINK_COLOR} from "./ColorPicker";
import {Content} from "antd/es/layout/layout";
import i18next from "i18next";
import {useEffect} from "react";

export const ThemeDefault = {
  themeType: "default",
  colorPrimary: "#5734d3",
  borderRadius: 6,
  isCompact: false,
};

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
  const themeData = props.themeData ?? ThemeDefault;
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
    const mergedData = Object.assign(Object.assign(Object.assign({}, ThemeDefault), {themeType}), ThemesInfo[themeType]);
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
      <Layout style={{}}>
        <Content style={{width: "800px", margin: "0 auto"}}>
          <Card
            title={i18next.t("theme:My Theme")}
          >
            <Form
              form={form}
              initialValues={themeData}
              onValuesChange={onThemeChange}
              labelCol={{span: 4}}
              wrapperCol={{span: 20}}
              style={{width: "800px", margin: "0 auto"}}
            >
              <Form.Item label={i18next.t("theme:Theme")} name="themeType">
                <ThemePicker />
              </Form.Item>
              <Form.Item label={i18next.t("theme:Primary Color")} name="colorPrimary">
                <ColorPicker />
              </Form.Item>
              <Form.Item label={i18next.t("theme:Border Radius")} name="borderRadius">
                <RadiusPicker />
              </Form.Item>
              <Form.Item label={i18next.t("theme:Compact")} name="isCompact">
                <Radio.Group>
                  <Radio value={false}>{i18next.t("theme:default")}</Radio>
                  <Radio value={true}>{i18next.t("theme:compact")}</Radio>
                </Radio.Group>
              </Form.Item>
            </Form>
          </Card>
        </Content>
      </Layout>
    </ConfigProvider>
  );
}
