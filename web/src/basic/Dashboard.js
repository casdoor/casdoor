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

import {useCallback, useEffect, useMemo, useState} from "react";
import {Card, Col, Grid, Row, Spin, Tour} from "antd";
import * as echarts from "echarts";
import i18next from "i18next";
import * as DashboardBackend from "../backend/DashboardBackend";
import * as Setting from "../Setting";
import * as TourConfig from "../TourConfig";

const getDateArray = () => {
  const currentDate = new Date();
  const dateArray = [];
  for (let i = 30; i >= 0; i--) {
    const date = new Date(currentDate);
    date.setDate(date.getDate() - i);
    dateArray.push(`${date.getMonth() + 1}-${date.getDate()}`);
  }
  return dateArray;
};

const formatAxisValue = (value) => {
  const numberValue = Number(value);
  if (!Number.isFinite(numberValue)) {
    return "";
  }
  const absValue = Math.abs(numberValue);

  if (absValue >= 1e9) {
    return `${(numberValue / 1e9).toFixed(1)}B`;
  }
  if (absValue >= 1e6) {
    return `${(numberValue / 1e6).toFixed(1)}M`;
  }
  if (absValue >= 1e3) {
    return `${(numberValue / 1e3).toFixed(1)}K`;
  }
  return `${numberValue}`;
};

const getYAxisLayout = (values) => {
  const maxValue = (values ?? []).reduce((max, item) => Math.max(max, Math.abs(Number(item) || 0)), 0);
  const label = formatAxisValue(maxValue);
  const left = Math.min(44, Math.max(32, label.length * 6 + 10));
  const nameGap = Math.min(68, Math.max(48, label.length * 4 + 12));
  return {left, nameGap};
};

const attachEChart = (chartDom, option) => {
  const myChart = echarts.getInstanceByDom(chartDom) ?? echarts.init(chartDom);
  myChart.setOption(option);
  myChart.resize();

  const handleResize = () => myChart.resize();
  window.addEventListener("resize", handleResize);

  return () => {
    window.removeEventListener("resize", handleResize);
    myChart.dispose();
  };
};

const LineChartWidget = ({chartId, seriesConfig, height, data}) => {
  useEffect(() => {
    if (!data || !document.getElementById(chartId)) {
      return;
    }

    const seriesValues = seriesConfig.flatMap(s => s.data ?? data?.[s.dataKey] ?? []);
    const yAxisLayout = getYAxisLayout(seriesValues);

    const option = {
      tooltip: {
        trigger: "axis",
        axisPointer: {type: "line"},
        confine: true,
      },
      legend: {type: "scroll", data: seriesConfig.map(s => s.name), top: 0, left: "center", width: "90%", pageButtonPosition: "end"},
      grid: {left: yAxisLayout.left, right: "4%", bottom: 28, top: 32, containLabel: true},
      xAxis: {type: "category", name: i18next.t("Date"), nameLocation: "middle", nameGap: 28, boundaryGap: false, data: getDateArray()},
      yAxis: {type: "value", name: i18next.t("Total"), nameLocation: "middle", nameGap: yAxisLayout.nameGap, splitArea: {show: true}, axisLabel: {formatter: formatAxisValue}},
      series: seriesConfig.map(s => ({
        name: s.name,
        type: "line",
        data: s.data ?? data[s.dataKey],
        smooth: true,
        ...(s.color ? {lineStyle: {color: s.color}, itemStyle: {color: s.color}} : {}),
      })),
    };

    return attachEChart(document.getElementById(chartId), option);
  }, [data, chartId, height, seriesConfig]);

  if (!data) {
    return <div style={{display: "flex", justifyContent: "center", alignItems: "center", height}}><Spin size="large" /></div>;
  }
  return <div id={chartId} style={{width: "100%", height}} />;
};

