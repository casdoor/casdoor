// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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
import {Button, Card, List, Result, Space} from "antd";
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
      scopeDescriptions: [],
      granting: false,
      oAuthParams: Util.getOAuthGetParameters(),
    };
  }

  getApplicationObj() {
    return this.props.application;
  }

  componentDidMount() {
    this.getApplication();
    this.loadScopeDescriptions();
  }

  componentDidUpdate(prevProps) {
    if (this.props.application !== prevProps.application) {
      this.loadScopeDescriptions();
    }
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

        this.props.onUpdateApplication(res.data);
      });
  }

  loadScopeDescriptions() {
    const {oAuthParams} = this.state;
    const application = this.getApplicationObj();
    if (!oAuthParams?.scope || !application) {
      return;
    }
    // Check if urlPar scope is within application scopes
    const scopes = oAuthParams.scope.split(" ").map(s => s.trim()).filter(Boolean);
    const customScopes = application.customScopes || [];
    const customScopesMap = {};
    customScopes.forEach(s => {
      if (s?.scope) {
        customScopesMap[s.scope] = s;
      }
    });

    const scopeDescriptions = scopes
      .map(scope => {
        const item = customScopesMap[scope];
        if (item) {
          return {
            ...item,
            displayName: item.displayName || item.scope,
          };
        }
        return {
          scope: scope,
          displayName: scope,
          description: i18next.t("consent:This scope is not defined in the application"),
        };
      })
      .filter(Boolean);

    this.setState({
      scopeDescriptions: scopeDescriptions,
    });
  }

  handleGrant() {
    const {oAuthParams, scopeDescriptions} = this.state;
    const application = this.getApplicationObj();

    this.setState({granting: true});

    const consent = {
      owner: application.owner,
      application: application.owner + "/" + application.name,
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
    const application = this.getApplicationObj();

    if (application === undefined) {
      return null;
    }

    if (!application) {
      return (
        <Result
          status="error"
          title={i18next.t("general:Invalid application")}
        />
      );
    }

    const {scopeDescriptions, granting} = this.state;
    const isScopeEmpty = scopeDescriptions.length === 0;

    return (
      <div className="login-content">
        <div className={Setting.isDarkTheme(this.props.themeAlgorithm) ? "login-panel-dark" : "login-panel"}>
          <div className="login-form">
            <Card
              style={{
                padding: "32px",
                width: 450,
                borderRadius: "12px",
                boxShadow: "0 10px 25px rgba(0, 0, 0, 0.05)",
                border: "1px solid #f0f0f0",
              }}
            >
              <div style={{textAlign: "center", marginBottom: 24}}>
                {application.logo && (
                  <div style={{marginBottom: 16}}>
                    <img
                      src={application.logo}
                      alt={application.displayName || application.name}
                      style={{height: 56, objectFit: "contain"}}
                    />
                  </div>
                )}
                <h2 style={{margin: 0, fontWeight: 600, fontSize: "24px"}}>
                  {i18next.t("consent:Authorization Request")}
                </h2>
              </div>

              <div style={{marginBottom: 32}}>
                <p style={{fontSize: 15, color: "#666", textAlign: "center", lineHeight: "1.6"}}>
                  <span style={{fontWeight: 600, color: "#000"}}>{application.displayName || application.name}</span>
                  {" "}{i18next.t("consent:wants to access your account")}
                </p>
                {application.homepageUrl && (
                  <div style={{textAlign: "center", marginTop: 4}}>
                    <a href={application.homepageUrl} target="_blank" rel="noopener noreferrer" style={{fontSize: 13, color: "#1890ff"}}>
                      {application.homepageUrl}
                    </a>
                  </div>
                )}
              </div>

              <div style={{marginBottom: 32}}>
                <div style={{fontSize: 14, color: "#8c8c8c", marginBottom: 16}}>
                  <LockOutlined style={{marginRight: 8}} /> {i18next.t("consent:This application is requesting")}
                </div>
                <div style={{display: "flex", justifyContent: "center"}}>
                  <List
                    size="small"
                    dataSource={scopeDescriptions}
                    style={{width: "100%"}}
                    renderItem={item => (
                      <List.Item style={{borderBottom: "none", width: "100%"}}>
                        <div style={{display: "inline-grid", gridTemplateColumns: "16px auto", columnGap: 8, alignItems: "start"}}>
                          <CheckOutlined style={{color: "#52c41a", fontSize: "14px", marginTop: "4px", justifySelf: "center"}} />
                          <div style={{fontWeight: 500, fontSize: "14px", lineHeight: "22px"}}>{item.displayName || item.scope}</div>
                        </div>
                        <div style={{fontSize: "12px", color: "#8c8c8c", marginTop: 2}}>{item.description}</div>
                      </List.Item>
                    )}
                  />
                </div>
              </div>

              <div style={{textAlign: "center", marginBottom: 24}}>
                <Space size={16}>
                  <Button
                    type="primary"
                    size="large"
                    shape="round"
                    onClick={() => this.handleGrant()}
                    loading={granting}
                    disabled={granting || isScopeEmpty}
                    style={{minWidth: 120, height: 44, fontWeight: 500}}
                  >
                    {i18next.t("consent:Allow")}
                  </Button>
                  <Button
                    size="large"
                    shape="round"
                    onClick={() => this.handleDeny()}
                    disabled={granting || isScopeEmpty}
                    style={{minWidth: 120, height: 44, fontWeight: 500}}
                  >
                    {i18next.t("consent:Deny")}
                  </Button>
                </Space>
              </div>

              <div style={{padding: "16px", backgroundColor: "#fafafa", borderRadius: "8px", border: "1px solid #f0f0f0"}}>
                <p style={{margin: 0, fontSize: 12, color: "#8c8c8c", textAlign: "center", lineHeight: "1.5"}}>
                  {i18next.t("consent:By clicking Allow, you allow this app to use your information")}
                </p>
              </div>
            </Card>
          </div>
        </div>
      </div>
    );
  }
}

export default withRouter(ConsentPage);
