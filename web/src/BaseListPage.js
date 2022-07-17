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
import {Button, Input, Space} from "antd";
import {SearchOutlined} from "@ant-design/icons";
import Highlighter from "react-highlight-words";

class BaseListPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      data: [],
      pagination: {
        current: 1,
        pageSize: 10,
      },
      loading: false,
      searchText: "",
      searchedColumn: "",
      isAuthenticated: false,
    };
  }

  UNSAFE_componentWillMount() {
    const {pagination} = this.state;
    this.fetch({pagination});
  }

    getColumnSearchProps = dataIndex => ({
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
      onFilterDropdownVisibleChange: visible => {
        if (visible) {
          setTimeout(() => this.searchInput.select(), 100);
        }
      },
      render: text =>
        this.state.searchedColumn === dataIndex ? (
          <Highlighter
            highlightStyle={{backgroundColor: "#ffc069", padding: 0}}
            searchWords={[this.state.searchText]}
            autoEscape
            textToHighlight={text ? text.toString() : ""}
          />
        ) : (
          text
        ),
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

    render() {
      return (
        <div>
          {
            this.renderTable(this.state.data)
          }
        </div>
      );
    }
}

export default BaseListPage;
