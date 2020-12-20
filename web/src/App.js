// Copyright 2020 The casbin Authors. All Rights Reserved.
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

import React, {Component} from 'react';
import './App.css';
import * as Setting from "./Setting";
import {DownOutlined, LogoutOutlined, SettingOutlined} from '@ant-design/icons';
import {Avatar, BackTop, Dropdown, Layout, Menu} from 'antd';
import {Switch, Route, withRouter, Redirect} from 'react-router-dom'
import * as AccountBackend from "./backend/AccountBackend";
import OrganizationListPage from "./OrganizationListPage";
import OrganizationEditPage from "./OrganizationEditPage";
import UserListPage from "./UserListPage";
import UserEditPage from "./UserEditPage";

const { Header, Footer } = Layout;

class App extends Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      selectedMenuKey: 0,
      account: undefined,
    };

    Setting.initServerUrl();
  }

  componentWillMount() {
    this.updateMenuKey();
    this.getAccount();
  }

  updateMenuKey() {
    // eslint-disable-next-line no-restricted-globals
    const uri = location.pathname;
    if (uri === '/') {
      this.setState({ selectedMenuKey: 0 });
    } else if (uri.includes('organizations')) {
      this.setState({ selectedMenuKey: 1 });
    } else if (uri.includes('users')) {
      this.setState({ selectedMenuKey: 2 });
    } else {
      this.setState({ selectedMenuKey: -1 });
    }
  }

  onLogined() {
    this.getAccount();
  }

  onUpdateAccount(account) {
    this.setState({
      account: account
    });
  }

  getAccount() {
    AccountBackend.getAccount()
      .then((res) => {
        const account = Setting.parseJson(res.data);
        if (window.location.pathname === '/' && account === null) {
          Setting.goToLink("/");
        }
        this.setState({
          account: account,
        });

        if (account !== undefined && account !== null) {
          window.mouselogUserId = account.username;
        }
      });
  }

  logout() {
    this.setState({
      expired: false,
      submitted: false,
    });

    AccountBackend.logout()
      .then((res) => {
        if (res.status === 'ok') {
          this.setState({
            account: null
          });

          Setting.showMessage("success", `Successfully logged out, redirected to homepage`);

          Setting.goToLink("/");
        } else {
          Setting.showMessage("error", `Logout failed: ${res.msg}`);
        }
      });
  }

  handleRightDropdownClick(e) {
    if (e.key === '0') {
      this.props.history.push(`/account`);
    } else if (e.key === '1') {
      this.logout();
    }
  }

  renderRightDropdown() {
    const menu = (
      <Menu onClick={this.handleRightDropdownClick.bind(this)}>
        <Menu.Item key='0'>
          <SettingOutlined />
          My Account
        </Menu.Item>
        <Menu.Item key='1'>
          <LogoutOutlined />
          Logout
        </Menu.Item>
      </Menu>
    );

    return (
      <Dropdown key="4" overlay={menu} >
        <a className="ant-dropdown-link" href="#" style={{float: 'right'}}>
          <Avatar style={{ backgroundColor: Setting.getAvatarColor(this.state.account.name), verticalAlign: 'middle' }} size="large">
            {Setting.getShortName(this.state.account.name)}
          </Avatar>
          &nbsp;
          &nbsp;
          {Setting.isMobile() ? null : Setting.getShortName(this.state.account.name)} &nbsp; <DownOutlined />
          &nbsp;
          &nbsp;
          &nbsp;
        </a>
      </Dropdown>
    )
  }

  renderAccount() {
    let res = [];

    if (this.state.account !== null && this.state.account !== undefined) {
      res.push(this.renderRightDropdown());
    } else {
      res.push(
        <Menu.Item key="1" style={{float: 'right', marginRight: '20px'}}>
          <a href="/register">
            Register
          </a>
        </Menu.Item>
      );
      res.push(
        <Menu.Item key="2" style={{float: 'right'}}>
          <a href="/login">
            Login
          </a>
        </Menu.Item>
      );
      res.push(
        <Menu.Item key="4" style={{float: 'right'}}>
          <a href="/">
            Home
          </a>
        </Menu.Item>
      );
    }

    return res;
  }

  renderMenu() {
    let res = [];

    // if (this.state.account === null || this.state.account === undefined) {
    //   return [];
    // }

    res.push(
      <Menu.Item key="0">
        <a href="/">
          Home
        </a>
      </Menu.Item>
    );
    res.push(
      <Menu.Item key="1">
        <a href="/organizations">
          Organizations
        </a>
      </Menu.Item>
    );
    res.push(
      <Menu.Item key="2">
        <a href="/users">
          Users
        </a>
      </Menu.Item>
    );

    return res;
  }

  renderHomeIfLogined(component) {
    if (this.state.account !== null && this.state.account !== undefined) {
      return <Redirect to='/' />
    } else {
      return component;
    }
  }

  renderLoginIfNotLogined(component) {
    if (this.state.account === null) {
      return <Redirect to='/login' />
    } else if (this.state.account === undefined) {
      return null;
    }
    else {
      return component;
    }
  }

  isStartPages() {
    return window.location.pathname.startsWith('/login') ||
      window.location.pathname.startsWith('/register') ||
      window.location.pathname === '/';
  }

  renderContent() {
    return (
      <div>
        <Header style={{ padding: '0', marginBottom: '3px'}}>
          {
            Setting.isMobile() ? null : <a className="logo" href={"/"} />
          }
          <Menu
            // theme="dark"
            mode={(Setting.isMobile() && this.isStartPages()) ? "inline" : "horizontal"}
            defaultSelectedKeys={[`${this.state.selectedMenuKey}`]}
            style={{ lineHeight: '64px' }}
          >
            {
              this.renderMenu()
            }
            {
              this.renderAccount()
            }
          </Menu>
        </Header>
        <Switch>
          <Route exact path="/organizations" component={OrganizationListPage}/>
          <Route exact path="/organizations/:organizationName" component={OrganizationEditPage}/>
          <Route exact path="/users" component={UserListPage}/>
          <Route exact path="/users/:userName" component={UserEditPage}/>
        </Switch>
      </div>
    )
  }

  renderFooter() {
    // How to keep your footer where it belongs ?
    // https://www.freecodecamp.org/neyarnws/how-to-keep-your-footer-where-it-belongs-59c6aa05c59c/

    return (
      <Footer id="footer" style={
        {
          borderTop: '1px solid #e8e8e8',
          backgroundColor: 'white',
          textAlign: 'center',
        }
      }>
        Made with <span style={{color: 'rgb(255, 255, 255)'}}>❤️</span> by <a style={{fontWeight: "bold", color: "black"}} target="_blank" href="https://casbin.org">Casbin</a>
      </Footer>
    )
  }

  render() {
    return (
      <div id="parent-area">
        <BackTop />
        <div id="content-wrap">
          {
            this.renderContent()
          }
        </div>
        {
          this.renderFooter()
        }
      </div>
    );
  }
}

export default withRouter(App);
