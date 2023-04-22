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
import {Spin} from "antd";
import moment from "moment";
import ChatMenu from "./ChatMenu";
import ChatBox from "./ChatBox";
import * as Setting from "./Setting";
import * as ChatBackend from "./backend/ChatBackend";
import * as MessageBackend from "./backend/MessageBackend";
import i18next from "i18next";
import BaseListPage from "./BaseListPage";

class ChatPage extends BaseListPage {
  newChat(chat) {
    const randomName = Setting.getRandomName();
    return {
      owner: "admin", // this.props.account.applicationName,
      name: `chat_${randomName}`,
      createdTime: moment().format(),
      updatedTime: moment().format(),
      organization: this.props.account.owner,
      displayName: `New Chat - ${randomName}`,
      type: "AI",
      category: chat !== undefined ? chat.category : "Chat Category - 1",
      user1: `${this.props.account.owner}/${this.props.account.name}`,
      user2: "",
      users: [`${this.props.account.owner}/${this.props.account.name}`],
      messageCount: 0,
    };
  }

  newMessage(text) {
    const randomName = Setting.getRandomName();
    return {
      owner: "admin", // this.props.account.messagename,
      name: `message_${randomName}`,
      createdTime: moment().format(),
      organization: this.props.account.owner,
      chat: this.state.chatName,
      author: `${this.props.account.owner}/${this.props.account.name}`,
      text: text,
    };
  }

  sendMessage(text) {
    const newMessage = this.newMessage(text);
    MessageBackend.addMessage(newMessage)
      .then((res) => {
        if (res.status === "ok") {
          this.getMessages(this.state.chatName);
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to add")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  getMessages(chatName) {
    MessageBackend.getChatMessages(chatName)
      .then((messages) => {
        this.setState({
          messages: messages,
        });

        Setting.scrollToDiv(`chatbox-list-item-${messages.length}`);
      });
  }

  addChat(chat) {
    const newChat = this.newChat(chat);
    ChatBackend.addChat(newChat)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully added"));
          this.setState({
            chatName: newChat.name,
            messages: null,
          });
          this.getMessages(newChat.name);

          const {pagination} = this.state;
          this.fetch({pagination});
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to add")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteChat(chats, i, chat) {
    ChatBackend.deleteChat(chat)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully deleted"));
          const data = Setting.deleteRow(this.state.data, i);
          const j = Math.min(i, data.length - 1);
          if (j < 0) {
            this.setState({
              chatName: undefined,
              messages: undefined,
              data: data,
            });
          } else {
            const focusedChat = data[j];
            this.setState({
              chatName: focusedChat.name,
              messages: null,
              data: data,
            });
            this.getMessages(focusedChat.name);
          }
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to delete")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  renderTable(chats) {
    const onSelectChat = (i) => {
      const chat = chats[i];
      this.setState({
        chatName: chat.name,
        messages: null,
      });
      this.getMessages(chat.name);
    };

    const onAddChat = () => {
      const chat = this.state.data.filter(chat => chat.name === this.state.chatName)[0];
      this.addChat(chat);
    };

    const onDeleteChat = (i) => {
      const chat = chats[i];
      this.deleteChat(chats, i, chat);
    };

    if (this.state.loading) {
      return (
        <div style={{display: "flex", justifyContent: "center", alignItems: "center"}}>
          <Spin size="large" tip={i18next.t("login:Loading")} style={{paddingTop: "10%"}} />
        </div>
      );
    }

    return (
      <div style={{display: "flex", height: "calc(100vh - 140px)"}}>
        <div style={{width: "250px", height: "100%", backgroundColor: "white", borderRight: "1px solid rgb(245,245,245)"}}>
          <ChatMenu chats={chats} onSelectChat={onSelectChat} onAddChat={onAddChat} onDeleteChat={onDeleteChat} />
        </div>
        <div style={{flex: 1, height: "100%", backgroundColor: "white", position: "relative"}}>
          {
            this.state.messages === null ? null : (
              <div style={{
                position: "absolute",
                top: -50,
                left: 0,
                right: 0,
                bottom: 0,
                backgroundImage: "url(https://cdn.casbin.org/img/casdoor-logo_1185x256.png)",
                backgroundPosition: "center",
                backgroundRepeat: "no-repeat",
                backgroundSize: "200px auto",
                backgroundBlendMode: "luminosity",
                filter: "grayscale(80%) brightness(140%) contrast(90%)",
                opacity: 0.5,
              }}>
              </div>
            )
          }
          <ChatBox messages={this.state.messages} sendMessage={(text) => {this.sendMessage(text);}} account={this.props.account} />
        </div>
      </div>
    );
  }

  fetch = (params = {}) => {
    let field = params.searchedColumn, value = params.searchText;
    const sortField = params.sortField, sortOrder = params.sortOrder;
    if (params.category !== undefined && params.category !== null) {
      field = "category";
      value = params.category;
    } else if (params.type !== undefined && params.type !== null) {
      field = "type";
      value = params.type;
    }
    this.setState({loading: true});
    ChatBackend.getChats("admin", params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            loading: false,
            data: res.data,
            pagination: {
              ...params.pagination,
              total: res.data2,
            },
            searchText: params.searchText,
            searchedColumn: params.searchedColumn,
          });

          const chats = res.data;
          if (this.state.chatName === undefined && chats.length > 0) {
            const chat = chats[0];
            this.getMessages(chat.name);
            this.setState({
              chatName: chat.name,
            });
          }
        } else {
          if (Setting.isResponseDenied(res)) {
            this.setState({
              loading: false,
              isAuthorized: false,
            });
          }
        }
      });
  };
}

export default ChatPage;
