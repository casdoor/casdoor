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

import {ArrowUpOutlined} from "@ant-design/icons";
import {Card, Col, Row, Statistic, Tour} from "antd";
import * as echarts from "echarts";
import i18next from "i18next";
import React from "react";
import * as DashboardBackend from "../backend/DashboardBackend";
import * as Setting from "../Setting";
import * as TourConfig from "../TourConfig";

const Dashboard = (props) => {
  const [dashboardData, setDashboardData] = React.useState(null);
  const [isTourVisible, setIsTourVisible] = React.useState(TourConfig.getTourVisible());
  const nextPathName = TourConfig.getNextUrl("home");

  React.useEffect(() => {
    window.addEventListener("storageTourChanged", handleTourChange);
    return () => window.removeEventListener("storageTourChanged", handleTourChange);
  }, []);

  React.useEffect(() => {
    if (!Setting.isLocalAdminUser(props.account)) {
      props.history.push("/apps");
    }
  }, [props.account]);

  React.useEffect(() => {
    if (!Setting.isLocalAdminUser(props.account)) {
      return;
    }
    DashboardBackend.getDashboard(props.account.owner).then((res) => {
      if (res.status === "ok") {
        setDashboardData(res.data);
      } else {
        Setting.showMessage("error", res.msg);
      }
    });
  }, [props.owner]);

  const handleTourChange = () => {
    setIsTourVisible(TourConfig.getTourVisible());
  };

  const setIsTourToLocal = () => {
    TourConfig.setIsTourVisible(false);
    setIsTourVisible(false);
  };

  const handleTourComplete = () => {
    if (nextPathName !== "") {
      props.history.push("/" + nextPathName);
      TourConfig.setIsTourVisible(true);
    }
  };

  const getSteps = () => {
    const steps = TourConfig.TourObj["home"];
    steps.map((item, index) => {
      item.target = () => document.getElementById(item.id) || null;
      if (index === steps.length - 1) {
        item.nextButtonProps = {
          children: TourConfig.getNextButtonChild(nextPathName),
        };
      }
    });
    return steps;
  };

  const renderEChart = () => {
    if (dashboardData === null) {
      return;
    }

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
        {name: i18next.t("general:Organizations"), type: "line", data: dashboardData.organizationCounts},
        {name: i18next.t("general:Users"), type: "line", data: dashboardData.userCounts},
        {name: i18next.t("general:Providers"), type: "line", data: dashboardData.providerCounts},
        {name: i18next.t("general:Applications"), type: "line", data: dashboardData.applicationCounts},
        {name: i18next.t("general:Subscriptions"), type: "line", data: dashboardData.subscriptionCounts},
      ],
    };
    myChart.setOption(option);

    return (
      <Row id="statistic" gutter={80}>
        <Col span={50}>
          <Card bordered={false} bodyStyle={{width: "100%", height: "150px", display: "flex", alignItems: "center", justifyContent: "center"}}>
            <Statistic title={i18next.t("home:Total users")} fontSize="100px" value={dashboardData.userCounts[30]} valueStyle={{fontSize: "30px"}} style={{width: "200px", paddingLeft: "10px"}} />
          </Card>
        </Col>
        <Col span={50}>
          <Card bordered={false} bodyStyle={{width: "100%", height: "150px", display: "flex", alignItems: "center", justifyContent: "center"}}>
            <Statistic title={i18next.t("home:New users today")} fontSize="100px" value={dashboardData.userCounts[30] - dashboardData.userCounts[30 - 1]} valueStyle={{fontSize: "30px"}} prefix={<ArrowUpOutlined />} style={{width: "200px", paddingLeft: "10px"}} />
          </Card>
        </Col>
        <Col span={50}>
          <Card bordered={false} bodyStyle={{width: "100%", height: "150px", display: "flex", alignItems: "center", justifyContent: "center"}}>
            <Statistic title={i18next.t("home:New users past 7 days")} value={dashboardData.userCounts[30] - dashboardData.userCounts[30 - 7]} valueStyle={{fontSize: "30px"}} prefix={<ArrowUpOutlined />} style={{width: "200px", paddingLeft: "10px"}} />
          </Card>
        </Col>
        <Col span={50}>
          <Card bordered={false} bodyStyle={{width: "100%", height: "150px", display: "flex", alignItems: "center", justifyContent: "center"}}>
            <Statistic title={i18next.t("home:New users past 30 days")} value={dashboardData.userCounts[30] - dashboardData.userCounts[30 - 30]} valueStyle={{fontSize: "30px"}} prefix={<ArrowUpOutlined />} style={{width: "200px", paddingLeft: "10px"}} />
          </Card>
        </Col>
      </Row>
    );
  };

  return (
    <div style={{display: "flex", justifyContent: "center", flexDirection: "column", alignItems: "center"}}>
      {renderEChart()}
      <div id="echarts-chart" style={{width: "80%", height: "400px", textAlign: "center", marginTop: "20px"}} />
      <Tour
        open={isTourVisible}
        onClose={setIsTourToLocal}
        steps={getSteps()}
        indicatorsRender={(current, total) => (
          <span>
            {current + 1} / {total}
          </span>
        )}
        onFinish={handleTourComplete}
      />
    </div>
  );
};

export default Dashboard;
