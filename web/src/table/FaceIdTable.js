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
import {Button, Col, Input, Row, Table} from "antd";
import i18next from "i18next";
import * as Setting from "../Setting";
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
        title: i18next.t("general:FaceIdData"),
        dataIndex: "faceIdData",
        key: "faceIdData",
        render: (text, record, index) => {
          const front = text.slice(0, 3).join(", ");
          const back = text.slice(-3).join(", ");
          return "[" + front + " ... " + back + "]";
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

    return (
      <Table scroll={{x: "max-content"}} columns={columns} dataSource={this.props.table} size="middle" bordered pagination={false}
        title={() => (
          <div>
            {i18next.t("user:Face IDs")}&nbsp;&nbsp;&nbsp;&nbsp;
            <Button disabled={this.props.table?.length >= 5} style={{marginRight: "5px"}} type="primary" size="small" onClick={() => this.setState({openFaceRecognitionModal: true})}>
              {i18next.t("general:Add Face Id")}
            </Button>
            <Suspense fallback={null}>
              <FaceRecognitionModal
                visible={this.state.openFaceRecognitionModal}
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
