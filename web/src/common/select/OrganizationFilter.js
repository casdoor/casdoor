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
import * as Setting from "../../Setting";
import {Dropdown, Space} from "antd";
import "../../App.less";
import {ApartmentOutlined, CheckOutlined} from "@ant-design/icons";
import * as OrganizationBackend from "../../backend/OrganizationBackend";
import * as Conf from "../../Conf";
import i18next from "i18next";

class OrganizationFilter extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      organizations: [],
      organization: Setting.getOrganization(),
    };
  }

  getOrganizations() {
    OrganizationBackend.getOrganizationNames("admin")
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            organizations: res.data.map((organization) => organization.name),
          });
        }
      });
  }

  renderItem(organizationName, key) {
    return Setting.getItem(
      <Space>
        {organizationName}
        {key === Setting.getOrganization() ? <CheckOutlined style={{marginLeft: "5px"}} /> : null}
      </Space>,
      key
    );
  }

  getOrganizationItems() {
    const items = [];

    items.push(this.renderItem(i18next.t(`organization:${Conf.DefaultOrganization}`), Conf.DefaultOrganization));

    this.state.organizations.forEach((organization) => items.push(this.renderItem(organization, organization)));
    return items;
  }

  render() {
    if (!Setting.isAdminUser(this.props.account)) {
      return null;
    }

    const onClick = (e) => {
      Setting.setOrganization(e.key);
      this.setState({
        organization: e.key,
      });
    };

    if (this.state.organizations.length === 0) {
      this.getOrganizations();
    }

    return (
      <Dropdown menu={{
        items: this.getOrganizationItems(),
        onClick,
        selectable: true,
        multiple: false,
        selectedKeys: [Setting.getOrganization()],
      }}>
        <div className="select-box">
          <ApartmentOutlined style={{fontSize: "24px", color: "#4d4d4d"}} />
        </div>
      </Dropdown>
    );
  }
}

export default OrganizationFilter;
