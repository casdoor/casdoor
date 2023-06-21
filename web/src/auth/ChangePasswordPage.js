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
import {Button, Col, Form, Input, Row} from "antd";
import * as Setting from "../Setting";
import i18next from "i18next";
import * as ApplicationBackend from "../backend/ApplicationBackend";
import * as AuthBackend from "./AuthBackend";
import CustomGithubCorner from "../common/CustomGithubCorner";
import * as OrganizationBackend from "../backend/OrganizationBackend";
import {authConfig} from "./Auth";

class ChangePasswordPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      applicationName: props.application?.name ?? (props.match?.params?.applicationName ?? authConfig.appName) ?? null,
      owner: props.owner ?? (props.match?.params?.owner ?? null),
    };
  }

  componentDidMount() {
    if (this.getApplicationObj() === undefined) {
      if (this.state.applicationName !== undefined) {
        this.getApplication();
      }
    }
  }

  getApplication() {
    if (this.state.applicationName === undefined) {
      return;
    }

    if (this.state.owner === null || this.state.type === "saml") {
      ApplicationBackend.getApplication("admin", this.state.applicationName)
        .then((application) => {
          this.onUpdateApplication(application);
        });
    } else {
      OrganizationBackend.getDefaultApplication("admin", this.state.owner)
        .then((res) => {
          if (res.status === "ok") {
            const application = res.data;
            this.onUpdateApplication(application);
            this.setState({
              applicationName: res.data.name,
            });
          } else {
            this.onUpdateApplication(null);
            Setting.showMessage("error", res.msg);
          }
        });
    }
  }

  getApplicationObj() {
    return this.props.application;
  }

  onUpdateApplication(application) {
    this.props.onUpdateApplication(application);
  }

  onFinish(values) {
    AuthBackend.setNewPassword(values)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("changePassword:Password successfully changed"));
          const link = Setting.getFromLink();
          Setting.goToLink(link);
        } else {
          Setting.showMessage("error", i18next.t(`signup:${res.msg}`));
        }
      });
  }

  onFinishFailed(values, errorFields, outOfDate) {
    this.form.current.scrollToField(errorFields[0].name);
  }

  renderForm(application) {
    return (
      <Form
        labelCol={{span: 8}}
        wrapperCol={{span: 16}}
        ref={this.form}
        name="changePassword"
        onFinish={(values) => this.onFinish(values)}
        onFinishFailed={(errorInfo) => this.onFinishFailed(errorInfo.values, errorInfo.errorFields, errorInfo.outOfDate)}
        initialValues={{
          application: application.name,
          organization: application.organization,
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
          name="organization"
          hidden={true}
          rules={[
            {
              required: true,
              message: "Please input your organization!",
            },
          ]}
        >
        </Form.Item>
        <Form.Item
          name="currentPassword"
          label={i18next.t("changePassword:Current password")}
          rules={[
            {
              required: true,
              message: i18next.t("changePassword:Please input your old password"),
            },
          ]}
          hasFeedback
        >
          <Input.Password />
        </Form.Item>
        <Form.Item
          name="password"
          label={i18next.t("changePassword:New password")}
          rules={[
            {
              required: true,
              min: 6,
              message: i18next.t("changePassword:Please input your password, at least 6 characters!"),
            },
          ]}
          hasFeedback
        >
          <Input.Password />
        </Form.Item>
        <Form.Item
          name="confirm"
          label={i18next.t("changePassword:Re-enter new")}
          dependencies={["password"]}
          hasFeedback
          rules={[
            {
              required: true,
              message: i18next.t("changePassword:Please confirm your password!"),
            },
            ({getFieldValue}) => ({
              validator(rule, value) {
                if (!value || getFieldValue("password") === value) {
                  return Promise.resolve();
                }

                return Promise.reject(i18next.t("changePassword:Your confirmed password is inconsistent with the password!"));
              },
            }),
          ]}
        >
          <Input.Password />
        </Form.Item>

        <Form.Item
          wrapperCol={{span: 24}}>
          <Button type="primary" htmlType="submit">
            {i18next.t("changePassword:Change password")}
          </Button>
        </Form.Item>
      </Form>

    );
  }

  render() {
    const application = this.getApplicationObj();
    if (application === undefined || application === null) {
      return null;
    }

    return (
      <React.Fragment>
        <CustomGithubCorner />
        <div className="change-password-content" style={{boxShadow: Setting.isMobile() ? "none" : null}}>
          {Setting.inIframe() || Setting.isMobile() ? null : <div dangerouslySetInnerHTML={{__html: application.formCss}} />}
          {Setting.inIframe() || !Setting.isMobile() ? null : <div dangerouslySetInnerHTML={{__html: application.formCssMobile}} />}
          <div className="login-panel" >
            <div className="side-image" style={{display: application.formOffset !== 4 ? "none" : null}}>
              <div dangerouslySetInnerHTML={{__html: application.formSideHtml}} />
            </div>
            <div className="login-form">
              {
                Setting.renderHelmet(application)
              }
              {
                Setting.renderLogo(application)
              }
              <h1 style={{fontSize: "28px", fontWeight: "400", marginTop: "10px", marginBottom: "40px"}}>{i18next.t("changePassword:Change password")}</h1>
              <Row type="flex" justify="center" align="middle">
                <Col span={16} style={{width: 600}}>
                  {
                    this.renderForm(application)
                  }
                </Col>
              </Row>
            </div>
          </div>
        </div>
      </React.Fragment>
    );
  }
}

export default ChangePasswordPage;
