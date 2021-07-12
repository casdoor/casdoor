
import React from 'react';
import { Component } from 'react';
import { Layout, Menu, Breadcrumb, Avatar } from 'antd';
import * as Setting from "../Setting.js";
import CustomGithubCorner from "../CustomGithubCorner";
import {Link, Redirect, Route, Switch, withRouter} from 'react-router-dom'
import i18next from 'i18next';
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
    console.log(this.props.children);
    return (
      <Layout style={{ minHeight: '100vh' }}>
        <Sider  collapsible collapsed={collapsed} onCollapse={this.onCollapse} theme ='light'>
          <Link to={"/"}><div className='logo' key="logo" /></Link>
          <div>
          <Menu theme="light" mode="inline" defaultSelectedKeys={['1']}>
            <Menu.Item key="1">
              <span>{i18next.t("general:Home")}</span>
              <Link to={"/"}></Link>
            </Menu.Item>
            <SubMenu
              key="sub1"
              title={<span><span>Forms</span></span>}
            >
               <Menu.Item key="2">{i18next.t("general:Organizations")}<Link to="/organizations"></Link></Menu.Item>
               <Menu.Item key="3">{i18next.t("general:Users")}<Link to="/Users"></Link></Menu.Item>
               <Menu.Item key="4">{i18next.t("general:Providers")}<Link to="/Providers"></Link></Menu.Item>
               <Menu.Item key="5">{i18next.t("general:Applications")}<Link to="/Applications"></Link></Menu.Item>
               <Menu.Item key="6">{i18next.t("general:Tokens")}<Link to="/Tokens"></Link></Menu.Item>
            </SubMenu>
            <Menu.Item key="7" onClick={() => window.location.href = "/swagger"}>
              <span>Swagger</span>
            </Menu.Item>
          </Menu>
          </div>
        </Sider>
        <Layout>
          <Header style={{ background: '#fff', textAlign: 'right', padding: 0 }}>
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