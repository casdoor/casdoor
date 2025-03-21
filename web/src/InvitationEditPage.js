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
import {Button, Card, Col, Input, InputNumber, Row, Select} from "antd";
import * as InvitationBackend from "./backend/InvitationBackend";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as ApplicationBackend from "./backend/ApplicationBackend";
import * as Setting from "./Setting";
import i18next from "i18next";
import copy from "copy-to-clipboard";
import * as GroupBackend from "./backend/GroupBackend";

const {Option} = Select;

class InvitationEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      organizationName: props.organizationName !== undefined ? props.organizationName : props.match.params.organizationName,
      invitationName: props.match.params.invitationName,
      invitation: null,
      organizations: [],
      applications: [],
      groups: [],
      mode: props.location.mode !== undefined ? props.location.mode : "edit",
    };
  }

  UNSAFE_componentWillMount() {
    this.getInvitation();
    this.getOrganizations();
    this.getApplicationsByOrganization(this.state.organizationName);
    this.getGroupsByOrganization(this.state.organizationName);
  }

  getInvitation() {
    InvitationBackend.getInvitation(this.state.organizationName, this.state.invitationName)
      .then((res) => {
        if (res.data === null) {
          this.props.history.push("/404");
          return;
        }

        this.setState({
          invitation: res.data,
        });
      });
  }

  getOrganizations() {
    OrganizationBackend.getOrganizations("admin")
      .then((res) => {
        this.setState({
          organizations: res.data || [],
        });
      });
  }

  getApplicationsByOrganization(organizationName) {
    ApplicationBackend.getApplicationsByOrganization("admin", organizationName)
      .then((res) => {
        this.setState({
          applications: res.data || [],
        });
      });
  }

  getGroupsByOrganization(organizationName) {
    GroupBackend.getGroups(organizationName)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            groups: res.data,
          });
        }
      });
  }

  parseInvitationField(key, value) {
    if ([""].includes(key)) {
      value = Setting.myParseInt(value);
    }
    return value;
  }

  updateInvitationField(key, value) {
    value = this.parseInvitationField(key, value);

    const invitation = this.state.invitation;
    invitation[key] = value;
    this.setState({
      invitation: invitation,
    });
  }

  copySignupLink() {
    let defaultApplication;
    if (this.state.invitation.owner === "built-in") {
      defaultApplication = "app-built-in";
    } else {
      const selectedOrganization = Setting.getArrayItem(this.state.organizations, "name", this.state.invitation.owner);
      defaultApplication = selectedOrganization.defaultApplication;
      if (!defaultApplication) {
        Setting.showMessage("error", i18next.t("invitation:You need to first specify a default application for organization: ") + selectedOrganization.name);
        return;
      }
    }
    copy(`${window.location.origin}/signup/${defaultApplication}?invitationCode=${this.state.invitation?.defaultCode}`);
    Setting.showMessage("success", i18next.t("general:Copied to clipboard successfully"));
  }

  renderInvitation() {
    const isCreatedByPlan = this.state.invitation.tag === "auto_created_invitation_for_plan";
    return (
      <Card size="small" title={
        <div>
          {this.state.mode === "add" ? i18next.t("invitation:New Invitation") : i18next.t("invitation:Edit Invitation")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button onClick={() => this.submitInvitationEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" onClick={() => this.submitInvitationEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          <Button style={{marginLeft: "20px"}} onClick={_ => this.copySignupLink()}>
            {i18next.t("application:Copy signup page URL")}
          </Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} onClick={() => this.deleteInvitation()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      } style={(Setting.isMobile()) ? {margin: "5px"} : {}} type="inner">
        <Row style={{marginTop: "10px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} disabled={!Setting.isAdminUser(this.props.account) || isCreatedByPlan} value={this.state.invitation.owner} onChange={(value => {this.updateInvitationField("owner", value); this.getApplicationsByOrganization(value);this.getGroupsByOrganization(value);})}>
              {
                this.state.organizations.map((organization, index) => <Option key={index} value={organization.name}>{organization.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Name"), i18next.t("general:Name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.invitation.name} disabled={isCreatedByPlan} onChange={e => {
              this.updateInvitationField("name", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Display name"), i18next.t("general:Display name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.invitation.displayName} onChange={e => {
              this.updateInvitationField("displayName", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("invitation:Code"), i18next.t("invitation:Code - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.invitation.code} onChange={e => {
              const regex = /[^a-zA-Z0-9]/;
              if (!regex.test(e.target.value)) {
                this.updateInvitationField("defaultCode", e.target.value);
              }
              this.updateInvitationField("code", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("invitation:Default code"), i18next.t("invitation:Default code - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.invitation.defaultCode} onChange={e => {
              this.updateInvitationField("defaultCode", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("invitation:Quota"), i18next.t("invitation:Quota - Tooltip"))} :
          </Col>
          <Col span={22} >
            <InputNumber min={0} value={this.state.invitation.quota} onChange={value => {
              this.updateInvitationField("quota", value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("invitation:Used count"), i18next.t("invitation:Used count - Tooltip"))} :
          </Col>
          <Col span={22} >
            <InputNumber min={0} max={this.state.invitation.quota} value={this.state.invitation.usedCount} onChange={value => {
              this.updateInvitationField("usedCount", value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Application"), i18next.t("general:Application - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.invitation.application}
              onChange={(value => {this.updateInvitationField("application", value);})}
              options={[
                {label: i18next.t("general:All"), value: "All"},
                ...this.state.applications.map((application) => Setting.getOption(application.name, application.name)),
              ]} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("provider:Signup group"), i18next.t("provider:Signup group - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.invitation.signupGroup} onChange={(value => {this.updateInvitationField("signupGroup", value);})}>
              <Option key={""} value={""}>
                {i18next.t("general:Default")}
              </Option>
              {
                this.state.groups.map((group, index) => <Option key={index} value={`${group.owner}/${group.name}`}>{group.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("signup:Username"), i18next.t("signup:Username - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.invitation.username} onChange={e => {
              this.updateInvitationField("username", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Email"), i18next.t("general:Email - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.invitation.email} onChange={e => {
              this.updateInvitationField("email", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Phone"), i18next.t("general:Phone - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.invitation.phone} onChange={e => {
              this.updateInvitationField("phone", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:State"), i18next.t("general:State - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.invitation.state} onChange={(value => {
              this.updateInvitationField("state", value);
            })}
            options={[
              {value: "Active", name: i18next.t("subscription:Active")},
              {value: "Suspended", name: i18next.t("subscription:Suspended")},
            ].map((item) => Setting.getOption(item.name, item.value))}
            />
          </Col>
        </Row>
      </Card>
    );
  }

  submitInvitationEdit(exitAfterSave) {
    const invitation = Setting.deepCopy(this.state.invitation);
    InvitationBackend.updateInvitation(this.state.organizationName, this.state.invitationName, invitation)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully saved"));
          this.setState({
            invitationName: this.state.invitation.name,
          });

          if (exitAfterSave) {
            this.props.history.push("/invitations");
          } else {
            this.props.history.push(`/invitations/${this.state.invitation.owner}/${this.state.invitation.name}`);
          }
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
          this.updateInvitationField("name", this.state.invitationName);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteInvitation() {
    InvitationBackend.deleteInvitation(this.state.invitation)
      .then((res) => {
        if (res.status === "ok") {
          this.props.history.push("/invitations");
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
          this.state.invitation !== null ? this.renderInvitation() : null
        }
        <div style={{marginTop: "20px", marginLeft: "40px"}}>
          <Button size="large" onClick={() => this.submitInvitationEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" size="large" onClick={() => this.submitInvitationEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          <Button style={{marginLeft: "20px"}} size="large" onClick={_ => this.copySignupLink()}>
            {i18next.t("application:Copy signup page URL")}
          </Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} size="large" onClick={() => this.deleteInvitation()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      </div>
    );
  }
}

export default InvitationEditPage;
