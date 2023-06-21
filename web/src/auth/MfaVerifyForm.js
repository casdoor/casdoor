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

import {Button, Col, Form, Input, Row} from "antd";
import i18next from "i18next";
import {CopyOutlined, UserOutlined} from "@ant-design/icons";
import {SendCodeInput} from "../common/SendCodeInput";
import * as Setting from "../Setting";
import React from "react";
import copy from "copy-to-clipboard";
import {CountryCodeSelect} from "../common/select/CountryCodeSelect";
import {EmailMfaType} from "./MfaSetupPage";

export const mfaAuth = "mfaAuth";
export const mfaSetup = "mfaSetup";

export const MfaSmsVerifyForm = ({mfaProps, application, onFinish, method}) => {
  const [dest, setDest] = React.useState(mfaProps.secret ?? "");
  const [form] = Form.useForm();
  const mfaType = mfaProps.mfaType;

  return (
    <Form
      form={form}
      style={{width: "300px"}}
      onFinish={onFinish}
      initialValues={{
        countryCode: mfaProps.countryCode,
      }}
    >
      {mfaProps.secret !== undefined && mfaProps.secret !== "" ?
        <div style={{marginBottom: 20, textAlign: "left", gap: 8}}>
          {mfaType === EmailMfaType ? i18next.t("mfa:Your email is") : i18next.t("mfa:Your phone is")} {mfaProps.secret}
        </div> :
        (<React.Fragment>
          <p>{mfaType === EmailMfaType ? i18next.t("mfa:Please bind your email first, the system will automatically uses the mail for multi-factor authentication") :
            i18next.t("mfa:Please bind your phone first, the system automatically uses the phone for multi-factor authentication")}
          </p>
          <Input.Group compact style={{width: "300Px", marginBottom: "30px"}}>
            {mfaType === EmailMfaType ? null :
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
                style={{width: mfaType === EmailMfaType ? "100% " : "70%"}}
                onChange={(e) => {setDest(e.target.value);}}
                prefix={<UserOutlined />}
                placeholder={mfaType === EmailMfaType ? i18next.t("general:Email") : i18next.t("general:Phone")}
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
          onButtonClickArgs={[dest, mfaType === EmailMfaType ? "email" : "phone", Setting.getApplicationName(application)]}
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

  return (
    <Form
      form={form}
      style={{width: "300px"}}
      onFinish={onFinish}
    >
      <Row type="flex" justify="center" align="middle">
        <Col>
        </Col>
      </Row>

      <Row type="flex" justify="center" align="middle">
        <Col>
          {Setting.getLabel(
            i18next.t("mfa:Multi-factor secret"),
            i18next.t("mfa:Multi-factor secret - Tooltip")
          )}
        :
        </Col>
        <Col>
          <Input value={mfaProps.secret} />
        </Col>
        <Col>
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
        </Col>
      </Row>

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
