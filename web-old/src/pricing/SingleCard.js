// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

import i18next from "i18next";
import React from "react";
import {Button, Card, Col, Row} from "antd";
import * as Setting from "../Setting";
import {withRouter} from "react-router-dom";

const {Meta} = Card;

class SingleCard extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
    };
  }

  renderCard(plan, isSingle, link) {
    return (
      <Col style={{minWidth: "320px", paddingLeft: "20px", paddingRight: "20px", paddingBottom: "20px", marginBottom: "20px", paddingTop: "0px"}} span={6}>
        <Card
          hoverable
          onClick={() => Setting.isMobile() ? window.location.href = link : null}
          style={isSingle ? {width: "320px", height: "100%"} : {width: "100%", height: "100%", paddingTop: "0px"}}
          title={<h2>{plan.displayName}</h2>}
        >
          <Col>
            <Row>
              <div style={{textAlign: "left"}} className="px-10 mt-5">
                <span style={{fontSize: "40px", fontWeight: 700}}>{Setting.getCurrencySymbol(plan.currency)} {plan.price}</span>
                <span style={{fontSize: "18px", fontWeight: 600, color: "gray"}}> {plan.period === "Yearly" ? i18next.t("plan:per year") : i18next.t("plan:per month")}</span>
              </div>
            </Row>

            <Row style={{height: "90px", paddingTop: "15px"}}>
              <div style={{textAlign: "left", fontSize: "18px"}}>
                <Meta description={plan.description} />
              </div>
            </Row>

            {/* <ul style={{listStyleType: "none", paddingLeft: "0px", textAlign: "left"}}>
              {(plan.options ?? []).map((option) => {
              // eslint-disable-next-line react/jsx-key
                return <li>
                  <svg style={{height: "1rem", width: "1rem", fill: "green", marginRight: "10px"}} xmlns="http://www.w3.org/2000/svg"
                    viewBox="0 0 20 20">
                    <path d="M0 11l2-2 5 5L18 3l2 2L7 18z"></path>
                  </svg>
                  <span style={{fontSize: "16px"}}>{option}</span>
                </li>;
              })}
            </ul> */}

            <Row style={{paddingTop: "15px"}}>
              <Button style={{width: "100%", height: "50px", borderRadius: "0px", bottom: "0", left: "0"}} type="primary" key="subscribe" onClick={() => window.location.href = link}>
                {
                  i18next.t("pricing:Getting started")
                }
              </Button>
            </Row>
          </Col>
        </Card>
      </Col>
    );
  }

  render() {
    return this.renderCard(this.props.plan, this.props.isSingle, this.props.link);
  }
}

export default withRouter(SingleCard);
