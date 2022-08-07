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
import {Button, Card, Col, Input, Row, Select, Switch} from "antd";
import * as ModelBackend from "./backend/ModelBackend";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as Setting from "./Setting";
import i18next from "i18next";
import TextArea from "antd/es/input/TextArea";

const {Option} = Select;

class ModelEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      organizationName: props.organizationName !== undefined ? props.organizationName : props.match.params.organizationName,
      modelName: props.match.params.modelName,
      model: null,
      organizations: [],
      users: [],
      models: [],
      mode: props.location.mode !== undefined ? props.location.mode : "edit",
    };
  }

  UNSAFE_componentWillMount() {
    this.getModel();
    this.getOrganizations();
  }

  getModel() {
    ModelBackend.getModel(this.state.organizationName, this.state.modelName)
      .then((model) => {
        this.setState({
          model: model,
        });

        this.getModels(model.owner);
      });
  }

  getOrganizations() {
    OrganizationBackend.getOrganizations("admin")
      .then((res) => {
        this.setState({
          organizations: (res.msg === undefined) ? res : [],
        });
      });
  }

  getModels(organizationName) {
    ModelBackend.getModels(organizationName)
      .then((res) => {
        this.setState({
          models: res,
        });
      });
  }

  parseModelField(key, value) {
    if ([""].includes(key)) {
      value = Setting.myParseInt(value);
    }
    return value;
  }

  updateModelField(key, value) {
    value = this.parseModelField(key, value);

    let model = this.state.model;
    model[key] = value;
    this.setState({
      model: model,
    });
  }

  renderModel() {
    return (
      <Card size="small" title={
        <div>
          {this.state.mode === "add" ? i18next.t("model:New Model") : i18next.t("model:Edit Model")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button onClick={() => this.submitModelEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" onClick={() => this.submitModelEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} onClick={() => this.deleteModel()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      } style={(Setting.isMobile()) ? {margin: "5px"} : {}} type="inner">
        <Row style={{marginTop: "10px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.model.owner} onChange={(value => {this.updateModelField("owner", value);})}>
              {
                this.state.organizations.map((organization, index) => <Option key={index} value={organization.name}>{organization.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Name"), i18next.t("general:Name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.model.name} onChange={e => {
              this.updateModelField("name", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Display name"), i18next.t("general:Display name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.model.displayName} onChange={e => {
              this.updateModelField("displayName", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("model:Model text"), i18next.t("model:Model text - Tooltip"))} :
          </Col>
          <Col span={22}>
            <TextArea rows={10} value={this.state.model.modelText} onChange={e => {
              this.updateModelField("modelText", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 19 : 2}>
            {Setting.getLabel(i18next.t("general:Is enabled"), i18next.t("general:Is enabled - Tooltip"))} :
          </Col>
          <Col span={1} >
            <Switch checked={this.state.model.isEnabled} onChange={checked => {
              this.updateModelField("isEnabled", checked);
            }} />
          </Col>
        </Row>
      </Card>
    );
  }

  submitModelEdit(willExist) {
    let model = Setting.deepCopy(this.state.model);
    ModelBackend.updateModel(this.state.organizationName, this.state.modelName, model)
      .then((res) => {
        if (res.msg === "") {
          Setting.showMessage("success", "Successfully saved");
          this.setState({
            modelName: this.state.model.name,
          });

          if (willExist) {
            this.props.history.push("/models");
          } else {
            this.props.history.push(`/models/${this.state.model.owner}/${this.state.model.name}`);
          }
        } else {
          Setting.showMessage("error", res.msg);
          this.updateModelField("name", this.state.modelName);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `Failed to connect to server: ${error}`);
      });
  }

  deleteModel() {
    ModelBackend.deleteModel(this.state.model)
      .then(() => {
        this.props.history.push("/models");
      })
      .catch(error => {
        Setting.showMessage("error", `Model failed to delete: ${error}`);
      });
  }

  render() {
    return (
      <div>
        {
          this.state.model !== null ? this.renderModel() : null
        }
        <div style={{marginTop: "20px", marginLeft: "40px"}}>
          <Button size="large" onClick={() => this.submitModelEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" size="large" onClick={() => this.submitModelEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} size="large" onClick={() => this.deleteModel()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      </div>
    );
  }
}

export default ModelEditPage;
