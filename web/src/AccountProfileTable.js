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

import {Col, Row, Switch, Table} from "antd";
import React from "react";
import i18next from "i18next";

class AccountProfileTable extends React.Component {
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

  renderTable(table) {
    const columns = [
      {
        title: i18next.t("organization:Name"),
        dateIndex: 'name',
        key: 'name',
        render: (text, record, index) => {
          return (
            <p>{i18next.t("user:" + text.name)}</p>
          )
        }
      },
      {
        title: i18next.t("organization:Visible"),
        dataIndex: 'visible',
        key: 'visible',
        render: (text, record, index) => {
          return (
            <Switch checked={text} onChange={checked => {
              this.updateField(table, index, 'visible', checked);
            }} />
          )
        }
      },
    ];

    return (
      <Table scroll={{x: 'max-content'}} rowKey="name" columns={columns} dataSource={table} size="middle" bordered pagination={false}
             title={() => (
               <div>
                 {this.props.title}&nbsp;&nbsp;&nbsp;&nbsp;
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

export default AccountProfileTable;