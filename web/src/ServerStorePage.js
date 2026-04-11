// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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
import {Button, Card, Col, Empty, Input, Row, Select, Spin, Tag, Typography} from "antd";
import moment from "moment";
import * as Setting from "./Setting";
import * as ServerBackend from "./backend/ServerBackend";
import i18next from "i18next";

const {Text, Title} = Typography;

class ServerStorePage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      onlineListLoading: false,
      onlineServerList: [],
      creatingOnlineServerId: "",
      onlineNameFilter: "",
      onlineCategoryFilter: [],
    };
  }

  componentDidMount() {
    this.fetchOnlineServers();
  }

  fetchOnlineServers = () => {
    this.setState({
      onlineListLoading: true,
      onlineNameFilter: "",
      onlineCategoryFilter: [],
    });

    ServerBackend.getOnlineServers()
      .then((res) => {
        if (res.status === "ok") {
          const onlineServerList = this.normalizeOnlineServers(this.getOnlineServersFromResponse(res.data));
          this.setState({
            onlineServerList: onlineServerList,
            onlineListLoading: false,
          });
        } else {
          this.setState({onlineListLoading: false});
          Setting.showMessage("error", `${i18next.t("general:Failed to get")}: ${res.msg}`);
        }
      })
      .catch(error => {
        this.setState({onlineListLoading: false});
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  };

  getOnlineServerName = (onlineServer) => {
    const source = onlineServer.id || onlineServer.name || `server_${Setting.getRandomName()}`;
    const normalized = String(source).toLowerCase().replace(/[^a-z0-9_-]/g, "_").replace(/_+/g, "_").replace(/^_+|_+$/g, "");
    return normalized || `server_${Setting.getRandomName()}`;
  };

  createServerFromOnline = (onlineServer) => {
    const owner = Setting.getRequestOrganization(this.props.account);
    const serverName = this.getOnlineServerName(onlineServer);
    const serverUrl = onlineServer.endpoint;

    if (!serverUrl) {
      Setting.showMessage("error", i18next.t("server:Production endpoint is empty"));
      return;
    }

    const randomName = Setting.getRandomName();
    const newServer = {
      owner: owner,
      name: serverName + randomName,
      createdTime: moment().format(),
      displayName: onlineServer.name || serverName,
      url: serverUrl,
      application: "",
    };

    this.setState({creatingOnlineServerId: onlineServer.id});
    ServerBackend.addServer(newServer)
      .then((res) => {
        this.setState({creatingOnlineServerId: ""});
        if (res.status === "ok") {
          this.props.history.push({pathname: `/servers/${newServer.owner}/${newServer.name}`, mode: "add"});
          Setting.showMessage("success", i18next.t("general:Successfully added"));
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to add")}: ${res.msg}`);
        }
      })
      .catch(error => {
        this.setState({creatingOnlineServerId: ""});
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  };

  normalizeOnlineServers = (onlineServers) => {
    return onlineServers.map((server, index) => {
      const categoriesRaw = [server?.category].filter((category) => typeof category === "string" && category.trim() !== "");

      return {
        id: server.id ?? `${server.name ?? "server"}-${index}`,
        name: server.name ?? "",
        nameText: (server.name ?? "").toLowerCase(),
        categoriesRaw: categoriesRaw,
        categoriesLower: categoriesRaw.map((category) => category.toLowerCase()),
        endpoint: server.endpoints?.production ?? server.endpoint ?? "",
        description: server.description ?? "",
        website: server?.maintainer?.website ?? server?.website,
      };
    }).filter(server => server.endpoint.startsWith("http"));
  };

  getWebsiteUrl = (website) => {
    if (!website) {
      return "";
    }

    return /^https?:\/\//i.test(website) ? website : `https://${website}`;
  };

  getOnlineServersFromResponse = (data) => {
    if (Array.isArray(data?.servers)) {
      return data.servers;
    }

    if (Array.isArray(data)) {
      return data;
    }

    if (Array.isArray(data?.data)) {
      return data.data;
    }

    return [];
  };

  getOnlineCategoryOptions = () => {
    const categories = this.state.onlineServerList.flatMap((server) => server.categoriesRaw || []);
    return [...new Set(categories)].sort((a, b) => a.localeCompare(b)).map((category) => ({label: category, value: category.toLowerCase()}));
  };

  getFilteredOnlineServers = () => {
    const nameFilter = this.state.onlineNameFilter.trim().toLowerCase();
    const categoryFilter = this.state.onlineCategoryFilter;

    return this.state.onlineServerList.filter((server) => {
      const nameMatched = !nameFilter || server.nameText.includes(nameFilter);
      const categoryMatched = categoryFilter.length === 0 || categoryFilter.some((category) => server.categoriesLower.includes(category));
      return nameMatched && categoryMatched;
    });
  };

  renderServerCard = (server) => {
    return (
      <Col xs={24} sm={12} md={8} lg={6} key={server.id} style={{marginBottom: "16px"}}>
        <Card
          title={server.name || "-"}
          hoverable
          style={{height: "100%"}}
          extra={
            <Button
              type="primary"
              size="small"
              loading={this.state.creatingOnlineServerId === server.id}
              onClick={(e) => {
                e.stopPropagation();
                this.createServerFromOnline(server);
              }}
            >
              {i18next.t("general:Add")}
            </Button>
          }
        >
          <div style={{minHeight: "48px", marginBottom: "8px"}}>
            <Text type="secondary">{server.description || "-"}</Text>
          </div>
          <div style={{marginBottom: "8px"}}>
            <Text strong>{i18next.t("general:Url")}: </Text>
            {server.website ? (
              <a target="_blank" rel="noreferrer" href={this.getWebsiteUrl(server.endpoint)}>{server.endpoint}</a>
            ) : (
              <Text>-</Text>
            )}
          </div>
          <div style={{marginBottom: "8px"}}>
            <Text strong>{i18next.t("general:Website")}: </Text>
            {server.website ? (
              <a target="_blank" rel="noreferrer" href={this.getWebsiteUrl(server.website)}>{server.website}</a>
            ) : (
              <Text>-</Text>
            )}
          </div>
          <div>
            {(server.categoriesRaw || []).map((category) => <Tag key={`${server.id}-${category}`}>{category}</Tag>)}
          </div>
        </Card>
      </Col>
    );
  };

  render() {
    const filteredServers = this.getFilteredOnlineServers();

    return (
      <div style={{padding: "16px"}}>
        <div style={{display: "flex", gap: "8px", marginBottom: "12px"}}>
          <Input
            allowClear
            placeholder={i18next.t("general:Name")}
            value={this.state.onlineNameFilter}
            onChange={(e) => this.setState({onlineNameFilter: e.target.value})}
          />
          <Select
            mode="multiple"
            allowClear
            placeholder={i18next.t("general:Category")}
            value={this.state.onlineCategoryFilter}
            onChange={(values) => this.setState({onlineCategoryFilter: values})}
            options={this.getOnlineCategoryOptions()}
            style={{minWidth: "260px"}}
          />
          <Button onClick={() => this.setState({onlineNameFilter: "", onlineCategoryFilter: []})}>
            {i18next.t("general:Clear")}
          </Button>
          <Button onClick={this.fetchOnlineServers}>
            {i18next.t("general:Refresh")}
          </Button>
        </div>
        <Title level={4} style={{marginBottom: "12px"}}>{i18next.t("general:MCP Store")}</Title>
        {this.state.onlineListLoading ? (
          <div style={{textAlign: "center", padding: "36px 0"}}>
            <Spin />
          </div>
        ) : filteredServers.length === 0 ? (
          <Empty description={i18next.t("general:No data")} />
        ) : (
          <Row gutter={16}>
            {filteredServers.map((server) => this.renderServerCard(server))}
          </Row>
        )}
      </div>
    );
  }
}

export default ServerStorePage;
