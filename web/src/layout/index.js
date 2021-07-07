
import React from 'react';
import { Component } from 'react';
import { Layout, Menu, Breadcrumb, Avatar } from 'antd';
import * as Setting from "../Setting.js";
import CustomGithubCorner from "../CustomGithubCorner";

const { Header, Footer, Sider, Content } = Layout;

// 引入子菜单组件
const SubMenu = Menu.SubMenu; 

export default class BasicLayout extends Component {
  state = {
    collapsed: false,
  };
  onCollapse = collapsed => {
    console.log(collapsed);
    this.setState({ collapsed });
  };

  render() {
    const { collapsed } = this.state;
    return (
      <Layout style={{ minHeight: '100vh' }}>
        <Sider  collapsible collapsed={collapsed} onCollapse={this.onCollapse} theme ='light'>
          <div className='logo' key="logo"></div>
          <Menu theme="light" mode="inline" defaultSelectedKeys={['1']}>
            <Menu.Item key="1">
              <span>Home</span>
            </Menu.Item>
            <SubMenu
              key="sub1"
              title={<span><span>Forms</span></span>}
            >
               <Menu.Item key="2">Organizations</Menu.Item>
               <Menu.Item key="3">Users</Menu.Item>
               <Menu.Item key="4">Providers</Menu.Item>
               <Menu.Item key="5">Applications</Menu.Item>
               <Menu.Item key="6">Tokens</Menu.Item>
            </SubMenu>
          </Menu>
        </Sider>
        <Layout>
          <Header style={{ background: '#fff', textAlign: 'center', padding: 0 }}>
          <CustomGithubCorner/>
          </Header>
          <Content style={{ margin: '24px 16px 0' }}>
            <Breadcrumb style={{ margin: '16px 0' }}>
            </Breadcrumb>
            <div style={{ padding: 24, background: '#fff', minHeight: 360 }}>
              {this.props.children}
            </div>
          </Content>
          <Footer style={{ textAlign: 'center' }}>
            Made with <span
             style={{color: 'rgb(255, 255, 255)'}}>❤️</span>
              by <a style={{fontWeight: "bold", color: "black"}} target="_blank" href="https://casbin.org" rel="noreferrer">
              Casbin</a>
          </Footer>
        </Layout>
      </Layout>
    )
  }
}