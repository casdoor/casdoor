import {Button, Col, Form, Input, Row} from "antd";
import i18next from "i18next";
import {CopyOutlined, LockOutlined, UserOutlined} from "@ant-design/icons";
import {SendCodeInput} from "../common/SendCodeInput";
import * as Setting from "../Setting";
import React from "react";
import QRCode from "qrcode.react";
import copy from "copy-to-clipboard";

export const TfaSmsVerifyForm = ({application, onFinish}) => {
  const [dest, setDest] = React.useState("");
  const [form] = Form.useForm();
  return (
    <Form
      form={form}
      style={{width: "300px"}}
      onFinish={onFinish}
    >
      <Form.Item
        name="dest"
        rules={[{required: true, message: i18next.t("login:Please input your Phone or email!")}]}
      >
        <Input
          onChange={(e) => {setDest(e.target.value);}}
          prefix={<LockOutlined />}
          placeholder={i18next.t("general:Phone or email")}
        />
      </Form.Item>
      <Form.Item
        name="passcode"
        rules={[{required: true, message: i18next.t("login:Please input your code!")}]}
      >
        <SendCodeInput
          method={"tfa"}
          onButtonClickArgs={[dest, "email", Setting.getApplicationName(application)]}
          application={application}
        />
      </Form.Item>
      <Form.Item>
        <Button
          style={{marginTop: 24}}
          loading={false}
          block
          type="primary"
          htmlType="submit"
        >
          {i18next.t("two-factor:Next step")}
        </Button>
      </Form.Item>
    </Form>
  );
};

export const TfaTotpVerifyForm = ({tfaProps, onFinish}) => {
  const [form] = Form.useForm();

  return (
    <Form
      form={form}
      style={{width: "300px"}}
      onFinish={onFinish}
    >
      <Row type="flex" justify="center" align="middle">
        <Col>
          <QRCode value={tfaProps.url} size={200} />
        </Col>
      </Row>

      <Row type="flex" justify="center" align="middle">
        <Col>
          {Setting.getLabel(
            i18next.t("two-factor:Two-factor secret"),
            i18next.t("two-factor:Two-factor secret - Tooltip")
          )}
        :
        </Col>
        <Col>
          <Input value={tfaProps.secret} />
        </Col>
        <Col>
          <Button
            type="primary"
            shape="round"
            icon={<CopyOutlined />}
            onClick={() => {
              copy(`${tfaProps.secret}`);
              Setting.showMessage(
                "success",
                i18next.t("two-factor:Two-factor secret to clipboard successfully")
              );
            }}
          />
        </Col>
      </Row>

      <Form.Item
        name="passcode"
        rules={[{required: true, message: "Please input your passcode"}]}
      >
        <Input
          style={{marginTop: 24}}
          prefix={<UserOutlined />}
          placeholder={i18next.t("two-factor:Passcode")}
        />
      </Form.Item>

      <Form.Item>
        <Button
          style={{marginTop: 24}}
          loading={false}
          block
          type="primary"
          htmlType="submit"
        >
          {i18next.t("two-factor:Next step")}
        </Button>
      </Form.Item>
    </Form>
  );
};
