// Copyright 2021 The casbin Authors. All Rights Reserved.
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
import * as Setting from "../Setting";
import SingleCard from "./SingleCard";

class HomePage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
    };
  }

  getItems() {
    let items = [];
    if (Setting.isAdminUser(this.props.account)) {
      items = [
        {link: "/organizations", name: "Organizations", organizer: "User containers"},
        {link: "/users", name: "Users", organizer: "Users under all organizations"},
        {link: "/providers", name: "Providers", organizer: "OAuth providers"},
        {link: "/applications", name: "Applications", organizer: "Applications that requires authentication"},
      ];
    }

    for (let i = 0; i < items.length; i ++) {
      items[i].logo = `https://cdn.casbin.com/static/img${items[i].link}.png`;
      items[i].createdTime = "";
    }

    return items
  }

  renderCards() {
    const items = this.getItems();

    if (Setting.isMobile()) {
      return (
        <Card bodyStyle={{padding: 0}}>
          {
            items.map(item => {
              return (
                <SingleCard logo={item.logo} link={item.link} title={item.name} desc={item.organizer} isSingle={items.length === 1} />
              )
            })
          }
        </Card>
      )
    } else {
      return (
        <div>
          <Row>
            <Col span={1}>
            </Col>
            <Col span={22}>
              <Row style={{marginLeft: "-20px", marginRight: "-20px", marginTop: "20px"}} gutter={24}>
                {
                  items.map(item => {
                    return (
                      <SingleCard logo={item.logo} link={item.link} title={item.name} desc={item.organizer} time={item.createdTime} isSingle={items.length === 1} />
                    )
                  })
                }
              </Row>
            </Col>
            <Col span={1}>
            </Col>
          </Row>
        </div>
      )
    }
  }

  render() {
    return (
      <div>
        &nbsp;
        <Row style={{width: "100%"}}>
          <Col span={24} style={{display: "flex", justifyContent:  "center"}} >
            {
              this.renderCards()
            }
          </Col>
        </Row>
      </div>
    )
  }
}

export default HomePage;
