// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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
import {Button, Result, Spin} from "antd";
import * as PaymentBackend from "./backend/PaymentBackend";
import * as PricingBackend from "./backend/PricingBackend";
import * as SubscriptionBackend from "./backend/SubscriptionBackend";
import * as UserBackend from "./backend/UserBackend";
import * as Setting from "./Setting";
import i18next from "i18next";

class PaymentResultPage extends React.Component {
  constructor(props) {
    super(props);
    const params = new URLSearchParams(window.location.search);
    this.state = {
      classes: props,
      owner: props.match?.params?.organizationName ?? props.match?.params?.owner ?? null,
      paymentName: props.match?.params?.paymentName ?? null,
      pricingName: props.pricingName ?? props.match?.params?.pricingName ?? null,
      subscriptionName: params.get("subscription"),
      payment: null,
      pricing: props.pricing ?? null,
      subscription: props.subscription ?? null,
      timeout: null,
      user: null,
      source: params.get("source"),
    };
  }

  UNSAFE_componentWillMount() {
    this.getPayment();
  }

  getUser() {
    UserBackend.getUser(this.props.account.owner, this.props.account.name)
      .then((res) => {
        if (res.data === null) {
          this.props.history.push("/404");
          return;
        }

        if (res.status === "error") {
          Setting.showMessage("error", res.msg);
          return;
        }

        this.setState({
          user: res.data,
        });
      });
  }

  componentWillUnmount() {
    if (this.state.timeout !== null) {
      clearTimeout(this.state.timeout);
    }
  }

  setStateAsync(state) {
    return new Promise((resolve, reject) => {
      this.setState(state, () => {
        resolve();
      });
    });
  }

  onUpdatePricing(pricing) {
    this.props.onUpdatePricing(pricing);
  }

  async getPayment() {
    if (!(this.state.owner && (this.state.paymentName || (this.state.pricingName && this.state.subscriptionName)))) {
      return ;
    }
    try {
      // loading price & subscription
      if (this.state.pricingName && this.state.subscriptionName) {
        if (!this.state.pricing) {
          const res = await PricingBackend.getPricing(this.state.owner, this.state.pricingName);
          if (res.status !== "ok") {
            throw new Error(res.msg);
          }
          const pricing = res.data;
          await this.setStateAsync({
            pricing: pricing,
          });
        }
        if (!this.state.subscription) {
          const res = await SubscriptionBackend.getSubscription(this.state.owner, this.state.subscriptionName);
          if (res.status !== "ok") {
            throw new Error(res.msg);
          }
          const subscription = res.data;
          await this.setStateAsync({
            subscription: subscription,
          });
        }
        const paymentName = this.state.subscription.payment;
        await this.setStateAsync({
          paymentName: paymentName,
        });
        this.onUpdatePricing(this.state.pricing);
      }
      const res = await PaymentBackend.getPayment(this.state.owner, this.state.paymentName);
      if (res.status !== "ok") {
        throw new Error(res.msg);
      }
      const payment = res.data;
      this.processAutoReturn(payment);

      await this.setStateAsync({
        payment: payment,
      });
      if (payment.state === "Created") {
        if (["PayPal", "Stripe", "AirWallex", "Alipay", "WeChat Pay", "Balance"].includes(payment.type)) {
          this.setState({
            timeout: setTimeout(async() => {
              await PaymentBackend.notifyPayment(this.state.owner, this.state.paymentName);
              this.getPayment();
            }, 1000),
          });
        } else {
          this.setState({
            timeout: setTimeout(() => this.getPayment(), 1000),
          });
        }
      }

      if (payment.state === "Paid") {
        if (this.props.account) {
          this.getUser();
        }
      }
    } catch (err) {
      Setting.showMessage("error", err.message);
      return;
    }
  }

  goToPaymentUrl(payment) {
    if (payment.returnUrl === undefined || payment.returnUrl === null || payment.returnUrl === "") {
      Setting.goToLink(`${window.location.origin}/products/${payment.owner}/${payment.productName}/buy`);
    } else {
      // append some usefull callback parameter to SuccessUrl      
      const returnUrl = new URL(payment.returnUrl);
      const params = new URLSearchParams(returnUrl.search);
      params.set("owner", payment.owner);
      params.set("paymentName", payment.name);
      returnUrl.search = params.toString();

      Setting.goToLink(returnUrl.toString());
    }
  }

