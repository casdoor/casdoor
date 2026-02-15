// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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
import {Button, Card, Col, List, Result, Row, Spin} from "antd";
import {CheckOutlined, LockOutlined} from "@ant-design/icons";
import * as ApplicationBackend from "../backend/ApplicationBackend";
import * as ConsentBackend from "../backend/ConsentBackend";
import * as Setting from "../Setting";
import i18next from "i18next";
import {withRouter} from "react-router-dom";
import * as Util from "./Util";

class ConsentPage extends React.Component {
  constructor(props) {
    super(props);
    const params = new URLSearchParams(window.location.search);
    this.state = {
      applicationName: props.match?.params?.applicationName || params.get("application"),
      application: null,
      scopeDescriptions: [],
      loading: true,
      granting: false,
      oAuthParams: Util.getOAuthGetParameters(),
    };
  }

  componentDidMount() {
    this.getApplication();
    this.loadScopeDescriptions();
  }

  getApplication() {
    if (!this.state.applicationName) {
      return;
    }

    ApplicationBackend.getApplication("admin", this.state.applicationName)
      .then((res) => {
        if (res.status === "error") {
          Setting.showMessage("error", res.msg);
          return;
        }

        this.setState({
          application: res.data,
        });
      });
  }

  loadScopeDescriptions() {
    const {oAuthParams} = this.state;
    if (!oAuthParams?.scope) {
      this.setState({loading: false});
      return;
    }

    const scopes = oAuthParams.scope.split(" ");
    const scopeDescriptions = scopes.map(scope => {
      // Use default descriptions or create generic ones
      const defaultDescriptions = {
        "openid": {scope: "openid", displayName: "OpenID", description: "Verify your identity"},
        "profile": {scope: "profile", displayName: "Profile", description: "View your basic profile information"},
        "email": {scope: "email", displayName: "Email", description: "View your email address"},
        "address": {scope: "address", displayName: "Address", description: "View your address"},
        "phone": {scope: "phone", displayName: "Phone", description: "View your phone number"},
        "offline_access": {scope: "offline_access", displayName: "Offline Access", description: "Maintain access when you are not actively using the application"},
      };

      return defaultDescriptions[scope] || {
        scope: scope,
        displayName: scope,
        description: `Access to ${scope}`,
      };
    });

    this.setState({
      scopeDescriptions: scopeDescriptions,
      loading: false,
    });
  }

  handleGrant() {
    const {oAuthParams, application, scopeDescriptions} = this.state;
    this.setState({granting: true});

    const consent = {
      owner: application.owner,
      application: application.name,
      grantedScopes: scopeDescriptions.map(s => s.scope),
    };

    ConsentBackend.grantConsent(consent, oAuthParams)
      .then((res) => {
        if (res.status === "ok") {
          // res.data contains the authorization code
          const code = res.data;
          const concatChar = oAuthParams?.redirectUri?.includes("?") ? "&" : "?";
          const redirectUrl = `${oAuthParams.redirectUri}${concatChar}code=${code}&state=${oAuthParams.state}`;
          Setting.goToLink(redirectUrl);
        } else {
          Setting.showMessage("error", res.msg);
          this.setState({granting: false});
        }
      });
  }

  handleDeny() {
    const {oAuthParams} = this.state;
    const concatChar = oAuthParams?.redirectUri?.includes("?") ? "&" : "?";
    Setting.goToLink(`${oAuthParams.redirectUri}${concatChar}error=access_denied&error_description=User denied consent&state=${oAuthParams.state}`);
  }

  render() {
    const {application, scopeDescriptions, loading, granting} = this.state;

    if (loading) {
      return (
        <div style={{display: "flex", justifyContent: "center", alignItems: "center", minHeight: "100vh"}}>
          <Spin size="large" tip={i18next.t("login:Loading")} />
        </div>
      );
    }

    if (!application) {
      return (
        <Result
          status="error"
          title={i18next.t("general:Invalid application")}
        />
      );
    }

    return (
      <div style={{
        display: "flex",
        justifyContent: "center",
        alignItems: "center",
        minHeight: "100vh",
        backgroundColor: "#f0f2f5",
      }}>
        <Row>
          <Col span={24}>
            <Card
              style={{
                width: 500,
                boxShadow: "0 4px 8px rgba(0, 0, 0, 0.1)",
              }}
            >
              <div style={{textAlign: "center", marginBottom: 24}}>
                {application.logo && (
                  <img
                    src={application.logo}
                    alt={application.displayName || application.name}
                    style={{width: 80, height: 80, marginBottom: 16}}
                  />
                )}
                <h2 style={{margin: 0}}>
                  {i18next.t("consent:Authorization Request")}
                </h2>
              </div>

              <div style={{marginBottom: 24}}>
                <p style={{fontSize: 16, textAlign: "center"}}>
                  <strong>{application.displayName || application.name}</strong>
                  {" "}{i18next.t("consent:wants to access your account")}
                </p>
                {application.homepageUrl && (
                  <p style={{textAlign: "center", color: "#666"}}>
                    <a href={application.homepageUrl} target="_blank" rel="noopener noreferrer">
                      {application.homepageUrl}
                    </a>
                  </p>
                )}
              </div>

              <div style={{marginBottom: 24}}>
                <h3 style={{marginBottom: 16}}>
                  <LockOutlined /> {i18next.t("consent:This application is requesting")}:
                </h3>
                <List
                  size="small"
                  dataSource={scopeDescriptions}
                  renderItem={item => (
                    <List.Item>
                      <List.Item.Meta
                        avatar={<CheckOutlined style={{color: "#52c41a"}} />}
                        title={item.displayName || item.scope}
                        description={item.description}
                      />
                    </List.Item>
                  )}
                />
              </div>

              <div style={{textAlign: "center"}}>
                <Button
                  type="primary"
                  size="large"
                  onClick={() => this.handleGrant()}
                  loading={granting}
                  style={{marginRight: 16, minWidth: 100}}
                >
                  {i18next.t("consent:Allow")}
                </Button>
                <Button
                  size="large"
                  onClick={() => this.handleDeny()}
                  disabled={granting}
                  style={{minWidth: 100}}
                >
                  {i18next.t("consent:Deny")}
                </Button>
              </div>

              <div style={{marginTop: 24, padding: "12px 16px", backgroundColor: "#f5f5f5", borderRadius: 4}}>
                <p style={{margin: 0, fontSize: 12, color: "#666"}}>
                  {i18next.t("consent:By clicking Allow, you allow this app to use your information")}
                </p>
              </div>
            </Card>
          </Col>
        </Row>
      </div>
    );
  }
}

export default withRouter(ConsentPage);
