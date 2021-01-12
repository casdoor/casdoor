import React from "react";
import {Button, Checkbox, Col, Form, Input, Row} from "antd";
import {LockOutlined, UserOutlined} from "@ant-design/icons";

class Face extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
    };
  }

  renderForm() {
    return (
      <Form
        name="normal_login"
        initialValues={{ remember: true }}
        // onFinish={this.onFinish.bind(this)}
        style={{width: "250px"}}
        size="large"
      >
        <Form.Item
          name="username"
          rules={[{ required: true, message: 'Please input your Username!' }]}
        >
          <Input
            prefix={<UserOutlined className="site-form-item-icon" />}
            placeholder="username"
          />
        </Form.Item>
        <Form.Item
          name="password"
          rules={[{ required: true, message: 'Please input your Password!' }]}
        >
          <Input
            prefix={<LockOutlined className="site-form-item-icon" />}
            type="password"
            placeholder="password"
          />
        </Form.Item>
        <Form.Item>
          <Form.Item name="remember" valuePropName="checked" noStyle>
            <Checkbox style={{float: "left"}}>Auto login</Checkbox>
          </Form.Item>
          <a style={{float: "right"}} href="">
            Forgot password?
          </a>
        </Form.Item>

        <Form.Item>
          <Button
            type="primary"
            htmlType="submit"
            style={{width: "100%"}}
          >
            Login
          </Button>
          <div style={{float: "right"}}>
            No account yet, <a href="/register">sign up now</a>
          </div>
        </Form.Item>
      </Form>
    );
  }

  render() {
    return (
      <Row>
        <Col span={24} style={{display: "flex", justifyContent:  "center"}} >
          <div style={{marginTop: "80px", textAlign: "center"}}>
            <img src={this.props.application.logo} alt={this.props.application.displayName} style={{marginBottom: '50px'}}/>
            {
              this.renderForm(this.props.application)
            }
          </div>
        </Col>
      </Row>
    )
  }
}

export default Face;
