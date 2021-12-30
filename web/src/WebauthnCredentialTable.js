import React from "react";
import {Button, Table} from 'antd';
import i18next from "i18next";
import * as UserWebauthnBackend from "./backend/UserWebauthnBackend";
import * as Setting from "./Setting";

class WebAuthnCredentialTable extends React.Component {

  render() {
    const datasource = this.props.table
    const columns = [
      {
        title: i18next.t("user:WebAuthn credentials"),
        dataIndex: 'ID',
        key: 'ID',
      },
      {
        title: i18next.t("general:Action"),
        key: 'action',
        render: (text, record, index) => {
          return (<Button style={{marginTop: '5px', marginBottom: '5px', marginRight: '5px'}} type="danger" onClick={() => {
            this.deleteRow(this.props.table, index)
          }}>{i18next.t("general:Delete")}</Button>)
        }
      }
    ]
    return (
      <Table scroll={{x: 'max-content'}} rowKey={record => record.ID} columns={columns} dataSource={datasource} size="middle" bordered pagination={false}
        title={() => (
          <div>
            {i18next.t("user:WebAuthn credentials")}&nbsp;&nbsp;&nbsp;&nbsp;
            <Button style={{marginRight: "5px"}} type="primary" size="small" onClick={() => {this.registerWebAuthn()}}>{i18next.t("general:Add")}</Button>
          </div>
        )}
      />)
  }

  deleteRow(table, i) {
    table = Setting.deleteRow(table, i);
    this.props.updateTable(table);
  }


  registerWebAuthn() {
    UserWebauthnBackend.registerWebauthnCredential().then((res) => {
      if (res.msg === "") {
        Setting.showMessage("success", `Successfully added webauthn credentials`);
      } else {
        Setting.showMessage("error", res.msg);
      }
      this.props.refresh()
    }).catch(error => {
      Setting.showMessage("error", `Failed to connect to server: ${error}`);
    });
  }
}

export default WebAuthnCredentialTable;