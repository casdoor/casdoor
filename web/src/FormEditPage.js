// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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
import {Button, Card, Col, Input, Row, Select} from "antd";
import * as FormBackend from "./backend/FormBackend";
import * as Setting from "./Setting";
import i18next from "i18next";
import FormItemTable from "./table/FormItemTable";
import UserListPage from "./UserListPage";
import ApplicationListPage from "./ApplicationListPage";
import ProviderListPage from "./ProviderListPage";
import OrganizationListPage from "./OrganizationListPage";

const {Option} = Select;

class FormEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      formName: props.match.params.formName,
      form: null,
      formItems: [],
    };
  }

  UNSAFE_componentWillMount() {
    this.getForm();
  }

  getForm() {
    FormBackend.getForm(this.props.account.owner, this.state.formName)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            form: res.data,
          });
        }
      });
  }

  updateFormField(key, value) {
    const form = this.state.form;
    form[key] = value;
    this.setState({
      form: form,
    });
  }

  renderForm() {
    return (
      <Card size="small" title={
        <div>
          {i18next.t("form:Edit Form")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button onClick={() => this.submitFormEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary"
            onClick={() => this.submitFormEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
        </div>
      } style={{marginLeft: "5px"}} type="inner">
        <Row style={{marginTop: "10px"}}>
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Name"), i18next.t("general:Name - Tooltip"))} :
          </Col>
          <Col span={22}>
            <Input
              value={this.state.form.name}
              disabled={true}
              onChange={e => {this.updateFormField("name", e.target.value);}}
            />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}}>
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Display name"), i18next.t("general:Display name - Tooltip"))} :
          </Col>
          <Col span={22}>
            <Input value={this.state.form.displayName} onChange={e => {
              this.updateFormField("displayName", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}}>
          <Col style={{marginTop: "5px"}} span={Setting.isMobile() ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Type"), i18next.t("cert:Type - Tooltip"))} :
          </Col>
          <Col span={22}>
            <Select
              style={{width: "100%"}}
              value={this.state.form.type}
              onChange={value => {
                this.updateFormField("type", value);
                this.updateFormField("name", value);
                this.updateFormField("displayName", value);
                const defaultItems = new FormItemTable({formType: value}).getItems();
                this.updateFormField("formItems", defaultItems);
              }}
            >
              {Setting.getFormTypeOptions().map(option => (
                <Option key={option.id} value={option.id}>{i18next.t(option.name)}</Option>
              ))}
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}}>
          <Col style={{marginTop: "5px"}} span={Setting.isMobile() ? 22 : 2}>
            {Setting.getLabel(i18next.t("user:Tag"), i18next.t("product:Tag - Tooltip"))} :
          </Col>
          <Col span={22}>
            <Input value={this.state.form.tag} onChange={e => {
              this.updateFormField("tag", e.target.value);
              this.updateFormField("name", e.target.value ? `${this.state.form.type}-tag-${e.target.value}` : this.state.form.type);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}}>
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("form:Form items"), i18next.t("form:Form items - Tooltip"))} :
          </Col>
          <Col span={22}>
            <FormItemTable
              title={i18next.t("form:Form items")}
              table={this.state.form.formItems}
              onUpdateTable={(value) => {
                this.updateFormField("formItems", value);
              }}
              formType={this.state.form.type}
            />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}}>
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Preview"), i18next.t("general:Preview - Tooltip"))} :
          </Col>
          <Col span={22}>
            {
              this.renderListPreview()
            }
          </Col>
        </Row>
      </Card>
    );
  }

  renderListPreview() {
    let listPageComponent = null;

    if (this.state.form.type === "users") {
      listPageComponent = (<UserListPage {...this.props} formItems={this.state.form.formItems} />);
    } else if (this.state.form.type === "applications") {
      listPageComponent = (<ApplicationListPage {...this.props} formItems={this.state.form.formItems} />);
    } else if (this.state.form.type === "providers") {
      listPageComponent = (<ProviderListPage {...this.props} formItems={this.state.form.formItems} />);
    } else if (this.state.form.type === "organizations") {
      listPageComponent = (<OrganizationListPage {...this.props} formItems={this.state.form.formItems} />);
    }

    return (
      <div style={{position: "relative", border: "1px solid rgb(217,217,217)", height: "600px", cursor: "pointer"}} onClick={(e) => {Setting.openLink(`/${this.state.form.type}`);}}>
        <div style={{position: "relative", height: "100%", overflow: "auto"}}>
          <div style={{display: "inline-block", position: "relative", zIndex: 1, pointerEvents: "none"}}>
            {listPageComponent}
          </div>
        </div>
        <div style={{position: "absolute", top: 0, left: 0, right: 0, bottom: 0, zIndex: 10, background: "rgba(0,0,0,0.4)", pointerEvents: "none"}} />
      </div>
    );
  }

  submitFormEdit(exitAfterSave) {
    const form = Setting.deepCopy(this.state.form);
    FormBackend.updateForm(this.state.form.owner, this.state.formName, form)
      .then((res) => {
        if (res.status === "ok") {
          if (res.data) {
            Setting.showMessage("success", i18next.t("general:Successfully saved"));
            this.setState({
              formName: this.state.form.name,
            });
            if (exitAfterSave) {
              this.props.history.push("/forms");
            } else {
              this.props.history.push(`/forms/${this.state.form.name}`);
            }
          } else {
            Setting.showMessage("error", i18next.t("general:Failed to save"));
            this.updateFormField("name", this.state.formName);
          }
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${error}`);
      });
  }

  render() {
    return (
      <div>
        {
          this.state.form !== null ? this.renderForm() : null
        }
        <div style={{marginTop: "20px", marginLeft: "40px"}}>
          <Button size="large" onClick={() => this.submitFormEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" size="large"
            onClick={() => this.submitFormEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
        </div>
      </div>
    );
  }
}

export default FormEditPage;
