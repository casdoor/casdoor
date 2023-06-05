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

import {Tree, message} from "antd";
import React from "react";
import * as GroupBackend from "./backend/GroupBackend";
import * as UserBackend from "./backend/UserBackend";

function convertToTreeData(groups, parentGroupId) {
  const treeData = [];

  for (const group of groups) {
    if (group.parentGroupId === parentGroupId) {
      const node = {
        title: group.displayName,
        key: group.id,
      };
      const children = convertToTreeData(groups, group.id);
      if (children.length > 0) {
        node.children = children;
      }
      treeData.push(node);
    }
  }
  return treeData;
}

class GroupTreePage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      organizationName: props.organizationName !== undefined ? props.organizationName : props.match.params.organizationName,
      treeData: [],
    };
  }

  UNSAFE_componentWillMount() {
    this.getTreeData();
    // this.getUsers();

  }

  getTreeData() {
    GroupBackend.getGroups(this.state.organizationName).then((res) => {
      if (res.status === "ok") {
        const treeData = [
          {
            title: this.state.organizationName,
            key: "0",
            children: convertToTreeData(res.data, this.state.organizationName),
          },
        ];
        // eslint-disable-next-line no-console
        console.log(treeData);
        this.setState({
          treeData: treeData,
        });
      } else {
        message.error(res.msg);
      }
    });
  }

  getUsers(id) {
    UserBackend.getGroupUsers(id)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            users: res.data,
          });
        }
      });
  }

  render() {
    const onSelect = (selectedKeys, info) => {
      // eslint-disable-next-line no-console
      console.log("selected", selectedKeys, info);
    };
    const onCheck = (checkedKeys, info) => {
      // eslint-disable-next-line no-console
      console.log("onCheck", checkedKeys, info);
    };

    if (this.state.treeData.length === 0) {
      return null;
    }
    return (
      <Tree
        checkable
        defaultExpandedKeys={["0"]}
        defaultSelectedKeys={["0"]}
        onSelect={onSelect}
        onCheck={onCheck}
        treeData={this.state.treeData}
      />
    );
  }
}

export default GroupTreePage;
