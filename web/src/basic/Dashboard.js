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

import React, {useCallback, useEffect, useMemo, useRef, useState} from "react";
import {Card, Col, Progress, Row, Statistic, Tour} from "antd";
import Loading from "../common/Loading";
import {ApartmentOutlined, AppstoreOutlined, ArrowUpOutlined, KeyOutlined, SafetyOutlined, TeamOutlined} from "@ant-design/icons";
import * as echarts from "echarts";
import i18next from "i18next";
import * as DashboardBackend from "../backend/DashboardBackend";
import * as Setting from "../Setting";
import * as TourConfig from "../TourConfig";

// Multi-hue palette: blue / sky / cyan / teal / indigo / violet, with one purple (#5734d3) added.
const CHART_COLORS = [
  "#1677ff", // blue-500
  "#0ea5e9", // sky-500
  "#06b6d4", // cyan-500
  "#14b8a6", // teal-500
  "#6366f1", // indigo-500
  "#8b5cf6", // violet-500
  "#0958d9", // blue-700
  "#0284c7", // sky-700
  "#0891b2", // cyan-700
  "#0f766e", // teal-700
  "#5734d3", // purple (primary)
  "#7c3aed", // violet-700
  "#38bdf8", // sky-300
  "#5eead4", // teal-300
];

// Reusable ECharts container that handles initialization, option updates, and resize cleanup.
const EchartsWidget = React.memo(({option, style}) => {
  const containerRef = useRef(null);
  const chartRef = useRef(null);

  useEffect(() => {
    if (!containerRef.current) {
      return;
    }
    const chart = echarts.init(containerRef.current);
    chartRef.current = chart;

    const observer = new ResizeObserver(() => chart.resize());
    observer.observe(containerRef.current);

    return () => {
      observer.disconnect();
      chart.dispose();
      chartRef.current = null;
    };
  }, []);

  useEffect(() => {
    if (chartRef.current && option) {
      chartRef.current.setOption(option, {notMerge: true});
    }
  }, [option]);

  return <div ref={containerRef} style={style} />;
});

EchartsWidget.displayName = "EchartsWidget";

function buildDateArray() {
  const arr = [];
  const now = new Date();
  for (let i = 30; i >= 0; i--) {
    const d = new Date(now);
    d.setDate(d.getDate() - i);
    arr.push(`${d.getMonth() + 1}-${d.getDate()}`);
  }
  return arr;
}

const DATE_ARRAY = buildDateArray();

function buildTrendOption(dashboardData) {
  if (!dashboardData) {
    return null;
  }
  const series = [
    {name: i18next.t("general:Users"), data: dashboardData.userCounts},
    {name: i18next.t("general:Applications"), data: dashboardData.applicationCounts},
    {name: i18next.t("application:Providers"), data: dashboardData.providerCounts},
    {name: i18next.t("general:Organizations"), data: dashboardData.organizationCounts},
    {name: i18next.t("general:Roles"), data: dashboardData.roleCounts},
    {name: i18next.t("general:Permissions"), data: dashboardData.permissionCounts},
    {name: i18next.t("general:Groups"), data: dashboardData.groupCounts},
    {name: i18next.t("general:Resources"), data: dashboardData.resourceCounts},
    {name: i18next.t("general:Certs"), data: dashboardData.certCounts},
    {name: i18next.t("general:Subscriptions"), data: dashboardData.subscriptionCounts},
    {name: i18next.t("general:Models"), data: dashboardData.modelCounts},
    {name: i18next.t("general:Transactions"), data: dashboardData.transactionCounts},
    {name: i18next.t("general:Adapters"), data: dashboardData.adapterCounts},
    {name: i18next.t("general:Enforcers"), data: dashboardData.enforcerCounts},
  ];

  return {
    color: CHART_COLORS,
    tooltip: {trigger: "axis"},
    legend: {
      type: "scroll",
      top: 0,
      data: series.map(s => s.name),
      selected: {
        [i18next.t("general:Adapters")]: false,
        [i18next.t("general:Enforcers")]: false,
        [i18next.t("general:Models")]: false,
        [i18next.t("general:Subscriptions")]: false,
      },
    },
    grid: {left: "3%", right: "4%", bottom: "3%", top: "22%", containLabel: true},
    xAxis: {type: "category", boundaryGap: false, data: DATE_ARRAY},
    yAxis: {type: "value", minInterval: 1},
    series: series.map(s => ({
      name: s.name,
      type: "line",
      smooth: true,
      symbol: "none",
      data: s.data,
    })),
  };
}

