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

import {ArrowDownOutlined, ArrowUpOutlined, InfoCircleOutlined} from "@ant-design/icons";
import {Card, Col, Progress, Row, Spin, Tour} from "antd";
import * as echarts from "echarts";
import i18next from "i18next";
import React from "react";
import * as DashboardBackend from "../backend/DashboardBackend";
import * as Setting from "../Setting";
import * as TourConfig from "../TourConfig";

const MiniChart = ({data, color = "#1890ff"}) => {
  const chartIdRef = React.useRef(null);
  const chartId = React.useMemo(() => {
    if (!chartIdRef.current) {
      chartIdRef.current = `mini-chart-${Date.now()}-${Math.floor(Math.random() * 10000)}`;
    }
    return chartIdRef.current;
  }, []);

  React.useEffect(() => {
    if (!data || data.length === 0) {return;}

    const chartDom = document.getElementById(chartId);
    if (!chartDom) {return;}

    const myChart = echarts.init(chartDom);
    const option = {
      grid: {left: 0, right: 0, top: 0, bottom: 0},
      xAxis: {
        type: "category",
        show: false,
        data: data.map((_, i) => i),
      },
      yAxis: {
        type: "value",
        show: false,
      },
      series: [
        {
          data: data,
          type: "line",
          smooth: true,
          symbol: "none",
          lineStyle: {width: 2, color: color},
          areaStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              {offset: 0, color: color + "40"},
              {offset: 1, color: color + "10"},
            ]),
          },
        },
      ],
    };
    myChart.setOption(option);

    return () => {
      myChart.dispose();
    };
  }, [data, color, chartId]);

  return <div id={chartId} style={{width: "100%", height: "46px"}} />;
};

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

  const renderEChart = () => {
    const chartDom = document.getElementById("echarts-chart");

    if (dashboardData === null) {
      if (chartDom) {
        const instance = echarts.getInstanceByDom(chartDom);
        if (instance) {
          instance.dispose();
        }
      }
      return (
        <div style={{display: "flex", justifyContent: "center", alignItems: "center"}}>
          <Spin size="large" tip={i18next.t("login:Loading")} style={{paddingTop: "10%"}} />
        </div>
      );
    }

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

    // Validate data array length
    if (!dashboardData.userCounts || dashboardData.userCounts.length < 31) {
      return (
        <div style={{display: "flex", justifyContent: "center", alignItems: "center", padding: "40px"}}>
          <span style={{color: "rgba(0,0,0,0.45)"}}>No data available</span>
        </div>
      );
    }

    // Calculate statistics for cards
    const totalUsers = dashboardData.userCounts[30];
    const usersToday = Math.max(0, dashboardData.userCounts[30] - dashboardData.userCounts[29]);
    const usersWeek = Math.max(0, dashboardData.userCounts[30] - dashboardData.userCounts[23]);
    const usersMonth = Math.max(0, dashboardData.userCounts[30] - dashboardData.userCounts[0]);

    // Calculate percentage changes (using last 7 days vs previous 7 days)
    const weekBeforeLast = dashboardData.userCounts[23] - dashboardData.userCounts[16];
    let weeklyGrowth = 0;
    if (weekBeforeLast > 0) {
      weeklyGrowth = ((usersWeek - weekBeforeLast) / weekBeforeLast * 100).toFixed(1);
    } else if (weekBeforeLast === 0 && usersWeek > 0) {
      weeklyGrowth = 100;
    } else if (weekBeforeLast === 0 && usersWeek < 0) {
      weeklyGrowth = -100;
    }

    // Get last 7 days data for mini charts
    const last7Days = dashboardData.userCounts.slice(24, 31);

    // Calculate daily sales average
    const dailySales = totalUsers > 0 ? (totalUsers / 30).toFixed(0) : 0;

    // Calculate monthly growth percentage (relative to the previous 30 days period)
    const previousMonthTotal = dashboardData.userCounts[0];
    let monthlyGrowthPercent = 0; // Capped at 100% for progress bar display
    let monthlyGrowthDisplay = 0; // Actual percentage shown as text
    if (previousMonthTotal > 0) {
      monthlyGrowthPercent = Math.min((usersMonth / previousMonthTotal * 100), 100); // Progress bar max is 100%
      monthlyGrowthDisplay = (usersMonth / previousMonthTotal * 100).toFixed(1);
    } else if (usersMonth > 0) {
      monthlyGrowthPercent = 100;
      monthlyGrowthDisplay = 100;
    }

    return (
      <div style={{width: "100%", maxWidth: "1400px"}}>
        <Row id="statistic" gutter={[16, 16]}>
          <Col xs={24} sm={12} lg={6}>
            <Card
              bordered={false}
              style={{
                boxShadow: "0 1px 2px rgba(0,0,0,0.03), 0 1px 6px rgba(0,0,0,0.03)",
                borderRadius: "2px",
              }}
            >
              <div style={{display: "flex", alignItems: "center", marginBottom: "8px"}}>
                <span style={{fontSize: "14px", color: "rgba(0,0,0,0.45)"}}>{i18next.t("home:Total users")}</span>
                <InfoCircleOutlined style={{marginLeft: "4px", color: "rgba(0,0,0,0.25)", fontSize: "12px"}} />
              </div>
              <div style={{fontSize: "30px", fontWeight: "500", marginBottom: "8px", color: "rgba(0,0,0,0.85)"}}>
                {totalUsers.toLocaleString()}
              </div>
              <div style={{display: "flex", alignItems: "center", fontSize: "12px"}}>
                <span style={{color: "rgba(0,0,0,0.45)"}}>{i18next.t("home:Daily sales")}</span>
                <span style={{marginLeft: "auto", color: "rgba(0,0,0,0.85)"}}>
                  {dailySales}
                </span>
              </div>
            </Card>
          </Col>

          <Col xs={24} sm={12} lg={6}>
            <Card
              bordered={false}
              style={{
                boxShadow: "0 1px 2px rgba(0,0,0,0.03), 0 1px 6px rgba(0,0,0,0.03)",
                borderRadius: "2px",
              }}
            >
              <div style={{display: "flex", alignItems: "center", marginBottom: "8px"}}>
                <span style={{fontSize: "14px", color: "rgba(0,0,0,0.45)"}}>{i18next.t("home:New users today")}</span>
                <InfoCircleOutlined style={{marginLeft: "4px", color: "rgba(0,0,0,0.25)", fontSize: "12px"}} />
              </div>
              <div style={{fontSize: "30px", fontWeight: "500", marginBottom: "8px", color: "rgba(0,0,0,0.85)"}}>
                {usersToday.toLocaleString()}
              </div>
              <div style={{height: "46px", marginTop: "8px"}}>
                <MiniChart data={last7Days} color="#1890ff" />
              </div>
            </Card>
          </Col>

          <Col xs={24} sm={12} lg={6}>
            <Card
              bordered={false}
              style={{
                boxShadow: "0 1px 2px rgba(0,0,0,0.03), 0 1px 6px rgba(0,0,0,0.03)",
                borderRadius: "2px",
              }}
            >
              <div style={{display: "flex", alignItems: "center", marginBottom: "8px"}}>
                <span style={{fontSize: "14px", color: "rgba(0,0,0,0.45)"}}>{i18next.t("home:New users past 7 days")}</span>
                <InfoCircleOutlined style={{marginLeft: "4px", color: "rgba(0,0,0,0.25)", fontSize: "12px"}} />
              </div>
              <div style={{fontSize: "30px", fontWeight: "500", marginBottom: "8px", color: "rgba(0,0,0,0.85)"}}>
                {usersWeek.toLocaleString()}
              </div>
              <div style={{display: "flex", alignItems: "center", fontSize: "12px"}}>
                <span style={{color: "rgba(0,0,0,0.45)"}}>{i18next.t("home:Week over week")}</span>
                <span style={{marginLeft: "auto", color: weeklyGrowth >= 0 ? "#52c41a" : "#ff4d4f"}}>
                  {weeklyGrowth >= 0 ? <ArrowUpOutlined /> : <ArrowDownOutlined />} {Math.abs(weeklyGrowth)}%
                </span>
              </div>
            </Card>
          </Col>

          <Col xs={24} sm={12} lg={6}>
            <Card
              bordered={false}
              style={{
                boxShadow: "0 1px 2px rgba(0,0,0,0.03), 0 1px 6px rgba(0,0,0,0.03)",
                borderRadius: "2px",
              }}
            >
              <div style={{display: "flex", alignItems: "center", marginBottom: "8px"}}>
                <span style={{fontSize: "14px", color: "rgba(0,0,0,0.45)"}}>{i18next.t("home:New users past 30 days")}</span>
                <InfoCircleOutlined style={{marginLeft: "4px", color: "rgba(0,0,0,0.25)", fontSize: "12px"}} />
              </div>
              <div style={{fontSize: "30px", fontWeight: "500", marginBottom: "8px", color: "rgba(0,0,0,0.85)"}}>
                {usersMonth.toLocaleString()}
              </div>
              <div>
                <Progress
                  percent={monthlyGrowthPercent}
                  strokeColor="#52c41a"
                  showInfo={false}
                  size="small"
                />
                <div style={{display: "flex", alignItems: "center", fontSize: "12px", marginTop: "4px"}}>
                  <span style={{color: "rgba(0,0,0,0.45)"}}>{i18next.t("home:Monthly growth")}</span>
                  <span style={{marginLeft: "auto", color: "rgba(0,0,0,0.85)"}}>
                    {monthlyGrowthDisplay}%
                  </span>
                </div>
              </div>
            </Card>
          </Col>
        </Row>
      </div>
    );
  };

  return (
    <div style={{padding: "24px", background: "#f0f2f5", minHeight: "100vh"}}>
      <div style={{maxWidth: "1400px", margin: "0 auto"}}>
        {renderEChart()}
        <Card
          bordered={false}
          style={{
            marginTop: "24px",
            boxShadow: "0 1px 2px rgba(0,0,0,0.03), 0 1px 6px rgba(0,0,0,0.03)",
            borderRadius: "2px",
          }}
        >
          <div id="echarts-chart" style={{width: "100%", height: "400px"}} />
        </Card>
      </div>
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
