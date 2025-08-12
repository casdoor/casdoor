import {CopyOutlined} from "@ant-design/icons";
import {Button, Checkbox, Col, Form, Input, QRCode, Space} from "antd";
import copy from "copy-to-clipboard";
import i18next from "i18next";
import React from "react";
import * as Setting from "../../Setting";

export const MfaVerifyTotpForm = ({mfaProps, onFinish}) => {
  const [form] = Form.useForm();

  const handleFinish = (values) => {
    onFinish({
      passcode: values.passcode,
      enableMfaRemember: values.enableMfaRemember,
    });
  };

  const renderSecret = () => {
    if (!mfaProps.secret) {
      return null;
    }

    return (
      <React.Fragment>
        <Col span={24} style={{display: "flex", justifyContent: "center"}}>
          <QRCode
            errorLevel="H"
            value={mfaProps.url}
            icon={"https://cdn.casdoor.com/static/favicon.png"}
          />
        </Col>
        <p style={{textAlign: "center"}}>{i18next.t("mfa:Scan the QR code with your Authenticator App")}</p>
        <p style={{textAlign: "center"}}>{i18next.t("mfa:Or copy the secret to your Authenticator App")}</p>
        <Col span={24}>
          <Space>
            <Input value={mfaProps.secret} />
            <Button type="primary" shape="round" icon={<CopyOutlined />} onClick={() => {
              copy(`${mfaProps.secret}`);
              Setting.showMessage("success", i18next.t("general:Copied to clipboard successfully"));
            }} />
          </Space>
        </Col>
      </React.Fragment>
    );
  };
  return (
    <Form
      form={form}
      style={{width: "300px"}}
      onFinish={handleFinish}
      initialValues={{
        enableMfaRemember: false,
      }}
    >
      {renderSecret()}
      <Form.Item
        name="passcode"
        rules={[{required: true, message: "Please input your passcode"}]}
      >
        <Input.OTP
          style={{marginTop: 24}}
          onChange={() => {
            form.submit();
          }}
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

export default MfaVerifyTotpForm;
