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
import {Button, Card, Col, Input, Row, Select} from "antd";
import * as ChatBackend from "./backend/ChatBackend";
import * as MessageBackend from "./backend/MessageBackend";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as UserBackend from "./backend/UserBackend";
import * as Setting from "./Setting";
import i18next from "i18next";

const {TextArea} = Input;

class MessageEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      messageName: props.match.params.messageName,
      message: null,
      organizations: [],
      chats: [],
      users: [],
      mode: props.location.mode !== undefined ? props.location.mode : "edit",
    };
  }

  UNSAFE_componentWillMount() {
    this.getMessage();
    this.getOrganizations();
    this.getChats();
  }

  getMessage() {
    MessageBackend.getMessage("admin", this.state.messageName)
      .then((message) => {
        this.setState({
          message: message,
        });

        this.getUsers(message.organization);
      });
  }

  getOrganizations() {
    OrganizationBackend.getOrganizations("admin")
      .then((res) => {
        this.setState({
          organizations: (res.msg === undefined) ? res : [],
        });
      });
  }

  getChats() {
    ChatBackend.getChats("admin")
      .then((res) => {
        this.setState({
          chats: (res.msg === undefined) ? res : [],
        });
      });
  }

  getUsers(organizationName) {
    UserBackend.getUsers(organizationName)
      .then((res) => {
        this.setState({
          users: res,
        });
      });
  }

  parseMessageField(key, value) {
    if ([].includes(key)) {
      value = Setting.myParseInt(value);
    }
    return value;
  }

  updateMessageField(key, value) {
    value = this.parseMessageField(key, value);

    const message = this.state.message;
    message[key] = value;
    this.setState({
      message: message,
    });
  }

  renderMessage() {
    return (
      <Card size="small" title={
        <div>
          {this.state.mode === "add" ? i18next.t("message:New Message") : i18next.t("message:Edit Message")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button onClick={() => this.submitMessageEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" onClick={() => this.submitMessageEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} onClick={() => this.deleteMessage()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      } style={(Setting.isMobile()) ? {margin: "5px"} : {}} type="inner">
        <Row style={{marginTop: "10px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} disabled={!Setting.isAdminUser(this.props.account)} style={{width: "100%"}} value={this.state.message.organization} onChange={(value => {this.updateMessageField("organization", value);})}
              options={this.state.organizations.map((organization) => Setting.getOption(organization.name, organization.name))
              } />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Name"), i18next.t("general:Name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.message.name} onChange={e => {
              this.updateMessageField("name", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("message:Chat"), i18next.t("message:Chat - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.message.chat} onChange={(value => {this.updateMessageField("chat", value);})}
              options={this.state.chats.map((chat) => Setting.getOption(chat.name, chat.name))
              } />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("message:Author"), i18next.t("message:Author - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.message.author} onChange={(value => {this.updateMessageField("author", value);})}
              options={this.state.users.map((user) => Setting.getOption(`${user.owner}/${user.name}`, `${user.owner}/${user.name}`))
              } />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("message:Text"), i18next.t("message:Text - Tooltip"))} :
          </Col>
          <Col span={22}>
            <TextArea rows={10} value={this.state.message.text} onChange={e => {
              this.updateMessageField("text", e.target.value);
            }} />
          </Col>
        </Row>
      </Card>
    );
  }

  submitMessageEdit(willExist) {
    const message = Setting.deepCopy(this.state.message);
    MessageBackend.updateMessage(this.state.message.owner, this.state.messageName, message)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully saved"));
          this.setState({
            messageName: this.state.message.name,
          });

          if (willExist) {
            this.props.history.push("/messages");
          } else {
            this.props.history.push(`/messages/${this.state.message.name}`);
          }
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
          this.updateMessageField("name", this.state.messageName);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteMessage() {
    MessageBackend.deleteMessage(this.state.message)
      .then((res) => {
        if (res.status === "ok") {
          this.props.history.push("/messages");
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
          this.state.message !== null ? this.renderMessage() : null
        }
        <div style={{marginTop: "20px", marginLeft: "40px"}}>
          <Button size="large" onClick={() => this.submitMessageEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" size="large" onClick={() => this.submitMessageEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} size="large" onClick={() => this.deleteMessage()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      </div>
    );
  }
}

export default MessageEditPage;
