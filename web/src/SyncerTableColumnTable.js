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
import {DownOutlined, DeleteOutlined, UpOutlined} from '@ant-design/icons';
import {Button, Col, Input, Row, Select, Table, Tooltip} from 'antd';
import * as Setting from "./Setting";
import i18next from "i18next";

const { Option } = Select;

class SyncerTableColumnTable extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
    };
  }

  updateTable(table) {
    this.props.onUpdateTable(table);
  }

  updateField(table, index, key, value) {
    table[index][key] = value;
    this.updateTable(table);
  }

  addRow(table) {
    let row = {name: `column${table.length}`, type: "string", values: []};
    if (table === undefined) {
      table = [];
    }
    table = Setting.addRow(table, row);
    this.updateTable(table);
  }

  deleteRow(table, i) {
    table = Setting.deleteRow(table, i);
    this.updateTable(table);
  }

  upRow(table, i) {
    table = Setting.swapRow(table, i - 1, i);
    this.updateTable(table);
  }

  downRow(table, i) {
    table = Setting.swapRow(table, i, i + 1);
    this.updateTable(table);
  }

  renderTable(table) {
    const columns = [
      {
        title: i18next.t("syncer:Column name"),
        dataIndex: 'name',
        key: 'name',
        render: (text, record, index) => {
          return (
            <Input value={text} onChange={e => {
              this.updateField(table, index, 'name', e.target.value);
            }} />
          )
        }
      },
      {
        title: i18next.t("syncer:Column type"),
        dataIndex: 'type',
        key: 'type',
        render: (text, record, index) => {
          return (
            <Select virtual={false} style={{width: '100%'}} value={text} onChange={(value => {this.updateField(table, index, 'type', value);})}>
              {
                ['string', 'integer', 'boolean']
                  .map((item, index) => <Option key={index} value={item}>{item}</Option>)
              }
            </Select>
          )
        }
      },
      {
        title: i18next.t("syncer:Casdoor column"),
        dataIndex: 'casdoorName',
        key: 'casdoorName',
        render: (text, record, index) => {
          return (
            <Select virtual={false} style={{width: '100%'}} value={text} onChange={(value => {this.updateField(table, index, 'casdoorName', value);})}>
              {
                ['Name', 'CreatedTime', 'UpdatedTime', 'Id', 'Type', 'Password', 'PasswordSalt', 'DisplayName', 'Avatar', 'PermanentAvatar', 'Email', 'Phone',
                  'Location', 'Address', 'Affiliation', 'Title', 'IdCardType', 'IdCard', 'Homepage', 'Bio', 'Tag', 'Region', 'Language', 'Gender', 'Birthday',
                  'Education', 'Score', 'Ranking', 'IsDefaultAvatar', 'IsOnline', 'IsAdmin', 'IsGlobalAdmin', 'IsForbidden', 'IsDeleted', 'CreatedIp']
                  .map((item, index) => <Option key={index} value={item}>{item}</Option>)
              }
            </Select>
          )
        }
      },
      {
        title: i18next.t("general:Action"),
        key: 'action',
        width: '100px',
        render: (text, record, index) => {
          return (
            <div>
              <Tooltip placement="bottomLeft" title={i18next.t("general:Up")}>
                <Button style={{marginRight: "5px"}} disabled={index === 0} icon={<UpOutlined />} size="small" onClick={() => this.upRow(table, index)} />
              </Tooltip>
              <Tooltip placement="topLeft" title={i18next.t("general:Down")}>
                <Button style={{marginRight: "5px"}} disabled={index === table.length - 1} icon={<DownOutlined />} size="small" onClick={() => this.downRow(table, index)} />
              </Tooltip>
              <Tooltip placement="topLeft" title={i18next.t("general:Delete")}>
                <Button icon={<DeleteOutlined />} size="small" onClick={() => this.deleteRow(table, index)} />
              </Tooltip>
            </div>
          );
        }
      },
    ];

    return (
      <Table rowKey="index" columns={columns} dataSource={table} size="middle" bordered pagination={false}
             title={() => (
               <div>
                 {this.props.title}&nbsp;&nbsp;&nbsp;&nbsp;
                 <Button style={{marginRight: "5px"}} type="primary" size="small" onClick={() => this.addRow(table)}>{i18next.t("general:Add")}</Button>
               </div>
             )}
      />
    );
  }

  render() {
    return (
      <div>
        <Row style={{marginTop: '20px'}} >
          <Col span={24}>
            {
              this.renderTable(this.props.table)
            }
          </Col>
        </Row>
      </div>
    )
  }
}

export default SyncerTableColumnTable;
