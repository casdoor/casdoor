import React from "react";
import {Button, Col, Popconfirm, Row, Table} from 'antd';
import moment from "moment";
import {Link} from 'react-router-dom'
import * as Setting from "./Setting";
import * as ProviderBackend from "./backend/ProviderBackend";

class ProviderListPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      providers: null,
    };
  }

  componentWillMount() {
    this.getProviders();
  }

  getProviders() {
    ProviderBackend.getProviders("admin")
      .then((res) => {
        this.setState({
          providers: res,
        });
      });
  }

  newProvider() {
    return {
      owner: "admin", // this.props.account.providername,
      name: `provider_${this.state.providers.length}`,
      createdTime: moment().format(),
      displayName: `New Provider - ${this.state.providers.length}`,
      type: "github",
      clientId: "",
      clientSecret: "",
      providerUrl: "https://github.com/organizations/xxx/settings/applications/1234567",
    }
  }

  addProvider() {
    const newProvider = this.newProvider();
    ProviderBackend.addProvider(newProvider)
      .then((res) => {
          Setting.showMessage("success", `Provider added successfully`);
          this.setState({
            providers: Setting.prependRow(this.state.providers, newProvider),
          });
        }
      )
      .catch(error => {
        Setting.showMessage("error", `Provider failed to add: ${error}`);
      });
  }

  deleteProvider(i) {
    ProviderBackend.deleteProvider(this.state.providers[i])
      .then((res) => {
          Setting.showMessage("success", `Provider deleted successfully`);
          this.setState({
            providers: Setting.deleteRow(this.state.providers, i),
          });
        }
      )
      .catch(error => {
        Setting.showMessage("error", `Provider failed to delete: ${error}`);
      });
  }

  renderTable(providers) {
    const columns = [
      {
        title: 'Name',
        dataIndex: 'name',
        key: 'name',
        width: '120px',
        sorter: (a, b) => a.name.localeCompare(b.name),
        render: (text, record, index) => {
          return (
            <Link to={`/providers/${text}`}>{text}</Link>
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
        title: 'Type',
        dataIndex: 'type',
        key: 'type',
        width: '150px',
        sorter: (a, b) => a.type.localeCompare(b.type),
      },
      {
        title: 'Client Id',
        dataIndex: 'clientId',
        key: 'clientId',
        width: '150px',
        sorter: (a, b) => a.clientId.localeCompare(b.clientId),
      },
      {
        title: 'Client Secret',
        dataIndex: 'clientSecret',
        key: 'clientSecret',
        width: '150px',
        sorter: (a, b) => a.clientSecret.localeCompare(b.clientSecret),
      },
      {
        title: 'Provider Url',
        dataIndex: 'providerUrl',
        key: 'providerUrl',
        width: '150px',
        sorter: (a, b) => a.providerUrl.localeCompare(b.providerUrl),
      },
      {
        title: 'Action',
        dataIndex: '',
        key: 'op',
        width: '170px',
        render: (text, record, index) => {
          return (
            <div>
              <Button style={{marginTop: '10px', marginBottom: '10px', marginRight: '10px'}} type="primary" onClick={() => Setting.goToLink(`/providers/${record.name}`)}>Edit</Button>
              <Popconfirm
                title={`Sure to delete provider: ${record.name} ?`}
                onConfirm={() => this.deleteProvider(index)}
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
        <Table columns={columns} dataSource={providers} rowKey="name" size="middle" bordered pagination={{pageSize: 100}}
               title={() => (
                 <div>
                   Providers&nbsp;&nbsp;&nbsp;&nbsp;
                   <Button type="primary" size="small" onClick={this.addProvider.bind(this)}>Add</Button>
                 </div>
               )}
               loading={providers === null}
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
              this.renderTable(this.state.providers)
            }
          </Col>
          <Col span={1}>
          </Col>
        </Row>
      </div>
    );
  }
}

export default ProviderListPage;