function buildProviderOption(providerData) {
  if (!providerData || providerData.length === 0) {
    return null;
  }
  return {
    color: CHART_COLORS,
    tooltip: {trigger: "item", formatter: "{b}: {c} ({d}%)"},
    legend: {
      type: "scroll",
      orient: "vertical",
      right: 8,
      left: "56%",
      top: "center",
      textStyle: {fontSize: 12},
    },
    series: [{
      type: "pie",
      radius: ["42%", "68%"],
      center: ["22%", "50%"],
      avoidLabelOverlap: true,
      itemStyle: {borderRadius: 5, borderColor: "#fff", borderWidth: 2},
      label: {show: false},
      emphasis: {
        label: {show: true, fontSize: 13, fontWeight: "bold"},
        itemStyle: {shadowBlur: 10, shadowOffsetX: 0, shadowColor: "rgba(0,0,0,0.25)"},
      },
      data: providerData.map(p => ({
        name: p.type || i18next.t("general:None"),
        value: p.count,
      })),
    }],
  };
}

function buildHeatmapOption(heatmapData) {
  if (!heatmapData || !heatmapData.data) {
    return null;
  }

  const range = (heatmapData.dateRange && heatmapData.dateRange.length === 2)
    ? heatmapData.dateRange
    : (() => {
      const end = new Date();
      const start = new Date(end);
      start.setFullYear(end.getFullYear() - 1);
      return [start.toISOString().slice(0, 10), end.toISOString().slice(0, 10)];
    })();

  return {
    tooltip: {
      position: "top",
      formatter: (params) => {
        const [date, count] = params.data;
        return `${date} &nbsp; <b>${count}</b>`;
      },
    },
    visualMap: {
      min: 0,
      max: Math.max(heatmapData.maxCount, 1),
      show: false,
      inRange: {color: ["#f3f0ff", "#d9d1f7", "#b5a8ef", "#7c5ce0", "#5734d3"]},
    },
    calendar: {
      top: 28,
      left: 36,
      right: 8,
      bottom: 0,
      range,
      cellSize: [13, 13],
      itemStyle: {
        color: "#f3f0ff",
        borderWidth: 2,
        borderColor: "#fff",
        borderRadius: 2,
      },
      dayLabel: {
        firstDay: 0,
        nameMap: ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"],
        fontSize: 10,
        color: "#888",
      },
      monthLabel: {
        fontSize: 11,
        color: "#555",
      },
      yearLabel: {show: false},
      splitLine: {show: false},
    },
    series: [{
      type: "heatmap",
      coordinateSystem: "calendar",
      data: heatmapData.data.map(d => [d.date, d.count]),
      itemStyle: {borderRadius: 2},
    }],
  };
}