  processAutoReturn(payment) {
    // only auto return from pay callback
    if (this.state.source !== "pay") {
      return;
    }
    const {state, returnType} = payment;
    if (returnType === "autoClose") {
      window.close();
      return;
    }
    if (state === "Paid") {
      if (returnType === "paidAutoRedirect") {
        this.goToPaymentUrl(payment);
        return;
      }
      if (returnType === "paidAutoClose") {
        window.close();
        return;
      }
    }
  }

  render() {
    const payment = this.state.payment;

    if (payment === null) {
      return null;
    }

    if (payment.state === "Paid") {
      if (payment.isRecharge) {
        return (
          <div className="login-content">
            {
              Setting.renderHelmet(payment)
            }
            <Result
              status="success"
              title={`${i18next.t("payment:Recharged successfully")}`}
              subTitle={`${i18next.t("payment:You have successfully recharged")} ${payment.price} ${Setting.getCurrencyText(payment)}, ${i18next.t("payment:Your current balance is")} ${this.state.user?.balance} ${Setting.getCurrencyText(payment)}`}
              extra={[
                <Button type="primary" key="returnUrl" onClick={() => {
                  this.goToPaymentUrl(payment);
                }}>
                  {i18next.t("payment:Return to Website")}
                </Button>,
              ]}
            />
          </div>
        );
      }
      return (
        <div className="login-content">
          {
            Setting.renderHelmet(payment)
          }
          <Result
            status="success"
            title={`${i18next.t("payment:You have successfully completed the payment")}: ${payment.productDisplayName}`}
            subTitle={i18next.t("payment:Please click the below button to return to the original website")}
            extra={[
              <Button type="primary" key="returnUrl" onClick={() => {
                this.goToPaymentUrl(payment);
              }}>
                {i18next.t("payment:Return to Website")}
              </Button>,
            ]}
          />
        </div>
      );
    } else if (payment.state === "Created") {
      return (
        <div className="login-content">
          {
            Setting.renderHelmet(payment)
          }
          <Result
            status="info"
            title={`${i18next.t("payment:The payment is still under processing")}: ${payment.productDisplayName}, ${i18next.t("payment:the current state is")}: ${payment.state}, ${i18next.t("payment:please wait for a few seconds...")}`}
            subTitle={i18next.t("payment:Please click the below button to return to the original website")}
            extra={[
              <Spin key="returnUrl" size="large" tip={i18next.t("payment:Processing...")} />,
            ]}
          />
        </div>
      );
    } else if (payment.state === "Canceled") {
      return (
        <div className="login-content">
          {
            Setting.renderHelmet(payment)
          }
          <Result
            status="warning"
            title={`${i18next.t("payment:The payment has been canceled")}: ${payment.productDisplayName}, ${i18next.t("payment:the current state is")}: ${payment.state}`}
            subTitle={i18next.t("payment:Please click the below button to return to the original website")}
            extra={[
              <Button type="primary" key="returnUrl" onClick={() => {
                this.goToPaymentUrl(payment);
              }}>
                {i18next.t("payment:Return to Website")}
              </Button>,
            ]}
          />
        </div>
      );
    } else if (payment.state === "Timeout") {
      return (
        <div className="login-content">
          {
            Setting.renderHelmet(payment)
          }
          <Result
            status="warning"
            title={`${i18next.t("payment:The payment has time out")}: ${payment.productDisplayName}, ${i18next.t("payment:the current state is")}: ${payment.state}`}
            subTitle={i18next.t("payment:Please click the below button to return to the original website")}
            extra={[
              <Button type="primary" key="returnUrl" onClick={() => {
                this.goToPaymentUrl(payment);
              }}>
                {i18next.t("payment:Return to Website")}
              </Button>,
            ]}
          />
        </div>
      );
    } else {
      return (
        <div className="login-content">
          {
            Setting.renderHelmet(payment)
          }
          <Result
            status="error"
            title={`${i18next.t("payment:The payment has failed")}: ${payment.productDisplayName}, ${i18next.t("payment:the current state is")}: ${payment.state}`}
            subTitle={`${i18next.t("payment:Failed reason")}: ${payment.message}`}
            extra={[
              <Button type="primary" key="returnUrl" onClick={() => {
                this.goToPaymentUrl(payment);
              }}>
                {i18next.t("payment:Return to Website")}
              </Button>,
            ]}
          />
        </div>
      );
    }
  }
}

export default PaymentResultPage;
