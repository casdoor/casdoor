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
import {Link} from "react-router-dom";
import {Form, Input, Select, Checkbox, Button, Row, Col} from 'antd';
import * as Setting from "../Setting";
import * as AuthBackend from "./AuthBackend";

const { Option } = Select;

const formItemLayout = {
  labelCol: {
    xs: {
      span: 24,
    },
    sm: {
      span: 8,
    },
  },
  wrapperCol: {
    xs: {
      span: 24,
    },
    sm: {
      span: 16,
    },
  },
};

const tailFormItemLayout = {
  wrapperCol: {
    xs: {
      span: 24,
      offset: 0,
    },
    sm: {
      span: 16,
      offset: 8,
    },
  },
};

class RegisterPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
    };

    this.form = React.createRef();
  }

  onFinish(values) {
    AuthBackend.register(values)
      .then((res) => {
        if (res.status === 'ok') {
          this.props.history.push('/result');
        } else {
          Setting.showMessage("error", `Failed to register: ${res.msg}`);
        }
      });
  }

  onFinishFailed(values, errorFields, outOfDate) {
    this.form.current.scrollToField(errorFields[0].name);
  }

  renderForm() {
    const prefixSelector = (
      <Form.Item name="prefix" noStyle>
        <Select
          style={{
            width: 80,
          }}
        >
          <Option value="1">+1</Option>
          <Option value="86">+86</Option>
        </Select>
      </Form.Item>
    );

    return (
      <Form
        {...formItemLayout}
        ref={this.form}
        name="register"
        onFinish={(values) => this.onFinish(values)}
        onFinishFailed={(errorInfo) => this.onFinishFailed(errorInfo.values, errorInfo.errorFields, errorInfo.outOfDate)}
        initialValues={{
          prefix: '86',
        }}
        style={{width: !Setting.isMobile() ? "400px" : "250px"}}
        size="large"
      >
        <Form.Item
          name="name"
          label="Username"
          rules={[
            {
              required: true,
              message: 'Please input your username',
              whitespace: true,
            },
          ]}
        >
          <Input />
        </Form.Item>
        <Form.Item
          name="displayName"
          label="Display name"
          rules={[
            {
              required: true,
              message: 'Please input your display name',
              whitespace: true,
            },
          ]}
        >
          <Input />
        </Form.Item>
        <Form.Item
          name="affiliation"
          label="Affiliation"
          rules={[
            {
              required: true,
              message: 'Please input your affiliation',
              whitespace: true,
            },
          ]}
        >
          <Input />
        </Form.Item>
        <Form.Item
          name="email"
          label="Email"
          rules={[
            {
              type: 'email',
              message: 'The input is not valid Email!',
            },
            {
              required: true,
              message: 'Please input your Email',
            },
          ]}
        >
          <Input />
        </Form.Item>
        <Form.Item
          name="password"
          label="Password"
          rules={[
            {
              required: true,
              message: 'Please input your password',
            },
          ]}
          hasFeedback
        >
          <Input.Password />
        </Form.Item>
        <Form.Item
          name="confirm"
          label="Confirm"
          dependencies={['password']}
          hasFeedback
          rules={[
            {
              required: true,
              message: 'Please confirm your password',
            },
            ({ getFieldValue }) => ({
              validator(rule, value) {
                if (!value || getFieldValue('password') === value) {
                  return Promise.resolve();
                }

                return Promise.reject('Your confirmed password is inconsistent with the password');
              },
            }),
          ]}
        >
          <Input.Password />
        </Form.Item>
        <Form.Item
          name="phone"
          label="Phone number"
          rules={[
            {
              required: true,
              message: 'Please confirm your phone number',
            },
          ]}
        >
          <Input
            addonBefore={prefixSelector}
            style={{
              width: '100%',
            }}
          />
        </Form.Item>
        <Form.Item name="agreement" valuePropName="checked" {...tailFormItemLayout}>
          <Checkbox>
            Accept&nbsp;
            <Link to={"/agreement"}>
              Terms of Use
            </Link>
          </Checkbox>
        </Form.Item>
        <Form.Item {...tailFormItemLayout}>
          <Button type="primary" htmlType="submit">
            Sign Up
          </Button>
          &nbsp;&nbsp;&nbsp;Have account?&nbsp;
          <Link to={"/login"}>
            sign in now
          </Link>
        </Form.Item>
      </Form>
    )
  }

  render() {
    return (
      <div>
        &nbsp;
        <Row>
          <Col span={24} style={{display: "flex", justifyContent:  "center"}} >
            {
              this.renderForm()
            }
          </Col>
        </Row>
      </div>
    )
  }
}

export default RegisterPage;
