import {Card, ConfigProvider, Form, Radio, theme} from "antd";
import ThemePicker from "./ThemePicker";
import ColorPicker from "./ColorPicker";
import RadiusPicker from "./RadiusPicker";
import * as React from "react";
import {PINK_COLOR} from "./colorUtil";

const ThemeDefault = {
  themeType: "default",
  colorPrimary: "#1677FF",
  borderRadius: 6,
  compact: "default",
};

const ThemesInfo = {
  default: {},
  dark: {
    borderRadius: 2,
  },
  lark: {
    colorPrimary: "#00B96B",
    borderRadius: 4,
  },
  comic: {
    colorPrimary: PINK_COLOR,
    borderRadius: 16,
  },
};

export default function ThemeEditor() {
  const [themeData, setThemeData] = React.useState(ThemeDefault);

  const onThemeChange = (_, nextThemeData) => {
    setThemeData(nextThemeData);
  };

  const {compact, themeType, ...themeToken} = themeData;
  const isLight = themeType !== "dark";
  const [form] = Form.useForm();

  const algorithmFn = React.useMemo(() => {
    const algorithms = [isLight ? theme.defaultAlgorithm : theme.darkAlgorithm];

    if (compact === "compact") {
      algorithms.push(theme.compactAlgorithm);
    }

    return algorithms;
  }, [isLight, compact]);

  // ================================ Themes ================================
  React.useEffect(() => {
    const mergedData = Object.assign(Object.assign(Object.assign({}, ThemeDefault), {themeType}), ThemesInfo[themeType]);
    setThemeData(mergedData);
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
      <Card
        title={"主题"}
      >
        <Form
          form={form}
          initialValues={themeData}
          onValuesChange={onThemeChange}
          labelCol={{span: 4}}
          wrapperCol={{span: 20}}
        >
          <Form.Item label={"type"} name="themeType">
            <ThemePicker />
          </Form.Item>

          <Form.Item label={"color"} name="colorPrimary">
            <ColorPicker />
          </Form.Item>
          <Form.Item label={"radius"} name="borderRadius">
            <RadiusPicker />
          </Form.Item>
          <Form.Item label={"compact"} name="compact">
            <Radio.Group>
              <Radio value="default">default</Radio>
              <Radio value="compact">compact</Radio>
            </Radio.Group>
          </Form.Item>
        </Form>
      </Card>
    </ConfigProvider>
  );
}
