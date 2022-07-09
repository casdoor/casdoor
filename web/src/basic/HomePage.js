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
import {Card, Col, Row} from "antd";
import * as ApplicationBackend from "../backend/ApplicationBackend";
import * as Setting from "../Setting";
import SingleCard from "./SingleCard";
import i18next from "i18next";

class HomePage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      applications: null,
    };
  }

  UNSAFE_componentWillMount() {
    this.getApplicationsByOrganization(this.props.account.owner);
  }

  getApplicationsByOrganization(organizationName) {
    ApplicationBackend.getApplicationsByOrganization("admin", organizationName)
      .then((res) => {
        this.setState({
          applications: (res.msg === undefined) ? res : [],
        });
      });
  }

  getItems() {
    let items = [];
    if (Setting.isAdminUser(this.props.account)) {
      items = [
        {link: "/organizations", name: i18next.t("general:Organizations"), organizer: i18next.t("general:User containers")},
        {link: "/users", name: i18next.t("general:Users"), organizer: i18next.t("general:Users under all organizations")},
        {link: "/providers", name: i18next.t("general:Providers"), organizer: i18next.t("general:OAuth providers")},
        {link: "/applications", name: i18next.t("general:Applications"), organizer: i18next.t("general:Applications that require authentication")},
      ];

      for (let i = 0; i < items.length; i ++) {
        let filename = items[i].link;
        if (filename === "/account") {
          filename = "/users";
        }
        items[i].logo = `https://cdn.casbin.com/static/img${filename}.png`;
        items[i].createdTime = "";
      }
    } else {
      this.state.applications.forEach(application => {
        console.log(application);
        items.push({
          link: application.homepageUrl, name: application.displayName, organizer: application.description, logo: application.logo, createdTime: "",
        });
      });
    }

    return items;
  }

  renderCards() {
    if (this.state.applications === null) {
      return null;
    }

    const items = this.getItems();

    if (Setting.isMobile()) {
      return (
        <Card bodyStyle={{padding: 0}}>
          {
            items.map(item => {
              return (
                <SingleCard logo={item.logo} link={item.link} title={item.name} desc={item.organizer} isSingle={items.length === 1} />
              );
            })
          }
        </Card>
      );
    } else {
      return (
        <div style={{marginRight: "15px", marginLeft: "15px"}}>
          <Row style={{marginLeft: "-20px", marginRight: "-20px", marginTop: "20px"}} gutter={24}>
            {
              items.map(item => {
                return (
                  <SingleCard logo={item.logo} link={item.link} title={item.name} desc={item.organizer} time={item.createdTime} isSingle={items.length === 1} key={item.name} />
                );
              })
            }
          </Row>
        </div>
      );
    }
  }

  render() {
    return (
      <div>
        <Row style={{width: "100%"}}>
          <Col span={24} style={{display: "flex", justifyContent:  "center"}} >
            {
              this.renderCards()
            }
          </Col>
        </Row>
      </div>
    );
  }
}

export default HomePage;
