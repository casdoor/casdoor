// Copyright 2021 The casbin Authors. All Rights Reserved.
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
import {withRouter} from "react-router-dom";
import * as PaymentBackend from "./backend/PaymentBackend"

class PaySuccessCallback extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      msg: null,
      paymentId: this.GetQueryString("token"),
      payerId: this.GetQueryString("PayerID")
    };
  }

  // eslint-disable-next-line react/no-deprecated
  componentWillMount() {
    PaymentBackend.PaySuccess(this.state.paymentId, this.state.payerId).then(res => {
      window.location.href = res
    })
  }

  GetQueryString(key){
    let reg = new RegExp("(^|&)"+ key +"=([^&]*)(&|$)");
    let r = window.location.search.substr(1).match(reg);
    if(r!=null)return  unescape(r[2]); return null;
  }

  render() {
    return (
      <div style={{textAlign: "center"}}>
        {
          (this.state.msg === null) ? (
            <Spin size="large" tip="支付完成，正在跳转..." style={{paddingTop: "10%"}} />
          ) : (
            <div style={{display: "inline"}}>
              <Result
                status="error"
                title="Login Error"
                subTitle={this.state.msg}
                extra={[
                  <Button type="primary" key="details">
                    Details
                  </Button>,
                  <Button key="help">Help</Button>,
                ]}
              />
            </div>
          )
        }
      </div>
    );
  }
}

export default withRouter(PaySuccessCallback);
