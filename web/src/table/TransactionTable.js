// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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
import {Button, Input, Space, Table} from "antd";
import {SearchOutlined} from "@ant-design/icons";
import Highlighter from "react-highlight-words";
import i18next from "i18next";
import {getTransactionTableColumns} from "./TransactionTableColumns";

class TransactionTable extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      searchText: "",
      searchedColumn: "",
    };
  }

  getColumnSearchProps = (dataIndex) => ({
    filterDropdown: ({setSelectedKeys, selectedKeys, confirm, clearFilters}) => (
      <div style={{padding: 8}}>
        <Input
          ref={node => {
            this.searchInput = node;
          }}
          placeholder={i18next.t("general:Please input your search")}
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
            {i18next.t("general:Search")}
          </Button>
          <Button onClick={() => this.handleReset(clearFilters)} size="small" style={{width: 90}}>
            {i18next.t("forget:Reset")}
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
            {i18next.t("general:Filter")}
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

      return highlightContent;
    },
  });

  handleSearch = (selectedKeys, confirm, dataIndex) => {
    confirm();
    this.setState({
      searchText: selectedKeys[0],
      searchedColumn: dataIndex,
    });
  };

  handleReset = clearFilters => {
    clearFilters();
    this.setState({searchText: ""});
  };

  render() {
    const columns = getTransactionTableColumns({
      includeOrganization: false,
      includeUser: this.props.includeUser || false,
      includeTag: !this.props.hideTag,
      includeActions: false,
      getColumnSearchProps: this.getColumnSearchProps,
      account: null,
      onEdit: null,
      onDelete: null,
    });

    return (
      <Table
        scroll={{x: "max-content"}}
        columns={columns}
        dataSource={this.props.transactions}
        rowKey={(record) => `${record.owner}/${record.name}`}
        size="small"
        bordered
        pagination={{
          pageSize: 10,
          showSizeChanger: true,
          pageSizeOptions: ["10", "20", "50", "100"],
        }}
      />
    );
  }
}

export default TransactionTable;
