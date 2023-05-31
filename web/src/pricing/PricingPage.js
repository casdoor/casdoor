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

import React from "react";
import {Card, Col, Row} from "antd";
import * as PricingBackend from "../backend/PricingBackend";
import * as PlanBackend from "../backend/PlanBackend";
import CustomGithubCorner from "../common/CustomGithubCorner";
import * as Setting from "../Setting";
import SingleCard from "./SingleCard";
import i18next from "i18next";

class PricingPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      applications: null,
      pricingName: (props.pricingName ?? props.match?.params?.pricingName) ?? null,
      pricing: props.pricing,
      plans: null,
      loading: false,
    };
  }

  componentDidMount() {
    this.setState({
      applications: [],
    });

    if (this.state.pricing) {
      this.loadPlans();
    } else {
      this.loadPricing(this.state.pricingName);
    }
    this.setState({
      loading: true,
    });
  }

  componentDidUpdate() {
    if (this.state.pricing &&
      this.state.pricing.plans?.length !== this.state.plans?.length && !this.state.loading) {
      this.setState({loading: true});
      this.loadPlans();
    }
  }

  loadPlans() {
    const plans = this.state.pricing.plans.map((plan) =>
      PlanBackend.getPlanById(plan, true));

    Promise.all(plans)
      .then(results => {
        this.setState({
          plans: results,
          loading: false,
        });
      })
      .catch(error => {
        Setting.showMessage("error", `Failed to get plans: ${error}`);
      });
  }

  loadPricing(pricingName) {
    if (pricingName === undefined) {
      return;
    }

    PricingBackend.getPricing("built-in", pricingName)
      .then((result) => {
        this.setState({
          loading: false,
          pricing: result,
        });
        this.onUpdatePricing(result);
      });
  }

  onUpdatePricing(pricing) {
    this.props.onUpdatePricing(pricing);
  }

  renderCards() {

    const getUrlByPlan = (plan) => {
      const pricing = this.state.pricing;
      const signUpUrl = `/signup/${pricing.application}?plan=${plan}&pricing=${pricing.name}`;
      return `${window.location.origin}${signUpUrl}`;
    };

    if (Setting.isMobile()) {
      return (
        <Card style={{border: "none"}} bodyStyle={{padding: 0}}>
          {
            this.state.plans.map(item => {
              return (
                <SingleCard link={getUrlByPlan(item.name)} key={item.name} plan={item} isSingle={this.state.plans.length === 1} />
              );
            })
          }
        </Card>
      );
    } else {
      return (
        <div style={{marginRight: "15px", marginLeft: "15px"}}>
          <Row style={{justifyContent: "center"}} gutter={24}>
            {
              this.state.plans.map(item => {
                return (
                  <SingleCard style={{marginRight: "5px", marginLeft: "5px"}} link={getUrlByPlan(item.name)} key={item.name} plan={item} isSingle={this.state.plans.length === 1} />
                );
              })
            }
          </Row>
        </div>
      );
    }
  }

  render() {
    if (this.state.loading || this.state.plans === null || this.state.plans === undefined) {
      return null;
    }

    const pricing = this.state.pricing;

    return (
      <React.Fragment>
        <CustomGithubCorner />
        <div className="login-content">
          <div className="login-panel">
            <div className="login-form">
              <h1 style={{fontSize: "48px", marginTop: "0px", marginBottom: "15px"}}>{pricing.displayName}</h1>
              <span style={{fontSize: "20px"}}>{pricing.description}</span>
              <Row style={{width: "100%", marginTop: "40px"}}>
                <Col span={24} style={{display: "flex", justifyContent: "center"}} >
                  {
                    this.renderCards()
                  }
                </Col>
              </Row>
              <Row style={{justifyContent: "center"}}>
                {pricing && pricing.trialDuration > 0
                  ? <i>{i18next.t("pricing:Free")} {pricing.trialDuration}-{i18next.t("pricing:days trial available!")}</i>
                  : null}
              </Row>
            </div>
          </div>
        </div>
      </React.Fragment>
    );
  }
}

export default PricingPage;
