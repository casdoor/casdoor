import React from "react";
import {Button, Col, Popconfirm, Row, Table} from 'antd';
import moment from "moment";
import * as Setting from "./Setting";
import * as OrganizationBackend from "./backend/OrganizationBackend";

class OrganizationListPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      organizations: null,
    };
  }

  componentWillMount() {
    this.getOrganizations();
  }

  getOrganizations() {
    OrganizationBackend.getOrganizations("admin")
      .then((res) => {
        this.setState({
          organizations: res,
        });
      });
  }

  newOrganization() {
    return {
      owner: "admin", // this.props.account.organizationname,
      name: `organization_${this.state.organizations.length}`,
      createdTime: moment().format(),
      displayName: `New Organization - ${this.state.organizations.length}`,
      websiteUrl: "https://example.com",
    }
  }

  addOrganization() {
    const newOrganization = this.newOrganization();
    OrganizationBackend.addOrganization(newOrganization)
      .then((res) => {
          Setting.showMessage("success", `Organization added successfully`);
          this.setState({
            organizations: Setting.prependRow(this.state.organizations, newOrganization),
          });
        }
      )
      .catch(error => {
        Setting.showMessage("error", `Organization failed to add: ${error}`);
      });
  }

  deleteOrganization(i) {
    OrganizationBackend.deleteOrganization(this.state.organizations[i])
      .then((res) => {
          Setting.showMessage("success", `Organization deleted successfully`);
          this.setState({
            organizations: Setting.deleteRow(this.state.organizations, i),
          });
        }
      )
      .catch(error => {
        Setting.showMessage("error", `Organization failed to delete: ${error}`);
      });
  }

  renderTable(organizations) {
    const columns = [
      {
        title: 'Name',
        dataIndex: 'name',
        key: 'name',
        width: '120px',
        sorter: (a, b) => a.name.localeCompare(b.name),
        render: (text, record, index) => {
          return (
            <a href={`/organizations/${text}`}>{text}</a>
          )
        }
      },
      {
        title: 'Created Time',
        dataIndex: 'createdTime',
        key: 'createdTime',
        width: '160px',
        sorter: (a, b) => a.createdTime.localeCompare(b.createdTime),
        render: (text, record, index) => {
          return Setting.getFormattedDate(text);
        }
      },
      {
        title: 'Display Name',
        dataIndex: 'displayName',
        key: 'displayName',
        // width: '100px',
        sorter: (a, b) => a.displayName.localeCompare(b.displayName),
      },
      {
        title: 'Website Url',
        dataIndex: 'websiteUrl',
        key: 'websiteUrl',
        width: '300px',
        sorter: (a, b) => a.websiteUrl.localeCompare(b.websiteUrl),
      },
      {
        title: 'Action',
        dataIndex: '',
        key: 'op',
        width: '170px',
        render: (text, record, index) => {
          return (
            <div>
              <Button style={{marginTop: '10px', marginBottom: '10px', marginRight: '10px'}} type="primary" onClick={() => Setting.goToLink(`/organizations/${record.name}`)}>Edit</Button>
              <Popconfirm
                title={`Sure to delete organization: ${record.name} ?`}
                onConfirm={() => this.deleteOrganization(index)}
              >
                <Button style={{marginBottom: '10px'}} type="danger">Delete</Button>
              </Popconfirm>
            </div>
          )
        }
      },
    ];

    return (
      <div>
        <Table columns={columns} dataSource={organizations} rowKey="name" size="middle" bordered pagination={{pageSize: 100}}
               title={() => (
                 <div>
                   Organizations&nbsp;&nbsp;&nbsp;&nbsp;
                   <Button type="primary" size="small" onClick={this.addOrganization.bind(this)}>Add</Button>
                 </div>
               )}
               loading={organizations === null}
        />
      </div>
    );
  }

  render() {
    return (
      <div>
        <Row style={{width: "100%"}}>
          <Col span={1}>
          </Col>
          <Col span={22}>
            {
              this.renderTable(this.state.organizations)
            }
          </Col>
          <Col span={1}>
          </Col>
        </Row>
      </div>
    );
  }
}

export default OrganizationListPage;
