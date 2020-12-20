import React from "react";
import {Button, Card, Col, Input, Row} from 'antd';
import {LinkOutlined} from "@ant-design/icons";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as Setting from "./Setting";

class OrganizationEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      organizationName: props.match.params.organizationName,
      organization: null,
      tasks: [],
      resources: [],
    };
  }

  componentWillMount() {
    this.getOrganization();
  }

  getOrganization() {
    OrganizationBackend.getOrganization("admin", this.state.organizationName)
      .then((organization) => {
        this.setState({
          organization: organization,
        });
      });
  }

  parseOrganizationField(key, value) {
    // if ([].includes(key)) {
    //   value = Setting.myParseInt(value);
    // }
    return value;
  }

  updateOrganizationField(key, value) {
    value = this.parseOrganizationField(key, value);

    let organization = this.state.organization;
    organization[key] = value;
    this.setState({
      organization: organization,
    });
  }

  renderOrganization() {
    return (
      <Card size="small" title={
        <div>
          Edit Organization&nbsp;&nbsp;&nbsp;&nbsp;
          <Button type="primary" onClick={this.submitOrganizationEdit.bind(this)}>Save</Button>
        </div>
      } style={{marginLeft: '5px'}} type="inner">
        <Row style={{marginTop: '10px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            Name:
          </Col>
          <Col span={22} >
            <Input value={this.state.organization.name} onChange={e => {
              this.updateOrganizationField('name', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            Display Name:
          </Col>
          <Col span={22} >
            <Input value={this.state.organization.displayName} onChange={e => {
              this.updateOrganizationField('displayName', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            Website Url:
          </Col>
          <Col span={22} >
            <Input value={this.state.organization.websiteUrl} onChange={e => {
              this.updateOrganizationField('websiteUrl', e.target.value);
            }} />
          </Col>
        </Row>
      </Card>
    )
  }

  submitOrganizationEdit() {
    let organization = Setting.deepCopy(this.state.organization);
    OrganizationBackend.updateOrganization(this.state.organization.owner, this.state.organizationName, organization)
      .then((res) => {
        if (res) {
          Setting.showMessage("success", `Successfully saved`);
          this.setState({
            organizationName: this.state.organization.name,
          });
          this.props.history.push(`/organizations/${this.state.organization.name}`);
        } else {
          Setting.showMessage("error", `failed to save: server side failure`);
          this.updateOrganizationField('name', this.state.organizationName);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `failed to save: ${error}`);
      });
  }

  render() {
    return (
      <div>
        <Row style={{width: "100%"}}>
          <Col span={1}>
          </Col>
          <Col span={22}>
            {
              this.state.organization !== null ? this.renderOrganization() : null
            }
          </Col>
          <Col span={1}>
          </Col>
        </Row>
        <Row style={{margin: 10}}>
          <Col span={2}>
          </Col>
          <Col span={18}>
            <Button type="primary" size="large" onClick={this.submitOrganizationEdit.bind(this)}>Save</Button>
          </Col>
        </Row>
      </div>
    );
  }
}

export default OrganizationEditPage;
