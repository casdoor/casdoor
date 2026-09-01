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
import {CaptchaModal} from "./common/modal/CaptchaModal";
import * as ApplicationBackend from "./backend/ApplicationBackend";
import * as Setting from "./Setting";

class CaptchaPage extends React.Component {
  constructor(props) {
    super(props);
    const params = new URLSearchParams(this.props.location.search);
    this.state = {
      owner: "admin",
      application: null,
      clientId: params.get("client_id"),
      applicationName: params.get("state"),
      redirectUri: params.get("redirect_uri"),
    };
  }

  componentDidMount() {
    this.getApplication();
  }

  onUpdateApplication(application) {
    this.setState({
      application: application,
    });
  }

  getApplication() {
    if (this.state.applicationName === null) {
      return null;
    }

    ApplicationBackend.getApplication(this.state.owner, this.state.applicationName)
      .then((res) => {
        if (res.status === "error") {
          this.onUpdateApplication(null);
          this.setState({
            msg: res.msg,
          });
          return ;
        }
        this.onUpdateApplication(res.data);
      });
  }

  getCaptchaProviderItems(application) {
    const providers = application?.providers;

    if (providers === undefined || providers === null) {
      return null;
    }

    return providers.filter(providerItem => {
      if (providerItem.provider === undefined || providerItem.provider === null) {
        return false;
      }

      return providerItem.provider.category === "Captcha";
    });
  }

  callback(values) {
    Setting.goToLink(`${this.state.redirectUri}?code=${values.captchaToken}&type=${values.captchaType}&secret=${values.clientSecret}&applicationId=${values.applicationId}`);
  }

  renderCaptchaModal(application) {
    const captchaProviderItems = this.getCaptchaProviderItems(application);
    if (captchaProviderItems === null) {
      return null;
    }
    const alwaysProviderItems = captchaProviderItems.filter(providerItem => providerItem.rule === "Always");
    const dynamicProviderItems = captchaProviderItems.filter(providerItem => providerItem.rule === "Dynamic");
    const provider = alwaysProviderItems.length > 0
      ? alwaysProviderItems[0].provider
      : dynamicProviderItems[0].provider;

    return <CaptchaModal
      owner={provider.owner}
      name={provider.name}
      visible={true}
      onOk={(captchaType, captchaToken, clientSecret) => {
        const values = {
          captchaType: captchaType,
          captchaToken: captchaToken,
          clientSecret: clientSecret,
          applicationId: `${provider.owner}/${provider.name}`,
        };
        this.callback(values);
      }}
      onCancel={() => this.callback({captchaType: "none", captchaToken: "", clientSecret: ""})}
      isCurrentProvider={true}
    />;
  }
  render() {
    return (
      this.renderCaptchaModal(this.state.application)
    );
  }
}

export default CaptchaPage;
