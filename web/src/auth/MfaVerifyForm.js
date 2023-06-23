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

import {Button, Col, Form, Input, QRCode, Space} from "antd";
import i18next from "i18next";
import {CopyOutlined, UserOutlined} from "@ant-design/icons";
import {SendCodeInput} from "../common/SendCodeInput";
import * as Setting from "../Setting";
import React, {useEffect} from "react";
import copy from "copy-to-clipboard";
import {CountryCodeSelect} from "../common/select/CountryCodeSelect";
import {EmailMfaType, SmsMfaType} from "./MfaSetupPage";

export const mfaAuth = "mfaAuth";
export const mfaSetup = "mfaSetup";

export const MfaSmsVerifyForm = ({mfaProps, application, onFinish, method, user}) => {
  const [dest, setDest] = React.useState(mfaProps.secret ?? "");
  const [form] = Form.useForm();

  useEffect(() => {
    if (mfaProps.mfaType === SmsMfaType) {
      setDest(user.phone);
    }

    if (mfaProps.mfaType === EmailMfaType) {
      setDest(user.email);
    }
  }, [mfaProps.mfaType]);

  const isEmail = () => {
    return mfaProps.mfaType === EmailMfaType;
  };

  return (
    <Form
      form={form}
      style={{width: "300px"}}
      onFinish={onFinish}
      initialValues={{
        countryCode: mfaProps.countryCode,
      }}
    >
      {dest !== "" ?
        <div style={{marginBottom: 20, textAlign: "left", gap: 8}}>
          {isEmail() ? i18next.t("mfa:Your email is") : i18next.t("mfa:Your phone is")} {dest}
        </div> :
        (<React.Fragment>
          <p>{isEmail() ? i18next.t("mfa:Please bind your email first, the system will automatically uses the mail for multi-factor authentication") :
            i18next.t("mfa:Please bind your phone first, the system automatically uses the phone for multi-factor authentication")}
          </p>
          <Input.Group compact style={{width: "300Px", marginBottom: "30px"}}>
            {isEmail() ? null :
              <Form.Item
                name="countryCode"
                noStyle
                rules={[
                  {
                    required: false,
                    message: i18next.t("signup:Please select your country code!"),
                  },
                ]}
              >
                <CountryCodeSelect
                  initValue={mfaProps.countryCode}
                  style={{width: "30%"}}
                  countryCodes={application.organizationObj.countryCodes}
                />
              </Form.Item>
            }
            <Form.Item
              name="dest"
              noStyle
              rules={[{required: true, message: i18next.t("login:Please input your Email or Phone!")}]}
            >
              <Input
                style={{width: isEmail() ? "100% " : "70%"}}
                onChange={(e) => {setDest(e.target.value);}}
                prefix={<UserOutlined />}
                placeholder={isEmail() ? i18next.t("general:Email") : i18next.t("general:Phone")}
              />
            </Form.Item>
          </Input.Group>
        </React.Fragment>
        )
      }
      <Form.Item
        name="passcode"
        rules={[{required: true, message: i18next.t("login:Please input your code!")}]}
      >
        <SendCodeInput
          countryCode={form.getFieldValue("countryCode")}
          method={method}
          onButtonClickArgs={[mfaProps.secret || dest, isEmail() ? "email" : "phone", Setting.getApplicationName(application)]}
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
          {i18next.t("forget:Next Step")}
        </Button>
      </Form.Item>
    </Form>
  );
};

export const MfaTotpVerifyForm = ({mfaProps, onFinish}) => {
  const [form] = Form.useForm();

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
        <p style={{textAlign: "center"}}>{i18next.t("mfa:Scan the QR code with your authenticator app")}</p>
        <p style={{textAlign: "center"}}>{i18next.t("mfa:Or copy the secret to your authenticator app")}</p>
        <Col span={24}>
          <Space>
            <Input value={mfaProps.secret} />
            <Button
              type="primary"
              shape="round"
              icon={<CopyOutlined />}
              onClick={() => {
                copy(`${mfaProps.secret}`);
                Setting.showMessage(
                  "success",
                  i18next.t("mfa:Multi-factor secret to clipboard successfully")
                );
              }}
            />
          </Space>
        </Col>
      </React.Fragment>
    );
  };
  return (
    <Form
      form={form}
      style={{width: "300px"}}
      onFinish={onFinish}
    >
      {renderSecret()}
      <Form.Item
        name="passcode"
        rules={[{required: true, message: "Please input your passcode"}]}
      >
        <Input
          style={{marginTop: 24}}
          prefix={<UserOutlined />}
          placeholder={i18next.t("mfa:Passcode")}
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
          {i18next.t("forget:Next Step")}
        </Button>
      </Form.Item>
    </Form>
  );
};
