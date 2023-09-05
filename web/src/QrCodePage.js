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
import {useEffect, useState} from "react";
import QRCode from "qrcode.react";
import {Col, Row} from "antd";
import * as PaymentBackend from "./backend/PaymentBackend";
import * as Setting from "./Setting";

export default function QrCodePage({providerDisplay, owner, paymentName, payUrl, successUrl, size}) {
  window.console.log(payUrl, successUrl);
  const [paymentState, setPaymentState] = useState("Created");

  useEffect(() => {
    if (!owner || !paymentName) {
      return ;
    }

    const notifyTask = async() => {
      try {
        const res = await PaymentBackend.notifyPayment(owner, paymentName);
        if (res.status !== "ok") {
          throw new Error(res.msg);
        }
        const payment = res.data;
        setPaymentState(payment.state);
      } catch (err) {
        Setting.showMessage("error", err.message);
        return ;
      }
    };

    const timer = setInterval(() => {
      if (paymentState === "Created") {
        notifyTask();
      }
    }, 2000);

    return () => {
      clearInterval(timer);
    };
  });

  useEffect(() => {
    if (paymentState !== "Created") {
      // the successUrl is redirected from payUrl after pay success, not the product's returnUrl
      Setting.goToLink(successUrl);
    }
  }, [paymentState]);

  if (!payUrl || !successUrl) {
    return null;
  }
  return (
    <Col>
      <Row style={{justifyContent: "center"}}>
        {providerDisplay}
      </Row>
      <Row style={{marginTop: "10px", justifyContent: "center"}}>
        <QRCode value={payUrl} size={size} />
      </Row>
    </Col>
  );
}
