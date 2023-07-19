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
import i18next from "i18next";
import {Button, Form, Input} from "antd";
import * as Setting from "../Setting";
import * as PasswordChecker from "../common/PasswordChecker";
import * as UserBackend from "../backend/UserBackend";

export function ChangePasswordForm({application, userOwner, userName, onSuccess, onFail}) {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);

  const setPassword = (values) => {
    setLoading(true);

    UserBackend.setPassword(values.userOwner, values.userName, values.oldPassword, values.newPassword)
      .then((res) => {
        if (res.status === "ok") {
          onSuccess(values);
        } else {
          onFail(res);
        }
      }).finally(() => setLoading(false));
  };

  return (
    <Form
      labelCol={{span: 8}}
      wrapperCol={{span: 16}}
      form={form}
      name="changePassword"
      onFinish={(values) => setPassword(values)}
      initialValues={{
        application: application.name,
        userOwner: userOwner,
        userName: userName,
      }}
      size="large"
      layout={Setting.isMobile() ? "vertical" : "horizontal"}
      style={{width: Setting.isMobile() ? "300px" : "400px"}}
    >
      <Form.Item
        name="application"
        hidden={true}
        rules={[
          {
            required: true,
            message: "Please input your application!",
          },
        ]}
      >
      </Form.Item>
      <Form.Item
        name="userOwner"
        hidden={true}
        rules={[
          {
            required: true,
            message: "Please input user owner!",
          },
        ]}
      >
      </Form.Item>
      <Form.Item
        name="userName"
        hidden={true}
        rules={[
          {
            required: true,
            message: "Please input user name!",
          },
        ]}
      >
      </Form.Item>
      <Form.Item
        name="oldPassword"
        label={i18next.t("user:Old Password")}
        rules={[
          {
            required: true,
            message: i18next.t("user:Empty input!"),
          },
        ]}
        hasFeedback
      >
        <Input.Password />
      </Form.Item>
      <Form.Item
        name="newPassword"
        label={i18next.t("user:New Password")}
        rules={[
          {
            required: true,
            message: i18next.t("user:Empty input!"),
          },
          {
            validator(rule, value) {
              const errorMsg = PasswordChecker.checkPasswordComplexity(value, application.organizationObj.passwordOptions);
              if (errorMsg === "") {
                return Promise.resolve();
              } else {
                return Promise.reject(errorMsg);
              }
            },
          },
        ]}
        hasFeedback
      >
        <Input.Password />
      </Form.Item>
      <Form.Item
        name="confirm"
        label={i18next.t("user:Re-enter New")}
        dependencies={["newPassword"]}
        hasFeedback
        rules={[
          {
            required: true,
            message: i18next.t("user:Empty input!"),
          },
          ({getFieldValue}) => ({
            validator(rule, value) {
              if (!value || getFieldValue("newPassword") === value) {
                return Promise.resolve();
              }

              return Promise.reject(i18next.t("signup:Your confirmed password is inconsistent with the password!"));
            },
          }),
        ]}
      >
        <Input.Password />
      </Form.Item>

      <Form.Item
        wrapperCol={{span: 24}}>
        <Button type="primary" htmlType="submit" loading={loading}>
          {i18next.t("changePassword:Change password")}
        </Button>
      </Form.Item>
    </Form>
  );
}
