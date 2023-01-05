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
import {Button, Popconfirm, Result, Table, Upload} from "antd";
import {UploadOutlined} from "@ant-design/icons";
import copy from "copy-to-clipboard";
import * as Setting from "./Setting";
import * as ResourceBackend from "./backend/ResourceBackend";
import i18next from "i18next";
import {Link} from "react-router-dom";
import BaseListPage from "./BaseListPage";

class ResourceListPage extends BaseListPage {
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
      fileList: [],
      uploading: false,
      isAuthorized: true,
    };
  }

  deleteResource(i) {
    ResourceBackend.deleteResource(this.state.data[i])
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully deleted"));
          this.setState({
            data: Setting.deleteRow(this.state.data, i),
            pagination: {total: this.state.pagination.total - 1},
          });
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to delete")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  handleUpload(info) {
    this.setState({uploading: true});
    const filename = info.fileList[0].name;
    const fullFilePath = `resource/${this.props.account.owner}/${this.props.account.name}/${filename}`;
    ResourceBackend.uploadResource(this.props.account.owner, this.props.account.name, "custom", "ResourceListPage", fullFilePath, info.file)
      .then(res => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("application:File uploaded successfully"));
          window.location.reload();
        } else {
          Setting.showMessage("error", res.msg);
        }
      }).finally(() => {
        this.setState({uploading: false});
      });
  }

  renderUpload() {
    return (
      <Upload maxCount={1} accept="image/*,video/*,audio/*,.pdf,.doc,.docx,.csv,.xls,.xlsx" showUploadList={false}
        beforeUpload={file => {return false;}} onChange={info => {this.handleUpload(info);}}>
        <Button icon={<UploadOutlined />} loading={this.state.uploading} type="primary" size="small">
          {i18next.t("resource:Upload a file...")}
        </Button>
      </Upload>
    );
  }

  renderTable(resources) {
    const columns = [
      {
        title: i18next.t("general:Provider"),
        dataIndex: "provider",
        key: "provider",
        width: "150px",
        fixed: "left",
        sorter: true,
        ...this.getColumnSearchProps("provider"),
        render: (text, record, index) => {
          return (
            <Link to={`/providers/${record.owner}/${text}`}>
              {text}
            </Link>
          );
        },
      },
      {
        title: i18next.t("resource:Application"),
        dataIndex: "application",
        key: "application",
        width: "80px",
        sorter: true,
        ...this.getColumnSearchProps("application"),
        render: (text, record, index) => {
          return (
            <Link to={`/applications/${record.organization}/${text}`}>
              {text}
            </Link>
          );
        },
      },
      {
        title: i18next.t("resource:User"),
        dataIndex: "user",
        key: "user",
        width: "80px",
        sorter: true,
        ...this.getColumnSearchProps("user"),
        render: (text, record, index) => {
          return (
            <Link to={`/users/${record.owner}/${record.user}`}>
              {text}
            </Link>
          );
        },
      },
      {
        title: i18next.t("resource:Parent"),
        dataIndex: "parent",
        key: "parent",
        width: "80px",
        sorter: true,
        ...this.getColumnSearchProps("parent"),
      },
      {
        title: i18next.t("general:Name"),
        dataIndex: "name",
        key: "name",
        width: "150px",
        sorter: true,
        ...this.getColumnSearchProps("name"),
      },
      {
        title: i18next.t("general:Created time"),
        dataIndex: "createdTime",
        key: "createdTime",
        width: "150px",
        sorter: true,
        render: (text, record, index) => {
          return Setting.getFormattedDate(text);
        },
      },
      {
        title: i18next.t("resource:Tag"),
        dataIndex: "tag",
        key: "tag",
        width: "80px",
        sorter: true,
        ...this.getColumnSearchProps("tag"),
      },
      // {
      //   title: i18next.t("resource:File name"),
      //   dataIndex: 'fileName',
      //   key: 'fileName',
      //   width: '120px',
      //   sorter: (a, b) => a.fileName.localeCompare(b.fileName),
      // },
      {
        title: i18next.t("resource:Type"),
        dataIndex: "fileType",
        key: "fileType",
        width: "80px",
        sorter: true,
        ...this.getColumnSearchProps("fileType"),
      },
      {
        title: i18next.t("resource:Format"),
        dataIndex: "fileFormat",
        key: "fileFormat",
        width: "80px",
        sorter: true,
        ...this.getColumnSearchProps("fileFormat"),
      },
      {
        title: i18next.t("resource:File size"),
        dataIndex: "fileSize",
        key: "fileSize",
        width: "100px",
        sorter: true,
        render: (text, record, index) => {
          return Setting.getFriendlyFileSize(text);
        },
      },
      {
        title: i18next.t("general:Preview"),
        dataIndex: "preview",
        key: "preview",
        width: "100px",
        render: (text, record, index) => {
          if (record.fileType === "image") {
            return (
              <a target="_blank" rel="noreferrer" href={record.url}>
                <img src={record.url} alt={record.name} width={200} />
              </a>
            );
          } else if (record.fileType === "video") {
            return (
              <video width={200} controls>
                <source src={record.url} type="video/mp4" />
              </video>
            );
          }
        },
      },
      {
        title: i18next.t("general:URL"),
        dataIndex: "url",
        key: "url",
        width: "120px",
        render: (text, record, index) => {
          return (
            <div>
              <Button type="normal" onClick={() => {
                copy(record.url);
                Setting.showMessage("success", i18next.t("resource:Link copied to clipboard successfully"));
              }}
              >
                {i18next.t("resource:Copy Link")}
              </Button>
            </div>
          );
        },
      },
      {
        title: i18next.t("general:Action"),
        dataIndex: "",
        key: "op",
        width: "70px",
        fixed: (Setting.isMobile()) ? "false" : "right",
        render: (text, record, index) => {
          return (
            <div>
              {/* <Button style={{marginTop: '10px', marginBottom: '10px', marginRight: '10px'}} type="primary" onClick={() => this.props.history.push(`/resources/${record.name}`)}>{i18next.t("general:Edit")}</Button>*/}
              <Popconfirm
                title={`Sure to delete resource: ${record.name} ?`}
                onConfirm={() => this.deleteResource(index)}
                okText={i18next.t("user:OK")}
                cancelText={i18next.t("user:Cancel")}
              >
                <Button type="primary" danger>{i18next.t("general:Delete")}</Button>
              </Popconfirm>
            </div>
          );
        },
      },
    ];

    const paginationProps = {
      total: this.state.pagination.total,
      showQuickJumper: true,
      showSizeChanger: true,
      showTotal: () => i18next.t("general:{total} in total").replace("{total}", this.state.pagination.total),
    };

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
        <Table scroll={{x: "max-content"}} columns={columns} dataSource={resources} rowKey="name" size="middle" bordered pagination={paginationProps}
          title={() => (
            <div>
              {i18next.t("general:Resources")}&nbsp;&nbsp;&nbsp;&nbsp;
              {/* <Button type="primary" size="small" onClick={this.addResource.bind(this)}>{i18next.t("general:Add")}</Button>*/}
              {
                this.renderUpload()
              }
            </div>
          )}
          loading={this.state.loading}
          onChange={this.handleTableChange}
        />
      </div>
    );
  }

  fetch = (params = {}) => {
    const field = params.searchedColumn, value = params.searchText;
    const sortField = params.sortField, sortOrder = params.sortOrder;
    this.setState({loading: true});
    ResourceBackend.getResources(this.props.account.owner, this.props.account.name, params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            loading: false,
            data: res.data,
            pagination: {
              ...params.pagination,
              total: res.data2,
            },
            searchText: params.searchText,
            searchedColumn: params.searchedColumn,
          });
        } else {
          if (res.msg.includes("Please login first")) {
            this.setState({
              loading: false,
              isAuthorized: false,
            });
          }
        }
      });
  };
}

export default ResourceListPage;
