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
import {Card, Col, Radio, Row} from "antd";
import * as PricingBackend from "../backend/PricingBackend";
import * as PlanBackend from "../backend/PlanBackend";
import CustomGithubCorner from "../common/CustomGithubCorner";
import * as Setting from "../Setting";
import SingleCard from "./SingleCard";
import i18next from "i18next";

class PricingPage extends React.Component {
  constructor(props) {
    super(props);
    const params = new URLSearchParams(window.location.search);
    this.state = {
      classes: props,
      applications: null,
      owner: props.owner ?? (props.match?.params?.owner ?? null),
      pricingName: (props.pricingName ?? props.match?.params?.pricingName) ?? null,
      userName: params.get("user"),
      pricing: props.pricing,
      plans: null,
      periods: null,
      selectedPeriod: null,
      loading: false,
    };
  }

  componentDidMount() {
    this.setState({
      applications: [],
    });
    if (this.state.userName) {
      Setting.showMessage("info", `${i18next.t("pricing:paid-user do not have active subscription or pending subscription, please select a plan to buy")}`);
    }
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
      PlanBackend.getPlan(this.state.owner, plan, true));

    Promise.all(plans)
      .then(results => {
        const hasError = results.some(result => result.status === "error");
        if (hasError) {
          Setting.showMessage("error", i18next.t("pricing:Failed to get plans"));
          return;
        }
        const plans = results.map(result => result.data);
        const periods = [... new Set(plans.map(plan => plan.period).filter(period => period !== ""))];
        this.setState({
          plans: plans,
          periods: periods,
          selectedPeriod: periods?.[0],
          loading: false,
        });
      })
      .catch(error => {
        Setting.showMessage("error", i18next.t("pricing:Failed to get plans") + `: ${error}`);
      });
  }

  loadPricing(pricingName) {
    if (!pricingName) {
      return;
    }
    PricingBackend.getPricing(this.state.owner, pricingName)
      .then((res) => {
        if (res.status === "error") {
          Setting.showMessage("error", res.msg);
          return;
        }
        this.setState({
          loading: false,
          pricing: res.data,
        });
        this.onUpdatePricing(res.data);
      });
  }

  onUpdatePricing(pricing) {
    this.props.onUpdatePricing(pricing);
  }

  renderSelectPeriod() {
    if (!this.state.periods || this.state.periods.length <= 1) {
      return null;
    }
    return (
      <Radio.Group
        value={this.state.selectedPeriod}
        size="large"
        buttonStyle="solid"
        onChange={e => {
          this.setState({selectedPeriod: e.target.value});
        }}
      >
        {
          this.state.periods.map(period => {
            return (
              <Radio.Button key={period} value={period}>{period}</Radio.Button>
            );
          })
        }
      </Radio.Group>
    );
  }

  renderCards() {
    const getUrlByPlan = (planName) => {
      const pricing = this.state.pricing;
      let signUpUrl = `/signup/${pricing.application}?plan=${planName}&pricing=${pricing.name}`;
      if (this.state.userName) {
        signUpUrl = `/buy-plan/${pricing.owner}/${pricing.name}?plan=${planName}&user=${this.state.userName}`;
      }
      return `${window.location.origin}${signUpUrl}`;
    };

    if (Setting.isMobile()) {
      return (
        <Card style={{border: "none"}} styles={{body: {padding: 0}}}>
          {
            this.state.plans.map(item => {
              return item.period === this.state.selectedPeriod ? (
                <SingleCard link={getUrlByPlan(item.name)} key={item.name} plan={item} isSingle={this.state.plans.length === 1} />
              ) : null;
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
                return item.period === this.state.selectedPeriod ? (
                  <SingleCard style={{marginRight: "5px", marginLeft: "5px"}} link={getUrlByPlan(item.name)} key={item.name} plan={item} isSingle={this.state.plans.length === 1} />
                ) : null;
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
                    this.renderSelectPeriod()
                  }
                </Col>
              </Row>
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
