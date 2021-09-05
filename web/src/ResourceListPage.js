// Copyright 2021 The casbin Authors. All Rights Reserved.
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
import {Button, Popconfirm, Table, Upload} from 'antd';
import {UploadOutlined} from "@ant-design/icons";
import copy from 'copy-to-clipboard';
import * as Setting from "./Setting";
import * as ResourceBackend from "./backend/ResourceBackend";
import i18next from "i18next";
import {Link} from "react-router-dom";

class ResourceListPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      resources: null,
      fileList: [],
      uploading: false,
    };
  }

  UNSAFE_componentWillMount() {
    this.getResources();
  }

  getResources() {
    ResourceBackend.getResources("admin")
      .then((res) => {
        this.setState({
          resources: res,
        });
      });
  }

  deleteResource(i) {
    ResourceBackend.deleteResource(this.state.resources[i])
      .then((res) => {
        Setting.showMessage("success", `Resource deleted successfully`);
          this.setState({
            resources: Setting.deleteRow(this.state.resources, i),
          });
        }
      )
      .catch(error => {
        Setting.showMessage("error", `Resource failed to delete: ${error}`);
      });
  }

  handleUpload(info) {
    this.setState({uploading: true});
    const filename = info.fileList[0].name;
    const fullFilePath = `resource/${this.props.account.owner}/${this.props.account.name}/${filename}`;
    ResourceBackend.uploadResource(this.props.account.owner, this.props.account.name, "custom", this.props.account.name, fullFilePath, info.file)
      .then(res => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("application:File uploaded successfully"));
          window.location.reload();
        } else {
          Setting.showMessage("error", res.msg);
        }
      }).finally(() => {
      this.setState({uploading: false});
    })
  }

  renderUpload() {
    return (
      <Upload maxCount={1} accept="image/*,video/*" showUploadList={false}
              beforeUpload={file => {return false}} onChange={info => {this.handleUpload(info)}}>
        <Button icon={<UploadOutlined />} loading={this.state.uploading} type="primary" size="small">
          {i18next.t("resource:Upload a file...")}
        </Button>
      </Upload>
    )
  }

  renderTable(resources) {
    const columns = [
      {
        title: i18next.t("general:Provider"),
        dataIndex: 'provider',
        key: 'provider',
        width: '150px',
        fixed: 'left',
        sorter: (a, b) => a.provider.localeCompare(b.provider),
        render: (text, record, index) => {
          return (
            <Link to={`/providers/${text}`}>
              {text}
            </Link>
          )
        }
      },
      {
        title: i18next.t("general:Name"),
        dataIndex: 'name',
        key: 'name',
        width: '150px',
        sorter: (a, b) => a.name.localeCompare(b.name),
      },
      {
        title: i18next.t("general:Created time"),
        dataIndex: 'createdTime',
        key: 'createdTime',
        width: '160px',
        sorter: (a, b) => a.createdTime.localeCompare(b.createdTime),
        render: (text, record, index) => {
          return Setting.getFormattedDate(text);
        }
      },
      {
        title: i18next.t("resource:Tag"),
        dataIndex: 'tag',
        key: 'tag',
        width: '80px',
        sorter: (a, b) => a.tag.localeCompare(b.tag),
      },
      {
        title: i18next.t("resource:User"),
        dataIndex: 'user',
        key: 'user',
        width: '80px',
        sorter: (a, b) => a.user.localeCompare(b.user),
        render: (text, record, index) => {
          return (
            <Link to={`/users/${record.owner}/${record.user}`}>
              {text}
            </Link>
          )
        }
      },
      {
        title: i18next.t("resource:Application"),
        dataIndex: 'application',
        key: 'application',
        width: '80px',
        sorter: (a, b) => a.application.localeCompare(b.application),
        render: (text, record, index) => {
          return (
            <Link to={`/applications/${text}`}>
              {text}
            </Link>
          )
        }
      },
      {
        title: i18next.t("resource:Parent"),
        dataIndex: 'parent',
        key: 'parent',
        width: '80px',
        sorter: (a, b) => a.parent.localeCompare(b.parent),
      },
      {
        title: i18next.t("resource:File name"),
        dataIndex: 'fileName',
        key: 'fileName',
        width: '120px',
        sorter: (a, b) => a.fileName.localeCompare(b.fileName),
      },
      {
        title: i18next.t("resource:File type"),
        dataIndex: 'fileType',
        key: 'fileType',
        width: '120px',
        sorter: (a, b) => a.fileType.localeCompare(b.fileType),
      },
      {
        title: i18next.t("resource:File format"),
        dataIndex: 'fileFormat',
        key: 'fileFormat',
        width: '130px',
        sorter: (a, b) => a.fileFormat.localeCompare(b.fileFormat),
      },
      {
        title: i18next.t("resource:File size"),
        dataIndex: 'fileSize',
        key: 'fileSize',
        width: '120px',
        sorter: (a, b) => a.fileSize - b.fileSize,
        render: (text, record, index) => {
          return Setting.getFriendlyFileSize(text);
        }
      },
      {
        title: i18next.t("general:Preview"),
        dataIndex: 'preview',
        key: 'preview',
        width: '100px',
        render: (text, record, index) => {
          if (record.fileType === "image") {
            return (
              <a target="_blank" href={record.url}>
                <img src={record.url} alt={record.name} width={100} />
              </a>
            )
          } else if (record.fileType === "video") {
            return (
              <div>
                <video width={100} controls>
                  <source src={text} type="video/mp4" />
                </video>
              </div>
            )
          }
        }
      },
      {
        title: i18next.t("general:URL"),
        dataIndex: 'url',
        key: 'url',
        width: '120px',
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
          )
        }
      },
      {
        title: i18next.t("general:Action"),
        dataIndex: '',
        key: 'op',
        width: '70px',
        fixed: (Setting.isMobile()) ? "false" : "right",
        render: (text, record, index) => {
          return (
            <div>
              {/*<Button style={{marginTop: '10px', marginBottom: '10px', marginRight: '10px'}} type="primary" onClick={() => this.props.history.push(`/resources/${record.name}`)}>{i18next.t("general:Edit")}</Button>*/}
              <Popconfirm
                title={`Sure to delete resource: ${record.name} ?`}
                onConfirm={() => this.deleteResource(index)}
                okText={i18next.t("user:OK")}
                cancelText={i18next.t("user:Cancel")}
              >
                <Button type="danger">{i18next.t("general:Delete")}</Button>
              </Popconfirm>
            </div>
          )
        }
      },
    ];

    return (
      <div>
        <Table scroll={{x: 'max-content'}} columns={columns} dataSource={resources} rowKey="name" size="middle" bordered pagination={{pageSize: 100}}
               title={() => (
                 <div>
                   {i18next.t("general:Resources")}&nbsp;&nbsp;&nbsp;&nbsp;
                   {/*<Button type="primary" size="small" onClick={this.addResource.bind(this)}>{i18next.t("general:Add")}</Button>*/}
                   {
                     this.renderUpload()
                   }
                 </div>
               )}
               loading={resources === null}
        />
      </div>
    );
  }

  render() {
    return (
      <div>
        {
          this.renderTable(this.state.resources)
        }
      </div>
    );
  }
}

export default ResourceListPage;
