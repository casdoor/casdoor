import React from "react";
import {Button, Checkbox, Col, Form, Input, Row} from "antd";
import {LockOutlined, UserOutlined} from "@ant-design/icons";
import * as ApplicationBackend from "./backend/ApplicationBackend";
import * as AccountBackend from "./backend/AccountBackend";
import * as Setting from "./Setting";

class Face extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      applicationName: props.match === undefined ? null : props.match.params.applicationName,
      application: null,
    };
  }

  componentWillMount() {
    this.getApplication();
  }

  getApplication() {
    ApplicationBackend.getApplication("admin", this.state.applicationName)
      .then((application) => {
        this.setState({
          application: application,
        });
      });
  }

  getApplicationObj() {
    if (this.props.application !== undefined) {
      return this.props.application;
    } else {
      return this.state.application;
    }
  }

  onFinish(values) {
    AccountBackend.login(values)
      .then((res) => {
        if (res.status === 'ok') {
          this.props.onLogined();
          Setting.showMessage("success", `Logged in successfully`);
          Setting.goToLink("/");
        } else {
          Setting.showMessage("error", `Log in failed：${res.msg}`);
        }
      });
  };

  renderForm() {
    return (
      <Form
        name="normal_login"
        initialValues={{ remember: true }}
        onFinish={this.onFinish.bind(this)}
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
    const application = this.getApplicationObj();

    return (
      <Row>
        <Col span={24} style={{display: "flex", justifyContent:  "center"}} >
          <div style={{marginTop: "80px", textAlign: "center"}}>
            <img src={application?.logo} alt={application?.displayName} style={{marginBottom: '50px'}}/>
            {
              this.renderForm(application)
            }
          </div>
        </Col>
      </Row>
    )
  }
}

export default Face;
