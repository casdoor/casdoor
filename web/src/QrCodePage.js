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

import React from "react";
import QRCode from "qrcode.react";
import {Button, Col, Row} from "antd";
import * as PaymentBackend from "./backend/PaymentBackend";
import * as Setting from "./Setting";
import * as ProviderBackend from "./backend/ProviderBackend";
import i18next from "i18next";

class QrCodePage extends React.Component {
  constructor(props) {
    super(props);
    const params = new URLSearchParams(window.location.search);
    this.state = {
      classes: props,
      owner: props.owner ?? (props.match?.params?.owner ?? null),
      paymentName: props.paymentName ?? (props.match?.params?.paymentName ?? null),
      providerName: props.providerName ?? params.get("providerName"),
      payUrl: props.payUrl ?? params.get("payUrl"),
      successUrl: props.successUrl ?? params.get("successUrl"),
      provider: props.provider ?? null,
      payment: null,
      timer: null,
    };
  }

  async getProvider() {
    if (!this.state.owner || !this.state.providerName) {
      return ;
    }
    try {
      const res = await ProviderBackend.getProvider(this.state.owner, this.state.providerName);
      if (res.status !== "ok") {
        throw new Error(res.msg);
      }
      const provider = res.data;
      this.setState({
        provider: provider,
      });
    } catch (err) {
      Setting.showMessage("error", err.message);
      return ;
    }
  }

  setNotifyTask() {
    if (!this.state.owner || !this.state.paymentName) {
      return ;
    }

    const notifyTask = async() => {
      try {
        const res = await PaymentBackend.notifyPayment(this.state.owner, this.state.paymentName);
        if (res.status !== "ok") {
          throw new Error(res.msg);
        }
        const payment = res.data;
        if (payment.state !== "Created") {
          Setting.goToLink(this.state.successUrl);
        }
      } catch (err) {
        Setting.showMessage("error", err.message);
        return ;
      }
    };

    this.setState({
      timer: setTimeout(async() => {
        await notifyTask();
        this.setNotifyTask();
      }, 2000),
    });
  }

  componentDidMount() {
    if (this.props.onUpdateApplication) {
      this.props.onUpdateApplication(null);
    }
    this.getProvider();
    this.setNotifyTask();
  }

  componentWillUnmount() {
    clearInterval(this.state.timer);
  }

  renderProviderInfo(provider) {
    if (!provider) {
      return null;
    }
    const text = i18next.t(`product:${provider.type}`);
    return (
      <Button style={{height: "50px", borderWidth: "2px"}} shape="round" icon={
        <img style={{marginRight: "10px"}} width={36} height={36} src={Setting.getProviderLogoURL(provider)} alt={provider.displayName} />
      } size={"large"} >
        {
          text
        }
      </Button>
    );
  }

  render() {
    if (!this.state.payUrl || !this.state.successUrl || !this.state.owner || !this.state.paymentName) {
      return null;
    }
    return (
      <div className="login-content">
        <Col>
          <Row style={{justifyContent: "center"}}>
            {this.renderProviderInfo(this.state.provider)}
          </Row>
          <Row style={{marginTop: "10px", justifyContent: "center"}}>
            <QRCode value={this.state.payUrl} size={this.props.size ?? 200} />
          </Row>
        </Col>
      </div>
    );
  }
}

export default QrCodePage;
