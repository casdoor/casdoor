// Copyright 2023 The casbin Authors. All Rights Reserved.
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
import {Button, Card, Col, Input, InputNumber, Row, Select, Switch} from "antd";
import {LinkOutlined} from "@ant-design/icons";
import * as ProviderBackend from "./backend/ProviderBackend";
import * as SiteBackend from "./backend/SiteBackend";
import * as CertBackend from "./backend/CertBackend";
import * as RuleBackend from "./backend/RuleBackend";
import * as ApplicationBackend from "./backend/ApplicationBackend";
import * as Setting from "./Setting";
import i18next from "i18next";
import RuleTable from "./table/RuleTable";
import * as OrganizationBackend from "./backend/OrganizationBackend";

const {Option} = Select;

class SiteEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      owner: props.match.params.organizationName,
      siteName: props.match.params.siteName,
      rules: [],
      providers: [],
      site: null,
      certs: null,
      applications: null,
      organizations: [],
    };
  }

  UNSAFE_componentWillMount() {
    this.getOrganizations();
    this.getSite();
    this.getCerts();
    this.getRules();
    this.getApplications();
    this.getAlertProviders();
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

  getSite() {
    SiteBackend.getSite(this.state.site?.owner || this.state.owner, this.state.siteName)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            site: res.data,
          });
        } else {
          Setting.showMessage("error", `Failed to get site: ${res.msg}`);
        }
      });
  }

  getCerts() {
    CertBackend.getCerts(this.state.owner)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            certs: res.data,
          });
        } else {
          Setting.showMessage("error", `Failed to get certs: ${res.msg}`);
        }
      });
  }

  getRules() {
    RuleBackend.getRules(this.state.owner)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            rules: res.data,
          });
        } else {
          Setting.showMessage("error", `Failed to get rules: ${res.msg}`);
        }
      });
  }

  getApplications(owner) {
    ApplicationBackend.getApplicationsByOrganization("admin", owner || this.state.owner)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            applications: res.data,
          });
        } else {
          Setting.showMessage("error", `Failed to get applications: ${res.msg}`);
        }
      });
  }

  getAlertProviders() {
    ProviderBackend.getProviders()
      .then((res) => {
        if (res.status === "ok") {
          const data = [];
          for (let i = 0; i < res.data.length; i++) {
            const provider = res.data[i];
            if (provider.category === "SMS" || provider.category === "Email") {
              data.push(provider.category + "/" + provider.name);
            }
          }
          this.setState({
            providers: data,
          });
        } else {
          Setting.showMessage("error", `Failed to get providers: ${res.msg}`);
        }
      });
  }

  parseSiteField(key, value) {
    if (["score"].includes(key)) {
      value = Setting.myParseInt(value);
    }
    return value;
  }

  updateSiteField(key, value) {
    value = this.parseSiteField(key, value);

    const site = this.state.site;
    site[key] = value;
    this.setState({
      site: site,
    });
  }

  renderSite() {
    return (
      <Card size="small" title={
        <div>
          {i18next.t("site:Edit Site")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button type="primary" onClick={this.submitSiteEdit.bind(this)}>{i18next.t("general:Save")}</Button>
        </div>
      } style={{marginLeft: "5px"}} type="inner">
        <Row style={{marginTop: "10px"}} >
          <Col style={{marginTop: "5px"}} span={2}>
            {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} disabled={!Setting.isAdminUser(this.props.account)} value={this.state.site.owner} onChange={(value => {
              this.updateSiteField("owner", value);
              this.getApplications(value);
            })}>
              {
                this.state.organizations.map((organization, index) => <Option key={index} value={organization.name}>{organization.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "10px"}} >
          <Col style={{marginTop: "5px"}} span={2}>
            {i18next.t("general:Name")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.site.name} onChange={e => {
              this.updateSiteField("name", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={2}>
            {i18next.t("general:Display name")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.site.displayName} onChange={e => {
              this.updateSiteField("displayName", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={2}>
            {i18next.t("general:Tag")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.site.tag} onChange={e => {
              this.updateSiteField("tag", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={2}>
            {i18next.t("site:Domain")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.site.domain} onChange={e => {
              this.updateSiteField("domain", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={2}>
            {i18next.t("site:Other domains")}:
          </Col>
          <Col span={22} >
            <Select virtual={false} mode="tags" style={{width: "100%"}} value={this.state.site.otherDomains} onChange={(value => {this.updateSiteField("otherDomains", value);})}>
              {
                this.state.site.otherDomains?.map((item, index) => <Option key={index} value={item}>{item}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={2}>
            {i18next.t("site:Need redirect")}:
          </Col>
          <Col span={1} >
            <Switch checked={this.state.site.needRedirect} onChange={checked => {
              this.updateSiteField("needRedirect", checked);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={2}>
            {i18next.t("site:Disable verbose")}:
          </Col>
          <Col span={1} >
            <Switch checked={this.state.site.disableVerbose} onChange={checked => {
              this.updateSiteField("disableVerbose", checked);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={2}>
            {i18next.t("site:Rules")}:
          </Col>
          <Col span={22} >
            <RuleTable
              title={"Rules"}
              account={this.props.account}
              sources={this.state.rules}
              rules={this.state.site.rules}
              onUpdateRules={(value) => this.updateSiteField("rules", value)}
            />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col span={2} style={{marginTop: "5px"}}>
            {i18next.t("site:Enable alert")}:
          </Col>
          <Col span={22} >
            <Switch checked={this.state.site.enableAlert} onChange={checked => {
              this.updateSiteField("enableAlert", checked);
            }} />
          </Col>
        </Row>
        {
          this.state.site.enableAlert ? (
            <Row style={{marginTop: "20px"}} >
              <Col span={2} style={{marginTop: "5px"}}>
                {i18next.t("site:Alert interval")}:
              </Col>
              <Col span={22} >
                <InputNumber min={1} value={this.state.site.alertInterval} addonAfter={i18next.t("usage:seconds")} onChange={value => {
                  this.updateSiteField("alertInterval", value);
                }} />
              </Col>
            </Row>
          ) : null
        }
        {
          this.state.site.enableAlert ? (
            <Row style={{marginTop: "20px"}} >
              <Col span={2} style={{marginTop: "5px"}}>
                {i18next.t("site:Alert try times")}:
              </Col>
              <Col span={22} >
                <InputNumber min={1} value={this.state.site.alertTryTimes} onChange={value => {
                  this.updateSiteField("alertTryTimes", value);
                }} />
              </Col>
            </Row>
          ) : null
        }
        {
          this.state.site.enableAlert ? (
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={2}>
                {i18next.t("site:Alert providers")}:
              </Col>
              <Col span={22} >
                <Select virtual={false} mode="tags" style={{width: "100%"}} value={this.state.site.alertProviders} onChange={(value => {this.updateSiteField("alertProviders", value);})}>
                  {
                    this.state.providers.map((item, index) => <Option key={index} value={item}>{item}</Option>)
                  }
                </Select>
              </Col>
            </Row>
          ) : null
        }
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={2}>
            {i18next.t("site:Challenges")}:
          </Col>
          <Col span={22} >
            <Select virtual={false} mode="tags" style={{width: "100%"}} value={this.state.site.challenges} onChange={(value => {this.updateSiteField("challenges", value);})}>
              {
                this.state.site.challenges?.map((item, index) => <Option key={index} value={item}>{item}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={2}>
            {i18next.t("site:Host")}:
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined />} value={this.state.site.host} onChange={e => {
              this.updateSiteField("host", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={2}>
            {i18next.t("site:Port")}:
          </Col>
          <Col span={22} >
            <InputNumber min={0} max={65535} value={this.state.site.port} onChange={value => {
              this.updateSiteField("port", value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={2}>
            {i18next.t("site:Hosts")}:
          </Col>
          <Col span={22} >
            <Select virtual={false} mode="tags" style={{width: "100%"}} value={this.state.site.hosts} onChange={(value => {this.updateSiteField("hosts", value);})}>
              {
                this.state.site.hosts?.map((item, index) => <Option key={index} value={item}>{item}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={2}>
            {i18next.t("site:Public IP")}:
          </Col>
          <Col span={22} >
            <Input disabled={true} value={this.state.site.publicIp} onChange={e => {
              this.updateSiteField("publicIp", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={2}>
            {i18next.t("site:Mode")}:
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.site.sslMode} onChange={(value => {this.updateSiteField("sslMode", value);})}>
              {
                [
                  {id: "HTTP", name: "HTTP"},
                  {id: "HTTPS and HTTP", name: "HTTPS and HTTP"},
                  {id: "HTTPS Only", name: "HTTPS Only"},
                  {id: "Static Folder", name: "Static Folder"},
                ].map((item, index) => <Option key={index} value={item.id}>{item.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={2}>
            {i18next.t("site:SSL cert")}:
          </Col>
          <Col span={22} >
            <Select disabled={true} virtual={false} style={{width: "100%"}} showSearch value={this.state.site.sslCert} onChange={(value => {
              this.updateSiteField("sslCert", value);
            })}>
              {
                this.state.certs?.map((cert, index) => <Option key={index} value={cert.name}>{cert.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={2}>
            {i18next.t("site:Casdoor app")}:
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} showSearch value={this.state.site.casdoorApplication} onChange={(value => {
              this.updateSiteField("casdoorApplication", value);
            })}>
              {
                this.state.applications?.map((application, index) => <Option key={index} value={application.name}>{application.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={2}>
            {i18next.t("site:Status")}:
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.site.status} onChange={(value => {this.updateSiteField("status", value);})}>
              {
                [
                  {id: "Active", name: "Active"},
                  {id: "Inactive", name: "Inactive"},
                ].map((item, index) => <Option key={index} value={item.id}>{item.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
      </Card>
    );
  }

  submitSiteEdit() {
    const site = Setting.deepCopy(this.state.site);
    SiteBackend.updateSite(this.state.owner, this.state.siteName, site)
      .then((res) => {
        if (res.status === "error") {
          Setting.showMessage("error", `Failed to save: ${res.msg}`);
          this.updateSiteField("name", this.state.siteName);
        } else {
          Setting.showMessage("success", "Successfully saved");
          this.setState({
            owner: this.state.site.owner,
            siteName: this.state.site.name,
          });
          this.props.history.push(`/sites/${this.state.site.owner}/${this.state.site.name}`);
          this.getSite();
        }
      })
      .catch(error => {
        Setting.showMessage("error", `failed to save: ${error}`);
      });
  }

  render() {
    return (
      <div>
        <Row style={{width: "100%"}}>
          <Col span={1}>
          </Col>
          <Col span={22}>
            {
              this.state.site !== null ? this.renderSite() : null
            }
          </Col>
          <Col span={1}>
          </Col>
        </Row>
        <Row style={{margin: 10}}>
          <Col span={2}>
          </Col>
          <Col span={18}>
            <Button type="primary" size="large" onClick={this.submitSiteEdit.bind(this)}>{i18next.t("general:Save")}</Button>
          </Col>
        </Row>
      </div>
    );
  }
}

export default SiteEditPage;
