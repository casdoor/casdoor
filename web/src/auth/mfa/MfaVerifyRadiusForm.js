import {Button, Checkbox, Form, Input} from "antd";
import i18next from "i18next";
import React from "react";
import {mfaAuth} from "./MfaVerifyForm";

export const MfaVerifyRadiusForm = ({mfaProps, application, onFinish, method, user}) => {
  const [form] = Form.useForm();
  return (
    <Form
      form={form}
      style={{width: "300px"}}
      onFinish={onFinish}
      initialValues={{
        countryCode: mfaProps.countryCode,
        enableMfaRemember: false,
      }}
    >
      {
        method === mfaAuth ? null : (<Form.Item
          name="dest"
          noStyle
          rules={[{required: true, message: i18next.t("login:Please input your RADIUS username!")}]}
        >
          <Input
            style={{width: "100%"}}
            placeholder={i18next.t("signup:Username")}
          />
        </Form.Item>)
      }
      <Form.Item
        name="passcode"
        noStyle
        rules={[{required: true, message: i18next.t("login:Please input your RADIUS password!")}]}
      >
        <Input
          style={{width: "100%", marginTop: 12}}
          placeholder={i18next.t("general:Password")}
        />
      </Form.Item>
      <Form.Item
        name="enableMfaRemember"
        valuePropName="checked"
      >
        <Checkbox>
          {i18next.t("mfa:Remember this account for {hour} hours").replace("{hour}", mfaProps?.mfaRememberInHours)}
        </Checkbox>
      </Form.Item>
      <Form.Item>
        <Button
          style={{marginTop: 24}}
          loading={false}
          block
          type="primary"
          htmlType="submit"
        >
          {i18next.t("forget:Next Step")}
        </Button>
      </Form.Item>
    </Form>
  );
};

export default MfaVerifyRadiusForm;
