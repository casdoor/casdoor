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
import {Button, Result, Spin, Form, Input,} from "antd";
import {withRouter} from "react-router-dom";
import * as PaymentBackend from "./backend/PaymentBackend"
import * as ApplicationBackend from "./backend/ApplicationBackend"

class Pay extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      msg: null,
      classes: props,
      clientId: this.GetQueryString("clientId"),
      invoice: this.GetQueryString("invoice"),
      price: this.GetQueryString("price"),
      currency: this.GetQueryString("currency"),
      description: this.GetQueryString("description"),
      redirectUri: this.GetQueryString("redirectUri"),
      waiting: false,
      applicationName: "",
      wrongMsg: "",
      spinMsg: "正在生成订单..."
    };
  }

  UNSAFE_componentWillMount() {
    this.getApplication();
  }

  getApplication(){
    ApplicationBackend.getApplicationByClientId(this.state.clientId).then(res => {
      if(res == null){
        this.setState({
          wrongMsg: "Invalid clientID"
        })
      }else {
        console.log("res")
        console.log(res)
        this.setState({
          applicationName : res.displayName
        })
      }
    })
  }
  GetQueryString(key){
    let reg = new RegExp("(^|&)"+ key +"=([^&]*)(&|$)");
    let r = window.location.search.substr(1).match(reg);
    if(r!=null)return  unescape(r[2]); return null;
  }

  getCurrencyString(currency){
    if(currency === "USD"){
      return "$"
    }else if (currency === "CNY"){
      return "￥"
    }else if (currency === "EUR"){
      return "€"
    }
    return "$"
  }


  pay(){
    let payItem = {
      invoice: this.state.invoice,
      price : this.state.price,
      currency : this.state.currency,
      description : this.state.description
    }
    this.setState({
      waiting: true
    })

    PaymentBackend.PaypalPal(payItem, this.state.clientId, this.state.redirectUri).then(res => {
      console.log(res)
      if (res.indexOf("http") !== -1){
        this.setState({
          spinMsg: "前往支付页面"
        })
        window.location.replace(res);
      }
      else {
        this.setState({
          wrongMsg : res
        })
      }
    })
  }

  render() {
    const data = [
      {
        title: 'Invoice',
      },
      {
        title: 'Amount',
      },
      {
        title: 'Description',
      }
    ];
    return (
        <div>
          {
            (this.state.wrongMsg === "") ?
            (<div style={{textAlign: "center"}}>
              {
                (this.state.waiting) ? (
                    <Spin size="large" tip={this.state.spinMsg} style={{paddingTop: "10%"}} />
                ) : (
                    <div style={{display: "inline"}}>
                      <p style={{fontSize: 20}}>{`尊敬的用户,您正在通过 casdoor 向 ${this.state.applicationName} 进行付款,请确认`}</p>
                      <Form
                          size="large"
                          name="pay"
                          labelCol={{ span: 8 }}
                          wrapperCol={{ span: 8 }}
                          initialValues={{ remember: true }}
                          autoComplete="off"
                      >
                        <Form.Item
                            label="Invoice"
                        >
                          <Input style={{color: "black"}} value={this.state.invoice} bordered={false} disabled />
                        </Form.Item>

                        <Form.Item
                            label="Account"
                        >
                          <Input style={{color: "black"}} disabled bordered={false} value={`${this.getCurrencyString(this.state.currency)} ${this.state.price}`}/>
                        </Form.Item>

                        <Form.Item
                            label="Description"
                        >
                          <Input style={{color: "black"}} bordered={false} disabled value={`${this.state.description}`}/>
                        </Form.Item>

                        <Form.Item wrapperCol={{ offset: 8, span: 8 }}>
                          <Button size="large" type="primary" onClick={() => this.pay()}>
                            确认付款
                          </Button>
                        </Form.Item>
                      </Form>
                    </div>
                )
              }
            </div>) :
            ( <div style={{display: "inline"}}>
                  <Result
                      status="error"
                      title="Pay Error"
                      subTitle={this.state.wrongMsg}
                  />
                </div>)
          }
        </div>
    );
  }
}

export default withRouter(Pay);
