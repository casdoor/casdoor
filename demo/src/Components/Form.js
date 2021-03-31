import React from "react";
import logo from "../assets/logo.png";
import { Form, Input, Button, Checkbox } from "antd";
import {UserOutlined, LockOutlined} from '@ant-design/icons';
import ProviderLogin from "./ProviderLogin";




const FormComponent = () => {
  const [form] = Form.useForm();
  return (
    <Form
      form={form}
      layout="vertical"
      style={{
        paddingLeft: "40%",
        paddingRight: "40%",
        paddingTop: "10%",
        paddingBottom: "10%",
      }}

    >
      <a href="/" ><img style={{marginBottom:10}} src={logo} width="100%" alt="casbin-logo"/></a>
      <Form.Item
        name="username"
        rules={[{ required: true, message: "Please input your Username!" }]}
      >
        <Input
          prefix={<UserOutlined />}
          placeholder="username"
        />
      </Form.Item>
      <Form.Item
        name="password"
        rules={[{ required: true, message: "Please input your Password!" }]}
      >
        <Input
          prefix={<LockOutlined />}
          type="password"
          placeholder="password"
        />
      </Form.Item>
      <Form.Item>
        <Form.Item name="remember" valuePropName="checked" noStyle>
          <Checkbox
            style={{ float: "left" }}
          >
            Auto login
          </Checkbox>
        </Form.Item>
        <a style={{ float: "right" }} href="/">
          Forgot password?
        </a>
      </Form.Item>

      <Form.Item>
        <Button
          type="primary"
          htmlType="submit"
          style={{ width: "100%" }}
        >
          Login
        </Button>
        <div style={{ float: "right" }}>
          No account yet, <a href="/register">sign up now</a>
        </div>
      </Form.Item>
      <div style={{textAlign:"center"}}><ProviderLogin/></div>
    </Form>
  );
}
export default FormComponent;