const Dashboard = (props) => {
  const [dashboardData, setDashboardData] = useState(null);
  const [providerData, setProviderData] = useState(null);
  const [mfaData, setMfaData] = useState(null);
  const [heatmapData, setHeatmapData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [isTourVisible, setIsTourVisible] = useState(TourConfig.getTourVisible());
  const nextPathName = TourConfig.getNextUrl("home");

  const getOrganizationName = useCallback(() => {
    if (!Setting.isAdminUser(props.account) && Setting.isLocalAdminUser(props.account)) {
      return props.account.owner;
    }
    const stored = localStorage.getItem("organization");
    return stored === "All" ? "" : (stored || "");
  }, [props.account]);

  const loadAllData = useCallback(() => {
    if (!Setting.isLocalAdminUser(props.account)) {
      return;
    }
    const org = getOrganizationName();
    setLoading(true);
    setDashboardData(null);
    setProviderData(null);
    setMfaData(null);
    setHeatmapData(null);

    const applyResult = (settled, setter) => {
      if (settled.status === "rejected") {
        Setting.showMessage("error", settled.reason?.message ?? String(settled.reason));
        return;
      }
      const res = settled.value;
      if (res.status === "ok") {
        setter(res.data);
      } else {
        Setting.showMessage("error", res.msg);
      }
    };

    // allSettled ensures a failing endpoint doesn't prevent other charts from rendering.
    Promise.allSettled([
      DashboardBackend.getDashboard(org),
      DashboardBackend.getDashboardProviders(org),
      DashboardBackend.getDashboardMfa(org),
      DashboardBackend.getDashboardHeatmap(org),
    ]).then(([dashResult, provResult, mfaResult, heatResult]) => {
      applyResult(dashResult, setDashboardData);
      applyResult(provResult, setProviderData);
      applyResult(mfaResult, setMfaData);
      applyResult(heatResult, setHeatmapData);
    }).finally(() => setLoading(false));
  }, [props.account, getOrganizationName]);

  useEffect(() => {
    if (!Setting.isLocalAdminUser(props.account)) {
      props.history.push("/apps");
      return;
    }
    loadAllData();
  }, [props.account]);

  useEffect(() => {
    window.addEventListener("storageOrganizationChanged", loadAllData);
    return () => window.removeEventListener("storageOrganizationChanged", loadAllData);
  }, [loadAllData]);

  useEffect(() => {
    const handleTourChange = () => setIsTourVisible(TourConfig.getTourVisible());
    window.addEventListener("storageTourChanged", handleTourChange);
    return () => window.removeEventListener("storageTourChanged", handleTourChange);
  }, []);

  const trendOption = useMemo(() => buildTrendOption(dashboardData), [dashboardData]);
  const providerOption = useMemo(() => buildProviderOption(providerData), [providerData]);
  const heatmapOption = useMemo(() => buildHeatmapOption(heatmapData), [heatmapData]);

  const mfaRate = useMemo(() => {
    if (!mfaData || mfaData.total === 0) {
      return 0;
    }
    return parseFloat((mfaData.enabled / mfaData.total * 100).toFixed(2));
  }, [mfaData]);

  const getSteps = () => {
    const steps = TourConfig.TourObj["home"];
    steps.forEach((item, index) => {
      item.target = () => document.getElementById(item.id) || null;
      if (index === steps.length - 1) {
        item.nextButtonProps = {children: TourConfig.getNextButtonChild(nextPathName)};
      }
    });
    return steps;
  };

  const handleTourClose = () => {
    TourConfig.setIsTourVisible(false);
    setIsTourVisible(false);
  };

  const handleTourFinish = () => {
    if (nextPathName !== "") {
      props.history.push("/" + nextPathName);
      TourConfig.setIsTourVisible(true);
    }
  };

  if (loading && !dashboardData) {
    return (
      <Loading type="page" tip={i18next.t("login:Loading")} />
    );
  }

  const cardStyle = {borderRadius: 8, border: "1px solid #e8e8e8"};
  const gutter = [16, 16];

  const userCounts = dashboardData?.userCounts ?? Array(31).fill(0);
  const orgCounts = dashboardData?.organizationCounts ?? Array(31).fill(0);
  const tokenCounts = dashboardData?.tokenCounts ?? Array(31).fill(0);
  const appCounts = dashboardData?.applicationCounts ?? Array(31).fill(0);
  const provCounts = dashboardData?.providerCounts ?? Array(31).fill(0);

  return (
    <div style={{padding: Setting.isMobile() ? "12px" : "24px"}}>

      {/* ── Row 1: Key metrics (8 cards × lg:3 = 24 cols → single row on lg+) ── */}
      <Row id="statistic" gutter={gutter}>
        <Col xs={12} sm={8} md={6} lg={3}>
          <Card variant="borderless" style={cardStyle}>
            <Statistic
              title={i18next.t("home:Total users")}
              value={userCounts[30]}
              prefix={<TeamOutlined style={{color: "#1677ff"}} />}
              valueStyle={{color: "#1677ff"}}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6} lg={3}>
          <Card variant="borderless" style={cardStyle}>
            <Statistic
              title={i18next.t("home:New users today")}
              value={Math.max(0, userCounts[30] - userCounts[29])}
              prefix={<ArrowUpOutlined style={{color: "#0958d9"}} />}
              valueStyle={{color: "#0958d9"}}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6} lg={3}>
          <Card variant="borderless" style={cardStyle}>
            <Statistic
              title={i18next.t("home:New users / 7 days")}
              value={Math.max(0, userCounts[30] - userCounts[23])}
              prefix={<ArrowUpOutlined style={{color: "#0958d9"}} />}
              valueStyle={{color: "#0958d9"}}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6} lg={3}>
          <Card variant="borderless" style={cardStyle}>
            <Statistic
              title={i18next.t("home:New users / 30 days")}
              value={Math.max(0, userCounts[30] - userCounts[0])}
              prefix={<ArrowUpOutlined style={{color: "#0958d9"}} />}
              valueStyle={{color: "#0958d9"}}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6} lg={3}>
          <Card variant="borderless" style={cardStyle}>
            <Statistic
              title={i18next.t("general:Organizations")}
              value={orgCounts[30]}
              prefix={<ApartmentOutlined style={{color: "#6366f1"}} />}
              valueStyle={{color: "#6366f1"}}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6} lg={3}>
          <Card variant="borderless" style={cardStyle}>
            <Statistic
              title={i18next.t("general:Tokens")}
              value={tokenCounts[30]}
              prefix={<KeyOutlined style={{color: "#14b8a6"}} />}
              valueStyle={{color: "#14b8a6"}}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6} lg={3}>
          <Card variant="borderless" style={cardStyle}>
            <Statistic
              title={i18next.t("general:Applications")}
              value={appCounts[30]}
              prefix={<AppstoreOutlined style={{color: "#5734d3"}} />}
              valueStyle={{color: "#5734d3"}}
            />
          </Card>
        </Col>
        <Col xs={12} sm={8} md={6} lg={3}>
          <Card variant="borderless" style={cardStyle}>
            <Statistic
              title={i18next.t("application:Providers")}
              value={provCounts[30]}
              prefix={<SafetyOutlined style={{color: "#0891b2"}} />}
              valueStyle={{color: "#0891b2"}}
            />
          </Card>
        </Col>
      </Row>

      {/* ── Row 2: 30-day trend + Provider distribution ── */}
      <Row gutter={gutter} style={{marginTop: 16}}>
        <Col xs={24} xl={14}>
          <Card
            title={i18next.t("home:Past 30 Days")}
            variant="borderless"
            style={cardStyle}
          >
            <EchartsWidget option={trendOption} style={{height: 320}} />
          </Card>
        </Col>
        <Col xs={24} xl={10}>
          <Card
            title={i18next.t("application:Providers")}
            variant="borderless"
            style={{...cardStyle, height: "100%"}}
          >
            {providerData && providerData.length > 0
              ? <EchartsWidget option={providerOption} style={{height: 320}} />
              : (
                <div style={{height: 320, display: "flex", alignItems: "center", justifyContent: "center", color: "#bbb"}}>
                  {i18next.t("general:None")}
                </div>
              )
            }
          </Card>
        </Col>
      </Row>

      {/* ── Row 3: MFA coverage + Activity heatmap ── */}
      <Row gutter={gutter} style={{marginTop: 16}}>
        <Col xs={24} xl={8}>
          <Card
            title={i18next.t("user:MFA accounts")}
            variant="borderless"
            style={{...cardStyle, height: "100%"}}
            styles={{body: {height: "calc(100% - 57px)", display: "flex", alignItems: "center", justifyContent: "center"}}}
          >
            {mfaData
              ? (
                <div style={{display: "flex", flexDirection: "row", alignItems: "center", justifyContent: "center", gap: 40, padding: "0 16px"}}>
                  <Progress
                    type="circle"
                    percent={mfaRate}
                    size={130}
                    strokeColor={{
                      "0%": "#4228a8",
                      "100%": "#5734d3",
                    }}
                    format={pct => (
                      <span>
                        <div style={{fontSize: 22, fontWeight: "bold", color: "#5734d3", lineHeight: 1.2}}>{Number(pct).toFixed(2)}%</div>
                        <div style={{fontSize: 12, color: "#999", marginTop: 4}}>MFA</div>
                      </span>
                    )}
                  />
                  <div style={{display: "flex", flexDirection: "column", gap: 16, fontSize: 14}}>
                    <div>
                      <div style={{color: "#5734d3", fontWeight: 600, fontSize: 22}}>{mfaData.enabled}</div>
                      <div style={{color: "#888"}}>{i18next.t("general:Enabled")}</div>
                    </div>
                    <div>
                      <div style={{color: "#9b82e8", fontWeight: 600, fontSize: 22}}>{mfaData.disabled}</div>
                      <div style={{color: "#888"}}>{i18next.t("general:Disable")}</div>
                    </div>
                  </div>
                </div>
              )
              : (
                <div style={{color: "#bbb"}}>
                  {i18next.t("general:None")}
                </div>
              )
            }
          </Card>
        </Col>
        <Col xs={24} xl={16}>
          <Card
            title={i18next.t("application:Signin")}
            variant="borderless"
            style={{...cardStyle, height: "100%"}}
          >
            <EchartsWidget option={heatmapOption} style={{height: 180}} />
          </Card>
        </Col>
      </Row>

      <Tour
        open={Setting.isMobile() ? false : isTourVisible}
        onClose={handleTourClose}
        steps={getSteps()}
        indicatorsRender={(current, total) => (
          <span>{current + 1} / {total}</span>
        )}
        onFinish={handleTourFinish}
      />
    </div>
  );
};

export default Dashboard;
