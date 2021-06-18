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
import {Button, Col, Row} from "antd";
import * as ApplicationBackend from "../backend/ApplicationBackend";
import * as Setting from "../Setting";
import i18next from "i18next";
import AffiliationSelect from "../common/AffiliationSelect";
import * as UserBackend from "../backend/UserBackend";

class PromptPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      type: props.type,
      applicationName: props.applicationName !== undefined ? props.applicationName : (props.match === undefined ? null : props.match.params.applicationName),
      application: null,
      user: Setting.deepCopy(this.props.account),
    };
  }

  UNSAFE_componentWillMount() {
    this.getApplication();
  }

  getApplication() {
    if (this.state.applicationName === null) {
      return;
    }

    ApplicationBackend.getApplication("admin", this.state.applicationName)
      .then((application) => {
        this.setState({
          application: application,
        });
      });
  }

  getApplicationObj() {
    if (this.props.application !== undefined) {
      return this.props.application;
    } else {
      return this.state.application;
    }
  }

  parseUserField(key, value) {
    // if ([].includes(key)) {
    //   value = Setting.myParseInt(value);
    // }
    return value;
  }

  updateUserField(key, value) {
    value = this.parseUserField(key, value);

    let user = this.state.user;
    user[key] = value;
    this.setState({
      user: user,
    });
  }

  renderAffiliation(application) {
    const signupItems = application.signupItems.filter(signupItem => signupItem.name === "Affiliation");
    if (signupItems.length === 0) {
      return null;
    }

    if (!signupItems[0].prompted) {
      return null;
    }

    return (
      <AffiliationSelect labelSpan={6} application={application} user={this.state.user} onUpdateUserField={(key, value) => { return this.updateUserField(key, value)}} />
    )
  }

  renderContent(application) {
    return (
      <div style={{width: '400px'}}>
        {
          this.renderAffiliation(application)
        }
      </div>
    )
  }

  isAnswered() {
    if (this.state.user === null) {
      return false;
    }

    return this.state.user.affiliation !== "";
  }

  submitUserEdit() {
    let user = Setting.deepCopy(this.state.user);
    UserBackend.updateUser(this.state.user.owner, this.state.user.name, user)
      .then((res) => {
        if (res.msg === "") {
          Setting.showMessage("success", `Successfully saved`);

          Setting.goToLogin(this, this.getApplicationObj());
        } else {
          Setting.showMessage("error", res.msg);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `Failed to connect to server: ${error}`);
      });
  }

  render() {
    const application = this.getApplicationObj();
    if (application === null) {
      return null;
    }

    return (
      <Row>
        <Col span={24} style={{display: "flex", justifyContent: "center"}}>
          <div style={{marginTop: "80px", marginBottom: "50px", textAlign: "center"}}>
            {
              Setting.renderHelmet(application)
            }
            {
              Setting.renderLogo(application)
            }
            {
              this.renderContent(application)
            }
            <Row style={{margin: 10}}>
              <Col span={18}>
              </Col>
            </Row>
            <div style={{marginTop: "50px"}}>
              <Button disabled={!this.isAnswered()} type="primary" size="large" onClick={this.submitUserEdit.bind(this)}>{i18next.t("signup:Submit and complete")}</Button>
            </div>
          </div>
        </Col>
      </Row>
    )
  }
}

export default PromptPage;
