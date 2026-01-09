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

import {Card, Col, Row, Spin, Tour} from "antd";
import * as echarts from "echarts";
import i18next from "i18next";
import React from "react";
import * as DashboardBackend from "../backend/DashboardBackend";
import * as Setting from "../Setting";
import * as TourConfig from "../TourConfig";

const Dashboard = (props) => {
  const [dashboardData, setDashboardData] = React.useState(null);
  const [usersByProvider, setUsersByProvider] = React.useState(null);
  const [loginHeatmap, setLoginHeatmap] = React.useState(null);
  const [mfaCoverage, setMfaCoverage] = React.useState(null);
  const [isTourVisible, setIsTourVisible] = React.useState(TourConfig.getTourVisible());
  const padding = 20;
  const gutter = 16;
  const chartHeight = 260;
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

  const fetchDashboardData = (organization) => {
    Promise.allSettled([
      DashboardBackend.getDashboard(organization),
      DashboardBackend.getDashboardUsersByProvider(organization),
      DashboardBackend.getDashboardLoginHeatmap(organization),
      DashboardBackend.getDashboardMfaCoverage(organization),
    ]).then(([dashboardRes, usersByProviderRes, loginHeatmapRes, mfaCoverageRes]) => {
      const handleResult = (result, setter) => {
        if (result.status === "fulfilled") {
          if (result.value?.status === "ok") {
            setter(result.value.data);
          } else {
            Setting.showMessage("error", result.value?.msg ?? "Request failed");
          }
        } else {
          Setting.showMessage("error", result.reason?.message ?? "Request failed");
        }
      };

      handleResult(dashboardRes, setDashboardData);
      handleResult(usersByProviderRes, setUsersByProvider);
      handleResult(loginHeatmapRes, setLoginHeatmap);
      handleResult(mfaCoverageRes, setMfaCoverage);
    });
  };

  React.useEffect(() => {
    if (!Setting.isLocalAdminUser(props.account)) {
      return;
    }

    const organization = getOrganizationName();
    fetchDashboardData(organization);
  }, [props.owner]);

  const handleTourChange = () => {
    setIsTourVisible(TourConfig.getTourVisible());
  };

  const handleOrganizationChange = () => {
    if (!Setting.isLocalAdminUser(props.account)) {
      return;
    }

    setDashboardData(null);
    setUsersByProvider(null);
    setLoginHeatmap(null);
    setMfaCoverage(null);

    const organization = getOrganizationName();
    fetchDashboardData(organization);
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

  const attachEChart = (chartDom, option, setOptionOpts) => {
    const myChart = echarts.getInstanceByDom(chartDom) ?? echarts.init(chartDom);
    myChart.setOption(option, setOptionOpts);
    myChart.resize();

    const handleResize = () => {
      myChart.resize();
    };
    window.addEventListener("resize", handleResize);

    return () => {
      window.removeEventListener("resize", handleResize);
      myChart.dispose();
    };
  };

  const ChartWidget = ({chartId, seriesConfig, height, yAxisName}) => {
    const seriesConfigKey = React.useMemo(() => JSON.stringify(seriesConfig), [seriesConfig]);

    React.useEffect(() => {
      if (dashboardData === null) {
        return;
      }

      const chartDom = document.getElementById(chartId);
      if (!chartDom) {
        return;
      }

      const dateArray = getDateArray();
      const dateLabel = i18next.t("Date", {defaultValue: "Date"});

      const option = {
        tooltip: {
          trigger: "axis",
          axisPointer: {type: "line"},
          confine: true,
          extraCssText: "pointer-events:none;",
          position: (point, params, dom, rect, size) => {
            if (!size || !Array.isArray(size.viewSize) || !Array.isArray(size.contentSize)) {
              return [point[0] + 12, point[1] + 12];
            }
            const [mouseX, mouseY] = point;
            const [viewWidth, viewHeight] = size.viewSize;
            const [boxWidth, boxHeight] = size.contentSize;

            let x = mouseX + 12;
            let y = mouseY - boxHeight - 12;

            if (x + boxWidth > viewWidth) {
              x = mouseX - boxWidth - 12;
            }
            if (x < 0) {
              x = 0;
            }

            if (y < 0) {
              y = mouseY + 12;
            }
            if (y + boxHeight > viewHeight) {
              y = Math.max(0, viewHeight - boxHeight - 12);
            }

            return [x, y];
          },
        },
        legend: {type: "scroll", data: seriesConfig.map(s => s.name), top: 0, left: "center", width: "90%", pageButtonPosition: "end"},
        grid: {left: 48, right: "4%", bottom: 28, top: 32, containLabel: true},
        xAxis: {type: "category", name: dateLabel, nameLocation: "middle", nameGap: 28, data: dateArray},
        yAxis: {type: "value", name: yAxisName, nameLocation: "middle", nameGap: 28, splitArea: {show: true}},
        series: seriesConfig.map(s => {
          return {
            name: s.name,
            type: "line",
            data: s.data ?? dashboardData[s.dataKey],
            smooth: true,
            ...(s.color ? {lineStyle: {color: s.color}, itemStyle: {color: s.color}} : {}),
          };
        }),
      };

      return attachEChart(chartDom, option);
    }, [dashboardData, chartId, height, seriesConfigKey, yAxisName]);

    return <div id={chartId} role="img" aria-label={chartId} style={{width: "100%", height}} />;
  };

  const MfaCoverageWidget = ({chartId, height}) => {
    React.useEffect(() => {
      if (mfaCoverage === null) {
        return;
      }

      const chartDom = document.getElementById(chartId);
      if (!chartDom) {
        return;
      }

      const items = Array.isArray(mfaCoverage) ? mfaCoverage : [];
      const organizations = items.map(i => i.organization);

      const adminDisabled = items.map(i => i.adminDisabled ?? 0);
      const userDisabled = items.map(i => i.userDisabled ?? 0);
      const adminEnabled = items.map(i => i.adminEnabled ?? 0);
      const userEnabled = items.map(i => i.userEnabled ?? 0);

      const adminLabel = i18next.t("ldap:Admin");
      const userLabel = i18next.t("general:User");
      const enabledLabel = i18next.t("general:Enabled");
      const disabledLabel = i18next.t("general:Disable");
      const enabledLabelWrapped = `(${enabledLabel})`;
      const disabledLabelWrapped = `(${disabledLabel})`;
      const organizationLabel = i18next.t("general:Organization");
      const totalUsersLabel = i18next.t("Total", {defaultValue: "Total"});

      const option = {
        tooltip: {
          trigger: "axis",
          axisPointer: {type: "shadow"},
        },
        legend: {
          top: 0,
          left: "center",
        },
        grid: {
          left: 48,
          right: "4%",
          bottom: 28,
          top: 32,
          containLabel: true,
        },
        xAxis: {type: "category", name: organizationLabel, nameLocation: "middle", nameGap: 28, data: organizations},
        yAxis: {type: "value", name: totalUsersLabel, nameLocation: "middle", nameGap: 36, minInterval: 1},
        series: [
          {name: `${adminLabel} ${disabledLabelWrapped}`, type: "bar", stack: "total", data: adminDisabled, itemStyle: {color: "#EE6666"}, barMaxWidth: 28},
          {name: `${userLabel} ${disabledLabelWrapped}`, type: "bar", stack: "total", data: userDisabled, itemStyle: {color: "#FC8452"}, barMaxWidth: 28},
          {name: `${adminLabel} ${enabledLabelWrapped}`, type: "bar", stack: "total", data: adminEnabled, itemStyle: {color: "#3BA272"}, barMaxWidth: 28},
          {name: `${userLabel} ${enabledLabelWrapped}`, type: "bar", stack: "total", data: userEnabled, itemStyle: {color: "#91CC75"}, barMaxWidth: 28},
        ],
      };

      return attachEChart(chartDom, option);
    }, [mfaCoverage, chartId, height]);

    return <div id={chartId} role="img" aria-label={chartId} style={{width: "100%", height}} />;
  };

  const DonutWidget = ({chartId, height}) => {
    React.useEffect(() => {
      if (usersByProvider === null) {
        return;
      }

      const chartDom = document.getElementById(chartId);
      if (!chartDom) {
        return;
      }

      const providerData = Object.entries(usersByProvider ?? {})
        .filter(([, value]) => Number(value) > 0)
        .sort((a, b) => Number(b[1]) - Number(a[1]))
        .map(([name, value]) => {
          const iconUrl = Setting.getProviderLogoURL({category: "OAuth", type: name});
          const baseItem = {
            name,
            value: Number(value),
          };

          if (!iconUrl) {
            return {
              ...baseItem,
              label: {show: true, position: "outside", formatter: "{b}"},
            };
          }

          return {
            ...baseItem,
            label: {
              show: true,
              position: "outside",
              formatter: "{icon|} {b}",
              rich: {
                icon: {
                  width: 12,
                  height: 12,
                  backgroundColor: {image: iconUrl},
                },
              },
            },
          };
        });

      const option = {
        title: {
          text: i18next.t("Identity Source", {defaultValue: "Identity Source"}),
          left: "center",
          bottom: 0,
          textStyle: {
            fontSize: 12,
            fontWeight: 400,
            color: "#8c8c8c",
          },
        },
        tooltip: {
          trigger: "item",
          formatter: (params) => `${params.name}: ${params.value} (${params.percent}%)`,
        },
        legend: {show: false},
        series: [
          {
            type: "pie",
            radius: ["0%", "90%"],
            center: ["50%", "46%"],
            labelLine: {show: true, length: 12, length2: 10},
            itemStyle: {borderColor: "#fff", borderWidth: 2},
            data: providerData,
          },
        ],
      };

      return attachEChart(chartDom, option);
    }, [usersByProvider, chartId, height]);

    return <div id={chartId} role="img" aria-label={chartId} style={{width: "100%", height}} />;
  };

  const HeatmapWidget = ({chartId, height}) => {
    React.useEffect(() => {
      if (loginHeatmap === null) {
        return;
      }

      const chartDom = document.getElementById(chartId);
      if (!chartDom) {
        return;
      }

      const xAxis = loginHeatmap.xAxis ?? [];
      const yAxis = loginHeatmap.yAxis ?? [];
      const data = loginHeatmap.data ?? [];
      const maxValue = Number(loginHeatmap.max ?? 0) || 0;

      const hourLabel = `${i18next.t("general:Timestamp")}/hour`;

      const option = {
        tooltip: {
          position: "top",
          formatter: (params) => {
            const hour = xAxis[params.value[0]] ?? params.value[0];
            const day = yAxis[params.value[1]] ?? params.value[1];
            return `${day} ${hour}:00<br/>${i18next.t("application:Signin")}${i18next.t("system:Count")}: ${params.value[2]}`;
          },
        },
        grid: {left: "4%", right: "4%", top: 12, bottom: 28, containLabel: true},
        xAxis: {type: "category", name: hourLabel, nameLocation: "middle", nameGap: 28, data: xAxis, splitArea: {show: true}},
        yAxis: {
          type: "category", zlevel: 1, z: 10, name: i18next.t("general:Date", {defaultValue: "Date"}), nameLocation: "middle", nameGap: 36, nameTextStyle: {fontSize: 11}, data: yAxis, splitArea: {show: true},
        },
        visualMap: {
          show: false,
          min: 0,
          max: Math.max(1, maxValue),
          inRange: {color: ["#f5f5f5", "#91CC75", "#3ba272"]},
        },
        series: [
          {
            type: "heatmap",
            data: data,
            zlevel: 0,
            z: 0,
            label: {
              show: true,
              formatter: (params) => {
                const v = Array.isArray(params.value) ? params.value[2] : params.value;
                return v ? `${v}` : "";
              },
              fontSize: 10,
              color: "#000",
            },
            emphasis: {
              itemStyle: {
                shadowBlur: 10,
                shadowColor: "rgba(0, 0, 0, 0.3)",
              },
            },
          },
        ],
      };

      return attachEChart(chartDom, option, {notMerge: true});
    }, [loginHeatmap, chartId, height]);

    return <div id={chartId} role="img" aria-label={chartId} style={{width: "100%", height}} />;
  };

  const renderStatistics = () => {
    if (dashboardData === null) {
      return (
        <div style={{display: "flex", justifyContent: "center", alignItems: "center", padding: "40px"}}>
          <Spin size="large" tip={i18next.t("login:Loading")} />
        </div>
      );
    }

    return null;
  };

  const colSpan = 12;

  return (
    <div style={{padding: `${padding}px`, maxWidth: "1400px", margin: "0 auto", overflow: "hidden"}}>
      {renderStatistics()}

      <Row gutter={[gutter, gutter]}>
        <Col xs={colSpan}>
          <Card title={i18next.t("home:Past 30 Days")} size="small" style={{height: "100%"}} styles={{body: {padding: 12}}}>
            <ChartWidget chartId="past-30-days-chart" height={chartHeight} yAxisName={i18next.t("Total", {defaultValue: "Total"})}
              seriesConfig={[
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
              ]}
            />
          </Card>
        </Col>

        <Col xs={colSpan}>
          <Card title={i18next.t("general:MFA items")} size="small" style={{height: "100%"}} styles={{body: {padding: 12}}}>
            <MfaCoverageWidget chartId="mfa-coverage-chart" height={chartHeight} />
          </Card>
        </Col>

        <Col xs={colSpan}>
          <Card title={i18next.t("Identity Provider Distribution", {defaultValue: "Identity Provider Distribution"})} size="small" style={{height: "100%"}} styles={{body: {padding: 12}}}>
            <DonutWidget chartId="users-by-provider-chart" height={chartHeight} />
          </Card>
        </Col>

        <Col xs={colSpan}>
          <Card title={i18next.t("general:Records")} size="small" style={{height: "100%"}} styles={{body: {padding: 12}}}>
            <HeatmapWidget chartId="login-time-heatmap" height={chartHeight} />
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
