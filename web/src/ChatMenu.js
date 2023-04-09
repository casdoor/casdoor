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
import {Menu} from "antd";
import {MailOutlined} from "@ant-design/icons";

class ChatMenu extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      openKeys: ["0"],
    };
  }

  chatsToItems(chats) {
    const categories = {};
    chats.forEach((chat) => {
      if (!categories[chat.category]) {
        categories[chat.category] = [];
      }
      categories[chat.category].push(chat);
    });

    return Object.keys(categories).map((category, index) => {
      return {
        key: `${index}`,
        icon: <MailOutlined />,
        label: category,
        children: categories[category].map((chat) => ({
          key: chat.id,
          label: chat.displayName,
        })),
      };
    });
  }

  // 处理菜单展开事件
  onOpenChange = (keys) => {
    const rootSubmenuKeys = this.props.chats.map((_, index) => `${index}`);
    const latestOpenKey = keys.find((key) => this.state.openKeys.indexOf(key) === -1);

    if (rootSubmenuKeys.indexOf(latestOpenKey) === -1) {
      this.setState({openKeys: keys});
    } else {
      this.setState({openKeys: latestOpenKey ? [latestOpenKey] : []});
    }
  };

  render() {
    const items = this.chatsToItems(this.props.chats);

    return (
      <Menu
        mode="inline"
        openKeys={this.state.openKeys}
        onOpenChange={this.onOpenChange}
        style={{
          width: 256,
        }}
        items={items}
      />
    );
  }
}

export default ChatMenu;
