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
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as UserBackend from "./backend/UserBackend";
import * as Setting from "./Setting";
import i18next from "i18next";

class ChatEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      chatName: props.match.params.chatName,
      chat: null,
      organizations: [],
      users: [],
      mode: props.location.mode !== undefined ? props.location.mode : "edit",
    };
  }

  UNSAFE_componentWillMount() {
    this.getChat();
    this.getOrganizations();
  }

  getChat() {
    ChatBackend.getChat("admin", this.state.chatName)
      .then((chat) => {
        if (chat === null) {
          this.props.history.push("/404");
        }
        this.setState({
          chat: chat,
        });

        this.getUsers(chat.organization);
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

  getUsers(organizationName) {
    UserBackend.getUsers(organizationName)
      .then((res) => {
        this.setState({
          users: res,
        });
      });
  }

  parseChatField(key, value) {
    if ([].includes(key)) {
      value = Setting.myParseInt(value);
    }
    return value;
  }

  updateChatField(key, value) {
    value = this.parseChatField(key, value);

    const chat = this.state.chat;
    chat[key] = value;
    this.setState({
      chat: chat,
    });
  }

  renderChat() {
    return (
      <Card size="small" title={
        <div>
          {this.state.mode === "add" ? i18next.t("chat:New Chat") : i18next.t("chat:Edit Chat")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button onClick={() => this.submitChatEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" onClick={() => this.submitChatEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} onClick={() => this.deleteChat()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      } style={(Setting.isMobile()) ? {margin: "5px"} : {}} type="inner">
        <Row style={{marginTop: "10px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} disabled={!Setting.isAdminUser(this.props.account)} style={{width: "100%"}} value={this.state.chat.organization} onChange={(value => {this.updateChatField("organization", value);})}
              options={this.state.organizations.map((organization) => Setting.getOption(organization.name, organization.name))
              } />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Name"), i18next.t("general:Name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.chat.name} onChange={e => {
              this.updateChatField("name", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Display name"), i18next.t("general:Display name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.chat.displayName} onChange={e => {
              this.updateChatField("displayName", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("provider:Type"), i18next.t("provider:Type - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.chat.type} onChange={(value => {
              this.updateChatField("type", value);
            })}
            options={[
              {value: "Single", name: i18next.t("chat:Single")},
              {value: "Group", name: i18next.t("chat:Group")},
              {value: "AI", name: i18next.t("chat:AI")},
            ].map((item) => Setting.getOption(item.name, item.value))}
            />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("provider:Category"), i18next.t("provider:Category - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.chat.category} onChange={e => {
              this.updateChatField("category", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("chat:User1"), i18next.t("chat:User1 - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.chat.user1} onChange={(value => {this.updateChatField("user1", value);})}
              options={this.state.users.map((user) => Setting.getOption(`${user.owner}/${user.name}`, `${user.owner}/${user.name}`))
              } />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("chat:User2"), i18next.t("chat:User2 - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.chat.user2} onChange={(value => {this.updateChatField("user2", value);})}
              options={this.state.users.map((user) => Setting.getOption(`${user.owner}/${user.name}`, `${user.owner}/${user.name}`))
              } />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Users"), i18next.t("chat:Users - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} mode="multiple" style={{width: "100%"}} value={this.state.chat.users}
              onChange={(value => {this.updateChatField("users", value);})}
              options={this.state.users.map((user) => Setting.getOption(`${user.owner}/${user.name}`, `${user.owner}/${user.name}`))}
            />
          </Col>
        </Row>
      </Card>
    );
  }

  submitChatEdit(willExist) {
    const chat = Setting.deepCopy(this.state.chat);
    ChatBackend.updateChat(this.state.chat.owner, this.state.chatName, chat)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully saved"));
          this.setState({
            chatName: this.state.chat.name,
          });

          if (willExist) {
            this.props.history.push("/chats");
          } else {
            this.props.history.push(`/chats/${this.state.chat.name}`);
          }
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
          this.updateChatField("name", this.state.chatName);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteChat() {
    ChatBackend.deleteChat(this.state.chat)
      .then((res) => {
        if (res.status === "ok") {
          this.props.history.push("/chats");
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
          this.state.chat !== null ? this.renderChat() : null
        }
        <div style={{marginTop: "20px", marginLeft: "40px"}}>
          <Button size="large" onClick={() => this.submitChatEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" size="large" onClick={() => this.submitChatEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} size="large" onClick={() => this.deleteChat()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      </div>
    );
  }
}

export default ChatEditPage;
