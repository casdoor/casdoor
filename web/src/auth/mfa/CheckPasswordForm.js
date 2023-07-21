import {LockOutlined} from "@ant-design/icons";
import {Button, Form, Input} from "antd";
import i18next from "i18next";
import React from "react";
import * as UserBackend from "../../backend/UserBackend";

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
      style={{width: "300px", marginTop: "20px"}}
      onFinish={onFinish}
    >
      <Form.Item
        name="password"
        rules={[{required: true, message: i18next.t("login:Please input your password!")}]}
      >
        <Input.Password
          prefix={<LockOutlined />}
          placeholder={i18next.t("general:Password")}
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
          {i18next.t("forget:Next Step")}
        </Button>
      </Form.Item>
    </Form>
  );
}

export default CheckPasswordForm;
