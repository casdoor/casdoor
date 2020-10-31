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

import React, { useEffect, useState } from 'react';
import * as Setting from './setting';
import { DownOutlined, LogoutOutlined, SettingOutlined } from '@ant-design/icons';
import { Avatar, BackTop, Dropdown, Layout, Menu } from 'antd';
import { Route, Routes, useNavigate } from 'react-router-dom';
import * as AccountBackend from './backend/account-backend';
import './app.css';
import { UserRoutes } from './user/user-routes';

const { Header, Footer } = Layout;

interface Account {
  username: string;
  name: string;
}

function isStartPages() {
  return (
    window.location.pathname.startsWith('/login') || window.location.pathname.startsWith('/register') || window.location.pathname === '/'
  );
}

function AppFooter() {
  return (
    <Footer
      id="footer"
      style={{
        borderTop: '1px solid #e8e8e8',
        backgroundColor: 'white',
        textAlign: 'center',
      }}
    >
      Made with <span style={{ color: 'rgb(255, 255, 255)' }}>❤️</span> by{' '}
      <a style={{ fontWeight: 'bold', color: 'black' }} rel="noreferrer" target="_blank" href="https://casbin.org">
        Casbin
      </a>
    </Footer>
  );
}

function AppMenu() {
  const [selectedMenuKey, setSelectedMenuKey] = useState(0);
  const [account, setAccount] = useState<Account | undefined>(undefined);
  const navigate = useNavigate();

  useEffect(() => {
    // TODO: Waiting for consolidation backend
    // if (window.location.pathname !== '/' && window.location.pathname !== '/login' && window.location.pathname !== '/register' && !account) {
    //   history.replace('/login');
    //   return;
    // }
    updateMenu();
    // getAccount();
  }, []);

  function handleRightDropdownClick(e: any) {
    if (e.key === 'account') {
      navigate(`/account`);
    } else if (e.key === 'logout') {
      logout();
    }
  }

  function updateMenu() {
    const uri = window.location.pathname;
    if (uri === '/') {
      setSelectedMenuKey(0);
    } else if (uri.includes('users')) {
      setSelectedMenuKey(1);
    } else {
      setSelectedMenuKey(-1);
    }
  }

  function getAccount() {
    AccountBackend.getAccount().then((res) => {
      const account = Setting.parseJson(res.data);
      setAccount(account);
      if (account) {
        // @ts-ignore Mouselog plugins
        window.mouselogUserId = account.username;
      }
    });
  }

  function logout() {
    AccountBackend.logout().then((res) => {
      // if (res.statusText === 'ok') {
      //   setAccount(undefined);
      //   Setting.showMessage('success', `Successfully logged out, redirected to homepage`);
      //   history.replace('/');
      // } else {
      //   Setting.showMessage('error', `Logout failed: ${res.msg}`);
      // }
    });
  }

  return (
    <Menu
      // theme="dark"
      mode={Setting.isMobile() && isStartPages() ? 'inline' : 'horizontal'}
      defaultSelectedKeys={[`${selectedMenuKey}`]}
      style={{ lineHeight: '64px' }}
    >
      <Menu.Item key="home">
        <a href="/">Home</a>
      </Menu.Item>
      <Menu.Item key="user">
        <a href="/users">Users</a>
      </Menu.Item>
      {account ? (
        <Dropdown
          key="4"
          overlay={
            <Menu onClick={handleRightDropdownClick}>
              <Menu.Item key="account">
                <SettingOutlined />
                My Account
              </Menu.Item>
              <Menu.Item key="logout">
                <LogoutOutlined />
                Logout
              </Menu.Item>
            </Menu>
          }
        >
          {/*eslint-disable-next-line*/}
          <a className="ant-dropdown-link" href="#" style={{ float: 'right' }}>
            <Avatar
              style={{
                backgroundColor: Setting.getAvatarColor(account.name),
                verticalAlign: 'middle',
              }}
              size="large"
            >
              {Setting.getShortName(account.name)}
            </Avatar>
            &nbsp; &nbsp;
            {Setting.isMobile() ? null : Setting.getShortName(account.name)} &nbsp;
            <DownOutlined />
            &nbsp; &nbsp; &nbsp;
          </a>
        </Dropdown>
      ) : (
        <>
          <Menu.Item key="register" style={{ float: 'right', marginRight: '20px' }}>
            <a href="/register">Register</a>
          </Menu.Item>
          <Menu.Item key="login" style={{ float: 'right' }}>
            <a href="/login">Login</a>
          </Menu.Item>
        </>
      )}
    </Menu>
  );
}

function AppHeader() {
  return (
    <Header style={{ padding: '0', marginBottom: '3px' }}>
      {/*eslint-disable-next-line*/}
      {Setting.isMobile() ? null : <a href="/" className="logo" />}
      <AppMenu />
    </Header>
  );
}

function App() {
  return (
    <div id="parent-area">
      <AppHeader />
      <BackTop />
      <div id="content-wrap">
        <Routes>
          <Route path="users/*" element={<UserRoutes />} />
        </Routes>
      </div>
      <AppFooter />
    </div>
  );
}

export default App;
