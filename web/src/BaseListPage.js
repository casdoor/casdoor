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

import React from "react";
import {Button, Input, Result, Space, Tour} from "antd";
import {SearchOutlined} from "@ant-design/icons";
import Highlighter from "react-highlight-words";
import i18next from "i18next";
import * as Setting from "./Setting";
import * as TourConfig from "./TourConfig";

class BaseListPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      organizationName: this.props.match?.params.organizationName || Setting.getRequestOrganization(this.props.account),
      data: [],
      pagination: {
        current: 1,
        pageSize: 10,
      },
      loading: false,
      searchText: "",
      searchedColumn: "",
      isAuthorized: true,
      isTourVisible: TourConfig.getTourVisible(),
    };
  }

  handleOrganizationChange = () => {
    this.setState({
      organizationName: this.props.match?.params.organizationName || Setting.getRequestOrganization(this.props.account),
    },
    () => {
      const {pagination} = this.state;
      this.fetch({pagination});
    });
  };

  handleTourChange = () => {
    this.setState({isTourVisible: TourConfig.getTourVisible()});
  };

  componentDidMount() {
    window.addEventListener("storageOrganizationChanged", this.handleOrganizationChange);
    window.addEventListener("storageTourChanged", this.handleTourChange);
    if (!Setting.isAdminUser(this.props.account)) {
      Setting.setOrganization("All");
    }
  }

  componentWillUnmount() {
    if (this.state.intervalId !== null) {
      clearInterval(this.state.intervalId);
    }
    window.removeEventListener("storageTourChanged", this.handleTourChange);
    window.removeEventListener("storageOrganizationChanged", this.handleOrganizationChange);
  }

  UNSAFE_componentWillMount() {
    const {pagination} = this.state;
    this.fetch({pagination});
  }

  getColumnSearchProps = (dataIndex, customRender = null) => ({
    filterDropdown: ({setSelectedKeys, selectedKeys, confirm, clearFilters}) => (
      <div style={{padding: 8}}>
        <Input
          ref={node => {
            this.searchInput = node;
          }}
          placeholder={`Search ${dataIndex}`}
          value={selectedKeys[0]}
          onChange={e => setSelectedKeys(e.target.value ? [e.target.value] : [])}
          onPressEnter={() => this.handleSearch(selectedKeys, confirm, dataIndex)}
          style={{marginBottom: 8, display: "block"}}
        />

        <Space>
          <Button
            type="primary"
            onClick={() => this.handleSearch(selectedKeys, confirm, dataIndex)}
            icon={<SearchOutlined />}
            size="small"
            style={{width: 90}}
          >
                        Search
          </Button>
          <Button onClick={() => this.handleReset(clearFilters)} size="small" style={{width: 90}}>
                        Reset
          </Button>
          <Button
            type="link"
            size="small"
            onClick={() => {
              confirm({closeDropdown: false});
              this.setState({
                searchText: selectedKeys[0],
                searchedColumn: dataIndex,
              });
            }}
          >
                        Filter
          </Button>
        </Space>
      </div>
    ),
    filterIcon: filtered => <SearchOutlined style={{color: filtered ? "#1890ff" : undefined}} />,
    onFilter: (value, record) =>
      record[dataIndex]
        ? record[dataIndex].toString().toLowerCase().includes(value.toLowerCase())
        : "",
    filterDropdownProps: {
      onOpenChange: visible => {
        if (visible) {
          setTimeout(() => this.searchInput.select(), 100);
        }
      },
    },
    render: (text, record, index) => {
      const highlightContent = this.state.searchedColumn === dataIndex ? (
        <Highlighter
          highlightStyle={{backgroundColor: "#ffc069", padding: 0}}
          searchWords={[this.state.searchText]}
          autoEscape
          textToHighlight={text ? text.toString() : ""}
        />
      ) : (
        text
      );

      return customRender ? customRender({text, record, index}, highlightContent) : highlightContent;
    },
  });

  handleSearch = (selectedKeys, confirm, dataIndex) => {
    this.fetch({searchText: selectedKeys[0], searchedColumn: dataIndex, pagination: this.state.pagination});
  };

  handleReset = clearFilters => {
    clearFilters();
    const {pagination} = this.state;
    this.fetch({pagination});
  };

  handleTableChange = (pagination, filters, sorter) => {
    this.fetch({
      sortField: sorter.field,
      sortOrder: sorter.order,
      pagination,
      ...filters,
      searchText: this.state.searchText,
      searchedColumn: this.state.searchedColumn,
    });
  };

  setIsTourVisible = () => {
    TourConfig.setIsTourVisible(false);
    this.setState({isTourVisible: false});
  };

  getSteps = () => {
    const nextPathName = TourConfig.getNextUrl();
    const steps = TourConfig.getSteps();
    steps.map((item, index) => {
      if (!index) {
        item.target = () => document.querySelector(".ant-table");
      } else {
        item.target = () => document.getElementById(item.id) || null;
      }
      if (index === steps.length - 1) {
        item.nextButtonProps = {
          children: TourConfig.getNextButtonChild(nextPathName),
        };
      }
    });
    return steps;
  };

  handleTourComplete = () => {
    const nextPathName = TourConfig.getNextUrl();
    if (nextPathName !== "") {
      this.props.history.push("/" + nextPathName);
      TourConfig.setIsTourVisible(true);
    }
  };

  render() {
    if (!this.state.isAuthorized) {
      return (
        <Result
          status="403"
          title="403 Unauthorized"
          subTitle={i18next.t("general:Sorry, you do not have permission to access this page or logged in status invalid.")}
          extra={<a href="/"><Button type="primary">{i18next.t("general:Back Home")}</Button></a>}
        />
      );
    }

    return (
      <div>
        {
          this.renderTable(this.state.data)
        }
        <Tour
          open={Setting.isMobile() ? false : this.state.isTourVisible}
          onClose={this.setIsTourVisible}
          steps={this.getSteps()}
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

export default BaseListPage;
