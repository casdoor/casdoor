// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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
import {Button, Card, Col, Descriptions, Input, Modal, Row, Select} from 'antd';
import {InfoCircleTwoTone} from "@ant-design/icons";
import * as PaymentBackend from "./backend/PaymentBackend";
import * as Setting from "./Setting";
import i18next from "i18next";

const { Option } = Select;

class PaymentEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      organizationName: props.organizationName !== undefined ? props.organizationName : props.match.params.organizationName,
      paymentName: props.match.params.paymentName,
      payment: null,
      isModalVisible: false,
      mode: props.location.mode !== undefined ? props.location.mode : "edit",
    };
  }

  UNSAFE_componentWillMount() {
    this.getPayment();
  }

  getPayment() {
    PaymentBackend.getPayment("admin", this.state.paymentName)
      .then((payment) => {
        this.setState({
          payment: payment,
        });
      });
  }

  parsePaymentField(key, value) {
    if ([""].includes(key)) {
      value = Setting.myParseInt(value);
    }
    return value;
  }

  updatePaymentField(key, value) {
    value = this.parsePaymentField(key, value);

    let payment = this.state.payment;
    payment[key] = value;
    this.setState({
      payment: payment,
    });
  }

  issueInvoice() {
    alert("111")
  }

  downloadInvoice() {
    Setting.openLinkSafe(this.state.payment.invoiceUrl);
  }

  renderModal() {
    const ths = this;
    const handleIssueInvoice = () => {
      ths.issueInvoice();
    };

    const handleCancel = () => {
      this.setState({
        isModalVisible: false,
      });
    };

    return (
      <Modal title={
        <div>
          <InfoCircleTwoTone twoToneColor="rgb(45,120,213)" />
          {" " + i18next.t("payment:Confirm your invoice information")}
        </div>
      }
             visible={this.state.isModalVisible}
             onOk={handleIssueInvoice}
             onCancel={handleCancel}
             okText={i18next.t("payment:Issue Invoice")}
             cancelText={i18next.t("general:Cancel")}>
        <p>
          {
            i18next.t("payment:Please carefully check your invoice information. Once the invoice is issued, it cannot be withdrawn or modified.")
          }
          <br/>
          <br/>
          <Descriptions size={"small"} bordered>
            <Descriptions.Item label={i18next.t("payment:Person name")} span={3}>{this.state.payment?.personName}</Descriptions.Item>
            <Descriptions.Item label={i18next.t("payment:Person ID card")} span={3}>{this.state.payment?.personIdCard}</Descriptions.Item>
            <Descriptions.Item label={i18next.t("payment:Person Email")} span={3}>{this.state.payment?.personEmail}</Descriptions.Item>
            <Descriptions.Item label={i18next.t("payment:Person phone")} span={3}>{this.state.payment?.personPhone}</Descriptions.Item>
            <Descriptions.Item label={i18next.t("payment:Invoice type")} span={3}>{this.state.payment?.invoiceType === "Individual" ? i18next.t("payment:Individual") : i18next.t("payment:Organization")}</Descriptions.Item>
            <Descriptions.Item label={i18next.t("payment:Invoice title")} span={3}>{this.state.payment?.invoiceTitle}</Descriptions.Item>
            <Descriptions.Item label={i18next.t("payment:Invoice tax ID")} span={3}>{this.state.payment?.invoiceTaxId}</Descriptions.Item>
            <Descriptions.Item label={i18next.t("payment:Invoice remark")} span={3}>{this.state.payment?.invoiceRemark}</Descriptions.Item>
          </Descriptions>
        </p>
      </Modal>
    )
  }

  renderPayment() {
    return (
      <Card size="small" title={
        <div>
          {this.state.mode === "add" ? i18next.t("payment:New Payment") : i18next.t("payment:Edit Payment")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button onClick={() => this.submitPaymentEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: '20px'}} type="primary" onClick={() => this.submitPaymentEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: '20px'}} onClick={() => this.deletePayment()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      } style={(Setting.isMobile())? {margin: '5px'}:{}} type="inner">
        <Row style={{marginTop: '10px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled={true} value={this.state.payment.organization} onChange={e => {
              // this.updatePaymentField('organization', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Name"), i18next.t("general:Name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled={true} value={this.state.payment.name} onChange={e => {
              // this.updatePaymentField('name', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Display name"), i18next.t("general:Display name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled={true} value={this.state.payment.displayName} onChange={e => {
              this.updatePaymentField('displayName', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Provider"), i18next.t("general:Provider - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled={true} value={this.state.payment.provider} onChange={e => {
              // this.updatePaymentField('provider', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("payment:Type"), i18next.t("payment:Type - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled={true} value={this.state.payment.type} onChange={e => {
              // this.updatePaymentField('type', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("payment:Product"), i18next.t("payment:Product - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled={true} value={this.state.payment.productName} onChange={e => {
              // this.updatePaymentField('productName', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("payment:Price"), i18next.t("payment:Price - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled={true} value={this.state.payment.price} onChange={e => {
              // this.updatePaymentField('amount', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("payment:Currency"), i18next.t("payment:Currency - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled={true} value={this.state.payment.currency} onChange={e => {
              // this.updatePaymentField('currency', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("payment:State"), i18next.t("payment:State - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled={true} value={this.state.payment.state} onChange={e => {
              // this.updatePaymentField('state', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("payment:Message"), i18next.t("payment:Message - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled={true} value={this.state.payment.message} onChange={e => {
              // this.updatePaymentField('message', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("payment:Person name"), i18next.t("payment:Person name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled={this.state.payment.invoiceUrl !== ""} value={this.state.payment.personName} onChange={e => {
              this.updatePaymentField('personName', e.target.value);
              if (this.state.payment.invoiceType === "Individual") {
                this.updatePaymentField('invoiceTitle', e.target.value);
                this.updatePaymentField('invoiceTaxId', "");
              }
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("payment:Person ID card"), i18next.t("payment:Person ID card - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled={this.state.payment.invoiceUrl !== ""} value={this.state.payment.personIdCard} onChange={e => {
              this.updatePaymentField('personIdCard', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("payment:Person Email"), i18next.t("payment:Person Email - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled={this.state.payment.invoiceUrl !== ""} value={this.state.payment.personEmail} onChange={e => {
              this.updatePaymentField('personEmail', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("payment:Person phone"), i18next.t("payment:Person phone - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled={this.state.payment.invoiceUrl !== ""} value={this.state.payment.personPhone} onChange={e => {
              this.updatePaymentField('personPhone', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("payment:Invoice type"), i18next.t("payment:Invoice type - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select disabled={this.state.payment.invoiceUrl !== ""} virtual={false} style={{width: '100%'}} value={this.state.payment.invoiceType} onChange={(value => {
              this.updatePaymentField('invoiceType', value);
              if (value === "Individual") {
                this.updatePaymentField('invoiceTitle', this.state.payment.personName);
                this.updatePaymentField('invoiceTaxId', "");
              }
            })}>
              {
                [
                  {id: 'Individual', name: i18next.t("payment:Individual")},
                  {id: 'Organization', name: i18next.t("payment:Organization")},
                ].map((item, index) => <Option key={index} value={item.id}>{item.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("payment:Invoice title"), i18next.t("payment:Invoice title - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled={this.state.payment.invoiceUrl !== "" || this.state.payment.invoiceType === "Individual"} value={this.state.payment.invoiceTitle} onChange={e => {
              this.updatePaymentField('invoiceTitle', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("payment:Invoice tax ID"), i18next.t("payment:Invoice tax ID - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled={this.state.payment.invoiceUrl !== "" || this.state.payment.invoiceType === "Individual"} value={this.state.payment.invoiceTaxId} onChange={e => {
              this.updatePaymentField('invoiceTaxId', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("payment:Invoice remark"), i18next.t("payment:Invoice remark - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled={this.state.payment.invoiceUrl !== ""} value={this.state.payment.invoiceRemark} onChange={e => {
              this.updatePaymentField('invoiceRemark', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("payment:Invoice URL"), i18next.t("payment:Invoice URL - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled={true} value={this.state.payment.invoiceUrl} onChange={e => {
              this.updatePaymentField('invoiceUrl', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("payment:Invoice actions"), i18next.t("payment:Invoice actions - Tooltip"))} :
          </Col>
          <Col span={22} >
            {
              this.state.payment.invoiceUrl === "" ? (
                <Button type={"primary"} onClick={() => {
                  const errorText = this.checkError();
                  if (errorText !== "") {
                    Setting.showMessage("error", errorText);
                    return;
                  }

                  this.setState({
                    isModalVisible: true,
                  });
                }}>{i18next.t("payment:Issue Invoice")}</Button>
              ) : (
                <Button type={"primary"} onClick={() => this.downloadInvoice(false)}>{i18next.t("payment:Download Invoice")}</Button>
              )
            }
          </Col>
        </Row>
      </Card>
    )
  }

  checkError() {
    if (!Setting.isValidPersonName(this.state.payment.personName)) {
      return i18next.t("signup:Please input your real name!");
    }

    if (!Setting.isValidIdCard(this.state.payment.personIdCard)) {
      return i18next.t("signup:Please input the correct ID card number!");
    }

    if (!Setting.isValidEmail(this.state.payment.personEmail)) {
      return i18next.t("signup:The input is not valid Email!");
    }

    if (!Setting.isValidPhone(this.state.payment.personPhone)) {
      return i18next.t("signup:The input is not valid Phone!");
    }

    if (!Setting.isValidPhone(this.state.payment.personPhone)) {
      return i18next.t("signup:The input is not valid Phone!");
    }

    if (this.state.payment.invoiceType === "Individual") {
      if (this.state.payment.invoiceTitle !== this.state.payment.personName) {
        return i18next.t("signup:The input is not invoice title!");
      }

      if (this.state.payment.invoiceTaxId !== "") {
        return i18next.t("signup:The input is not invoice Tax ID!");
      }
    } else {
      if (!Setting.isValidInvoiceTitle(this.state.payment.invoiceTitle)) {
        return i18next.t("signup:The input is not invoice title!");
      }

      if (!Setting.isValidTaxId(this.state.payment.invoiceTaxId)) {
        return i18next.t("signup:The input is not invoice Tax ID!");
      }
    }

    return "";
  }

  submitPaymentEdit(willExist) {
    const errorText = this.checkError();
    if (errorText !== "") {
      Setting.showMessage("error", errorText);
      return;
    }

    let payment = Setting.deepCopy(this.state.payment);
    PaymentBackend.updatePayment(this.state.payment.owner, this.state.paymentName, payment)
      .then((res) => {
        if (res.msg === "") {
          Setting.showMessage("success", `Successfully saved`);
          this.setState({
            paymentName: this.state.payment.name,
          });

          if (willExist) {
            this.props.history.push(`/payments`);
          } else {
            this.props.history.push(`/payments/${this.state.payment.name}`);
          }
        } else {
          Setting.showMessage("error", res.msg);
          this.updatePaymentField('name', this.state.paymentName);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `Failed to connect to server: ${error}`);
      });
  }

  deletePayment() {
    PaymentBackend.deletePayment(this.state.payment)
      .then(() => {
        this.props.history.push(`/payments`);
      })
      .catch(error => {
        Setting.showMessage("error", `Payment failed to delete: ${error}`);
      });
  }

  render() {
    return (
      <div>
        {
          this.state.payment !== null ? this.renderPayment() : null
        }
        {
          this.renderModal()
        }
        <div style={{marginTop: '20px', marginLeft: '40px'}}>
          <Button size="large" onClick={() => this.submitPaymentEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: '20px'}} type="primary" size="large" onClick={() => this.submitPaymentEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: '20px'}} size="large" onClick={() => this.deletePayment()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      </div>
    );
  }
}

export default PaymentEditPage;
