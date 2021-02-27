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

import React from 'react';
import i18next from "i18next";
import {Button, Form, Input} from 'antd';
import * as OAuthAppBackend from "./backend/OAuthAppBackend";
import * as Setting from "./Setting";

class RegisterOAuthApplication extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      account: props.account
    };
  }

  cancel() {
    this.props.history.push(`/oauth/`);
  }

  onFinish(values) {
    let client = {
      name: values.name,
      domain: values.domain,
      callback: values.callback,
      userId: this.state.account.name
    }
    OAuthAppBackend.registerOAuthApp(client).then((res) => {
      if (res) {
        Setting.showMessage("success", `Successfully Registered`);
        this.props.history.push(`/oauth`);
      } else {
        Setting.showMessage("error", `failed to register: server side failure`);
        this.props.history.push(`/oauth`);
      }
    })
    .catch(error => {
      Setting.showMessage("error", `failed to register: ${error}`);
    });
  };

  render() {
    return(
      <div>
        <h2 style={{marginLeft:"600px", marginTop:"50px", fontSize:"24px"}}>{i18next.t("oauth:Register a new OAuth application")}</h2>
        <Form 
          style={{marginLeft:"600px", marginTop:"20px", fontSize:"14px"}}
          onFinish={this.onFinish.bind(this)}
          layout="vertical"
          size="large"
        >
          <Form.Item
            name="name"
            label={i18next.t("oauth:Application Name")}
            rules={[{ required: true, message: 'Please input your application name!' }]}
          >
            <Input style={{width:"450px"}} placeholder="casdoor"/>
          </Form.Item>
          <Form.Item
            name="domain"
            label={i18next.t("oauth:Homepage URL")}
            rules={[{ required: true, message: 'Please input your Homepage URL!' }]}
          >
            <Input type="text" style={{width:"450px"}} placeholder="localhost"/>
          </Form.Item>
          <Form.Item
            name="callback"
            label={i18next.t("oauth:Callback URL")}
            rules={[{ required: true, message: 'Please input your Callback URL!' }]}
          >
            <Input type="text" style={{width:"450px"}} placeholder="localhost"/>
          </Form.Item>
          <Form.Item>
            <Button style={{marginTop:"20px"}} type="primary" shape="round" htmlType="submit">{i18next.t("oauth:Register Application")}</Button>
            <Button type="text" onClick={this.cancel.bind(this)}>{i18next.t("general:Cancel")}</Button>
          </Form.Item>
        </Form>
      </div>
    )
  }
}

export default RegisterOAuthApplication;