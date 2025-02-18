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
    window.addEventListener("storageOrganizationChanged", handleOrganizationChange);
    return () => window.removeEventListener("storageOrganizationChanged", handleOrganizationChange);
  }, [props.owner]);

  React.useEffect(() => {
    if (!Setting.isLocalAdminUser(props.account)) {
      props.history.push("/apps");
    }
  }, [props.account]);

  const getOrganizationName = () => {
    let organization = localStorage.getItem("organization") === "All" ? "" : localStorage.getItem("organization");
    if (!Setting.isAdminUser(props.account) && Setting.isLocalAdminUser(props.account)) {
      organization = props.account.owner;
    }
    return organization;
  };

  React.useEffect(() => {
    if (!Setting.isLocalAdminUser(props.account)) {
      return;
    }

    const organization = getOrganizationName();
    DashboardBackend.getDashboard(organization).then((res) => {
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

  const handleOrganizationChange = () => {
    if (!Setting.isLocalAdminUser(props.account)) {
      return;
    }

    const organization = getOrganizationName();
    DashboardBackend.getDashboard(organization).then((res) => {
      if (res.status === "ok") {
        setDashboardData(res.data);
      } else {
        Setting.showMessage("error", res.msg);
      }
    });
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
        i18next.t("general:Roles"),
        i18next.t("general:Groups"),
        i18next.t("general:Resources"),
        i18next.t("general:Certs"),
        i18next.t("general:Permissions"),
        i18next.t("general:Transactions"),
        i18next.t("general:Models"),
        i18next.t("general:Adapters"),
        i18next.t("general:Enforcers"),
      ], top: "10%"},
      grid: {left: "3%", right: "4%", bottom: "0", top: "25%", containLabel: true},
      xAxis: {type: "category", boundaryGap: false, data: dateArray},
      yAxis: {type: "value"},
      series: [
        {name: i18next.t("general:Organizations"), type: "line", data: dashboardData.organizationCounts},
        {name: i18next.t("general:Users"), type: "line", data: dashboardData.userCounts},
        {name: i18next.t("general:Providers"), type: "line", data: dashboardData.providerCounts},
        {name: i18next.t("general:Applications"), type: "line", data: dashboardData.applicationCounts},
        {name: i18next.t("general:Subscriptions"), type: "line", data: dashboardData.subscriptionCounts},
        {name: i18next.t("general:Roles"), type: "line", data: dashboardData.roleCounts},
        {name: i18next.t("general:Groups"), type: "line", data: dashboardData.groupCounts},
        {name: i18next.t("general:Resources"), type: "line", data: dashboardData.resourceCounts},
        {name: i18next.t("general:Certs"), type: "line", data: dashboardData.certCounts},
        {name: i18next.t("general:Permissions"), type: "line", data: dashboardData.permissionCounts},
        {name: i18next.t("general:Transactions"), type: "line", data: dashboardData.transactionCounts},
        {name: i18next.t("general:Models"), type: "line", data: dashboardData.modelCounts},
        {name: i18next.t("general:Adapters"), type: "line", data: dashboardData.adapterCounts},
        {name: i18next.t("general:Enforcers"), type: "line", data: dashboardData.enforcerCounts},
      ],
    };
    myChart.setOption(option);

    const cardStyles = {
      body: {
        width: Setting.isMobile() ? "340px" : "100%",
        height: Setting.isMobile() ? "100px" : "150px",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
      },
    };

    return (
      <Row id="statistic" gutter={80} justify={"center"}>
        <Col span={50} style={{marginBottom: "10px"}}>
          <Card variant="borderless" styles={cardStyles}>
            <Statistic title={i18next.t("home:Total users")} fontSize="100px" value={dashboardData.userCounts[30]} valueStyle={{fontSize: "30px"}} style={{width: "200px", paddingLeft: "10px"}} />
          </Card>
        </Col>
        <Col span={50} style={{marginBottom: "10px"}}>
          <Card variant="borderless" styles={cardStyles}>
            <Statistic title={i18next.t("home:New users today")} fontSize="100px" value={dashboardData.userCounts[30] - dashboardData.userCounts[30 - 1]} valueStyle={{fontSize: "30px"}} prefix={<ArrowUpOutlined />} style={{width: "200px", paddingLeft: "10px"}} />
          </Card>
        </Col>
        <Col span={50} style={{marginBottom: "10px"}}>
          <Card variant="borderless" styles={cardStyles}>
            <Statistic title={i18next.t("home:New users past 7 days")} value={dashboardData.userCounts[30] - dashboardData.userCounts[30 - 7]} valueStyle={{fontSize: "30px"}} prefix={<ArrowUpOutlined />} style={{width: "200px", paddingLeft: "10px"}} />
          </Card>
        </Col>
        <Col span={50} style={{marginBottom: "10px"}}>
          <Card variant="borderless" styles={cardStyles}>
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
        open={Setting.isMobile() ? false : isTourVisible}
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
