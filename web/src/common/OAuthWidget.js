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
import {Button, Col, Row} from "antd";
import i18next from "i18next";
import * as UserBackend from "../backend/UserBackend";
import * as Setting from "../Setting";
import * as Provider from "../auth/Provider";
import * as AuthBackend from "../auth/AuthBackend";

class OAuthWidget extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      addressOptions: [],
      affiliationOptions: [],
      oAuthState: "",
    };
  }

  componentWillMount() {
    this.getAddressOptions(this.props.application);
    this.getAffiliationOptions(this.props.application, this.props.user);
    this.getOAuthState();
  }

  getAddressOptions(application) {
    if (application.affiliationUrl === "") {
      return;
    }

    const addressUrl = application.affiliationUrl.split("|")[0];
    UserBackend.getAddressOptions(addressUrl)
      .then((addressOptions) => {
        this.setState({
          addressOptions: addressOptions,
        });
      });
  }

  getAffiliationOptions(application, user) {
    if (application.affiliationUrl === "") {
      return;
    }

    const affiliationUrl = application.affiliationUrl.split("|")[1];
    const code = user.address[user.address.length - 1];
    UserBackend.getAffiliationOptions(affiliationUrl, code)
      .then((affiliationOptions) => {
        this.setState({
          affiliationOptions: affiliationOptions,
        });
      });
  }

  getOAuthState() {
    AuthBackend.getOAuthState().then((res) => {
      this.setState({
        oAuthState: res.state,
      });
    });
}

  updateUserField(key, value) {
    this.props.onUpdateUserField(key, value);
  }

  unlinked() {
    this.props.onUnlinked();
  }

  getProviderLink(user, provider) {
    if (provider.type === "GitHub") {
      return `https://github.com/${this.getUserProperty(user, provider.type, "username")}`;
    } else if (provider.type === "Google") {
      return "https://mail.google.com";
    } else {
      return "";
    }
  }

  getUserProperty(user, providerType, propertyName) {
    const key = `oauth_${providerType}_${propertyName}`;
    if (user.properties === null) {return "";}
    return user.properties[key];
  }

  unlinkUser(providerType) {
    const body = {
      providerType: providerType,
    };
    AuthBackend.unlink(body)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", "Unlinked successfully");

          this.unlinked();
        } else {
          Setting.showMessage("error", `Failed to unlink: ${res.msg}`);
        }
      });
  }

  renderIdp(user, application, providerItem) {
    const provider = providerItem.provider;
    const linkedValue = user[provider.type.toLowerCase()];
    const profileUrl = this.getProviderLink(user, provider);
    const id = this.getUserProperty(user, provider.type, "id");
    const username = this.getUserProperty(user, provider.type, "username");
    const displayName = this.getUserProperty(user, provider.type, "displayName");
    const email = this.getUserProperty(user, provider.type, "email");
    let avatarUrl = this.getUserProperty(user, provider.type, "avatarUrl");

    if (avatarUrl === "" || avatarUrl === undefined) {
      avatarUrl = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAB4AAAAeCAQAAACROWYpAAAAHElEQVR42mNkoAAwjmoe1TyqeVTzqOZRzcNZMwB18wAfEFQkPQAAAABJRU5ErkJggg==";
    }

    let name = (username === undefined) ? displayName : `${displayName} (${username})`;
    if (name === undefined) {
      if (id !== undefined) {
        name = id;
      } else if (email !== undefined) {
        name = email;
      } else {
        name = linkedValue;
      }
    }

    return (
      <Row key={provider.name} style={{marginTop: "20px"}} >
        <Col style={{marginTop: "5px"}} span={this.props.labelSpan}>
          {
            Setting.getProviderLogo(provider)
          }
          <span style={{marginLeft: "5px"}}>
            {
              `${provider.type}:`
            }
          </span>
        </Col>
        <Col span={24 - this.props.labelSpan} >
          <img style={{marginRight: "10px"}} width={30} height={30} src={avatarUrl} alt={name} referrerPolicy="no-referrer" />
          <span style={{width: this.props.labelSpan === 3 ? "300px" : "130px", display: (Setting.isMobile()) ? "inline" : "inline-block"}}>
            {
              linkedValue === "" ? (
                "(empty)"
              ) : (
                profileUrl === "" ? name : (
                  <a target="_blank" rel="noreferrer" href={profileUrl}>
                    {
                      name
                    }
                  </a>
                )
              )
            }
          </span>
          {
            linkedValue === "" ? (
              <a key={provider.displayName} href={Provider.getAuthUrl(application, provider, "link", this.state.oAuthState)}>
                <Button style={{marginLeft: "20px", width: "80px"}} type="primary">{i18next.t("user:Link")}</Button>
              </a>
            ) : (
              <Button disabled={!providerItem.canUnlink} style={{marginLeft: "20px", width: "80px"}} onClick={() => this.unlinkUser(provider.type)}>{i18next.t("user:Unlink")}</Button>
            )
          }
        </Col>
      </Row>
    );
  }

  render() {
    return this.renderIdp(this.props.user, this.props.application, this.props.providerItem);
  }
}

export default OAuthWidget;