const MfaCoverageWidget = ({chartId, height, data}) => {
  useEffect(() => {
    if (!data || !document.getElementById(chartId)) {
      return;
    }

    const items = Array.isArray(data) ? data : [];
    const organizations = items.map(i => i.organization);
    const adminDisabled = items.map(i => i.adminDisabled ?? 0);
    const userDisabled = items.map(i => i.userDisabled ?? 0);
    const adminEnabled = items.map(i => i.adminEnabled ?? 0);
    const userEnabled = items.map(i => i.userEnabled ?? 0);

    const yAxisLayout = getYAxisLayout([...adminDisabled, ...userDisabled, ...adminEnabled, ...userEnabled]);

    const option = {
      tooltip: {trigger: "axis", axisPointer: {type: "shadow"}},
      legend: {top: 0, left: "center"},
      grid: {left: yAxisLayout.left, right: "4%", bottom: 28, top: 32, containLabel: true},
      xAxis: {type: "category", name: i18next.t("general:Organization"), nameLocation: "middle", nameGap: 28, data: organizations},
      yAxis: {type: "value", name: i18next.t("Total"), nameLocation: "middle", nameGap: yAxisLayout.nameGap, minInterval: 1, axisLabel: {formatter: formatAxisValue}},
      series: [
        {name: `${i18next.t("ldap:Admin")} (${i18next.t("general:Disable")})`, type: "bar", stack: "total", data: adminDisabled, itemStyle: {color: "#EE6666"}, barMaxWidth: 28},
        {name: `${i18next.t("general:User")} (${i18next.t("general:Disable")})`, type: "bar", stack: "total", data: userDisabled, itemStyle: {color: "#FC8452"}, barMaxWidth: 28},
        {name: `${i18next.t("ldap:Admin")} (${i18next.t("general:Enabled")})`, type: "bar", stack: "total", data: adminEnabled, itemStyle: {color: "#3BA272"}, barMaxWidth: 28},
        {name: `${i18next.t("general:User")} (${i18next.t("general:Enabled")})`, type: "bar", stack: "total", data: userEnabled, itemStyle: {color: "#91CC75"}, barMaxWidth: 28},
      ],
    };

    return attachEChart(document.getElementById(chartId), option);
  }, [data, chartId, height]);

  if (!data) {
    return <div style={{display: "flex", justifyContent: "center", alignItems: "center", height}}><Spin size="large" /></div>;
  }
  return <div id={chartId} style={{width: "100%", height}} />;
};

const DonutWidget = ({chartId, height, data}) => {
  useEffect(() => {
    if (!data || !document.getElementById(chartId)) {
      return;
    }

    const providerData = Object.entries(data ?? {})
      .filter(([, value]) => Number(value) > 0)
      .sort((a, b) => Number(b[1]) - Number(a[1]))
      .map(([name, value]) => {
        const iconUrl = Setting.getProviderLogoURL({category: "OAuth", type: name});
        const labelConfig = iconUrl
          ? {formatter: "{icon|} {b}", rich: {icon: {width: 12, height: 12, backgroundColor: {image: iconUrl}}}}
          : {formatter: "{b}"};

        return {name, value: Number(value), label: {show: true, position: "outside", ...labelConfig}};
      });

    const option = {
      title: {text: i18next.t("provider:Third-party"), left: "center", bottom: 0, textStyle: {fontSize: 12, fontWeight: 400, color: "#8c8c8c"}},
      tooltip: {trigger: "item", formatter: (p) => `${p.name}: ${p.value} (${p.percent}%)`},
      legend: {show: false},
      series: [{type: "pie", radius: ["0%", "85%"], center: ["50%", "46%"], labelLine: {show: true, length: 12, length2: 10}, itemStyle: {borderColor: "#fff", borderWidth: 1}, data: providerData}],
    };

    return attachEChart(document.getElementById(chartId), option);
  }, [data, chartId, height]);

  if (!data) {
    return <div style={{display: "flex", justifyContent: "center", alignItems: "center", height}}><Spin size="large" /></div>;
  }
  return <div id={chartId} style={{width: "100%", height}} />;
};

