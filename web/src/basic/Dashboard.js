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
import {Card, Col, Row, Spin, Statistic, Tour} from "antd";
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

    setDashboardData(null);

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

  const getDateArray = () => {
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
    return dateArray;
  };

  const ChartWidget = ({chartId, title, seriesConfig}) => {
    React.useEffect(() => {
      if (dashboardData === null) {
        return;
      }

      const chartDom = document.getElementById(chartId);
      if (!chartDom) {
        return;
      }

      const myChart = echarts.init(chartDom);
      const dateArray = getDateArray();

      const option = {
        title: {text: title, left: "center"},
        tooltip: {trigger: "axis"},
        legend: {
          data: seriesConfig.map(s => s.name),
          top: "10%",
        },
        grid: {left: "3%", right: "4%", bottom: "3%", top: "20%", containLabel: true},
        xAxis: {type: "category", boundaryGap: false, data: dateArray},
        yAxis: {type: "value"},
        series: seriesConfig.map(s => ({
          name: s.name,
          type: "line",
          data: dashboardData[s.dataKey],
          smooth: true,
        })),
      };
      myChart.setOption(option);

      return () => {
        myChart.dispose();
      };
    }, [dashboardData, chartId, title, seriesConfig]);

    return <div id={chartId} style={{width: "100%", height: "350px"}} />;
  };

  const renderStatistics = () => {
    if (dashboardData === null) {
      return (
        <div style={{display: "flex", justifyContent: "center", alignItems: "center", padding: "40px"}}>
          <Spin size="large" tip={i18next.t("login:Loading")} />
        </div>
      );
    }

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
      <Row id="statistic" gutter={16} justify={"center"} style={{marginBottom: "20px"}}>
        <Col xs={24} sm={12} lg={6} style={{marginBottom: "10px"}}>
          <Card variant="borderless" styles={cardStyles}>
            <Statistic title={i18next.t("home:Total users")} fontSize="100px" value={dashboardData.userCounts[30]} valueStyle={{fontSize: "30px"}} style={{width: "200px", paddingLeft: "10px"}} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6} style={{marginBottom: "10px"}}>
          <Card variant="borderless" styles={cardStyles}>
            <Statistic title={i18next.t("home:New users today")} fontSize="100px" value={dashboardData.userCounts[30] - dashboardData.userCounts[30 - 1]} valueStyle={{fontSize: "30px"}} prefix={<ArrowUpOutlined />} style={{width: "200px", paddingLeft: "10px"}} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6} style={{marginBottom: "10px"}}>
          <Card variant="borderless" styles={cardStyles}>
            <Statistic title={i18next.t("home:New users past 7 days")} value={dashboardData.userCounts[30] - dashboardData.userCounts[30 - 7]} valueStyle={{fontSize: "30px"}} prefix={<ArrowUpOutlined />} style={{width: "200px", paddingLeft: "10px"}} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6} style={{marginBottom: "10px"}}>
          <Card variant="borderless" styles={cardStyles}>
            <Statistic title={i18next.t("home:New users past 30 days")} value={dashboardData.userCounts[30] - dashboardData.userCounts[30 - 30]} valueStyle={{fontSize: "30px"}} prefix={<ArrowUpOutlined />} style={{width: "200px", paddingLeft: "10px"}} />
          </Card>
        </Col>
      </Row>
    );
  };

  return (
    <div style={{padding: "20px", maxWidth: "1400px", margin: "0 auto"}}>
      {renderStatistics()}

      <Row gutter={[16, 16]}>
        <Col xs={24} lg={12}>
          <Card title={i18next.t("home:User & Organization Metrics")} style={{height: "100%"}}>
            <ChartWidget chartId="users-orgs-chart" title={i18next.t("home:Past 30 Days")} seriesConfig={[
              {name: i18next.t("general:Users"), dataKey: "userCounts"},
              {name: i18next.t("general:Organizations"), dataKey: "organizationCounts"},
            ]} />
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <Card title={i18next.t("home:Authentication & Applications")} style={{height: "100%"}}>
            <ChartWidget chartId="auth-apps-chart" title={i18next.t("home:Past 30 Days")} seriesConfig={[
              {name: i18next.t("general:Providers"), dataKey: "providerCounts"},
              {name: i18next.t("general:Applications"), dataKey: "applicationCounts"},
            ]} />
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <Card title={i18next.t("home:Access Control")} style={{height: "100%"}}>
            <ChartWidget chartId="access-control-chart" title={i18next.t("home:Past 30 Days")} seriesConfig={[
              {name: i18next.t("general:Roles"), dataKey: "roleCounts"},
              {name: i18next.t("general:Groups"), dataKey: "groupCounts"},
              {name: i18next.t("general:Permissions"), dataKey: "permissionCounts"},
              {name: i18next.t("general:Enforcers"), dataKey: "enforcerCounts"},
            ]} />
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <Card title={i18next.t("home:Resources & Security")} style={{height: "100%"}}>
            <ChartWidget chartId="resources-security-chart" title={i18next.t("home:Past 30 Days")} seriesConfig={[
              {name: i18next.t("general:Resources"), dataKey: "resourceCounts"},
              {name: i18next.t("general:Certs"), dataKey: "certCounts"},
              {name: i18next.t("general:Subscriptions"), dataKey: "subscriptionCounts"},
            ]} />
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <Card title={i18next.t("home:Transactions & Payments")} style={{height: "100%"}}>
            <ChartWidget chartId="transactions-chart" title={i18next.t("home:Past 30 Days")} seriesConfig={[
              {name: i18next.t("general:Transactions"), dataKey: "transactionCounts"},
            ]} />
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <Card title={i18next.t("home:System & Integration")} style={{height: "100%"}}>
            <ChartWidget chartId="system-integration-chart" title={i18next.t("home:Past 30 Days")} seriesConfig={[
              {name: i18next.t("general:Models"), dataKey: "modelCounts"},
              {name: i18next.t("general:Adapters"), dataKey: "adapterCounts"},
            ]} />
          </Card>
        </Col>
      </Row>

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
