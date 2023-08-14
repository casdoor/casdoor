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
import {Card, Col, Row, Statistic} from "antd";
import * as ApplicationBackend from "../backend/ApplicationBackend";
import * as DashboardBackend from "../backend/DashboardBackend";
import * as Setting from "../Setting";
import SingleCard from "./SingleCard";
import i18next from "i18next";
import {ArrowUpOutlined} from "@ant-design/icons";
import * as echarts from "echarts";

class HomePage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      applications: null,
      dashboardData: {},
    };
  }

  UNSAFE_componentWillMount() {
    this.getApplicationsByOrganization(this.props.account.owner);
    this.getDashboard();
  }

  getApplicationsByOrganization(organizationName) {
    ApplicationBackend.getApplicationsByOrganization("admin", organizationName)
      .then((res) => {
        this.setState({
          applications: res.data || [],
        });
      });
  }

  getDashboard() {
    DashboardBackend.getDashboard()
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            dashboardData: res.data,
          }, () => {
            this.renderEChart();
          });
        } else {
          Setting.showMessage("error", res.msg);
        }
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

      for (let i = 0; i < items.length; i++) {
        let filename = items[i].link;
        if (filename === "/account") {
          filename = "/users";
        }
        items[i].logo = `${Setting.StaticBaseUrl}/img${filename}.png`;
        items[i].createdTime = "";
      }
    } else {
      this.state.applications.forEach(application => {
        let homepageUrl = application.homepageUrl;
        if (homepageUrl === "<custom-url>") {
          homepageUrl = this.props.account.homepage;
        }

        items.push({
          link: homepageUrl, name: application.displayName, organizer: application.description, logo: application.logo, createdTime: "",
        });
      });
    }

    return items;
  }

  renderEChart() {
    const chartDom = document.getElementById("echarts-chart");
    const myChart = echarts.init(chartDom);
    const currentDate = new Date();
    const dateArray = [];
    for (let i = 6; i >= 0; i--) {
      const date = new Date(currentDate);
      date.setDate(date.getDate() - i);
      const month = parseInt(date.getMonth()) + 1;
      const day = parseInt(date.getDate());
      const formattedDate = `${month}-${day}`;
      dateArray.push(formattedDate);
    }
    const option = {
      title: {text: "Past Seven Days"},
      tooltip: {trigger: "axis"},
      legend: {data: ["User", "Provider", "Application", "Organization", "Subscription"]},
      grid: {left: "3%", right: "4%", bottom: "3%", containLabel: true},
      xAxis: {type: "category", boundaryGap: false, data: dateArray},
      yAxis: {type: "value"},
      series: [
        {name: "User", type: "line", data: this.state.dashboardData.usersCount},
        {name: "Provider", type: "line", data: this.state.dashboardData.providersCount},
        {name: "Application", type: "line", data: this.state.dashboardData.applicationsCount},
        {name: "Organization", type: "line", data: this.state.dashboardData.organizationsCount},
        {name: "Subscription", type: "line", data: this.state.dashboardData.subscriptionsCount},
      ],
    };
    myChart.setOption(option);
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
                <SingleCard key={item.link} logo={item.logo} link={item.link} title={item.name} desc={item.organizer} isSingle={items.length === 1} />
              );
            })
          }
        </Card>
      );
    } else {
      return (
        <Row gutter={80}>
          <Col span={50}>
            <Card bordered={false} bodyStyle={{width: "100%", height: "150px", display: "flex", alignItems: "center", justifyContent: "center"}}>
              <Statistic title="Total users" fontSize="100px" value={this.state.dashboardData.TotalUsersCount} valueStyle={{fontSize: "30px"}} style={{width: "200px", paddingLeft: "10px"}} />
            </Card>
          </Col>
          <Col span={50}>
            <Card bordered={false} bodyStyle={{width: "100%", height: "150px", display: "flex", alignItems: "center", justifyContent: "center"}}>
              <Statistic title="New users today" fontSize="100px" value={this.state.dashboardData.TodayNewUsersCount} valueStyle={{fontSize: "30px"}} prefix={<ArrowUpOutlined />} style={{width: "200px", paddingLeft: "10px"}} />
            </Card>
          </Col>
          <Col span={50}>
            <Card bordered={false} bodyStyle={{width: "100%", height: "150px", display: "flex", alignItems: "center", justifyContent: "center"}}>
              <Statistic title="New users past 7 days" value={this.state.dashboardData.PastSevenDaysNewUsersCount} valueStyle={{fontSize: "30px"}} prefix={<ArrowUpOutlined />} style={{width: "200px", paddingLeft: "10px"}} />
            </Card>
          </Col>
        </Row>
      );
    }
  }

  render() {
    return (
      <div style={{display: "flex", justifyContent: "center", flexDirection: "column", alignItems: "center"}}>
        <Row style={{width: "100%"}}>
          <Col span={24} style={{display: "flex", justifyContent: "center"}} >
            {
              this.renderCards()
            }
          </Col>
        </Row>
        <div id="echarts-chart"
          style={{width: "80%", height: "400px", textAlign: "center", marginTop: "20px"}}></div>
      </div>
    );
  }
}

export default HomePage;
