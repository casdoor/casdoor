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

import React, {createRef} from "react";
import {Card, Col, Row, Spin, Statistic, Tour} from "antd";
import {ArrowUpOutlined} from "@ant-design/icons";
import * as ApplicationBackend from "../backend/ApplicationBackend";
import * as DashboardBackend from "../backend/DashboardBackend";
import * as echarts from "echarts";
import * as Setting from "../Setting";
import SingleCard from "./SingleCard";
import i18next from "i18next";

class HomePage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      applications: null,
      dashboardData: null,
      isTourVisible: Setting.getTourVisible(),
    };
    this.ref1 = createRef();
    this.ref2 = createRef();
    this.steps = [
      {
        title: "Welcome to casdoor",
        description: "You can learn more about the use of CasDoor at https://casdoor.org/.",
        cover: (
          <img
            alt="casdoor.png"
            src="https://cdn.casbin.org/img/casdoor-logo_1185x256.png"
          />
        ),
        target: null,
      },
      {
        title: "Statistic cards",
        description: "Here are four statistic cards for user information.",
        target: () => this.ref1,
      },
      {
        title: "Import users",
        description: "You can add new users or update existing Casdoor users by uploading a XLSX file of user information.",
        target: () => this.ref2,
        nextButtonProps: {
          children: "Go to \"Organizations list\"",
        },
      },
    ];
  }

  UNSAFE_componentWillMount() {
    this.getApplicationsByOrganization(this.props.account.owner);
    this.getDashboard();
  }

  componentDidMount() {
    window.addEventListener("storageTourChanged", this.handleTourChange);
  }

  componentWillUnmount() {
    window.removeEventListener("storageTourChanged", this.handleTourChange);
  }

  handleTourChange = () => {
    this.setState({isTourVisible: Setting.getTourVisible()});
  };

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

  setIsTourVisible = () => {
    Setting.setIsTourVisible(false);
    this.setState({isTourVisible: false});
  };

  handleTourComplete = () => {
    this.props.history.push("/syncers");
  };

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
    const data = this.state.dashboardData;

    const chartDom = document.getElementById("echarts-chart");
    const myChart = echarts.init(chartDom);
    const currentDate = new Date();
    const dateArray = [];
    for (let i = 30; i >= 0; i--) {
      const date = new Date(currentDate);
      date.setDate(date.getDate() - i);
      const month = parseInt(date.getMonth()) + 1;
      const day = parseInt(date.getDate());
      const formattedDate = `${month}-${day}`;
      dateArray.push(formattedDate);
    }
    const option = {
      title: {text: i18next.t("home:Past 30 Days")},
      tooltip: {trigger: "axis"},
      legend: {data: [
        i18next.t("general:Users"),
        i18next.t("general:Providers"),
        i18next.t("general:Applications"),
        i18next.t("general:Organizations"),
        i18next.t("general:Subscriptions"),
      ]},
      grid: {left: "3%", right: "4%", bottom: "3%", containLabel: true},
      xAxis: {type: "category", boundaryGap: false, data: dateArray},
      yAxis: {type: "value"},
      series: [
        {name: i18next.t("general:Organizations"), type: "line", data: data?.organizationCounts},
        {name: i18next.t("general:Users"), type: "line", data: data?.userCounts},
        {name: i18next.t("general:Providers"), type: "line", data: data?.providerCounts},
        {name: i18next.t("general:Applications"), type: "line", data: data?.applicationCounts},
        {name: i18next.t("general:Subscriptions"), type: "line", data: data?.subscriptionCounts},
      ],
    };
    myChart.setOption(option);
  }

  renderCards() {
    const data = this.state.dashboardData;
    if (data === null) {
      return (
        <div style={{display: "flex", justifyContent: "center", alignItems: "center", marginTop: "10%"}}>
          <Spin size="large" tip={i18next.t("login:Loading")} style={{paddingTop: "10%"}} />
        </div>
      );
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
              <Statistic title={i18next.t("home:Total users")} fontSize="100px" value={data?.userCounts[30]} valueStyle={{fontSize: "30px"}} style={{width: "200px", paddingLeft: "10px"}} />
            </Card>
          </Col>
          <Col span={50}>
            <Card bordered={false} bodyStyle={{width: "100%", height: "150px", display: "flex", alignItems: "center", justifyContent: "center"}}>
              <Statistic title={i18next.t("home:New users today")} fontSize="100px" value={data?.userCounts[30] - data?.userCounts[30 - 1]} valueStyle={{fontSize: "30px"}} prefix={<ArrowUpOutlined />} style={{width: "200px", paddingLeft: "10px"}} />
            </Card>
          </Col>
          <Col span={50}>
            <Card bordered={false} bodyStyle={{width: "100%", height: "150px", display: "flex", alignItems: "center", justifyContent: "center"}}>
              <Statistic title={i18next.t("home:New users past 7 days")} value={data?.userCounts[30] - data?.userCounts[30 - 7]} valueStyle={{fontSize: "30px"}} prefix={<ArrowUpOutlined />} style={{width: "200px", paddingLeft: "10px"}} />
            </Card>
          </Col>
          <Col span={50}>
            <Card bordered={false} bodyStyle={{width: "100%", height: "150px", display: "flex", alignItems: "center", justifyContent: "center"}}>
              <Statistic title={i18next.t("home:New users past 30 days")} value={data?.userCounts[30] - data?.userCounts[30 - 30]} valueStyle={{fontSize: "30px"}} prefix={<ArrowUpOutlined />} style={{width: "200px", paddingLeft: "10px"}} />
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
          <Col span={24} style={{display: "flex", justifyContent: "center"}} ref={ref => this.ref1 = ref}>
            {
              this.renderCards()
            }
          </Col>
        </Row>
        <div id="echarts-chart"
          style={{width: "80%", height: "400px", textAlign: "center", marginTop: "20px"}} ref={ref => this.ref2 = ref}></div>
        <Tour
          open={this.state.isTourVisible}
          onClose={this.setIsTourVisible}
          steps={this.steps}
          indicatorsRender={(current, total) => (
            <span>
              {current + 1} / {total}
            </span>
          )}
          onFinish={this.handleTourComplete}
        />
      </div>
    );
  }
}

export default HomePage;
