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
import {Button, Card, Col, Input, Row, Select, Switch} from "antd";
import {LinkOutlined} from "@ant-design/icons";
import * as ProviderBackend from "./backend/ProviderBackend";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as CertBackend from "./backend/CertBackend";
import * as Setting from "./Setting";
import i18next from "i18next";
import {renderNotificationProviderFields} from "./provider/NotificationProviderFields";
import {renderEmailProviderFields} from "./provider/EmailProviderFields";
import {renderSmsProviderFields} from "./provider/SmsProviderFields";
import {renderMfaProviderFields} from "./provider/MfaProviderFields";
import {renderSamlProviderFields} from "./provider/SamlProviderFields";
import {renderOAuthProviderFields} from "./provider/OAuthProviderFields";
import {renderCaptchaProviderFields} from "./provider/CaptchaProviderFields";
import {renderPaymentProviderFields} from "./provider/PaymentProviderFields";
import {renderWeb3ProviderFields} from "./provider/Web3ProviderFields";
import {renderStorageProviderFields} from "./provider/StorageProviderFields";
import {renderFaceIdProviderFields} from "./provider/FaceIDProviderFields";
import {renderIDVerificationProviderFields} from "./provider/IDVerificationProviderFields";

const {Option} = Select;
const {TextArea} = Input;

const defaultUserMapping = {
  id: "id",
  username: "username",
  displayName: "displayName",
  email: "email",
  avatarUrl: "avatarUrl",
  phone: "phone",
  countryCode: "country_code",
  firstName: "given_name",
  lastName: "family_name",
  region: "region",
  location: "location",
  affiliation: "affiliation",
  title: "title",
};

const defaultEmailMapping = {
  fromName: "fromName",
  fromAddress: "fromAddress",
  toAddress: "toAddress",
  subject: "subject",
  content: "content",
};

const defaultSmsMapping = {
  phoneNumber: "phoneNumber",
  content: "content",
};

class ProviderEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      providerName: props.match.params.providerName,
      owner: props.organizationName !== undefined ? props.organizationName : props.match.params.organizationName,
      provider: null,
      certs: [],
      organizations: [],
      mode: props.location.mode !== undefined ? props.location.mode : "edit",
    };
  }

  UNSAFE_componentWillMount() {
    this.getOrganizations();
    this.getProvider();
    this.getCerts(this.state.owner);
  }

  getProvider() {
    ProviderBackend.getProvider(this.state.owner, this.state.providerName)
      .then((res) => {
        if (res.data === null) {
          this.props.history.push("/404");
          return;
        }

        if (res.status === "ok") {
          const provider = res.data;
          if (provider.type === "Custom HTTP Email") {
            if (!provider.userMapping) {
              provider.userMapping = provider.userMapping || defaultEmailMapping;
            }
            if (!provider.userMapping?.fromName) {
              provider.userMapping = defaultEmailMapping;
            }
          } else if (provider.type === "Custom HTTP SMS") {
            if (!provider.userMapping) {
              provider.userMapping = provider.userMapping || defaultSmsMapping;
            }
            if (!provider.userMapping?.phoneNumber) {
              provider.userMapping = defaultSmsMapping;
            }
          } else {
            provider.userMapping = provider.userMapping || defaultUserMapping;
          }
          this.setState({
            provider: provider,
          });
        } else {
          Setting.showMessage("error", res.msg);
        }
      });
  }

  getOrganizations() {
    if (Setting.isAdminUser(this.props.account)) {
      OrganizationBackend.getOrganizations("admin")
        .then((res) => {
          this.setState({
            organizations: res.data || [],
          });
        });
    }
  }

  getCerts(owner) {
    CertBackend.getCerts(owner)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            certs: res.data || [],
          });
        }
      });
  }

  parseProviderField(key, value) {
    if (["port"].includes(key)) {
      value = Setting.myParseInt(value);
    }
    return value;
  }

  updateProviderField(key, value) {
    value = this.parseProviderField(key, value);

    const provider = this.state.provider;
    if (key === "owner" && provider["owner"] !== value) {
      // the provider change the owner, reset the cert
      provider["cert"] = "";
      this.getCerts(value);
    }

    provider[key] = value;

    if (provider["type"] === "WeChat") {
      if (!provider["clientId"]) {
        provider["signName"] = "media";
        provider["disableSsl"] = true;
      }
      if (!provider["clientId2"]) {
        provider["signName"] = "open";
        provider["disableSsl"] = false;
      }
      if (!provider["disableSsl"]) {
        provider["signName"] = "open";
      }
    }

    this.setState({
      provider: provider,
    });
  }

  updateUserMappingField(key, value) {
    const requiredKeys = ["id", "username", "displayName"];
    const provider = this.state.provider;

    if (provider.type === "Custom HTTP Email") {
      if (value === "") {
        Setting.showMessage("error", i18next.t("provider:This field is required"));
        return;
      }
    } else {
      if (value === "" && requiredKeys.includes(key)) {
        Setting.showMessage("error", i18next.t("provider:This field is required"));
        return;
      }
    }

    if (value === "") {
      delete provider.userMapping[key];
    } else {
      provider.userMapping[key] = value;
    }

    this.setState({
      provider: provider,
    });
  }

  renderUserMappingInput() {
    return (
      <React.Fragment>
        {Setting.getLabel(i18next.t("general:ID"), i18next.t("general:ID - Tooltip"))} :
        <Input value={this.state.provider.userMapping.id} onChange={e => {
          this.updateUserMappingField("id", e.target.value);
        }} />
        {Setting.getLabel(i18next.t("signup:Username"), i18next.t("signup:Username - Tooltip"))} :
        <Input value={this.state.provider.userMapping.username} onChange={e => {
          this.updateUserMappingField("username", e.target.value);
        }} />
        {Setting.getLabel(i18next.t("general:Display name"), i18next.t("general:Display name - Tooltip"))} :
        <Input value={this.state.provider.userMapping.displayName} onChange={e => {
          this.updateUserMappingField("displayName", e.target.value);
        }} />
        {Setting.getLabel(i18next.t("general:Email"), i18next.t("general:Email - Tooltip"))} :
        <Input value={this.state.provider.userMapping.email} onChange={e => {
          this.updateUserMappingField("email", e.target.value);
        }} />
        {Setting.getLabel(i18next.t("general:Avatar"), i18next.t("general:Avatar - Tooltip"))} :
        <Input value={this.state.provider.userMapping.avatarUrl} onChange={e => {
          this.updateUserMappingField("avatarUrl", e.target.value);
        }} />
        {Setting.getLabel(i18next.t("general:Phone"), i18next.t("general:Phone - Tooltip"))} :
        <Input value={this.state.provider.userMapping.phone} onChange={e => {
          this.updateUserMappingField("phone", e.target.value);
        }} />
        {Setting.getLabel(i18next.t("user:Country code"), i18next.t("user:Country code - Tooltip"))} :
        <Input value={this.state.provider.userMapping.countryCode} onChange={e => {
          this.updateUserMappingField("countryCode", e.target.value);
        }} />
        {Setting.getLabel(i18next.t("general:First name"), i18next.t("general:First name - Tooltip"))} :
        <Input value={this.state.provider.userMapping.firstName} onChange={e => {
          this.updateUserMappingField("firstName", e.target.value);
        }} />
        {Setting.getLabel(i18next.t("general:Last name"), i18next.t("general:Last name - Tooltip"))} :
        <Input value={this.state.provider.userMapping.lastName} onChange={e => {
          this.updateUserMappingField("lastName", e.target.value);
        }} />
        {Setting.getLabel(i18next.t("provider:Region"), i18next.t("provider:Region - Tooltip"))} :
        <Input value={this.state.provider.userMapping.region} onChange={e => {
          this.updateUserMappingField("region", e.target.value);
        }} />
        {Setting.getLabel(i18next.t("user:Location"), i18next.t("user:Location - Tooltip"))} :
        <Input value={this.state.provider.userMapping.location} onChange={e => {
          this.updateUserMappingField("location", e.target.value);
        }} />
        {Setting.getLabel(i18next.t("user:Affiliation"), i18next.t("user:Affiliation - Tooltip"))} :
        <Input value={this.state.provider.userMapping.affiliation} onChange={e => {
          this.updateUserMappingField("affiliation", e.target.value);
        }} />
        {Setting.getLabel(i18next.t("general:Title"), i18next.t("general:Title - Tooltip"))} :
        <Input value={this.state.provider.userMapping.title} onChange={e => {
          this.updateUserMappingField("title", e.target.value);
        }} />
      </React.Fragment>
    );
  }

  renderEmailMappingInput() {
    return (
      <React.Fragment>
        {Setting.getLabel(i18next.t("provider:From name"), i18next.t("provider:From name - Tooltip"))} :
        <Input value={this.state.provider.userMapping.fromName} onChange={e => {
          this.updateUserMappingField("fromName", e.target.value);
        }} />
        {Setting.getLabel(i18next.t("provider:From address"), i18next.t("provider:From address - Tooltip"))} :
        <Input value={this.state.provider.userMapping.fromAddress} onChange={e => {
          this.updateUserMappingField("fromAddress", e.target.value);
        }} />
        {Setting.getLabel(i18next.t("provider:To address"), i18next.t("provider:To address - Tooltip"))} :
        <Input value={this.state.provider.userMapping.toAddress} onChange={e => {
          this.updateUserMappingField("toAddress", e.target.value);
        }} />
        {Setting.getLabel(i18next.t("provider:Subject"), i18next.t("provider:Subject - Tooltip"))} :
        <Input value={this.state.provider.userMapping.subject} onChange={e => {
          this.updateUserMappingField("subject", e.target.value);
        }} />
        {Setting.getLabel(i18next.t("provider:Email content"), i18next.t("provider:Email content - Tooltip"))} :
        <Input value={this.state.provider.userMapping.content} onChange={e => {
          this.updateUserMappingField("content", e.target.value);
        }} />
      </React.Fragment>
    );
  }

  renderSmsMappingInput() {
    return (
      <React.Fragment>
        {Setting.getLabel(i18next.t("general:Phone"), i18next.t("general:Phone - Tooltip"))} :
        <Input value={this.state.provider.userMapping.phoneNumber} onChange={e => {
          this.updateUserMappingField("phoneNumber", e.target.value);
        }} />
        {Setting.getLabel(i18next.t("provider:Content"), i18next.t("provider:Content - Tooltip"))} :
        <Input value={this.state.provider.userMapping.content} onChange={e => {
          this.updateUserMappingField("content", e.target.value);
        }} />
      </React.Fragment>
    );
  }

  getClientIdLabel(provider) {
    switch (provider.category) {
    case "OAuth":
      if (provider.type === "Apple") {
        return Setting.getLabel(i18next.t("provider:Service ID identifier"), i18next.t("provider:Service ID identifier - Tooltip"));
      } else {
        return Setting.getLabel(i18next.t("provider:Client ID"), i18next.t("provider:Client ID - Tooltip"));
      }
    case "Email":
      return Setting.getLabel(i18next.t("signup:Username"), i18next.t("signup:Username - Tooltip"));
    case "SMS":
      if (provider.type === "Volc Engine SMS" || provider.type === "Amazon SNS" || provider.type === "Baidu Cloud SMS") {
        return Setting.getLabel(i18next.t("general:Access key"), i18next.t("general:Access key - Tooltip"));
      } else if (provider.type === "Huawei Cloud SMS") {
        return Setting.getLabel(i18next.t("provider:App key"), i18next.t("provider:App key - Tooltip"));
      } else if (provider.type === "UCloud SMS") {
        return Setting.getLabel(i18next.t("provider:Public key"), i18next.t("provider:Public key - Tooltip"));
      } else if (provider.type === "Msg91 SMS" || provider.type === "Infobip SMS" || provider.type === "OSON SMS") {
        return Setting.getLabel(i18next.t("provider:Sender Id"), i18next.t("provider:Sender Id - Tooltip"));
      } else {
        return Setting.getLabel(i18next.t("provider:Client ID"), i18next.t("provider:Client ID - Tooltip"));
      }
    case "Captcha":
      if (provider.type === "Aliyun Captcha") {
        return Setting.getLabel(i18next.t("general:Access key"), i18next.t("general:Access key - Tooltip"));
      } else {
        return Setting.getLabel(i18next.t("provider:Site key"), i18next.t("provider:Site key - Tooltip"));
      }
    case "Notification":
      if (provider.type === "DingTalk") {
        return Setting.getLabel(i18next.t("general:Access key"), i18next.t("general:Access key - Tooltip"));
      } else {
        return Setting.getLabel(i18next.t("provider:Client ID"), i18next.t("provider:Client ID - Tooltip"));
      }
    case "ID Verification":
      if (provider.type === "Alibaba Cloud") {
        return Setting.getLabel(i18next.t("general:Access key"), i18next.t("general:Access key - Tooltip"));
      } else {
        return Setting.getLabel(i18next.t("provider:Client ID"), i18next.t("provider:Client ID - Tooltip"));
      }
    default:
      return Setting.getLabel(i18next.t("provider:Client ID"), i18next.t("provider:Client ID - Tooltip"));
    }
  }

  getClientSecretLabel(provider) {
    switch (provider.category) {
    case "OAuth":
      if (provider.type === "Apple") {
        return Setting.getLabel(i18next.t("provider:Team ID"), i18next.t("provider:Team ID - Tooltip"));
      } else {
        return Setting.getLabel(i18next.t("provider:Client secret"), i18next.t("provider:Client secret - Tooltip"));
      }
    case "Storage":
      if (provider.type === "Google Cloud Storage") {
        return Setting.getLabel(i18next.t("provider:Service account JSON"), i18next.t("provider:Service account JSON - Tooltip"));
      } else {
        return Setting.getLabel(i18next.t("provider:Client secret"), i18next.t("provider:Client secret - Tooltip"));
      }
    case "Email":
      if (provider.type === "Azure ACS" || provider.type === "SendGrid") {
        return Setting.getLabel(i18next.t("provider:Secret key"), i18next.t("provider:Secret key - Tooltip"));
      } else {
        return Setting.getLabel(i18next.t("general:Password"), i18next.t("general:Password - Tooltip"));
      }
    case "SMS":
      if (provider.type === "Volc Engine SMS" || provider.type === "Amazon SNS" || provider.type === "Baidu Cloud SMS" || provider.type === "OSON SMS") {
        return Setting.getLabel(i18next.t("provider:Secret access key"), i18next.t("provider:Secret access key - Tooltip"));
      } else if (provider.type === "Huawei Cloud SMS") {
        return Setting.getLabel(i18next.t("provider:App secret"), i18next.t("provider:AppSecret - Tooltip"));
      } else if (provider.type === "UCloud SMS") {
        return Setting.getLabel(i18next.t("provider:Private Key"), i18next.t("provider:Private Key - Tooltip"));
      } else if (provider.type === "Msg91 SMS") {
        return Setting.getLabel(i18next.t("provider:Auth Key"), i18next.t("provider:Auth Key - Tooltip"));
      } else if (provider.type === "Infobip SMS") {
        return Setting.getLabel(i18next.t("provider:Api Key"), i18next.t("provider:Api Key - Tooltip"));
      } else {
        return Setting.getLabel(i18next.t("provider:Client secret"), i18next.t("provider:Client secret - Tooltip"));
      }
    case "Captcha":
      if (provider.type === "Aliyun Captcha") {
        return Setting.getLabel(i18next.t("provider:Secret access key"), i18next.t("provider:Secret access key - Tooltip"));
      } else {
        return Setting.getLabel(i18next.t("provider:Secret key"), i18next.t("provider:Secret key - Tooltip"));
      }
    case "Notification":
      if (provider.type === "Line" || provider.type === "Telegram" || provider.type === "Bark" || provider.type === "DingTalk" || provider.type === "Discord" || provider.type === "Slack" || provider.type === "Pushover" || provider.type === "Pushbullet") {
        return Setting.getLabel(i18next.t("provider:Secret key"), i18next.t("provider:Secret key - Tooltip"));
      } else if (provider.type === "Lark" || provider.type === "Microsoft Teams" || provider.type === "WeCom") {
        return Setting.getLabel(i18next.t("provider:Endpoint"), i18next.t("provider:Endpoint - Tooltip"));
      } else {
        return Setting.getLabel(i18next.t("provider:Client secret"), i18next.t("provider:Client secret - Tooltip"));
      }
    case "ID Verification":
      if (provider.type === "Alibaba Cloud") {
        return Setting.getLabel(i18next.t("provider:Secret access key"), i18next.t("provider:Secret access key - Tooltip"));
      } else {
        return Setting.getLabel(i18next.t("provider:Client secret"), i18next.t("provider:Client secret - Tooltip"));
      }
    default:
      return Setting.getLabel(i18next.t("provider:Client secret"), i18next.t("provider:Client secret - Tooltip"));
    }
  }

  getClientId2Label(provider) {
    switch (provider.category) {
    case "OAuth":
      if (provider.type === "Apple") {
        return Setting.getLabel(i18next.t("provider:Key ID"), i18next.t("provider:Key ID - Tooltip"));
      } else {
        return Setting.getLabel(i18next.t("provider:Client ID 2"), i18next.t("provider:Client ID 2 - Tooltip"));
      }
    case "Email":
      return Setting.getLabel(i18next.t("provider:From address"), i18next.t("provider:From address - Tooltip"));
    default:
      if (provider.type === "Aliyun Captcha") {
        return Setting.getLabel(i18next.t("provider:Scene"), i18next.t("provider:Scene - Tooltip"));
      } else if (provider.type === "WeChat Pay" || provider.type === "CUCloud") {
        return Setting.getLabel(i18next.t("provider:App ID"), i18next.t("provider:App ID - Tooltip"));
      } else {
        return Setting.getLabel(i18next.t("provider:Client ID 2"), i18next.t("provider:Client ID 2 - Tooltip"));
      }
    }
  }

  getClientSecret2Label(provider) {
    switch (provider.category) {
    case "OAuth":
      if (provider.type === "Apple") {
        return Setting.getLabel(i18next.t("provider:Key text"), i18next.t("provider:Key text - Tooltip"));
      } else {
        return Setting.getLabel(i18next.t("provider:Client secret 2"), i18next.t("provider:Client secret 2 - Tooltip"));
      }
    case "Email":
      return Setting.getLabel(i18next.t("provider:From name"), i18next.t("provider:From name - Tooltip"));
    default:
      if (provider.type === "Aliyun Captcha") {
        return Setting.getLabel(i18next.t("provider:App key"), i18next.t("provider:App key - Tooltip"));
      } else {
        return Setting.getLabel(i18next.t("provider:Client secret 2"), i18next.t("provider:Client secret 2 - Tooltip"));
      }
    }
  }

  getProviderSubTypeOptions(type) {
    if (type === "WeCom" || type === "Infoflow") {
      return (
        [
          {id: "Internal", name: i18next.t("provider:Internal")},
          {id: "Third-party", name: i18next.t("provider:Third-party")},
        ]
      );
    } else if (type === "WeChat") {
      return (
        [
          {id: "Web", name: i18next.t("provider:Web")},
          {id: "Mobile", name: i18next.t("provider:Mobile")},
        ]
      );
    } else {
      return [];
    }
  }

  getAppIdRow(provider) {
    let text = "";
    let tooltip = "";

    if (provider.category === "OAuth") {
      if (provider.type === "WeCom" && provider.subType === "Internal") {
        text = i18next.t("provider:Agent ID");
        tooltip = i18next.t("provider:Agent ID - Tooltip");
      } else if (provider.type === "Infoflow") {
        text = i18next.t("provider:Agent ID");
        tooltip = i18next.t("provider:Agent ID - Tooltip");
      } else if (provider.type === "AzureADB2C") {
        text = i18next.t("provider:User flow");
        tooltip = i18next.t("provider:User flow - Tooltip");
      }
    } else if (provider.category === "SMS") {
      if (provider.type === "Twilio SMS" || provider.type === "Azure ACS") {
        text = i18next.t("provider:Sender number");
        tooltip = i18next.t("provider:Sender number - Tooltip");
      } else if (provider.type === "Tencent Cloud SMS") {
        text = i18next.t("provider:App ID");
        tooltip = i18next.t("provider:App ID - Tooltip");
      } else if (provider.type === "Volc Engine SMS") {
        text = i18next.t("provider:SMS account");
        tooltip = i18next.t("provider:SMS account - Tooltip");
      } else if (provider.type === "Huawei Cloud SMS") {
        text = i18next.t("provider:Channel No.");
        tooltip = i18next.t("provider:Channel No. - Tooltip");
      } else if (provider.type === "Amazon SNS") {
        text = i18next.t("provider:Region");
        tooltip = i18next.t("provider:Region - Tooltip");
      } else if (provider.type === "Baidu Cloud SMS") {
        text = i18next.t("provider:Endpoint");
        tooltip = i18next.t("provider:Endpoint - Tooltip");
      } else if (provider.type === "Infobip SMS") {
        text = i18next.t("provider:Base URL");
        tooltip = i18next.t("provider:Base URL - Tooltip");
      } else if (provider.type === "UCloud SMS") {
        text = i18next.t("provider:Project Id");
        tooltip = i18next.t("provider:Project Id - Tooltip");
      }
    } else if (provider.category === "Email") {
      if (provider.type === "SUBMAIL") {
        text = i18next.t("provider:App ID");
        tooltip = i18next.t("provider:App ID - Tooltip");
      }
    } else if (provider.category === "Notification") {
      if (provider.type === "Viber") {
        text = i18next.t("provider:Domain");
        tooltip = i18next.t("provider:Domain - Tooltip");
      } else if (provider.type === "Line" || provider.type === "Matrix" || provider.type === "Rocket Chat") {
        text = i18next.t("provider:App Key");
        tooltip = i18next.t("provider:App Key - Tooltip");
      } else if (provider.type === "CUCloud") {
        text = "Topic name";
        tooltip = "Topic name - Tooltip";
      }
    }

    if (text === "" && tooltip === "") {
      return null;
    } else {
      return (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(text, tooltip)} :
          </Col>
          <Col span={22} >
            <Input value={provider.appId} onChange={e => {
              this.updateProviderField("appId", e.target.value);
            }} />
          </Col>
        </Row>
      );
    }
  }

  getReceiverRow(provider) {
    let text = "";
    let tooltip = "";

    if (provider.type === "Telegram" || provider.type === "Pushover" || provider.type === "Pushbullet" || provider.type === "Slack" || provider.type === "Discord" || provider.type === "Line" || provider.type === "Twitter" || provider.type === "Reddit" || provider.type === "Rocket Chat" || provider.type === "Viber") {
      text = i18next.t("provider:Chat ID");
      tooltip = i18next.t("provider:Chat ID - Tooltip");
    } else if (provider.type === "Custom HTTP" || provider.type === "Webpush" || provider.type === "Matrix") {
      text = i18next.t("provider:Endpoint");
      tooltip = i18next.t("provider:Endpoint - Tooltip");
    }

    if (text === "" && tooltip === "") {
      return (
        <React.Fragment>
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel("Test Notification", "Test Notification")} :
          </Col>
        </React.Fragment>
      );
    } else {
      return (
        <React.Fragment>
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(text, tooltip)} :
          </Col>
          <Col span={6} >
            <Input value={provider.receiver} onChange={e => {
              this.updateProviderField("receiver", e.target.value);
            }} />
          </Col>
        </React.Fragment>
      );
    }
  }

  loadSamlConfiguration() {
    const parser = new DOMParser();
    const rawXml = this.state.provider.metadata.replace("\n", "");
    const xmlDoc = parser.parseFromString(rawXml, "text/xml");
    const cert = xmlDoc.querySelector("X509Certificate").childNodes[0].nodeValue.replace(" ", "");
    const endpoint = xmlDoc.querySelector("SingleSignOnService").getAttribute("Location");
    const issuerUrl = xmlDoc.querySelector("EntityDescriptor").getAttribute("entityID");
    this.updateProviderField("idP", cert);
    this.updateProviderField("endpoint", endpoint);
    this.updateProviderField("issuerUrl", issuerUrl);
  }

  fetchSamlMetadata() {
    this.setState({
      metadataLoading: true,
    });
    fetch(this.state.requestUrl, {
      method: "GET",
    }).then(res => {
      if (!res.ok) {
        return Promise.reject("error");
      }
      return res.text();
    }).then(text => {
      this.updateProviderField("metadata", text);
      this.parseSamlMetadata();
      Setting.showMessage("success", i18next.t("general:Successfully added"));
    }).catch(err => {
      Setting.showMessage("error", err.message);
    }).finally(() => {
      this.setState({
        metadataLoading: false,
      });
    });
  }

  parseSamlMetadata() {
    try {
      this.loadSamlConfiguration();
      Setting.showMessage("success", i18next.t("provider:Parse metadata successfully"));
    } catch (err) {
      Setting.showMessage("error", i18next.t("provider:Can not parse metadata"));
    }
  }

  renderProvider() {
    return (
      <Card size="small" title={
        <div>
          {this.state.mode === "add" ? i18next.t("provider:New Provider") : i18next.t("provider:Edit Provider")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button onClick={() => this.submitProviderEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" onClick={() => this.submitProviderEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} onClick={() => this.deleteProvider()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      } style={(Setting.isMobile()) ? {margin: "5px"} : {}} type="inner">
        <Row style={{marginTop: "10px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Name"), i18next.t("general:Name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.provider.name} onChange={e => {
              this.updateProviderField("name", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Display name"), i18next.t("general:Display name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.provider.displayName} onChange={e => {
              this.updateProviderField("displayName", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} disabled={!Setting.isAdminUser(this.props.account)} value={this.state.provider.owner} onChange={(value => {this.updateProviderField("owner", value);})}>
              {Setting.isAdminUser(this.props.account) ? <Option key={"admin"} value={"admin"}>{i18next.t("provider:admin (Shared)")}</Option> : null}
              {
                this.state.organizations.map((organization, index) => <Option key={index} value={organization.name}>{organization.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Category"), i18next.t("general:Category - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.provider.category} onChange={(value => {
              this.updateProviderField("category", value);
              if (value === "OAuth") {
                this.updateProviderField("type", "Google");
              } else if (value === "Email") {
                this.updateProviderField("type", "Default");
                this.updateProviderField("host", "smtp.example.com");
                this.updateProviderField("port", 465);
                this.updateProviderField("sslMode", "Auto");
                this.updateProviderField("title", "Casdoor Verification Code");
                this.updateProviderField("content", Setting.getDefaultHtmlEmailContent());
                this.updateProviderField("metadata", Setting.getDefaultInvitationHtmlEmailContent());
                this.updateProviderField("receiver", this.props.account.email);
              } else if (value === "SMS") {
                this.updateProviderField("type", "Twilio SMS");
              } else if (value === "Storage") {
                this.updateProviderField("type", "AWS S3");
              } else if (value === "SAML") {
                this.updateProviderField("type", "Keycloak");
              } else if (value === "Payment") {
                this.updateProviderField("type", "PayPal");
              } else if (value === "Captcha") {
                this.updateProviderField("type", "Default");
              } else if (value === "Web3") {
                this.updateProviderField("type", "MetaMask");
              } else if (value === "Notification") {
                this.updateProviderField("type", "Telegram");
              } else if (value === "Face ID") {
                this.updateProviderField("type", "Alibaba Cloud Facebody");
              } else if (value === "MFA") {
                this.updateProviderField("type", "RADIUS");
                this.updateProviderField("host", "");
                this.updateProviderField("port", 1812);
              } else if (value === "ID Verification") {
                this.updateProviderField("type", "Jumio");
                this.updateProviderField("endpoint", "");
              }
            })}>
              {
                [
                  {id: "Captcha", name: "Captcha"},
                  {id: "Email", name: "Email"},
                  {id: "ID Verification", name: "ID Verification"},
                  {id: "MFA", name: "MFA"},
                  {id: "Notification", name: "Notification"},
                  {id: "OAuth", name: "OAuth"},
                  {id: "Payment", name: "Payment"},
                  {id: "SAML", name: "SAML"},
                  {id: "SMS", name: "SMS"},
                  {id: "Storage", name: "Storage"},
                  {id: "Web3", name: "Web3"},
                  {id: "Face ID", name: "Face ID"},
                ]
                  .sort((a, b) => a.name.localeCompare(b.name))
                  .map((providerCategory, index) => <Option key={index} value={providerCategory.id}>{providerCategory.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Type"), i18next.t("general:Type - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} showSearch value={this.state.provider.type} onChange={(value => {
              this.updateProviderField("type", value);
              if (value === "Local File System") {
                this.updateProviderField("domain", Setting.getFullServerUrl());
              } else if (value.startsWith("Custom") && this.state.provider.category === "OAuth") {
                this.updateProviderField("customAuthUrl", "https://door.casdoor.com/login/oauth/authorize");
                this.updateProviderField("scopes", "openid profile email");
                this.updateProviderField("customTokenUrl", "https://door.casdoor.com/api/login/oauth/access_token");
                this.updateProviderField("customUserInfoUrl", "https://door.casdoor.com/api/userinfo");
              } else if (value === "Custom HTTP SMS") {
                this.updateProviderField("endpoint", "https://example.com/send-custom-http-sms");
                this.updateProviderField("method", "GET");
                this.updateProviderField("title", "code");
              } else if (value === "Custom HTTP Email") {
                this.updateProviderField("endpoint", "https://example.com/send-custom-http-email");
                this.updateProviderField("method", "POST");
              } else if (value === "Custom HTTP") {
                this.updateProviderField("method", "GET");
                this.updateProviderField("title", "");
              }
            })}>
              {
                Setting.getProviderTypeOptions(this.state.provider.category)
                  .sort((a, b) => a.name.localeCompare(b.name))
                  .map((providerType, index) => <Option key={index} value={providerType.id}>
                    <img width={20} height={20} style={{marginBottom: "3px", marginRight: "10px"}} src={Setting.getProviderLogoURL({category: this.state.provider.category, type: providerType.id})} alt={providerType.id} />
                    {providerType.name}
                  </Option>)
              }
            </Select>
          </Col>
        </Row>
        {
          this.state.provider.type !== "WeCom" && this.state.provider.type !== "Infoflow" && this.state.provider.type !== "WeChat" ? null : (
            <React.Fragment>
              <Row style={{marginTop: "20px"}} >
                <Col style={{marginTop: "5px"}} span={2}>
                  {Setting.getLabel(i18next.t("provider:Sub type"), i18next.t("provider:Sub type - Tooltip"))} :
                </Col>
                <Col span={22} >
                  <Select virtual={false} style={{width: "100%"}} value={this.state.provider.subType} onChange={value => {
                    this.updateProviderField("subType", value);
                  }}>
                    {
                      this.getProviderSubTypeOptions(this.state.provider.type).map((providerSubType, index) => <Option key={index} value={providerSubType.id}>{providerSubType.name}</Option>)
                    }
                  </Select>
                </Col>
              </Row>
              {
                this.state.provider.type !== "WeCom" ? null : (
                  <React.Fragment>
                    <Row style={{marginTop: "20px"}} >
                      <Col style={{marginTop: "5px"}} span={2}>
                        {Setting.getLabel(i18next.t("general:Method"), i18next.t("provider:Method - Tooltip"))} :
                      </Col>
                      <Col span={22} >
                        <Select virtual={false} style={{width: "100%"}} value={this.state.provider.method} onChange={value => {
                          this.updateProviderField("method", value);
                        }}>
                          {
                            [
                              {id: "Normal", name: i18next.t("application:Normal")},
                              {id: "Silent", name: i18next.t("provider:Silent")},
                            ].map((method, index) => <Option key={index} value={method.id}>{method.name}</Option>)
                          }
                        </Select>
                      </Col>
                    </Row>
                    <Row style={{marginTop: "20px"}} >
                      <Col style={{marginTop: "5px"}} span={2}>
                        {Setting.getLabel(i18next.t("provider:Scope"), i18next.t("provider:Scope - Tooltip"))} :
                      </Col>
                      <Col span={22} >
                        <Select virtual={false} style={{width: "100%"}} value={this.state.provider.scopes} onChange={value => {
                          this.updateProviderField("scopes", value);
                        }}>
                          <Option key="snsapi_userinfo" value="snsapi_userinfo">snsapi_userinfo</Option>
                          <Option key="snsapi_privateinfo" value="snsapi_privateinfo">snsapi_privateinfo</Option>
                        </Select>
                      </Col>
                    </Row>
                    <Row style={{marginTop: "20px"}} >
                      <Col style={{marginTop: "5px"}} span={2}>
                        {Setting.getLabel(i18next.t("provider:Use id as name"), i18next.t("provider:Use id as name - Tooltip"))} :
                      </Col>
                      <Col span={22} >
                        <Switch checked={this.state.provider.disableSsl} onChange={checked => {
                          this.updateProviderField("disableSsl", checked);
                        }} />
                      </Col>
                    </Row>
                  </React.Fragment>)
              }
            </React.Fragment>
          )
        }
        {
          this.state.provider.category === "OAuth" ? renderOAuthProviderFields(
            this.state.provider,
            this.updateProviderField.bind(this),
            this.renderUserMappingInput.bind(this)
          ) : null
        }
        {
          (this.state.provider.category === "Captcha" && this.state.provider.type === "Default") ||
          (this.state.provider.category === "Web3") ||
          (this.state.provider.category === "MFA") ||
          (this.state.provider.category === "Storage" && this.state.provider.type === "Local File System") ||
          (this.state.provider.category === "SMS" && this.state.provider.type === "Custom HTTP SMS") ||
          (this.state.provider.category === "Email" && this.state.provider.type === "Custom HTTP Email") ||
          (this.state.provider.category === "Notification" && (this.state.provider.type === "Google Chat" || this.state.provider.type === "Custom HTTP" || this.state.provider.type === "Balance")) ? null : (
              <React.Fragment>
                {
                  (this.state.provider.category === "Storage" && this.state.provider.type === "Google Cloud Storage") ||
                  (this.state.provider.category === "Email" && (this.state.provider.type === "Azure ACS" || this.state.provider.type === "SendGrid")) ||
                  (this.state.provider.category === "Notification" && (this.state.provider.type === "Line" || this.state.provider.type === "Telegram" || this.state.provider.type === "Bark" || this.state.provider.type === "Discord" || this.state.provider.type === "Slack" || this.state.provider.type === "Pushbullet" || this.state.provider.type === "Pushover" || this.state.provider.type === "Lark" || this.state.provider.type === "Microsoft Teams" || this.state.provider.type === "WeCom")) ? null : (
                      <Row style={{marginTop: "20px"}} >
                        <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                          {this.getClientIdLabel(this.state.provider)} :
                        </Col>
                        <Col span={22} >
                          <Input value={this.state.provider.clientId} onChange={e => {
                            this.updateProviderField("clientId", e.target.value);
                          }} />
                        </Col>
                      </Row>
                    )
                }
                <Row style={{marginTop: "20px"}} >
                  <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                    {this.getClientSecretLabel(this.state.provider)} :
                  </Col>
                  <Col span={22} >
                    <Input value={this.state.provider.clientSecret} onChange={e => {
                      this.updateProviderField("clientSecret", e.target.value);
                    }} />
                  </Col>
                </Row>
              </React.Fragment>
            )
        }
        {
          this.state.provider.category !== "Email" && this.state.provider.type !== "WeChat" && this.state.provider.type !== "Apple" && this.state.provider.type !== "Aliyun Captcha" && this.state.provider.type !== "WeChat Pay" && this.state.provider.type !== "Twitter" && this.state.provider.type !== "Reddit" && this.state.provider.type !== "CUCloud" ? null : (
            <React.Fragment>
              <Row style={{marginTop: "20px"}} >
                <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                  {this.getClientId2Label(this.state.provider)} :
                </Col>
                <Col span={22} >
                  <Input value={this.state.provider.clientId2} onChange={e => {
                    this.updateProviderField("clientId2", e.target.value);
                  }} />
                </Col>
              </Row>
              {
                (this.state.provider.type === "WeChat Pay" || this.state.provider.type === "CUCloud") || (this.state.provider.category === "Email" && (this.state.provider.type === "Azure ACS")) ? null : (
                  <Row style={{marginTop: "20px"}} >
                    <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                      {this.getClientSecret2Label(this.state.provider)} :
                    </Col>
                    <Col span={22} >
                      {
                        (this.state.provider.category === "OAuth" && this.state.provider.type === "Apple") ? (
                          <TextArea autoSize={{minRows: 1, maxRows: 20}} value={this.state.provider.clientSecret2} onChange={e => {
                            this.updateProviderField("clientSecret2", e.target.value);
                          }} />
                        ) : (
                          <Input value={this.state.provider.clientSecret2} onChange={e => {
                            this.updateProviderField("clientSecret2", e.target.value);
                          }} />
                        )
                      }
                    </Col>
                  </Row>
                )
              }
            </React.Fragment>
          )
        }
        {this.getAppIdRow(this.state.provider)}
        {
          this.state.provider.category === "Notification" ? renderNotificationProviderFields(
            this.state.provider,
            this.updateProviderField.bind(this),
            this.getReceiverRow.bind(this)
          ) : this.state.provider.category === "Email" ? renderEmailProviderFields(
            this.state.provider,
            this.updateProviderField.bind(this),
            this.renderEmailMappingInput.bind(this),
            this.props.account
          ) : ["SMS"].includes(this.state.provider.category) ? renderSmsProviderFields(
            this.state.provider,
            this.updateProviderField.bind(this),
            this.renderSmsMappingInput.bind(this),
            this.props.account
          ) : this.state.provider.category === "MFA" ? renderMfaProviderFields(
            this.state.provider,
            this.updateProviderField.bind(this)
          ) : this.state.provider.category === "SAML" ? renderSamlProviderFields(
            this.state.provider,
            this.updateProviderField.bind(this),
            {
              requestUrl: this.state.requestUrl,
              setRequestUrl: (value) => this.setState({requestUrl: value}),
              metadataLoading: this.state.metadataLoading,
              fetchSamlMetadata: this.fetchSamlMetadata.bind(this),
              parseSamlMetadata: this.parseSamlMetadata.bind(this),
            }
          ) : null
        }
        {this.state.provider.category === "Payment" ? renderPaymentProviderFields(
          this.state.provider,
          this.updateProviderField.bind(this),
          this.state.certs
        ) : null}
        {this.state.provider.category === "Web3" ? renderWeb3ProviderFields(
          this.state.provider,
          this.updateProviderField.bind(this)
        ) : null}
        {this.state.provider.category === "Storage" ? renderStorageProviderFields(
          this.state.provider,
          this.updateProviderField.bind(this)
        ) : null}
        {this.state.provider.category === "Face ID" ? renderFaceIdProviderFields(
          this.state.provider,
          this.updateProviderField.bind(this)
        ) : null}
        {this.state.provider.category === "ID Verification" ? renderIDVerificationProviderFields(
          this.state.provider,
          this.updateProviderField.bind(this)
        ) : null}
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("provider:Provider URL"), i18next.t("provider:Provider URL - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined />} value={this.state.provider.providerUrl} onChange={e => {
              this.updateProviderField("providerUrl", e.target.value);
            }} />
          </Col>
        </Row>
        {
          this.state.provider.category === "Captcha" ? renderCaptchaProviderFields(
            this.state.provider,
            this.state.providerName
          ) : null
        }
      </Card>
    );
  }

  submitProviderEdit(exitAfterSave) {
    const provider = Setting.deepCopy(this.state.provider);
    ProviderBackend.updateProvider(this.state.owner, this.state.providerName, provider)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully saved"));
          this.setState({
            owner: this.state.provider.owner,
            providerName: this.state.provider.name,
          });

          if (exitAfterSave) {
            this.props.history.push("/providers");
          } else {
            this.props.history.push(`/providers/${this.state.provider.owner}/${this.state.provider.name}`);
          }
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
          this.updateProviderField("name", this.state.providerName);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteProvider() {
    ProviderBackend.deleteProvider(this.state.provider)
      .then((res) => {
        if (res.status === "ok") {
          this.props.history.push("/providers");
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to delete")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  render() {
    return (
      <div>
        {
          this.state.provider !== null ? this.renderProvider() : null
        }
        <div style={{marginTop: "20px", marginLeft: "40px"}}>
          <Button size="large" onClick={() => this.submitProviderEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" size="large" onClick={() => this.submitProviderEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} size="large" onClick={() => this.deleteProvider()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      </div>
    );
  }
}

export default ProviderEditPage;
