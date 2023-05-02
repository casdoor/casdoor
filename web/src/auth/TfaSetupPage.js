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

import React, {useState} from "react";
import {Button, Col, Form, Input, Result, Row, Steps} from "antd";
import * as Setting from "../Setting";
import i18next from "i18next";
import * as TwoFactorBackend from "../backend/TfaBackend";
import {CheckOutlined, KeyOutlined, UserOutlined} from "@ant-design/icons";

import * as UserBackend from "../backend/UserBackend";
import {TfaSmsVerifyForm, TfaTotpVerifyForm} from "./TfaVerifyForms";

const {Step} = Steps;
export const SmsTfaType = "sms";
export const TotpTfaType = "app";

function CheckPasswordForm({user, onSuccess, onFail}) {
  const [form] = Form.useForm();

  const onFinish = ({password}) => {
    const data = {...user, password};
    UserBackend.checkUserPassword(data)
      .then((res) => {
        if (res.status === "ok") {
          onSuccess(res);
        } else {
          onFail(res);
        }
      })
      .finally(() => {
        form.setFieldsValue({password: ""});
      });
  };

  return (
    <Form
      form={form}
      style={{width: "300px"}}
      onFinish={onFinish}
    >
      <Form.Item
        name="password"
        rules={[{required: true, message: "Please input your password"}]}
      >
        <Input.Password
          prefix={<UserOutlined />}
          placeholder={i18next.t("two-factor:Password")}
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
}

export function TfaVerityForm({tfaProps, application, onSuccess, onFail}) {
  const [form] = Form.useForm();

  const onFinish = ({passcode}) => {
    const type = tfaProps.type;
    const data = {passcode, type};
    TwoFactorBackend.twoFactorSetupVerity(data)
      .then((res) => {
        if (res.status === "ok") {
          onSuccess(res);
        } else {
          onFail(res);
        }
      })
      .catch((error) => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      })
      .finally(() => {
        form.setFieldsValue({passcode: ""});
      });
  };

  if (tfaProps.type === SmsTfaType) {
    return <TfaSmsVerifyForm onFinish={onFinish} application={application} />;
  } else if (tfaProps.type === TotpTfaType) {
    return <TfaTotpVerifyForm onFinish={onFinish} tfaProps={tfaProps} />;
  } else {
    return <div></div>;
  }
}

function EnableTfaForm({userId, tfaProps, onSuccess, onFail}) {
  const [loading, setLoading] = useState(false);
  const requestEnableTotp = () => {
    const data = {
      type: tfaProps.type,
      userId: userId,
    };
    setLoading(true);
    TwoFactorBackend.twoFactorSetupEnable(data).then(res => {
      if (res.status === "ok") {
        onSuccess(res);
      } else {
        onFail(res);
      }
    }
    ).finally(() => {
      setLoading(false);
    });
  };

  return (
    <div style={{width: "400px"}}>
      <p>{i18next.t(
        "two-factor:Please save this recovery code. Once your device cannot provide an authentication code, you can reset two-factor authentication by this recovery code")}</p>
      <br />
      <code style={{fontStyle: "solid"}}>{tfaProps.recoveryCodes[0]}</code>
      <Button style={{marginTop: 24}} loading={loading} onClick={() => {
        requestEnableTotp();
      }} block type="primary">
        {i18next.t("two-factor:Enable")}
      </Button>
    </div>
  );
}

class TfaSetupPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      account: props.account,
      current: 0,
      type: props.type ?? SmsTfaType,
      tfaProps: null,
    };
  }

  getApplication() {
    return {
      owner: "admin",
      name: this.state.account.signupApplication,
    };
  }

  getUser() {
    return {
      name: this.state.account.name,
      owner: this.state.account.owner,
    };
  }

  getUserId() {
    return this.state.account.owner + "/" + this.state.account.name;
  }

  renderStep() {
    switch (this.state.current) {
    case 0:
      return <CheckPasswordForm
        user={this.getUser()}
        onSuccess={() => {
          TwoFactorBackend.twoFactorSetupInitiate({
            userId: this.getUserId(),
            type: this.state.type,
          }).then((res) => {
            if (res.status === "ok") {
              this.setState({
                current: this.state.current + 1,
                TFAProps: res.data,
              });
            } else {
              Setting.showMessage("error", i18next.t("tfa:initiate failed"));
            }
          });
        }}
        onFail={(res) => {
          Setting.showMessage("error", i18next.t("tfa:initiate failed"));
        }}
      />;
    case 1:
      return <TfaVerityForm tfaProps={{type: this.state.type, ...this.state.TFAProps}}
        application={this.getApplication()}
        onSuccess={() => {
          this.setState({
            current: this.state.current + 1,
          });
        }}
        onFail={(res) => {
          Setting.showMessage("error", i18next.t(`signup:${res.msg}`));
        }}
      />;
    case 2:
      return <EnableTfaForm userId={this.getUserId()} tfaProps={{type: this.state.type, ...this.state.TFAProps}}
        onSuccess={() => {
          Setting.showMessage("success", i18next.t("two-factor:Enabled successfully"));
          Setting.goToLinkSoft(this, "/account");
        }}
        onFail={(res) => {
          Setting.showMessage("error", i18next.t(`signup:${res.msg}`));
        }} />;
    default:
      return null;
    }
  }

  render() {
    if (!this.props.account) {
      return (
        <Result
          status="403"
          title="403 Unauthorized"
          subTitle={i18next.t("general:Sorry, you do not have permission to access this page or logged in status invalid.")}
          extra={<a href="/"><Button type="primary">{i18next.t("general:Back Home")}</Button></a>}
        />
      );
    }

    return (
      <Row>
        <Col span={24} style={{justifyContent: "center"}}>
          <Row>
            <Col span={24}>
              <div style={{textAlign: "center", fontSize: "28px"}}>
                {i18next.t("two-factor:Protect your account with two-factor authentication")}</div>
              <div style={{textAlign: "center", fontSize: "16px", marginTop: "10px"}}>{i18next.t("two-factor:Each time you sign in to your Account, you'll need your password and a authentication code")}</div>
            </Col>
          </Row>
          <Row>
            <Col span={24}>
              <Steps current={this.state.current} style={{
                width: "90%",
                maxWidth: "500px",
                margin: "auto",
                marginTop: "80px",
              }} >
                <Step title={i18next.t("two-factor:Verify Password")} icon={<UserOutlined />} />
                <Step title={i18next.t("two-factor:Verify Code")} icon={<KeyOutlined />} />
                <Step title={i18next.t("two-factor:Enable")} icon={<CheckOutlined />} />
              </Steps>
            </Col>
          </Row>
        </Col>
        <Col span={24} style={{display: "flex", justifyContent: "center"}}>
          <div style={{marginTop: "10px", textAlign: "center"}}>{this.renderStep()}</div>
        </Col>
      </Row>
    );
  }
}

export default TfaSetupPage;
