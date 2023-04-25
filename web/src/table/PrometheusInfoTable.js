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

import React from "react";
import {Table} from "antd";

class PrometheusInfoTable extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      table: props.table,
    };
  }
  render() {
    const latencyColumns = [
      {
        title: "Name",
        dataIndex: "name",
        key: "name",
      },
      {
        title: "Method",
        dataIndex: "method",
        key: "method",
      },
      {
        title: "Count",
        dataIndex: "count",
        key: "count",
      },
      {
        title: "Latency(ms)",
        dataIndex: "latency",
        key: "latency",
      },
    ];
    const throughputColumns = [
      {
        title: "Name",
        dataIndex: "name",
        key: "name",
      },
      {
        title: "Method",
        dataIndex: "method",
        key: "method",
      },
      {
        title: "Throughput",
        dataIndex: "throughput",
        key: "throughput",
      },
    ];
    if (this.state.table === "latency") {
      return (
        <div style={{height: "300px", overflow: "auto"}}>
          <Table columns={latencyColumns} dataSource={this.props.prometheusInfo.apiLatency} pagination={false} />
        </div>
      );
    } else if (this.state.table === "throughput") {
      return (
        <div style={{height: "300px", overflow: "auto"}}>
            Total Throughput: {this.props.prometheusInfo.totalThroughput}
          <Table columns={throughputColumns} dataSource={this.props.prometheusInfo.apiThroughput} pagination={false} />
        </div>
      );
    }
  }
}

export default PrometheusInfoTable;
