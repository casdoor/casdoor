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

import {DeleteOutlined, EditOutlined, HolderOutlined, PlusOutlined, UsergroupAddOutlined} from "@ant-design/icons";
import {Button, Col, Empty, Row, Space, Tree} from "antd";
import i18next from "i18next";
import moment from "moment/moment";
import React from "react";
import * as GroupBackend from "./backend/GroupBackend";
import * as Setting from "./Setting";
import OrganizationSelect from "./common/select/OrganizationSelect";
import UserListPage from "./UserListPage";

class GroupTreePage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      owner: Setting.isAdminUser(this.props.account) ? "" : this.props.account.owner,
      organizationName: props.organizationName !== undefined ? props.organizationName : props.match.params.organizationName,
      groupName: this.props.match?.params.groupName,
      groupId: "",
      treeData: [],
      selectedKeys: [this.props.match?.params.groupName],
    };
  }

  UNSAFE_componentWillMount() {
    this.getTreeData();
  }

  componentDidUpdate(prevProps, prevState, snapshot) {
    if (this.state.organizationName !== prevState.organizationName) {
      this.getTreeData();
    }

    if (prevState.treeData !== this.state.treeData) {
      this.setTreeExpandedKeys();
    }
  }

  getTreeData() {
    GroupBackend.getGroups(this.state.organizationName, true).then((res) => {
      if (res.status === "ok") {
        this.setState({
          treeData: res.data,
        });
      } else {
        Setting.showMessage("error", res.msg);
      }
    });
  }

  setTreeTitle(treeData) {
    const haveChildren = Array.isArray(treeData.children) && treeData.children.length > 0;
    const isSelected = this.state.groupName === treeData.key;
    return {
      id: treeData.id,
      key: treeData.key,
      title: <Space>
        {treeData.type === "Physical" ? <UsergroupAddOutlined /> : <HolderOutlined />}
        <span>{treeData.title}</span>
        {isSelected && (
          <React.Fragment>
            <PlusOutlined
              style={{
                visibility: "visible",
                color: "inherit",
                transition: "color 0.3s",
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.color = "rgba(89,54,213,0.6)";
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.color = "inherit";
              }}
              onMouseDown={(e) => {
                e.currentTarget.style.color = "rgba(89,54,213,0.4)";
              }}
              onMouseUp={(e) => {
                e.currentTarget.style.color = "rgba(89,54,213,0.6)";
              }}
              onClick={(e) => {
                e.stopPropagation();
                sessionStorage.setItem("groupTreeUrl", window.location.pathname);
                this.addGroup();
              }}
            />
            <EditOutlined
              style={{
                visibility: "visible",
                color: "inherit",
                transition: "color 0.3s",
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.color = "rgba(89,54,213,0.6)";
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.color = "inherit";
              }}
              onMouseDown={(e) => {
                e.currentTarget.style.color = "rgba(89,54,213,0.4)";
              }}
              onMouseUp={(e) => {
                e.currentTarget.style.color = "rgba(89,54,213,0.6)";
              }}
              onClick={(e) => {
                e.stopPropagation();
                sessionStorage.setItem("groupTreeUrl", window.location.pathname);
                this.props.history.push(`/groups/${this.state.organizationName}/${treeData.key}`);
              }}
            />
            {!haveChildren &&
            <DeleteOutlined
              style={{
                visibility: "visible",
                color: "inherit",
                transition: "color 0.3s",
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.color = "rgba(89,54,213,0.6)";
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.color = "inherit";
              }}
              onMouseDown={(e) => {
                e.currentTarget.style.color = "rgba(89,54,213,0.4)";
              }}
              onMouseUp={(e) => {
                e.currentTarget.style.color = "rgba(89,54,213,0.6)";
              }}
              onClick={(e) => {
                e.stopPropagation();
                GroupBackend.deleteGroup({owner: treeData.owner, name: treeData.key})
                  .then((res) => {
                    if (res.status === "ok") {
                      Setting.showMessage("success", i18next.t("general:Successfully deleted"));
                      this.getTreeData();
                    } else {
                      Setting.showMessage("error", `${i18next.t("general:Failed to delete")}: ${res.msg}`);
                    }
                  })
                  .catch(error => {
                    Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
                  });
              }}
            />
            }
          </React.Fragment>
        )}
      </Space>,
      children: haveChildren ? treeData.children.map(i => this.setTreeTitle(i)) : [],
    };
  }

  setTreeExpandedKeys = () => {
    const expandedKeys = [];
    const setExpandedKeys = (nodes) => {
      for (const node of nodes) {
        expandedKeys.push(node.key);
        if (node.children) {
          setExpandedKeys(node.children);
        }
      }
    };
    setExpandedKeys(this.state.treeData);
    this.setState({
      expandedKeys: expandedKeys,
    });
  };

  renderTree() {
    const onSelect = (selectedKeys, info) => {
      this.setState({
        selectedKeys: selectedKeys,
        groupName: info.node.key,
        groupId: info.node.id,
      });
      this.props.history.push(`/trees/${this.state.organizationName}/${info.node.key}`);
    };
    const onExpand = (expandedKeysValue) => {
      this.setState({
        expandedKeys: expandedKeysValue,
      });
    };

    if (this.state.treeData.length === 0) {
      return <Empty />;
    }

    const treeData = this.state.treeData.map(i => this.setTreeTitle(i));
    return (
      <Tree
        blockNode={true}
        defaultSelectedKeys={[this.state.groupName]}
        defaultExpandAll={true}
        selectedKeys={this.state.selectedKeys}
        expandedKeys={this.state.expandedKeys}
        onSelect={onSelect}
        onExpand={onExpand}
        showIcon={true}
        treeData={treeData}
      />
    );
  }

  renderOrganizationSelect() {
    if (Setting.isAdminUser()) {
      return (
        <OrganizationSelect
          initValue={this.state.organizationName}
          style={{width: "100%"}}
          onChange={(value) => {
            this.setState({
              organizationName: value,
            });
            this.props.history.push(`/trees/${value}`);
          }}
        />
      );
    }
  }

  newGroup(isRoot) {
    const randomName = Setting.getRandomName();
    return {
      owner: this.state.organizationName,
      name: `group_${randomName}`,
      createdTime: moment().format(),
      updatedTime: moment().format(),
      displayName: `New Group - ${randomName}`,
      type: "Virtual",
      parentGroupId: isRoot ? this.state.organizationName : this.state.groupId,
      isTopGroup: isRoot,
      isEnabled: true,
    };
  }

  addGroup(isRoot = false) {
    const newGroup = this.newGroup(isRoot);
    GroupBackend.addGroup(newGroup)
      .then((res) => {
        if (res.status === "ok") {
          sessionStorage.setItem("groupTreeUrl", window.location.pathname);
          this.props.history.push({pathname: `/groups/${newGroup.owner}/${newGroup.name}`, mode: "add"});
          Setting.showMessage("success", i18next.t("general:Successfully added"));
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to add")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  render() {
    return (
      <div style={{
        flex: 1,
        backgroundColor: "white",
        padding: "5px 5px 2px 5px",
      }}>
        <Row>
          <Col span={5}>
            <Row>
              <Col span={24} style={{textAlign: "left"}}>
                {this.renderOrganizationSelect()}
              </Col>
            </Row>
            <Row>
              <Col span={24} style={{marginTop: "10px", textAlign: "left"}}>
                <Button size={"small"}
                  onClick={() => {
                    this.setState({
                      selectedKeys: [],
                      groupName: null,
                      groupId: null,
                    });
                    this.props.history.push(`/trees/${this.state.organizationName}`);
                  }}
                >
                  {i18next.t("group:Show all")}
                </Button>
                <Button size={"small"} type={"primary"} style={{marginLeft: "10px"}} onClick={() => this.addGroup(true)}>
                  {i18next.t("general:Add")}
                </Button>
              </Col>
            </Row>
            <Row style={{marginTop: 10}}>
              <Col span={24} style={{textAlign: "left"}}>
                {this.renderTree()}
              </Col>
            </Row>
          </Col>
          <Col span={19}>
            <UserListPage
              organizationName={this.state.organizationName}
              groupName={this.state.groupName}
              groupId={this.state.groupId}
              {...this.props}
            />
          </Col>
        </Row>
      </div>
    );
  }
}

export default GroupTreePage;
