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

import React, {Suspense, lazy} from "react";
import {Button, Col, Input, Row, Table, Upload} from "antd";
import i18next from "i18next";
import * as Setting from "../Setting";
import {UploadOutlined} from "@ant-design/icons";
import * as ResourceBackend from "../backend/ResourceBackend";
const FaceRecognitionModal = lazy(() => import("../common/modal/FaceRecognitionModal"));

class FaceIdTable extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      openFaceRecognitionModal: false,
    };
  }

  updateTable(table) {
    this.props.onUpdateTable(table);
  }

  updateField(table, index, key, value) {
    table[index][key] = value;
    this.updateTable(table);
  }

  deleteRow(table, i) {
    table = Setting.deleteRow(table, i);
    this.updateTable(table);
  }

  addFaceId(table, faceIdData) {
    const faceId = {
      name: Setting.getRandomName(),
      faceIdData: faceIdData,
    };
    if (table === undefined || table === null) {
      table = [];
    }
    table = Setting.addRow(table, faceId);
    this.updateTable(table);
  }

  addFaceImage(table, imageUrl) {
    const faceId = {
      name: Setting.getRandomName(),
      imageUrl: imageUrl,
      faceIdData: [],
    };
    if (table === undefined || table === null) {
      table = [];
    }
    table = Setting.addRow(table, faceId);
    this.updateTable(table);
  }

  renderTable(table) {
    const columns = [
      {
        title: i18next.t("general:Name"),
        dataIndex: "name",
        key: "name",
        width: "200px",
        render: (text, record, index) => {
          return (
            <Input defaultValue={text} onChange={e => {
              this.updateField(table, index, "name", e.target.value);
            }} />
          );
        },
      },
      {
        title: i18next.t("general:Data"),
        dataIndex: "faceIdData",
        key: "faceIdData",
        render: (text, record, index) => {
          const front = text.slice(0, 3).join(", ");
          const back = text.slice(-3).join(", ");
          return "[" + front + " ... " + back + "]";
        },
      },
      {
        title: i18next.t("general:URL"),
        dataIndex: "imageUrl",
        key: "imageUrl",
        render: (text, record, index) => {
          return text;
        },
      },
      {
        title: i18next.t("general:Action"),
        key: "action",
        width: "100px",
        render: (text, record, index) => {
          return (
            <Button style={{marginTop: "5px", marginBottom: "5px", marginRight: "5px"}} type="primary" danger onClick={() => {this.deleteRow(table, index);}}>
              {i18next.t("general:Delete")}
            </Button>
          );
        },
      },
    ];

    const handleUpload = (info) => {
      this.setState({uploading: true});
      const filename = info.fileList[0].name;
      const fullFilePath = `resource/${this.props.account.owner}/${this.props.account.name}/${filename}`;
      ResourceBackend.uploadResource(this.props.account.owner, this.props.account.name, "custom", "ResourceListPage", fullFilePath, info.file)
        .then(res => {
          if (res.status === "ok") {
            Setting.showMessage("success", i18next.t("application:File uploaded successfully"));

            this.addFaceImage(table, res.data);
          } else {
            Setting.showMessage("error", res.msg);
          }
        }).finally(() => {
          this.setState({uploading: false});
        });
    };

    return (
      <Table scroll={{x: "max-content"}} columns={columns} dataSource={this.props.table} size="middle" bordered pagination={false}
        title={() => (
          <div>
            {i18next.t("user:Face IDs")}&nbsp;&nbsp;&nbsp;&nbsp;
            <Button disabled={this.props.table?.length >= 5} style={{marginRight: "5px"}} type="primary" size="small" onClick={() => this.setState({openFaceRecognitionModal: true, withImage: false})}>
              {i18next.t("application:Add Face ID")}
            </Button>
            <Button disabled={this.props.table?.length >= 5} style={{marginRight: "5px"}} type="primary" size="small" onClick={() => this.setState({openFaceRecognitionModal: true, withImage: true})}>
              {i18next.t("application:Add Face ID with Image")}
            </Button>
            <Upload maxCount={1} accept="image/*" showUploadList={false}
              beforeUpload={file => {return false;}} onChange={info => {handleUpload(info);}}>
              <Button id="upload-button" icon={<UploadOutlined />} loading={this.state.uploading} type="primary" size="small">
                {i18next.t("resource:Upload a file...")}
              </Button>
            </Upload>
            <Suspense fallback={null}>
              <FaceRecognitionModal
                visible={this.state.openFaceRecognitionModal}
                withImage={this.state.withImage}
                onOk={(faceIdData) => {
                  this.addFaceId(table, faceIdData);
                  this.setState({openFaceRecognitionModal: false});
                }}
                onCancel={() => this.setState({openFaceRecognitionModal: false})}
              />
            </Suspense>
          </div>
        )}
      />
    );
  }

  render() {
    return (
      <div>
        <Row style={{marginTop: "20px"}}>
          <Col span={24}>
            {
              this.renderTable(this.props.table)
            }
          </Col>
        </Row>
      </div>
    );
  }
}

export default FaceIdTable;
