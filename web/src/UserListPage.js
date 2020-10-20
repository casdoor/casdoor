import React from "react";
import {Button, Col, Popconfirm, Row, Table} from 'antd';
import moment from "moment";
import * as Setting from "./Setting";
import * as UserBackend from "./backend/UserBackend";

class UserListPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      users: null,
    };
  }

  componentWillMount() {
    this.getUsers();
  }

  getUsers() {
    UserBackend.getUsers("admin")
      .then((res) => {
        this.setState({
          users: res,
        });
      });
  }

  newUser() {
    return {
      owner: "admin", // this.props.account.username,
      name: `user_${this.state.users.length}`,
      title: `New User - ${this.state.users.length}`,
      createdTime: moment().format(),
      Url: "",
    }
  }

  addUser() {
    const newUser = this.newUser();
    UserBackend.addUser(newUser)
      .then((res) => {
          Setting.showMessage("success", `User added successfully`);
          this.setState({
            users: Setting.prependRow(this.state.users, newUser),
          });
        }
      )
      .catch(error => {
        Setting.showMessage("error", `User failed to add: ${error}`);
      });
  }

  deleteUser(i) {
    UserBackend.deleteUser(this.state.users[i])
      .then((res) => {
          Setting.showMessage("success", `User deleted successfully`);
          this.setState({
            users: Setting.deleteRow(this.state.users, i),
          });
        }
      )
      .catch(error => {
        Setting.showMessage("error", `User failed to delete: ${error}`);
      });
  }

  renderTable(users) {
    const columns = [
      {
        title: 'Name',
        dataIndex: 'name',
        key: 'name',
        width: '120px',
        sorter: (a, b) => a.name.localeCompare(b.name),
        render: (text, record, index) => {
          return (
            <a href={`/users/${text}`}>{text}</a>
          )
        }
      },
      {
        title: 'Title',
        dataIndex: 'title',
        key: 'title',
        // width: '80px',
        sorter: (a, b) => a.title.localeCompare(b.title),
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
        title: 'Url',
        dataIndex: 'url',
        key: 'url',
        width: '150px',
        sorter: (a, b) => a.url.localeCompare(b.url),
        render: (text, record, index) => {
          return (
            <a target="_blank" href={text}>
              {
                text
              }
            </a>
          )
        }
      },
      {
        title: 'Action',
        dataIndex: '',
        key: 'op',
        width: '220px',
        render: (text, record, index) => {
          return (
            <div>
              <Button style={{marginTop: '10px', marginBottom: '10px', marginRight: '10px'}} type="primary" onClick={() => Setting.goToLink(`/users/${record.name}`)}>Edit</Button>
              <Popconfirm
                title={`Sure to delete user: ${record.name} ?`}
                onConfirm={() => this.deleteUser(index)}
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
        <Table columns={columns} dataSource={users} rowKey="name" size="middle" bordered pagination={{pageSize: 100}}
               title={() => (
                 <div>
                   Users&nbsp;&nbsp;&nbsp;&nbsp;
                   <Button type="primary" size="small" onClick={this.addUser.bind(this)}>Add</Button>
                 </div>
               )}
               loading={users === null}
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
              this.renderTable(this.state.users)
            }
          </Col>
          <Col span={1}>
          </Col>
        </Row>
      </div>
    );
  }
}

export default UserListPage;
