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
import {Button, Col, Input, Result, Row, Steps} from "antd";
import * as Setting from "../Setting";
import i18next from "i18next";
import * as TwoFactorBackend from "../backend/TwoFactorAuthBackend";
import {CheckOutlined, CopyOutlined, KeyOutlined, UserOutlined} from "@ant-design/icons";
import QRCode from "qrcode.react";

import copy from "copy-to-clipboard";

const {Step} = Steps;

function CheckPassword({user, onSuccess, onFail}) {
  return (
    <form style={{width: "300px"}} >
      <Input
        prefix={<UserOutlined />}
        placeholder={i18next.t("two-factor:Password")}
        type="password"
      />
      <Button style={{marginTop: 24}}
        type="primary" htmlType="submit">
        {i18next.t("two-factor:Next step")}
      </Button>
    </form>
  );
}

function VerityTotp({totp, onSuccess, onFail}) {
  return (
    <form style={{width: "300px"}} >
      <QRCode value={totp.url} size={200} />
      <Row type="flex" justify="center" align="middle" >
        <Col>{Setting.getLabel(i18next.t("two-factor:Two-factor secret"), i18next.t("two-factor:Two-factor secret - Tooltip"))} :</Col>
      </Row>
      <Row type="flex" justify="center" align="middle" >
        <Col><Input value={totp.secret} /></Col>
        <Button type="primary" shape="round" icon={<CopyOutlined />} onClick={() => {
          copy(`${totp.secret}`);
          Setting.showMessage("success", i18next.t("two-factor:Two-factor secret to clipboard successfully"));
        }}></Button>
      </Row>
      <Input
        style={{marginTop: 24}}
        prefix={<UserOutlined />}
        placeholder={i18next.t("two-factor:Passcode")}
        type="text"
      />
      <Button style={{marginTop: 24}} block
        type="primary"
        htmlType="submit">
        {i18next.t("two-factor:Next step")}
      </Button>
    </form>
  );
}

function EnableTotp({user, totp, onSuccess, onFail}) {
  const [loading, setLoading] = useState(false);
  const requestEnableTotp = () => {
    const data = {
      userId: user.owner + "/" + user.name,
      secret: totp.secret,
      recoveryCode: totp.recoveryCode,
    };
    setLoading(true);
    TwoFactorBackend.twoFactorEnable(data).then(res => {
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
      <code style={{fontStyle: "solid"}}>{totp.recoveryCode}</code>
      <Button style={{marginTop: 24}} loading={loading} onClick={() => {
        requestEnableTotp();
      }} block type="primary">
        {i18next.t("two-factor:Enable")}
      </Button>
    </div>
  );
}

class TotpPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      account: props.account,
      current: 0,
      totp: null,
    };
  }

  getUser() {
    return {
      name: this.state.userName,
      owner: this.state.userOwner,
    };
  }

  getUserId() {
    return this.state.userOwner + "/" + this.state.userName;
  }

  renderStep() {
    switch (this.state.current) {
    case 0:
      return <CheckPassword
        user={this.getUser()}
        onSuccess={() => {
          TwoFactorBackend.twoFactorSetupInitiate({
            userId: this.getUserId(),
          }).then((res) => {
            if (res.status === "ok") {
              this.setState({
                totp: res.data,
                current: this.state.current + 1,
              });
            } else {
              Setting.showMessage("error",
                i18next.t(`signup:${res.msg}`));
            }
          });
        }}
        onFail={(res) => {
          Setting.showMessage("error",
            i18next.t(`signup:${res.msg}`));
        }}
      />;
    case 1:
      return <VerityTotp totp={this.state?.totp}
        onSuccess={() => {
          this.setState({
            current: this.state.current + 1,
          });
        }}
        onFail={(res) => {
          Setting.showMessage("error",
            i18next.t(`signup:${res.msg}`));
        }}
      />;
    case 2:
      return <EnableTotp user={this.getUser()} totp={this.state?.totp}
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

            </Col>
          </Row>
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

export default TotpPage;