const HeatmapWidget = ({chartId, height, data}) => {
  useEffect(() => {
    if (!data || !document.getElementById(chartId)) {
      return;
    }

    const {xAxis = [], yAxis = [], data: mapData = [], max: maxValue = 0} = data;
    const option = {
      tooltip: {position: "top", formatter: (p) => `${yAxis[p.value[1]]} ${xAxis[p.value[0]]}:00<br/>${i18next.t("application:Signin")}${i18next.t("system:Count")}: ${p.value[2]}`},
      grid: {left: "4%", right: "4%", top: 12, bottom: 28, containLabel: true},
      xAxis: {type: "category", name: i18next.t("Hour"), nameLocation: "middle", nameGap: 28, data: xAxis, splitArea: {show: true}},
      yAxis: {z: 5, type: "category", name: i18next.t("general:Date"), nameLocation: "middle", nameGap: 36, nameTextStyle: {fontSize: 11}, data: yAxis, splitArea: {show: true}},
      visualMap: {show: false, min: 0, max: Math.max(1, Number(maxValue)), inRange: {color: ["#f5f5f5", "#91CC75", "#3ba272"]}},
      series: [{type: "heatmap", data: mapData, label: {show: true, formatter: (p) => (p.value[2] ? `${p.value[2]}` : ""), fontSize: 10, color: "#000"}, emphasis: {itemStyle: {shadowBlur: 10, shadowColor: "rgba(0, 0, 0, 0.3)"}}}],
    };

    return attachEChart(document.getElementById(chartId), option);
  }, [data, chartId, height]);

  if (!data) {
    return <div style={{display: "flex", justifyContent: "center", alignItems: "center", height}}><Spin size="large" /></div>;
  }
  return <div id={chartId} style={{width: "100%", height}} />;
};

const Dashboard = (props) => {
  const [dashboardData, setDashboardData] = useState(null);
  const [usersByProvider, setUsersByProvider] = useState(null);
  const [loginHeatmap, setLoginHeatmap] = useState(null);
  const [mfaCoverage, setMfaCoverage] = useState(null);
  const [isTourVisible, setIsTourVisible] = useState(TourConfig.getTourVisible());

  const screens = Grid.useBreakpoint();
  const nextPathName = TourConfig.getNextUrl("home");
  const chartHeight = screens.xs ? "300px" : "clamp(300px, 40vh, 550px)";

  const getOrganizationName = useCallback(() => {
    let organization = localStorage.getItem("organization") === "All" ? "" : localStorage.getItem("organization");
    if (!Setting.isAdminUser(props.account) && Setting.isLocalAdminUser(props.account)) {
      organization = props.account.owner;
    }
    return organization;
  }, [props.account]);

  const fetchDashboardData = useCallback((organization) => {
    Promise.allSettled([
      DashboardBackend.getDashboard(organization),
      DashboardBackend.getDashboardUsersByProvider(organization),
      DashboardBackend.getDashboardLoginHeatmap(organization),
      DashboardBackend.getDashboardMfaCoverage(organization),
    ]).then(([dashRes, userRes, heatmapRes, mfaRes]) => {
      const handle = (res, setter) => {
        if (res.status === "fulfilled" && res.value?.status === "ok") {
          setter(res.value.data);
        } else {
          Setting.showMessage("error", res.value?.msg || res.reason?.message || "Request failed");
        }
      };
      handle(dashRes, setDashboardData);
      handle(userRes, setUsersByProvider);
      handle(heatmapRes, setLoginHeatmap);
      handle(mfaRes, setMfaCoverage);
    });
  }, []);

  useEffect(() => {
    if (!Setting.isLocalAdminUser(props.account)) {
      props.history.push("/apps");
      return;
    }
    fetchDashboardData(getOrganizationName());
  }, [props.account, props.owner, getOrganizationName, fetchDashboardData, props.history]);

  useEffect(() => {
    const handleTourChange = () => setIsTourVisible(TourConfig.getTourVisible());
    const handleOrgChange = () => {
      if (!Setting.isLocalAdminUser(props.account)) {
        return;
      }
      setDashboardData(null);
      setUsersByProvider(null);
      setLoginHeatmap(null);
      setMfaCoverage(null);
      fetchDashboardData(getOrganizationName());
    };

    window.addEventListener("storageTourChanged", handleTourChange);
    window.addEventListener("storageOrganizationChanged", handleOrgChange);
    return () => {
      window.removeEventListener("storageTourChanged", handleTourChange);
      window.removeEventListener("storageOrganizationChanged", handleOrgChange);
    };
  }, [props.account, getOrganizationName, fetchDashboardData]);

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

  const lineChartConfig = useMemo(() => [
    {name: i18next.t("general:Users"), dataKey: "userCounts"},
    {name: i18next.t("general:Providers"), dataKey: "providerCounts"},
    {name: i18next.t("general:Applications"), dataKey: "applicationCounts"},
    {name: i18next.t("general:Organizations"), dataKey: "organizationCounts"},
    {name: i18next.t("general:Subscriptions"), dataKey: "subscriptionCounts"},
    {name: i18next.t("general:Roles"), dataKey: "roleCounts"},
    {name: i18next.t("general:Groups"), dataKey: "groupCounts"},
    {name: i18next.t("general:Resources"), dataKey: "resourceCounts"},
    {name: i18next.t("general:Certs"), dataKey: "certCounts"},
    {name: i18next.t("general:Permissions"), dataKey: "permissionCounts"},
    {name: i18next.t("general:Transactions"), dataKey: "transactionCounts"},
    {name: i18next.t("general:Models"), dataKey: "modelCounts"},
    {name: i18next.t("general:Adapters"), dataKey: "adapterCounts"},
    {name: i18next.t("general:Enforcers"), dataKey: "enforcerCounts"},
  ], []);

  return (
    <div>
      <Row gutter={[16, 16]}>
        <Col xs={12}>
          <Card title={i18next.t("home:Past 30 Days")} size="small" style={{height: "100%"}} styles={{body: {padding: 0}}}>
            <LineChartWidget chartId="past-30-days-chart" height={chartHeight} seriesConfig={lineChartConfig} data={dashboardData} />
          </Card>
        </Col>
        <Col xs={12}>
          <Card title={i18next.t("general:MFA items")} size="small" style={{height: "100%"}} styles={{body: {padding: 0}}}>
            <MfaCoverageWidget chartId="mfa-coverage-chart" height={chartHeight} data={mfaCoverage} />
          </Card>
        </Col>
        <Col xs={12}>
          <Card title={i18next.t("user:3rd-party logins")} size="small" style={{height: "100%"}} styles={{body: {padding: 0}}}>
            <DonutWidget chartId="users-by-provider-chart" height={chartHeight} data={usersByProvider} />
          </Card>
        </Col>
        <Col xs={12}>
          <Card title={i18next.t("general:Records")} size="small" style={{height: "100%"}} styles={{body: {padding: 0}}}>
            <HeatmapWidget chartId="login-time-heatmap" height={chartHeight} data={loginHeatmap} />
          </Card>
        </Col>
      </Row>

      <Tour
        open={Setting.isMobile() ? false : isTourVisible}
        onClose={() => {
          TourConfig.setIsTourVisible(false);
          setIsTourVisible(false);
        }}
        steps={getSteps()}
        indicatorsRender={(current, total) => <span>{current + 1} / {total}</span>}
        onFinish={() => {
          if (nextPathName) {
            props.history.push("/" + nextPathName);
            TourConfig.setIsTourVisible(true);
          }
        }}
      />
    </div>
  );
};

export default Dashboard;
